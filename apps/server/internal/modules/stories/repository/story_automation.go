package storiesrepository

import (
	"context"
	"errors"
	"fmt"
	"time"

	storydomain "github.com/complexus-tech/projects-api/internal/modules/stories/domain"
	storyreadsql "github.com/complexus-tech/projects-api/internal/modules/stories/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

const (
	maximumStoryAutomationBatchSize = 1000

	storyAutoArchiveLockName     = "stories:auto-archive:v1"
	storyAutoCloseLockName       = "stories:auto-close:v1"
	storySprintMigrationLockName = "stories:sprint-migration:v1"

	storyAutoCloseActivityReason  = "Story automation moved this story to a closed state after it was inactive for the configured period."
	sprintMigrationActivityReason = "Sprint automation moved this story into the next sprint because incomplete work migration is enabled."
)

var errStoryAutomationRepositoryNotConfigured = errors.New("story automation repository is not configured")

// storyAutomationQueries is the complete SQL surface used by the three story
// automation transactions. Keeping it narrow makes transaction ordering and
// rollback behavior independently testable.
type storyAutomationQueries interface {
	LockStoryAutomation(context.Context, storyreadsql.LockStoryAutomationParams) error
	ArchiveEligibleStoriesBatch(
		context.Context,
		storyreadsql.ArchiveEligibleStoriesBatchParams,
	) ([]uuid.UUID, error)
	CloseEligibleStoriesBatch(
		context.Context,
		storyreadsql.CloseEligibleStoriesBatchParams,
	) ([]storyreadsql.CloseEligibleStoriesBatchRow, error)
	InsertStoryAutoCloseActivities(
		context.Context,
		storyreadsql.InsertStoryAutoCloseActivitiesParams,
	) (int64, error)
	MigrateEligibleSprintStoriesBatch(
		context.Context,
		storyreadsql.MigrateEligibleSprintStoriesBatchParams,
	) ([]storyreadsql.MigrateEligibleSprintStoriesBatchRow, error)
	InsertSprintMigrationActivities(
		context.Context,
		storyreadsql.InsertSprintMigrationActivitiesParams,
	) (int64, error)
	InsertSprintMigrationAuditEvents(
		context.Context,
		storyreadsql.InsertSprintMigrationAuditEventsParams,
	) (int64, error)
}

// ArchiveEligibleStoriesBatch archives one stable, bounded page in a single
// advisory-locked transaction.
func (r *repo) ArchiveEligibleStoriesBatch(
	ctx context.Context,
	batch storydomain.StoryAutoArchiveBatch,
) (storydomain.StoryAutoArchiveResult, error) {
	batchSize, asOf, err := r.validateStoryAutomationBatch(ctx, batch.AsOf, batch.BatchSize, false, uuid.Nil)
	if err != nil {
		return storydomain.StoryAutoArchiveResult{}, err
	}

	var result storydomain.StoryAutoArchiveResult
	err = r.runStoryAutomationTransaction(ctx, func(queries storyAutomationQueries) error {
		if lockErr := queries.LockStoryAutomation(ctx, storyreadsql.LockStoryAutomationParams{
			LockName: storyAutoArchiveLockName,
		}); lockErr != nil {
			return fmt.Errorf("lock story auto-archive: %w", lockErr)
		}

		storyIDs, queryErr := queries.ArchiveEligibleStoriesBatch(
			ctx,
			storyreadsql.ArchiveEligibleStoriesBatchParams{AsOf: asOf, BatchSize: batchSize},
		)
		if queryErr != nil {
			return fmt.Errorf("archive eligible stories: %w", queryErr)
		}
		if validateErr := validateStoryAutomationIDs(storyIDs, int(batchSize), "archive eligible stories"); validateErr != nil {
			return validateErr
		}
		result.Archived = len(storyIDs)
		return nil
	})
	if err != nil {
		return storydomain.StoryAutoArchiveResult{}, fmt.Errorf("archive eligible stories batch: %w", err)
	}
	return result, nil
}

// CloseEligibleStoriesBatch performs the status transition and writes every
// corresponding activity before the transaction is allowed to commit.
func (r *repo) CloseEligibleStoriesBatch(
	ctx context.Context,
	batch storydomain.StoryAutoCloseBatch,
) (storydomain.StoryAutoCloseResult, error) {
	batchSize, asOf, err := r.validateStoryAutomationBatch(
		ctx, batch.AsOf, batch.BatchSize, true, batch.SystemUserID,
	)
	if err != nil {
		return storydomain.StoryAutoCloseResult{}, err
	}

	var result storydomain.StoryAutoCloseResult
	err = r.runStoryAutomationTransaction(ctx, func(queries storyAutomationQueries) error {
		if lockErr := queries.LockStoryAutomation(ctx, storyreadsql.LockStoryAutomationParams{
			LockName: storyAutoCloseLockName,
		}); lockErr != nil {
			return fmt.Errorf("lock story auto-close: %w", lockErr)
		}

		rows, queryErr := queries.CloseEligibleStoriesBatch(
			ctx,
			storyreadsql.CloseEligibleStoriesBatchParams{AsOf: asOf, BatchSize: batchSize},
		)
		if queryErr != nil {
			return fmt.Errorf("close eligible stories: %w", queryErr)
		}
		params, mapErr := autoCloseActivityParams(rows, batch.SystemUserID, asOf, int(batchSize))
		if mapErr != nil {
			return mapErr
		}
		if len(rows) == 0 {
			result = storydomain.StoryAutoCloseResult{}
			return nil
		}

		inserted, insertErr := queries.InsertStoryAutoCloseActivities(ctx, params)
		if insertErr != nil {
			return fmt.Errorf("insert story auto-close activities: %w", insertErr)
		}
		if inserted != int64(len(rows)) {
			return fmt.Errorf(
				"insert story auto-close activities: inserted %d rows, want %d",
				inserted,
				len(rows),
			)
		}
		result = storydomain.StoryAutoCloseResult{
			Closed:             len(rows),
			ActivitiesRecorded: len(rows),
		}
		return nil
	})
	if err != nil {
		return storydomain.StoryAutoCloseResult{}, fmt.Errorf("close eligible stories batch: %w", err)
	}
	return result, nil
}

// MigrateEligibleSprintStoriesBatch moves incomplete work and records both
// its human-readable activity and durable audit event atomically.
func (r *repo) MigrateEligibleSprintStoriesBatch(
	ctx context.Context,
	batch storydomain.SprintStoryMigrationBatch,
) (storydomain.SprintStoryMigrationResult, error) {
	batchSize, asOf, err := r.validateStoryAutomationBatch(
		ctx, batch.AsOf, batch.BatchSize, true, batch.SystemUserID,
	)
	if err != nil {
		return storydomain.SprintStoryMigrationResult{}, err
	}

	var result storydomain.SprintStoryMigrationResult
	err = r.runStoryAutomationTransaction(ctx, func(queries storyAutomationQueries) error {
		if lockErr := queries.LockStoryAutomation(ctx, storyreadsql.LockStoryAutomationParams{
			LockName: storySprintMigrationLockName,
		}); lockErr != nil {
			return fmt.Errorf("lock sprint story migration: %w", lockErr)
		}

		rows, queryErr := queries.MigrateEligibleSprintStoriesBatch(
			ctx,
			storyreadsql.MigrateEligibleSprintStoriesBatchParams{AsOf: asOf, BatchSize: batchSize},
		)
		if queryErr != nil {
			return fmt.Errorf("migrate eligible sprint stories: %w", queryErr)
		}
		activityParams, auditParams, mapErr := sprintMigrationSideEffectParams(
			rows, batch.SystemUserID, asOf, int(batchSize),
		)
		if mapErr != nil {
			return mapErr
		}
		if len(rows) == 0 {
			result = storydomain.SprintStoryMigrationResult{}
			return nil
		}

		activities, activityErr := queries.InsertSprintMigrationActivities(ctx, activityParams)
		if activityErr != nil {
			return fmt.Errorf("insert sprint migration activities: %w", activityErr)
		}
		if activities != int64(len(rows)) {
			return fmt.Errorf(
				"insert sprint migration activities: inserted %d rows, want %d",
				activities,
				len(rows),
			)
		}

		auditEvents, auditErr := queries.InsertSprintMigrationAuditEvents(ctx, auditParams)
		if auditErr != nil {
			return fmt.Errorf("insert sprint migration audit events: %w", auditErr)
		}
		if auditEvents != int64(len(rows)) {
			return fmt.Errorf(
				"insert sprint migration audit events: inserted %d rows, want %d",
				auditEvents,
				len(rows),
			)
		}

		result = storydomain.SprintStoryMigrationResult{
			Migrated:            len(rows),
			ActivitiesRecorded:  len(rows),
			AuditEventsRecorded: len(rows),
		}
		return nil
	})
	if err != nil {
		return storydomain.SprintStoryMigrationResult{}, fmt.Errorf("migrate eligible sprint stories batch: %w", err)
	}
	return result, nil
}

func (r *repo) validateStoryAutomationBatch(
	ctx context.Context,
	asOf time.Time,
	batchSize int,
	requireActor bool,
	actorID uuid.UUID,
) (int32, time.Time, error) {
	if ctx == nil {
		return 0, time.Time{}, errors.New("story automation context is required")
	}
	if r == nil || r.runStoryAutomationTransaction == nil {
		return 0, time.Time{}, errStoryAutomationRepositoryNotConfigured
	}
	if asOf.IsZero() {
		return 0, time.Time{}, errors.New("story automation as-of time is required")
	}
	if requireActor && actorID == uuid.Nil {
		return 0, time.Time{}, errors.New("story automation system user is required")
	}
	if batchSize <= 0 || batchSize > maximumStoryAutomationBatchSize {
		return 0, time.Time{}, fmt.Errorf(
			"story automation batch size must be between 1 and %d",
			maximumStoryAutomationBatchSize,
		)
	}
	converted, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, time.Time{}, fmt.Errorf("convert story automation batch size: %w", err)
	}
	return converted, asOf.UTC(), nil
}

