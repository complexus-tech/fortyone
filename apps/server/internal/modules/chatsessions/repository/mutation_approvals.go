package chatsessionsrepository

import (
	"context"
	"errors"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	mutationApprovalLeaseExpiredFailure = "execution_lease_expired"
	mutationApprovalSafeRetryResolution = "safe_retry_prepared"
	mutationApprovalWrongOriginFailure  = "retry_requires_original_approval"
)

// ClaimMutationApproval leases a validated prepared call. Only an expired
// ready lease can be reclaimed; an expired executing lease becomes terminally
// uncertain and is never re-executed.
func (r *repo) ClaimMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.ClaimMutationApproval")
	defer span.End()

	var result chatsessions.CoreMutationApprovalExecution
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		leaseToken := uuid.New()
		execution, err := queries.ClaimNewMutationApproval(ctx, chatsessionssql.ClaimNewMutationApprovalParams{
			ToolCallID:  params.ToolCallID,
			Fingerprint: params.Fingerprint,
			LeaseToken:  &leaseToken,
			SessionID:   params.SessionID,
			UserID:      params.UserID,
			WorkspaceID: params.WorkspaceID,
		})
		if err == nil {
			result, err = toClaimedMutationApprovalExecution(execution)
			return err
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("claim mutation approval: %w", err)
		}

		execution, err = lockMutationApproval(ctx, queries, params)
		if errors.Is(err, chatsessions.ErrNotFound) {
			result, err = claimUnresolvedMutationApprovalFingerprint(ctx, queries, params, leaseToken)
			return err
		}
		if err != nil {
			return err
		}
		if execution.Fingerprint != params.Fingerprint {
			return chatsessions.ErrMutationApprovalConflict
		}

		switch execution.Status {
		case "ready":
			current := execution
			execution, err = reclaimExpiredReadyMutationApproval(ctx, queries, params, leaseToken)
			if errors.Is(err, pgx.ErrNoRows) {
				result, err = toCoreMutationApprovalExecution(current)
				return err
			}
			if err != nil {
				return err
			}
			result, err = toClaimedMutationApprovalExecution(execution)
			return err
		case "retry_ready":
			current := execution
			execution, err = claimRetryReadyMutationApproval(ctx, queries, params, leaseToken)
			if errors.Is(err, pgx.ErrNoRows) {
				result, err = toCoreMutationApprovalExecution(current)
				return err
			}
			if err != nil {
				return err
			}
			result, err = toClaimedMutationApprovalExecution(execution)
			return err
		case "executing":
			current := execution
			execution, err = failExpiredExecutingMutationApproval(ctx, queries, params)
			if errors.Is(err, pgx.ErrNoRows) {
				result, err = toCoreMutationApprovalExecution(current)
				return err
			}
			if err != nil {
				return err
			}
			result, err = toCoreMutationApprovalExecution(execution)
			return err
		default:
			result, err = toCoreMutationApprovalExecution(execution)
			return err
		}
	})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	return result, nil
}

// StartMutationApproval crosses the no-retry boundary. A caller may execute a
// tool only after receiving the ephemeral started response for its exact lease.
func (r *repo) StartMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.StartMutationApproval")
	defer span.End()

	var result chatsessions.CoreMutationApprovalExecution
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		execution, err := queries.StartMutationApproval(ctx, chatsessionssql.StartMutationApprovalParams{
			SessionID:   params.SessionID,
			UserID:      params.UserID,
			WorkspaceID: params.WorkspaceID,
			ToolCallID:  params.ToolCallID,
			Fingerprint: params.Fingerprint,
			LeaseToken:  &params.LeaseToken,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			execution, err = lockMutationApproval(ctx, queries, params)
			if err != nil {
				return err
			}
			if execution.Fingerprint != params.Fingerprint {
				return chatsessions.ErrMutationApprovalConflict
			}
			result, err = toCoreMutationApprovalExecution(execution)
			if err != nil {
				return err
			}
			if result.State == chatsessions.MutationApprovalExecutionCompleted || result.State == chatsessions.MutationApprovalExecutionFailed {
				return nil
			}
			return chatsessions.ErrMutationApprovalLease
		}
		if err != nil {
			return fmt.Errorf("start mutation approval: %w", err)
		}
		result = chatsessions.CoreMutationApprovalExecution{State: chatsessions.MutationApprovalExecutionStarted}
		return nil
	})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	return result, nil
}
