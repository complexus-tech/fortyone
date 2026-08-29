package emailthread

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/google/uuid"
)

const (
	Provider                  = "brevo_email"
	DefaultReplyDomain        = "reply.fortyone.app"
	defaultReplyTokenLifetime = 180 * 24 * time.Hour
	defaultTokenBytes         = 32
)

type Store interface {
	CreateEmailThread(context.Context, messaging.EmailThreadInput) (messaging.EmailThreadRecord, bool, error)
	CreateEmailReplyTokenAlias(context.Context, messaging.EmailReplyTokenInput) (messaging.EmailReplyTokenRecord, bool, error)
	AppendEmailMessage(context.Context, messaging.EmailMessageInput) (messaging.EmailMessageRecord, bool, error)
}

type GuidancePreparer interface {
	PrepareGuidance(context.Context, GuidanceInput) (PreparedGuidance, error)
}

type ThreadContext struct {
	Version       int             `json:"version"`
	Source        string          `json:"source"`
	WorkspaceSlug string          `json:"workspaceSlug"`
	Targets       []TargetContext `json:"targets,omitempty"`
}

type TargetContext struct {
	Kind        string    `json:"kind"`
	ID          uuid.UUID `json:"id"`
	TeamID      uuid.UUID `json:"teamId,omitempty"`
	ParentID    uuid.UUID `json:"parentId,omitempty"`
	DisplayName string    `json:"displayName"`
}

func EncodeThreadContext(context ThreadContext) (json.RawMessage, error) {
	context.Version = 1
	context.Source = strings.TrimSpace(context.Source)
	context.WorkspaceSlug = strings.TrimSpace(context.WorkspaceSlug)
	if context.Source == "" || context.WorkspaceSlug == "" {
		return nil, errors.New("email thread context requires a source and workspace slug")
	}
	encoded, err := json.Marshal(context)
	if err != nil {
		return nil, fmt.Errorf("encode email thread context: %w", err)
	}
	return encoded, nil
}

type Service struct {
	store       Store
	domain      string
	now         func() time.Time
	randomBytes func([]byte) (int, error)
}

type GuidanceInput struct {
	WorkspaceID       uuid.UUID
	UserID            uuid.UUID
	RecipientEmail    string
	ExternalThreadID  string
	InternetMessageID string
	Subject           string
	Content           string
	Context           json.RawMessage
}

type PreparedGuidance struct {
	Thread  messaging.EmailThreadRecord
	Message messaging.EmailMessageRecord
	ReplyTo string
	Token   string
}

type ReplyInput struct {
	Thread            messaging.EmailThreadRecord
	ReplyToken        string
	InternetMessageID string
	InReplyTo         string
	Subject           string
	Content           string
	Kind              string
	IdempotencyKey    string
	Context           json.RawMessage
}

type PreparedReply struct {
	Message    messaging.EmailMessageRecord
	ReplyTo    string
	Token      string
	References []string
}

func New(store Store) (*Service, error) {
	if store == nil {
		return nil, errors.New("email thread store is required")
	}
	return &Service{
		store:       store,
		domain:      DefaultReplyDomain,
		now:         time.Now,
		randomBytes: rand.Read,
	}, nil
}

func (s *Service) PrepareGuidance(ctx context.Context, input GuidanceInput) (PreparedGuidance, error) {
	if s == nil || s.store == nil {
		return PreparedGuidance{}, errors.New("email thread service is not configured")
	}
	input = normalizeGuidanceInput(input)
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.RecipientEmail == "" ||
		input.ExternalThreadID == "" || input.InternetMessageID == "" || input.Subject == "" || input.Content == "" {
		return PreparedGuidance{}, errors.New("email guidance thread requires workspace, user, recipient, thread, message, subject, and content")
	}
	token, tokenHash, err := s.newReplyToken()
	if err != nil {
		return PreparedGuidance{}, err
	}
	expiresAt := s.now().UTC().Add(defaultReplyTokenLifetime)
	thread, _, err := s.store.CreateEmailThread(ctx, messaging.EmailThreadInput{
		Provider:              Provider,
		WorkspaceID:           input.WorkspaceID,
		UserID:                input.UserID,
		RecipientEmail:        input.RecipientEmail,
		ExternalThreadID:      input.ExternalThreadID,
		RootInternetMessageID: input.InternetMessageID,
		Context:               objectOrEmpty(input.Context),
		ReplyTokenHash:        tokenHash,
		ReplyTokenExpiresAt:   expiresAt,
	})
	if err != nil {
		return PreparedGuidance{}, fmt.Errorf("prepare Maya email thread: %w", err)
	}
	message, _, err := s.store.AppendEmailMessage(ctx, messaging.EmailMessageInput{
		ThreadID:          thread.ID,
		WorkspaceID:       thread.WorkspaceID,
		UserID:            thread.UserID,
		IdempotencyKey:    "guidance:" + input.InternetMessageID,
		Direction:         messaging.EmailMessageDirectionOutbound,
		Role:              messaging.EmailMessageRoleAssistant,
		Kind:              messaging.EmailMessageKindGuidance,
		InternetMessageID: input.InternetMessageID,
		Subject:           input.Subject,
		Content:           input.Content,
		Context:           objectOrEmpty(input.Context),
	})
	if err != nil {
		return PreparedGuidance{}, fmt.Errorf("record Maya guidance email: %w", err)
	}
	return PreparedGuidance{
		Thread:  thread,
		Message: message,
		ReplyTo: s.replyAddress(token),
		Token:   token,
	}, nil
}

