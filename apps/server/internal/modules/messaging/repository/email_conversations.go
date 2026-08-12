package messagingrepository

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	defaultEmailMessagePageSize = 50
	maximumEmailMessagePageSize = 100
	maximumEmailErrorRunes      = 2_000
)

// HasEarlierInboundEvent enforces provider arrival order after a worker has
// acquired the per-thread lease. It prevents a later reply from superseding or
// confirming state before an earlier queued reply for the same thread scope.
func (r *Repository) HasEarlierInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID string,
	currentID uuid.UUID,
) (bool, error) {
	if r == nil || r.db == nil {
		return false, errors.New("messaging repository is not configured")
	}
	provider = strings.TrimSpace(provider)
	externalWorkspaceID = strings.TrimSpace(externalWorkspaceID)
	if provider == "" || externalWorkspaceID == "" || currentID == uuid.Nil {
		return false, messaging.ErrInvalidEmailConversation
	}
	var earlier bool
	err := r.db.GetContext(ctx, &earlier, `
		SELECT EXISTS (
			SELECT 1
			FROM messaging_inbound_events candidate
			INNER JOIN messaging_inbound_events current ON current.id = $3
			WHERE candidate.provider = $1
			  AND candidate.external_workspace_id = $2
			  AND candidate.id <> current.id
			  AND (candidate.received_at, candidate.id) < (current.received_at, current.id)
			  AND candidate.status IN ('pending', 'processing', 'failed')
			  AND candidate.attempt_count < 20
		)
	`, provider, externalWorkspaceID, currentID)
	if err != nil {
		return false, fmt.Errorf("check earlier email reply event: %w", err)
	}
	return earlier, nil
}

// AcquireEmailThreadProcessingLease obtains a session advisory lock for one
// actor-bound email thread. It serializes independent inbound events across
// worker processes until the returned release function is called.
func (r *Repository) AcquireEmailThreadProcessingLease(
	ctx context.Context,
	lease messaging.EmailThreadProcessingLease,
) (func() error, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("messaging repository is not configured")
	}
	if lease.ThreadID == uuid.Nil || lease.WorkspaceID == uuid.Nil || lease.UserID == uuid.Nil {
		return nil, messaging.ErrInvalidEmailConversation
	}
	connection, err := r.db.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("reserve email thread lock connection: %w", err)
	}
	lockKey := lease.WorkspaceID.String() + ":" + lease.ThreadID.String() + ":" + lease.UserID.String()
	var locked bool
	err = connection.GetContext(ctx, &locked, `SELECT pg_try_advisory_lock(hashtextextended($1, 0))`, lockKey)
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("acquire email thread lock: %w", err)
	}
	if !locked {
		_ = connection.Close()
		return nil, newLeaseBusyError("messaging email thread")
	}
	return func() error {
		defer connection.Close()
		var released bool
		if err := connection.GetContext(context.WithoutCancel(ctx), &released, `SELECT pg_advisory_unlock(hashtextextended($1, 0))`, lockKey); err != nil {
			return fmt.Errorf("release email thread lock: %w", err)
		}
		if !released {
			return errors.New("email thread lock was not held by its reserved connection")
		}
		return nil
	}, nil
}

