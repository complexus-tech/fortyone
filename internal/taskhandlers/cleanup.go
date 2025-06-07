package taskhandlers

import (
	"context"
	"fmt"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jmoiron/sqlx"
)

// CleanupHandlers handles cleanup tasks with database access
type CleanupHandlers struct {
	log          *logger.Logger
	db           *sqlx.DB
	systemUserID uuid.UUID
}

// NewCleanupHandlers creates a new CleanupHandlers instance
func NewCleanupHandlers(log *logger.Logger, db *sqlx.DB, systemUserID uuid.UUID) *CleanupHandlers {
	return &CleanupHandlers{
		log:          log,
		db:           db,
		systemUserID: systemUserID,
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

// HandleStoryAutoArchive processes the story auto-archive task
func (c *CleanupHandlers) HandleStoryAutoArchive(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing StoryAutoArchive task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.ProcessStoryAutoArchive(ctx, c.db, c.log); err != nil {
		c.log.Error(ctx, "Failed to process story auto-archive", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("story auto-archive failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed StoryAutoArchive task", "task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleStoryAutoClose processes the story auto-close task
func (c *CleanupHandlers) HandleStoryAutoClose(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing StoryAutoClose task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.ProcessStoryAutoClose(ctx, c.db, c.log, c.systemUserID); err != nil {
		c.log.Error(ctx, "Failed to process story auto-close", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("story auto-close failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed StoryAutoClose task", "task_id", t.ResultWriter().TaskID())
	return nil
}

// HandleSprintStoryMigration processes the sprint story migration task
func (c *CleanupHandlers) HandleSprintStoryMigration(ctx context.Context, t *asynq.Task) error {
	c.log.Info(ctx, "HANDLER: Processing SprintStoryMigration task", "task_id", t.ResultWriter().TaskID())

	if err := jobs.ProcessSprintStoryMigration(ctx, c.db, c.log, c.systemUserID); err != nil {
		c.log.Error(ctx, "Failed to process sprint story migration", "error", err, "task_id", t.ResultWriter().TaskID())
		return fmt.Errorf("sprint story migration failed: %w", err)
	}

	c.log.Info(ctx, "HANDLER: Successfully processed SprintStoryMigration task", "task_id", t.ResultWriter().TaskID())
	return nil
}
