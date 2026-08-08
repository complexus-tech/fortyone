package messagingrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrNotFound = sql.ErrNoRows

var ErrLeaseBusy = errors.New("messaging lease is busy")

const (
	messagingLeaseDuration    = 2 * time.Minute
	messagingLeaseRetryMargin = 5 * time.Second
	inboundRecoveryBaseDelay  = 10 * time.Minute
	inboundRecoveryMaxShift   = 5
)

// LeaseBusyError indicates that another worker still owns an active processing
// lease. Callers must retry rather than treating the existing row as completed.
type LeaseBusyError struct {
	Resource   string
	RetryAfter time.Duration
}

func (e *LeaseBusyError) Error() string {
	return fmt.Sprintf("%s: %v; retry after %s", e.Resource, ErrLeaseBusy, e.RetryAfter)
}

func (e *LeaseBusyError) Unwrap() error {
	return ErrLeaseBusy
}

// LeaseRetryAfter extracts the safe retry delay from a busy messaging lease.
func LeaseRetryAfter(err error) (time.Duration, bool) {
	var busy *LeaseBusyError
	if !errors.As(err, &busy) {
		return 0, false
	}
	if busy.RetryAfter <= 0 {
		return messagingLeaseDuration + messagingLeaseRetryMargin, true
	}
	return busy.RetryAfter, true
}

func newLeaseBusyError(resource string) *LeaseBusyError {
	return &LeaseBusyError{
		Resource:   resource,
		RetryAfter: messagingLeaseDuration + messagingLeaseRetryMargin,
	}
}

type Repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

type NonceInput struct {
	Provider            string
	Purpose             string
	NonceHash           []byte
	WorkspaceID         uuid.UUID
	UserID              *uuid.UUID
	ExternalWorkspaceID string
	ExternalUserID      string
	Payload             json.RawMessage
	ExpiresAt           time.Time
}

type NonceRecord struct {
	ID                  uuid.UUID       `db:"id"`
	Provider            string          `db:"provider"`
	Purpose             string          `db:"purpose"`
	WorkspaceID         uuid.UUID       `db:"workspace_id"`
	UserID              *uuid.UUID      `db:"user_id"`
	ExternalWorkspaceID *string         `db:"external_workspace_id"`
	ExternalUserID      *string         `db:"external_user_id"`
	Payload             json.RawMessage `db:"payload"`
	ExpiresAt           time.Time       `db:"expires_at"`
	ConsumedAt          *time.Time      `db:"consumed_at"`
}

type NonceConsumeInput struct {
	Provider    string
	Purpose     string
	NonceHash   []byte
	WorkspaceID *uuid.UUID
	UserID      *uuid.UUID
	Now         time.Time
}

