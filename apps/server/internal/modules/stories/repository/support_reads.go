package storiesrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	maxStoryActivityPage     = 10_000
	maxStoryActivityPageSize = 100
)

func (r *repo) GetTeamEstimateScheme(ctx context.Context, scope storydomain.ReadScope, teamID uuid.UUID) (string, error) {
	if err := validateReadScope(scope); err != nil {
		return "", err
	}
	scheme, err := r.reads.GetVisibleTeamEstimateScheme(ctx, storyreadsql.GetVisibleTeamEstimateSchemeParams{
		ActorID: scope.ActorID, TeamID: teamID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return "", fmt.Errorf("read visible team estimate scheme: %w", err)
	}
	return scheme, nil
}

func (r *repo) FindFirstStatusByCategory(
	ctx context.Context,
	scope storydomain.ReadScope,
	teamID uuid.UUID,
	category string,
) (*uuid.UUID, error) {
	if err := validateReadScope(scope); err != nil {
		return nil, err
	}
	category = strings.TrimSpace(category)
	if teamID == uuid.Nil || category == "" {
		return nil, storydomain.ErrInvalidReadQuery
	}
	statusID, err := r.reads.FindVisibleFirstStatusByCategory(ctx, storyreadsql.FindVisibleFirstStatusByCategoryParams{
		ActorID: scope.ActorID, TeamID: teamID, WorkspaceID: scope.WorkspaceID, Category: &category,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("find visible default story status: %w", err)
	}
	return &statusID, nil
}

func (r *repo) ResolveKeyResult(
	ctx context.Context,
	scope storydomain.ReadScope,
	keyResultID uuid.UUID,
) (storydomain.MutationKeyResultReference, error) {
	if err := validateReadScope(scope); err != nil {
		return storydomain.MutationKeyResultReference{}, err
	}
	row, err := r.reads.ResolveVisibleStoryKeyResult(ctx, storyreadsql.ResolveVisibleStoryKeyResultParams{
		ActorID: scope.ActorID, KeyResultID: keyResultID, WorkspaceID: &scope.WorkspaceID,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return storydomain.MutationKeyResultReference{}, storydomain.ErrNotFound
	}
	if err != nil {
		return storydomain.MutationKeyResultReference{}, fmt.Errorf("resolve visible story key result: %w", err)
	}
	return storydomain.MutationKeyResultReference{ObjectiveID: row.ObjectiveID, Name: row.Name}, nil
}

func (r *repo) ListVisibleStoryLinks(ctx context.Context, scope storydomain.ReadScope, storyID uuid.UUID) ([]storydomain.StoryLink, error) {
	if err := validateReadScope(scope); err != nil {
		return nil, err
	}
	rows, err := r.reads.ListVisibleStoryLinks(ctx, storyreadsql.ListVisibleStoryLinksParams{
		ActorID: scope.ActorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if err != nil {
		return nil, fmt.Errorf("list visible story links: %w", err)
	}
	result := make([]storydomain.StoryLink, len(rows))
	for index, row := range rows {
		result[index] = storydomain.StoryLink{
			ID: row.LinkID, Title: row.Title, URL: row.URL, StoryID: row.StoryID,
			CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
	}
	return result, nil
}

func (r *repo) ListVisibleStoryActivities(
	ctx context.Context,
	scope storydomain.ReadScope,
	storyID uuid.UUID,
	page, pageSize int,
) ([]storydomain.ActivityWithUser, bool, error) {
	if err := validateReadScope(scope); err != nil {
		return nil, false, err
	}
	offset, limit, err := storyActivityPage(page, pageSize)
	if err != nil {
		return nil, false, err
	}
	rows, err := r.reads.ListVisibleStoryActivities(ctx, storyreadsql.ListVisibleStoryActivitiesParams{
		ActorID: scope.ActorID, StoryID: storyID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
		RowOffset:              offset, RowLimit: limit,
	})
	if err != nil {
		return nil, false, fmt.Errorf("list visible story activities: %w", err)
	}
	hasMore := len(rows) > pageSize
	if hasMore {
		rows = rows[:pageSize]
	}
	result := make([]storydomain.ActivityWithUser, len(rows))
	for index, row := range rows {
		result[index] = mapVisibleStoryActivity(row, scope.WorkspaceID)
	}
	return result, hasMore, nil
}

func (r *repo) GetStatusCategory(ctx context.Context, scope storydomain.ReadScope, statusID uuid.UUID) (string, error) {
	if err := validateReadScope(scope); err != nil {
		return "", err
	}
	category, err := r.reads.GetVisibleStoryStatusCategory(ctx, storyreadsql.GetVisibleStoryStatusCategoryParams{
		ActorID: scope.ActorID, StatusID: statusID, WorkspaceID: scope.WorkspaceID,
		UnrestrictedTeamAccess: scope.UnrestrictedTeamAccess,
		AllowedTeamIds:         cloneUUIDs(scope.AllowedTeamIDs),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", storydomain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("read visible story status category: %w", err)
	}
	return category, nil
}

func storyActivityPage(page, pageSize int) (int32, int32, error) {
	if page < 1 || page > maxStoryActivityPage || pageSize < 1 || pageSize > maxStoryActivityPageSize {
		return 0, 0, storydomain.ErrInvalidReadQuery
	}
	offset, err := safecast.Int64ToInt32(int64(page-1) * int64(pageSize))
	if err != nil {
		return 0, 0, fmt.Errorf("%w: activity page offset is outside the supported range", storydomain.ErrInvalidReadQuery)
	}
	limit, err := safecast.Int64ToInt32(int64(pageSize) + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: activity page size is outside the supported range", storydomain.ErrInvalidReadQuery)
	}
	return offset, limit, nil
}

func mapVisibleStoryActivity(
	row storyreadsql.ListVisibleStoryActivitiesRow,
	workspaceID uuid.UUID,
) storydomain.ActivityWithUser {
	return storydomain.ActivityWithUser{
		ID: row.ActivityID, StoryID: row.StoryID, UserID: row.UserID,
		Type: row.ActivityType, Field: row.FieldChanged, CurrentValue: row.CurrentValue,
		OldValue: activityJSONValue(row.OldValue), NewValue: activityJSONValue(row.NewValue),
		Reason: row.Reason, CreatedAt: row.CreatedAt, WorkspaceID: workspaceID,
		User: storydomain.ActivityUser{
			ID: row.UserID, Username: row.Username, FullName: row.FullName,
			AvatarURL: row.AvatarURL, IsActive: row.IsActive, IsSystem: row.IsSystem,
		},
	}
}

func activityJSONValue(value []byte) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	var decoded any
	if err := json.Unmarshal(value, &decoded); err != nil {
		return string(value)
	}
	return decoded
}
