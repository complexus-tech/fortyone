package taskhandlers

import (
	"context"

	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	"github.com/complexus-tech/projects-api/pkg/emailcopy"
	"github.com/complexus-tech/projects-api/pkg/emailthread"
	"github.com/complexus-tech/projects-api/pkg/jobs"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/mailer"
	"github.com/hibiken/asynq"
)

// GuidanceHandlerDependencies contains the typed read capabilities and
// delivery services used by scheduled deadline-guidance emails.
type GuidanceHandlerDependencies struct {
	Log               *logger.Logger
	Objectives        jobs.ObjectiveOverdueStore
	Stories           jobs.OverdueStoryStore
	WeeklyDigest      jobs.WeeklyDigestStore
	FeedbackDigest    feedback.DigestStore
	Mailer            mailer.Service
	CopyGenerator     emailcopy.Generator
	ThreadPreparation emailthread.GuidancePreparer
}

// GuidanceHandlers owns scheduled, user-facing deadline guidance. It remains
// separate from destructive retention and state-transition handlers.
type GuidanceHandlers struct {
	log               *logger.Logger
	objectives        jobs.ObjectiveOverdueStore
	stories           jobs.OverdueStoryStore
	weeklyDigest      jobs.WeeklyDigestStore
	feedbackDigest    feedback.DigestStore
	mailer            mailer.Service
	copyGenerator     emailcopy.Generator
	threadPreparation emailthread.GuidancePreparer
}

func NewGuidanceHandlers(dependencies GuidanceHandlerDependencies) *GuidanceHandlers {
	return &GuidanceHandlers{
		log:               dependencies.Log,
		objectives:        dependencies.Objectives,
		stories:           dependencies.Stories,
		weeklyDigest:      dependencies.WeeklyDigest,
		feedbackDigest:    dependencies.FeedbackDigest,
		mailer:            dependencies.Mailer,
		copyGenerator:     dependencies.CopyGenerator,
		threadPreparation: dependencies.ThreadPreparation,
	}
}

func (h *GuidanceHandlers) HandleOverdueStoriesEmail(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "OverdueStoriesEmail", "overdue stories email", func(ctx context.Context) error {
		return jobs.ProcessOverdueStoriesEmail(
			ctx,
			h.stories,
			h.log,
			h.mailer,
			h.copyGenerator,
			h.threadPreparation,
		)
	})
}

func (h *GuidanceHandlers) HandleObjectiveOverdueEmail(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "ObjectiveOverdueEmail", "objective overdue email", func(ctx context.Context) error {
		return jobs.ProcessObjectiveOverdue(
			ctx,
			h.objectives,
			h.log,
			h.mailer,
			h.copyGenerator,
			h.threadPreparation,
		)
	})
}

func (h *GuidanceHandlers) HandleWeeklyDigestEmail(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "WeeklyDigestEmail", "weekly digest email", func(ctx context.Context) error {
		return jobs.ProcessWeeklyDigestEmail(
			ctx,
			h.weeklyDigest,
			h.log,
			h.mailer,
			h.copyGenerator,
			h.threadPreparation,
		)
	})
}

func (h *GuidanceHandlers) HandleFeedbackDigestEmail(ctx context.Context, task *asynq.Task) error {
	return h.handle(ctx, task, "FeedbackDigestEmail", "feedback digest email", func(ctx context.Context) error {
		return jobs.ProcessFeedbackDigestEmail(
			ctx,
			h.feedbackDigest,
			h.log,
			h.mailer,
			h.copyGenerator,
			h.threadPreparation,
		)
	})
}

func (h *GuidanceHandlers) handle(
	ctx context.Context,
	task *asynq.Task,
	taskName string,
	operationName string,
	operation func(context.Context) error,
) error {
	if h == nil {
		return runScheduledTask(ctx, task, nil, "guidance", taskName, operationName, nil)
	}
	return runScheduledTask(ctx, task, h.log, "guidance", taskName, operationName, operation)
}
