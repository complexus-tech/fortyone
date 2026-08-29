package emailreply

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"strings"
)

const (
	Provider                   = "brevo_email"
	InboundProcessedEvent      = "inboundEmailProcessed"
	MaximumInboundWebhookBytes = 5 << 20
	maxInboundItemsPerBatch    = 100
	maxInboundMailboxes        = 100
	maxReplyTokenLength        = 256
	minReplyTokenLength        = 16
	replyAddressPrefix         = "maya+"
)

var (
	ErrInvalidPayload    = errors.New("invalid Brevo inbound email payload")
	ErrInvalidReplyToken = errors.New("invalid Maya email reply token")
)

type InboundWebhook struct {
	Items []json.RawMessage `json:"items"`
}

type InboundEmail struct {
	UUIDs                      []string  `json:"Uuid"`
	MessageID                  string    `json:"MessageId"`
	InReplyTo                  *string   `json:"InReplyTo"`
	From                       Mailbox   `json:"From"`
	To                         Mailboxes `json:"To"`
	Recipients                 Mailboxes `json:"Recipients"`
	Cc                         Mailboxes `json:"Cc"`
	ReplyTo                    *Mailbox  `json:"ReplyTo"`
	SentAtDate                 string    `json:"SentAtDate"`
	Subject                    string    `json:"Subject"`
	RawHTMLBody                *string   `json:"RawHtmlBody"`
	RawTextBody                *string   `json:"RawTextBody"`
	ExtractedMarkdownMessage   string    `json:"ExtractedMarkdownMessage"`
	ExtractedMarkdownSignature *string   `json:"ExtractedMarkdownSignature"`
	Attachments                []struct {
		Name          string `json:"Name"`
		ContentType   string `json:"ContentType"`
		ContentLength int64  `json:"ContentLength"`
		ContentID     string `json:"ContentID"`
		DownloadToken string `json:"DownloadToken"`
	} `json:"Attachments"`
}

type Mailbox struct {
	Name    string `json:"Name,omitempty"`
	Address string `json:"Address"`
}

func (m *Mailbox) UnmarshalJSON(data []byte) error {
	if m == nil {
		return errors.New("cannot decode mailbox into nil receiver")
	}
	var address string
	if err := json.Unmarshal(data, &address); err == nil {
		m.Address = strings.TrimSpace(address)
		return nil
	}
	type mailboxAlias Mailbox
	var decoded mailboxAlias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return fmt.Errorf("decode mailbox: %w", err)
	}
	*m = Mailbox(decoded)
	m.Name = strings.TrimSpace(m.Name)
	m.Address = strings.TrimSpace(m.Address)
	return nil
}

type Mailboxes []Mailbox

func (m *Mailboxes) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		*m = nil
		return nil
	}
	var mailboxes []Mailbox
	if err := json.Unmarshal(data, &mailboxes); err == nil {
		*m = mailboxes
		return nil
	}
	var single Mailbox
	if err := json.Unmarshal(data, &single); err != nil {
		return fmt.Errorf("decode mailbox list: %w", err)
	}
	*m = []Mailbox{single}
	return nil
}

func decodeInboundWebhook(raw []byte) (InboundWebhook, error) {
	if len(raw) == 0 || len(raw) > MaximumInboundWebhookBytes {
		return InboundWebhook{}, fmt.Errorf("%w: body must be between 1 and %d bytes", ErrInvalidPayload, MaximumInboundWebhookBytes)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var payload InboundWebhook
	if err := decoder.Decode(&payload); err != nil {
		return InboundWebhook{}, fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return InboundWebhook{}, err
	}
	if len(payload.Items) == 0 {
		return InboundWebhook{}, fmt.Errorf("%w: items must not be empty", ErrInvalidPayload)
	}
	if len(payload.Items) > maxInboundItemsPerBatch {
		return InboundWebhook{}, fmt.Errorf("%w: items exceeds limit of %d", ErrInvalidPayload, maxInboundItemsPerBatch)
	}
	return payload, nil
}

func decodeInboundEmail(raw json.RawMessage) (InboundEmail, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	var email InboundEmail
	if err := decoder.Decode(&email); err != nil {
		return InboundEmail{}, fmt.Errorf("%w: decode item: %v", ErrInvalidPayload, err)
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return InboundEmail{}, err
	}
	if strings.TrimSpace(email.From.Address) == "" {
		return InboundEmail{}, fmt.Errorf("%w: item sender is required", ErrInvalidPayload)
	}
	if len(email.To) == 0 && len(email.Recipients) == 0 {
		return InboundEmail{}, fmt.Errorf("%w: item recipient is required", ErrInvalidPayload)
	}
	if len(email.To)+len(email.Recipients)+len(email.Cc) > maxInboundMailboxes {
		return InboundEmail{}, fmt.Errorf("%w: item mailbox count exceeds limit of %d", ErrInvalidPayload, maxInboundMailboxes)
	}
	return email, nil
}

func ensureJSONEOF(decoder *json.Decoder) error {
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err == nil {
		return fmt.Errorf("%w: multiple JSON values", ErrInvalidPayload)
	} else if !errors.Is(err, io.EOF) {
		return fmt.Errorf("%w: %v", ErrInvalidPayload, err)
	}
	return nil
}

func extractReplyToken(email InboundEmail) (string, error) {
	addresses := make([]Mailbox, 0, len(email.Recipients)+len(email.To))
	addresses = append(addresses, email.Recipients...)
	addresses = append(addresses, email.To...)

	tokens := make(map[string]struct{})
	for _, mailbox := range addresses {
		address, err := parsedEmailAddress(mailbox.Address)
		if err != nil {
			continue
		}
		at := strings.LastIndexByte(address, '@')
		if at <= 0 {
			continue
		}
		localPart := address[:at]
		if !strings.HasPrefix(strings.ToLower(localPart), replyAddressPrefix) {
			continue
		}
		token := localPart[len(replyAddressPrefix):]
		if !validReplyToken(token) {
			return "", ErrInvalidReplyToken
		}
		tokens[token] = struct{}{}
	}
	if len(tokens) != 1 {
		return "", ErrInvalidReplyToken
	}
	for token := range tokens {
		return token, nil
	}
	return "", ErrInvalidReplyToken
}

func validReplyToken(token string) bool {
	if len(token) < minReplyTokenLength || len(token) > maxReplyTokenLength {
		return false
	}
	for _, char := range token {
		if (char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') || char == '-' || char == '_' {
			continue
		}
		return false
	}
	return true
}

func normalizedEmailAddress(value string) (string, error) {
	address, err := parsedEmailAddress(value)
	if err != nil {
		return "", err
	}
	return strings.ToLower(address), nil
}

func parsedEmailAddress(value string) (string, error) {
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil {
		return "", fmt.Errorf("parse email address: %w", err)
	}
	address := strings.TrimSpace(parsed.Address)
	if address == "" {
		return "", errors.New("email address is empty")
	}
	return address, nil
}

func externalEventID(email InboundEmail, raw json.RawMessage) string {
	messageID := strings.TrimSpace(email.MessageID)
	if messageID != "" && len(messageID) <= 998 {
		return messageID
	}
	canonical := raw
	var payload any
	if err := json.Unmarshal(raw, &payload); err == nil {
		if encoded, marshalErr := json.Marshal(payload); marshalErr == nil {
			canonical = encoded
		}
	}
	digest := sha256.Sum256(canonical)
	return "sha256:" + hex.EncodeToString(digest[:])
}