func (r *Repository) CreateNonce(ctx context.Context, input NonceInput) error {
	if r == nil || r.db == nil {
		return errors.New("messaging repository is not configured")
	}
	payload := input.Payload
	if len(payload) == 0 {
		payload = json.RawMessage(`{}`)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO messaging_nonces (
			provider, purpose, nonce_hash, workspace_id, user_id,
			external_workspace_id, external_user_id, payload, expires_at
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), $8, $9)
	`, input.Provider, input.Purpose, input.NonceHash, input.WorkspaceID, input.UserID,
		input.ExternalWorkspaceID, input.ExternalUserID, payload, input.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create messaging nonce: %w", err)
	}
	return nil
}

// ConsumeNonce atomically marks a nonce used and returns its bound identity.
func (r *Repository) ConsumeNonce(ctx context.Context, input NonceConsumeInput) (NonceRecord, error) {
	if r == nil || r.db == nil {
		return NonceRecord{}, errors.New("messaging repository is not configured")
	}
	var row NonceRecord
	err := r.db.GetContext(ctx, &row, `
		UPDATE messaging_nonces
		SET consumed_at = $4,
		    user_id = COALESCE(user_id, $6)
		WHERE provider = $1
		  AND purpose = $2
		  AND nonce_hash = $3
		  AND consumed_at IS NULL
		  AND expires_at > $4
		  AND (CAST($5 AS uuid) IS NULL OR workspace_id = $5)
		  AND (CAST($6 AS uuid) IS NULL OR user_id IS NULL OR user_id = $6)
		RETURNING id, provider, purpose, workspace_id, user_id,
		          external_workspace_id, external_user_id, payload, expires_at, consumed_at
	`, input.Provider, input.Purpose, input.NonceHash, input.Now, input.WorkspaceID, input.UserID)
	if err != nil {
		return NonceRecord{}, err
	}
	return row, nil
}

type InboundEventInput struct {
	Provider            string
	WorkspaceID         *uuid.UUID
	InstallGeneration   *uuid.UUID
	ExternalWorkspaceID string
	ExternalEventID     string
	EventType           string
	PayloadEncrypted    string
}

type InboundEventRecord struct {
	ID                  uuid.UUID  `db:"id"`
	WorkspaceID         *uuid.UUID `db:"workspace_id"`
	InstallGeneration   *uuid.UUID `db:"installation_generation"`
	ExternalWorkspaceID string     `db:"external_workspace_id"`
	ExternalEventID     string     `db:"external_event_id"`
	Status              string     `db:"status"`
	AttemptCount        int        `db:"attempt_count"`
	RecoveryGeneration  int        `db:"recovery_generation"`
	ProcessedAt         *time.Time `db:"processed_at"`
	PayloadEncrypted    *string    `db:"payload_encrypted"`
}

// RegisterInboundEvent persists the provider event identity before an HTTP
// acknowledgement. created is false for a retry already present in the inbox.
func (r *Repository) RegisterInboundEvent(ctx context.Context, input InboundEventInput) (record InboundEventRecord, created bool, err error) {
	if r == nil || r.db == nil {
		return InboundEventRecord{}, false, errors.New("messaging repository is not configured")
	}
	err = r.db.GetContext(ctx, &record, `
		INSERT INTO messaging_inbound_events (
			provider, workspace_id, installation_generation, external_workspace_id, external_event_id, event_type, payload_encrypted
		) VALUES ($1, $2, $3, $4, $5, $6, NULLIF($7, ''))
		ON CONFLICT (provider, external_workspace_id, external_event_id) DO NOTHING
		RETURNING id, workspace_id, installation_generation, external_workspace_id, external_event_id, status, attempt_count,
		          recovery_generation, processed_at, payload_encrypted
	`, input.Provider, input.WorkspaceID, input.InstallGeneration, input.ExternalWorkspaceID, input.ExternalEventID, input.EventType, input.PayloadEncrypted)
	if err == nil {
		return record, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return InboundEventRecord{}, false, fmt.Errorf("register messaging inbound event: %w", err)
	}
	err = r.db.GetContext(ctx, &record, `
		UPDATE messaging_inbound_events mie
		SET payload_encrypted = COALESCE(mie.payload_encrypted, NULLIF($4, ''))
		WHERE mie.provider = $1 AND mie.external_workspace_id = $2 AND mie.external_event_id = $3
		RETURNING mie.id, mie.workspace_id, mie.installation_generation, mie.external_workspace_id, mie.external_event_id, mie.status,
		          mie.attempt_count, mie.recovery_generation, mie.processed_at, mie.payload_encrypted
	`, input.Provider, input.ExternalWorkspaceID, input.ExternalEventID, input.PayloadEncrypted)
	if err != nil {
		return InboundEventRecord{}, false, fmt.Errorf("read messaging inbound event: %w", err)
	}
	return record, false, nil
}

// ClaimRecoverableInboundEvents assigns a durable queue generation to inbox
// work that is absent, failed, or held by an expired processing lease.
func (r *Repository) ClaimRecoverableInboundEvents(ctx context.Context, provider string, limit int) ([]InboundEventRecord, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("messaging repository is not configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows := make([]InboundEventRecord, 0)
	err := r.db.SelectContext(ctx, &rows, `
		WITH candidates AS (
			SELECT id
			FROM messaging_inbound_events
			WHERE provider = $1
			  AND payload_encrypted IS NOT NULL
			  AND attempt_count < 20
			  AND (
				(status = 'pending' AND updated_at < NOW() - INTERVAL '30 seconds')
				OR (status = 'failed' AND updated_at < NOW() - INTERVAL '5 minutes')
				OR (status = 'processing' AND updated_at < NOW() - ($3 * INTERVAL '1 second'))
			  )
			  AND (
				recovery_enqueued_at IS NULL
				OR recovery_enqueued_at < NOW() - (
					($4 * POWER(2, LEAST(recovery_generation, $5))) * INTERVAL '1 second'
				)
			  )
			ORDER BY COALESCE(recovery_enqueued_at, received_at) ASC, received_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT $2
		)
		UPDATE messaging_inbound_events mie
		SET recovery_generation = mie.recovery_generation + 1,
		    recovery_enqueued_at = NOW()
		FROM candidates
		WHERE mie.id = candidates.id
		RETURNING mie.id, mie.workspace_id, mie.installation_generation, mie.external_workspace_id, mie.external_event_id, mie.status,
		          mie.attempt_count, mie.recovery_generation, mie.processed_at,
		          mie.payload_encrypted
	`, provider, limit, int64(messagingLeaseDuration/time.Second),
		int64(inboundRecoveryBaseDelay/time.Second), inboundRecoveryMaxShift)
	if err != nil {
		return nil, fmt.Errorf("claim recoverable messaging inbound events: %w", err)
	}
	return rows, nil
}

// MarkInboundEventQueued records a successful database-to-queue handoff. The
// recovery scanner uses this timestamp with exponential backoff so it cannot
// amplify a queued event while integrations workers are unavailable.
func (r *Repository) MarkInboundEventQueued(ctx context.Context, id uuid.UUID) error {
	if r == nil || r.db == nil {
		return errors.New("messaging repository is not configured")
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE messaging_inbound_events
		SET recovery_enqueued_at = NOW(), updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'failed', 'processing')
	`, id)
	if err != nil {
		return fmt.Errorf("mark messaging inbound event queued: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read queued messaging inbound event result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("mark messaging inbound event queued: %w", ErrNotFound)
	}
	return nil
}

