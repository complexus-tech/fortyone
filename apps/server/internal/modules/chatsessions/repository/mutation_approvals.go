package chatsessionsrepository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

const (
	mutationApprovalReadyLeaseInterval     = "30 seconds"
	mutationApprovalExecutionLeaseInterval = "3 minutes"
	mutationApprovalLeaseExpiredFailure    = "execution_lease_expired"
	mutationApprovalSafeRetryResolution    = "safe_retry_prepared"
	mutationApprovalWrongOriginFailure     = "retry_requires_original_approval"
)

type dbMutationApprovalExecution struct {
	SessionID   string `db:"session_id"`
	ToolCallID  string `db:"tool_call_id"`
	Fingerprint string `db:"fingerprint"`
	Status      string `db:"status"`
	// Keep the database representation nullable. database/sql cannot scan a
	// NULL JSONB value into json.RawMessage, while []byte correctly receives
	// nil for prepared and executing rows.
	Output         []byte         `db:"output"`
	LeaseToken     uuid.NullUUID  `db:"lease_token"`
	LeaseExpiresAt sql.NullTime   `db:"lease_expires_at"`
	FailureCode    sql.NullString `db:"failure_code"`
}

const mutationApprovalExecutionColumns = `
	execution.session_id,
	execution.tool_call_id,
	execution.fingerprint,
	execution.status,
	execution.output,
	execution.lease_token,
	execution.lease_expires_at,
	execution.failure_code
`

const mutationApprovalExecutionSelect = `
	SELECT ` + mutationApprovalExecutionColumns + `
	FROM chat_mutation_approval_executions AS execution
	JOIN chat_sessions AS session ON session.id = execution.session_id
	WHERE execution.session_id = $1
		AND execution.user_id = $2
		AND execution.workspace_id = $3
		AND execution.tool_call_id = $4
		AND session.user_id = $2
		AND session.workspace_id = $3
		AND session.deleted_at IS NULL
`

// ClaimMutationApproval leases a validated prepared call. Only an expired
// ready lease can be reclaimed; an expired executing lease becomes terminally
// uncertain and is never re-executed.
func (r *repo) ClaimMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.ClaimMutationApproval")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("begin mutation approval claim: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	leaseToken := uuid.New()
	const insertQuery = `
		INSERT INTO chat_mutation_approval_executions (
			session_id,
			user_id,
			workspace_id,
			tool_call_id,
			fingerprint,
			status,
			lease_token,
			lease_expires_at,
			attempt_count
		)
		SELECT session.id, session.user_id, session.workspace_id, $4, $5,
		       'ready', $6, CURRENT_TIMESTAMP + CAST($7 AS interval), 1
		FROM chat_sessions AS session
		WHERE session.id = $1
			AND session.user_id = $2
			AND session.workspace_id = $3
			AND session.deleted_at IS NULL
		ON CONFLICT DO NOTHING
		RETURNING tool_call_id, fingerprint, status, output, lease_token, lease_expires_at, failure_code
	`

	var execution dbMutationApprovalExecution
	err = tx.GetContext(
		ctx,
		&execution,
		insertQuery,
		params.SessionID,
		params.UserID,
		params.WorkspaceID,
		params.ToolCallID,
		params.Fingerprint,
		leaseToken,
		mutationApprovalReadyLeaseInterval,
	)
	if err == nil {
		result, convertErr := toClaimedMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit mutation approval claim: %w", commitErr)
		}
		return result, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("claim mutation approval: %w", err)
	}

	execution, err = getMutationApprovalExecutionForUpdate(ctx, tx, params)
	if errors.Is(err, chatsessions.ErrNotFound) {
		return claimUnresolvedMutationApprovalFingerprint(ctx, tx, params, leaseToken)
	}
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	if execution.Fingerprint != params.Fingerprint {
		return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalConflict
	}

	switch execution.Status {
	case "ready":
		currentExecution := execution
		execution, err = reclaimExpiredReadyMutationApproval(ctx, tx, params, leaseToken)
		if errors.Is(err, sql.ErrNoRows) {
			return commitMutationApprovalState(tx, currentExecution)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		result, convertErr := toClaimedMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit reclaimed mutation approval: %w", commitErr)
		}
		return result, nil
	case "retry_ready":
		currentExecution := execution
		execution, err = claimRetryReadyMutationApproval(ctx, tx, params, leaseToken)
		if errors.Is(err, sql.ErrNoRows) {
			return commitMutationApprovalState(tx, currentExecution)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		result, convertErr := toClaimedMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit safe retry mutation approval claim: %w", commitErr)
		}
		return result, nil
	case "executing":
		currentExecution := execution
		execution, err = failExpiredExecutingMutationApproval(ctx, tx, params)
		if errors.Is(err, sql.ErrNoRows) {
			return commitMutationApprovalState(tx, currentExecution)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		return commitMutationApprovalState(tx, execution)
	default:
		return commitMutationApprovalState(tx, execution)
	}
}

