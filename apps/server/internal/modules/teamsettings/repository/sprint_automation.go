package teamsettingsrepository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	teamsettingssql "github.com/complexus-tech/projects-api/internal/modules/teamsettings/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maxSprintAutomationBatchSize = 100

type automatedSprintAuditMetadata struct {
	Name                string `json:"name"`
	SprintNumber        int    `json:"sprint_number"`
	StartDate           string `json:"start_date"`
	EndDate             string `json:"end_date"`
	UpcomingTarget      int    `json:"upcoming_target"`
	SprintDurationWeeks int    `json:"sprint_duration_weeks"`
}

type disabledSprintAutomationAuditMetadata struct {
	TeamName       string     `json:"team_name"`
	Reason         string     `json:"reason"`
	InactivityDays int        `json:"inactivity_days"`
	GraceDays      int        `json:"grace_days"`
	LastActivityAt *time.Time `json:"last_activity_at"`
}

func (r *repo) ListSprintAutomationTeams(
	ctx context.Context,
	query teamsettings.SprintAutomationQuery,
) ([]teamsettings.SprintAutomationTeamRef, error) {
	batchSize, err := validateSprintAutomationPage(query.Cursor, query.BatchSize)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSprintAutomationTeams(ctx, teamsettingssql.ListSprintAutomationTeamsParams{
		HasCursor: query.Cursor.Valid, AfterWorkspaceID: query.Cursor.WorkspaceID,
		AfterTeamID: query.Cursor.TeamID, BatchSize: batchSize,
	})
	if err != nil {
		return nil, fmt.Errorf("list sprint automation teams: %w", mapDatabaseError(err))
	}
	refs := make([]teamsettings.SprintAutomationTeamRef, len(rows))
	for index, row := range rows {
		refs[index] = teamsettings.SprintAutomationTeamRef{
			WorkspaceID: row.WorkspaceID,
			TeamID:      row.TeamID,
		}
	}
	return refs, nil
}

