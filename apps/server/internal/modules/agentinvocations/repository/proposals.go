package agentinvocationsrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
	"github.com/google/uuid"
)

func (r *Repository) PublishProposal(ctx context.Context, run domain.Run, publish func(context.Context, domain.Run) (uuid.UUID, error)) (domain.Run, error) {
	if publish == nil {
		return domain.Run{}, domain.ErrInvalidInput
	}
	return r.withLease(ctx, run, func(txCtx context.Context, queries *invocationsql.Queries, current domain.Run) (domain.Run, error) {
		if current.ProposalReplyID != nil {
			return current, nil
		}
		if current.Proposal == nil || !current.Proposal.RequiresApproval || current.ApprovedAt != nil || current.Outcome != nil {
			return domain.Run{}, domain.ErrConflict
		}
		id, err := publish(txCtx, current)
		if err != nil {
			return domain.Run{}, err
		}
		if id == uuid.Nil {
			return domain.Run{}, domain.ErrInvalidInput
		}
		row, err := queries.SaveAgentInvocationProposalPublication(txCtx, invocationsql.SaveAgentInvocationProposalPublicationParams{
			ID: current.ID, LeaseToken: current.LeaseToken, LeaseGeneration: current.LeaseGeneration, ProposalReplyID: &id,
		})
		return mapResult(row, err, domain.ErrLeaseLost)
	})
}

func (r *Repository) AwaitApproval(ctx context.Context, run domain.Run) (domain.Run, error) {
	row, err := r.query(ctx).AwaitAgentInvocationApproval(ctx, invocationsql.AwaitAgentInvocationApprovalParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration,
	})
	return mapResult(row, err, domain.ErrLeaseLost)
}
