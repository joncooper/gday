package gmail

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// Service wraps the Gmail API service
type Service struct {
	srv *gmail.Service
}

// Message represents a simplified email message
type Message struct {
	ID          string
	ThreadID    string
	Date        time.Time
	From        string
	To          string
	Subject     string
	Snippet     string
	Body        string
	BodyHTML    string
	Labels      []string
	Attachments []Attachment
	IsUnread    bool
}

// Attachment represents an email attachment
type Attachment struct {
	ID       string
	Filename string
	MimeType string
	Size     int64
}

// NewService creates a new Gmail service
func NewService(ctx context.Context, client *http.Client) (*Service, error) {
	srv, err := gmail.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gmail service: %w", err)
	}
	return &Service{srv: srv}, nil
}

// ListMessages lists recent emails
func (s *Service) ListMessages(ctx context.Context, maxResults int64, query string, labelIDs []string) ([]*Message, error) {
	req := s.srv.Users.Messages.List("me").MaxResults(maxResults)
	if query != "" {
		req = req.Q(query)
	}
	if len(labelIDs) > 0 {
		req = req.LabelIds(labelIDs...)
	}

	resp, err := req.Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	messages := make([]*Message, 0, len(resp.Messages))
	for _, m := range resp.Messages {
		msg, err := s.GetMessage(ctx, m.Id, false)
		if err != nil {
			continue // Skip messages that fail to load
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// GetMessage retrieves a single message
func (s *Service) GetMessage(ctx context.Context, id string, includeBody bool) (*Message, error) {
	format := "metadata"
	if includeBody {
		format = "full"
	}

	msg, err := s.srv.Users.Messages.Get("me", id).Format(format).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get message: %w", err)
	}

	return parseMessage(msg, includeBody), nil
}

// GetThread retrieves a thread with all messages
func (s *Service) GetThread(ctx context.Context, threadID string) ([]*Message, error) {
	thread, err := s.srv.Users.Threads.Get("me", threadID).Format("full").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	messages := make([]*Message, 0, len(thread.Messages))
	for _, m := range thread.Messages {
		messages = append(messages, parseMessage(m, true))
	}

	return messages, nil
}

// SearchMessages searches for messages matching a query
func (s *Service) SearchMessages(ctx context.Context, query string, maxResults int64) ([]*Message, error) {
	return s.ListMessages(ctx, maxResults, query, nil)
}

// SendMessage sends a new email
func (s *Service) SendMessage(ctx context.Context, to, subject, body string, cc, bcc []string) (*Message, error) {
	// Build the message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", to))
	if len(cc) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(cc, ", ")))
	}
	if len(bcc) > 0 {
		msgBuilder.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(bcc, ", ")))
	}
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuilder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)

	rawMsg := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))
	message := &gmail.Message{Raw: rawMsg}

	sent, err := s.srv.Users.Messages.Send("me", message).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %w", err)
	}

	return s.GetMessage(ctx, sent.Id, false)
}

// ReplyToMessage sends a reply to an existing message
func (s *Service) ReplyToMessage(ctx context.Context, messageID, body string) (*Message, error) {
	// Get original message
	orig, err := s.GetMessage(ctx, messageID, true)
	if err != nil {
		return nil, err
	}

	// Build reply subject
	subject := orig.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}

	// Get references and message-id for threading
	origMsg, err := s.srv.Users.Messages.Get("me", messageID).Format("full").Do()
	if err != nil {
		return nil, err
	}

	var messageIDHeader, references string
	for _, h := range origMsg.Payload.Headers {
		switch strings.ToLower(h.Name) {
		case "message-id":
			messageIDHeader = h.Value
		case "references":
			references = h.Value
		}
	}

	// Build new references header
	if references != "" {
		references = references + " " + messageIDHeader
	} else {
		references = messageIDHeader
	}

	// Build the reply message
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", orig.From))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuilder.WriteString(fmt.Sprintf("In-Reply-To: %s\r\n", messageIDHeader))
	msgBuilder.WriteString(fmt.Sprintf("References: %s\r\n", references))
	msgBuilder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)

	rawMsg := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))
	message := &gmail.Message{
		Raw:      rawMsg,
		ThreadId: orig.ThreadID,
	}

	sent, err := s.srv.Users.Messages.Send("me", message).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to send reply: %w", err)
	}

	return s.GetMessage(ctx, sent.Id, false)
}

// DownloadAttachment downloads an attachment to the specified directory
func (s *Service) DownloadAttachment(ctx context.Context, messageID, attachmentID, filename, outDir string) (string, error) {
	att, err := s.srv.Users.Messages.Attachments.Get("me", messageID, attachmentID).Do()
	if err != nil {
		return "", fmt.Errorf("failed to get attachment: %w", err)
	}

	data, err := base64.URLEncoding.DecodeString(att.Data)
	if err != nil {
		return "", fmt.Errorf("failed to decode attachment: %w", err)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output directory: %w", err)
	}

	// Write file
	outPath := filepath.Join(outDir, filename)
	if err := os.WriteFile(outPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write attachment: %w", err)
	}

	return outPath, nil
}

// GetLabels returns all labels
func (s *Service) GetLabels(ctx context.Context) ([]string, error) {
	resp, err := s.srv.Users.Labels.List("me").Do()
	if err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	labels := make([]string, 0, len(resp.Labels))
	for _, l := range resp.Labels {
		labels = append(labels, l.Name)
	}
	return labels, nil
}

