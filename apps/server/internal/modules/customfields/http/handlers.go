package customfieldshttp

import (
	"context"
	"errors"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/customfields/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Service interface {
	List(context.Context, domain.Scope) ([]domain.Field, error)
	Create(context.Context, domain.Scope, domain.Definition) (domain.Field, error)
	Update(context.Context, domain.Scope, uuid.UUID, domain.Definition) (domain.Field, error)
	Archive(context.Context, domain.Scope, uuid.UUID) error
	Snapshot(context.Context, domain.Scope, uuid.UUID) (domain.Snapshot, error)
	Batch(context.Context, domain.Scope, []uuid.UUID) (domain.BatchSnapshot, error)
	Patch(context.Context, domain.Scope, uuid.UUID, domain.ValuePatch) (domain.Snapshot, error)
	Report(context.Context, domain.Scope, domain.ReportInput) (domain.Report, error)
}
type Handlers struct{ service Service }

func New(service Service) *Handlers { return &Handlers{service: service} }

func requestScope(ctx context.Context, r *http.Request, team bool) (domain.Scope, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return domain.Scope{}, domain.ErrForbidden
	}
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return domain.Scope{}, domain.ErrForbidden
	}
	scope := domain.Scope{ActorID: actorID, WorkspaceID: workspace.ID}
	if team {
		scope.TeamID, err = parseID(r, "teamId")
	}
	return scope, err
}
func parseID(r *http.Request, name string) (uuid.UUID, error) {
	id, err := uuid.Parse(web.Params(r, name))
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domain.ErrInvalid
	}
	return id, nil
}
func errorStatus(err error) int {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrInvalid), errors.Is(err, domain.ErrLimit):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
func respond(ctx context.Context, w http.ResponseWriter, result any, err error, status int) error {
	if err != nil {
		return web.RespondError(ctx, w, err, errorStatus(err))
	}
	w.Header().Set("Cache-Control", "private, no-store")
	return web.Respond(ctx, w, result, status)
}
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, true)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	result, err := h.service.List(ctx, scope)
	return respond(ctx, w, result, err, http.StatusOK)
}
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, true)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	input, err := decodeDefinition(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Create(ctx, scope, input)
	return respond(ctx, w, result, err, http.StatusCreated)
}
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, true)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	id, err := parseID(r, "fieldId")
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	input, err := decodeDefinition(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Update(ctx, scope, id, input)
	return respond(ctx, w, result, err, http.StatusOK)
}
func (h *Handlers) Archive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, true)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	id, err := parseID(r, "fieldId")
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	return respond(ctx, w, nil, h.service.Archive(ctx, scope, id), http.StatusOK)
}
func (h *Handlers) Snapshot(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, false)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	id, err := parseID(r, "id")
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	result, err := h.service.Snapshot(ctx, scope, id)
	return respond(ctx, w, result, err, http.StatusOK)
}

func (h *Handlers) Batch(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, false)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	var input domain.BatchInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Batch(ctx, scope, input.StoryIDs)
	return respond(ctx, w, result, err, http.StatusOK)
}
func (h *Handlers) Patch(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, false)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	id, err := parseID(r, "id")
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	var input domain.ValuePatch
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Patch(ctx, scope, id, input)
	return respond(ctx, w, result, err, http.StatusOK)
}
func (h *Handlers) Report(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	scope, err := requestScope(ctx, r, false)
	if err != nil {
		return respond(ctx, w, nil, err, 0)
	}
	var input domain.ReportInput
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.service.Report(ctx, scope, input)
	return respond(ctx, w, result, err, http.StatusOK)
}
