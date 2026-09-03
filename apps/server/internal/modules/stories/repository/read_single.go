package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *repo) GetVisibleStory(
	ctx context.Context,
	scope storydomain.ReadScope,
	storyID uuid.UUID,
) (storydomain.Story, error) {
	if r.reads == nil {
		return storydomain.Story{}, errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return storydomain.Story{}, err
	}
	if storyID == uuid.Nil {
		return storydomain.Story{}, fmt.Errorf("%w: story id is required", errInvalidReadQuery)
	}

	row, err := r.reads.GetVisibleStory(ctx, storyreadsql.GetVisibleStoryParams{
		ActorID:                scope.ActorID,
		StoryID:                storyID,
		WorkspaceID:            scope.WorkspaceID,
		BypassActorMembership:  false,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return storydomain.Story{}, visibleStoryReadError(err)
	}

	story, err := mapVisibleStory(row)
	if err != nil {
		return storydomain.Story{}, err
	}
	if err := r.enrichVisibleStory(ctx, scope, &story); err != nil {
		return storydomain.Story{}, err
	}
	return story, nil
}

func (r *repo) QueryVisibleStoryByRef(
	ctx context.Context,
	scope storydomain.ReadScope,
	teamCode string,
	sequenceID int,
) (storydomain.Story, error) {
	storyID, err := r.GetVisibleStoryIDByRef(ctx, scope, teamCode, sequenceID)
	if err != nil {
		return storydomain.Story{}, err
	}
	return r.GetVisibleStory(ctx, scope, storyID)
}

// QueryCredentialVisibleStoryByRef serves a pre-authorized integration
// credential. The caller supplies a restricted team scope; this path never
// grants unrestricted access and does not rely on a human membership row.
func (r *repo) QueryCredentialVisibleStoryByRef(
	ctx context.Context,
	scope storydomain.ReadScope,
	teamCode string,
	sequenceID int,
) (storydomain.Story, error) {
	if r.reads == nil {
		return storydomain.Story{}, errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return storydomain.Story{}, err
	}
	if scope.UnrestrictedTeamAccess {
		return storydomain.Story{}, fmt.Errorf("%w: integration story reads require restricted team access", errInvalidReadQuery)
	}
	teamCode = strings.ToUpper(strings.TrimSpace(teamCode))
	if teamCode == "" || sequenceID < 1 || sequenceID > int(^uint32(0)>>1) {
		return storydomain.Story{}, fmt.Errorf("%w: valid team code and sequence are required", errInvalidReadQuery)
	}
	sequence := int32(sequenceID)
	storyID, err := r.reads.GetVisibleStoryIDByRef(ctx, storyreadsql.GetVisibleStoryIDByRefParams{
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		TeamCode:               teamCode,
		SequenceID:             &sequence,
		BypassActorMembership:  true,
		UnrestrictedTeamAccess: false,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return storydomain.Story{}, visibleStoryReadError(err)
	}
	row, err := r.reads.GetVisibleStory(ctx, storyreadsql.GetVisibleStoryParams{
		ActorID:                scope.ActorID,
		StoryID:                storyID,
		WorkspaceID:            scope.WorkspaceID,
		BypassActorMembership:  true,
		UnrestrictedTeamAccess: false,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return storydomain.Story{}, visibleStoryReadError(err)
	}
	return mapVisibleStory(row)
}

func (r *repo) GetVisibleStoryIDByRef(
	ctx context.Context,
	scope storydomain.ReadScope,
	teamCode string,
	sequenceID int,
) (uuid.UUID, error) {
	if r.reads == nil {
		return uuid.Nil, errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return uuid.Nil, err
	}
	teamCode = strings.ToUpper(strings.TrimSpace(teamCode))
	if teamCode == "" || sequenceID < 1 || sequenceID > int(^uint32(0)>>1) {
		return uuid.Nil, fmt.Errorf("%w: valid team code and sequence are required", errInvalidReadQuery)
	}
	sequence := int32(sequenceID)

	storyID, err := r.reads.GetVisibleStoryIDByRef(ctx, storyreadsql.GetVisibleStoryIDByRefParams{
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		TeamCode:               teamCode,
		SequenceID:             &sequence,
		BypassActorMembership:  false,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return uuid.Nil, visibleStoryReadError(err)
	}
	return storyID, nil
}

func (r *repo) enrichVisibleStory(
	ctx context.Context,
	scope storydomain.ReadScope,
	story *storydomain.Story,
) error {
	subStories, err := r.listVisibleSubStories(ctx, scope, []uuid.UUID{story.ID})
	if err != nil {
		return err
	}
	story.SubStories = subStories[story.ID]
	if story.SubStories == nil {
		story.SubStories = []storydomain.StoryList{}
	}

	rows, err := r.reads.ListVisibleStoryAssociations(ctx, storyreadsql.ListVisibleStoryAssociationsParams{
		StoryID:                story.ID,
		WorkspaceID:            scope.WorkspaceID,
		ActorID:                scope.ActorID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return fmt.Errorf("list visible story associations: %w", err)
	}
	story.Associations, err = mapVisibleAssociations(rows)
	return err
}

func visibleStoryReadError(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.ErrNotFound
	}
	return fmt.Errorf("read visible story: %w", err)
}
