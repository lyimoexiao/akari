package smtp

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"github.com/lyimoexiao/akari/internal/config"
)

// Mailer wraps an SMTP client configuration and provides convenience
// methods for sending plain-text and HTML emails.
type Mailer struct {
	cfg  config.SMTPConfig
	auth smtp.Auth
	addr string
}

// New creates a new Mailer from the SMTP configuration block.
func New(cfg *config.SMTPConfig) *Mailer {
	addr := net.JoinHostPort(cfg.Host, cfg.Port)

	var auth smtp.Auth
	if cfg.Username != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.Host)
	}

	return &Mailer{
		cfg:  *cfg,
		auth: auth,
		addr: addr,
	}
}

// Send sends a plain-text email.
//   - to:   one or more recipient addresses
//   - subject:  email subject
//   - body: plain-text body
func (m *Mailer) Send(to []string, subject, body string) error {
	msg := m.buildMessage(to, subject, body, "text/plain; charset=UTF-8")
	return m.send(to, msg)
}

// SendHTML sends an HTML email.
//   - to:   one or more recipient addresses
//   - subject:  email subject
//   - body: HTML body
func (m *Mailer) SendHTML(to []string, subject, htmlBody string) error {
	msg := m.buildMessage(to, subject, htmlBody, "text/html; charset=UTF-8")
	return m.send(to, msg)
}

// SendWithTemplate sends an email with a custom body and Content-Type.
// Useful when you want to send multipart messages or custom headers.
func (m *Mailer) SendWithTemplate(to []string, subject, body, contentType string) error {
	msg := m.buildMessage(to, subject, body, contentType)
	return m.send(to, msg)
}

// send delivers the raw message bytes to the SMTP server.
func (m *Mailer) send(to []string, msg []byte) error {
	if m.cfg.SSL {
		return m.sendSSL(to, msg)
	}
	return m.sendSTARTTLS(to, msg)
}

// Close is a no-op for the standard smtp client; exists for interface
// consistency (e.g. wire cleanup).
func (m *Mailer) Close() error {
	return nil
}

// ─────────────────────────────────────────────────────────────────
// Internal helpers
// ─────────────────────────────────────────────────────────────────

// buildMessage composes the full RFC 822 message with headers.
func (m *Mailer) buildMessage(to []string, subject, body, contentType string) []byte {
	headers := make([]string, 0, 8)
	headers = append(headers, fmt.Sprintf("From: %s", m.cfg.From))
	headers = append(headers, fmt.Sprintf("To: %s", strings.Join(to, ", ")))
	headers = append(headers, fmt.Sprintf("Subject: %s", subject))
	headers = append(headers, fmt.Sprintf("MIME-Version: 1.0"))
	headers = append(headers, fmt.Sprintf("Content-Type: %s", contentType))
	headers = append(headers, fmt.Sprintf("Date: %s", time.Now().Format(time.RFC1123Z)))
	headers = append(headers, "") // end of headers
	headers = append(headers, body)

	return []byte(strings.Join(headers, "\r\n"))
}

// sendSTARTTLS connects with plain SMTP then upgrades via STARTTLS.
func (m *Mailer) sendSTARTTLS(to []string, msg []byte) error {
	client, err := smtp.Dial(m.addr)
	if err != nil {
		return fmt.Errorf("smtp dial: %w", err)
	}
	defer client.Close()

	// STARTTLS
	if err = client.StartTLS(&tls.Config{ServerName: m.cfg.Host}); err != nil {
		return fmt.Errorf("smtp starttls: %w", err)
	}

	// Auth
	if m.auth != nil {
		if err = client.Auth(m.auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	// Send
	if err = client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %q: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}

// sendSSL connects over a direct TLS connection (SMTPS, port 465).
func (m *Mailer) sendSSL(to []string, msg []byte) error {
	conn, err := tls.Dial("tcp", m.addr, &tls.Config{ServerName: m.cfg.Host})
	if err != nil {
		return fmt.Errorf("smtp tls dial: %w", err)
	}

	client, err := smtp.NewClient(conn, m.cfg.Host)
	if err != nil {
		conn.Close()
		return fmt.Errorf("smtp new client: %w", err)
	}
	defer client.Close()

	// Auth
	if m.auth != nil {
		if err = client.Auth(m.auth); err != nil {
			return fmt.Errorf("smtp auth: %w", err)
		}
	}

	// Send
	if err = client.Mail(m.cfg.From); err != nil {
		return fmt.Errorf("smtp mail from: %w", err)
	}
	for _, rcpt := range to {
		if err = client.Rcpt(rcpt); err != nil {
			return fmt.Errorf("smtp rcpt %q: %w", rcpt, err)
		}
	}
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("smtp data: %w", err)
	}
	if _, err = w.Write(msg); err != nil {
		return fmt.Errorf("smtp write: %w", err)
	}
	if err = w.Close(); err != nil {
		return fmt.Errorf("smtp data close: %w", err)
	}
	return client.Quit()
}
