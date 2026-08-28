package jobs

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	storyAutoArchiveBatchSize     = 1000
	storyAutoCloseBatchSize       = 1000
	sprintStoryMigrationBatchSize = 500
	storyAutomationMaxBatches     = 100
)

var errStoryAutomationBacklogRemaining = errors.New("story automation backlog remains")

// StoryAutoArchiveStore is the single persistence capability needed by the
// auto-archive job.
type StoryAutoArchiveStore interface {
	ArchiveEligibleStoriesBatch(
		context.Context,
		storydomain.StoryAutoArchiveBatch,
	) (storydomain.StoryAutoArchiveResult, error)
}

// StoryAutoCloseStore is the single atomic transition-and-activity capability
// needed by the auto-close job.
type StoryAutoCloseStore interface {
	CloseEligibleStoriesBatch(
		context.Context,
		storydomain.StoryAutoCloseBatch,
	) (storydomain.StoryAutoCloseResult, error)
}

// SprintStoryMigrationStore atomically owns the sprint transition, activity,
// and audit writes used by the migration job.
type SprintStoryMigrationStore interface {
	MigrateEligibleSprintStoriesBatch(
		context.Context,
		storydomain.SprintStoryMigrationBatch,
	) (storydomain.SprintStoryMigrationResult, error)
}

// StoryAutomationStore is the composition-root convenience contract. Each job
// still accepts only the narrower capability it actually uses.
type StoryAutomationStore interface {
	StoryAutoArchiveStore
	StoryAutoCloseStore
	SprintStoryMigrationStore
}

// ProcessStoryAutoArchive archives eligible completed and cancelled stories in
// bounded transactions evaluated against one application-owned UTC clock.
func ProcessStoryAutoArchive(ctx context.Context, store StoryAutoArchiveStore, log *logger.Logger) error {
	return processStoryAutoArchiveAt(ctx, store, log, time.Now().UTC())
}

