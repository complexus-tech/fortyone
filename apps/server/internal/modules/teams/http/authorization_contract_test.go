package teamshttp

import (
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
)

func TestTeamMutationRoutesEnforceServerRoles(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("routes.go")
	if err != nil {
		t.Fatalf("read team routes: %v", err)
	}
	routes := string(source)

	required := []string{
		`app.Post("/workspaces/{workspaceSlug}/teams", h.Create, auth, workspace, memberAndAdmin)`,
		`app.Put("/workspaces/{workspaceSlug}/teams/{id}", h.Update, auth, workspace, adminOnly)`,
		`app.Delete("/workspaces/{workspaceSlug}/teams/{id}", h.Delete, auth, workspace, adminOnly)`,
		`app.Post("/workspaces/{workspaceSlug}/teams/{id}/members", h.AddMember, auth, workspace, gzip, adminOnly)`,
		`app.Put("/workspaces/{workspaceSlug}/teams/{id}/members/{userId}/ai-context", h.UpdateMemberAIContext, auth, workspace, adminOnly)`,
		`app.Delete("/workspaces/{workspaceSlug}/teams/{id}/members/{userId}", h.RemoveMember, auth, workspace, adminOnly)`,
		`app.Delete("/workspaces/{workspaceSlug}/teams/{id}/membership", h.LeaveTeam, auth, workspace)`,
	}
	for _, contract := range required {
		if !strings.Contains(routes, contract) {
			t.Fatalf("team route is missing authorization contract %q", contract)
		}
	}
}

func TestLeaveTeamHandlerBindsAuthenticatedActor(t *testing.T) {
	t.Parallel()

	source, err := os.ReadFile("teams.go")
	if err != nil {
		t.Fatalf("read team handlers: %v", err)
	}
	handler := sourceBetween(t, string(source), "func (h *Handlers) LeaveTeam", "func teamMembershipMutationStatus")

	for _, required := range []string{
		"mid.GetUserID(ctx)",
		"ActorID:     actorID",
		"WorkspaceID: workspace.ID",
		"h.teams.LeaveTeam(ctx, input)",
	} {
		if !strings.Contains(handler, required) {
			t.Fatalf("self-leave handler must contain %q", required)
		}
	}
	for _, forbidden := range []string{`web.Params(r, "userId")`, "h.teams.RemoveMember("} {
		if strings.Contains(handler, forbidden) {
			t.Fatalf("self-leave handler must not accept a caller-selected member: %q", forbidden)
		}
	}
}

func TestTeamMembershipMutationStatusHidesScopedMisses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		err  error
		want int
	}{
		{err: teams.ErrTeamNotFound, want: http.StatusNotFound},
		{err: teams.ErrTeamMemberNotFound, want: http.StatusNotFound},
		{err: errors.New("database unavailable"), want: http.StatusInternalServerError},
	}
	for _, tt := range tests {
		if got := teamMembershipMutationStatus(tt.err); got != tt.want {
			t.Fatalf("teamMembershipMutationStatus(%v) = %d, want %d", tt.err, got, tt.want)
		}
	}
}

func sourceBetween(t *testing.T, source, start, end string) string {
	t.Helper()
	startIndex := strings.Index(source, start)
	if startIndex < 0 {
		t.Fatalf("source marker %q not found", start)
	}
	endIndex := strings.Index(source[startIndex+len(start):], end)
	if endIndex < 0 {
		t.Fatalf("source marker %q not found after %q", end, start)
	}
	return source[startIndex : startIndex+len(start)+endIndex]
}