// StartMutationApproval crosses the no-retry boundary. A caller may execute a
// tool only after receiving the ephemeral started response for its exact lease.
func (r *repo) StartMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.StartMutationApproval")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("begin mutation approval start: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const startQuery = `
		UPDATE chat_mutation_approval_executions AS execution
		SET status = 'executing',
			started_at = CURRENT_TIMESTAMP,
			failed_at = NULL,
			failure_code = NULL,
			lease_expires_at = CURRENT_TIMESTAMP + CAST($7 AS interval),
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.lease_token = $6
			AND execution.status IN ('ready', 'retry_ready')
			AND execution.lease_expires_at > CURRENT_TIMESTAMP
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
			AND session.deleted_at IS NULL
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err = tx.GetContext(
		ctx,
		&execution,
		startQuery,
		params.SessionID,
		params.UserID,
		params.WorkspaceID,
		params.ToolCallID,
		params.Fingerprint,
		params.LeaseToken,
		mutationApprovalExecutionLeaseInterval,
	)
	if errors.Is(err, sql.ErrNoRows) {
		execution, err = getMutationApprovalExecutionForUpdate(ctx, tx, params)
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		if execution.Fingerprint != params.Fingerprint {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalConflict
		}
		result, convertErr := toCoreMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if result.State == chatsessions.MutationApprovalExecutionCompleted || result.State == chatsessions.MutationApprovalExecutionFailed {
			if commitErr := tx.Commit(); commitErr != nil {
				return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit terminal mutation approval start lookup: %w", commitErr)
			}
			return result, nil
		}
		return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalLease
	}
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("start mutation approval: %w", err)
	}
	if commitErr := tx.Commit(); commitErr != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit mutation approval start: %w", commitErr)
	}
	return chatsessions.CoreMutationApprovalExecution{State: chatsessions.MutationApprovalExecutionStarted}, nil
}

// CompleteMutationApproval stores the first output for the exact executing
// lease. Repeating completion returns the original durable output.
func (r *repo) CompleteMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams, output json.RawMessage) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.CompleteMutationApproval")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("begin mutation approval completion: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const completeQuery = `
		UPDATE chat_mutation_approval_executions AS execution
		SET status = 'completed',
			output = CAST($7 AS jsonb),
			completed_at = CURRENT_TIMESTAMP,
			lease_token = NULL,
			lease_expires_at = NULL,
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.lease_token = $6
			AND execution.status = 'executing'
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err = tx.GetContext(
		ctx,
		&execution,
		completeQuery,
		params.SessionID,
		params.UserID,
		params.WorkspaceID,
		params.ToolCallID,
		params.Fingerprint,
		params.LeaseToken,
		string(output),
	)
	if errors.Is(err, sql.ErrNoRows) {
		execution, err = getMutationApprovalExecutionForUpdate(ctx, tx, params)
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		if execution.Fingerprint != params.Fingerprint {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalConflict
		}
		result, convertErr := toCoreMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if result.State != chatsessions.MutationApprovalExecutionCompleted && result.State != chatsessions.MutationApprovalExecutionFailed {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalLease
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit mutation approval completion lookup: %w", commitErr)
		}
		return result, nil
	}
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("complete mutation approval: %w", err)
	}

	return commitMutationApprovalState(tx, execution)
}

