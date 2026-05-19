package health

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// PostfixChecker checks Postfix via SMTP EHLO handshake.
type PostfixChecker struct {
	Address string
	Timeout time.Duration
}

func (c *PostfixChecker) Check(ctx context.Context) (bool, int) {
	start := time.Now()
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", c.Address, timeout)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	// Read greeting
	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	greeting := string(buf[:n])
	if !strings.HasPrefix(greeting, "220") {
		return false, int(time.Since(start).Milliseconds())
	}

	// Send EHLO
	fmt.Fprintf(conn, "EHLO healthcheck\r\n")
	n, err = conn.Read(buf)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	response := string(buf[:n])
	if !strings.Contains(response, "250") {
		return false, int(time.Since(start).Milliseconds())
	}

	// Send QUIT
	fmt.Fprintf(conn, "QUIT\r\n")

	return true, int(time.Since(start).Milliseconds())
}

// DovecotChecker checks Dovecot via IMAP greeting.
type DovecotChecker struct {
	Address string
	Timeout time.Duration
}

func (c *DovecotChecker) Check(ctx context.Context) (bool, int) {
	start := time.Now()
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", c.Address, timeout)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	buf := make([]byte, 512)
	n, err := conn.Read(buf)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}

	greeting := string(buf[:n])
	if !strings.Contains(greeting, "* OK") {
		return false, int(time.Since(start).Milliseconds())
	}

	return true, int(time.Since(start).Milliseconds())
}

// RspamdChecker checks Rspamd via HTTP ping endpoint.
type RspamdChecker struct {
	URL     string
	Timeout time.Duration
}

func (c *RspamdChecker) Check(ctx context.Context) (bool, int) {
	start := time.Now()
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(c.URL)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}

	if !strings.Contains(string(body), "pong") {
		return false, int(time.Since(start).Milliseconds())
	}

	return true, int(time.Since(start).Milliseconds())
}

// RedisChecker checks Redis via PING command.
type RedisChecker struct {
	Address string
	Timeout time.Duration
}

func (c *RedisChecker) Check(ctx context.Context) (bool, int) {
	start := time.Now()
	timeout := c.Timeout
	if timeout == 0 {
		timeout = 3 * time.Second
	}

	conn, err := net.DialTimeout("tcp", c.Address, timeout)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}
	defer conn.Close()

	conn.SetDeadline(time.Now().Add(timeout))

	// Send PING in Redis protocol
	fmt.Fprintf(conn, "*1\r\n$4\r\nPING\r\n")

	buf := make([]byte, 128)
	n, err := conn.Read(buf)
	if err != nil {
		return false, int(time.Since(start).Milliseconds())
	}

	response := string(buf[:n])
	if !strings.Contains(response, "+PONG") {
		return false, int(time.Since(start).Milliseconds())
	}

	return true, int(time.Since(start).Milliseconds())
}