// ReleaseInboundEventRecovery makes a failed queue handoff eligible for the
// next recovery scan without disturbing a newer recovery generation.
func (r *Repository) ReleaseInboundEventRecovery(ctx context.Context, id uuid.UUID, generation int) error {
	if r == nil || r.db == nil {
		return errors.New("messaging repository is not configured")
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE messaging_inbound_events
		SET recovery_enqueued_at = NULL
		WHERE id = $1 AND recovery_generation = $2
	`, id, generation)
	if err != nil {
		return fmt.Errorf("release messaging inbound event recovery: %w", err)
	}
	return nil
}

// GetInboundEvent returns the canonical encrypted inbox payload used by queue
// workers. Provider payloads never need to be copied into the queue backend.
func (r *Repository) GetInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (InboundEventRecord, error) {
	if r == nil || r.db == nil {
		return InboundEventRecord{}, errors.New("messaging repository is not configured")
	}
	var record InboundEventRecord
	err := r.db.GetContext(ctx, &record, `
		SELECT id, workspace_id, installation_generation, external_workspace_id, external_event_id, status, attempt_count,
		       recovery_generation, processed_at, payload_encrypted
		FROM messaging_inbound_events
		WHERE provider = $1 AND external_workspace_id = $2 AND external_event_id = $3
	`, provider, externalWorkspaceID, externalEventID)
	if err != nil {
		return InboundEventRecord{}, fmt.Errorf("get messaging inbound event: %w", err)
	}
	return record, nil
}

// StartInboundEvent claims an event for processing. claimed is false with a nil
// error only when the event already reached a terminal state. An active claim
// returns LeaseBusyError so queue workers retain and retry their task.
func (r *Repository) StartInboundEvent(ctx context.Context, provider, externalWorkspaceID, externalEventID string) (InboundEventRecord, bool, error) {
	if r == nil || r.db == nil {
		return InboundEventRecord{}, false, errors.New("messaging repository is not configured")
	}
	var row InboundEventRecord
	err := r.db.GetContext(ctx, &row, `
		UPDATE messaging_inbound_events
		SET status = 'processing', attempt_count = attempt_count + 1, last_error = NULL,
		    recovery_enqueued_at = NULL, updated_at = NOW()
		WHERE provider = $1
		  AND external_workspace_id = $2
		  AND external_event_id = $3
		  AND (
			status IN ('pending', 'failed')
			OR (status = 'processing' AND updated_at < NOW() - ($4 * INTERVAL '1 second'))
		  )
		RETURNING id, workspace_id, installation_generation, external_workspace_id, external_event_id, status, attempt_count, processed_at
	`, provider, externalWorkspaceID, externalEventID, int64(messagingLeaseDuration/time.Second))
	if err == nil {
		return row, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return InboundEventRecord{}, false, fmt.Errorf("start messaging inbound event: %w", err)
	}
	err = r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, installation_generation, external_workspace_id, external_event_id, status, attempt_count, processed_at
		FROM messaging_inbound_events
		WHERE provider = $1 AND external_workspace_id = $2 AND external_event_id = $3
	`, provider, externalWorkspaceID, externalEventID)
	if err != nil {
		return InboundEventRecord{}, false, fmt.Errorf("read completed messaging inbound event: %w", err)
	}
	switch row.Status {
	case "completed", "ignored", "cancelled":
		return row, false, nil
	case "processing":
		return row, false, newLeaseBusyError("messaging inbound event")
	default:
		return row, false, fmt.Errorf("messaging inbound event is unexpectedly unclaimed in status %q", row.Status)
	}
}