func processStoryAutoArchiveAt(
	ctx context.Context,
	store StoryAutoArchiveStore,
	log *logger.Logger,
	asOf time.Time,
) error {
	if err := validateStoryAutomationDependencies(ctx, store, log, asOf, "auto-archive"); err != nil {
		return err
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStoryAutoArchive")
	defer span.End()
	asOf = asOf.UTC()

	log.Info(ctx, "Processing story auto-archive", "as_of", asOf)
	summary, err := runStoryAutomationBatches(
		ctx,
		"story auto-archive",
		storyAutoArchiveBatchSize,
		storyAutomationSideEffectsNone,
		func(batchCtx context.Context, batchSize int) (storyAutomationBatchResult, error) {
			result, batchErr := store.ArchiveEligibleStoriesBatch(batchCtx, storydomain.StoryAutoArchiveBatch{
				AsOf: asOf, BatchSize: batchSize,
			})
			return storyAutomationBatchResult{Stories: result.Archived}, batchErr
		},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	recordStoryAutomationCompletion(ctx, span, log, "story auto-archive", summary)
	return nil
}

// ProcessStoryAutoClose moves eligible inactive stories to a cancelled status
// and records their activities atomically.
func ProcessStoryAutoClose(
	ctx context.Context,
	store StoryAutoCloseStore,
	log *logger.Logger,
	systemUserID uuid.UUID,
) error {
	return processStoryAutoCloseAt(ctx, store, log, systemUserID, time.Now().UTC())
}

func processStoryAutoCloseAt(
	ctx context.Context,
	store StoryAutoCloseStore,
	log *logger.Logger,
	systemUserID uuid.UUID,
	asOf time.Time,
) error {
	if err := validateStoryAutomationDependencies(ctx, store, log, asOf, "auto-close"); err != nil {
		return err
	}
	if systemUserID == uuid.Nil {
		return errors.New("story auto-close system user is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessStoryAutoClose")
	defer span.End()
	asOf = asOf.UTC()

	log.Info(ctx, "Processing story auto-close", "as_of", asOf)
	summary, err := runStoryAutomationBatches(
		ctx,
		"story auto-close",
		storyAutoCloseBatchSize,
		storyAutomationSideEffectsActivity,
		func(batchCtx context.Context, batchSize int) (storyAutomationBatchResult, error) {
			result, batchErr := store.CloseEligibleStoriesBatch(batchCtx, storydomain.StoryAutoCloseBatch{
				AsOf: asOf, SystemUserID: systemUserID, BatchSize: batchSize,
			})
			return storyAutomationBatchResult{
				Stories: result.Closed, Activities: result.ActivitiesRecorded,
			}, batchErr
		},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	recordStoryAutomationCompletion(ctx, span, log, "story auto-close", summary)
	return nil
}

// ProcessSprintStoryMigration moves incomplete stories from yesterday's ended
// sprints into the next scoped sprint and records activity plus audit evidence.
func ProcessSprintStoryMigration(
	ctx context.Context,
	store SprintStoryMigrationStore,
	log *logger.Logger,
	systemUserID uuid.UUID,
) error {
	return processSprintStoryMigrationAt(ctx, store, log, systemUserID, time.Now().UTC())
}

func processSprintStoryMigrationAt(
	ctx context.Context,
	store SprintStoryMigrationStore,
	log *logger.Logger,
	systemUserID uuid.UUID,
	asOf time.Time,
) error {
	if err := validateStoryAutomationDependencies(ctx, store, log, asOf, "sprint story migration"); err != nil {
		return err
	}
	if systemUserID == uuid.Nil {
		return errors.New("sprint story migration system user is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessSprintStoryMigration")
	defer span.End()
	asOf = asOf.UTC()

	log.Info(ctx, "Processing sprint story migration", "as_of", asOf)
	summary, err := runStoryAutomationBatches(
		ctx,
		"sprint story migration",
		sprintStoryMigrationBatchSize,
		storyAutomationSideEffectsActivityAndAudit,
		func(batchCtx context.Context, batchSize int) (storyAutomationBatchResult, error) {
			result, batchErr := store.MigrateEligibleSprintStoriesBatch(
				batchCtx,
				storydomain.SprintStoryMigrationBatch{
					AsOf: asOf, SystemUserID: systemUserID, BatchSize: batchSize,
				},
			)
			return storyAutomationBatchResult{
				Stories:     result.Migrated,
				Activities:  result.ActivitiesRecorded,
				AuditEvents: result.AuditEventsRecorded,
			}, batchErr
		},
	)
	if err != nil {
		span.RecordError(err)
		return err
	}

	recordStoryAutomationCompletion(ctx, span, log, "sprint story migration", summary)
	return nil
}

type storyAutomationSideEffects uint8

const (
	storyAutomationSideEffectsNone storyAutomationSideEffects = iota
	storyAutomationSideEffectsActivity
	storyAutomationSideEffectsActivityAndAudit
)

type storyAutomationBatchResult struct {
	Stories     int
	Activities  int
	AuditEvents int
}

type storyAutomationRunSummary struct {
	Stories     int64
	Activities  int64
	AuditEvents int64
	Batches     int
}

type storyAutomationBatchProcessor func(context.Context, int) (storyAutomationBatchResult, error)

func runStoryAutomationBatches(
	ctx context.Context,
	operation string,
	batchSize int,
	sideEffects storyAutomationSideEffects,
	process storyAutomationBatchProcessor,
) (storyAutomationRunSummary, error) {
	var summary storyAutomationRunSummary
	for batchIndex := 0; batchIndex < storyAutomationMaxBatches; batchIndex++ {
		if err := ctx.Err(); err != nil {
			return storyAutomationRunSummary{}, fmt.Errorf(
				"%s cancelled after %d stories: %w", operation, summary.Stories, err,
			)
		}

		result, err := process(ctx, batchSize)
		if err != nil {
			return storyAutomationRunSummary{}, fmt.Errorf("%s batch %d: %w", operation, batchIndex+1, err)
		}
		if err := validateStoryAutomationBatchResult(result, batchSize, sideEffects); err != nil {
			return storyAutomationRunSummary{}, fmt.Errorf("%s batch %d: %w", operation, batchIndex+1, err)
		}

		summary.Stories += int64(result.Stories)
		summary.Activities += int64(result.Activities)
		summary.AuditEvents += int64(result.AuditEvents)
		if result.Stories > 0 {
			summary.Batches++
		}
		if result.Stories < batchSize {
			return summary, nil
		}
	}

	return storyAutomationRunSummary{}, fmt.Errorf(
		"%s after %d stories: %w", operation, summary.Stories, errStoryAutomationBacklogRemaining,
	)
}

func validateStoryAutomationBatchResult(
	result storyAutomationBatchResult,
	batchSize int,
	sideEffects storyAutomationSideEffects,
) error {
	if result.Stories < 0 || result.Stories > batchSize {
		return fmt.Errorf("invalid story count %d, want 0..%d", result.Stories, batchSize)
	}
	if result.Activities < 0 || result.AuditEvents < 0 {
		return fmt.Errorf(
			"invalid side-effect counts activities=%d audit_events=%d",
			result.Activities,
			result.AuditEvents,
		)
	}

	wantActivities, wantAuditEvents := 0, 0
	if sideEffects >= storyAutomationSideEffectsActivity {
		wantActivities = result.Stories
	}
	if sideEffects == storyAutomationSideEffectsActivityAndAudit {
		wantAuditEvents = result.Stories
	}
	if result.Activities != wantActivities || result.AuditEvents != wantAuditEvents {
		return fmt.Errorf(
			"incomplete side effects stories=%d activities=%d audit_events=%d",
			result.Stories,
			result.Activities,
			result.AuditEvents,
		)
	}
	return nil
}

func validateStoryAutomationDependencies(
	ctx context.Context,
	store any,
	log *logger.Logger,
	asOf time.Time,
	operation string,
) error {
	if ctx == nil {
		return fmt.Errorf("story %s context is required", operation)
	}
	if store == nil {
		return fmt.Errorf("story %s store is required", operation)
	}
	if log == nil {
		return fmt.Errorf("story %s logger is required", operation)
	}
	if asOf.IsZero() {
		return fmt.Errorf("story %s as-of time is required", operation)
	}
	return nil
}

func recordStoryAutomationCompletion(
	ctx context.Context,
	span trace.Span,
	log *logger.Logger,
	operation string,
	summary storyAutomationRunSummary,
) {
	span.AddEvent(operation+" completed", trace.WithAttributes(
		attribute.Int64("stories.processed", summary.Stories),
		attribute.Int64("activities.recorded", summary.Activities),
		attribute.Int64("audit_events.recorded", summary.AuditEvents),
		attribute.Int("batches.processed", summary.Batches),
	))
	log.Info(
		ctx,
		operation+" completed",
		"stories_processed", summary.Stories,
		"activities_recorded", summary.Activities,
		"audit_events_recorded", summary.AuditEvents,
		"batches_processed", summary.Batches,
	)
}
