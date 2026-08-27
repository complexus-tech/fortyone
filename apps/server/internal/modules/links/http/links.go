package linkshttp

import (
	"context"
	"errors"
	"net/http"

	links "github.com/complexus-tech/projects-api/internal/modules/links/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidLinkID  = errors.New("link id is not in its proper form")
	ErrInvalidStoryID = errors.New("story id is not in its proper form")
)

func mutationStatus(err error) int {
	if errors.Is(err, links.ErrNotFound) {
		return http.StatusNotFound
	}
	return http.StatusInternalServerError
}

type Handlers struct {
	links *links.Service
	log   *logger.Logger
}

func New(log *logger.Logger, links *links.Service) *Handlers {
	return &Handlers{
		links: links,
		log:   log,
	}
}

func (h *Handlers) CreateLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	var nl NewLink
	if err := web.Decode(r, &nl); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	link, err := h.links.CreateLink(ctx, toCoreNewLink(nl, workspace.ID))
	if err != nil {
		web.RespondError(ctx, w, err, mutationStatus(err))
		return nil
	}

	return web.Respond(ctx, w, toLink(link), http.StatusCreated)
}

func (h *Handlers) UpdateLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	linkIdParam := web.Params(r, "id")
	linkID, err := uuid.Parse(linkIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLinkID, http.StatusBadRequest)
		return nil
	}

	var ul UpdateLink
	if err := web.Decode(r, &ul); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.links.UpdateLink(ctx, linkID, workspace.ID, toCoreUpdateLink(ul)); err != nil {
		web.RespondError(ctx, w, err, mutationStatus(err))
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) DeleteLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	linkIdParam := web.Params(r, "id")
	linkID, err := uuid.Parse(linkIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidLinkID, http.StatusBadRequest)
		return nil
	}

	if err := h.links.DeleteLink(ctx, linkID, workspace.ID); err != nil {
		web.RespondError(ctx, w, err, mutationStatus(err))
		return nil
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}
