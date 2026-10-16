// Package agentinvocationsrepository persists surface-neutral assistant work.
package agentinvocationsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maximumAttempts = domain.MaximumAttempts

type Repository struct {
	pool    *pgxpool.Pool
	queries *invocationsql.Queries
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool, queries: invocationsql.New(pool)}
}

func (r *Repository) query(ctx context.Context) *invocationsql.Queries {
	if tx, ok := database.UnitTransaction(ctx); ok {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *Repository) Get(ctx context.Context, workspaceID, id uuid.UUID) (domain.Run, error) {
	row, err := r.query(ctx).GetAgentInvocation(ctx, invocationsql.GetAgentInvocationParams{WorkspaceID: workspaceID, ID: id})
	return mapResult(row, err, domain.ErrNotFound)
}

func (r *Repository) List(ctx context.Context, workspaceID, entityID uuid.UUID) ([]domain.Run, error) {
	rows, err := r.query(ctx).ListAgentInvocations(ctx, invocationsql.ListAgentInvocationsParams{WorkspaceID: workspaceID, EntityID: entityID})
	if err != nil {
		return nil, err
	}
	runs := make([]domain.Run, 0, len(rows))
	for _, row := range rows {
		run, err := toRun(row)
		if err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

func (r *Repository) Control(ctx context.Context, control domain.Control, operation string) (domain.Run, error) {
	nextStatus := domain.StatusQueued
	switch operation {
	case "cancel":
		nextStatus = domain.StatusCancelled
	case "retry", "approve":
	default:
		return domain.Run{}, domain.ErrInvalidInput
	}
	row, err := r.query(ctx).ControlAgentInvocation(ctx, invocationsql.ControlAgentInvocationParams{
		ID: control.RunID, WorkspaceID: control.WorkspaceID, RequesterID: control.RequesterID,
		ExpectedVersion: control.ExpectedVersion, Operation: operation, NextStatus: string(nextStatus),
	})
	return mapResult(row, err, domain.ErrConflict)
}

func (r *Repository) Pending(ctx context.Context, limit int) ([]uuid.UUID, error) {
	if limit < 1 || limit > 100 {
		return nil, domain.ErrInvalidInput
	}
	return r.query(ctx).ListPendingAgentInvocationIDs(ctx, invocationsql.ListPendingAgentInvocationIDsParams{RowLimit: int32(limit)})
}

func (r *Repository) Claim(ctx context.Context, id uuid.UUID) (domain.Run, error) {
	token := uuid.New()
	row, err := r.query(ctx).ClaimAgentInvocation(ctx, invocationsql.ClaimAgentInvocationParams{ID: id, LeaseToken: &token})
	return mapResult(row, err, domain.ErrLeaseLost)
}

func (r *Repository) SaveProposal(ctx context.Context, run domain.Run, proposal domain.Proposal) (domain.Run, error) {
	if err := proposal.Validate(); err != nil {
		return domain.Run{}, err
	}
	encoded, err := json.Marshal(proposal)
	if err != nil {
		return domain.Run{}, err
	}
	row, err := r.query(ctx).SaveAgentInvocationProposal(ctx, invocationsql.SaveAgentInvocationProposalParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration,
		Proposal: encoded, Summary: proposal.Summary, NextStatus: string(domain.StatusRunning),
	})
	return mapResult(row, err, domain.ErrLeaseLost)
}

// Execute commits the action and its receipt together. Recovered leases inspect
// the receipt before calling the adapter, preventing a committed action replay.
func (r *Repository) Execute(ctx context.Context, run domain.Run, apply func(context.Context, domain.Run) (domain.Outcome, error)) (domain.Run, error) {
	if apply == nil {
		return domain.Run{}, domain.ErrInvalidInput
	}
	return r.withLease(ctx, run, func(txCtx context.Context, queries *invocationsql.Queries, current domain.Run) (domain.Run, error) {
		if current.Outcome != nil {
			return current, nil
		}
		if current.Proposal == nil || (current.Proposal.RequiresApproval && current.ApprovedAt == nil) {
			return domain.Run{}, domain.ErrConflict
		}
		outcome, err := apply(txCtx, current)
		if err != nil {
			return domain.Run{}, err
		}
		if err := outcome.Validate(); err != nil {
			return domain.Run{}, err
		}
		encoded, err := json.Marshal(outcome)
		if err != nil {
			return domain.Run{}, err
		}
		row, err := queries.SaveAgentInvocationOutcome(txCtx, invocationsql.SaveAgentInvocationOutcomeParams{
			ID: current.ID, LeaseToken: current.LeaseToken, LeaseGeneration: current.LeaseGeneration, Outcome: encoded,
		})
		return mapResult(row, err, domain.ErrLeaseLost)
	})
}

// Publish commits the source reply and its stable ID together. Delivery of
// notifications is a later retryable stage and cannot lose the durable reply.
func (r *Repository) Publish(ctx context.Context, run domain.Run, publish func(context.Context, domain.Run) (uuid.UUID, error)) (domain.Run, error) {
	if publish == nil {
		return domain.Run{}, domain.ErrInvalidInput
	}
	return r.withLease(ctx, run, func(txCtx context.Context, queries *invocationsql.Queries, current domain.Run) (domain.Run, error) {
		if current.ReplyCommentID != nil {
			return current, nil
		}
		if current.Outcome == nil {
			return domain.Run{}, domain.ErrConflict
		}
		replyID, err := publish(txCtx, current)
		if err != nil {
			return domain.Run{}, err
		}
		if replyID == uuid.Nil {
			return domain.Run{}, domain.ErrInvalidInput
		}
		row, err := queries.SaveAgentInvocationPublication(txCtx, invocationsql.SaveAgentInvocationPublicationParams{
			ID: current.ID, LeaseToken: current.LeaseToken, LeaseGeneration: current.LeaseGeneration, ReplyID: &replyID,
		})
		return mapResult(row, err, domain.ErrLeaseLost)
	})
}

func (r *Repository) Complete(ctx context.Context, run domain.Run) (domain.Run, error) {
	if run.Outcome == nil || run.ReplyCommentID == nil {
		return domain.Run{}, domain.ErrConflict
	}
	nextStatus := domain.StatusSucceeded
	if run.Proposal != nil && run.Proposal.NeedsInput {
		nextStatus = domain.StatusNeedsInput
	}
	row, err := r.query(ctx).CompleteAgentInvocation(ctx, invocationsql.CompleteAgentInvocationParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration, NextStatus: string(nextStatus), ReplyID: run.ReplyCommentID,
	})
	return mapResult(row, err, domain.ErrLeaseLost)
}

func (r *Repository) Fail(ctx context.Context, run domain.Run, summary string, retry bool) error {
	rows, err := r.query(ctx).FailAgentInvocation(ctx, invocationsql.FailAgentInvocationParams{
		ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration, Retry: retry, MaximumAttempts: maximumAttempts, Summary: summary,
		RetrySeconds: float64(min(max(run.RetryAttemptCount, 1)*5, 30)),
	})
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrLeaseLost
	}
	return nil
}