// FailMutationApproval records that execution may have crossed the mutation
// boundary without a durable result. Completed rows always win this race.
func (r *repo) FailMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams, failureCode string) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.FailMutationApproval")
	defer span.End()

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("begin mutation approval failure: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	const failQuery = `
		UPDATE chat_mutation_approval_executions AS execution
		SET status = 'failed_uncertain',
			lease_token = NULL,
			lease_expires_at = NULL,
			failed_at = CURRENT_TIMESTAMP,
			failure_code = $7,
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.lease_token = $6
			AND (
				execution.status = 'executing'
				OR (
					execution.status IN ('ready', 'retry_ready')
					AND $7 = 'start_transition_uncertain'
				)
			)
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err = tx.GetContext(
		ctx,
		&execution,
		failQuery,
		params.SessionID,
		params.UserID,
		params.WorkspaceID,
		params.ToolCallID,
		params.Fingerprint,
		params.LeaseToken,
		failureCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		execution, err = getMutationApprovalExecutionForUpdate(ctx, tx, params)
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		if execution.Fingerprint != params.Fingerprint {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalConflict
		}
		result, convertErr := toCoreMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if result.State != chatsessions.MutationApprovalExecutionCompleted && result.State != chatsessions.MutationApprovalExecutionFailed {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalLease
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit mutation approval failure lookup: %w", commitErr)
		}
		return result, nil
	}
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("fail mutation approval: %w", err)
	}

	return commitMutationApprovalState(tx, execution)
}

// ReconcileMutationApproval applies an explicit, independently verified
// resolution to a terminally uncertain execution. Verified-not-applied resets
// the row with an already-expired ready lease so a later approval must claim a
// fresh lease before any tool can run.
func (r *repo) ReconcileMutationApproval(
	ctx context.Context,
	params chatsessions.MutationApprovalExecutionParams,
	reconciliation chatsessions.MutationApprovalReconciliation,
) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.chatsessions.ReconcileMutationApproval")
	defer span.End()

	evidence, err := json.Marshal(reconciliation.Evidence)
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("marshal mutation approval reconciliation evidence: %w", err)
	}

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("begin mutation approval reconciliation: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var execution dbMutationApprovalExecution
	switch reconciliation.Resolution {
	case chatsessions.MutationApprovalReconciliationVerifiedCompleted:
		const query = `
			UPDATE chat_mutation_approval_executions AS execution
			SET status = 'completed',
				output = CAST($8 AS jsonb),
				completed_at = CURRENT_TIMESTAMP,
				lease_token = NULL,
				lease_expires_at = NULL,
				failed_at = NULL,
				failure_code = NULL,
				last_reconciliation_resolution = $6,
				last_reconciliation_evidence = CAST($7 AS jsonb),
				last_reconciled_at = CURRENT_TIMESTAMP,
				reconciliation_count = reconciliation_count + 1,
				updated_at = CURRENT_TIMESTAMP
			FROM chat_sessions AS session
			WHERE execution.session_id = $1
				AND execution.user_id = $2
				AND execution.workspace_id = $3
				AND execution.tool_call_id = $4
				AND execution.fingerprint = $5
				AND execution.status = 'failed_uncertain'
				AND session.id = execution.session_id
				AND session.user_id = $2
				AND session.workspace_id = $3
			RETURNING execution.tool_call_id, execution.fingerprint, execution.status,
			          execution.output, execution.lease_token,
			          execution.lease_expires_at, execution.failure_code
		`
		err = tx.GetContext(ctx, &execution, query,
			params.SessionID, params.UserID, params.WorkspaceID,
			params.ToolCallID, params.Fingerprint,
			reconciliation.Resolution, string(evidence), string(reconciliation.Output),
		)
	case chatsessions.MutationApprovalReconciliationVerifiedNotApplied:
		const query = `
			UPDATE chat_mutation_approval_executions AS execution
			SET status = 'ready',
				output = NULL,
				completed_at = NULL,
				lease_token = $8,
				lease_expires_at = CURRENT_TIMESTAMP,
				started_at = NULL,
				failed_at = NULL,
				failure_code = NULL,
				attempt_count = attempt_count + 1,
				last_reconciliation_resolution = $6,
				last_reconciliation_evidence = CAST($7 AS jsonb),
				last_reconciled_at = CURRENT_TIMESTAMP,
				reconciliation_count = reconciliation_count + 1,
				updated_at = CURRENT_TIMESTAMP
			FROM chat_sessions AS session
			WHERE execution.session_id = $1
				AND execution.user_id = $2
				AND execution.workspace_id = $3
				AND execution.tool_call_id = $4
				AND execution.fingerprint = $5
				AND execution.status = 'failed_uncertain'
				AND session.id = execution.session_id
				AND session.user_id = $2
				AND session.workspace_id = $3
			RETURNING execution.tool_call_id, execution.fingerprint, execution.status,
			          execution.output, execution.lease_token,
			          execution.lease_expires_at, execution.failure_code
		`
		err = tx.GetContext(ctx, &execution, query,
			params.SessionID, params.UserID, params.WorkspaceID,
			params.ToolCallID, params.Fingerprint,
			reconciliation.Resolution, string(evidence), uuid.New(),
		)
	default:
		return chatsessions.CoreMutationApprovalExecution{}, errors.New("unsupported mutation approval reconciliation resolution")
	}

	if errors.Is(err, sql.ErrNoRows) {
		execution, err = getMutationApprovalExecutionForUpdate(ctx, tx, params)
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		if execution.Fingerprint != params.Fingerprint {
			return chatsessions.CoreMutationApprovalExecution{}, chatsessions.ErrMutationApprovalConflict
		}
		return commitMutationApprovalState(tx, execution)
	}
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("reconcile mutation approval: %w", err)
	}

	return commitMutationApprovalState(tx, execution)
}