func (r *Repository) CompleteInboundEvent(ctx context.Context, id uuid.UUID, status, message string) error {
	if status != "completed" && status != "ignored" && status != "failed" {
		return fmt.Errorf("invalid messaging event status %q", status)
	}
	processedAt := any(nil)
	if status == "completed" || status == "ignored" {
		processedAt = time.Now().UTC()
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE messaging_inbound_events
		SET status = $2,
		    last_error = NULLIF($3, ''),
		    processed_at = COALESCE($4, processed_at),
		    updated_at = NOW()
		WHERE id = $1 AND status = 'processing'
	`, id, status, message, processedAt)
	if err != nil {
		return fmt.Errorf("complete messaging inbound event: %w", err)
	}
	return nil
}

type ConversationInput struct {
	Provider            string
	WorkspaceID         uuid.UUID
	ExternalWorkspaceID string
	ExternalChannelID   string
	ExternalThreadID    string
	UserID              uuid.UUID
}

type MessageRecord struct {
	ExternalMessageID *string   `db:"external_message_id"`
	Role              string    `db:"role"`
	Content           string    `db:"content"`
	CreatedAt         time.Time `db:"created_at"`
}

func (r *Repository) UpsertConversation(ctx context.Context, input ConversationInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.GetContext(ctx, &id, `
		INSERT INTO messaging_conversations (
			provider, workspace_id, external_workspace_id, external_channel_id, external_thread_id, user_id
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (
			provider, workspace_id, external_workspace_id, external_channel_id, external_thread_id, user_id
		) DO UPDATE SET updated_at = NOW()
		RETURNING id
	`, input.Provider, input.WorkspaceID, input.ExternalWorkspaceID, input.ExternalChannelID, input.ExternalThreadID, input.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("upsert messaging conversation: %w", err)
	}
	return id, nil
}

func (r *Repository) AppendMessage(ctx context.Context, conversationID uuid.UUID, externalMessageID, role, content string) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO messaging_messages (conversation_id, external_message_id, role, content)
		VALUES ($1, NULLIF($2, ''), $3, $4)
		ON CONFLICT (conversation_id, external_message_id, role)
		WHERE external_message_id IS NOT NULL
		DO NOTHING
	`, conversationID, externalMessageID, role, content)
	if err != nil {
		return fmt.Errorf("append messaging message: %w", err)
	}
	return nil
}

func (r *Repository) ListRecentMessages(ctx context.Context, conversationID uuid.UUID, limit int) ([]MessageRecord, error) {
	if limit <= 0 || limit > 50 {
		limit = 20
	}
	rows := make([]MessageRecord, 0, limit)
	err := r.db.SelectContext(ctx, &rows, `
		SELECT external_message_id, role, content, created_at
		FROM (
			SELECT external_message_id, role, content, created_at
			FROM messaging_messages
			WHERE conversation_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		) recent
		ORDER BY created_at ASC
	`, conversationID, limit)
	if err != nil {
		return nil, fmt.Errorf("list messaging messages: %w", err)
	}
	return rows, nil
}

type OutboundDeliveryInput struct {
	Provider                string
	WorkspaceID             uuid.UUID
	UserID                  *uuid.UUID
	InstallGeneration       *uuid.UUID
	ExternalWorkspaceID     string
	ExternalRecipientUserID string
	InboundEventID          *uuid.UUID
	IdempotencyKey          string
	ExternalChannelID       string
	ExternalThreadID        string
	Content                 string
	Purpose                 string
	ExpiresAt               *time.Time
}

type OutboundDeliveryRecord struct {
	ID                      uuid.UUID  `db:"id"`
	WorkspaceID             uuid.UUID  `db:"workspace_id"`
	UserID                  *uuid.UUID `db:"user_id"`
	InstallGeneration       *uuid.UUID `db:"installation_generation"`
	ExternalWorkspaceID     string     `db:"external_workspace_id"`
	ExternalRecipientUserID *string    `db:"external_recipient_user_id"`
	InboundEventID          *uuid.UUID `db:"inbound_event_id"`
	IdempotencyKey          string     `db:"idempotency_key"`
	ExternalChannelID       string     `db:"external_channel_id"`
	ExternalThreadID        *string    `db:"external_thread_id"`
	ExternalMessageID       *string    `db:"external_message_id"`
	Content                 *string    `db:"content"`
	Status                  string     `db:"status"`
	AttemptCount            int        `db:"attempt_count"`
	Purpose                 string     `db:"purpose"`
	ExpiresAt               *time.Time `db:"expires_at"`
}

// ListRecoverableOutboundDeliveries returns persisted messages that can be
// retried independently of their original HTTP request or worker process.
func (r *Repository) ListRecoverableOutboundDeliveries(ctx context.Context, provider string, limit int) ([]OutboundDeliveryRecord, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("messaging repository is not configured")
	}
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	records := make([]OutboundDeliveryRecord, 0)
	err := r.db.SelectContext(ctx, &records, `
			SELECT id, workspace_id, user_id, installation_generation, external_workspace_id, external_recipient_user_id, inbound_event_id, idempotency_key,
			       external_channel_id, external_thread_id, external_message_id,
			       content, purpose, expires_at, status, attempt_count
		FROM messaging_outbound_deliveries
		WHERE provider = $1
		  AND content IS NOT NULL
		  AND attempt_count < 20
		  AND (
			(status IN ('pending', 'failed') AND updated_at < NOW() - INTERVAL '5 minutes')
			OR (status = 'delivering' AND updated_at < NOW() - ($3 * INTERVAL '1 second'))
		  )
		ORDER BY created_at ASC
		LIMIT $2
	`, provider, limit, int64(messagingLeaseDuration/time.Second))
	if err != nil {
		return nil, fmt.Errorf("list recoverable messaging outbound deliveries: %w", err)
	}
	return records, nil
}

