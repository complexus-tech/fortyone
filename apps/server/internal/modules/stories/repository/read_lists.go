package storiesrepository

import (
	"context"
	"fmt"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/google/uuid"
)

func (r *repo) ListMyVisibleStories(
	ctx context.Context,
	scope storydomain.ReadScope,
) ([]storydomain.StoryList, error) {
	if r.reads == nil {
		return nil, errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return nil, err
	}

	rows, err := r.reads.ListMyVisibleStories(ctx, storyreadsql.ListMyVisibleStoriesParams{
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
		ResultLimit:            maxMyStoriesResultCount,
	})
	if err != nil {
		return nil, fmt.Errorf("list visible actor stories: %w", err)
	}

	storyList, err := mapMyVisibleStories(rows)
	if err != nil {
		return nil, err
	}
	if err := r.attachVisibleSubStories(ctx, scope, storyList); err != nil {
		return nil, err
	}
	return storyList, nil
}

func (r *repo) ListVisibleStoriesByCategory(
	ctx context.Context,
	scope storydomain.ReadScope,
	teamID uuid.UUID,
	categoryValue string,
	page int,
	pageSize int,
	includeSubStories bool,
) ([]storydomain.StoryList, bool, error) {
	if r.reads == nil {
		return nil, false, errReadRepositoryNotConfigured
	}
	if err := validateReadScope(scope); err != nil {
		return nil, false, err
	}
	if teamID == uuid.Nil {
		return nil, false, fmt.Errorf("%w: team id is required", errInvalidReadQuery)
	}
	category, err := parseStoryCategory(categoryValue)
	if err != nil {
		return nil, false, err
	}
	offset, limit, err := categoryPage(page, pageSize)
	if err != nil {
		return nil, false, err
	}
	categoryParam := string(category)

	rows, err := r.reads.ListVisibleStoriesByCategory(ctx, storyreadsql.ListVisibleStoriesByCategoryParams{
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		TeamID:                 teamID,
		IncludeSubStories:      includeSubStories,
		Category:               &categoryParam,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
		PageOffset:             offset,
		PageLimit:              limit,
	})
	if err != nil {
		return nil, false, fmt.Errorf("list visible category stories: %w", err)
	}

	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	storyList, err := mapVisibleCategoryStories(rows)
	if err != nil {
		return nil, false, err
	}
	if err := r.attachVisibleSubStories(ctx, scope, storyList); err != nil {
		return nil, false, err
	}
	return storyList, hasMore, nil
}

func (r *repo) attachVisibleSubStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	storyList []storydomain.StoryList,
) error {
	parentIDs := make([]uuid.UUID, len(storyList))
	for index := range storyList {
		parentIDs[index] = storyList[index].ID
	}
	byParent, err := r.listVisibleSubStories(ctx, scope, parentIDs)
	if err != nil {
		return err
	}
	for index := range storyList {
		storyList[index].SubStories = byParent[storyList[index].ID]
		if storyList[index].SubStories == nil {
			storyList[index].SubStories = []storydomain.StoryList{}
		}
	}
	return nil
}

func (r *repo) listVisibleSubStories(
	ctx context.Context,
	scope storydomain.ReadScope,
	parentIDs []uuid.UUID,
) (map[uuid.UUID][]storydomain.StoryList, error) {
	byParent := make(map[uuid.UUID][]storydomain.StoryList, len(parentIDs))
	if len(parentIDs) == 0 {
		return byParent, nil
	}
	rows, err := r.reads.ListVisibleSubStories(ctx, storyreadsql.ListVisibleSubStoriesParams{
		ActorID:                scope.ActorID,
		WorkspaceID:            scope.WorkspaceID,
		ParentIds:              append([]uuid.UUID(nil), parentIDs...),
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("list visible sub-stories: %w", err)
	}
	for _, row := range rows {
		story, mapErr := mapVisibleSubStory(row)
		if mapErr != nil {
			return nil, mapErr
		}
		if story.Parent == nil {
			return nil, fmt.Errorf("map visible sub-story %s: parent id is missing", story.ID)
		}
		byParent[*story.Parent] = append(byParent[*story.Parent], story)
	}
	return byParent, nil
}
