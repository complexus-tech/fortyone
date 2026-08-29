package jobs

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"time"

	teamsettingsdomain "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	sprintAutomationBatchSize         = 100
	sprintAutomationMaxBatches        = 100
	sprintAutomationTeamMaxAttempts   = 3
	sprintAutomationRetryDelay        = 10 * time.Millisecond
	sprintAutomationInactivityDays    = 90
	sprintAutomationSettingsGraceDays = 30
	sprintAutomationDisabledReason    = "No human sprint planning activity in the last 90 days"
)

var errSprintAutomationBacklogRemaining = errors.New("sprint automation backlog remains")

// SprintAutomationStore is the worker-owned persistence capability. The
// adapter pages durable settings and owns each team's transactional locks,
// schedule reconciliation, creation, counters, and audit writes.
type SprintAutomationStore interface {
	ListSprintAutomationTeams(
		context.Context,
		teamsettingsdomain.SprintAutomationQuery,
	) ([]teamsettingsdomain.SprintAutomationTeamRef, error)
	RunSprintAutomationForTeam(
		context.Context,
		teamsettingsdomain.SprintAutomationTeamRef,
		time.Time,
	) (teamsettingsdomain.SprintAutomationRunResult, error)
	ListSprintAutomationInactivityCandidates(
		context.Context,
		teamsettingsdomain.SprintAutomationInactivityQuery,
	) ([]teamsettingsdomain.SprintAutomationTeamRef, error)
	DisableSprintAutomationIfInactive(
		context.Context,
		teamsettingsdomain.SprintAutomationInactivityEligibility,
	) (bool, error)
}

// ProcessSprintAutoCreation reconciles and replenishes the bounded set of
// teams whose durable settings currently enable sprint automation.
func ProcessSprintAutoCreation(
	ctx context.Context,
	store SprintAutomationStore,
	log *logger.Logger,
) error {
	return processSprintAutoCreationAt(ctx, store, log, time.Now().UTC())
}