// StartOutboundDelivery claims an idempotent delivery. claimed is false with a
// nil error only when the delivery already succeeded. An active claim returns
// LeaseBusyError so queue workers retain and retry their task.
func (r *Repository) StartOutboundDelivery(ctx context.Context, input OutboundDeliveryInput) (OutboundDeliveryRecord, bool, error) {
	externalWorkspaceID := strings.TrimSpace(input.ExternalWorkspaceID)
	if externalWorkspaceID == "" {
		return OutboundDeliveryRecord{}, false, errors.New("messaging outbound delivery external workspace id is required")
	}
	purpose := strings.TrimSpace(input.Purpose)
	if purpose == "" {
		purpose = "provider_message"
	}
	var row OutboundDeliveryRecord
	err := r.db.GetContext(ctx, &row, `
		INSERT INTO messaging_outbound_deliveries (
			provider, workspace_id, user_id, installation_generation, external_workspace_id, external_recipient_user_id,
			inbound_event_id, idempotency_key, external_channel_id, external_thread_id,
			content, purpose, expires_at, status, attempt_count
		) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $9, NULLIF($10, ''),
		          NULLIF($11, ''), $12, $13, 'delivering', 1)
		ON CONFLICT (provider, workspace_id, idempotency_key) DO UPDATE SET
			status = 'delivering',
			attempt_count = messaging_outbound_deliveries.attempt_count + 1,
			content = COALESCE(messaging_outbound_deliveries.content, EXCLUDED.content),
			user_id = COALESCE(messaging_outbound_deliveries.user_id, EXCLUDED.user_id),
			installation_generation = COALESCE(messaging_outbound_deliveries.installation_generation, EXCLUDED.installation_generation),
			external_recipient_user_id = COALESCE(messaging_outbound_deliveries.external_recipient_user_id, EXCLUDED.external_recipient_user_id),
			expires_at = COALESCE(messaging_outbound_deliveries.expires_at, EXCLUDED.expires_at),
			last_error = NULL,
			updated_at = NOW()
		WHERE (
			messaging_outbound_deliveries.status IN ('pending', 'failed')
			OR (
			messaging_outbound_deliveries.status = 'delivering'
				AND messaging_outbound_deliveries.updated_at < NOW() - ($14 * INTERVAL '1 second')
			)
		)
		AND messaging_outbound_deliveries.external_workspace_id = EXCLUDED.external_workspace_id
		AND messaging_outbound_deliveries.purpose = EXCLUDED.purpose
		AND messaging_outbound_deliveries.user_id IS NOT DISTINCT FROM EXCLUDED.user_id
		AND messaging_outbound_deliveries.installation_generation IS NOT DISTINCT FROM EXCLUDED.installation_generation
		AND messaging_outbound_deliveries.external_recipient_user_id IS NOT DISTINCT FROM EXCLUDED.external_recipient_user_id
		RETURNING id, workspace_id, user_id, installation_generation, external_workspace_id, external_recipient_user_id, inbound_event_id, idempotency_key,
		          external_channel_id, external_thread_id, external_message_id, content, purpose, expires_at, status, attempt_count
	`, input.Provider, input.WorkspaceID, input.UserID, input.InstallGeneration, externalWorkspaceID, input.ExternalRecipientUserID,
		input.InboundEventID, input.IdempotencyKey, input.ExternalChannelID, input.ExternalThreadID,
		input.Content, purpose, input.ExpiresAt, int64(messagingLeaseDuration/time.Second))
	if err == nil {
		return row, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return OutboundDeliveryRecord{}, false, fmt.Errorf("start messaging outbound delivery: %w", err)
	}
	err = r.db.GetContext(ctx, &row, `
		SELECT id, workspace_id, user_id, installation_generation, external_workspace_id, external_recipient_user_id, inbound_event_id, idempotency_key,
		       external_channel_id, external_thread_id, external_message_id, content, purpose, expires_at, status, attempt_count
		FROM messaging_outbound_deliveries
		WHERE provider = $1 AND workspace_id = $2 AND idempotency_key = $3
	`, input.Provider, input.WorkspaceID, input.IdempotencyKey)
	if err != nil {
		return OutboundDeliveryRecord{}, false, fmt.Errorf("read messaging outbound delivery: %w", err)
	}
	if strings.TrimSpace(row.ExternalWorkspaceID) != externalWorkspaceID {
		return row, false, fmt.Errorf(
			"messaging outbound delivery external workspace mismatch: persisted %q, requested %q",
			row.ExternalWorkspaceID,
			externalWorkspaceID,
		)
	}
	if row.Purpose != purpose || !equalOptionalUUID(row.UserID, input.UserID) || !equalOptionalUUID(row.InstallGeneration, input.InstallGeneration) || strings.TrimSpace(valueOrEmptyString(row.ExternalRecipientUserID)) != strings.TrimSpace(input.ExternalRecipientUserID) {
		return row, false, errors.New("messaging outbound delivery actor or installation binding mismatch")
	}
	switch row.Status {
	case "delivered", "cancelled":
		return row, false, nil
	case "delivering":
		return row, false, newLeaseBusyError("messaging outbound delivery")
	default:
		return row, false, fmt.Errorf("messaging outbound delivery is unexpectedly unclaimed in status %q", row.Status)
	}
}

func equalOptionalUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func valueOrEmptyString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func (r *Repository) SetOutboundDeliveryContent(ctx context.Context, id uuid.UUID, content string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET content = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'delivering'
	`, id, content)
	if err != nil {
		return fmt.Errorf("set messaging outbound delivery content: %w", err)
	}
	if err := requireAffectedRow(result, "set messaging outbound delivery content"); err != nil {
		return err
	}
	return nil
}

func (r *Repository) CompleteOutboundDelivery(ctx context.Context, id uuid.UUID, externalMessageID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'delivered', external_message_id = NULLIF($2, ''),
		    delivered_at = NOW(), last_error = NULL, updated_at = NOW()
		WHERE id = $1 AND status = 'delivering'
	`, id, externalMessageID)
	if err != nil {
		return fmt.Errorf("complete messaging outbound delivery: %w", err)
	}
	if err := requireAffectedRow(result, "complete messaging outbound delivery"); err != nil {
		return err
	}
	return nil
}

func requireAffectedRow(result sql.Result, operation string) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s result: %w", operation, err)
	}
	if affected == 0 {
		return fmt.Errorf("%s: %w", operation, ErrNotFound)
	}
	return nil
}

func (r *Repository) FailOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'failed', last_error = $2, updated_at = NOW()
		WHERE id = $1 AND status = 'delivering'
	`, id, message)
	if err != nil {
		return fmt.Errorf("fail messaging outbound delivery: %w", err)
	}
	return nil
}

func (r *Repository) CancelOutboundDelivery(ctx context.Context, id uuid.UUID, message string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE messaging_outbound_deliveries
		SET status = 'cancelled', content = NULL, last_error = $2, updated_at = NOW()
		WHERE id = $1 AND status IN ('pending', 'delivering', 'failed')
	`, id, message)
	if err != nil {
		return fmt.Errorf("cancel messaging outbound delivery: %w", err)
	}
	return nil
}
