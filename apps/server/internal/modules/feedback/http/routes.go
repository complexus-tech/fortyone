package feedbackhttp

import (
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

type Config struct {
	DB          *sqlx.DB
	Log         *logger.Logger
	SecretKey   string
	Cache       *cache.Service
	Service     *feedback.Service
	Attachments *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Attachments, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey)
	createItemRateLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope:  "public-feedback-item",
		Limit:  10,
		Window: time.Hour,
	})
	createCommentRateLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope:  "public-feedback-comment",
		Limit:  60,
		Window: time.Hour,
	})
	voteRateLimit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{
		Scope:  "public-feedback-vote",
		Limit:  120,
		Window: time.Minute,
	})
	workspace := mid.Workspace(cfg.Log, cfg.DB, cfg.Cache)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)

	app.Get("/portals/{portalSlug}/feedback", h.GetPortal)
	app.Post("/portals/{portalSlug}/feedback/items", h.CreatePublicItem, auth, createItemRateLimit)
	app.Post("/portals/{portalSlug}/feedback/items/{itemId}/comments", h.CreatePublicComment, auth, createCommentRateLimit)
	app.Post("/portals/{portalSlug}/feedback/items/{itemId}/vote", h.TogglePublicVote, auth, voteRateLimit)
	app.Get("/workspaces/{workspaceSlug}/portals/{portalSlug}/feedback", h.GetWorkspacePortal)
	app.Get("/workspaces/{workspaceSlug}/feedback/portals", h.ListPortals, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/portals/{portalId}", h.UpdatePortal, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/boards", h.CreateBoard, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items", h.CreateItem, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/items/{itemId}/status", h.UpdateItemStatus, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/comments", h.CreateComment, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/vote", h.ToggleVote, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/story", h.CreateStoryFromItem, auth, workspace, adminOnly)
}
