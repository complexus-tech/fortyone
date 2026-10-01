package customfieldsrepository

import (
	"context"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	sql "github.com/complexus-tech/projects-api/internal/modules/customfields/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type CreationParticipant struct{ queries *sql.Queries }

func NewCreationParticipant(tx pgx.Tx) *CreationParticipant {
	return &CreationParticipant{queries: sql.New(tx)}
}
func (p *CreationParticipant) Close() { p.queries = nil }

func (p *CreationParticipant) Apply(ctx context.Context, scope domain.Scope, storyID uuid.UUID, values []domain.Value) error {
	if p.queries == nil {
		return domain.ErrConflict
	}
	if err := authorize(ctx, p.queries, scope, true, false); err != nil {
		return err
	}
	story, err := p.queries.LockStoryScope(ctx, sql.LockStoryScopeParams{StoryID: storyID, WorkspaceID: scope.WorkspaceID})
	if err != nil {
		return mapError(err)
	}
	if story.TeamID != scope.TeamID || story.ArchivedAt != nil {
		return domain.ErrConflict
	}
	_, err = applyValues(ctx, p.queries, scope, storyID, values, story.CustomFieldsVersion)
	return mapError(err)
}
