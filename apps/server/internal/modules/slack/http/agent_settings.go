package slackhttp

import (
	"context"
	"net/http"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type AppSlackAgentSettings struct {
	AssistantEnabled       bool   `json:"assistantEnabled"`
	WorkflowActionsEnabled bool   `json:"workflowActionsEnabled"`
	Guidance               string `json:"guidance"`
}

func (h *Handlers) GetAgentSettings(ctx context.Context, w http.ResponseWriter, _ *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	settings, err := h.service.GetAgentSettings(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, toAppSlackAgentSettings(settings), http.StatusOK)
}

func (h *Handlers) UpdateAgentSettings(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input AppSlackAgentSettings
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	settings, err := h.service.UpdateAgentSettings(ctx, workspace.ID, slack.CoreSlackAgentSettings{
		AssistantEnabled:       input.AssistantEnabled,
		WorkflowActionsEnabled: input.WorkflowActionsEnabled,
		Guidance:               input.Guidance,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, toAppSlackAgentSettings(settings), http.StatusOK)
}

func toAppSlackAgentSettings(settings slack.CoreSlackAgentSettings) AppSlackAgentSettings {
	return AppSlackAgentSettings{
		AssistantEnabled:       settings.AssistantEnabled,
		WorkflowActionsEnabled: settings.WorkflowActionsEnabled,
		Guidance:               settings.Guidance,
	}
}
