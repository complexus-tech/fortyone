package agentinvocationsrepository

import (
	"context"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) BeginFailure(ctx context.Context, run domain.Run, summary string) (domain.Run, error) {
	if strings.TrimSpace(summary) == "" || len(summary) > 2000 {
		return domain.Run{}, domain.ErrInvalidInput
	}
	row, err := r.query(ctx).BeginAgentInvocationFailure(ctx, invocationsql.BeginAgentInvocationFailureParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration, Summary: summary,
	})
	return mapResult(row, err, domain.ErrLeaseLost)
}

func (r *Repository) PublishFailure(ctx context.Context, run domain.Run, publish func(context.Context, domain.Run) (uuid.UUID, error)) (domain.Run, error) {
	if publish == nil {
		return domain.Run{}, domain.ErrInvalidInput
	}
	return r.withLease(ctx, run, func(txCtx context.Context, queries *invocationsql.Queries, current domain.Run) (domain.Run, error) {
		if !current.FailurePending || current.Outcome != nil {
			return domain.Run{}, domain.ErrConflict
		}
		if current.FailureReplyID != nil {
			return current, nil
		}
		id, err := publish(txCtx, current)
		if err != nil {
			return domain.Run{}, err
		}
		if id == uuid.Nil {
			return domain.Run{}, domain.ErrInvalidInput
		}
		row, err := queries.SaveAgentInvocationFailurePublication(txCtx, invocationsql.SaveAgentInvocationFailurePublicationParams{
			ID: current.ID, LeaseToken: current.LeaseToken, LeaseGeneration: current.LeaseGeneration, FailureReplyID: &id,
		})
		return mapResult(row, err, domain.ErrLeaseLost)
	})
}

func (r *Repository) DeferFailure(ctx context.Context, run domain.Run) error {
	count, err := r.query(ctx).DeferAgentInvocationFailure(ctx, invocationsql.DeferAgentInvocationFailureParams{ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration})
	if err != nil {
		return err
	}
	if count == 0 {
		return domain.ErrLeaseLost
	}
	return nil
}

func (r *Repository) CompleteFailure(ctx context.Context, run domain.Run) (domain.Run, error) {
	row, err := r.query(ctx).CompleteAgentInvocationFailure(ctx, invocationsql.CompleteAgentInvocationFailureParams{ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration})
	return mapResult(row, err, domain.ErrLeaseLost)
}
