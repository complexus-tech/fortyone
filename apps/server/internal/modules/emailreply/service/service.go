package emailreply

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

var (
	ErrReplyNotAuthorized = errors.New("email reply is not authorized")
	ErrWebhookRetry       = errors.New("email reply webhook must be retried")
)

// Store is intentionally narrow so the email thread repository can adapt to
// ingress without coupling this package to thread persistence details.
type Store interface {
	FindEmailThreadByReplyToken(ctx context.Context, input messaging.EmailReplyTokenLookup) (messaging.EmailThreadLookup, error)
	RegisterInboundEvent(ctx context.Context, input messagingrepository.InboundEventInput) (messagingrepository.InboundEventRecord, bool, error)
	MarkInboundEventQueued(ctx context.Context, id uuid.UUID) error
	ClaimRecoverableInboundEvents(ctx context.Context, provider string, limit int) ([]messagingrepository.InboundEventRecord, error)
	ReleaseInboundEventRecovery(ctx context.Context, id uuid.UUID, generation int) error
}

type Queue interface {
	EnqueueBrevoEmailReply(ctx context.Context, payload tasks.BrevoEmailReplyPayload) error
}

type Service struct {
	secret string
	store  Store
	queue  Queue
	codec  *PayloadCodec
	now    func() time.Time
}

type IngestResult struct {
	Accepted   int
	Duplicates int
	Ignored    int
}

type storedInboundEmail struct {
	ThreadID    uuid.UUID       `json:"threadId"`
	WorkspaceID uuid.UUID       `json:"workspaceId"`
	UserID      uuid.UUID       `json:"userId"`
	Email       json.RawMessage `json:"email"`
}

// StoredInboundEmail is the authenticated, encrypted work item recovered by
// the background processor. It deliberately exposes only the thread binding
// and parsed provider message after the payload codec has authenticated it.
type StoredInboundEmail struct {
	ThreadID    uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	Email       InboundEmail
}

func New(secret string, store Store, queue Queue) (*Service, error) {
	if store == nil {
		return nil, errors.New("email reply store is required")
	}
	if queue == nil {
		return nil, errors.New("email reply queue is required")
	}
	codec, err := NewPayloadCodec(secret)
	if err != nil {
		return nil, err
	}
	return &Service{
		secret: strings.TrimSpace(secret),
		store:  store,
		queue:  queue,
		codec:  codec,
		now:    time.Now,
	}, nil
}

// OpenStoredInboundEmail authenticates and decodes one encrypted inbox value.
// Provider JSON is never copied into the queue payload.
func (s *Service) OpenStoredInboundEmail(sealed string) (StoredInboundEmail, error) {
	if s == nil || s.codec == nil {
		return StoredInboundEmail{}, errors.New("email reply service is not configured")
	}
	plaintext, err := s.codec.Open(sealed)
	if err != nil {
		return StoredInboundEmail{}, err
	}
	var stored storedInboundEmail
	if err := json.Unmarshal(plaintext, &stored); err != nil {
		return StoredInboundEmail{}, fmt.Errorf("decode stored Brevo email reply: %w", err)
	}
	email, err := decodeInboundEmail(stored.Email)
	if err != nil {
		return StoredInboundEmail{}, err
	}
	return StoredInboundEmail{
		ThreadID:    stored.ThreadID,
		WorkspaceID: stored.WorkspaceID,
		UserID:      stored.UserID,
		Email:       email,
	}, nil
}

// SealProcessorState encrypts retry-critical email delivery state before it is
// stored in the generic messaging outbox. This includes the raw plus-address
// token; reply-token tables continue to persist hashes only.
func (s *Service) SealProcessorState(payload []byte) (string, error) {
	if s == nil || s.codec == nil {
		return "", errors.New("email reply service is not configured")
	}
	return s.codec.Seal(payload)
}

// OpenProcessorState authenticates retry-critical state frozen by
// SealProcessorState. Plaintext provider payloads are never accepted.
func (s *Service) OpenProcessorState(sealed string) ([]byte, error) {
	if s == nil || s.codec == nil {
		return nil, errors.New("email reply service is not configured")
	}
	return s.codec.Open(sealed)
}