func reclaimExpiredReadyMutationApproval(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (dbMutationApprovalExecution, error) {
	const query = `
		UPDATE chat_mutation_approval_executions AS execution
		SET lease_token = $6,
			lease_expires_at = CURRENT_TIMESTAMP + CAST($7 AS interval),
			attempt_count = attempt_count + 1,
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.status = 'ready'
			AND execution.lease_expires_at <= CURRENT_TIMESTAMP
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err := tx.GetContext(ctx, &execution, query,
		params.SessionID, params.UserID, params.WorkspaceID,
		params.ToolCallID, params.Fingerprint, leaseToken,
		mutationApprovalReadyLeaseInterval,
	)
	return execution, err
}

func claimRetryReadyMutationApproval(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (dbMutationApprovalExecution, error) {
	const query = `
		UPDATE chat_mutation_approval_executions AS execution
		SET lease_token = $6,
			lease_expires_at = CURRENT_TIMESTAMP + CAST($7 AS interval),
			attempt_count = attempt_count + 1,
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.status = 'retry_ready'
			AND (
				execution.lease_token IS NULL
				OR execution.lease_expires_at <= CURRENT_TIMESTAMP
			)
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
			AND session.deleted_at IS NULL
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err := tx.GetContext(ctx, &execution, query,
		params.SessionID, params.UserID, params.WorkspaceID,
		params.ToolCallID, params.Fingerprint, leaseToken,
		mutationApprovalReadyLeaseInterval,
	)
	return execution, err
}

func failExpiredExecutingMutationApproval(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
) (dbMutationApprovalExecution, error) {
	const query = `
		UPDATE chat_mutation_approval_executions AS execution
		SET status = 'failed_uncertain',
			lease_token = NULL,
			lease_expires_at = NULL,
			failed_at = CURRENT_TIMESTAMP,
			failure_code = $6,
			updated_at = CURRENT_TIMESTAMP
		FROM chat_sessions AS session
		WHERE execution.session_id = $1
			AND execution.user_id = $2
			AND execution.workspace_id = $3
			AND execution.tool_call_id = $4
			AND execution.fingerprint = $5
			AND execution.status = 'executing'
			AND execution.lease_expires_at <= CURRENT_TIMESTAMP
			AND session.id = execution.session_id
			AND session.user_id = $2
			AND session.workspace_id = $3
		RETURNING execution.tool_call_id, execution.fingerprint, execution.status, execution.output,
		          execution.lease_token, execution.lease_expires_at,
		          execution.failure_code
	`

	var execution dbMutationApprovalExecution
	err := tx.GetContext(ctx, &execution, query,
		params.SessionID, params.UserID, params.WorkspaceID,
		params.ToolCallID, params.Fingerprint,
		mutationApprovalLeaseExpiredFailure,
	)
	return execution, err
}

func getMutationApprovalExecutionForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
) (dbMutationApprovalExecution, error) {
	var execution dbMutationApprovalExecution
	if err := tx.GetContext(
		ctx,
		&execution,
		mutationApprovalExecutionSelect+" FOR UPDATE OF execution",
		params.SessionID,
		params.UserID,
		params.WorkspaceID,
		params.ToolCallID,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbMutationApprovalExecution{}, chatsessions.ErrNotFound
		}
		return dbMutationApprovalExecution{}, fmt.Errorf("get mutation approval execution: %w", err)
	}
	return execution, nil
}

