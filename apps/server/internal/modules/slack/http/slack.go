package slackhttp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	log     *logger.Logger
	service *slack.Service
}

func New(log *logger.Logger, service *slack.Service) *Handlers {
	return &Handlers{log: log, service: service}
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
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, AppCreateInstallSession{InstallURL: session.InstallURL}, http.StatusOK)
}

func (h *Handlers) DisconnectWorkspace(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.DisconnectWorkspace(ctx, workspace.ID); err != nil {
		if slack.IsNotFound(err) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) HandleSetup(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	slackError := r.URL.Query().Get("error")
	redirectURL, err := h.service.HandleSetup(ctx, code, state, slackError)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) ResyncChannels(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.SyncChannels(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}

func (h *Handlers) UpdateWorkspaceSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppUpdateWorkspaceSettingsRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.service.UpdateWorkspaceSettings(ctx, workspace.ID, slack.CoreUpdateWorkspaceSettingsInput{
		DefaultCreateMode: input.DefaultCreateMode,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, toAppWorkspaceSettings(settings), http.StatusOK)
}

func (h *Handlers) CreateChannelLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppCreateChannelLinkRequest
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	teamID, err := uuid.Parse(input.TeamID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	link, err := h.service.CreateChannelLink(ctx, workspace.ID, userID, slack.CoreCreateChannelLinkInput{
		SlackChannelID: input.SlackChannelID,
		TeamID:         teamID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, toAppChannelLink(link), http.StatusCreated)
}

func (h *Handlers) DeleteChannelLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.DeleteChannelLink(ctx, workspace.ID, linkID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) HandleEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleEvents(rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if response.Challenge != "" {
		return writeRawJSON(w, http.StatusOK, response)
	}
	w.WriteHeader(http.StatusOK)
	return nil
}

func (h *Handlers) HandleCommands(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleCommand(ctx, rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return writeRawJSON(w, http.StatusOK, response)
}

func (h *Handlers) HandleInteractivity(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	response, err := h.service.HandleInteractivity(ctx, rawBody)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if response.ContentType != "" {
		w.Header().Set("Content-Type", response.ContentType)
	}
	status := response.StatusCode
	if status == 0 {
		status = http.StatusOK
	}
	if len(response.Body) > 0 {
		w.WriteHeader(status)
		_, writeErr := w.Write(response.Body)
		return writeErr
	}
	w.WriteHeader(status)
	return nil
}

func writeRawJSON(w http.ResponseWriter, status int, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}
