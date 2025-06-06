package taskhandlers

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

// CleanupHandlers handles cleanup tasks with database access
type CleanupHandlers struct {
	log *logger.Logger
	db  *sqlx.DB
}

// NewCleanupHandlers creates a new CleanupHandlers instance
func NewCleanupHandlers(log *logger.Logger, db *sqlx.DB) *CleanupHandlers {
	return &CleanupHandlers{
		log: log,
		db:  db,
	}
}

// HandleTokenCleanup processes the token cleanup task
func (c *CleanupHandlers) HandleTokenCleanup(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing TokenCleanup task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.PurgeExpiredTokens(ctx, c.db, c.log); err != nil {
		c.log.Error(ctx, "Failed to purge expired tokens", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("token cleanup failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed TokenCleanup task", "task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleDeleteStories processes the delete stories cleanup task
func (c *CleanupHandlers) HandleDeleteStories(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing DeleteStories task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.PurgeDeletedStories(ctx, c.db, c.log); err != nil {
		c.log.Error(ctx, "Failed to purge deleted stories", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("delete stories cleanup failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed DeleteStories task", "task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleWebhookCleanup processes the webhook cleanup task
func (c *CleanupHandlers) HandleWebhookCleanup(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing WebhookCleanup task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.PurgeOldStripeWebhookEvents(ctx, c.db, c.log); err != nil {
		c.log.Error(ctx, "Failed to purge old webhook events", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("webhook cleanup failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed WebhookCleanup task", "task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleSprintAutoCreation processes the sprint auto-creation task
func (c *CleanupHandlers) HandleSprintAutoCreation(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing SprintAutoCreation task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.ProcessSprintAutoCreation(ctx, c.db, c.log); err != nil {
		c.log.Error(ctx, "Failed to process sprint auto-creation", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("sprint auto-creation failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed SprintAutoCreation task", "task_id", t.ResultWriter().TaskID())
	return nil
}