func claimUnresolvedMutationApprovalFingerprint(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
	leaseToken uuid.UUID,
) (chatsessions.CoreMutationApprovalExecution, error) {
	execution, err := getUnresolvedMutationApprovalFingerprintForUpdate(ctx, tx, params)
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}

	switch execution.Status {
	case "ready":
		// A ready row has never crossed Start. Preserve its identity as a
		// terminal known-not-run tombstone, then create a separate lease for the
		// newer approval. Rewriting the old identity would let a stale browser
		// replay it after the replacement completes.
		currentExecution := execution
		execution, err = terminalizeExpiredReadyAndClaimReplacement(
			ctx,
			tx,
			params,
			execution.SessionID,
			execution.ToolCallID,
			leaseToken,
		)
		if errors.Is(err, sql.ErrNoRows) {
			return commitMutationApprovalState(tx, currentExecution)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("replace expired ready mutation approval: %w", err)
		}
		result, convertErr := toClaimedMutationApprovalExecution(execution)
		if convertErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, convertErr
		}
		if commitErr := tx.Commit(); commitErr != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit transferred mutation approval: %w", commitErr)
		}
		return result, nil
	case "executing":
		currentExecution := execution
		existingParams := params
		existingParams.SessionID = execution.SessionID
		existingParams.ToolCallID = execution.ToolCallID
		execution, err = failExpiredExecutingMutationApproval(ctx, tx, existingParams)
		if errors.Is(err, sql.ErrNoRows) {
			return commitMutationApprovalState(tx, currentExecution)
		}
		if err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, err
		}
		return commitMutationApprovalState(tx, execution)
	case "retry_ready":
		if err := tx.Commit(); err != nil {
			return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit cross-origin safe retry quarantine: %w", err)
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:       chatsessions.MutationApprovalExecutionFailed,
			FailureCode: mutationApprovalWrongOriginFailure,
		}, nil
	default:
		return commitMutationApprovalState(tx, execution)
	}
}

func getUnresolvedMutationApprovalFingerprintForUpdate(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
) (dbMutationApprovalExecution, error) {
	// Quarantine follows the user and workspace rather than one chat. This
	// prevents a new conversation from replaying an identical mutation whose
	// outcome is still uncertain.
	const query = `
		SELECT ` + mutationApprovalExecutionColumns + `
		FROM chat_mutation_approval_executions AS execution
		JOIN chat_sessions AS session ON session.id = execution.session_id
		WHERE execution.user_id = $2
			AND execution.workspace_id = $3
			AND (
				execution.session_id <> $1
				OR execution.tool_call_id <> $4
			)
			AND execution.fingerprint = $5
			AND execution.status IN ('ready', 'retry_ready', 'executing', 'failed_uncertain')
			AND session.user_id = $2
			AND session.workspace_id = $3
		FOR UPDATE OF execution
	`

	var execution dbMutationApprovalExecution
	if err := tx.GetContext(ctx, &execution, query,
		params.SessionID, params.UserID, params.WorkspaceID,
		params.ToolCallID, params.Fingerprint,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dbMutationApprovalExecution{}, chatsessions.ErrNotFound
		}
		return dbMutationApprovalExecution{}, fmt.Errorf("get unresolved mutation approval fingerprint: %w", err)
	}
	return execution, nil
}

func terminalizeExpiredReadyAndClaimReplacement(
	ctx context.Context,
	tx *sqlx.Tx,
	params chatsessions.MutationApprovalExecutionParams,
	previousSessionID string,
	previousToolCallID string,
	leaseToken uuid.UUID,
) (dbMutationApprovalExecution, error) {
	expiredOutput, err := json.Marshal(map[string]any{
		"error":   expiredApprovalOutputMessage,
		"success": false,
	})
	if err != nil {
		return dbMutationApprovalExecution{}, fmt.Errorf("marshal expired approval output: %w", err)
	}

	const query = `
		WITH terminalized AS (
			UPDATE chat_mutation_approval_executions AS execution
			SET status = 'completed',
				output = CAST($10 AS jsonb),
				completed_at = CURRENT_TIMESTAMP,
				lease_token = NULL,
				lease_expires_at = NULL,
				failed_at = NULL,
				failure_code = NULL,
				updated_at = CURRENT_TIMESTAMP
			WHERE execution.session_id = $8
				AND execution.user_id = $2
				AND execution.workspace_id = $3
				AND execution.tool_call_id = $9
				AND execution.fingerprint = $5
				AND execution.status = 'ready'
				AND execution.lease_expires_at <= CURRENT_TIMESTAMP
			RETURNING 1
		)
		INSERT INTO chat_mutation_approval_executions (
			session_id,
			user_id,
			workspace_id,
			tool_call_id,
			fingerprint,
			status,
			lease_token,
			lease_expires_at,
			attempt_count
		)
		SELECT destination_session.id, destination_session.user_id,
		       destination_session.workspace_id, $4, $5, 'ready', $6,
		       CURRENT_TIMESTAMP + CAST($7 AS interval), 1
		FROM chat_sessions AS destination_session
		CROSS JOIN terminalized
		WHERE destination_session.id = $1
			AND destination_session.user_id = $2
			AND destination_session.workspace_id = $3
			AND destination_session.deleted_at IS NULL
		RETURNING session_id, tool_call_id, fingerprint, status, output,
		          lease_token, lease_expires_at, failure_code
	`

	var execution dbMutationApprovalExecution
	err = tx.GetContext(ctx, &execution, query,
		params.SessionID, params.UserID, params.WorkspaceID,
		params.ToolCallID, params.Fingerprint, leaseToken,
		mutationApprovalReadyLeaseInterval, previousSessionID, previousToolCallID,
		string(expiredOutput),
	)
	return execution, err
}