func (r *Repository) withLease(ctx context.Context, run domain.Run, operation func(context.Context, *invocationsql.Queries, domain.Run) (domain.Run, error)) (domain.Run, error) {
	var result domain.Run
	var err error
	for attempt := 0; attempt < 3; attempt++ {
		err = database.WithinUnitOfWork(ctx, r.pool, func(txCtx context.Context, tx pgx.Tx) error {
			queries := invocationsql.New(tx)
			row, err := queries.LockAgentInvocationLease(txCtx, invocationsql.LockAgentInvocationLeaseParams{ID: run.ID, LeaseToken: run.LeaseToken, LeaseGeneration: run.LeaseGeneration})
			current, err := mapResult(row, err, domain.ErrLeaseLost)
			if err != nil {
				return err
			}
			result, err = operation(txCtx, queries, current)
			return err
		})
		if !database.IsRetryableTransactionError(err) {
			break
		}
	}
	if err != nil {
		return domain.Run{}, err
	}
	return result, nil
}

func mapResult(row invocationsql.AgentInvocation, err, noRows error) (domain.Run, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Run{}, noRows
	}
	if err != nil {
		return domain.Run{}, err
	}
	return toRun(row)
}

func toRun(row invocationsql.AgentInvocation) (domain.Run, error) {
	run := domain.Run{
		ID: row.ID, WorkspaceID: row.WorkspaceID, RequesterID: row.RequesterID, TeamID: row.TeamID, Surface: row.Surface,
		SourceID: row.SourceID, ConversationID: row.ConversationID, EntityID: row.EntityID, SourceRevision: row.SourceRevision, SourceHash: row.SourceHash,
		ParentRunID: row.ParentRunID,
		Payload:     row.Payload, Status: domain.Status(row.Status), Summary: row.Summary, Version: row.Version,
		ReplyCommentID: row.ReplyID, ProposalReplyID: row.ProposalReplyID, ProposalDeliveredAt: row.ProposalDeliveredAt, FailureReplyID: row.FailureReplyID, FailurePending: row.FailurePending, LeaseToken: row.LeaseToken, LeaseGeneration: row.LeaseGeneration, LeaseExpiresAt: row.LeaseExpiresAt,
		AttemptCount: row.AttemptCount, RetryAttemptCount: row.RetryAttemptCount, ApprovedAt: row.ApprovedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if len(row.Proposal) > 0 {
		run.Proposal = &domain.Proposal{}
		if err := json.Unmarshal(row.Proposal, run.Proposal); err != nil {
			return domain.Run{}, fmt.Errorf("decode invocation proposal: %w", err)
		}
	}
	if len(row.Outcome) > 0 {
		run.Outcome = &domain.Outcome{}
		if err := json.Unmarshal(row.Outcome, run.Outcome); err != nil {
			return domain.Run{}, fmt.Errorf("decode invocation outcome: %w", err)
		}
	}
	return run, nil
}
