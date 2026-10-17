package nativepreviewhttp

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	domain "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/domain"
	service "github.com/complexus-tech/projects-api/internal/modules/nativepreviews/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *service.Service
}

func Routes(cfg Config, app *web.App) {
	handler := &Handlers{service: cfg.Service}
	app.Get("/workspaces/{workspaceSlug}/previews/native/{kind}/{id}", handler.Get,
		mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions), mid.Workspace(cfg.Log, cfg.WorkspaceResolver))
}

type Handlers struct{ service *service.Service }

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	w.Header().Set("Cache-Control", "private, no-store")
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return web.RespondError(ctx, w, domain.ErrForbidden, http.StatusForbidden)
	}
	target, err := parseTarget(web.Params(r, "kind"), web.Params(r, "id"), r.URL.Query().Get("teamId"))
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	preview, err := h.service.Get(ctx, workspace.ID, target, r.URL.Query().Get("refresh") == "true")
	if err != nil {
		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, domain.ErrNotFound):
			status = http.StatusNotFound
		case errors.Is(err, domain.ErrForbidden):
			status = http.StatusForbidden
		case errors.Is(err, domain.ErrInvalidTarget):
			status = http.StatusBadRequest
		}
		return web.RespondError(ctx, w, err, status)
	}
	return web.Respond(ctx, w, preview, http.StatusOK)
}

var storyReference = regexp.MustCompile(`(?i)^([a-z][a-z0-9]*)-([1-9][0-9]*)$`)

func parseTarget(kind, id, teamID string) (domain.Target, error) {
	target := domain.Target{Kind: kind}
	if kind != "story" && kind != "objective" && kind != "document" {
		return target, domain.ErrInvalidTarget
	}
	parsedID, err := uuid.Parse(id)
	if err != nil {
		parts := storyReference.FindStringSubmatch(id)
		if kind != "story" || len(parts) != 3 {
			return target, domain.ErrInvalidTarget
		}
		sequence, err := strconv.ParseInt(parts[2], 10, 32)
		if err != nil {
			return target, domain.ErrInvalidTarget
		}
		target.TeamCode, target.Sequence = strings.ToUpper(parts[1]), int32(sequence)
	} else {
		if parsedID == uuid.Nil {
			return target, domain.ErrInvalidTarget
		}
		target.ID = parsedID
	}
	if kind == "objective" {
		parsedTeam, err := uuid.Parse(teamID)
		if err != nil || parsedTeam == uuid.Nil {
			return target, domain.ErrInvalidTarget
		}
		target.TeamID = parsedTeam
	}
	return target, nil
}
