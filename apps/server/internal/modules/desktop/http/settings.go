package desktophttp

import (
	"context"
	"net/http"
	"time"

	users "github.com/complexus-tech/projects-api/internal/modules/users/service"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// SettingsUser is an explicit download allowlist, deliberately independent of
// auth/session/integration fields added to the user domain model in future.
type SettingsUser struct {
	ID                 uuid.UUID `json:"id"`
	Username           string    `json:"username"`
	FullName           string    `json:"fullName"`
	Email              string    `json:"email"`
	Timezone           string    `json:"timezone"`
	Role               *string   `json:"role,omitempty"`
	IsActive           bool      `json:"isActive"`
	WorkingDays        []int     `json:"workingDays"`
	WorkingStartMinute *int      `json:"workingStartMinute"`
	WorkingEndMinute   *int      `json:"workingEndMinute"`
}

func settingsUser(user users.CoreUser) SettingsUser {
	return SettingsUser{ID: user.ID, Username: user.Username, FullName: user.FullName, Email: user.Email, Timezone: user.Timezone, Role: user.Role, IsActive: user.IsActive, WorkingDays: user.WorkingDays, WorkingStartMinute: user.WorkingStartMinute, WorkingEndMinute: user.WorkingEndMinute}
}

func (h *Handlers) Settings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "no-store")
	identity, userID, err := requestIdentity(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	profile, err := h.Users.GetUser(ctx, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	workspace, err := h.Workspaces.Get(ctx, identity.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	settings, err := h.Workspaces.GetOrCreateWorkspaceSettings(ctx, identity.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	preferences, err := h.Users.GetAutomationPreferences(ctx, userID, identity.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	members, err := h.Users.List(ctx, identity.ID, users.CoreListUsersFilter{})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	teams, err := h.Teams.List(ctx, identity.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	mappedMembers := make([]SettingsUser, len(members))
	for i, member := range members {
		mappedMembers[i] = settingsUser(member)
	}
	mappedTeams := make([]Team, len(teams))
	for i, team := range teams {
		mappedTeams[i] = Team{ID: team.ID, Name: team.Name, Code: team.Code, Color: team.Color}
	}
	return web.Respond(ctx, w, map[string]any{
		"profile": settingsUser(profile),
		"workspace": map[string]any{"id": workspace.ID, "name": workspace.Name, "slug": workspace.Slug, "color": workspace.Color, "userRole": identity.UserRole,
			"storyTerm": settings.StoryTerm, "sprintTerm": settings.SprintTerm, "objectiveTerm": settings.ObjectiveTerm, "keyResultTerm": settings.KeyResultTerm,
			"objectiveEnabled": settings.ObjectiveEnabled, "keyResultEnabled": settings.KeyResultEnabled, "workingDays": settings.WorkingDays, "workingStartMinute": settings.WorkingStartMinute, "workingEndMinute": settings.WorkingEndMinute},
		"teams": mappedTeams, "members": mappedMembers,
		"preferences":  map[string]bool{"autoAssignSelf": preferences.AutoAssignSelf, "autoScheduling": preferences.AutoScheduling, "assignSelfOnBranchCopy": preferences.AssignSelfOnBranchCopy, "moveStoryToStartedOnBranch": preferences.MoveStoryToStartedOnBranch, "openStoryInDialog": preferences.OpenStoryInDialog},
		"downloadedAt": time.Now().UTC(),
	}, http.StatusOK)
}
