package chatsessionsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/domain"
	chatsessionssql "github.com/complexus-tech/projects-api/internal/modules/chatsessions/repository/sqlc"
	apptracing "github.com/complexus-tech/projects-api/pkg/tracing"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CompleteMutationApproval stores the first output for the exact executing
// lease. Repeating completion returns the original durable output.
func (r *repo) CompleteMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams, output json.RawMessage) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.CompleteMutationApproval")
	defer span.End()

	var result chatsessions.CoreMutationApprovalExecution
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		execution, err := queries.CompleteMutationApproval(ctx, chatsessionssql.CompleteMutationApprovalParams{
			Output:      output,
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
			if result.State != chatsessions.MutationApprovalExecutionCompleted && result.State != chatsessions.MutationApprovalExecutionFailed {
				return chatsessions.ErrMutationApprovalLease
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("complete mutation approval: %w", err)
		}
		result, err = toCoreMutationApprovalExecution(execution)
		return err
	})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	return result, nil
}

// FailMutationApproval records that execution may have crossed the mutation
// boundary without a durable result. Completed rows always win this race.
func (r *repo) FailMutationApproval(ctx context.Context, params chatsessions.MutationApprovalExecutionParams, failureCode string) (chatsessions.CoreMutationApprovalExecution, error) {
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.FailMutationApproval")
	defer span.End()

	var result chatsessions.CoreMutationApprovalExecution
	err := r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		execution, err := queries.FailMutationApproval(ctx, chatsessionssql.FailMutationApprovalParams{
			FailureCode: &failureCode,
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
			if result.State != chatsessions.MutationApprovalExecutionCompleted && result.State != chatsessions.MutationApprovalExecutionFailed {
				return chatsessions.ErrMutationApprovalLease
			}
			return nil
		}
		if err != nil {
			return fmt.Errorf("fail mutation approval: %w", err)
		}
		result, err = toCoreMutationApprovalExecution(execution)
		return err
	})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	return result, nil
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
	ctx, span := apptracing.AddSpanFromContext(ctx, "business.repository.chatsessions.ReconcileMutationApproval")
	defer span.End()

	evidence, err := json.Marshal(reconciliation.Evidence)
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, fmt.Errorf("marshal mutation approval reconciliation evidence: %w", err)
	}
	resolution := string(reconciliation.Resolution)

	var result chatsessions.CoreMutationApprovalExecution
	err = r.withinTransaction(ctx, func(queries chatsessionssql.Querier) error {
		var execution chatsessionssql.ChatMutationApprovalExecution
		var queryErr error
		switch reconciliation.Resolution {
		case chatsessions.MutationApprovalReconciliationVerifiedCompleted:
			execution, queryErr = queries.ReconcileMutationApprovalVerifiedCompleted(ctx, chatsessionssql.ReconcileMutationApprovalVerifiedCompletedParams{
				Output:      reconciliation.Output,
				Resolution:  &resolution,
				Evidence:    evidence,
				SessionID:   params.SessionID,
				UserID:      params.UserID,
				WorkspaceID: params.WorkspaceID,
				ToolCallID:  params.ToolCallID,
				Fingerprint: params.Fingerprint,
			})
		case chatsessions.MutationApprovalReconciliationVerifiedNotApplied:
			leaseToken := uuid.New()
			execution, queryErr = queries.ReconcileMutationApprovalVerifiedNotApplied(ctx, chatsessionssql.ReconcileMutationApprovalVerifiedNotAppliedParams{
				LeaseToken:  &leaseToken,
				Resolution:  &resolution,
				Evidence:    evidence,
				SessionID:   params.SessionID,
				UserID:      params.UserID,
				WorkspaceID: params.WorkspaceID,
				ToolCallID:  params.ToolCallID,
				Fingerprint: params.Fingerprint,
			})
		default:
			return errors.New("unsupported mutation approval reconciliation resolution")
		}

		if errors.Is(queryErr, pgx.ErrNoRows) {
			execution, queryErr = lockMutationApproval(ctx, queries, params)
			if queryErr != nil {
				return queryErr
			}
			if execution.Fingerprint != params.Fingerprint {
				return chatsessions.ErrMutationApprovalConflict
			}
			result, queryErr = toCoreMutationApprovalExecution(execution)
			return queryErr
		}
		if queryErr != nil {
			return fmt.Errorf("reconcile mutation approval: %w", queryErr)
		}
		result, queryErr = toCoreMutationApprovalExecution(execution)
		return queryErr
	})
	if err != nil {
		return chatsessions.CoreMutationApprovalExecution{}, err
	}
	return result, nil
}
