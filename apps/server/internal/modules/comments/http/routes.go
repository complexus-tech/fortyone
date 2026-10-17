package commentshttp

import (
	comments "github.com/complexus-tech/projects-api/internal/modules/comments/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *comments.Service
}

func Routes(cfg Config, app *web.App) {
	commentsService := cfg.Service
	h := New(commentsService)
	collaboration := newCollaborationHandlers(commentsService)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	writeScope := mid.RequireScopes(platformauth.ScopeCommentsWrite)
	readScope := mid.RequireScopes(platformauth.ScopeCommentsRead)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)

	app.Put("/workspaces/{workspaceSlug}/comments/{id}", h.UpdateComment, auth, writeScope, workspace)
	app.Delete("/workspaces/{workspaceSlug}/comments/{id}", h.DeleteComment, auth, writeScope, workspace)
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/comments/state", collaboration.GetState, auth, readScope, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{storyId}/comments/read", collaboration.MarkRead, auth, readScope, workspace)
	app.Put("/workspaces/{workspaceSlug}/stories/{storyId}/comments/{commentId}/seen", collaboration.MarkSeen, auth, readScope, workspace)
	app.Put("/workspaces/{workspaceSlug}/comments/{id}/reactions", collaboration.SetReaction, auth, writeScope, workspace)
	app.Delete("/workspaces/{workspaceSlug}/comments/{id}/reactions", collaboration.SetReaction, auth, writeScope, workspace)
	app.Put("/workspaces/{workspaceSlug}/comments/{id}/resolution", collaboration.SetResolution, auth, writeScope, workspace)
}
