package messagingrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/domain"
	messagingsql "github.com/complexus-tech/projects-api/internal/modules/messaging/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) RegisterEmailActionProposal(
	ctx context.Context,
	input messaging.EmailActionProposalInput,
) (messaging.EmailActionProposalRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailActionProposalInput(input)
	if err := validateEmailActionProposalInput(input); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	var (
		record  messaging.EmailActionProposalRecord
		created bool
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		if _, err := lockEmailThreadSequence(ctx, queries, input.ThreadID, input.WorkspaceID, input.UserID); err != nil {
			return err
		}
		existing, err := findEmailActionProposalByIdempotencyKey(ctx, queries, input.ThreadID, input.IdempotencyKey)
		if err == nil {
			if !emailActionProposalMatchesInput(existing, input) {
				return fmt.Errorf("%w: proposal idempotency key was reused with different content", messaging.ErrInvalidEmailProposal)
			}
			record = existing
			return nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		if err := queries.SupersedePendingEmailProposals(ctx, messagingsql.SupersedePendingEmailProposalsParams{
			Now: input.Now.UTC(), ThreadID: input.ThreadID,
		}); err != nil {
			return fmt.Errorf("supersede pending email proposal: %w", err)
		}
		row, err := queries.InsertEmailActionProposal(ctx, messagingsql.InsertEmailActionProposalParams{
			ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
			SourceMessageID: input.SourceMessageID, IdempotencyKey: input.IdempotencyKey,
			ActionKind: input.ActionKind, EntityType: input.EntityType, EntityID: input.EntityID,
			ExpectedEntityVersion: input.ExpectedEntityVersion, ProposedDiff: input.ProposedDiff,
			ExpiresAt: input.ExpiresAt.UTC(),
		})
		if err != nil {
			return fmt.Errorf("register email action proposal: %w", err)
		}
		record = toEmailActionProposalRecord(row)
		created = true
		return nil
	})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	return record, created, nil
}

func (repository *Repository) ListPendingEmailActionProposals(
	ctx context.Context,
	input messaging.EmailActionProposalListInput,
) ([]messaging.EmailActionProposalRecord, error) {
	if !repository.configured() {
		return nil, errors.New("messaging repository is not configured")
	}
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.Now.IsZero() {
		return nil, messaging.ErrInvalidEmailProposal
	}
	var records []messaging.EmailActionProposalRecord
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		now := input.Now.UTC()
		if err := queries.ExpirePendingEmailProposals(ctx, messagingsql.ExpirePendingEmailProposalsParams{
			Now: &now, ThreadID: input.ThreadID,
			WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		}); err != nil {
			return fmt.Errorf("expire stale email action proposal: %w", err)
		}
		rows, err := queries.ListPendingEmailActionProposals(ctx, messagingsql.ListPendingEmailActionProposalsParams{
			ThreadID: input.ThreadID, WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		})
		if err != nil {
			return fmt.Errorf("list pending email action proposals: %w", err)
		}
		records = make([]messaging.EmailActionProposalRecord, 0, len(rows))
		for _, row := range rows {
			records = append(records, toEmailActionProposalRecord(row))
		}
		return nil
	})
	return records, err
}