// CreateEmailThread atomically creates an actor-bound email thread and its
// first reply-token alias. Replaying the same external thread identity returns
// the existing thread only when its actor and recipient bindings agree.
func (r *Repository) CreateEmailThread(
	ctx context.Context,
	input messaging.EmailThreadInput,
) (messaging.EmailThreadRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailThreadRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailThreadInput(input)
	if err := validateEmailThreadInput(input); err != nil {
		return messaging.EmailThreadRecord{}, false, err
	}

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("begin email thread creation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	thread, created, err := insertOrReadEmailThread(ctx, tx, input)
	if err != nil {
		return messaging.EmailThreadRecord{}, false, err
	}
	if thread.UserID != input.UserID || !strings.EqualFold(thread.RecipientEmail, input.RecipientEmail) ||
		thread.RootInternetMessageID != input.RootInternetMessageID || !jsonObjectsEqual(thread.Context, input.Context) {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf(
			"%w: external thread is bound to different immutable context",
			messaging.ErrInvalidEmailConversation,
		)
	}
	if _, _, err := createEmailReplyTokenAlias(ctx, tx, messaging.EmailReplyTokenInput{
		ThreadID:    thread.ID,
		WorkspaceID: thread.WorkspaceID,
		UserID:      thread.UserID,
		TokenHash:   input.ReplyTokenHash,
		ExpiresAt:   input.ReplyTokenExpiresAt,
	}); err != nil {
		return messaging.EmailThreadRecord{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("commit email thread creation: %w", err)
	}
	return thread, created, nil
}

func (r *Repository) CreateEmailReplyTokenAlias(
	ctx context.Context,
	input messaging.EmailReplyTokenInput,
) (messaging.EmailReplyTokenRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailReplyTokenRecord{}, false, errors.New("messaging repository is not configured")
	}
	if err := validateEmailReplyTokenInput(input); err != nil {
		return messaging.EmailReplyTokenRecord{}, false, err
	}
	return createEmailReplyTokenAlias(ctx, r.db, input)
}

func createEmailReplyTokenAlias(
	ctx context.Context,
	db sqlx.ExtContext,
	input messaging.EmailReplyTokenInput,
) (messaging.EmailReplyTokenRecord, bool, error) {
	if err := validateEmailReplyTokenInput(input); err != nil {
		return messaging.EmailReplyTokenRecord{}, false, err
	}
	var record messaging.EmailReplyTokenRecord
	err := sqlx.GetContext(ctx, db, &record, `
		INSERT INTO messaging_email_reply_tokens (
			thread_id, workspace_id, user_id, token_hash, expires_at
		) VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (token_hash) DO NOTHING
		RETURNING id, thread_id, expires_at, revoked_at, created_at
	`, input.ThreadID, input.WorkspaceID, input.UserID, input.TokenHash, input.ExpiresAt.UTC())
	if err == nil {
		return record, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf("create email reply token alias: %w", err)
	}
	err = sqlx.GetContext(ctx, db, &record, `
		SELECT id, thread_id, expires_at, revoked_at, created_at
		FROM messaging_email_reply_tokens
		WHERE token_hash = $1
	`, input.TokenHash)
	if err != nil {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf("read email reply token alias: %w", err)
	}
	if record.ThreadID != input.ThreadID || !record.ExpiresAt.Equal(input.ExpiresAt.UTC()) {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf(
			"%w: reply token is already bound to another thread or expiration",
			messaging.ErrInvalidEmailReplyToken,
		)
	}
	return record, false, nil
}

func (r *Repository) FindEmailThreadByReplyToken(
	ctx context.Context,
	lookup messaging.EmailReplyTokenLookup,
) (messaging.EmailThreadLookup, error) {
	if r == nil || r.db == nil {
		return messaging.EmailThreadLookup{}, errors.New("messaging repository is not configured")
	}
	lookup.Provider = strings.TrimSpace(lookup.Provider)
	if lookup.Provider == "" || len(lookup.TokenHash) != sha256.Size || lookup.Now.IsZero() {
		return messaging.EmailThreadLookup{}, messaging.ErrInvalidEmailReplyToken
	}
	type lookupRow struct {
		messaging.EmailThreadRecord
		ReplyTokenID        uuid.UUID `db:"reply_token_id"`
		ReplyTokenExpiresAt time.Time `db:"reply_token_expires_at"`
	}
	var row lookupRow
	err := r.db.GetContext(ctx, &row, `
		SELECT thread.id, thread.workspace_id, thread.user_id, thread.provider,
		       thread.recipient_email, thread.external_thread_id,
		       thread.root_internet_message_id, thread.latest_internet_message_id,
		       thread.context, thread.summary, thread.summary_through_sequence,
		       thread.next_message_sequence, thread.created_at, thread.updated_at,
		       token.id AS reply_token_id, token.expires_at AS reply_token_expires_at
		FROM messaging_email_reply_tokens token
		INNER JOIN messaging_email_threads thread ON thread.id = token.thread_id
		INNER JOIN workspaces workspace ON workspace.workspace_id = thread.workspace_id
		INNER JOIN users actor ON actor.user_id = thread.user_id
		INNER JOIN workspace_members member
			ON member.workspace_id = thread.workspace_id
			AND member.user_id = thread.user_id
		WHERE token.token_hash = $1
		  AND thread.provider = $2
		  AND token.revoked_at IS NULL
		  AND token.expires_at > $3
		  AND workspace.deleted_at IS NULL
		  AND actor.is_active = true
		  AND actor.is_system = false
		  AND lower(actor.email) = lower(thread.recipient_email)
		  AND member.role IN ('admin', 'member', 'guest')
	`, lookup.TokenHash, lookup.Provider, lookup.Now.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailThreadLookup{}, messaging.ErrInvalidEmailReplyToken
	}
	if err != nil {
		return messaging.EmailThreadLookup{}, fmt.Errorf("find email thread by reply token: %w", err)
	}
	return messaging.EmailThreadLookup{
		Thread:              row.EmailThreadRecord,
		ReplyTokenID:        row.ReplyTokenID,
		ReplyTokenExpiresAt: row.ReplyTokenExpiresAt,
	}, nil
}

func (r *Repository) GetEmailThread(
	ctx context.Context,
	key messaging.EmailThreadKey,
) (messaging.EmailThreadRecord, error) {
	if r == nil || r.db == nil {
		return messaging.EmailThreadRecord{}, errors.New("messaging repository is not configured")
	}
	if key.ThreadID == uuid.Nil || key.WorkspaceID == uuid.Nil || key.UserID == uuid.Nil {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	var record messaging.EmailThreadRecord
	err := r.db.GetContext(ctx, &record, `
		SELECT thread.id, thread.workspace_id, thread.user_id, thread.provider,
		       thread.recipient_email, thread.external_thread_id,
		       thread.root_internet_message_id, thread.latest_internet_message_id,
		       thread.context, thread.summary, thread.summary_through_sequence,
		       thread.next_message_sequence, thread.created_at, thread.updated_at
		FROM messaging_email_threads thread
		INNER JOIN workspaces workspace ON workspace.workspace_id = thread.workspace_id
		INNER JOIN users actor ON actor.user_id = thread.user_id
		INNER JOIN workspace_members member
			ON member.workspace_id = thread.workspace_id
			AND member.user_id = thread.user_id
		WHERE thread.id = $1
		  AND thread.workspace_id = $2
		  AND thread.user_id = $3
		  AND workspace.deleted_at IS NULL
		  AND actor.is_active = true
		  AND actor.is_system = false
		  AND lower(actor.email) = lower(thread.recipient_email)
		  AND member.role IN ('admin', 'member', 'guest')
	`, key.ThreadID, key.WorkspaceID, key.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	if err != nil {
		return messaging.EmailThreadRecord{}, fmt.Errorf("get email thread: %w", err)
	}
	return record, nil
}

// AppendEmailMessage allocates a gap-free per-thread sequence while holding
// the thread row lock. Replaying an idempotency key returns the immutable
// existing message and does not advance the sequence cursor.
func (r *Repository) AppendEmailMessage(
	ctx context.Context,
	input messaging.EmailMessageInput,
) (messaging.EmailMessageRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailMessageRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailMessageInput(input)
	if err := validateEmailMessageInput(input); err != nil {
		return messaging.EmailMessageRecord{}, false, err
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailMessageRecord{}, false, fmt.Errorf("begin email message append: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	nextSequence, err := lockEmailThreadSequence(ctx, tx, input.ThreadID, input.WorkspaceID, input.UserID)
	if err != nil {
		return messaging.EmailMessageRecord{}, false, err
	}
	existing, err := findEmailMessageByIdempotencyKey(ctx, tx, input.ThreadID, input.IdempotencyKey)
	if err == nil {
		if !emailMessageMatchesInput(existing, input) {
			return messaging.EmailMessageRecord{}, false, fmt.Errorf(
				"%w: message idempotency key was reused with different content",
				messaging.ErrInvalidEmailConversation,
			)
		}
		if err := tx.Commit(); err != nil {
			return messaging.EmailMessageRecord{}, false, fmt.Errorf("commit email message replay: %w", err)
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailMessageRecord{}, false, err
	}

	var record messaging.EmailMessageRecord
	err = tx.GetContext(ctx, &record, `
		INSERT INTO messaging_email_messages (
			thread_id, workspace_id, user_id, sequence, inbound_event_id,
			idempotency_key, direction, role, kind, provider_message_id,
			internet_message_id, in_reply_to_message_id, subject, content,
			context, provider_metadata
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, NULLIF($10, ''),
			NULLIF($11, ''), NULLIF($12, ''), $13, $14, $15, $16
		)
		RETURNING id, thread_id, sequence, inbound_event_id, idempotency_key,
		          direction, role, kind, provider_message_id, internet_message_id,
		          in_reply_to_message_id, subject, content, context,
		          provider_metadata, created_at
	`, input.ThreadID, input.WorkspaceID, input.UserID, nextSequence, input.InboundEventID,
		input.IdempotencyKey, input.Direction, input.Role, input.Kind, input.ProviderMessageID,
		input.InternetMessageID, input.InReplyToMessageID, input.Subject, input.Content,
		input.Context, input.ProviderMetadata)
	if err != nil {
		return messaging.EmailMessageRecord{}, false, fmt.Errorf("append email message: %w", err)
	}
	result, err := tx.ExecContext(ctx, `
		UPDATE messaging_email_threads
		SET next_message_sequence = $2,
		    root_internet_message_id = CASE
		        WHEN root_internet_message_id = '' THEN $3
		        ELSE root_internet_message_id
		    END,
		    latest_internet_message_id = CASE
		        WHEN $3 = '' THEN latest_internet_message_id
		        ELSE $3
		    END,
		    updated_at = NOW()
		WHERE id = $1
	`, input.ThreadID, nextSequence+1, input.InternetMessageID)
	if err != nil {
		return messaging.EmailMessageRecord{}, false, fmt.Errorf("advance email thread cursor: %w", err)
	}
	if err := requireAffectedRow(result, "advance email thread cursor"); err != nil {
		return messaging.EmailMessageRecord{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailMessageRecord{}, false, fmt.Errorf("commit email message append: %w", err)
	}
	return record, true, nil
}

func (r *Repository) ListEmailMessages(
	ctx context.Context,
	input messaging.EmailMessagePageInput,
) (messaging.EmailMessagePage, error) {
	if r == nil || r.db == nil {
		return messaging.EmailMessagePage{}, errors.New("messaging repository is not configured")
	}
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.AfterSequence < 0 {
		return messaging.EmailMessagePage{}, messaging.ErrInvalidEmailConversation
	}
	limit := input.Limit
	if limit <= 0 || limit > maximumEmailMessagePageSize {
		limit = defaultEmailMessagePageSize
	}
	messages := make([]messaging.EmailMessageRecord, 0, limit+1)
	err := r.db.SelectContext(ctx, &messages, `
		SELECT message.id, message.thread_id, message.sequence,
		       message.inbound_event_id, message.idempotency_key, message.direction,
		       message.role, message.kind, message.provider_message_id,
		       message.internet_message_id, message.in_reply_to_message_id,
		       message.subject, message.content, message.context,
		       message.provider_metadata, message.created_at
		FROM messaging_email_messages message
		INNER JOIN messaging_email_threads thread ON thread.id = message.thread_id
		WHERE message.thread_id = $1
		  AND thread.workspace_id = $2
		  AND thread.user_id = $3
		  AND message.sequence > $4
		ORDER BY message.sequence ASC
		LIMIT $5
	`, input.ThreadID, input.WorkspaceID, input.UserID, input.AfterSequence, limit+1)
	if err != nil {
		return messaging.EmailMessagePage{}, fmt.Errorf("list email messages: %w", err)
	}
	hasMore := len(messages) > limit
	if hasMore {
		messages = messages[:limit]
	}
	nextSequence := input.AfterSequence
	if len(messages) > 0 {
		nextSequence = messages[len(messages)-1].Sequence
	}
	return messaging.EmailMessagePage{Messages: messages, NextSequence: nextSequence, HasMore: hasMore}, nil
}

func (r *Repository) UpdateEmailThreadSummary(
	ctx context.Context,
	input messaging.EmailThreadSummaryUpdate,
) (messaging.EmailThreadRecord, error) {
	if r == nil || r.db == nil {
		return messaging.EmailThreadRecord{}, errors.New("messaging repository is not configured")
	}
	input.Summary = strings.TrimSpace(input.Summary)
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil ||
		input.ExpectedSummaryThroughSequence < 0 || input.ThroughSequence < input.ExpectedSummaryThroughSequence ||
		(input.ThroughSequence > 0 && input.Summary == "") {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	var record messaging.EmailThreadRecord
	err := r.db.GetContext(ctx, &record, `
		UPDATE messaging_email_threads
		SET summary = $5,
		    summary_through_sequence = $6,
		    updated_at = NOW()
		WHERE id = $1
		  AND workspace_id = $2
		  AND user_id = $3
		  AND summary_through_sequence = $4
		  AND $6 < next_message_sequence
		RETURNING id, workspace_id, user_id, provider, recipient_email,
		          external_thread_id, root_internet_message_id,
		          latest_internet_message_id, context, summary,
		          summary_through_sequence, next_message_sequence,
		          created_at, updated_at
	`, input.ThreadID, input.WorkspaceID, input.UserID,
		input.ExpectedSummaryThroughSequence, input.Summary, input.ThroughSequence)
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailThreadRecord{}, messaging.ErrEmailSummaryConflict
	}
	if err != nil {
		return messaging.EmailThreadRecord{}, fmt.Errorf("update email thread summary: %w", err)
	}
	return record, nil
}

func insertOrReadEmailThread(
	ctx context.Context,
	tx *sqlx.Tx,
	input messaging.EmailThreadInput,
) (messaging.EmailThreadRecord, bool, error) {
	var record messaging.EmailThreadRecord
	err := tx.GetContext(ctx, &record, `
		INSERT INTO messaging_email_threads (
			provider, workspace_id, user_id, recipient_email, external_thread_id,
			root_internet_message_id, latest_internet_message_id, context
		) VALUES ($1, $2, $3, $4, $5, $6, $6, $7)
		ON CONFLICT (provider, workspace_id, external_thread_id) DO NOTHING
		RETURNING id, workspace_id, user_id, provider, recipient_email,
		          external_thread_id, root_internet_message_id,
		          latest_internet_message_id, context, summary,
		          summary_through_sequence, next_message_sequence,
		          created_at, updated_at
	`, input.Provider, input.WorkspaceID, input.UserID, input.RecipientEmail,
		input.ExternalThreadID, input.RootInternetMessageID, input.Context)
	if err == nil {
		return record, true, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("create email thread: %w", err)
	}
	err = tx.GetContext(ctx, &record, `
		SELECT id, workspace_id, user_id, provider, recipient_email,
		       external_thread_id, root_internet_message_id,
		       latest_internet_message_id, context, summary,
		       summary_through_sequence, next_message_sequence,
		       created_at, updated_at
		FROM messaging_email_threads
		WHERE provider = $1 AND workspace_id = $2 AND external_thread_id = $3
	`, input.Provider, input.WorkspaceID, input.ExternalThreadID)
	if err != nil {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("read email thread: %w", err)
	}
	return record, false, nil
}

func lockEmailThreadSequence(
	ctx context.Context,
	tx *sqlx.Tx,
	threadID, workspaceID, userID uuid.UUID,
) (int64, error) {
	var sequence int64
	err := tx.GetContext(ctx, &sequence, `
		SELECT next_message_sequence
		FROM messaging_email_threads
		WHERE id = $1 AND workspace_id = $2 AND user_id = $3
		FOR UPDATE
	`, threadID, workspaceID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, messaging.ErrInvalidEmailConversation
	}
	if err != nil {
		return 0, fmt.Errorf("lock email thread: %w", err)
	}
	return sequence, nil
}

func findEmailMessageByIdempotencyKey(
	ctx context.Context,
	db sqlx.ExtContext,
	threadID uuid.UUID,
	idempotencyKey string,
) (messaging.EmailMessageRecord, error) {
	var record messaging.EmailMessageRecord
	err := sqlx.GetContext(ctx, db, &record, `
		SELECT id, thread_id, sequence, inbound_event_id, idempotency_key,
		       direction, role, kind, provider_message_id, internet_message_id,
		       in_reply_to_message_id, subject, content, context,
		       provider_metadata, created_at
		FROM messaging_email_messages
		WHERE thread_id = $1 AND idempotency_key = $2
	`, threadID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailMessageRecord{}, fmt.Errorf("read email message by idempotency key: %w", err)
	}
	return record, err
}

func normalizeEmailThreadInput(input messaging.EmailThreadInput) messaging.EmailThreadInput {
	input.Provider = strings.TrimSpace(input.Provider)
	input.RecipientEmail = strings.ToLower(strings.TrimSpace(input.RecipientEmail))
	input.ExternalThreadID = strings.TrimSpace(input.ExternalThreadID)
	input.RootInternetMessageID = strings.TrimSpace(input.RootInternetMessageID)
	input.Context = normalizeJSONObject(input.Context)
	return input
}

func validateEmailThreadInput(input messaging.EmailThreadInput) error {
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.Provider == "" ||
		input.RecipientEmail == "" || input.ExternalThreadID == "" {
		return messaging.ErrInvalidEmailConversation
	}
	if len(input.ReplyTokenHash) != sha256.Size || input.ReplyTokenExpiresAt.IsZero() {
		return messaging.ErrInvalidEmailReplyToken
	}
	if !isJSONObject(input.Context) {
		return fmt.Errorf("%w: thread context must be a JSON object", messaging.ErrInvalidEmailConversation)
	}
	return nil
}

func validateEmailReplyTokenInput(input messaging.EmailReplyTokenInput) error {
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil ||
		len(input.TokenHash) != sha256.Size || input.ExpiresAt.IsZero() {
		return messaging.ErrInvalidEmailReplyToken
	}
	return nil
}

func normalizeEmailMessageInput(input messaging.EmailMessageInput) messaging.EmailMessageInput {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.Direction = strings.TrimSpace(input.Direction)
	input.Role = strings.TrimSpace(input.Role)
	input.Kind = strings.TrimSpace(input.Kind)
	input.ProviderMessageID = strings.TrimSpace(input.ProviderMessageID)
	input.InternetMessageID = strings.TrimSpace(input.InternetMessageID)
	input.InReplyToMessageID = strings.TrimSpace(input.InReplyToMessageID)
	input.Subject = strings.TrimSpace(input.Subject)
	input.Content = strings.TrimSpace(input.Content)
	input.Context = normalizeJSONObject(input.Context)
	input.ProviderMetadata = normalizeJSONObject(input.ProviderMetadata)
	return input
}

func validateEmailMessageInput(input messaging.EmailMessageInput) error {
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil ||
		input.IdempotencyKey == "" || input.Content == "" {
		return messaging.ErrInvalidEmailConversation
	}
	if input.Direction != messaging.EmailMessageDirectionInbound && input.Direction != messaging.EmailMessageDirectionOutbound {
		return fmt.Errorf("%w: unsupported email direction %q", messaging.ErrInvalidEmailConversation, input.Direction)
	}
	if input.Role != messaging.EmailMessageRoleUser && input.Role != messaging.EmailMessageRoleAssistant && input.Role != messaging.EmailMessageRoleSystem {
		return fmt.Errorf("%w: unsupported email role %q", messaging.ErrInvalidEmailConversation, input.Role)
	}
	switch input.Kind {
	case messaging.EmailMessageKindGuidance,
		messaging.EmailMessageKindReply,
		messaging.EmailMessageKindAnswer,
		messaging.EmailMessageKindProposal,
		messaging.EmailMessageKindConfirmation,
		messaging.EmailMessageKindReceipt,
		messaging.EmailMessageKindError:
	default:
		return fmt.Errorf("%w: unsupported email message kind %q", messaging.ErrInvalidEmailConversation, input.Kind)
	}
	if !isJSONObject(input.Context) || !isJSONObject(input.ProviderMetadata) {
		return fmt.Errorf("%w: email message metadata must be JSON objects", messaging.ErrInvalidEmailConversation)
	}
	return nil
}

func emailMessageMatchesInput(record messaging.EmailMessageRecord, input messaging.EmailMessageInput) bool {
	return record.ThreadID == input.ThreadID &&
		equalOptionalUUID(record.InboundEventID, input.InboundEventID) &&
		record.IdempotencyKey == input.IdempotencyKey &&
		record.Direction == input.Direction &&
		record.Role == input.Role &&
		record.Kind == input.Kind &&
		strings.TrimSpace(valueOrEmptyString(record.ProviderMessageID)) == input.ProviderMessageID &&
		strings.TrimSpace(valueOrEmptyString(record.InternetMessageID)) == input.InternetMessageID &&
		strings.TrimSpace(valueOrEmptyString(record.InReplyToMessageID)) == input.InReplyToMessageID &&
		record.Subject == input.Subject &&
		record.Content == input.Content &&
		jsonObjectsEqual(record.Context, input.Context) &&
		jsonObjectsEqual(record.ProviderMetadata, input.ProviderMetadata)
}

func normalizeJSONObject(value json.RawMessage) json.RawMessage {
	if len(value) == 0 {
		return json.RawMessage(`{}`)
	}
	return value
}

func isJSONObject(value json.RawMessage) bool {
	var object map[string]any
	return json.Unmarshal(value, &object) == nil && object != nil
}

func jsonObjectsEqual(left, right json.RawMessage) bool {
	var leftObject, rightObject map[string]any
	if json.Unmarshal(normalizeJSONObject(left), &leftObject) != nil || json.Unmarshal(normalizeJSONObject(right), &rightObject) != nil {
		return false
	}
	leftJSON, _ := json.Marshal(leftObject)
	rightJSON, _ := json.Marshal(rightObject)
	return string(leftJSON) == string(rightJSON)
}

func truncateEmailError(message string) string {
	runes := []rune(strings.TrimSpace(message))
	if len(runes) <= maximumEmailErrorRunes {
		return string(runes)
	}
	return string(runes[:maximumEmailErrorRunes])
}
