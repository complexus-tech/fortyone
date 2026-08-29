package okractivitiesrepository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	okractivitiesdomain "github.com/complexus-tech/projects-api/internal/modules/okractivities/domain"
	okractivitiessql "github.com/complexus-tech/projects-api/internal/modules/okractivities/repository/sqlc"
	platformdatabase "github.com/complexus-tech/projects-api/internal/platform/database"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errRepositoryNotConfigured = errors.New("OKR activities repository is not configured")

type Repository struct {
	queries        okractivitiessql.Querier
	runTransaction func(context.Context, func(okractivitiessql.Querier) error) error
}

func New(pool *pgxpool.Pool) *Repository {
	if pool == nil {
		return &Repository{}
	}
	queries := okractivitiessql.New(pool)
	transactor := platformdatabase.NewTransactor(pool)
	repository := newWithQueries(queries)
	repository.runTransaction = func(ctx context.Context, operation func(okractivitiessql.Querier) error) error {
		return transactor.WithinTransaction(ctx, pgx.TxOptions{}, func(tx pgx.Tx) error {
			return operation(queries.WithTx(tx))
		})
	}
	return repository
}

func newWithQueries(queries okractivitiessql.Querier) *Repository {
	repository := &Repository{queries: queries}
	if queries != nil {
		repository.runTransaction = func(ctx context.Context, operation func(okractivitiessql.Querier) error) error {
			return operation(queries)
		}
	}
	return repository
}

func (repository *Repository) configured() error {
	if repository == nil || repository.queries == nil {
		return errRepositoryNotConfigured
	}
	return nil
}

func (repository *Repository) Create(ctx context.Context, activity okractivitiesdomain.NewActivity) error {
	return repository.CreateBatch(ctx, []okractivitiesdomain.NewActivity{activity})
}

func (repository *Repository) CreateBatch(ctx context.Context, activities []okractivitiesdomain.NewActivity) error {
	normalized, err := (okractivitiesdomain.CreateBatchCommand{Activities: activities}).Normalize()
	if err != nil {
		return err
	}
	if len(normalized.Activities) == 0 {
		return nil
	}
	if repository == nil || repository.runTransaction == nil {
		return errRepositoryNotConfigured
	}
	err = repository.runTransaction(ctx, func(queries okractivitiessql.Querier) error {
		for _, activity := range normalized.Activities {
			field, currentValue, comment := activity.Field, activity.CurrentValue, activity.Comment
			rows, err := queries.CreateOKRActivity(ctx, okractivitiessql.CreateOKRActivityParams{
				KeyResultID:  activity.KeyResultID,
				ActivityType: okractivitiessql.OkrActivityType(activity.Type),
				UpdateType:   okractivitiessql.OkrUpdateType(activity.UpdateType),
				FieldChanged: &field, CurrentValue: &currentValue, Comment: &comment,
				ActorID: activity.UserID, ObjectiveID: activity.ObjectiveID,
				WorkspaceID: uuidPointer(activity.WorkspaceID),
			})
			if err != nil {
				return fmt.Errorf("create OKR activity: %w", err)
			}
			if rows != 1 {
				return okractivitiesdomain.ErrScopeMismatch
			}
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("create OKR activity batch: %w", err)
	}
	return nil
}

func (repository *Repository) List(
	ctx context.Context,
	query okractivitiesdomain.ListQuery,
) ([]okractivitiesdomain.Activity, bool, error) {
	normalized, err := query.Normalize()
	if err != nil {
		return nil, false, err
	}
	if err := repository.configured(); err != nil {
		return nil, false, err
	}
	offset, limit, err := activityPageBounds(normalized.Page, normalized.PageSize)
	if err != nil {
		return nil, false, err
	}
	var activities []okractivitiesdomain.Activity
	if normalized.KeyResultID != nil {
		rows, err := repository.queries.ListKeyResultActivities(ctx, okractivitiessql.ListKeyResultActivitiesParams{
			ActorID: normalized.ActorID, KeyResultID: normalized.KeyResultID,
			WorkspaceID: normalized.WorkspaceID, ResultOffset: offset, ResultLimit: limit,
		})
		if err != nil {
			return nil, false, fmt.Errorf("list key result activities: %w", err)
		}
		activities = make([]okractivitiesdomain.Activity, 0, len(rows))
		for _, row := range rows {
			activities = append(activities, activityFromValues(
				row.ActivityID, row.ObjectiveID, row.KeyResultID, row.UserID,
				row.ActivityType, row.UpdateType, row.FieldChanged, row.CurrentValue,
				row.Comment, row.CreatedAt, row.WorkspaceID, row.Username,
				row.FullName, row.AvatarURL, row.IsActive,
			))
		}
	} else {
		rows, err := repository.queries.ListObjectiveActivities(ctx, okractivitiessql.ListObjectiveActivitiesParams{
			ActorID: normalized.ActorID, ObjectiveID: normalized.ObjectiveID,
			WorkspaceID: normalized.WorkspaceID, ResultOffset: offset, ResultLimit: limit,
		})
		if err != nil {
			return nil, false, fmt.Errorf("list objective activities: %w", err)
		}
		activities = make([]okractivitiesdomain.Activity, 0, len(rows))
		for _, row := range rows {
			activities = append(activities, activityFromValues(
				row.ActivityID, row.ObjectiveID, row.KeyResultID, row.UserID,
				row.ActivityType, row.UpdateType, row.FieldChanged, row.CurrentValue,
				row.Comment, row.CreatedAt, row.WorkspaceID, row.Username,
				row.FullName, row.AvatarURL, row.IsActive,
			))
		}
	}
	hasMore := len(activities) > normalized.PageSize
	if hasMore {
		activities = activities[:normalized.PageSize]
	}
	return activities, hasMore, nil
}

func activityPageBounds(page, pageSize int) (int32, int32, error) {
	if page < 1 || pageSize < 1 || page-1 > math.MaxInt/pageSize {
		return 0, 0, fmt.Errorf("%w: pagination offset is outside the supported range", okractivitiesdomain.ErrInvalid)
	}
	offset, err := safecast.Int32((page - 1) * pageSize)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: pagination offset: %v", okractivitiesdomain.ErrInvalid, err)
	}
	limit, err := safecast.Int32(pageSize + 1)
	if err != nil {
		return 0, 0, fmt.Errorf("%w: pagination limit: %v", okractivitiesdomain.ErrInvalid, err)
	}
	return offset, limit, nil
}

func activityFromValues(
	id, objectiveID uuid.UUID,
	keyResultID *uuid.UUID,
	userID uuid.UUID,
	activityType, updateType, field, currentValue, comment string,
	createdAt *time.Time,
	workspaceID uuid.UUID,
	username, fullName, avatarURL string,
	isActive bool,
) okractivitiesdomain.Activity {
	created := time.Unix(0, 0).UTC()
	if createdAt != nil {
		created = createdAt.UTC()
	}
	return okractivitiesdomain.Activity{
		ID: id, ObjectiveID: objectiveID, KeyResultID: keyResultID, UserID: userID,
		Type:       okractivitiesdomain.ActivityType(activityType),
		UpdateType: okractivitiesdomain.UpdateType(updateType), Field: field,
		CurrentValue: currentValue, Comment: comment, CreatedAt: created,
		WorkspaceID: workspaceID,
		User: okractivitiesdomain.UserDetails{
			ID: userID, Username: username, FullName: fullName,
			AvatarURL: avatarURL, IsActive: isActive,
		},
	}
}

func uuidPointer(value uuid.UUID) *uuid.UUID {
	return &value
}
