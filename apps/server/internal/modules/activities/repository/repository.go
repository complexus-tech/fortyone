package activitiesrepository

import (
	"context"
	"errors"
	"fmt"

	activitiesdomain "github.com/complexus-tech/projects-api/internal/modules/activities/domain"
	activitiessql "github.com/complexus-tech/projects-api/internal/modules/activities/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("activities repository is not configured")

type Repository struct {
	queries activitiessql.Querier
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	return &Repository{queries: activitiessql.New(pool)}
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

func (repository *Repository) Create(ctx context.Context, activity activitiesdomain.NewActivity) error {
	if err := repository.configured(); err != nil {
		return err
	}

	rows, err := repository.queries.CreateActivity(ctx, activitiessql.CreateActivityParams{
		StoryID:      activity.StoryID,
		UserID:       activity.UserID,
		ActivityType: activity.Type,
		FieldChanged: activity.Field,
		CurrentValue: activity.CurrentValue,
		WorkspaceID:  activity.WorkspaceID,
	})
	if err != nil {
		return fmt.Errorf("create activity: %w", err)
	}
	if rows != 1 {
		return activitiesdomain.ErrScopeMismatch
	}
	return nil
}

func (repository *Repository) GetActivities(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
	workspaceID uuid.UUID,
	filters activitiesdomain.Filters,
) ([]activitiesdomain.Activity, error) {
	if err := repository.configured(); err != nil {
		return nil, err
	}
	resultLimit, err := activityLimit(limit)
	if err != nil {
		return nil, err
	}

	rows, err := repository.queries.ListActivitiesForMember(ctx, activitiessql.ListActivitiesForMemberParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StartDate:   filters.StartDate,
		EndDate:     filters.EndDate,
		ResultLimit: resultLimit,
	})
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}

	activities := make([]activitiesdomain.Activity, len(rows))
	for index, row := range rows {
		workspaceID := uuid.Nil
		if row.WorkspaceID != nil {
			workspaceID = *row.WorkspaceID
		}
		activities[index] = activitiesdomain.Activity{
			ID:           row.ActivityID,
			StoryID:      row.StoryID,
			UserID:       row.UserID,
			Type:         row.ActivityType,
			Field:        row.FieldChanged,
			CurrentValue: row.CurrentValue,
			CreatedAt:    row.CreatedAt,
			WorkspaceID:  workspaceID,
			User: activitiesdomain.UserDetails{
				ID:        row.UserID,
				Username:  row.Username,
				FullName:  optionalString(row.FullName),
				AvatarURL: optionalString(row.AvatarURL),
				IsActive:  row.IsActive,
			},
		}
	}
	return activities, nil
}

func activityLimit(limit int) (int32, error) {
	if limit < 1 || limit > 100 {
		return 0, activitiesdomain.ErrInvalidLimit
	}
	result, err := safecast.Int32(limit)
	if err != nil {
		return 0, activitiesdomain.ErrInvalidLimit
	}
	return result, nil
}

func optionalString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