// MarkAsRead marks a message as read
func (s *Service) MarkAsRead(ctx context.Context, messageID string) error {
	_, err := s.srv.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		RemoveLabelIds: []string{"UNREAD"},
	}).Do()
	return err
}

// MarkAsUnread marks a message as unread
func (s *Service) MarkAsUnread(ctx context.Context, messageID string) error {
	_, err := s.srv.Users.Messages.Modify("me", messageID, &gmail.ModifyMessageRequest{
		AddLabelIds: []string{"UNREAD"},
	}).Do()
	return err
}

// parseMessage converts a Gmail API message to our Message type
func parseMessage(m *gmail.Message, includeBody bool) *Message {
	msg := &Message{
		ID:       m.Id,
		ThreadID: m.ThreadId,
		Snippet:  m.Snippet,
		Labels:   m.LabelIds,
	}

	// Check if unread
	for _, l := range m.LabelIds {
		if l == "UNREAD" {
			msg.IsUnread = true
			break
		}
	}

	// Parse headers
	if m.Payload != nil {
		for _, h := range m.Payload.Headers {
			switch strings.ToLower(h.Name) {
			case "from":
				msg.From = h.Value
			case "to":
				msg.To = h.Value
			case "subject":
				msg.Subject = h.Value
			case "date":
				if t, err := parseDate(h.Value); err == nil {
					msg.Date = t
				}
			}
		}

		// Extract body if requested
		if includeBody {
			msg.Body, msg.BodyHTML = extractBody(m.Payload)
			msg.Attachments = extractAttachments(m.Payload)
		}
	}

	// Use internal date if header date failed
	if msg.Date.IsZero() && m.InternalDate > 0 {
		msg.Date = time.Unix(m.InternalDate/1000, 0)
	}

	return msg
}

// extractBody extracts the text and HTML body from message payload
func extractBody(payload *gmail.MessagePart) (string, string) {
	var textBody, htmlBody string

	// Check if this part has data
	if payload.Body != nil && payload.Body.Data != "" {
		decoded, err := base64.URLEncoding.DecodeString(payload.Body.Data)
		if err == nil {
			if strings.HasPrefix(payload.MimeType, "text/plain") {
				textBody = string(decoded)
			} else if strings.HasPrefix(payload.MimeType, "text/html") {
				htmlBody = string(decoded)
			}
		}
	}

	// Recursively check parts
	for _, part := range payload.Parts {
		text, html := extractBody(part)
		if textBody == "" {
			textBody = text
		}
		if htmlBody == "" {
			htmlBody = html
		}
	}

	// If we only have HTML, try to convert to plain text
	if textBody == "" && htmlBody != "" {
		textBody = htmlToText(htmlBody)
	}

	return textBody, htmlBody
}

// extractAttachments extracts attachment info from message payload
func extractAttachments(payload *gmail.MessagePart) []Attachment {
	var attachments []Attachment

	// Check if this part is an attachment
	if payload.Filename != "" && payload.Body != nil && payload.Body.AttachmentId != "" {
		attachments = append(attachments, Attachment{
			ID:       payload.Body.AttachmentId,
			Filename: payload.Filename,
			MimeType: payload.MimeType,
			Size:     payload.Body.Size,
		})
	}

	// Recursively check parts
	for _, part := range payload.Parts {
		attachments = append(attachments, extractAttachments(part)...)
	}

	return attachments
}

// parseDate attempts to parse various date formats
func parseDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		"Mon, 2 Jan 2006 15:04:05 -0700",
		"2 Jan 2006 15:04:05 -0700",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

// htmlToText converts HTML to plain text (basic implementation)
func htmlToText(html string) string {
	// Remove script and style elements
	re := regexp.MustCompile(`<script[^>]*>.*?</script>|<style[^>]*>.*?</style>`)
	text := re.ReplaceAllString(html, "")

	// Replace br and p tags with newlines
	text = regexp.MustCompile(`<br\s*/?>|</p>|</div>`).ReplaceAllString(text, "\n")

	// Remove all other HTML tags
	text = regexp.MustCompile(`<[^>]+>`).ReplaceAllString(text, "")

	// Decode common HTML entities
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")
	text = strings.ReplaceAll(text, "&quot;", "\"")
	text = strings.ReplaceAll(text, "&#39;", "'")

	// Clean up whitespace
	text = regexp.MustCompile(`\n{3,}`).ReplaceAllString(text, "\n\n")
	text = strings.TrimSpace(text)

	return text
}

// CreateDraft creates a draft email
func (s *Service) CreateDraft(ctx context.Context, to, subject, body string) (string, error) {
	var msgBuilder strings.Builder
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	msgBuilder.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(body)

	rawMsg := base64.URLEncoding.EncodeToString([]byte(msgBuilder.String()))
	draft := &gmail.Draft{
		Message: &gmail.Message{Raw: rawMsg},
	}

	created, err := s.srv.Users.Drafts.Create("me", draft).Do()
	if err != nil {
		return "", fmt.Errorf("failed to create draft: %w", err)
	}

	return created.Id, nil
}

// Stubbed interface for potential streaming
var _ io.Writer = (*os.File)(nil)