func (repository *Repository) FindLatestEmailActionProposalForControl(
	ctx context.Context,
	lookup messaging.EmailActionProposalControlLookup,
) (messaging.EmailActionProposalRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if lookup.ThreadID == uuid.Nil || lookup.WorkspaceID == uuid.Nil || lookup.UserID == uuid.Nil {
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	params := messagingsql.FindLatestEmailActionProposalForConfirmParams{
		ThreadID: lookup.ThreadID, WorkspaceID: lookup.WorkspaceID, UserID: lookup.UserID,
	}
	var (
		row messagingsql.MessagingEmailActionProposal
		err error
	)
	switch lookup.Control {
	case messaging.EmailActionProposalConfirmed:
		row, err = repository.queries.FindLatestEmailActionProposalForConfirm(ctx, params)
	case messaging.EmailActionProposalCancelled:
		row, err = repository.queries.FindLatestEmailActionProposalForCancel(
			ctx,
			messagingsql.FindLatestEmailActionProposalForCancelParams(params),
		)
	default:
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, false, nil
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("find latest email proposal for control: %w", err)
	}
	return toEmailActionProposalRecord(row), true, nil
}

func (repository *Repository) GetEmailActionProposal(
	ctx context.Context,
	key messaging.EmailActionProposalKey,
) (messaging.EmailActionProposalRecord, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, errors.New("messaging repository is not configured")
	}
	if key.ProposalID == uuid.Nil || key.ThreadID == uuid.Nil || key.WorkspaceID == uuid.Nil || key.UserID == uuid.Nil {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	row, err := repository.queries.GetEmailActionProposal(ctx, messagingsql.GetEmailActionProposalParams{
		ProposalID: key.ProposalID, ThreadID: key.ThreadID,
		WorkspaceID: key.WorkspaceID, UserID: key.UserID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("get email action proposal: %w", err)
	}
	return toEmailActionProposalRecord(row), nil
}

func (repository *Repository) DecideEmailActionProposal(
	ctx context.Context,
	decision messaging.EmailActionProposalDecision,
) (messaging.EmailActionProposalRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if err := validateEmailActionProposalDecision(decision); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	var (
		record    messaging.EmailActionProposalRecord
		duplicate bool
		resultErr error
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		if err := requireActiveEmailReplyToken(ctx, queries, decision); err != nil {
			return err
		}
		var err error
		record, err = lockEmailActionProposal(ctx, queries, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
		if err != nil {
			return err
		}
		now := decision.Now.UTC()
		if record.Status == messaging.EmailActionProposalPending && !now.Before(record.ExpiresAt.UTC()) {
			if err := transitionEmailActionProposalDecision(ctx, queries, record.ID, messaging.EmailActionProposalExpired, now); err != nil {
				return err
			}
			record, err = lockEmailActionProposal(ctx, queries, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
			resultErr = messaging.ErrEmailProposalExpired
			return err
		}
		if emailProposalDecisionAlreadyWon(record.Status, decision.Decision) {
			duplicate = true
			return nil
		}
		if record.Status == messaging.EmailActionProposalExpired {
			return messaging.ErrEmailProposalExpired
		}
		if record.Status != messaging.EmailActionProposalPending {
			return fmt.Errorf("%w: proposal is already %s", messaging.ErrEmailProposalConflict, record.Status)
		}
		if err := transitionEmailActionProposalDecision(ctx, queries, record.ID, decision.Decision, now); err != nil {
			return err
		}
		record, err = lockEmailActionProposal(ctx, queries, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
		return err
	})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	return record, duplicate, resultErr
}

func (repository *Repository) ClaimEmailActionProposalApply(
	ctx context.Context,
	claim messaging.EmailActionProposalApplyClaim,
) (messaging.EmailActionProposalRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if claim.ProposalID == uuid.Nil || claim.ThreadID == uuid.Nil || claim.WorkspaceID == uuid.Nil ||
		claim.UserID == uuid.Nil || claim.Now.IsZero() {
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	if claim.RetryAfter <= 0 {
		claim.RetryAfter = messagingLeaseDuration
	}
	var (
		record  messaging.EmailActionProposalRecord
		claimed bool
	)
	err := repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		var err error
		record, err = lockEmailActionProposal(ctx, queries, claim.ProposalID, claim.ThreadID, claim.WorkspaceID, claim.UserID)
		if err != nil {
			return err
		}
		now := claim.Now.UTC()
		switch record.Status {
		case messaging.EmailActionProposalApplied:
			return nil
		case messaging.EmailActionProposalConfirmed, messaging.EmailActionProposalFailed:
		case messaging.EmailActionProposalApplying:
			if record.ApplyingAt != nil && now.Sub(record.ApplyingAt.UTC()) < claim.RetryAfter {
				return messaging.ErrEmailProposalApplyBusy
			}
		default:
			return fmt.Errorf("%w: proposal in status %s cannot be applied", messaging.ErrEmailProposalConflict, record.Status)
		}
		row, err := queries.ClaimEmailActionProposalApply(ctx, messagingsql.ClaimEmailActionProposalApplyParams{
			Now: &now, ProposalID: record.ID,
		})
		if err != nil {
			return fmt.Errorf("claim email proposal apply: %w", err)
		}
		record = toEmailActionProposalRecord(row)
		claimed = true
		return nil
	})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	return record, claimed, nil
}

func (repository *Repository) CompleteEmailActionProposalApply(
	ctx context.Context,
	completion messaging.EmailActionProposalApplyCompletion,
) (messaging.EmailActionProposalRecord, bool, error) {
	if !repository.configured() {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	completion.Result = normalizeJSONObject(completion.Result)
	if err := validateEmailActionProposalApplyCompletion(completion); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	databaseApplyAttempt, err := safecast.Int32(completion.ApplyAttempt)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("validate email proposal apply attempt: %w", err)
	}
	var (
		record    messaging.EmailActionProposalRecord
		duplicate bool
	)
	err = repository.withinTransaction(ctx, func(queries messagingsql.Querier) error {
		var err error
		record, err = lockEmailActionProposal(ctx, queries, completion.ProposalID, completion.ThreadID, completion.WorkspaceID, completion.UserID)
		if err != nil {
			return err
		}
		if record.Status == completion.Status && record.ApplyAttempt == completion.ApplyAttempt {
			duplicate = true
			return nil
		}
		if record.Status != messaging.EmailActionProposalApplying || record.ApplyAttempt != completion.ApplyAttempt {
			return fmt.Errorf("%w: apply attempt %d is no longer active", messaging.ErrEmailProposalConflict, completion.ApplyAttempt)
		}
		now := completion.Now.UTC()
		var row messagingsql.MessagingEmailActionProposal
		if completion.Status == messaging.EmailActionProposalFailed {
			row, err = queries.MarkEmailActionProposalFailed(ctx, messagingsql.MarkEmailActionProposalFailedParams{
				Result: completion.Result, LastError: truncateEmailError(completion.ErrorMessage),
				Now: &now, ProposalID: record.ID, ApplyAttempt: databaseApplyAttempt,
			})
		} else {
			row, err = queries.MarkEmailActionProposalApplied(ctx, messagingsql.MarkEmailActionProposalAppliedParams{
				Result: completion.Result, Now: &now, ProposalID: record.ID, ApplyAttempt: databaseApplyAttempt,
			})
		}
		if err != nil {
			return fmt.Errorf("complete email proposal apply: %w", err)
		}
		record = toEmailActionProposalRecord(row)
		return nil
	})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	return record, duplicate, nil
}

func requireActiveEmailReplyToken(
	ctx context.Context,
	queries messagingsql.Querier,
	decision messaging.EmailActionProposalDecision,
) error {
	exists, err := queries.HasActiveEmailReplyToken(ctx, messagingsql.HasActiveEmailReplyTokenParams{
		ThreadID: decision.ThreadID, WorkspaceID: decision.WorkspaceID, UserID: decision.UserID,
		TokenHash: decision.ReplyTokenHash, Now: decision.Now.UTC(),
	})
	if err != nil {
		return fmt.Errorf("verify email proposal reply token: %w", err)
	}
	if !exists {
		return messaging.ErrInvalidEmailReplyToken
	}
	return nil
}

func transitionEmailActionProposalDecision(
	ctx context.Context,
	queries messagingsql.Querier,
	proposalID uuid.UUID,
	status messaging.EmailActionProposalStatus,
	now time.Time,
) error {
	switch status {
	case messaging.EmailActionProposalConfirmed,
		messaging.EmailActionProposalCancelled,
		messaging.EmailActionProposalExpired,
		messaging.EmailActionProposalSuperseded:
	default:
		return messaging.ErrInvalidEmailProposal
	}
	affected, err := queries.TransitionEmailActionProposalDecision(ctx, messagingsql.TransitionEmailActionProposalDecisionParams{
		Status: string(status), Now: &now, ProposalID: proposalID,
	})
	if err != nil {
		return fmt.Errorf("transition email action proposal: %w", err)
	}
	return requireAffectedRows(affected, "transition email action proposal")
}

func emailProposalDecisionAlreadyWon(status, decision messaging.EmailActionProposalStatus) bool {
	if status == decision {
		return true
	}
	if decision != messaging.EmailActionProposalConfirmed {
		return false
	}
	switch status {
	case messaging.EmailActionProposalApplying,
		messaging.EmailActionProposalApplied,
		messaging.EmailActionProposalFailed:
		return true
	default:
		return false
	}
}

func lockEmailActionProposal(
	ctx context.Context,
	queries messagingsql.Querier,
	proposalID, threadID, workspaceID, userID uuid.UUID,
) (messaging.EmailActionProposalRecord, error) {
	row, err := queries.LockEmailActionProposal(ctx, messagingsql.LockEmailActionProposalParams{
		ProposalID: proposalID, ThreadID: threadID, WorkspaceID: workspaceID, UserID: userID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("lock email action proposal: %w", err)
	}
	return toEmailActionProposalRecord(row), nil
}

func findEmailActionProposalByIdempotencyKey(
	ctx context.Context,
	queries messagingsql.Querier,
	threadID uuid.UUID,
	idempotencyKey string,
) (messaging.EmailActionProposalRecord, error) {
	row, err := queries.GetEmailActionProposalByIdempotencyKey(ctx, messagingsql.GetEmailActionProposalByIdempotencyKeyParams{
		ThreadID: threadID, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return messaging.EmailActionProposalRecord{}, err
		}
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("read email action proposal: %w", err)
	}
	return toEmailActionProposalRecord(row), nil
}

func toEmailActionProposalRecord(row messagingsql.MessagingEmailActionProposal) messaging.EmailActionProposalRecord {
	return messaging.EmailActionProposalRecord{
		ID: row.ID, ThreadID: row.ThreadID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		SourceMessageID: row.SourceMessageID, IdempotencyKey: row.IdempotencyKey,
		ActionKind: row.ActionKind, EntityType: row.EntityType, EntityID: row.EntityID,
		ExpectedEntityVersion: row.ExpectedEntityVersion, ProposedDiff: row.ProposedDiff,
		Status: messaging.EmailActionProposalStatus(row.Status), ApplyAttempt: int(row.ApplyAttempt),
		Result: row.Result, LastError: row.LastError, ExpiresAt: row.ExpiresAt,
		ConfirmedAt: row.ConfirmedAt, ApplyingAt: row.ApplyingAt, AppliedAt: row.AppliedAt,
		FailedAt: row.FailedAt, CancelledAt: row.CancelledAt, ExpiredAt: row.ExpiredAt,
		SupersededAt: row.SupersededAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func normalizeEmailActionProposalInput(input messaging.EmailActionProposalInput) messaging.EmailActionProposalInput {
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	input.ActionKind = strings.TrimSpace(input.ActionKind)
	input.EntityType = strings.TrimSpace(input.EntityType)
	input.ExpectedEntityVersion = strings.TrimSpace(input.ExpectedEntityVersion)
	input.ProposedDiff = normalizeJSONObject(input.ProposedDiff)
	return input
}

func validateEmailActionProposalInput(input messaging.EmailActionProposalInput) error {
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil ||
		input.SourceMessageID == uuid.Nil || input.EntityID == uuid.Nil || input.IdempotencyKey == "" ||
		input.ActionKind == "" || input.EntityType == "" || input.ExpectedEntityVersion == "" ||
		input.ExpiresAt.IsZero() || input.Now.IsZero() || !input.ExpiresAt.After(input.Now) ||
		!isJSONObject(input.ProposedDiff) {
		return messaging.ErrInvalidEmailProposal
	}
	return nil
}

func validateEmailActionProposalDecision(decision messaging.EmailActionProposalDecision) error {
	if decision.ProposalID == uuid.Nil || decision.ThreadID == uuid.Nil || decision.WorkspaceID == uuid.Nil ||
		decision.UserID == uuid.Nil || len(decision.ReplyTokenHash) != sha256DigestSize || decision.Now.IsZero() {
		return messaging.ErrInvalidEmailProposal
	}
	if decision.Decision != messaging.EmailActionProposalConfirmed &&
		decision.Decision != messaging.EmailActionProposalCancelled &&
		decision.Decision != messaging.EmailActionProposalSuperseded {
		return messaging.ErrInvalidEmailProposal
	}
	return nil
}

func validateEmailActionProposalApplyCompletion(completion messaging.EmailActionProposalApplyCompletion) error {
	if completion.ProposalID == uuid.Nil || completion.ThreadID == uuid.Nil || completion.WorkspaceID == uuid.Nil ||
		completion.UserID == uuid.Nil || completion.ApplyAttempt <= 0 || completion.Now.IsZero() ||
		!isJSONObject(completion.Result) {
		return messaging.ErrInvalidEmailProposal
	}
	switch completion.Status {
	case messaging.EmailActionProposalApplied:
		if strings.TrimSpace(completion.ErrorMessage) != "" {
			return messaging.ErrInvalidEmailProposal
		}
	case messaging.EmailActionProposalFailed:
		if strings.TrimSpace(completion.ErrorMessage) == "" {
			return messaging.ErrInvalidEmailProposal
		}
	default:
		return messaging.ErrInvalidEmailProposal
	}
	return nil
}

func emailActionProposalMatchesInput(
	record messaging.EmailActionProposalRecord,
	input messaging.EmailActionProposalInput,
) bool {
	return record.ThreadID == input.ThreadID && record.WorkspaceID == input.WorkspaceID &&
		record.UserID == input.UserID && record.SourceMessageID == input.SourceMessageID &&
		record.IdempotencyKey == input.IdempotencyKey && record.ActionKind == input.ActionKind &&
		record.EntityType == input.EntityType && record.EntityID == input.EntityID &&
		record.ExpectedEntityVersion == input.ExpectedEntityVersion &&
		record.ExpiresAt.Equal(input.ExpiresAt.UTC()) && jsonObjectsEqual(record.ProposedDiff, input.ProposedDiff)
}