func commitMutationApprovalState(
	tx *sqlx.Tx,
	execution dbMutationApprovalExecution,
) (chatsessions.CoreMutationApprovalExecution, error) {
	result, err := toCoreMutationApprovalExecution(execution)
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	if err := tx.Commit(); err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("commit mutation approval state: %w", err)
	}
	return result, nil
}

func toClaimedMutationApprovalExecution(execution dbMutationApprovalExecution) (chatsessions.CoreMutationApprovalExecution, error) {
	if (execution.Status != "ready" && execution.Status != "retry_ready") || !execution.LeaseToken.Valid || !execution.LeaseExpiresAt.Valid {
		return chatsessions.CoreMutationApprovalExecution{}, errors.New("claimed mutation approval has an invalid lease")
	}
	leaseToken := execution.LeaseToken.UUID
	leaseExpiresAt := execution.LeaseExpiresAt.Time
	return chatsessions.CoreMutationApprovalExecution{
		State:          chatsessions.MutationApprovalExecutionClaimed,
		LeaseToken:     &leaseToken,
		LeaseExpiresAt: &leaseExpiresAt,
	}, nil
}

func toCoreMutationApprovalExecution(execution dbMutationApprovalExecution) (chatsessions.CoreMutationApprovalExecution, error) {
	switch execution.Status {
	case "ready":
		if !execution.LeaseExpiresAt.Valid {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("ready mutation approval has no lease expiry")
		}
		leaseExpiresAt := execution.LeaseExpiresAt.Time
		return chatsessions.CoreMutationApprovalExecution{
			State:          chatsessions.MutationApprovalExecutionReady,
			LeaseExpiresAt: &leaseExpiresAt,
		}, nil
	case "retry_ready":
		if !execution.LeaseExpiresAt.Valid {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("safe retry mutation approval has no lease expiry")
		}
		leaseExpiresAt := execution.LeaseExpiresAt.Time
		return chatsessions.CoreMutationApprovalExecution{
			State:          chatsessions.MutationApprovalExecutionReady,
			LeaseExpiresAt: &leaseExpiresAt,
		}, nil
	case "executing":
		if !execution.LeaseExpiresAt.Valid {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("executing mutation approval has no lease expiry")
		}
		leaseExpiresAt := execution.LeaseExpiresAt.Time
		return chatsessions.CoreMutationApprovalExecution{
			State:          chatsessions.MutationApprovalExecutionExecuting,
			LeaseExpiresAt: &leaseExpiresAt,
		}, nil
	case "completed":
		if !json.Valid(execution.Output) {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("completed mutation approval has invalid output")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:  chatsessions.MutationApprovalExecutionCompleted,
			Output: append(json.RawMessage(nil), execution.Output...),
		}, nil
	case "failed_uncertain":
		if !execution.FailureCode.Valid {
			return chatsessions.CoreMutationApprovalExecution{}, errors.New("uncertain mutation approval has no failure code")
		}
		return chatsessions.CoreMutationApprovalExecution{
			State:       chatsessions.MutationApprovalExecutionFailed,
			FailureCode: execution.FailureCode.String,
		}, nil
	default:
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("unknown mutation approval status %q", execution.Status)
	}
}