func autoCloseActivityParams(
	rows []storyreadsql.CloseEligibleStoriesBatchRow,
	systemUserID uuid.UUID,
	asOf time.Time,
	batchSize int,
) (storyreadsql.InsertStoryAutoCloseActivitiesParams, error) {
	if len(rows) > batchSize {
		return storyreadsql.InsertStoryAutoCloseActivitiesParams{}, fmt.Errorf(
			"close eligible stories: returned %d rows, want at most %d",
			len(rows),
			batchSize,
		)
	}

	reason := storyAutoCloseActivityReason
	params := storyreadsql.InsertStoryAutoCloseActivitiesParams{
		SystemUserID: systemUserID,
		Reason:       &reason,
		AsOf:         asOf,
		StoryIds:     make([]uuid.UUID, 0, len(rows)),
		WorkspaceIds: make([]uuid.UUID, 0, len(rows)),
		TeamIds:      make([]uuid.UUID, 0, len(rows)),
		StatusIds:    make([]uuid.UUID, 0, len(rows)),
	}
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row.ID == uuid.Nil || row.WorkspaceID == uuid.Nil || row.TeamID == uuid.Nil || row.StatusID == uuid.Nil {
			return storyreadsql.InsertStoryAutoCloseActivitiesParams{}, errors.New(
				"close eligible stories: database returned invalid tenant routing metadata",
			)
		}
		if _, duplicate := seen[row.ID]; duplicate {
			return storyreadsql.InsertStoryAutoCloseActivitiesParams{}, errors.New(
				"close eligible stories: database returned a duplicate story",
			)
		}
		seen[row.ID] = struct{}{}
		params.StoryIds = append(params.StoryIds, row.ID)
		params.WorkspaceIds = append(params.WorkspaceIds, row.WorkspaceID)
		params.TeamIds = append(params.TeamIds, row.TeamID)
		params.StatusIds = append(params.StatusIds, row.StatusID)
	}
	return params, nil
}