func (r *repo) RunSprintAutomationForTeam(
	ctx context.Context,
	ref teamsettings.SprintAutomationTeamRef,
	asOf time.Time,
) (teamsettings.SprintAutomationRunResult, error) {
	if ref.WorkspaceID == uuid.Nil || ref.TeamID == uuid.Nil || asOf.IsZero() {
		return teamsettings.SprintAutomationRunResult{}, teamsettings.ErrInvalidSprintAutomationQuery
	}
	asOf = asOf.UTC()
	scheduleDate := utcDate(asOf)

	var pending teamsettings.SprintAutomationRunResult
	err := r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		if err := queries.LockSprintAutomation(ctx, teamsettingssql.LockSprintAutomationParams{
			WorkspaceID: ref.WorkspaceID,
			TeamID:      ref.TeamID,
		}); err != nil {
			return fmt.Errorf("lock sprint automation: %w", err)
		}
		locked, err := queries.LockSprintSettings(ctx, teamsettingssql.LockSprintSettingsParams{
			TeamID: ref.TeamID, WorkspaceID: ref.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("lock sprint automation settings: %w", err)
		}
		settings := mapLockedSprintSettings(locked)
		if !settings.AutoCreateSprints {
			return nil
		}

		rescheduled, err := reconcileSprintScheduleAt(
			ctx,
			queries,
			settings,
			scheduleDate,
			asOf,
			func(
				ctx context.Context,
				queries teamsettingssql.Querier,
				settings teamsettings.CoreTeamSprintSettings,
				sprint scheduledSprint,
				startDate, endDate time.Time,
			) error {
				return insertSprintAutomationAuditEvent(
					ctx, queries, settings.WorkspaceID, settings.TeamID,
					"sprint", sprint.ID, "sprint.auto_rescheduled",
					scheduleAuditMetadata(settings, sprint, startDate, endDate), asOf,
				)
			},
		)
		if err != nil {
			return err
		}
		pending.Rescheduled = rescheduled

		target, err := safecast.Int32(settings.UpcomingSprintsCount)
		if err != nil {
			return teamsettings.ErrInvalidUpcomingCount
		}
		existing, err := queries.CountUpcomingSprintsForAutomation(ctx, teamsettingssql.CountUpcomingSprintsForAutomationParams{
			TeamID: ref.TeamID, WorkspaceID: ref.WorkspaceID,
			ScheduleDate: scheduleDate, UpcomingTarget: target,
		})
		if err != nil {
			return fmt.Errorf("count upcoming sprints for automation: %w", err)
		}
		missing := settings.UpcomingSprintsCount - int(existing)
		if missing <= 0 {
			return nil
		}

		anchor := scheduleDate
		boundary, err := queries.GetSprintAutomationScheduleBoundary(ctx, teamsettingssql.GetSprintAutomationScheduleBoundaryParams{
			TeamID: ref.TeamID, WorkspaceID: ref.WorkspaceID, ScheduleDate: scheduleDate,
		})
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("get sprint automation schedule boundary: %w", err)
		}
		if err == nil {
			anchor = boundary
		}
		startDate := nextSprintStartAfter(anchor, settings.SprintStartDay)

		for index := 0; index < missing; index++ {
			sprintNumber := settings.NextAutoSprintNumber + index
			sprintStart := startDate.AddDate(0, 0, index*settings.SprintDurationWeeks*7)
			sprintEnd := sprintStart.AddDate(0, 0, settings.SprintDurationWeeks*7-1)
			sprintName := fmt.Sprintf("Sprint %d", sprintNumber)
			sprintID, err := queries.CreateAutomatedSprint(ctx, teamsettingssql.CreateAutomatedSprintParams{
				Name: sprintName, TeamID: ref.TeamID, WorkspaceID: ref.WorkspaceID,
				StartDate: sprintStart, EndDate: sprintEnd, CreatedAt: asOf, UpdatedAt: asOf,
			})
			if err != nil {
				return fmt.Errorf("create automated sprint %d: %w", sprintNumber, err)
			}
			if err := insertSprintAutomationAuditEvent(
				ctx, queries, ref.WorkspaceID, ref.TeamID,
				"sprint", sprintID, "sprint.auto_created",
				automatedSprintAuditMetadata{
					Name: sprintName, SprintNumber: sprintNumber,
					StartDate: sprintStart.Format(time.DateOnly), EndDate: sprintEnd.Format(time.DateOnly),
					UpcomingTarget: settings.UpcomingSprintsCount, SprintDurationWeeks: settings.SprintDurationWeeks,
				},
				asOf,
			); err != nil {
				return err
			}
		}

		createdCount, err := safecast.Int32(missing)
		if err != nil {
			return teamsettings.ErrInvalidUpcomingCount
		}
		expectedNext, err := safecast.Int32(settings.NextAutoSprintNumber)
		if err != nil {
			return teamsettings.ErrInvalidNextAutoNumber
		}
		rowsAffected, err := queries.AdvanceSprintAutomationCounter(ctx, teamsettingssql.AdvanceSprintAutomationCounterParams{
			CreatedCount: createdCount, UpdatedAt: asOf, TeamID: ref.TeamID,
			WorkspaceID: ref.WorkspaceID, ExpectedNextNumber: expectedNext,
		})
		if err != nil {
			return fmt.Errorf("advance sprint automation counter: %w", err)
		}
		if rowsAffected != 1 {
			return teamsettings.ErrConcurrentUpdate
		}
		pending.Created = missing
		return nil
	})
	if err != nil {
		return teamsettings.SprintAutomationRunResult{}, err
	}
	return pending, nil
}

