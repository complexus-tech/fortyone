package taskhandlers

import (
	"context"

	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/hibiken/asynq"
)

// StoryRetentionHandlerDependencies contains the database and object-store
// capabilities required by scheduled story retention. Neither capability
// exposes a database driver or storage credentials to the handler.
type StoryRetentionHandlerDependencies struct {
	Log     *logger.Logger
	Store   jobs.DeletedStoryRetentionStore
	Objects jobs.RetainedAttachmentObjectStore
}

// StoryRetentionHandlers owns the daily database purge and the independent,
// minute-cadence delivery of durable attachment-object deletions.
type StoryRetentionHandlers struct {
	log     *logger.Logger
	store   jobs.DeletedStoryRetentionStore
	objects jobs.RetainedAttachmentObjectStore
}

// NewStoryRetentionHandlers creates the story retention handler group.
func NewStoryRetentionHandlers(dependencies StoryRetentionHandlerDependencies) *StoryRetentionHandlers {
	return &StoryRetentionHandlers{
		log:     dependencies.Log,
		store:   dependencies.Store,
		objects: dependencies.Objects,
	}
}

// HandleDeletedStories retires expired soft-deleted stories transactionally.
func (h *StoryRetentionHandlers) HandleDeletedStories(ctx context.Context, task *asynq.Task) error {
	return h.handle(
		ctx,
		task,
		"DeleteStories",
		"deleted story retention",
		func(ctx context.Context) error {
			return jobs.PurgeDeletedStories(ctx, h.store, h.objects, h.log)
		},
	)
}

// HandleAttachmentObjectDeletions delivers fenced object-store deletions and
// retains no object names in handler logs or task payloads.
func (h *StoryRetentionHandlers) HandleAttachmentObjectDeletions(
	ctx context.Context,
	task *asynq.Task,
) error {
	return h.handle(
		ctx,
		task,
		"AttachmentObjectDeletions",
		"attachment object deletions",
		func(ctx context.Context) error {
			return jobs.ProcessAttachmentObjectDeletions(ctx, h.store, h.objects, h.log)
		},
	)
}

func (h *StoryRetentionHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "story retention", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "story retention", taskName, operationName, operation)
}
