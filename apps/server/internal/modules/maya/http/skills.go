package mayahttp

import (
	"context"
	"net/http"
	"time"

	maya "github.com/complexus-tech/projects-api/internal/modules/maya/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type AppSkill struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Instructions string    `json:"instructions"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

type AppSkillRequest struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Instructions string    `json:"instructions"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

func (request AppSkillRequest) content() maya.SkillContent {
	return maya.SkillContent{Name: request.Name, Description: request.Description, Instructions: request.Instructions}
}

func (h *Handlers) skillScope(ctx context.Context) (maya.SkillScope, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return maya.SkillScope{}, err
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return maya.SkillScope{}, err
	}
	allowed, err := h.workspaceCanUseMaya(ctx, workspace.ID)
	if err != nil {
		return maya.SkillScope{}, err
	}
	if !allowed {
		return maya.SkillScope{}, ErrMayaAccessRequired
	}
	return maya.SkillScope{WorkspaceID: workspace.ID, UserID: userID}, nil
}

func (h *Handlers) ListSkills(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.skillScope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	skills, err := h.service.ListSkills(ctx, scope)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	response := make([]AppSkill, 0, len(skills))
	for _, skill := range skills {
		response = append(response, AppSkill(skill))
	}
	return web.Respond(ctx, w, response, http.StatusOK)
}

func (h *Handlers) CreateSkill(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.skillScope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	var request AppSkillRequest
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	skill, err := h.service.CreateSkill(ctx, scope, request.content())
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, AppSkill(skill), http.StatusCreated)
}

func (h *Handlers) UpdateSkill(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.skillScope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	id, err := uuid.Parse(web.Params(r, "skillId"))
	if err != nil {
		return web.RespondError(ctx, w, maya.ErrInvalidSkill, http.StatusBadRequest)
	}
	var request AppSkillRequest
	if err := web.Decode(r, &request); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	skill, err := h.service.UpdateSkill(ctx, scope, id, request.content(), request.UpdatedAt)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, AppSkill(skill), http.StatusOK)
}

func (h *Handlers) DeleteSkill(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := h.skillScope(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	id, err := uuid.Parse(web.Params(r, "skillId"))
	if err != nil {
		return web.RespondError(ctx, w, maya.ErrInvalidSkill, http.StatusBadRequest)
	}
	if err := h.service.DeleteSkill(ctx, scope, id); err != nil {
		return web.RespondError(ctx, w, err, h.statusCode(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