func (r *repo) ListSprintAutomationInactivityCandidates(
	ctx context.Context,
	query teamsettings.SprintAutomationInactivityQuery,
) ([]teamsettings.SprintAutomationTeamRef, error) {
	if query.TeamCreatedBefore.IsZero() || query.SettingsUpdatedBefore.IsZero() {
		return nil, teamsettings.ErrInvalidSprintAutomationQuery
	}
	batchSize, err := validateSprintAutomationPage(query.Cursor, query.BatchSize)
	if err != nil {
		return nil, err
	}
	rows, err := r.queries.ListSprintAutomationInactivityCandidates(
		ctx,
		teamsettingssql.ListSprintAutomationInactivityCandidatesParams{
			TeamCreatedBefore: query.TeamCreatedBefore.UTC(), SettingsUpdatedBefore: query.SettingsUpdatedBefore.UTC(),
			HasCursor: query.Cursor.Valid, AfterWorkspaceID: query.Cursor.WorkspaceID,
			AfterTeamID: query.Cursor.TeamID, BatchSize: batchSize,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("list sprint automation inactivity candidates: %w", mapDatabaseError(err))
	}
	refs := make([]teamsettings.SprintAutomationTeamRef, len(rows))
	for index, row := range rows {
		refs[index] = teamsettings.SprintAutomationTeamRef{WorkspaceID: row.WorkspaceID, TeamID: row.TeamID}
	}
	return refs, nil
}

func (r *repo) DisableSprintAutomationIfInactive(
	ctx context.Context,
	eligibility teamsettings.SprintAutomationInactivityEligibility,
) (bool, error) {
	if err := validateSprintAutomationInactivityEligibility(eligibility); err != nil {
		return false, err
	}
	eligibility.TeamCreatedBefore = eligibility.TeamCreatedBefore.UTC()
	eligibility.SettingsUpdatedBefore = eligibility.SettingsUpdatedBefore.UTC()
	eligibility.ActivityBefore = eligibility.ActivityBefore.UTC()
	eligibility.DisabledAt = eligibility.DisabledAt.UTC()
	eligibility.Reason = strings.TrimSpace(eligibility.Reason)

	disabled := false
	err := r.withinTransaction(ctx, func(queries teamsettingssql.Querier) error {
		if err := queries.LockSprintAutomation(ctx, teamsettingssql.LockSprintAutomationParams{
			WorkspaceID: eligibility.WorkspaceID,
			TeamID:      eligibility.TeamID,
		}); err != nil {
			return fmt.Errorf("lock sprint automation: %w", err)
		}
		locked, err := queries.LockSprintSettings(ctx, teamsettingssql.LockSprintSettingsParams{
			TeamID: eligibility.TeamID, WorkspaceID: eligibility.WorkspaceID,
		})
		if err != nil {
			return fmt.Errorf("lock sprint automation settings: %w", err)
		}
		settings := mapLockedSprintSettings(locked)
		if !settings.AutoCreateSprints {
			return nil
		}
		snapshot, err := queries.GetSprintAutomationInactivitySnapshot(
			ctx,
			teamsettingssql.GetSprintAutomationInactivitySnapshotParams{
				TeamID: eligibility.TeamID, WorkspaceID: eligibility.WorkspaceID,
			},
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("get sprint automation inactivity snapshot: %w", err)
		}
		lastActivityAt := latestSprintPlanningActivity(snapshot)
		if snapshot.TeamCreatedAt.After(eligibility.TeamCreatedBefore) ||
			snapshot.SettingsUpdatedAt.After(eligibility.SettingsUpdatedBefore) ||
			(lastActivityAt != nil && !lastActivityAt.Before(eligibility.ActivityBefore)) {
			return nil
		}

		reason := eligibility.Reason
		rowsAffected, err := queries.DisableSprintAutomationIfInactive(
			ctx,
			teamsettingssql.DisableSprintAutomationIfInactiveParams{
				DisabledAt: eligibility.DisabledAt, DisabledReason: &reason,
				TeamID: eligibility.TeamID, WorkspaceID: eligibility.WorkspaceID,
				TeamCreatedBefore:     eligibility.TeamCreatedBefore,
				SettingsUpdatedBefore: eligibility.SettingsUpdatedBefore,
				ActivityBefore:        eligibility.ActivityBefore,
			},
		)
		if err != nil {
			return fmt.Errorf("disable inactive sprint automation: %w", err)
		}
		if rowsAffected == 0 {
			return nil
		}
		if rowsAffected != 1 {
			return teamsettings.ErrConcurrentUpdate
		}
		if err := insertSprintAutomationAuditEvent(
			ctx, queries, eligibility.WorkspaceID, eligibility.TeamID,
			"automation_setting", eligibility.TeamID, "sprint_automation.disabled",
			disabledSprintAutomationAuditMetadata{
				TeamName: snapshot.Name, Reason: eligibility.Reason,
				InactivityDays: eligibility.InactivityDays, GraceDays: eligibility.GraceDays,
				LastActivityAt: lastActivityAt,
			},
			eligibility.DisabledAt,
		); err != nil {
			return err
		}
		disabled = true
		return nil
	})
	if err != nil {
		return false, err
	}
	return disabled, nil
}

func validateSprintAutomationPage(
	cursor teamsettings.SprintAutomationCursor,
	batchSize int,
) (int32, error) {
	if batchSize <= 0 || batchSize > maxSprintAutomationBatchSize {
		return 0, teamsettings.ErrInvalidSprintAutomationQuery
	}
	if cursor.Valid && (cursor.WorkspaceID == uuid.Nil || cursor.TeamID == uuid.Nil) {
		return 0, teamsettings.ErrInvalidSprintAutomationQuery
	}
	converted, err := safecast.Int32(batchSize)
	if err != nil {
		return 0, teamsettings.ErrInvalidSprintAutomationQuery
	}
	return converted, nil
}

func validateSprintAutomationInactivityEligibility(
	eligibility teamsettings.SprintAutomationInactivityEligibility,
) error {
	if eligibility.WorkspaceID == uuid.Nil || eligibility.TeamID == uuid.Nil ||
		eligibility.TeamCreatedBefore.IsZero() || eligibility.SettingsUpdatedBefore.IsZero() ||
		eligibility.ActivityBefore.IsZero() || eligibility.DisabledAt.IsZero() ||
		strings.TrimSpace(eligibility.Reason) == "" ||
		eligibility.InactivityDays <= 0 || eligibility.GraceDays <= 0 {
		return teamsettings.ErrInvalidSprintAutomationQuery
	}
	return nil
}

func latestSprintPlanningActivity(
	row teamsettingssql.GetSprintAutomationInactivitySnapshotRow,
) *time.Time {
	var latest time.Time
	if row.HasLatestHumanStory {
		latest = row.LatestHumanStoryAt.UTC()
	}
	if row.HasLatestHumanSprintChange && (latest.IsZero() || row.LatestHumanSprintChangeAt.After(latest)) {
		latest = row.LatestHumanSprintChangeAt.UTC()
	}
	if latest.IsZero() {
		return nil
	}
	return &latest
}

func insertSprintAutomationAuditEvent(
	ctx context.Context,
	queries teamsettingssql.Querier,
	workspaceID, teamID uuid.UUID,
	entityType string,
	entityID uuid.UUID,
	eventType string,
	metadata any,
	createdAt time.Time,
) error {
	payload, err := json.Marshal(metadata)
	if err != nil {
		return fmt.Errorf("marshal sprint automation audit metadata: %w", err)
	}
	if err := queries.InsertSprintAutomationAuditEvent(
		ctx,
		teamsettingssql.InsertSprintAutomationAuditEventParams{
			WorkspaceID: workspaceID, TeamID: &teamID,
			EntityType: entityType, EntityID: &entityID,
			EventType: eventType, Metadata: payload, CreatedAt: createdAt.UTC(),
		},
	); err != nil {
		return fmt.Errorf("insert sprint automation audit event: %w", mapDatabaseError(err))
	}
	return nil
}

func utcDate(value time.Time) time.Time {
	value = value.UTC()
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
