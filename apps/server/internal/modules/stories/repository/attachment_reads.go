package storiesrepository

import (
	"context"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/google/uuid"
)

// AuthorizeVisibleStoryAttachmentRead shares the ordinary story read fence,
// but never fetches associations, sub-stories, or other hydrated records.
func (r *repo) AuthorizeVisibleStoryAttachmentRead(ctx context.Context, scope storydomain.ReadScope, storyID uuid.UUID) error {
	if r.reads == nil {
		return errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return err
	}
	if storyID == uuid.Nil {
		return fmt.Errorf("%w: story id is required", errInvalidReadQuery)
	}
	queries := r.reads
	if tx, ok := platformdatabase.UnitTransaction(ctx); ok {
		queries = storyreadsql.New(tx)
	}
	row, err := queries.GetVisibleStory(ctx, storyreadsql.GetVisibleStoryParams{
		ActorID: scope.ActorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		BypassActorMembership: false, UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds: cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return visibleStoryReadError(err)
	}
	// Ordinary story reads include trash for restore views. Durable attachment
	// links must stop granting fresh access while their story is deleted.
	if row.DeletedAt != nil {
		return storydomain.ErrNotFound
	}
	return nil
}
