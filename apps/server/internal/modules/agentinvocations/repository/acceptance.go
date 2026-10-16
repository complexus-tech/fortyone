package agentinvocationsrepository

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/jackc/pgx/v5"
)

// Accept joins the source's unit of work so its comment, durable dispatch and
// optional continuation transition commit together. The caller authorizes the
// source; persistence also checks current human tenant/team membership.
func (r *Repository) Accept(ctx context.Context, input domain.Input) (domain.Run, bool, error) {
	if err := input.Validate(); err != nil {
		return domain.Run{}, false, err
	}
	if len(input.Payload) == 0 {
		input.Payload = json.RawMessage(`{}`)
	}
	var payload bytes.Buffer
	if err := json.Compact(&payload, input.Payload); err != nil {
		return domain.Run{}, false, domain.ErrInvalidInput
	}
	input.Payload = payload.Bytes()
	identity := input
	identity.Payload = nil
	serialized, err := json.Marshal(identity)
	if err != nil {
		return domain.Run{}, false, fmt.Errorf("encode invocation identity: %w", err)
	}
	hash := sha256.Sum256(serialized)
	if _, joined := database.UnitTransaction(ctx); joined {
		return r.accept(ctx, input, hash[:])
	}
	var run domain.Run
	var created bool
	for attempt := 0; attempt < 3; attempt++ {
		err = database.WithinUnitOfWork(ctx, r.pool, func(txCtx context.Context, _ pgx.Tx) error {
			var acceptErr error
			run, created, acceptErr = r.accept(txCtx, input, hash[:])
			return acceptErr
		})
		if !database.IsRetryableTransactionError(err) {
			break
		}
	}
	if err != nil {
		return domain.Run{}, false, err
	}
	return run, created, nil
}

func (r *Repository) accept(ctx context.Context, input domain.Input, hash []byte) (domain.Run, bool, error) {
	queries := r.query(ctx)
	if input.ContinuationID != nil {
		parent, err := queries.LockAgentInvocationContinuation(ctx, invocationsql.LockAgentInvocationContinuationParams{
			ID: *input.ContinuationID, WorkspaceID: input.WorkspaceID, TeamID: input.TeamID,
			Surface: input.Surface, EntityID: input.EntityID, ConversationID: input.ConversationID,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Run{}, false, domain.ErrConflict
		}
		if err != nil {
			return domain.Run{}, false, err
		}
		if parent.Status != string(domain.StatusNeedsInput) && parent.Status != string(domain.StatusAwaitingApproval) {
			if parent.Status != string(domain.StatusSucceeded) {
				return domain.Run{}, false, domain.ErrConflict
			}
			// A replay must resolve to the same already-linked child. A new key
			// cannot continue a clarification a second time or reuse approval.
			row, err := queries.GetAgentInvocationByKey(ctx, keyParams(input))
			if errors.Is(err, pgx.ErrNoRows) {
				return domain.Run{}, false, domain.ErrConflict
			}
			if err != nil {
				return domain.Run{}, false, err
			}
			if row.ParentRunID == nil || *row.ParentRunID != parent.ID || !bytes.Equal(row.RequestHash, hash) {
				return domain.Run{}, false, domain.ErrConflict
			}
			run, err := toRun(row)
			return run, false, err
		}
	}
	row, err := queries.CreateAgentInvocation(ctx, invocationsql.CreateAgentInvocationParams{
		WorkspaceID: input.WorkspaceID, RequesterID: input.RequesterID, TeamID: input.TeamID,
		Surface: input.Surface, SourceID: input.SourceID, ConversationID: input.ConversationID, EntityID: input.EntityID, ParentRunID: input.ContinuationID,
		SourceRevision: input.SourceRevision, SourceHash: input.SourceHash, IdempotencyKey: input.IdempotencyKey, RequestHash: hash, Payload: input.Payload,
	})
	created := err == nil
	if errors.Is(err, pgx.ErrNoRows) {
		row, err = queries.GetAgentInvocationByKey(ctx, keyParams(input))
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Run{}, false, domain.ErrPermissionDenied
		}
		if err == nil && !bytes.Equal(row.RequestHash, hash) {
			return domain.Run{}, false, domain.ErrConflict
		}
	}
	if err != nil {
		return domain.Run{}, false, fmt.Errorf("accept invocation: %w", err)
	}
	if input.ContinuationID != nil {
		count, err := queries.CompleteAgentInvocationContinuation(ctx, invocationsql.CompleteAgentInvocationContinuationParams{ID: *input.ContinuationID})
		if err != nil {
			return domain.Run{}, false, err
		}
		if count != 1 {
			return domain.Run{}, false, domain.ErrConflict
		}
	}
	run, err := toRun(row)
	return run, created, err
}

func keyParams(input domain.Input) invocationsql.GetAgentInvocationByKeyParams {
	return invocationsql.GetAgentInvocationByKeyParams{WorkspaceID: input.WorkspaceID, RequesterID: input.RequesterID, Surface: input.Surface, IdempotencyKey: input.IdempotencyKey}
}