// ReplyTokenHash resolves the opaque plus-address token without retaining the
// raw value in durable processor state.
func ReplyTokenHash(email InboundEmail) ([]byte, error) {
	token, err := extractReplyToken(email)
	if err != nil {
		return nil, err
	}
	digest := sha256.Sum256([]byte(token))
	return digest[:], nil
}

func (s *Service) VerifyWebhookToken(provided string) bool {
	return s != nil && VerifyWebhookToken(s.secret, provided)
}

// Ingest persists authorized Brevo items in the durable messaging inbox and
// hands them to the integrations queue before returning success. Malformed or
// unauthorized individual items are permanent rejects and cannot poison later
// items in the same provider batch. A transient partial failure is safe: Brevo
// can retry the batch and the inbox/task identities suppress duplicates.
func (s *Service) Ingest(ctx context.Context, rawBody []byte) (IngestResult, error) {
	if s == nil || s.store == nil || s.queue == nil || s.codec == nil {
		return IngestResult{}, fmt.Errorf("%w: email reply service is not configured", ErrWebhookRetry)
	}
	payload, err := decodeInboundWebhook(rawBody)
	if err != nil {
		return IngestResult{}, err
	}

	result := IngestResult{}
	for index, rawItem := range payload.Items {
		created, err := s.ingestItem(ctx, rawItem)
		if err != nil {
			if isPermanentItemError(err) {
				result.Ignored++
				continue
			}
			return result, fmt.Errorf("ingest Brevo inbound email item %d: %w", index, err)
		}
		if created {
			result.Accepted++
		} else {
			result.Duplicates++
		}
	}
	return result, nil
}

// RecoverPendingEvents republishes durable inbox entries whose original queue
// handoff failed or whose worker lease expired. Each recovery claim advances a
// generation that becomes part of the task ID, so a retained Asynq task cannot
// suppress a later recovery attempt.
func (s *Service) RecoverPendingEvents(ctx context.Context) (int, error) {
	if s == nil || s.store == nil || s.queue == nil {
		return 0, errors.New("email reply recovery is not configured")
	}
	records, err := s.store.ClaimRecoverableInboundEvents(ctx, Provider, 500)
	if err != nil {
		return 0, fmt.Errorf("claim recoverable Brevo email replies: %w", err)
	}

	recovered := 0
	var recoveryErrors []error
	for _, record := range records {
		err := s.queue.EnqueueBrevoEmailReply(ctx, tasks.BrevoEmailReplyPayload{
			ExternalWorkspaceID: record.ExternalWorkspaceID,
			EventID:             record.ExternalEventID,
			RecoveryAttempt:     record.RecoveryGeneration,
		})
		if err == nil {
			recovered++
			continue
		}

		recoveryErrors = append(recoveryErrors, fmt.Errorf(
			"re-enqueue Brevo email reply %s: %w",
			record.ExternalEventID,
			err,
		))
		if releaseErr := s.store.ReleaseInboundEventRecovery(ctx, record.ID, record.RecoveryGeneration); releaseErr != nil {
			recoveryErrors = append(recoveryErrors, fmt.Errorf(
				"release Brevo email reply %s recovery claim: %w",
				record.ExternalEventID,
				releaseErr,
			))
		}
	}
	return recovered, errors.Join(recoveryErrors...)
}

