package smtp

import (
	"net"
	"testing"
	"time"

	"go.uber.org/zap"
)

func Test_Mailer_Send_returns_after_enqueue_when_SMTP_stalls(t *testing.T) {
	// Given
	server := newStallingServer(t)
	mailer := New(&Config{
		Host:      server.host,
		Port:      server.port,
		From:      "sender@example.com",
		SSL:       true,
		Timeout:   1,
		QueueSize: 1,
	}, zap.NewNop().Sugar())
	t.Cleanup(func() { _ = mailer.Close() })

	// When
	result := make(chan error, 1)
	go func() {
		result <- mailer.Send([]string{"recipient@example.com"}, "subject", "body")
	}()

	// Then
	select {
	case err := <-result:
		if err != nil {
			t.Fatalf("Send returned error: %v", err)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Send blocked on SMTP delivery")
	}
}

func Test_Mailer_delivery_closes_stalled_connection_after_timeout(t *testing.T) {
	// Given
	server := newStallingServer(t)
	mailer := New(&Config{
		Host:      server.host,
		Port:      server.port,
		From:      "sender@example.com",
		SSL:       true,
		Timeout:   1,
		QueueSize: 1,
	}, zap.NewNop().Sugar())
	t.Cleanup(func() { _ = mailer.Close() })

	// When
	if err := mailer.Send([]string{"recipient@example.com"}, "subject", "body"); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	// Then
	select {
	case <-server.closed:
	case <-time.After(1500 * time.Millisecond):
		t.Fatal("SMTP connection remained open past configured timeout")
	}
}

func Test_Mailer_Send_rejects_when_bounded_queue_is_full(t *testing.T) {
	// Given
	server := newStallingServer(t)
	mailer := New(&Config{
		Host:      server.host,
		Port:      server.port,
		From:      "sender@example.com",
		SSL:       true,
		Timeout:   2,
		QueueSize: 1,
	}, zap.NewNop().Sugar())
	t.Cleanup(func() { _ = mailer.Close() })
	if err := mailer.Send([]string{"first@example.com"}, "subject", "body"); err != nil {
		t.Fatalf("enqueue first email: %v", err)
	}
	select {
	case <-server.accepted:
	case <-time.After(time.Second):
		t.Fatal("SMTP worker did not begin first delivery")
	}
	if err := mailer.Send([]string{"second@example.com"}, "subject", "body"); err != nil {
		t.Fatalf("enqueue second email: %v", err)
	}

	// When
	err := mailer.Send([]string{"third@example.com"}, "subject", "body")

	// Then
	if err != ErrQueueFull {
		t.Fatalf("Send error = %v, want %v", err, ErrQueueFull)
	}
}

type stallingServer struct {
	host     string
	port     string
	accepted <-chan struct{}
	closed   <-chan struct{}
}

func newStallingServer(t *testing.T) stallingServer {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	accepted := make(chan struct{})
	closed := make(chan struct{})
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := listener.Accept()
		if acceptErr != nil {
			close(accepted)
			close(closed)
			return
		}
		close(accepted)
		go func() {
			<-stop
			_ = conn.Close()
		}()
		buffer := make([]byte, 1)
		_, _ = conn.Read(buffer)
		close(closed)
	}()
	t.Cleanup(func() {
		close(stop)
		_ = listener.Close()
		<-done
	})

	host, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		t.Fatalf("split listener address: %v", err)
	}
	return stallingServer{host: host, port: port, accepted: accepted, closed: closed}
}
