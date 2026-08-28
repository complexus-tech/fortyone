package messagingrepository

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	defaultEmailMessagePageSize = 50
	maximumEmailMessagePageSize = 100
	maximumEmailErrorRunes      = 2_000
)

func (repository *Repository) HasEarlierInboundEvent(
	ctx context.Context,
	provider, externalWorkspaceID string,
	currentID uuid.UUID,
) (bool, error) {
	if !repository.configured() {
		return false, errors.New("messaging repository is not configured")
	}
	provider = strings.TrimSpace(provider)
	externalWorkspaceID = strings.TrimSpace(externalWorkspaceID)
	if provider == "" || externalWorkspaceID == "" || currentID == uuid.Nil {
		return false, messaging.ErrInvalidEmailConversation
	}
	earlier, err := repository.queries.HasEarlierInboundEvent(ctx, messagingsql.HasEarlierInboundEventParams{
		CurrentID: currentID, Provider: provider, ExternalWorkspaceID: externalWorkspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("check earlier email reply event: %w", err)
	}
	return earlier, nil
}

// AcquireEmailThreadProcessingLease reserves one physical pool connection for
// the lifetime of the session advisory lock. Releasing the connection before
// unlocking would leak ownership into an unrelated borrower.
func (repository *Repository) AcquireEmailThreadProcessingLease(
	ctx context.Context,
	lease messaging.EmailThreadProcessingLease,
) (func() error, error) {
	if !repository.configured() || repository.pool == nil {
		return nil, errors.New("messaging repository is not configured")
	}
	if lease.ThreadID == uuid.Nil || lease.WorkspaceID == uuid.Nil || lease.UserID == uuid.Nil {
		return nil, messaging.ErrInvalidEmailConversation
	}
	connection, err := repository.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("reserve email thread lock connection: %w", err)
	}
	queries := messagingsql.New(connection)
	lockKey := lease.WorkspaceID.String() + ":" + lease.ThreadID.String() + ":" + lease.UserID.String()
	locked, err := queries.TryEmailThreadAdvisoryLock(ctx, messagingsql.TryEmailThreadAdvisoryLockParams{LockKey: lockKey})
	if err != nil {
		connection.Release()
		return nil, fmt.Errorf("acquire email thread lock: %w", err)
	}
	if !locked {
		connection.Release()
		return nil, newLeaseBusyError("messaging email thread")
	}
	return func() error {
		defer connection.Release()
		released, err := queries.ReleaseEmailThreadAdvisoryLock(
			context.WithoutCancel(ctx),
			messagingsql.ReleaseEmailThreadAdvisoryLockParams{LockKey: lockKey},
		)
		if err != nil {
			return fmt.Errorf("release email thread lock: %w", err)
		}
		if !released {
			return errors.New("email thread lock was not held by its reserved connection")
		}
		return nil
	}, nil
}

func (repository *Repository) CreateEmailThread(
	ctx context.Context,
	input messaging.EmailThreadInput,
) (messaging.EmailThreadRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailThreadRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailThreadInput(input)
	if err := validateEmailThreadInput(input); err != nil {
		return messaging.EmailThreadRecord{}, false, err
	}
	var (
		thread  messaging.EmailThreadRecord
		created bool
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		var err error
		thread, created, err = insertOrReadEmailThread(ctx, queries, input)
		if err != nil {
			return err
		}
		if thread.UserID != input.UserID || !strings.EqualFold(thread.RecipientEmail, input.RecipientEmail) ||
			thread.RootInternetMessageID != input.RootInternetMessageID || !jsonObjectsEqual(thread.Context, input.Context) {
			return fmt.Errorf("%w: external thread is bound to different immutable context", messaging.ErrInvalidEmailConversation)
		}
		_, _, err = createEmailReplyTokenAlias(ctx, queries, messaging.EmailReplyTokenInput{
			ThreadID: thread.ID, WorkspaceID: thread.WorkspaceID, UserID: thread.UserID,
			TokenHash: input.ReplyTokenHash, ExpiresAt: input.ReplyTokenExpiresAt,
		})
		return err
	})
	if err != nil {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("create email thread transaction: %w", err)
	}
	return thread, created, nil
}

func (repository *Repository) CreateEmailReplyTokenAlias(
	ctx context.Context,
	input messaging.EmailReplyTokenInput,
) (messaging.EmailReplyTokenRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailReplyTokenRecord{}, false, errors.New("messaging repository is not configured")
	}
	return createEmailReplyTokenAlias(ctx, repository.queries, input)
}

