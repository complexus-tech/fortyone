package sprintshttp

import (
	"context"
	"net/http"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	sprints "github.com/complexus-tech/projects-api/internal/modules/sprints/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Handlers struct {
	sprints     *sprints.Service
	attachments *attachments.Service
}

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

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	query, err := parseSprintListQuery(r.URL.Query(), workspace.ID, userID)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	if query.Page != nil {
		params := *query.Page
		page, pageSize := params.Page, params.PageSize

		sprints, err := h.sprints.ListQuery(ctx, query.Query)
		if err != nil {
			return respondSprintError(ctx, w, err)
		}

		hasMore := len(sprints) > pageSize
		if hasMore {
			sprints = sprints[:pageSize]
		}

		web.Respond(ctx, w, toAppSprintsResponse(sprints, page, pageSize, hasMore), http.StatusOK)
		return nil
	}

	sprints, err := h.sprints.ListQuery(ctx, query.Query)
	if err != nil {
		return respondSprintError(ctx, w, err)
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}

func (h *Handlers) Running(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprints, err := h.sprints.Running(ctx, workspace.ID, userID)
	if err != nil {
		return respondSprintError(ctx, w, err)
	}
	web.Respond(ctx, w, toAppSprints(sprints), http.StatusOK)
	return nil
}

func (h *Handlers) GetByID(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintID, ok := sprintPathID(ctx, w, r)
	if !ok {
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprint, err := h.sprints.GetByID(ctx, sprintID, workspace.ID, userID)
	if err != nil {
		return respondSprintError(ctx, w, err)
	}

	web.Respond(ctx, w, toAppSprint(sprint), http.StatusOK)
	return nil
}

func (h *Handlers) GetAnalytics(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintID, ok := sprintPathID(ctx, w, r)
	if !ok {
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	analytics, err := h.sprints.GetAnalytics(ctx, sprintID, workspace.ID, userID)
	if err != nil {
		return respondSprintError(ctx, w, err)
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
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	sprint := sprints.CoreNewSprint{
		Name: app.Name, Goal: app.Goal, ObjectiveID: app.Objective,
		TeamID: app.Team, WorkspaceID: workspace.ID,
		StartDate: app.StartDate.Time(), EndDate: app.EndDate.Time(),
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	result, err := h.sprints.Create(ctx, sprint, &userID)
	if err != nil {
		return respondSprintError(ctx, w, err)
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusCreated)
	return nil
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintID, ok := sprintPathID(ctx, w, r)
	if !ok {
		return nil
	}

	var app AppUpdateSprint
	if err := web.Decode(r, &app); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	result, err := h.sprints.UpdatePatch(
		ctx, sprintID, workspace.ID, userID, app.SprintPatch(), app.ExpectedUpdatedAt,
	)
	if err != nil {
		return respondSprintError(ctx, w, err)
	}

	web.Respond(ctx, w, toAppSprints([]sprints.CoreSprint{result})[0], http.StatusOK)
	return nil
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	sprintID, ok := sprintPathID(ctx, w, r)
	if !ok {
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	if err := h.sprints.Delete(ctx, sprintID, workspace.ID, &userID); err != nil {
		return respondSprintError(ctx, w, err)
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}
