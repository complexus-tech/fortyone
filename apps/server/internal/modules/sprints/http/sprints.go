package sprintshttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Handlers struct {
	sprints     *sprints.Service
	attachments *attachments.Service
}

var (
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

func New(sprints *sprints.Service, attachments *attachments.Service) *Handlers {
	return &Handlers{
		sprints:     sprints,
		attachments: attachments,
	}
}

func (h *Handlers) resolveUserAvatarURL(ctx context.Context, avatar string) string {
	if h.attachments == nil {
		return avatar
	}
	resolved, err := h.attachments.ResolveProfileImageURL(ctx, avatar, 24*time.Hour)
	if err != nil {
		return ""
	}
	return resolved
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	userID, _ := mid.GetUserID(ctx)

	sprints, err := h.sprints.List(ctx, workspace.ID, userID, filters)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}

func (h *Handlers) Running(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, _ := mid.GetUserID(ctx)

	sprints, err := h.sprints.Running(ctx, workspace.ID, userID)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}

func (h *Handlers) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	sprint, err := h.sprints.GetByID(ctx, sprintId, workspace.ID)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprint(sprint), http.StatusOK)
	return nil
}

func (h *Handlers) GetAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	analytics, err := h.sprints.GetAnalytics(ctx, sprintId, workspace.ID)
	if err != nil {
		return err
	}

	for i := range analytics.TeamAllocation {
		analytics.TeamAllocation[i].AvatarURL = h.resolveUserAvatarURL(ctx, analytics.TeamAllocation[i].AvatarURL)
	}

	web.Respond(ctx, w, toAppSprintAnalytics(analytics), http.StatusOK)
	return nil
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var app AppNewSprint
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	sprint := sprints.CoreNewSprint{
		Name:      app.Name,
		Goal:      app.Goal,
		Objective: app.Objective,
		Team:      app.Team,
		Workspace: workspace.ID,
		StartDate: app.StartDate.Time(),
		EndDate:   app.EndDate.Time(),
	}

	result, err := h.sprints.Create(ctx, sprint)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusCreated)
	return nil
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	var app AppUpdateSprint
	if err := web.Decode(r, &app); err != nil {
		return err
	}

	userID, _ := mid.GetUserID(ctx)

	existingSprints, err := h.sprints.List(ctx, workspace.ID, userID, map[string]any{"sprint_id": sprintId})
	if err != nil {
		return err
	}
	if len(existingSprints) == 0 {
		return errors.New("sprint not found in workspace")
	}

	sprint := sprints.CoreUpdateSprint{
		Name:      app.Name,
		Goal:      app.Goal,
		Objective: app.Objective,
		StartDate: app.StartDate.TimePtr(),
		EndDate:   app.EndDate.TimePtr(),
	}

	result, err := h.sprints.Update(ctx, sprintId, workspace.ID, sprint)
	if err != nil {
		return err
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusOK)
	return nil
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintIdParam := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(sprintIdParam)
	if err != nil {
		return errors.New("sprint id is not in its proper form")
	}

	if err := h.sprints.Delete(ctx, sprintId, workspace.ID); err != nil {
		return err
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
