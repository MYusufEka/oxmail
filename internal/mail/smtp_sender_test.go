package mail

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSMTPServer starts a fake SMTP server that accepts messages.
func mockSMTPServer(t *testing.T) (addr string, messages *[]string) {
	t.Helper()
	msgs := make([]string, 0)
	messages = &msgs

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	addr = listener.Addr().String()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			go handleSMTPConn(conn, &msgs)
		}
	}()

	t.Cleanup(func() { listener.Close() })
	return addr, messages
}

func handleSMTPConn(conn net.Conn, messages *[]string) {
	defer conn.Close()

	write := func(msg string) {
		conn.Write([]byte(msg + "\r\n"))
	}

	read := func() string {
		buf := make([]byte, 4096)
		n, _ := conn.Read(buf)
		return string(buf[:n])
	}

	write("220 localhost ESMTP MockServer")

	for {
		line := read()
		upper := strings.ToUpper(strings.TrimSpace(line))

		switch {
		case strings.HasPrefix(upper, "EHLO"):
			write("250-localhost")
			write("250 OK")
		case strings.HasPrefix(upper, "MAIL FROM:"):
			write("250 OK")
		case strings.HasPrefix(upper, "RCPT TO:"):
			write("250 OK")
		case strings.HasPrefix(upper, "DATA"):
			write("354 Start mail input")
			// Read until we get a line with just "."
			var body strings.Builder
			for {
				chunk := read()
				body.WriteString(chunk)
				if strings.Contains(chunk, "\r\n.\r\n") {
					break
				}
			}
			*messages = append(*messages, body.String())
			write("250 OK")
		case strings.HasPrefix(upper, "QUIT"):
			write("221 Bye")
			return
		default:
			write("250 OK")
		}
	}
}

func TestSMTPSender_Send_Success(t *testing.T) {
	addr, messages := mockSMTPServer(t)

	sender := NewSMTPSender(SMTPSenderConfig{
		Host: strings.Split(addr, ":")[0],
		Port: strings.Split(addr, ":")[1],
	})

	messageID, err := sender.Send(
		"user@example.com",
		[]string{"recipient@example.com"},
		nil,
		"Test Subject",
		"Hello plain text",
		"<p>Hello HTML</p>",
	)

	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
	assert.Contains(t, messageID, "@")
	assert.True(t, strings.HasPrefix(messageID, "<"))
	assert.True(t, strings.HasSuffix(messageID, ">"))

	require.Len(t, *messages, 1)
	msg := (*messages)[0]
	assert.Contains(t, msg, "From: user@example.com")
	assert.Contains(t, msg, "To: recipient@example.com")
	assert.Contains(t, msg, "Subject: Test Subject")
	assert.Contains(t, msg, "MIME-Version: 1.0")
	assert.Contains(t, msg, "multipart/alternative")
	assert.Contains(t, msg, "Hello plain text")
	assert.Contains(t, msg, "<p>Hello HTML</p>")
}

func TestSMTPSender_Send_MultipleRecipients(t *testing.T) {
	addr, messages := mockSMTPServer(t)

	sender := NewSMTPSender(SMTPSenderConfig{
		Host: strings.Split(addr, ":")[0],
		Port: strings.Split(addr, ":")[1],
	})

	messageID, err := sender.Send(
		"sender@example.com",
		[]string{"a@example.com", "b@example.com"},
		[]string{"cc@example.com"},
		"Multi Recipient",
		"body text",
		"",
	)

	require.NoError(t, err)
	assert.NotEmpty(t, messageID)
	require.Len(t, *messages, 1)
	msg := (*messages)[0]
	assert.Contains(t, msg, "To: a@example.com, b@example.com")
	assert.Contains(t, msg, "Cc: cc@example.com")
}

func TestSMTPSender_Send_TextOnly(t *testing.T) {
	addr, messages := mockSMTPServer(t)

	sender := NewSMTPSender(SMTPSenderConfig{
		Host: strings.Split(addr, ":")[0],
		Port: strings.Split(addr, ":")[1],
	})

	_, err := sender.Send(
		"sender@example.com",
		[]string{"to@example.com"},
		nil,
		"Plain Only",
		"just text",
		"",
	)

	require.NoError(t, err)
	require.Len(t, *messages, 1)
	msg := (*messages)[0]
	assert.Contains(t, msg, "text/plain")
	assert.Contains(t, msg, "just text")
}

func TestSMTPSender_Send_ConnectionError(t *testing.T) {
	sender := NewSMTPSender(SMTPSenderConfig{
		Host: "127.0.0.1",
		Port: "19999",
	})

	_, err := sender.Send(
		"user@example.com",
		[]string{"to@example.com"},
		nil,
		"Subject",
		"body",
		"",
	)

	assert.Error(t, err)
}

func TestSMTPSender_Send_EmptyRecipients(t *testing.T) {
	sender := NewSMTPSender(SMTPSenderConfig{
		Host: "127.0.0.1",
		Port: "587",
	})

	_, err := sender.Send(
		"user@example.com",
		nil,
		nil,
		"Subject",
		"body",
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one recipient")
}

func TestSMTPSender_Send_EmptyFrom(t *testing.T) {
	sender := NewSMTPSender(SMTPSenderConfig{
		Host: "127.0.0.1",
		Port: "587",
	})

	_, err := sender.Send(
		"",
		[]string{"to@example.com"},
		nil,
		"Subject",
		"body",
		"",
	)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "from address is required")
}

// SMTPSenderInterface allows mocking in handler tests.
type SMTPSenderInterface interface {
	Send(from string, to []string, cc []string, subject, bodyText, bodyHTML string) (string, error)
}