func (s *Service) ingestItem(ctx context.Context, rawItem json.RawMessage) (bool, error) {
	email, err := decodeInboundEmail(rawItem)
	if err != nil {
		return false, err
	}
	token, err := extractReplyToken(email)
	if err != nil {
		return false, ErrReplyNotAuthorized
	}
	senderEmail, err := normalizedEmailAddress(email.From.Address)
	if err != nil {
		return false, ErrReplyNotAuthorized
	}
	tokenHash := sha256.Sum256([]byte(token))
	lookup, err := s.store.FindEmailThreadByReplyToken(ctx, messaging.EmailReplyTokenLookup{
		Provider:  Provider,
		TokenHash: tokenHash[:],
		Now:       s.now().UTC(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) || errors.Is(err, messaging.ErrInvalidEmailReplyToken) || errors.Is(err, ErrReplyNotAuthorized) {
			return false, ErrReplyNotAuthorized
		}
		return false, fmt.Errorf("%w: resolve email reply thread: %v", ErrWebhookRetry, err)
	}
	thread := lookup.Thread
	if !validReplyThread(thread, senderEmail) {
		return false, ErrReplyNotAuthorized
	}

	canonicalEmail, err := processorEmailPayload(email)
	if err != nil {
		return false, fmt.Errorf("%w: minimize durable email reply payload: %v", ErrWebhookRetry, err)
	}
	storedPayload, err := json.Marshal(storedInboundEmail{
		ThreadID:    thread.ID,
		WorkspaceID: thread.WorkspaceID,
		UserID:      thread.UserID,
		Email:       canonicalEmail,
	})
	if err != nil {
		return false, fmt.Errorf("%w: encode durable email reply payload: %v", ErrWebhookRetry, err)
	}
	encryptedPayload, err := s.codec.Seal(storedPayload)
	if err != nil {
		return false, fmt.Errorf("%w: %v", ErrWebhookRetry, err)
	}

	eventID := externalEventID(email, rawItem)
	workspaceID := thread.WorkspaceID
	eventScope := workspaceID.String() + ":" + thread.ID.String()
	receipt, created, err := s.store.RegisterInboundEvent(ctx, messagingrepository.InboundEventInput{
		Provider:            Provider,
		WorkspaceID:         &workspaceID,
		ExternalWorkspaceID: eventScope,
		ExternalEventID:     eventID,
		EventType:           InboundProcessedEvent,
		PayloadEncrypted:    encryptedPayload,
	})
	if err != nil {
		return false, fmt.Errorf("%w: persist email reply receipt: %v", ErrWebhookRetry, err)
	}
	if !created && isTerminalInboundStatus(receipt.Status) {
		return false, nil
	}
	if err := s.queue.EnqueueBrevoEmailReply(ctx, tasks.BrevoEmailReplyPayload{
		ExternalWorkspaceID: eventScope,
		EventID:             eventID,
	}); err != nil {
		return false, fmt.Errorf("%w: enqueue email reply: %v", ErrWebhookRetry, err)
	}
	if err := s.store.MarkInboundEventQueued(ctx, receipt.ID); err != nil {
		return false, fmt.Errorf("%w: record email reply queue handoff: %v", ErrWebhookRetry, err)
	}
	return created, nil
}

// processorEmailPayload keeps only fields needed to authenticate, thread, and
// interpret the current reply. Raw HTML, parsed signatures, CC lists, and
// attachment download capabilities never enter durable conversation storage.
func processorEmailPayload(email InboundEmail) (json.RawMessage, error) {
	email.RawHTMLBody = nil
	email.ExtractedMarkdownSignature = nil
	email.Cc = nil
	email.ReplyTo = nil
	email.Attachments = nil
	if strings.TrimSpace(email.ExtractedMarkdownMessage) != "" {
		email.RawTextBody = nil
	}
	encoded, err := json.Marshal(email)
	if err != nil {
		return nil, err
	}
	return encoded, nil
}

func validReplyThread(thread messaging.EmailThreadRecord, senderEmail string) bool {
	if thread.ID == uuid.Nil || thread.WorkspaceID == uuid.Nil || thread.UserID == uuid.Nil {
		return false
	}
	expected, err := normalizedEmailAddress(thread.RecipientEmail)
	if err != nil {
		return false
	}
	return strings.EqualFold(expected, senderEmail)
}

func isTerminalInboundStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "completed", "ignored", "cancelled":
		return true
	default:
		return false
	}
}

func isPermanentItemError(err error) bool {
	return errors.Is(err, ErrInvalidPayload) ||
		errors.Is(err, ErrInvalidReplyToken) ||
		errors.Is(err, ErrReplyNotAuthorized)
}
