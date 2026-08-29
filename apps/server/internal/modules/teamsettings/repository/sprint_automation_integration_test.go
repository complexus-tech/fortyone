//go:build integration

package teamsettingsrepository

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/complexus-tech/projects-api/internal/testkit"
	"github.com/stretchr/testify/require"
)

func TestConcurrentSprintAutomationRunIsIdempotent(t *testing.T) {
	t.Parallel()

	postgres := testkit.NewPostgres(t)
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	fixture := newTeamSettingsFixture(t, ctx, postgres)
	asOf := time.Now().UTC().Truncate(time.Second)
	_, err := postgres.Pool.Exec(ctx, `
		INSERT INTO team_sprint_settings (
			team_id, workspace_id, auto_create_sprints, upcoming_sprints_count,
			sprint_duration_weeks, sprint_start_day, next_auto_sprint_number,
			last_auto_sprint_number, updated_at
		) VALUES ($1, $2, TRUE, 2, 2, 'Monday', 1, 0, $3)
		ON CONFLICT (team_id, workspace_id) DO UPDATE SET
			auto_create_sprints = TRUE,
			upcoming_sprints_count = 2,
			sprint_duration_weeks = 2,
			sprint_start_day = 'Monday',
			next_auto_sprint_number = 1,
			last_auto_sprint_number = 0,
			updated_at = EXCLUDED.updated_at
	`, fixture.teamA, fixture.workspaceA, asOf)
	require.NoError(t, err)

	repository := New(postgres.Pool)
	ref := teamsettings.SprintAutomationTeamRef{WorkspaceID: fixture.workspaceA, TeamID: fixture.teamA}
	start := make(chan struct{})
	errorsByRun := make(chan error, 2)
	var workers sync.WaitGroup
	for range 2 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			_, runErr := repository.RunSprintAutomationForTeam(ctx, ref, asOf)
			errorsByRun <- runErr
		}()
	}
	close(start)
	workers.Wait()
	close(errorsByRun)
	for runErr := range errorsByRun {
		if runErr != nil && !errors.Is(runErr, teamsettings.ErrConcurrentUpdate) {
			t.Fatalf("concurrent sprint automation run: %v", runErr)
		}
	}

	// A worker retry after either concurrent outcome must converge without
	// creating another schedule or consuming another sprint number.
	result, err := repository.RunSprintAutomationForTeam(ctx, ref, asOf)
	require.NoError(t, err)
	require.Zero(t, result.Created)

	var sprintCount, lastNumber, nextNumber, auditCount int
	err = postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM sprints
		WHERE team_id = $1
		  AND workspace_id = $2
		  AND schedule_managed_by_automation = TRUE
		  AND start_date > CAST($3 AS date)
	`, fixture.teamA, fixture.workspaceA, asOf).Scan(&sprintCount)
	require.NoError(t, err)
	err = postgres.Pool.QueryRow(ctx, `
		SELECT last_auto_sprint_number, next_auto_sprint_number
		FROM team_sprint_settings
		WHERE team_id = $1 AND workspace_id = $2
	`, fixture.teamA, fixture.workspaceA).Scan(&lastNumber, &nextNumber)
	require.NoError(t, err)
	err = postgres.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM audit_events
		WHERE team_id = $1
		  AND workspace_id = $2
		  AND event_type = 'sprint.auto_created'
	`, fixture.teamA, fixture.workspaceA).Scan(&auditCount)
	require.NoError(t, err)

	require.Equal(t, 2, sprintCount)
	require.Equal(t, 2, lastNumber)
	require.Equal(t, 3, nextNumber)
	require.Equal(t, 2, auditCount)
}