func processSprintAutoCreationAt(
	ctx context.Context,
	store SprintAutomationStore,
	log *logger.Logger,
	asOf time.Time,
) error {
	if ctx == nil {
		return errors.New("sprint automation context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.ProcessSprintAutoCreation")
	defer span.End()
	if store == nil {
		return errors.New("sprint automation store is required")
	}
	if log == nil {
		return errors.New("sprint automation logger is required")
	}
	if asOf.IsZero() {
		return errors.New("sprint automation as-of time is required")
	}
	asOf = asOf.UTC()

	log.Info(ctx, "Processing sprint auto-creation for teams")
	var cursor teamsettingsdomain.SprintAutomationCursor
	var processingErrors []error
	teamsProcessed := 0
	teamsWithNewSprints := 0
	sprintsCreated := 0
	sprintsRescheduled := 0

	for batch := 0; batch < sprintAutomationMaxBatches; batch++ {
		if err := ctx.Err(); err != nil {
			return errors.Join(joinSprintAutomationErrors(processingErrors), fmt.Errorf("process sprint automation: %w", err))
		}
		teams, err := store.ListSprintAutomationTeams(
			ctx,
			teamsettingsdomain.SprintAutomationQuery{Cursor: cursor, BatchSize: sprintAutomationBatchSize},
		)
		if err != nil {
			return errors.Join(joinSprintAutomationErrors(processingErrors), fmt.Errorf("list sprint automation teams: %w", err))
		}
		if len(teams) > sprintAutomationBatchSize {
			return errors.Join(
				joinSprintAutomationErrors(processingErrors),
				fmt.Errorf("sprint automation page exceeds limit: got %d, limit %d", len(teams), sprintAutomationBatchSize),
			)
		}
		if len(teams) == 0 {
			return finishSprintAutomationRun(
				ctx, span, log, processingErrors,
				teamsProcessed, teamsWithNewSprints, sprintsCreated, sprintsRescheduled,
			)
		}

		for _, team := range teams {
			result, err := runSprintAutomationTeamWithRetry(ctx, store, team, asOf)
			if err != nil {
				log.Error(ctx, "Failed to run sprint automation for team",
					"error", err,
					"team_id", team.TeamID,
					"workspace_id", team.WorkspaceID)
				processingErrors = append(processingErrors, fmt.Errorf("team %s in workspace %s: %w", team.TeamID, team.WorkspaceID, err))
				continue
			}
			teamsProcessed++
			sprintsCreated += result.Created
			sprintsRescheduled += result.Rescheduled
			if result.Created > 0 {
				teamsWithNewSprints++
			}
		}

		last := teams[len(teams)-1]
		nextCursor := teamsettingsdomain.SprintAutomationCursor{
			WorkspaceID: last.WorkspaceID,
			TeamID:      last.TeamID,
			Valid:       true,
		}
		if cursor.Valid && !sprintAutomationCursorAdvances(cursor, nextCursor) {
			return errors.Join(joinSprintAutomationErrors(processingErrors), errors.New("sprint automation cursor did not advance"))
		}
		cursor = nextCursor
		span.AddEvent("sprint automation batch processed", trace.WithAttributes(
			attribute.Int("batch", batch+1),
			attribute.Int("teams.scanned", len(teams)),
		))
		if len(teams) < sprintAutomationBatchSize {
			return finishSprintAutomationRun(
				ctx, span, log, processingErrors,
				teamsProcessed, teamsWithNewSprints, sprintsCreated, sprintsRescheduled,
			)
		}
	}

	return errors.Join(
		joinSprintAutomationErrors(processingErrors),
		fmt.Errorf("process sprint automation after %d teams: %w", teamsProcessed, errSprintAutomationBacklogRemaining),
	)
}

func runSprintAutomationTeamWithRetry(
	ctx context.Context,
	store SprintAutomationStore,
	team teamsettingsdomain.SprintAutomationTeamRef,
	asOf time.Time,
) (teamsettingsdomain.SprintAutomationRunResult, error) {
	for attempt := 1; attempt <= sprintAutomationTeamMaxAttempts; attempt++ {
		result, err := store.RunSprintAutomationForTeam(ctx, team, asOf)
		if err == nil {
			return result, nil
		}
		if !errors.Is(err, teamsettingsdomain.ErrConcurrentUpdate) || attempt == sprintAutomationTeamMaxAttempts {
			return teamsettingsdomain.SprintAutomationRunResult{}, err
		}
		if err := waitForSprintAutomationRetry(ctx); err != nil {
			return teamsettingsdomain.SprintAutomationRunResult{}, err
		}
	}
	return teamsettingsdomain.SprintAutomationRunResult{}, teamsettingsdomain.ErrConcurrentUpdate
}

func finishSprintAutomationRun(
	ctx context.Context,
	span trace.Span,
	log *logger.Logger,
	processingErrors []error,
	teamsProcessed, teamsWithNewSprints, sprintsCreated, sprintsRescheduled int,
) error {
	span.AddEvent("sprint auto-creation completed", trace.WithAttributes(
		attribute.Int("teams.processed", teamsProcessed),
		attribute.Int("teams.with.new.sprints", teamsWithNewSprints),
		attribute.Int("sprints.created", sprintsCreated),
		attribute.Int("sprints.rescheduled", sprintsRescheduled),
		attribute.Int("errors", len(processingErrors)),
	))
	log.Info(ctx, "Sprint auto-creation completed",
		"teams_processed", teamsProcessed,
		"teams_with_new_sprints", teamsWithNewSprints,
		"sprints_created", sprintsCreated,
		"sprints_rescheduled", sprintsRescheduled,
		"errors", len(processingErrors))
	return joinSprintAutomationErrors(processingErrors)
}

// DisableAutomationForInactiveTeams disables sprint creation after the
// existing 90-day inactivity and 30-day settings grace periods.
func DisableAutomationForInactiveTeams(
	ctx context.Context,
	store SprintAutomationStore,
	log *logger.Logger,
) error {
	return disableAutomationForInactiveTeamsAt(ctx, store, log, time.Now().UTC())
}

func disableAutomationForInactiveTeamsAt(
	ctx context.Context,
	store SprintAutomationStore,
	log *logger.Logger,
	asOf time.Time,
) error {
	if ctx == nil {
		return errors.New("sprint automation inactivity context is required")
	}
	ctx, span := web.AddSpan(ctx, "jobs.DisableAutomationForInactiveTeams")
	defer span.End()
	if store == nil {
		return errors.New("sprint automation store is required")
	}
	if log == nil {
		return errors.New("sprint automation logger is required")
	}
	if asOf.IsZero() {
		return errors.New("sprint automation as-of time is required")
	}
	asOf = asOf.UTC()
	teamCreatedBefore := asOf.AddDate(0, 0, -sprintAutomationInactivityDays)
	settingsUpdatedBefore := asOf.AddDate(0, 0, -sprintAutomationSettingsGraceDays)
	activityBefore := teamCreatedBefore

	log.Info(ctx, "Processing disable automation for inactive teams")
	var cursor teamsettingsdomain.SprintAutomationCursor
	var processingErrors []error
	teamsScanned := 0
	teamsDisabled := 0

	for batch := 0; batch < sprintAutomationMaxBatches; batch++ {
		if err := ctx.Err(); err != nil {
			return errors.Join(joinSprintAutomationErrors(processingErrors), fmt.Errorf("disable inactive sprint automation: %w", err))
		}
		teams, err := store.ListSprintAutomationInactivityCandidates(
			ctx,
			teamsettingsdomain.SprintAutomationInactivityQuery{
				TeamCreatedBefore: teamCreatedBefore, SettingsUpdatedBefore: settingsUpdatedBefore,
				Cursor: cursor, BatchSize: sprintAutomationBatchSize,
			},
		)
		if err != nil {
			return errors.Join(joinSprintAutomationErrors(processingErrors), fmt.Errorf("list sprint automation inactivity candidates: %w", err))
		}
		if len(teams) > sprintAutomationBatchSize {
			return errors.Join(
				joinSprintAutomationErrors(processingErrors),
				fmt.Errorf("sprint automation inactivity page exceeds limit: got %d, limit %d", len(teams), sprintAutomationBatchSize),
			)
		}
		if len(teams) == 0 {
			return finishSprintAutomationInactivityRun(ctx, span, log, processingErrors, teamsScanned, teamsDisabled)
		}

		for _, team := range teams {
			disabled, err := store.DisableSprintAutomationIfInactive(
				ctx,
				teamsettingsdomain.SprintAutomationInactivityEligibility{
					WorkspaceID: team.WorkspaceID, TeamID: team.TeamID,
					TeamCreatedBefore: teamCreatedBefore, SettingsUpdatedBefore: settingsUpdatedBefore,
					ActivityBefore: activityBefore, DisabledAt: asOf,
					Reason:         sprintAutomationDisabledReason,
					InactivityDays: sprintAutomationInactivityDays,
					GraceDays:      sprintAutomationSettingsGraceDays,
				},
			)
			teamsScanned++
			if err != nil {
				log.Error(ctx, "Failed to disable sprint automation for inactive team",
					"error", err,
					"team_id", team.TeamID,
					"workspace_id", team.WorkspaceID)
				processingErrors = append(processingErrors, fmt.Errorf("team %s in workspace %s: %w", team.TeamID, team.WorkspaceID, err))
				continue
			}
			if disabled {
				teamsDisabled++
			}
		}

		last := teams[len(teams)-1]
		nextCursor := teamsettingsdomain.SprintAutomationCursor{
			WorkspaceID: last.WorkspaceID,
			TeamID:      last.TeamID,
			Valid:       true,
		}
		if cursor.Valid && !sprintAutomationCursorAdvances(cursor, nextCursor) {
			return errors.Join(joinSprintAutomationErrors(processingErrors), errors.New("sprint automation inactivity cursor did not advance"))
		}
		cursor = nextCursor
		span.AddEvent("sprint automation inactivity batch processed", trace.WithAttributes(
			attribute.Int("batch", batch+1),
			attribute.Int("teams.scanned", len(teams)),
		))
		if len(teams) < sprintAutomationBatchSize {
			return finishSprintAutomationInactivityRun(ctx, span, log, processingErrors, teamsScanned, teamsDisabled)
		}
	}

	return errors.Join(
		joinSprintAutomationErrors(processingErrors),
		fmt.Errorf("disable inactive sprint automation after %d teams: %w", teamsScanned, errSprintAutomationBacklogRemaining),
	)
}

func finishSprintAutomationInactivityRun(
	ctx context.Context,
	span trace.Span,
	log *logger.Logger,
	processingErrors []error,
	teamsScanned, teamsDisabled int,
) error {
	span.AddEvent("sprint automation inactivity completed", trace.WithAttributes(
		attribute.Int("teams.scanned", teamsScanned),
		attribute.Int("teams.disabled", teamsDisabled),
		attribute.Int("errors", len(processingErrors)),
	))
	log.Info(ctx, "Disabled automation for inactive teams",
		"teams_scanned", teamsScanned,
		"teams_disabled", teamsDisabled,
		"errors", len(processingErrors))
	return joinSprintAutomationErrors(processingErrors)
}

func sprintAutomationCursorAdvances(
	current, next teamsettingsdomain.SprintAutomationCursor,
) bool {
	workspaceComparison := bytes.Compare(next.WorkspaceID[:], current.WorkspaceID[:])
	return workspaceComparison > 0 ||
		(workspaceComparison == 0 && bytes.Compare(next.TeamID[:], current.TeamID[:]) > 0)
}

func waitForSprintAutomationRetry(ctx context.Context) error {
	timer := time.NewTimer(sprintAutomationRetryDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func joinSprintAutomationErrors(processingErrors []error) error {
	return errors.Join(processingErrors...)
}
