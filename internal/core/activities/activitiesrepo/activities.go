package activitiesrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/core/activities"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type repo struct {
	db  *sqlx.DB
	log *logger.Logger
}

func New(log *logger.Logger, db *sqlx.DB) *repo {
	return &repo{
		db:  db,
		log: log,
	}
}

func (r *repo) RecordActivity(ctx context.Context, activity *activities.CoreActivity) (activities.CoreActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.activities.RecordActivity")
	defer span.End()

	da, err := r.recordActivity(ctx, activity)
	if err != nil {
		return activities.CoreActivity{}, fmt.Errorf("failed to insert activity: %w", err)
	}
	return toCoreActivity(da), nil
}

func (r *repo) recordActivity(ctx context.Context, activity *activities.CoreActivity) (dbActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.activities.recordActivity")
	defer span.End()

	q := `
		INSERT INTO story_activities (story_id, type, field, previous_value, current_value, created_at)
		VALUES (:story_id, :type, :field, :previous_value, :current_value, :created_at)
		RETURNING story_activities.*;
	`

	var da dbActivity
	stmt, err := r.db.PrepareNamedContext(ctx, q)
	if err != nil {
		errMsg := fmt.Sprintf("Failed to prepare named statement: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to prepare statement"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbActivity{}, err
	}
	defer stmt.Close()

	if err := stmt.GetContext(ctx, &da, toDBActivity(*activity)); err != nil {
		errMsg := fmt.Sprintf("Failed to insert activity: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to insert activity"), trace.WithAttributes(attribute.String("error", errMsg)))
		return dbActivity{}, err
	}

	r.log.Info(ctx, "Activity created successfully.")
	span.AddEvent("Activity created.", trace.WithAttributes(
		attribute.String("activity.story_id", activity.StoryID.String()),
	))

	return da, nil
}

// GetActivities returns all activities for a given story ID.
func (r *repo) GetActivities(ctx context.Context, storyID uuid.UUID) ([]activities.CoreActivity, error) {
	ctx, span := web.AddSpan(ctx, "business.repository.activities.GetActivities")
	defer span.End()

	q := `
		SELECT 
			activity_id,
			story_id,
			type,
			field,
			previous_value,
			current_value,
			created_at
		FROM story_activities
		WHERE story_id = :story_id
	`

	var activities []dbActivity
	if err := r.db.SelectContext(ctx, &activities, q, map[string]interface{}{"story_id": storyID}); err != nil {
		errMsg := fmt.Sprintf("Failed to get activities: %s", err)
		r.log.Error(ctx, errMsg)
		span.RecordError(errors.New("failed to get activities"), trace.WithAttributes(attribute.String("error", errMsg)))
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	return toCoreActivities(activities), nil
}
