package agentinvocationsrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
)

// BeginAction commits the external command and its invocation stage together.
// The invocation lock serializes cancellation with the authoritative consumer.
func (r *Repository) BeginAction(ctx context.Context, run domain.Run, begin func(context.Context, domain.Run) error) (domain.Run, error) {
	if begin == nil {
		return domain.Run{}, domain.ErrInvalidInput
	}
	return r.withLease(ctx, run, func(txCtx context.Context, queries *invocationsql.Queries, current domain.Run) (domain.Run, error) {
		if current.ActionStartedAt != nil {
			return current, nil
		}
		if current.Proposal == nil || !current.Proposal.HasAction() || current.Outcome != nil || (current.Proposal.RequiresApproval && current.ApprovedAt == nil) {
			return domain.Run{}, domain.ErrConflict
		}
		if err := begin(txCtx, current); err != nil {
			return domain.Run{}, err
		}
		row, err := queries.StartAgentInvocationAction(txCtx, invocationsql.StartAgentInvocationActionParams{
			ID: current.ID, LeaseToken: current.LeaseToken, LeaseGeneration: current.LeaseGeneration,
		})
		return mapResult(row, err, domain.ErrLeaseLost)
	})
}

func (r *Repository) DeferAction(ctx context.Context, run domain.Run) error {
	count, err := r.query(ctx).DeferAgentInvocationAction(ctx, invocationsql.DeferAgentInvocationActionParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration,
	})
	if err == nil && count == 0 {
		return domain.ErrLeaseLost
	}
	return err
}