func (s *Service) PrepareReply(ctx context.Context, input ReplyInput) (PreparedReply, error) {
	if s == nil || s.store == nil {
		return PreparedReply{}, errors.New("email thread service is not configured")
	}
	input.InternetMessageID = strings.TrimSpace(input.InternetMessageID)
	input.InReplyTo = strings.TrimSpace(input.InReplyTo)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Content = strings.TrimSpace(input.Content)
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.Thread.ID == uuid.Nil || input.InternetMessageID == "" || input.Subject == "" || input.Content == "" || input.IdempotencyKey == "" {
		return PreparedReply{}, errors.New("email reply requires a thread, message id, subject, content, and idempotency key")
	}
	token := strings.TrimSpace(input.ReplyToken)
	if token == "" {
		var tokenHash []byte
		var err error
		token, tokenHash, err = s.newReplyToken()
		if err != nil {
			return PreparedReply{}, err
		}
		if _, _, err := s.store.CreateEmailReplyTokenAlias(ctx, messaging.EmailReplyTokenInput{
			ThreadID:    input.Thread.ID,
			WorkspaceID: input.Thread.WorkspaceID,
			UserID:      input.Thread.UserID,
			TokenHash:   tokenHash,
			ExpiresAt:   s.now().UTC().Add(defaultReplyTokenLifetime),
		}); err != nil {
			return PreparedReply{}, fmt.Errorf("rotate Maya reply address: %w", err)
		}
	} else if !validOpaqueReplyToken(token) {
		return PreparedReply{}, errors.New("maya reply token is invalid")
	}
	message, _, err := s.store.AppendEmailMessage(ctx, messaging.EmailMessageInput{
		ThreadID:           input.Thread.ID,
		WorkspaceID:        input.Thread.WorkspaceID,
		UserID:             input.Thread.UserID,
		IdempotencyKey:     input.IdempotencyKey,
		Direction:          messaging.EmailMessageDirectionOutbound,
		Role:               messaging.EmailMessageRoleAssistant,
		Kind:               input.Kind,
		InternetMessageID:  input.InternetMessageID,
		InReplyToMessageID: input.InReplyTo,
		Subject:            input.Subject,
		Content:            input.Content,
		Context:            objectOrEmpty(input.Context),
	})
	if err != nil {
		return PreparedReply{}, fmt.Errorf("record Maya reply email: %w", err)
	}
	return PreparedReply{
		Message:    message,
		ReplyTo:    s.replyAddress(token),
		Token:      token,
		References: threadReferences(input.Thread, input.InReplyTo),
	}, nil
}

// NewReplyToken creates and persists a fresh thread-bound reply address token.
// Callers can freeze this token in a durable outbox before recording/sending a
// reply, then pass it to PrepareReply on every retry without alias churn.
func (s *Service) NewReplyToken(ctx context.Context, thread messaging.EmailThreadRecord) (string, error) {
	if s == nil || s.store == nil || thread.ID == uuid.Nil {
		return "", errors.New("email reply token requires a configured thread service and thread")
	}
	token, tokenHash, err := s.newReplyToken()
	if err != nil {
		return "", err
	}
	if _, _, err := s.store.CreateEmailReplyTokenAlias(ctx, messaging.EmailReplyTokenInput{
		ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
		TokenHash: tokenHash, ExpiresAt: s.now().UTC().Add(defaultReplyTokenLifetime),
	}); err != nil {
		return "", fmt.Errorf("rotate Maya reply address: %w", err)
	}
	return token, nil
}

func ThreadedEmail(prepared PreparedReply, recipient, htmlBody, plainText string) mailer.Email {
	return mailer.Email{
		To:            []string{recipient},
		Subject:       prepared.Message.Subject,
		Body:          htmlBody,
		PlainTextBody: plainText,
		IsHTML:        true,
		Sender:        mailer.SenderProfileMaya,
		ReplyTo:       prepared.ReplyTo,
		MessageID:     valueOrEmpty(prepared.Message.InternetMessageID),
		InReplyTo:     valueOrEmpty(prepared.Message.InReplyToMessageID),
		References:    prepared.References,
	}
}

func (s *Service) newReplyToken() (string, []byte, error) {
	raw := make([]byte, defaultTokenBytes)
	if _, err := s.randomBytes(raw); err != nil {
		return "", nil, fmt.Errorf("generate Maya reply token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	digest := sha256.Sum256([]byte(token))
	return token, digest[:], nil
}

func (s *Service) replyAddress(token string) string {
	return "maya+" + token + "@" + s.domain
}

func normalizeGuidanceInput(input GuidanceInput) GuidanceInput {
	input.RecipientEmail = strings.ToLower(strings.TrimSpace(input.RecipientEmail))
	input.ExternalThreadID = strings.TrimSpace(input.ExternalThreadID)
	input.InternetMessageID = strings.TrimSpace(input.InternetMessageID)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Content = strings.TrimSpace(input.Content)
	input.Context = objectOrEmpty(input.Context)
	return input
}

func objectOrEmpty(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`{}`)
	}
	return value
}

func threadReferences(thread messaging.EmailThreadRecord, inReplyTo string) []string {
	values := []string{thread.RootInternetMessageID, inReplyTo}
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func validOpaqueReplyToken(token string) bool {
	if len(token) < 16 || len(token) > 256 {
		return false
	}
	for _, character := range token {
		if (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
			(character >= '0' && character <= '9') || character == '-' || character == '_' {
			continue
		}
		return false
	}
	return true
}
