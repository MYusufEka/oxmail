package mail

import (
	"fmt"
	"io"
	"mime"
	"net"
	"strings"
	"time"

	"github.com/emersion/go-imap/v2"
	"github.com/emersion/go-imap/v2/imapclient"
	gomessage "github.com/emersion/go-message/mail"

	"github.com/MYusufEka/oxmail/internal/domain"
)

// goIMAPClient wraps the go-imap v2 client to implement imapClient.
type goIMAPClient struct {
	client *imapclient.Client
}

func newGoIMAPClient(addr string) (imapClient, error) {
	conn, err := net.DialTimeout("tcp", addr, 10*time.Second)
	if err != nil {
		return nil, fmt.Errorf("dial %s: %w", addr, err)
	}

	client := imapclient.New(conn, nil)
	return &goIMAPClient{client: client}, nil
}

func (c *goIMAPClient) Login(user, password string) error {
	cmd := c.client.Login(user, password)
	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("login failed: %w", err)
	}
	return nil
}

func (c *goIMAPClient) FetchMessages(page, limit int) ([]domain.MailMessage, int, error) {
	selectCmd := c.client.Select("INBOX", nil)
	mbox, err := selectCmd.Wait()
	if err != nil {
		return nil, 0, fmt.Errorf("select INBOX: %w", err)
	}

	total := int(mbox.NumMessages)
	if total == 0 {
		return []domain.MailMessage{}, 0, nil
	}

	// Calculate sequence range for pagination (newest first)
	end := total - (page-1)*limit
	start := end - limit + 1
	if start < 1 {
		start = 1
	}
	if end < 1 {
		return []domain.MailMessage{}, total, nil
	}

	seqSet := imap.SeqSet{}
	seqSet.AddRange(uint32(start), uint32(end))

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		UID:      true,
	}

	fetchCmd := c.client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	var messages []domain.MailMessage
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}

		mailMsg := envelopeToMailMessage(msg)
		messages = append(messages, mailMsg)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, 0, fmt.Errorf("fetch messages: %w", err)
	}

	// Reverse to show newest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, total, nil
}

func (c *goIMAPClient) FetchMessage(uid uint32) (*domain.MailMessage, error) {
	selectCmd := c.client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		return nil, fmt.Errorf("select INBOX: %w", err)
	}

	seqSet := imap.UIDSet{}
	seqSet.AddNum(imap.UID(uid))

	fetchOptions := &imap.FetchOptions{
		Envelope:    true,
		Flags:       true,
		UID:         true,
		BodySection: []*imap.FetchItemBodySection{{}},
	}

	fetchCmd := c.client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	msg := fetchCmd.Next()
	if msg == nil {
		fetchCmd.Close()
		return nil, fmt.Errorf("message not found: UID %d", uid)
	}

	mailMsg := envelopeToMailMessage(msg)

	// Parse MIME body
	for {
		item := msg.Next()
		if item == nil {
			break
		}
		section, ok := item.(imapclient.FetchItemDataBodySection)
		if !ok {
			continue
		}
		parseMIMEBody(&mailMsg, section.Literal)
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, fmt.Errorf("fetch message: %w", err)
	}

	return &mailMsg, nil
}

func (c *goIMAPClient) DeleteMessage(uid uint32) error {
	selectCmd := c.client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		return fmt.Errorf("select INBOX: %w", err)
	}

	seqSet := imap.UIDSet{}
	seqSet.AddNum(imap.UID(uid))

	storeCmd := c.client.Store(seqSet, &imap.StoreFlags{
		Op:    imap.StoreFlagsAdd,
		Flags: []imap.Flag{imap.FlagDeleted},
	}, nil)
	// Consume the store response
	if _, err := storeCmd.Collect(); err != nil {
		return fmt.Errorf("flag deleted: %w", err)
	}

	expungeCmd := c.client.Expunge()
	if _, err := expungeCmd.Collect(); err != nil {
		return fmt.Errorf("expunge: %w", err)
	}

	return nil
}

func (c *goIMAPClient) MarkRead(uid uint32, read bool) error {
	selectCmd := c.client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		return fmt.Errorf("select INBOX: %w", err)
	}

	seqSet := imap.UIDSet{}
	seqSet.AddNum(imap.UID(uid))

	op := imap.StoreFlagsAdd
	if !read {
		op = imap.StoreFlagsDel
	}

	storeCmd := c.client.Store(seqSet, &imap.StoreFlags{
		Op:    op,
		Flags: []imap.Flag{imap.FlagSeen},
	}, nil)
	// Consume the store response
	if _, err := storeCmd.Collect(); err != nil {
		return fmt.Errorf("store seen flag: %w", err)
	}

	return nil
}

