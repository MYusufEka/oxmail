package health

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestPostfixChecker_Success(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("220 smtp.example.com ESMTP Postfix\r\n"))
		buf := make([]byte, 256)
		conn.Read(buf)
		conn.Write([]byte("250-smtp.example.com\r\n250 HELP\r\n"))
	}()

	checker := &PostfixChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, latency := checker.Check(context.Background())
	if !healthy {
		t.Error("PostfixChecker should be healthy")
	}
	if latency < 0 {
		t.Error("latency should be >= 0")
	}
}

func TestPostfixChecker_InvalidGreeting(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("421 Service not available\r\n"))
	}()

	checker := &PostfixChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("PostfixChecker should be unhealthy with non-220 greeting")
	}
}

func TestPostfixChecker_ConnectionRefused(t *testing.T) {
	checker := &PostfixChecker{Address: "127.0.0.1:9999", Timeout: 1 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("PostfixChecker should be unhealthy on connection refused")
	}
}

func TestDovecotChecker_Success(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("* OK [CAPABILITY IMAP4rev1] Dovecot ready.\r\n"))
	}()

	checker := &DovecotChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, latency := checker.Check(context.Background())
	if !healthy {
		t.Error("DovecotChecker should be healthy")
	}
	if latency < 0 {
		t.Error("latency should be >= 0")
	}
}

func TestDovecotChecker_InvalidGreeting(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		conn.Write([]byte("* BYE Server shutting down\r\n"))
	}()

	checker := &DovecotChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("DovecotChecker should be unhealthy without * OK greeting")
	}
}

func TestDovecotChecker_ConnectionRefused(t *testing.T) {
	checker := &DovecotChecker{Address: "127.0.0.1:9998", Timeout: 1 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("DovecotChecker should be unhealthy on connection refused")
	}
}

func TestRspamdChecker_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/ping" {
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, "pong")
		}
	}))
	defer server.Close()

	checker := &RspamdChecker{URL: server.URL + "/ping", Timeout: 2 * time.Second}
	healthy, latency := checker.Check(context.Background())
	if !healthy {
		t.Error("RspamdChecker should be healthy")
	}
	if latency < 0 {
		t.Error("latency should be >= 0")
	}
}

func TestRspamdChecker_NonOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	checker := &RspamdChecker{URL: server.URL, Timeout: 2 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("RspamdChecker should be unhealthy on non-200 status")
	}
}

func TestRspamdChecker_URLNotFound(t *testing.T) {
	checker := &RspamdChecker{URL: "http://127.0.0.1:9997/ping", Timeout: 1 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("RspamdChecker should be unhealthy on connection error")
	}
}

func TestRedisChecker_Success(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 256)
		conn.Read(buf)
		conn.Write([]byte("+PONG\r\n"))
	}()

	checker := &RedisChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, latency := checker.Check(context.Background())
	if !healthy {
		t.Error("RedisChecker should be healthy")
	}
	if latency < 0 {
		t.Error("latency should be >= 0")
	}
}

func TestRedisChecker_InvalidResponse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer listener.Close()

	addr := listener.Addr().String()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 256)
		conn.Read(buf)
		conn.Write([]byte("-ERR unknown command\r\n"))
	}()

	checker := &RedisChecker{Address: addr, Timeout: 2 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("RedisChecker should be unhealthy without +PONG")
	}
}

func TestRedisChecker_ConnectionRefused(t *testing.T) {
	checker := &RedisChecker{Address: "127.0.0.1:9996", Timeout: 1 * time.Second}
	healthy, _ := checker.Check(context.Background())
	if healthy {
		t.Error("RedisChecker should be unhealthy on connection refused")
	}
}
