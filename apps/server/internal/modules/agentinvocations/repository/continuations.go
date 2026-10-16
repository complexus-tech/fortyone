package agentinvocationsrepository

import (
	"context"
	"errors"

	domain "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/domain"
	invocationsql "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) ResolveContinuation(ctx context.Context, query domain.ContinuationQuery) (*uuid.UUID, error) {
	queries := r.query(ctx)
	parentID, err := queries.GetAgentInvocationSourceParent(ctx, invocationsql.GetAgentInvocationSourceParentParams{
		WorkspaceID: query.WorkspaceID, RequesterID: query.RequesterID, Surface: query.Surface, EntityID: query.EntityID,
		ConversationID: query.ConversationID, SourceID: query.SourceID,
	})
	if err == nil {
		return parentID, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	id, err := queries.FindAgentInvocationClarification(ctx, invocationsql.FindAgentInvocationClarificationParams{
		WorkspaceID: query.WorkspaceID, Surface: query.Surface, EntityID: query.EntityID,
		ConversationID: query.ConversationID, ReplyToID: query.ReplyToID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}