func (c *goIMAPClient) SearchMessages(query string) ([]domain.MailMessage, error) {
	selectCmd := c.client.Select("INBOX", nil)
	if _, err := selectCmd.Wait(); err != nil {
		return nil, fmt.Errorf("select INBOX: %w", err)
	}

	criteria := &imap.SearchCriteria{
		Body: []string{query},
	}

	searchCmd := c.client.Search(criteria, nil)
	searchData, err := searchCmd.Wait()
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	allSeqNums := searchData.AllSeqNums()
	if len(allSeqNums) == 0 {
		return []domain.MailMessage{}, nil
	}

	seqSet := imap.SeqSet{}
	for _, num := range allSeqNums {
		seqSet.AddNum(num)
	}

	fetchOptions := &imap.FetchOptions{
		Envelope: true,
		Flags:    true,
		UID:      true,
	}

	fetchCmd := c.client.Fetch(seqSet, fetchOptions)
	defer fetchCmd.Close()

	var messages []domain.MailMessage
	for {
		msg := fetchCmd.Next()
		if msg == nil {
			break
		}
		messages = append(messages, envelopeToMailMessage(msg))
	}

	if err := fetchCmd.Close(); err != nil {
		return nil, fmt.Errorf("fetch search results: %w", err)
	}

	return messages, nil
}

func (c *goIMAPClient) Logout() error {
	cmd := c.client.Logout()
	return cmd.Wait()
}

func envelopeToMailMessage(msg *imapclient.FetchMessageData) domain.MailMessage {
	var mailMsg domain.MailMessage

	for {
		item := msg.Next()
		if item == nil {
			break
		}

		switch data := item.(type) {
		case imapclient.FetchItemDataUID:
			mailMsg.ID = int64(data.UID)
		case imapclient.FetchItemDataFlags:
			for _, flag := range data.Flags {
				if flag == imap.FlagSeen {
					mailMsg.Read = true
				}
			}
		case imapclient.FetchItemDataEnvelope:
			env := data.Envelope
			mailMsg.Subject = env.Subject
			mailMsg.ReceivedAt = env.Date
			mailMsg.MessageID = env.MessageID
			if len(env.InReplyTo) > 0 {
				mailMsg.InReplyTo = env.InReplyTo[0]
			}

			if len(env.From) > 0 {
				mailMsg.From = formatAddress(env.From[0])
			}
			for _, addr := range env.To {
				mailMsg.To = append(mailMsg.To, formatAddress(addr))
			}
			for _, addr := range env.Cc {
				mailMsg.CC = append(mailMsg.CC, formatAddress(addr))
			}

			// Simple thread grouping by In-Reply-To
			if len(env.InReplyTo) > 0 {
				mailMsg.ThreadID = env.InReplyTo[0]
			} else {
				mailMsg.ThreadID = env.MessageID
			}
		}
	}

	return mailMsg
}

func formatAddress(addr imap.Address) string {
	if addr.Name != "" {
		return fmt.Sprintf("%s <%s@%s>", addr.Name, addr.Mailbox, addr.Host)
	}
	return fmt.Sprintf("%s@%s", addr.Mailbox, addr.Host)
}

func parseMIMEBody(msg *domain.MailMessage, body io.Reader) {
	if body == nil {
		return
	}

	reader, err := gomessage.CreateReader(body)
	if err != nil {
		return
	}
	defer reader.Close()

	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}

		contentType, params, _ := mime.ParseMediaType(part.Header.Get("Content-Type"))
		disposition, dparams, _ := mime.ParseMediaType(part.Header.Get("Content-Disposition"))

		if disposition == "attachment" {
			filename := dparams["filename"]
			if filename == "" {
				filename = params["name"]
			}
			content, _ := io.ReadAll(part.Body)
			msg.Attachments = append(msg.Attachments, domain.Attachment{
				Filename: filename,
				Size:     int64(len(content)),
			})
			continue
		}

		switch {
		case strings.HasPrefix(contentType, "text/plain"):
			content, _ := io.ReadAll(part.Body)
			msg.BodyText = string(content)
		case strings.HasPrefix(contentType, "text/html"):
			content, _ := io.ReadAll(part.Body)
			msg.BodyHTML = string(content)
		}
	}
}
