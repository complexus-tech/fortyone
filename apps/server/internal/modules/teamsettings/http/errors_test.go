package teamsettingshttp

import (
	"net/http"
	"testing"

	teamsettings "github.com/complexus-tech/projects-api/internal/modules/teamsettings/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
)

func TestTeamSettingsErrorStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		err    error
		status int
	}{
		{name: "validation", err: teamsettings.ErrInvalidSprintDuration, status: http.StatusBadRequest},
		{name: "empty patch", err: teamsettings.ErrNoSettingsChanges, status: http.StatusBadRequest},
		{name: "schedule conflict", err: teamsettings.ErrSprintScheduleConflict, status: http.StatusConflict},
		{name: "concurrent update", err: teamsettings.ErrConcurrentUpdate, status: http.StatusConflict},
		{name: "not found", err: teamsettings.ErrTeamSettingsNotFound, status: http.StatusNotFound},
		{name: "team membership", err: teamsettings.ErrTeamMembershipRequired, status: http.StatusForbidden},
		{name: "admin policy", err: authorization.ErrWorkspaceAdminRequired, status: http.StatusForbidden},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := teamSettingsErrorStatus(test.err); got != test.status {
				t.Fatalf("teamSettingsErrorStatus() = %d, want %d", got, test.status)
			}
		})
	}
}