func sprintMigrationSideEffectParams(
	rows []storyreadsql.MigrateEligibleSprintStoriesBatchRow,
	systemUserID uuid.UUID,
	asOf time.Time,
	batchSize int,
) (
	storyreadsql.InsertSprintMigrationActivitiesParams,
	storyreadsql.InsertSprintMigrationAuditEventsParams,
	error,
) {
	if len(rows) > batchSize {
		return storyreadsql.InsertSprintMigrationActivitiesParams{},
			storyreadsql.InsertSprintMigrationAuditEventsParams{},
			fmt.Errorf("migrate eligible sprint stories: returned %d rows, want at most %d", len(rows), batchSize)
	}

	reason := sprintMigrationActivityReason
	activityParams := storyreadsql.InsertSprintMigrationActivitiesParams{
		SystemUserID:      systemUserID,
		Reason:            &reason,
		AsOf:              asOf,
		StoryIds:          make([]uuid.UUID, 0, len(rows)),
		WorkspaceIds:      make([]uuid.UUID, 0, len(rows)),
		TeamIds:           make([]uuid.UUID, 0, len(rows)),
		PreviousSprintIds: make([]uuid.UUID, 0, len(rows)),
		NewSprintIds:      make([]uuid.UUID, 0, len(rows)),
	}
	auditActor := systemUserID
	auditParams := storyreadsql.InsertSprintMigrationAuditEventsParams{
		SystemUserID:      &auditActor,
		AsOf:              asOf,
		StoryIds:          make([]uuid.UUID, 0, len(rows)),
		WorkspaceIds:      make([]uuid.UUID, 0, len(rows)),
		TeamIds:           make([]uuid.UUID, 0, len(rows)),
		PreviousSprintIds: make([]uuid.UUID, 0, len(rows)),
		NewSprintIds:      make([]uuid.UUID, 0, len(rows)),
	}
	seen := make(map[uuid.UUID]struct{}, len(rows))
	for _, row := range rows {
		if row.ID == uuid.Nil || row.WorkspaceID == uuid.Nil || row.TeamID == uuid.Nil ||
			row.PreviousSprintID == uuid.Nil || row.NewSprintID == uuid.Nil ||
			row.PreviousSprintID == row.NewSprintID {
			return storyreadsql.InsertSprintMigrationActivitiesParams{},
				storyreadsql.InsertSprintMigrationAuditEventsParams{},
				errors.New("migrate eligible sprint stories: database returned invalid tenant routing metadata")
		}
		if _, duplicate := seen[row.ID]; duplicate {
			return storyreadsql.InsertSprintMigrationActivitiesParams{},
				storyreadsql.InsertSprintMigrationAuditEventsParams{},
				errors.New("migrate eligible sprint stories: database returned a duplicate story")
		}
		seen[row.ID] = struct{}{}

		activityParams.StoryIds = append(activityParams.StoryIds, row.ID)
		activityParams.WorkspaceIds = append(activityParams.WorkspaceIds, row.WorkspaceID)
		activityParams.TeamIds = append(activityParams.TeamIds, row.TeamID)
		activityParams.PreviousSprintIds = append(activityParams.PreviousSprintIds, row.PreviousSprintID)
		activityParams.NewSprintIds = append(activityParams.NewSprintIds, row.NewSprintID)

		auditParams.StoryIds = append(auditParams.StoryIds, row.ID)
		auditParams.WorkspaceIds = append(auditParams.WorkspaceIds, row.WorkspaceID)
		auditParams.TeamIds = append(auditParams.TeamIds, row.TeamID)
		auditParams.PreviousSprintIds = append(auditParams.PreviousSprintIds, row.PreviousSprintID)
		auditParams.NewSprintIds = append(auditParams.NewSprintIds, row.NewSprintID)
	}
	return activityParams, auditParams, nil
}

func validateStoryAutomationIDs(ids []uuid.UUID, batchSize int, operation string) error {
	if len(ids) > batchSize {
		return fmt.Errorf("%s: returned %d rows, want at most %d", operation, len(ids), batchSize)
	}
	seen := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			return fmt.Errorf("%s: database returned an empty story ID", operation)
		}
		if _, duplicate := seen[id]; duplicate {
			return fmt.Errorf("%s: database returned a duplicate story ID", operation)
		}
		seen[id] = struct{}{}
	}
	return nil
}
