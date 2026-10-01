package dataexporthttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	exportdomain "github.com/complexus-tech/projects-api/internal/modules/dataexport/domain"
	dataexport "github.com/complexus-tech/projects-api/internal/modules/dataexport/service"
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
	Service           *dataexport.Service
}

func Routes(config Config, app *web.App) {
	auth := mid.Auth(config.Log, config.SecretKey, config.BrowserSessions)
	workspace := mid.Workspace(config.Log, config.WorkspaceResolver)
	app.Get("/workspaces/{workspaceSlug}/exports/work", handler(config.Service), auth, workspace)
}

func handler(service *dataexport.Service) web.Handler {
	return func(ctx context.Context, writer http.ResponseWriter, request *http.Request) error {
		workspace, err := mid.GetWorkspace(ctx)
		if err != nil {
			return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
		}
		actorID, err := mid.GetUserID(ctx)
		if err != nil {
			return web.RespondError(ctx, writer, err, http.StatusUnauthorized)
		}
		if workspace.UserRole != "admin" {
			return web.RespondError(ctx, writer, exportdomain.ErrForbidden, http.StatusForbidden)
		}
		var teamID *uuid.UUID
		if raw := request.URL.Query().Get("teamId"); raw != "" {
			parsed, err := uuid.Parse(raw)
			if err != nil || parsed == uuid.Nil {
				return web.RespondError(ctx, writer, exportdomain.ErrInvalid, http.StatusBadRequest)
			}
			teamID = &parsed
		}
		ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		envelope, err := service.Export(ctx, exportdomain.Scope{ActorID: actorID, WorkspaceID: workspace.ID, TeamID: teamID})
		if err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, exportdomain.ErrForbidden) {
				status = http.StatusForbidden
			}
			if errors.Is(err, exportdomain.ErrTooLarge) {
				status = http.StatusRequestEntityTooLarge
			}
			return web.RespondError(ctx, writer, err, status)
		}
		writer.Header().Set("Cache-Control", "no-store")
		return web.Respond(ctx, writer, envelope, http.StatusOK)
	}
}
