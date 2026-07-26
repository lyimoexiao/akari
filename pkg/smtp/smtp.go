package smtp

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	smtplib "net/smtp"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	defaultTimeout   = 10 * time.Second
	defaultQueueSize = 100
)

var (
	ErrQueueFull = errors.New("smtp queue is full")
	ErrClosed    = errors.New("smtp mailer is closed")
)

// Mailer queues email for bounded background SMTP delivery.
type Mailer struct {
	cfg     Config
	auth    smtplib.Auth
	addr    string
	timeout time.Duration
	jobs    chan outboundMessage
	cancel  context.CancelFunc
	logger  *zap.SugaredLogger

	mu     sync.RWMutex
	closed bool
	wg     sync.WaitGroup
}

type outboundMessage struct {
	to   []string
	data []byte
}

// New creates a Mailer and starts its delivery worker.
func New(cfg *Config, logger *zap.SugaredLogger) *Mailer {
	var auth smtplib.Auth
	if cfg.Username != "" {
		auth = smtplib.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}
	timeout := time.Duration(cfg.Timeout) * time.Second
	if timeout <= 0 {
		timeout = defaultTimeout
	}
	queueSize := cfg.QueueSize
	if queueSize <= 0 {
		queueSize = defaultQueueSize
	}
	workerCtx, cancel := context.WithCancel(context.Background())
	mailer := &Mailer{
		cfg:     *cfg,
		auth:    auth,
		addr:    net.JoinHostPort(cfg.Host, cfg.Port),
		timeout: timeout,
		jobs:    make(chan outboundMessage, queueSize),
		cancel:  cancel,
		logger:  logger,
	}
	mailer.wg.Add(1)
	go mailer.run(workerCtx)
	return mailer
}

// Send queues a plain-text email.
func (m *Mailer) Send(to []string, subject, body string) error {
	return m.send(to, m.buildMessage(to, subject, body, "text/plain; charset=UTF-8"))
}

// SendHTML queues an HTML email.
func (m *Mailer) SendHTML(to []string, subject, htmlBody string) error {
	return m.send(to, m.buildMessage(to, subject, htmlBody, "text/html; charset=UTF-8"))
}

// SendWithTemplate queues an email with a custom Content-Type.
func (m *Mailer) SendWithTemplate(to []string, subject, body, contentType string) error {
	return m.send(to, m.buildMessage(to, subject, body, contentType))
}

func (m *Mailer) send(to []string, data []byte) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	if m.closed {
		return ErrClosed
	}
	message := outboundMessage{
		to:   append([]string(nil), to...),
		data: append([]byte(nil), data...),
	}
	select {
	case m.jobs <- message:
		return nil
	default:
		return ErrQueueFull
	}
}

// Close stops the worker and cancels any active SMTP operation.
func (m *Mailer) Close() error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}
	m.closed = true
	m.cancel()
	m.mu.Unlock()
	m.wg.Wait()
	return nil
}

func (m *Mailer) run(ctx context.Context) {
	defer m.wg.Done()
	for {
		select {
		case <-ctx.Done():
			return
		case message := <-m.jobs:
			if err := m.deliver(ctx, message); err != nil && ctx.Err() == nil {
				m.logger.Errorw("smtp delivery failed", "recipients", len(message.to), "error", err)
			}
		}
	}
}

func (m *Mailer) deliver(parent context.Context, message outboundMessage) error {
	ctx, cancel := context.WithTimeout(parent, m.timeout)
	defer cancel()

	conn, err := (&net.Dialer{}).DialContext(ctx, "tcp", m.addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer conn.Close()
	deadline, ok := ctx.Deadline()
	if ok {
		if err := conn.SetDeadline(deadline); err != nil {
			return fmt.Errorf("smtp deadline: %w", err)
		}
	}
	stopDeadline := context.AfterFunc(ctx, func() {
		_ = conn.SetDeadline(time.Now())
	})
	defer stopDeadline()

	if m.cfg.SSL {
		return m.sendSSL(ctx, conn, message)
	}
	return m.sendSTARTTLS(conn, message)
}

func (m *Mailer) sendSTARTTLS(conn net.Conn, message outboundMessage) error {
	client, err := smtplib.NewClient(conn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()
	if err := client.StartTLS(m.tlsConfig()); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}
	return m.sendMessage(client, message)
}

func (m *Mailer) sendSSL(ctx context.Context, conn net.Conn, message outboundMessage) error {
	tlsConn := tls.Client(conn, m.tlsConfig())
	if err := tlsConn.HandshakeContext(ctx); err != nil {
		return fmt.Errorf("smtp tls handshake: %w", err)
	}
	client, err := smtplib.NewClient(tlsConn, m.cfg.Host)
	if err != nil {
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()
	return m.sendMessage(client, message)
}

func (m *Mailer) sendMessage(client *smtplib.Client, message outboundMessage) error {
	if m.auth != nil {
		if err := client.Auth(m.auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}
	if err := client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, recipient := range message.to {
		if err := client.Rcpt(recipient); err != nil {
			return fmt.Errorf("smtp rcpt %q: %w", recipient, err)
		}
	}
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err := writer.Write(message.data); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}

func (m *Mailer) buildMessage(to []string, subject, body, contentType string) []byte {
	headers := []string{
		fmt.Sprintf("From: %s", m.cfg.From),
		fmt.Sprintf("To: %s", strings.Join(to, ", ")),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		fmt.Sprintf("Content-Type: %s", contentType),
		fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)),
		"",
		body,
	}
	return []byte(strings.Join(headers, "\r\n"))
}

func (m *Mailer) tlsConfig() *tls.Config {
	return &tls.Config{ServerName: m.cfg.Host, MinVersion: tls.VersionTLS12}
}