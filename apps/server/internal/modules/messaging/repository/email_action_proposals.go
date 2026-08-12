package messagingrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

func (r *Repository) RegisterEmailActionProposal(
	ctx context.Context,
	input messaging.EmailActionProposalInput,
) (messaging.EmailActionProposalRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	input = normalizeEmailActionProposalInput(input)
	if err := validateEmailActionProposalInput(input); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("begin email action proposal registration: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := lockEmailThreadSequence(ctx, tx, input.ThreadID, input.WorkspaceID, input.UserID); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	existing, err := findEmailActionProposalByIdempotencyKey(ctx, tx, input.ThreadID, input.IdempotencyKey)
	if err == nil {
		if !emailActionProposalMatchesInput(existing, input) {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf(
				"%w: proposal idempotency key was reused with different content",
				messaging.ErrInvalidEmailProposal,
			)
		}
		if err := tx.Commit(); err != nil {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal replay: %w", err)
		}
		return existing, false, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE messaging_email_action_proposals
		SET status = CASE WHEN expires_at <= $2 THEN 'expired' ELSE 'superseded' END,
		    expired_at = CASE WHEN expires_at <= $2 THEN $2 ELSE NULL END,
		    superseded_at = CASE WHEN expires_at > $2 THEN $2 ELSE NULL END,
		    updated_at = NOW()
		WHERE thread_id = $1 AND status = 'pending'
	`, input.ThreadID, input.Now.UTC())
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("supersede pending email proposal: %w", err)
	}
	var record messaging.EmailActionProposalRecord
	err = tx.GetContext(ctx, &record, emailActionProposalInsertQuery(),
		input.ThreadID, input.WorkspaceID, input.UserID, input.SourceMessageID,
		input.IdempotencyKey, input.ActionKind, input.EntityType, input.EntityID,
		input.ExpectedEntityVersion, input.ProposedDiff, input.ExpiresAt.UTC())
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("register email action proposal: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email action proposal registration: %w", err)
	}
	return record, true, nil
}

func (r *Repository) ListPendingEmailActionProposals(
	ctx context.Context,
	input messaging.EmailActionProposalListInput,
) ([]messaging.EmailActionProposalRecord, error) {
	if r == nil || r.db == nil {
		return nil, errors.New("messaging repository is not configured")
	}
	if input.ThreadID == uuid.Nil || input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.Now.IsZero() {
		return nil, messaging.ErrInvalidEmailProposal
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("begin pending email proposal lookup: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.ExecContext(ctx, `
		UPDATE messaging_email_action_proposals
		SET status = 'expired', expired_at = $4, updated_at = NOW()
		WHERE thread_id = $1 AND workspace_id = $2 AND user_id = $3
		  AND status = 'pending' AND expires_at <= $4
	`, input.ThreadID, input.WorkspaceID, input.UserID, input.Now.UTC()); err != nil {
		return nil, fmt.Errorf("expire stale email action proposal: %w", err)
	}
	records := make([]messaging.EmailActionProposalRecord, 0, 1)
	if err := tx.SelectContext(ctx, &records, emailActionProposalSelectQuery()+`
		WHERE proposal.thread_id = $1
		  AND proposal.workspace_id = $2
		  AND proposal.user_id = $3
		  AND proposal.status = 'pending'
		ORDER BY proposal.created_at DESC
	`, input.ThreadID, input.WorkspaceID, input.UserID); err != nil {
		return nil, fmt.Errorf("list pending email action proposals: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit pending email proposal lookup: %w", err)
	}
	return records, nil
}

func (r *Repository) FindLatestEmailActionProposalForControl(
	ctx context.Context,
	lookup messaging.EmailActionProposalControlLookup,
) (messaging.EmailActionProposalRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if lookup.ThreadID == uuid.Nil || lookup.WorkspaceID == uuid.Nil || lookup.UserID == uuid.Nil {
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	statusPredicate := ""
	switch lookup.Control {
	case messaging.EmailActionProposalConfirmed:
		statusPredicate = "('pending', 'confirmed', 'applying', 'applied', 'failed')"
	case messaging.EmailActionProposalCancelled:
		statusPredicate = "('pending', 'cancelled')"
	default:
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	var record messaging.EmailActionProposalRecord
	err := r.db.GetContext(ctx, &record, emailActionProposalSelectQuery()+`
		WHERE proposal.thread_id = $1
		  AND proposal.workspace_id = $2
		  AND proposal.user_id = $3
		  AND proposal.status IN `+statusPredicate+`
		ORDER BY proposal.created_at DESC, proposal.id DESC
		LIMIT 1
	`, lookup.ThreadID, lookup.WorkspaceID, lookup.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, false, nil
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("find latest email proposal for control: %w", err)
	}
	return record, true, nil
}

func (r *Repository) GetEmailActionProposal(
	ctx context.Context,
	key messaging.EmailActionProposalKey,
) (messaging.EmailActionProposalRecord, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, errors.New("messaging repository is not configured")
	}
	if key.ProposalID == uuid.Nil || key.ThreadID == uuid.Nil || key.WorkspaceID == uuid.Nil || key.UserID == uuid.Nil {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	var record messaging.EmailActionProposalRecord
	err := r.db.GetContext(ctx, &record, emailActionProposalSelectQuery()+`
		WHERE proposal.id = $1 AND proposal.thread_id = $2
		  AND proposal.workspace_id = $3 AND proposal.user_id = $4
	`, key.ProposalID, key.ThreadID, key.WorkspaceID, key.UserID)
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("get email action proposal: %w", err)
	}
	return record, nil
}

func (r *Repository) DecideEmailActionProposal(
	ctx context.Context,
	decision messaging.EmailActionProposalDecision,
) (messaging.EmailActionProposalRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if err := validateEmailActionProposalDecision(decision); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("begin email proposal decision: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := requireActiveEmailReplyToken(ctx, tx, decision); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	record, err := selectEmailActionProposalForUpdate(ctx, tx, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	now := decision.Now.UTC()
	if record.Status == messaging.EmailActionProposalPending && !now.Before(record.ExpiresAt.UTC()) {
		if err := transitionEmailActionProposalDecision(ctx, tx, record.ID, messaging.EmailActionProposalExpired, now); err != nil {
			return messaging.EmailActionProposalRecord{}, false, err
		}
		record, err = selectEmailActionProposalForUpdate(ctx, tx, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
		if err != nil {
			return messaging.EmailActionProposalRecord{}, false, err
		}
		if err := tx.Commit(); err != nil {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit expired email proposal: %w", err)
		}
		return record, false, messaging.ErrEmailProposalExpired
	}
	if emailProposalDecisionAlreadyWon(record.Status, decision.Decision) {
		if err := tx.Commit(); err != nil {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal decision replay: %w", err)
		}
		return record, true, nil
	}
	if record.Status == messaging.EmailActionProposalExpired {
		return record, false, messaging.ErrEmailProposalExpired
	}
	if record.Status != messaging.EmailActionProposalPending {
		return record, false, fmt.Errorf("%w: proposal is already %s", messaging.ErrEmailProposalConflict, record.Status)
	}
	if err := transitionEmailActionProposalDecision(ctx, tx, record.ID, decision.Decision, now); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	record, err = selectEmailActionProposalForUpdate(ctx, tx, decision.ProposalID, decision.ThreadID, decision.WorkspaceID, decision.UserID)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal decision: %w", err)
	}
	return record, false, nil
}

func (r *Repository) ClaimEmailActionProposalApply(
	ctx context.Context,
	claim messaging.EmailActionProposalApplyClaim,
) (messaging.EmailActionProposalRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	if claim.ProposalID == uuid.Nil || claim.ThreadID == uuid.Nil || claim.WorkspaceID == uuid.Nil ||
		claim.UserID == uuid.Nil || claim.Now.IsZero() {
		return messaging.EmailActionProposalRecord{}, false, messaging.ErrInvalidEmailProposal
	}
	if claim.RetryAfter <= 0 {
		claim.RetryAfter = messagingLeaseDuration
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("begin email proposal apply claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	record, err := selectEmailActionProposalForUpdate(ctx, tx, claim.ProposalID, claim.ThreadID, claim.WorkspaceID, claim.UserID)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	now := claim.Now.UTC()
	switch record.Status {
	case messaging.EmailActionProposalApplied:
		if err := tx.Commit(); err != nil {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit applied email proposal replay: %w", err)
		}
		return record, false, nil
	case messaging.EmailActionProposalConfirmed, messaging.EmailActionProposalFailed:
	case messaging.EmailActionProposalApplying:
		if record.ApplyingAt != nil && now.Sub(record.ApplyingAt.UTC()) < claim.RetryAfter {
			return record, false, messaging.ErrEmailProposalApplyBusy
		}
	default:
		return record, false, fmt.Errorf(
			"%w: proposal in status %s cannot be applied",
			messaging.ErrEmailProposalConflict,
			record.Status,
		)
	}
	var claimed messaging.EmailActionProposalRecord
	err = tx.GetContext(ctx, &claimed, emailActionProposalUpdateReturningQuery(`
		SET status = 'applying',
		    apply_attempt = apply_attempt + 1,
		    applying_at = $2,
		    failed_at = NULL,
		    result = NULL,
		    last_error = NULL,
		    updated_at = NOW()
		WHERE id = $1
	`), record.ID, now)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("claim email proposal apply: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal apply claim: %w", err)
	}
	return claimed, true, nil
}

func (r *Repository) CompleteEmailActionProposalApply(
	ctx context.Context,
	completion messaging.EmailActionProposalApplyCompletion,
) (messaging.EmailActionProposalRecord, bool, error) {
	if r == nil || r.db == nil {
		return messaging.EmailActionProposalRecord{}, false, errors.New("messaging repository is not configured")
	}
	completion.Result = normalizeJSONObject(completion.Result)
	if err := validateEmailActionProposalApplyCompletion(completion); err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("begin email proposal apply completion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	record, err := selectEmailActionProposalForUpdate(ctx, tx, completion.ProposalID, completion.ThreadID, completion.WorkspaceID, completion.UserID)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, err
	}
	if record.Status == completion.Status && record.ApplyAttempt == completion.ApplyAttempt {
		if err := tx.Commit(); err != nil {
			return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal completion replay: %w", err)
		}
		return record, true, nil
	}
	if record.Status != messaging.EmailActionProposalApplying || record.ApplyAttempt != completion.ApplyAttempt {
		return record, false, fmt.Errorf(
			"%w: apply attempt %d is no longer active",
			messaging.ErrEmailProposalConflict,
			completion.ApplyAttempt,
		)
	}
	now := completion.Now.UTC()
	setClause := `
		SET status = 'applied', result = $3, last_error = NULL,
		    applied_at = $2, updated_at = NOW()
		WHERE id = $1 AND apply_attempt = $4
	`
	lastError := ""
	if completion.Status == messaging.EmailActionProposalFailed {
		setClause = `
			SET status = 'failed', result = $3, last_error = NULLIF($5, ''),
			    failed_at = $2, updated_at = NOW()
			WHERE id = $1 AND apply_attempt = $4
		`
		lastError = truncateEmailError(completion.ErrorMessage)
	}
	var completed messaging.EmailActionProposalRecord
	args := []any{record.ID, now, completion.Result, completion.ApplyAttempt}
	if completion.Status == messaging.EmailActionProposalFailed {
		args = append(args, lastError)
	}
	err = tx.GetContext(ctx, &completed, emailActionProposalUpdateReturningQuery(setClause), args...)
	if err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("complete email proposal apply: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return messaging.EmailActionProposalRecord{}, false, fmt.Errorf("commit email proposal apply completion: %w", err)
	}
	return completed, false, nil
}

func requireActiveEmailReplyToken(
	ctx context.Context,
	tx *sqlx.Tx,
	decision messaging.EmailActionProposalDecision,
) error {
	var exists bool
	err := tx.GetContext(ctx, &exists, `
		SELECT true
		FROM messaging_email_reply_tokens token
		INNER JOIN messaging_email_threads thread
			ON thread.id = token.thread_id
			AND thread.workspace_id = token.workspace_id
			AND thread.user_id = token.user_id
		INNER JOIN workspaces workspace ON workspace.workspace_id = thread.workspace_id
		INNER JOIN users actor ON actor.user_id = thread.user_id
		INNER JOIN workspace_members member
			ON member.workspace_id = thread.workspace_id
			AND member.user_id = thread.user_id
		WHERE token.thread_id = $1
		  AND token.workspace_id = $2
		  AND token.user_id = $3
		  AND token.token_hash = $4
		  AND token.revoked_at IS NULL
		  AND token.expires_at > $5
		  AND workspace.deleted_at IS NULL
		  AND actor.is_active = true
		  AND actor.is_system = false
		  AND lower(actor.email) = lower(thread.recipient_email)
		  AND member.role IN ('admin', 'member', 'guest')
	`, decision.ThreadID, decision.WorkspaceID, decision.UserID,
		decision.ReplyTokenHash, decision.Now.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.ErrInvalidEmailReplyToken
	}
	if err != nil {
		return fmt.Errorf("verify email proposal reply token: %w", err)
	}
	return nil
}

func transitionEmailActionProposalDecision(
	ctx context.Context,
	tx *sqlx.Tx,
	proposalID uuid.UUID,
	status messaging.EmailActionProposalStatus,
	now time.Time,
) error {
	column := ""
	switch status {
	case messaging.EmailActionProposalConfirmed:
		column = "confirmed_at"
	case messaging.EmailActionProposalCancelled:
		column = "cancelled_at"
	case messaging.EmailActionProposalExpired:
		column = "expired_at"
	case messaging.EmailActionProposalSuperseded:
		column = "superseded_at"
	default:
		return messaging.ErrInvalidEmailProposal
	}
	query := fmt.Sprintf(`
		UPDATE messaging_email_action_proposals
		SET status = $2, %s = $3, updated_at = NOW()
		WHERE id = $1 AND status = 'pending'
	`, column)
	result, err := tx.ExecContext(ctx, query, proposalID, status, now)
	if err != nil {
		return fmt.Errorf("transition email action proposal: %w", err)
	}
	return requireAffectedRow(result, "transition email action proposal")
}

func emailProposalDecisionAlreadyWon(
	status, decision messaging.EmailActionProposalStatus,
) bool {
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

func selectEmailActionProposalForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	proposalID, threadID, workspaceID, userID uuid.UUID,
) (messaging.EmailActionProposalRecord, error) {
	var record messaging.EmailActionProposalRecord
	err := tx.GetContext(ctx, &record, emailActionProposalSelectQuery()+`
		WHERE proposal.id = $1 AND proposal.thread_id = $2
		  AND proposal.workspace_id = $3 AND proposal.user_id = $4
		FOR UPDATE
	`, proposalID, threadID, workspaceID, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, messaging.ErrInvalidEmailProposal
	}
	if err != nil {
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("lock email action proposal: %w", err)
	}
	return record, nil
}

func findEmailActionProposalByIdempotencyKey(
	ctx context.Context,
	db sqlx.ExtContext,
	threadID uuid.UUID,
	idempotencyKey string,
) (messaging.EmailActionProposalRecord, error) {
	var record messaging.EmailActionProposalRecord
	err := sqlx.GetContext(ctx, db, &record, emailActionProposalSelectQuery()+`
		WHERE proposal.thread_id = $1 AND proposal.idempotency_key = $2
	`, threadID, idempotencyKey)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return messaging.EmailActionProposalRecord{}, fmt.Errorf("read email action proposal: %w", err)
	}
	return record, err
}

func emailActionProposalInsertQuery() string {
	return `
		INSERT INTO messaging_email_action_proposals (
			thread_id, workspace_id, user_id, source_message_id, idempotency_key,
			action_kind, entity_type, entity_id, expected_entity_version,
			proposed_diff, status, expires_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 'pending', $11)
		RETURNING ` + emailActionProposalColumns()
}

func emailActionProposalSelectQuery() string {
	return "SELECT " + emailActionProposalAliasedColumns() + " FROM messaging_email_action_proposals proposal "
}

func emailActionProposalUpdateReturningQuery(setClause string) string {
	return "UPDATE messaging_email_action_proposals " + setClause + " RETURNING " + emailActionProposalColumns()
}

func emailActionProposalColumns() string {
	return `id, thread_id, workspace_id, user_id, source_message_id,
	       idempotency_key, action_kind, entity_type, entity_id,
	       expected_entity_version, proposed_diff, status, apply_attempt,
	       COALESCE(result, CAST('{}' AS jsonb)) AS result, last_error,
	       expires_at, confirmed_at, applying_at, applied_at, failed_at,
	       cancelled_at, expired_at, superseded_at, created_at, updated_at`
}

func emailActionProposalAliasedColumns() string {
	return `proposal.id, proposal.thread_id, proposal.workspace_id, proposal.user_id,
	       proposal.source_message_id, proposal.idempotency_key,
	       proposal.action_kind, proposal.entity_type, proposal.entity_id,
	       proposal.expected_entity_version, proposal.proposed_diff,
	       proposal.status, proposal.apply_attempt,
	       COALESCE(proposal.result, CAST('{}' AS jsonb)) AS result,
	       proposal.last_error, proposal.expires_at, proposal.confirmed_at,
	       proposal.applying_at, proposal.applied_at, proposal.failed_at,
	       proposal.cancelled_at, proposal.expired_at, proposal.superseded_at,
	       proposal.created_at, proposal.updated_at`
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
		input.SourceMessageID == uuid.Nil || input.EntityID == uuid.Nil ||
		input.IdempotencyKey == "" || input.ActionKind == "" || input.EntityType == "" ||
		input.ExpectedEntityVersion == "" || input.ExpiresAt.IsZero() || input.Now.IsZero() ||
		!input.ExpiresAt.After(input.Now) || !isJSONObject(input.ProposedDiff) {
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
