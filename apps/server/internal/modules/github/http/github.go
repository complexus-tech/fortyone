package githubhttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	github "github.com/complexus-tech/projects-api/internal/modules/github/service"
	"github.com/complexus-tech/projects-api/internal/platform/authorization"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

// The durable gateway deliberately accepts a smaller, reviewable payload than
// the provider maximum. Oversized deliveries fail before signature parsing or
// persistence and can be investigated from safe receipt metadata.
const maxGitHubWebhookBodyBytes = 1 << 20

type Handlers struct {
	service *github.Service
	users   UserLookup
}

func New(service *github.Service, users UserLookup) *Handlers {
	return &Handlers{service: service, users: users}
}

func (h *Handlers) GetIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	integration, err := h.service.GetIntegration(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppIntegration(integration), http.StatusOK)
}

func (h *Handlers) CreateInstallSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	session, err := h.service.CreateInstallSession(ctx, workspace.ID, userID, workspace.Slug)
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	return web.Respond(ctx, w, AppCreateInstallSession{InstallURL: session.InstallURL}, http.StatusOK)
}

func (h *Handlers) HandleSetup(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	installationID, state, err := githubSetupCallbackParameters(r)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid GitHub installation callback"), http.StatusBadRequest)
	}
	redirectURL, err := h.service.HandleSetup(ctx, installationID, state)
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	// #nosec G710 -- HandleSetup atomically consumes opaque, identity-bound
	// installation state and constructs the destination from the configured
	// FortyOne website origin.
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func githubSetupCallbackParameters(r *http.Request) (int64, string, error) {
	installationValue, err := web.RequiredOpaqueQueryParameter(r.URL.Query(), "installation_id", 20)
	if err != nil {
		return 0, "", err
	}
	state, err := web.RequiredOpaqueQueryParameter(
		r.URL.Query(),
		"state",
		web.DefaultMaxQueryParameterBytes,
	)
	if err != nil {
		return 0, "", err
	}
	installationID, err := strconv.ParseInt(installationValue, 10, 64)
	if err != nil || installationID <= 0 {
		return 0, "", web.ErrInvalidQueryParameter
	}
	return installationID, state, nil
}

func (h *Handlers) ResyncRepositories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.ResyncRepositories(ctx, workspace.ID, userID); err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusInternalServerError))
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) CreateIssueSyncLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateIssueSyncLinkRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	link, err := h.service.CreateIssueSyncLink(ctx, workspace.ID, userID, toCoreIssueSyncLinkInput(input))
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	return web.Respond(ctx, w, toAppIssueSyncLink(link), http.StatusCreated)
}

func (h *Handlers) UpdateIssueSyncLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppUpdateIssueSyncLinkRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	link, err := h.service.UpdateIssueSyncLink(ctx, workspace.ID, userID, linkID, toCoreIssueSyncLinkUpdate(input))
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	return web.Respond(ctx, w, toAppIssueSyncLink(link), http.StatusOK)
}

func (h *Handlers) DeleteIssueSyncLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.DeleteIssueSyncLink(ctx, workspace.ID, userID, linkID); err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusInternalServerError))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) GetWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	integration, err := h.service.GetIntegration(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppWorkspaceSettings(integration.Settings), http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppUpdateWorkspaceSettingsRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.service.UpdateWorkspaceSettings(ctx, workspace.ID, userID, toCoreWorkspaceSettingsUpdate(input))
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	return web.Respond(ctx, w, toAppWorkspaceSettings(settings), http.StatusOK)
}

func (h *Handlers) GetTeamSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.service.GetTeamSettings(ctx, workspace.ID, teamID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppTeamSettings(settings), http.StatusOK)
}

func (h *Handlers) UpdateTeamSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	teamID, err := uuid.Parse(web.Params(r, "teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input AppUpdateTeamGitHubSettingsRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.service.UpdateTeamSettings(ctx, workspace.ID, userID, teamID, toCoreTeamSettingsUpdate(input))
	if err != nil {
		return web.RespondError(ctx, w, err, githubAdminErrorStatus(err, http.StatusBadRequest))
	}
	return web.Respond(ctx, w, toAppTeamSettings(settings), http.StatusOK)
}

func githubAdminErrorStatus(err error, fallback int) int {
	if errors.Is(err, authorization.ErrWorkspaceAdminRequired) {
		return http.StatusForbidden
	}
	return fallback
}
