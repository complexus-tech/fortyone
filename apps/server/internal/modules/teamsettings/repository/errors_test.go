package teamsettingsrepository

import (
	"context"
	"errors"
	"fmt"
	"testing"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestRepositoryRejectsEmptyPatchesBeforeOpeningTransaction(t *testing.T) {
	t.Parallel()

	repository := &repo{}
	actor := teamsettings.SystemAuditActor()
	teamID := uuid.New()
	workspaceID := uuid.New()
	if _, err := repository.UpdateSprintSettings(context.Background(), teamID, workspaceID, teamsettings.CoreUpdateTeamSprintSettings{}, actor); !errors.Is(err, teamsettings.ErrNoSettingsChanges) {
		t.Fatalf("empty sprint patch error = %v", err)
	}
	if _, err := repository.UpdateStoryAutomationSettings(context.Background(), teamID, workspaceID, teamsettings.CoreUpdateTeamStoryAutomationSettings{}, actor); !errors.Is(err, teamsettings.ErrNoSettingsChanges) {
		t.Fatalf("empty story patch error = %v", err)
	}
	if _, err := repository.UpdateEstimationSettings(context.Background(), teamID, workspaceID, teamsettings.CoreUpdateTeamEstimationSettings{}, actor); !errors.Is(err, teamsettings.ErrNoSettingsChanges) {
		t.Fatalf("empty estimation patch error = %v", err)
	}
}

func TestMapDatabaseError(t *testing.T) {
	t.Parallel()

	unchanged := errors.New("connection unavailable")
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{name: "nil", err: nil, wantErr: nil},
		{name: "no rows", err: fmt.Errorf("read settings: %w", pgx.ErrNoRows), wantErr: teamsettings.ErrTeamSettingsNotFound},
		{name: "serialization", err: &pgconn.PgError{Code: "40001"}, wantErr: teamsettings.ErrConcurrentUpdate},
		{name: "deadlock", err: &pgconn.PgError{Code: "40P01"}, wantErr: teamsettings.ErrConcurrentUpdate},
		{name: "sprint duration constraint", err: &pgconn.PgError{Code: "23514", ConstraintName: "team_sprint_settings_sprint_duration_weeks_check"}, wantErr: teamsettings.ErrInvalidSprintDuration},
		{name: "estimation constraint", err: &pgconn.PgError{Code: "23514", ConstraintName: "team_estimation_settings_scheme_check"}, wantErr: teamsettings.ErrInvalidEstimateScheme},
		{name: "unknown", err: unchanged, wantErr: unchanged},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if err := mapDatabaseError(test.err); !errors.Is(err, test.wantErr) {
				t.Fatalf("mapDatabaseError() = %v, want %v", err, test.wantErr)
			}
		})
	}
}