func createEmailReplyTokenAlias(
	ctx context.Context,
	queries messagingsql.Querier,
	input messaging.EmailReplyTokenInput,
) (messaging.EmailReplyTokenRecord, bool, error) {
	if err := validateEmailReplyTokenInput(input); err != nil {
		return messaging.EmailReplyTokenRecord{}, false, err
	}
	row, err := queries.InsertEmailReplyToken(ctx, messagingsql.InsertEmailReplyTokenParams{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		TokenHash: input.TokenHash, ExpiresAt: input.ExpiresAt.UTC(),
	})
	if err == nil {
		return emailReplyTokenRecord(row.ID, row.ThreadID, row.ExpiresAt, row.RevokedAt, row.CreatedAt), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf("create email reply token alias: %w", err)
	}
	existing, err := queries.GetEmailReplyToken(ctx, messagingsql.GetEmailReplyTokenParams{TokenHash: input.TokenHash})
	if err != nil {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf("read email reply token alias: %w", err)
	}
	record := emailReplyTokenRecord(existing.ID, existing.ThreadID, existing.ExpiresAt, existing.RevokedAt, existing.CreatedAt)
	if record.ThreadID != input.ThreadID || !record.ExpiresAt.Equal(input.ExpiresAt.UTC()) {
		return messaging.EmailReplyTokenRecord{}, false, fmt.Errorf(
			"%w: reply token is already bound to another thread or expiration",
			messaging.ErrInvalidEmailReplyToken,
		)
	}
	return record, false, nil
}

