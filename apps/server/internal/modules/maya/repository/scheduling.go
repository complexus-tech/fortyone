package mayarepository

import (
	"context"
	"fmt"
	"time"

	mayadomain "github.com/complexus-tech/projects-api/internal/modules/maya/domain"
	mayasql "github.com/complexus-tech/projects-api/internal/modules/maya/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *Repo) WithScheduleStoryLock(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
	reconcile func() error,
) error {
	if workspaceID == uuid.Nil || storyID == uuid.Nil || reconcile == nil {
		return mayadomain.ErrInvalidPlanInput
	}
	if err := r.configured(); err != nil {
		return err
	}

	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(queries mayasql.Querier) error {
		if err := queries.LockMayaStorySchedule(ctx, mayasql.LockMayaStoryScheduleParams{
			WorkspaceID: workspaceID,
			StoryID:     storyID,
		}); err != nil {
			return fmt.Errorf("lock Maya story schedule: %w", err)
		}
		return reconcile()
	})
	if err != nil {
		return fmt.Errorf("run Maya story schedule lock transaction: %w", err)
	}
	return nil
}

func (r *Repo) ListScheduleStoryRefsForUser(
	ctx context.Context,
	userID uuid.UUID,
) ([]mayadomain.ScheduleStoryRef, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	rows, err := r.queries.ListScheduleStoryRefsForUser(ctx, mayasql.ListScheduleStoryRefsForUserParams{
		UserID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("list stories for schedule reconciliation: %w", err)
	}
	refs := make([]mayadomain.ScheduleStoryRef, len(rows))
	for index, row := range rows {
		refs[index] = mayadomain.ScheduleStoryRef{
			WorkspaceID: row.WorkspaceID,
			StoryID:     row.StoryID,
		}
	}
	return refs, nil
}

// ClaimScheduleRecoveryStoryRefs leases inconsistent durable Maya ownerships
// without advancing their last-successful-reconciliation watermark. Failed
// rows therefore remain discoverable after the bounded retry delay.
func (r *Repo) ClaimScheduleRecoveryStoryRefs(
	ctx context.Context,
	limit int,
	retryBefore, interruptedRunBefore time.Time,
) ([]mayadomain.ScheduleRecoveryRef, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, fmt.Errorf("validate Maya schedule recovery limit: %w", err)
	}
	retryBefore = retryBefore.UTC()

	var rows []mayasql.ClaimScheduleRecoveryStoryRefsRow
	err = r.withinTransaction(ctx, pgx.TxOptions{}, func(queries mayasql.Querier) error {
		var queryErr error
		rows, queryErr = queries.ClaimScheduleRecoveryStoryRefs(ctx, mayasql.ClaimScheduleRecoveryStoryRefsParams{
			InterruptedRunBefore: interruptedRunBefore.UTC(),
			RetryBefore:          &retryBefore,
			RowLimit:             rowLimit,
		})
		if queryErr != nil {
			return fmt.Errorf("claim Maya schedule recovery stories: %w", queryErr)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("claim Maya schedule recovery transaction: %w", err)
	}

	refs := make([]mayadomain.ScheduleRecoveryRef, len(rows))
	for index, row := range rows {
		refs[index] = mayadomain.ScheduleRecoveryRef{
			ScheduleStoryRef: mayadomain.ScheduleStoryRef{
				WorkspaceID: row.WorkspaceID,
				StoryID:     row.StoryID,
			},
			InterruptedRunID: row.InterruptedRunID,
		}
	}
	return refs, nil
}

func (r *Repo) CompleteInterruptedScheduleRun(ctx context.Context, runID uuid.UUID, message string) error {
	if runID == uuid.Nil {
		return nil
	}
	if err := r.configured(); err != nil {
		return err
	}

	err := r.withinTransaction(ctx, pgx.TxOptions{}, func(queries mayasql.Querier) error {
		params := mayasql.FailInterruptedScheduleActionsParams{
			ErrorMessage: &message,
			RunID:        runID,
		}
		if _, err := queries.FailInterruptedScheduleActions(ctx, params); err != nil {
			return fmt.Errorf("fail interrupted Maya actions: %w", err)
		}
		if _, err := queries.CompleteInterruptedScheduleRun(ctx, mayasql.CompleteInterruptedScheduleRunParams{
			ErrorMessage: &message,
			RunID:        runID,
		}); err != nil {
			return fmt.Errorf("complete interrupted Maya run: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("complete interrupted Maya run transaction: %w", err)
	}
	return nil
}

func (r *Repo) ListMayaScheduleOwners(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) ([]uuid.UUID, error) {
	if err := r.configured(); err != nil {
		return nil, err
	}
	owners, err := r.queries.ListMayaScheduleOwners(ctx, mayasql.ListMayaScheduleOwnersParams{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return nil, fmt.Errorf("list Maya schedule owners: %w", err)
	}
	return owners, nil
}

func (r *Repo) StoryIsSchedulableForUser(
	ctx context.Context,
	workspaceID, storyID, userID uuid.UUID,
) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	schedulable, err := r.queries.StoryIsSchedulableForUser(ctx, mayasql.StoryIsSchedulableForUserParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return false, fmt.Errorf("check story schedule eligibility: %w", err)
	}
	return schedulable, nil
}

func (r *Repo) WorkspaceCanUseMaya(ctx context.Context, workspaceID uuid.UUID) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	hasAccess, err := r.queries.WorkspaceCanUseMaya(ctx, mayasql.WorkspaceCanUseMayaParams{
		WorkspaceID: workspaceID,
	})
	if err != nil {
		return false, fmt.Errorf("check Maya workspace access: %w", err)
	}
	return hasAccess, nil
}

func (r *Repo) StoryIsActiveForAutoScheduling(
	ctx context.Context,
	workspaceID, storyID uuid.UUID,
) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	active, err := r.queries.StoryIsActiveForAutoScheduling(ctx, mayasql.StoryIsActiveForAutoSchedulingParams{
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return false, fmt.Errorf("check story auto-scheduling lifecycle: %w", err)
	}
	return active, nil
}

// StoryScheduleOwnershipIsRetainable distinguishes temporary placement gaps
// from lifecycle states that must permanently retire Maya's ownership. An
// unassigned active story remains enrolled so a later assignment can resume
// scheduling, but terminal stories and removed or inactive owners do not.
func (r *Repo) StoryScheduleOwnershipIsRetainable(
	ctx context.Context,
	workspaceID, storyID, userID uuid.UUID,
) (bool, error) {
	if err := r.configured(); err != nil {
		return false, err
	}
	retainable, err := r.queries.StoryScheduleOwnershipIsRetainable(ctx, mayasql.StoryScheduleOwnershipIsRetainableParams{
		UserID:      userID,
		WorkspaceID: workspaceID,
		StoryID:     storyID,
	})
	if err != nil {
		return false, fmt.Errorf("check Maya schedule ownership retention: %w", err)
	}
	return retainable, nil
}
