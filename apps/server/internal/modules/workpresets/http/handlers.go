package workpresetshttp

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	domain "github.com/complexus-tech/projects-api/internal/modules/workpresets/domain"
	workpresets "github.com/complexus-tech/projects-api/internal/modules/workpresets/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/pagination"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type cursor struct {
	WorkspaceID uuid.UUID       `json:"workspaceId"`
	ActorID     uuid.UUID       `json:"actorId"`
	TeamID      uuid.UUID       `json:"teamId"`
	Kind        domain.Kind     `json:"kind"`
	Position    domain.Position `json:"position"`
	ExpiresAt   time.Time       `json:"expiresAt"`
}

type Handlers struct {
	service *workpresets.Service
	cursors pagination.CursorCodec[cursor]
}

func New(service *workpresets.Service, secret string) (*Handlers, error) {
	key, err := pagination.DeriveSigningKey("active", []byte(secret), "work-presets.list")
	if err != nil {
		return nil, err
	}
	codec, err := pagination.NewCursorCodec[cursor](key)
	if err != nil {
		return nil, err
	}
	return &Handlers{service: service, cursors: codec}, nil
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, workspaceID, err := identities(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	query := r.URL.Query()
	for key, values := range query {
		if (key != "teamId" && key != "kind" && key != "limit" && key != "cursor") || len(values) != 1 {
			return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
		}
	}
	teamID, err := uuid.Parse(query.Get("teamId"))
	if err != nil {
		return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
	}
	limit := 100
	if query.Has("limit") {
		limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil || limit < 1 || limit > 100 {
			return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
		}
	}
	filters := domain.List{TeamID: teamID, Kind: domain.Kind(query.Get("kind")), Limit: limit + 1}
	if query.Get("cursor") != "" {
		position, err := h.cursors.Decode(query.Get("cursor"))
		if err != nil || position.WorkspaceID != workspaceID || position.ActorID != actorID || position.TeamID != teamID || position.Kind != filters.Kind || position.ExpiresAt.Before(time.Now()) {
			return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
		}
		filters.Before = &position.Position
	}
	items, err := h.service.List(ctx, actorID, workspaceID, filters)
	if err != nil {
		return respondError(ctx, w, err)
	}
	next := ""
	if len(items) > limit {
		items = items[:limit]
		last := items[len(items)-1]
		next, err = h.cursors.Encode(cursor{WorkspaceID: workspaceID, ActorID: actorID, TeamID: teamID, Kind: filters.Kind, Position: domain.Position{CreatedAt: last.CreatedAt, ID: last.ID}, ExpiresAt: time.Now().Add(24 * time.Hour)})
		if err != nil {
			return respondError(ctx, w, err)
		}
	}
	return web.Respond(ctx, w, struct {
		Items      []domain.Preset `json:"items"`
		NextCursor string          `json:"nextCursor"`
	}{items, next}, http.StatusOK)
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, workspaceID, err := identities(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	var input domain.Input
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	preset, err := h.service.Create(ctx, actorID, workspaceID, input)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, preset, http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, workspaceID, err := identities(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
	}
	var input domain.Update
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	preset, err := h.service.Update(ctx, actorID, workspaceID, id, input)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, preset, http.StatusOK)
}

func (h *Handlers) Archive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	actorID, workspaceID, err := identities(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return web.RespondError(ctx, w, domain.ErrInvalidInput, http.StatusBadRequest)
	}
	if err := h.service.Archive(ctx, actorID, workspaceID, id); err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func identities(ctx context.Context) (uuid.UUID, uuid.UUID, error) {
	actorID, err := mid.GetUserID(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	return actorID, workspace.ID, nil
}

func respondError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	if errors.Is(err, domain.ErrNotFound) {
		status = http.StatusNotFound
	}
	if errors.Is(err, domain.ErrInvalidInput) {
		status = http.StatusBadRequest
	}
	return web.RespondError(ctx, w, err, status)
}
