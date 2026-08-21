package figmahttp

import (
	"context"
	"io"
	"net/http"

	figma "github.com/complexus-tech/projects-api/internal/modules/figma/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct{ service *figma.Service }

func New(service *figma.Service) *Handlers { return &Handlers{service: service} }

type urlInput struct {
	URL string `json:"url" validate:"required,url"`
}
type createStoryInput struct {
	URL         string     `json:"url"`
	TeamID      uuid.UUID  `json:"teamId"`
	StatusID    *uuid.UUID `json:"statusId"`
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
}

func (h *Handlers) GetIntegration(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	result, err := h.service.GetIntegration(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, result, http.StatusOK)
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
	authURL, err := h.service.CreateInstallSession(ctx, workspace.ID, userID, workspace.Slug)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, map[string]string{"authorizationUrl": authURL}, http.StatusOK)
}
func (h *Handlers) CompleteOAuth(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	redirect, err := h.service.CompleteOAuth(ctx, r.URL.Query().Get("code"), r.URL.Query().Get("state"))
	if err != nil {
		if redirect != "" {
			http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
			return nil
		}
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	http.Redirect(w, r, redirect, http.StatusTemporaryRedirect)
	return nil
}
func (h *Handlers) Disconnect(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	if err := h.service.Disconnect(ctx, workspace.ID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
func (h *Handlers) ResolveLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input urlInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	artifact, err := h.service.ResolveLink(ctx, workspace.ID, input.URL)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, artifact, http.StatusOK)
}
func (h *Handlers) ListStoryLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	links, err := h.service.ListStoryLinks(ctx, workspace.ID, storyID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, links, http.StatusOK)
}

func (h *Handlers) ListStoryHandoffStatuses(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	statuses, err := h.service.ListStoryHandoffStatuses(ctx, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, statuses, http.StatusOK)
}
func (h *Handlers) LinkStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	storyID, err := uuid.Parse(web.Params(r, "storyId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	var input urlInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	link, err := h.service.LinkStory(ctx, workspace.ID, userID, storyID, workspace.Slug, input.URL)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, link, http.StatusCreated)
}
func (h *Handlers) DeleteStoryLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.DeleteStoryLink(ctx, workspace.ID, linkID); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
func (h *Handlers) RefreshStoryLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	linkID, err := uuid.Parse(web.Params(r, "linkId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	link, err := h.service.RefreshStoryLink(ctx, workspace.ID, linkID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, link, http.StatusOK)
}
func (h *Handlers) CreateStoryFromLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input createStoryInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	story, link, err := h.service.CreateStoryFromLink(ctx, workspace.ID, userID, figma.CreateStoryInput{URL: input.URL, WorkspaceSlug: workspace.Slug, TeamID: input.TeamID, StatusID: input.StatusID, Title: input.Title, Description: input.Description})
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, map[string]any{
		"story":     map[string]any{"id": story.ID, "sequenceId": story.SequenceID, "teamCode": story.TeamCode, "title": story.Title},
		"figmaLink": link,
	}, http.StatusCreated)
}
func (h *Handlers) HandleWebhook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.HandleWebhook(ctx, body); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	return web.Respond(ctx, w, nil, http.StatusOK)
}