func (repository *Repository) FindEmailThreadByReplyToken(
	ctx context.Context,
	lookup messaging.EmailReplyTokenLookup,
) (messaging.EmailThreadLookup, error) {
	if !repository.configured() {
		return messaging.EmailThreadLookup{}, errors.New("messaging repository is not configured")
	}
	lookup.Provider = strings.TrimSpace(lookup.Provider)
	if lookup.Provider == "" || len(lookup.TokenHash) != sha256.Size || lookup.Now.IsZero() {
		return messaging.EmailThreadLookup{}, messaging.ErrInvalidEmailReplyToken
	}
	row, err := repository.queries.FindEmailThreadByReplyToken(ctx, messagingsql.FindEmailThreadByReplyTokenParams{
		TokenHash: lookup.TokenHash, Provider: lookup.Provider, Now: lookup.Now.UTC(),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailThreadLookup{}, messaging.ErrInvalidEmailReplyToken
	}
	if err != nil {
		return messaging.EmailThreadLookup{}, fmt.Errorf("find email thread by reply token: %w", err)
	}
	return messaging.EmailThreadLookup{
		Thread: messaging.EmailThreadRecord{
			ID: row.ID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
			Provider: row.Provider, RecipientEmail: row.RecipientEmail,
			ExternalThreadID: row.ExternalThreadID, RootInternetMessageID: row.RootInternetMessageID,
			LatestInternetMessageID: row.LatestInternetMessageID, Context: json.RawMessage(row.Context),
			Summary: row.Summary, SummaryThroughSequence: row.SummaryThroughSequence,
			NextMessageSequence: row.NextMessageSequence, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		},
		ReplyTokenID: row.ReplyTokenID, ReplyTokenExpiresAt: row.ReplyTokenExpiresAt,
	}, nil
}

func (repository *Repository) GetEmailThread(
	ctx context.Context,
	key messaging.EmailThreadKey,
) (messaging.EmailThreadRecord, error) {
	if !repository.configured() {
		return messaging.EmailThreadRecord{}, errors.New("messaging repository is not configured")
	}
	if key.ThreadID == uuid.Nil || key.WorkspaceID == uuid.Nil || key.UserID == uuid.Nil {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	row, err := repository.queries.GetAuthorizedEmailThread(ctx, messagingsql.GetAuthorizedEmailThreadParams{
		ThreadID: key.ThreadID, WorkspaceID: key.WorkspaceID, UserID: key.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	if err != nil {
		return messaging.EmailThreadRecord{}, fmt.Errorf("get email thread: %w", err)
	}
	return toEmailThreadRecord(row), nil
}

func (repository *Repository) AppendEmailMessage(
	ctx context.Context,
	input messaging.EmailMessageInput,
) (messaging.EmailMessageRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailMessageRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailMessageInput(input)
	if err := validateEmailMessageInput(input); err != nil {
		return messaging.EmailMessageRecord{}, false, err
	}
	var (
		record  messaging.EmailMessageRecord
		created bool
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		nextSequence, err := lockEmailThreadSequence(ctx, queries, input.ThreadID, input.WorkspaceID, input.UserID)
		if err != nil {
			return err
		}
		existing, err := findEmailMessageByIdempotencyKey(ctx, queries, input.ThreadID, input.IdempotencyKey)
		if err == nil {
			if !emailMessageMatchesInput(existing, input) {
				return fmt.Errorf("%w: message idempotency key was reused with different content", messaging.ErrInvalidEmailConversation)
			}
			record = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		row, err := queries.InsertEmailMessage(ctx, messagingsql.InsertEmailMessageParams{
			ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
			Sequence: nextSequence, InboundEventID: input.InboundEventID,
			IdempotencyKey: input.IdempotencyKey, Direction: input.Direction, Role: input.Role,
			Kind: input.Kind, ProviderMessageID: input.ProviderMessageID,
			InternetMessageID: input.InternetMessageID, InReplyToMessageID: input.InReplyToMessageID,
			Subject: input.Subject, Content: input.Content, Context: input.Context,
			ProviderMetadata: input.ProviderMetadata,
		})
		if err != nil {
			return fmt.Errorf("append email message: %w", err)
		}
		record = emailMessageRecord(
			row.ID, row.ThreadID, row.Sequence, row.InboundEventID, row.IdempotencyKey,
			row.Direction, row.Role, row.Kind, row.ProviderMessageID, row.InternetMessageID,
			row.InReplyToMessageID, row.Subject, row.Content, row.Context, row.ProviderMetadata, row.CreatedAt,
		)
		affected, err := queries.AdvanceEmailThreadCursor(ctx, messagingsql.AdvanceEmailThreadCursorParams{
			NextMessageSequence: nextSequence + 1, InternetMessageID: input.InternetMessageID,
			ThreadID: input.ThreadID,
		})
		if err != nil {
			return fmt.Errorf("advance email thread cursor: %w", err)
		}
		if err := requireAffectedRows(affected, "advance email thread cursor"); err != nil {
			return err
		}
		created = true
		return nil
	})
	if err != nil {
		return messaging.EmailMessageRecord{}, false, err
	}
	return record, created, nil
}

func (repository *Repository) ListEmailMessages(
	ctx context.Context,
	input messaging.EmailMessagePageInput,
) (messaging.EmailMessagePage, error) {
	if !repository.configured() {
		return messaging.EmailMessagePage{}, errors.New("messaging repository is not configured")
	}
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.AfterSequence < 0 {
		return messaging.EmailMessagePage{}, messaging.ErrInvalidEmailConversation
	}
	limit := input.Limit
	if limit <= 0 || limit > maximumEmailMessagePageSize {
		limit = defaultEmailMessagePageSize
	}
	rows, err := repository.queries.ListEmailMessages(ctx, messagingsql.ListEmailMessagesParams{
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		AfterSequence: input.AfterSequence, RowLimit: int32(limit + 1),
	})
	if err != nil {
		return messaging.EmailMessagePage{}, fmt.Errorf("list email messages: %w", err)
	}
	messages := make([]messaging.EmailMessageRecord, 0, len(rows))
	for _, row := range rows {
		messages = append(messages, emailMessageRecord(
			row.ID, row.ThreadID, row.Sequence, row.InboundEventID, row.IdempotencyKey,
			row.Direction, row.Role, row.Kind, row.ProviderMessageID, row.InternetMessageID,
			row.InReplyToMessageID, row.Subject, row.Content, row.Context, row.ProviderMetadata, row.CreatedAt,
		))
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

func (repository *Repository) UpdateEmailThreadSummary(
	ctx context.Context,
	input messaging.EmailThreadSummaryUpdate,
) (messaging.EmailThreadRecord, error) {
	if !repository.configured() {
		return messaging.EmailThreadRecord{}, errors.New("messaging repository is not configured")
	}
	input.Summary = strings.TrimSpace(input.Summary)
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil ||
		input.ExpectedSummaryThroughSequence < 0 || input.ThroughSequence < input.ExpectedSummaryThroughSequence ||
		(input.ThroughSequence > 0 && input.Summary == "") {
		return messaging.EmailThreadRecord{}, messaging.ErrInvalidEmailConversation
	}
	row, err := repository.queries.UpdateEmailThreadSummary(ctx, messagingsql.UpdateEmailThreadSummaryParams{
		Summary: input.Summary, ThroughSequence: input.ThroughSequence,
		ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		ExpectedSummaryThroughSequence: input.ExpectedSummaryThroughSequence,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailThreadRecord{}, messaging.ErrEmailSummaryConflict
	}
	if err != nil {
		return messaging.EmailThreadRecord{}, fmt.Errorf("update email thread summary: %w", err)
	}
	return toEmailThreadRecord(row), nil
}

func insertOrReadEmailThread(
	ctx context.Context,
	queries messagingsql.Querier,
	input messaging.EmailThreadInput,
) (messaging.EmailThreadRecord, bool, error) {
	row, err := queries.InsertEmailThread(ctx, messagingsql.InsertEmailThreadParams{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		RecipientEmail: input.RecipientEmail, ExternalThreadID: input.ExternalThreadID,
		RootInternetMessageID: input.RootInternetMessageID, Context: input.Context,
	})
	if err == nil {
		return toEmailThreadRecord(row), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("create email thread: %w", err)
	}
	existing, err := queries.GetEmailThreadByExternalID(ctx, messagingsql.GetEmailThreadByExternalIDParams{
		Provider: input.Provider, WorkspaceID: input.WorkspaceID, ExternalThreadID: input.ExternalThreadID,
	})
	if err != nil {
		return messaging.EmailThreadRecord{}, false, fmt.Errorf("read email thread: %w", err)
	}
	return toEmailThreadRecord(existing), false, nil
}

func lockEmailThreadSequence(
	ctx context.Context,
	queries messagingsql.Querier,
	threadID, workspaceID, userID uuid.UUID,
) (int64, error) {
	sequence, err := queries.LockEmailThreadSequence(ctx, messagingsql.LockEmailThreadSequenceParams{
		ThreadID: threadID, WorkspaceID: workspaceID, UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, messaging.ErrInvalidEmailConversation
	}
	if err != nil {
		return 0, fmt.Errorf("lock email thread: %w", err)
	}
	return sequence, nil
}

func findEmailMessageByIdempotencyKey(
	ctx context.Context,
	queries messagingsql.Querier,
	threadID uuid.UUID,
	idempotencyKey string,
) (messaging.EmailMessageRecord, error) {
	row, err := queries.GetEmailMessageByIdempotencyKey(ctx, messagingsql.GetEmailMessageByIdempotencyKeyParams{
		ThreadID: threadID, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return messaging.EmailMessageRecord{}, err
		}
		return messaging.EmailMessageRecord{}, fmt.Errorf("read email message by idempotency key: %w", err)
	}
	return emailMessageRecord(
		row.ID, row.ThreadID, row.Sequence, row.InboundEventID, row.IdempotencyKey,
		row.Direction, row.Role, row.Kind, row.ProviderMessageID, row.InternetMessageID,
		row.InReplyToMessageID, row.Subject, row.Content, row.Context, row.ProviderMetadata, row.CreatedAt,
	), nil
}

func toEmailThreadRecord(row messagingsql.MessagingEmailThread) messaging.EmailThreadRecord {
	return messaging.EmailThreadRecord{
		ID: row.ID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		Provider: row.Provider, RecipientEmail: row.RecipientEmail,
		ExternalThreadID: row.ExternalThreadID, RootInternetMessageID: row.RootInternetMessageID,
		LatestInternetMessageID: row.LatestInternetMessageID, Context: json.RawMessage(row.Context),
		Summary: row.Summary, SummaryThroughSequence: row.SummaryThroughSequence,
		NextMessageSequence: row.NextMessageSequence, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func emailReplyTokenRecord(
	id, threadID uuid.UUID,
	expiresAt time.Time,
	revokedAt *time.Time,
	createdAt time.Time,
) messaging.EmailReplyTokenRecord {
	return messaging.EmailReplyTokenRecord{
		ID: id, ThreadID: threadID, ExpiresAt: expiresAt, RevokedAt: revokedAt, CreatedAt: createdAt,
	}
}

func emailMessageRecord(
	id, threadID uuid.UUID,
	sequence int64,
	inboundEventID *uuid.UUID,
	idempotencyKey, direction, role, kind string,
	providerMessageID, internetMessageID, inReplyToMessageID *string,
	subject, content string,
	contextJSON, providerMetadata []byte,
	createdAt time.Time,
) messaging.EmailMessageRecord {
	return messaging.EmailMessageRecord{
		ID: id, ThreadID: threadID, Sequence: sequence, InboundEventID: inboundEventID,
		IdempotencyKey: idempotencyKey, Direction: direction, Role: role, Kind: kind,
		ProviderMessageID: providerMessageID, InternetMessageID: internetMessageID,
		InReplyToMessageID: inReplyToMessageID, Subject: subject, Content: content,
		Context: json.RawMessage(contextJSON), ProviderMetadata: json.RawMessage(providerMetadata),
		CreatedAt: createdAt,
	}
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
		record.Direction == input.Direction && record.Role == input.Role && record.Kind == input.Kind &&
		strings.TrimSpace(valueOrEmptyString(record.ProviderMessageID)) == input.ProviderMessageID &&
		strings.TrimSpace(valueOrEmptyString(record.InternetMessageID)) == input.InternetMessageID &&
		strings.TrimSpace(valueOrEmptyString(record.InReplyToMessageID)) == input.InReplyToMessageID &&
		record.Subject == input.Subject && record.Content == input.Content &&
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
