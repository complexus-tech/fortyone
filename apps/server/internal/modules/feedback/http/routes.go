package feedbackhttp

import (
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	feedback "github.com/complexus-tech/projects-api/internal/modules/feedback/service"
	teams "github.com/complexus-tech/projects-api/internal/modules/teams/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/cache"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	IngressSecret     string
	Cache             *cache.Service
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *feedback.Service
	Teams             *teams.Service
	Attachments       *attachments.Service
}

func Routes(cfg Config, app *web.App) {
	h := New(cfg.Service, cfg.Teams, cfg.Attachments, cfg.Log)
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	optionalAuth := mid.OptionalAuth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	createItemRateLimit := mid.PublicFeedbackRateLimit(cfg.Log, cfg.Cache, mid.PublicFeedbackRateLimitConfig{
		Scope:                "public-feedback-item",
		AuthenticatedLimit:   10,
		AnonymousLimit:       3,
		AnonymousGlobalLimit: 12,
		Window:               time.Hour,
		IngressSecret:        cfg.IngressSecret,
		ContributorResolver:  cfg.Service.ResolveContributorRateLimitIdentity,
	})
	createCommentRateLimit := mid.PublicFeedbackRateLimit(cfg.Log, cfg.Cache, mid.PublicFeedbackRateLimitConfig{
		Scope: "public-feedback-comment", AuthenticatedLimit: 60, AnonymousLimit: 30, AnonymousGlobalLimit: 120,
		Window: time.Hour, IngressSecret: cfg.IngressSecret,
		ContributorResolver: cfg.Service.ResolveContributorRateLimitIdentity,
	})
	voteRateLimit := mid.PublicFeedbackRateLimit(cfg.Log, cfg.Cache, mid.PublicFeedbackRateLimitConfig{
		Scope: "public-feedback-vote", AuthenticatedLimit: 120, AnonymousLimit: 120, AnonymousGlobalLimit: 300,
		Window: time.Hour, IngressSecret: cfg.IngressSecret,
		ContributorResolver: cfg.Service.ResolveContributorRateLimitIdentity,
	})
	identityRateLimit := mid.PublicFeedbackRateLimit(cfg.Log, cfg.Cache, mid.PublicFeedbackRateLimitConfig{
		Scope: "public-feedback-identity", AuthenticatedLimit: 30, AnonymousLimit: 10, AnonymousGlobalLimit: 30,
		Window: time.Hour, IngressSecret: cfg.IngressSecret,
		ContributorResolver: cfg.Service.ResolveContributorRateLimitIdentity,
	})
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	adminOnly := mid.RequireMinimumRole(cfg.Log, mid.RoleAdmin)
	memberAndAdmin := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)

	app.Get("/portals/{portalSlug}/feedback", h.GetPortal)
	app.Get("/portals/{portalSlug}/feedback/similar", h.ListPublicSimilarItems)
	app.Get("/portals/{portalSlug}/feedback/updates", h.ListPublicUpdates)
	app.Post("/portals/{portalSlug}/feedback/updates/seen", h.MarkUpdatesSeen)
	app.Get("/portals/{portalSlug}/feedback/updates/{updateSlug}", h.GetPublicUpdate)
	app.Post("/portals/{portalSlug}/feedback/verifications", h.RequestContributorVerification, optionalAuth, identityRateLimit)
	app.Post("/portals/{portalSlug}/feedback/verifications/confirm", h.ConfirmContributorVerification, optionalAuth, identityRateLimit)
	app.Get("/portals/{portalSlug}/feedback/session", h.GetContributorSession)
	app.Post("/portals/{portalSlug}/feedback/sessions/revoke", h.RevokeContributorSession)
	app.Post("/portals/{portalSlug}/feedback/preferences/exchange", h.ExchangePreferenceToken)
	app.Get("/portals/{portalSlug}/feedback/preferences", h.GetContributorPreferences)
	app.Put("/portals/{portalSlug}/feedback/preferences", h.UpdateContributorPreferences)
	app.Get("/portals/{portalSlug}/feedback/widget/config", h.GetPublicWidgetConfig)
	app.Post("/portals/{portalSlug}/feedback/widget/sessions", h.CreateWidgetContributorSession, identityRateLimit)
	app.Get("/portals/{portalSlug}/feedback/contributors/{authorId}", h.GetPublicContributor)
	app.Get("/portals/{portalSlug}/feedback/contributors/{authorId}/comments", h.ListPublicContributorComments)
	app.Get("/feedback/contributor/activity", h.ListContributorActivity, auth)
	app.Post("/portals/{portalSlug}/feedback/items", h.CreatePublicItem, optionalAuth, createItemRateLimit)
	app.Post("/portals/{portalSlug}/widget/feedback/items", h.CreateWidgetItem, optionalAuth, createItemRateLimit)
	app.Get("/portals/{portalSlug}/feedback/items/{itemId}/attachments/{attachmentId}", h.ResolvePublicItemAttachment)
	app.Post("/portals/{portalSlug}/feedback/items/{itemId}/comments", h.CreatePublicComment, optionalAuth, createCommentRateLimit)
	app.Post("/portals/{portalSlug}/feedback/items/{itemId}/vote", h.TogglePublicVote, optionalAuth, voteRateLimit)
	app.Get("/portals/{portalSlug}/feedback/items/{itemId}/follow", h.GetItemFollow, optionalAuth)
	app.Get("/portals/{portalSlug}/feedback/items/{itemReference}/canonical", h.ResolveCanonicalItem)
	app.Put("/portals/{portalSlug}/feedback/items/{itemId}/follow", h.FollowItem, optionalAuth)
	app.Delete("/portals/{portalSlug}/feedback/items/{itemId}/follow", h.UnfollowItem, optionalAuth)
	app.Get("/workspaces/{workspaceSlug}/portals/{portalSlug}/feedback", h.GetWorkspacePortal)
	app.Get("/workspaces/{workspaceSlug}/portals/{portalSlug}/feedback/contributors/{authorId}", h.GetWorkspacePublicContributor)
	app.Get("/workspaces/{workspaceSlug}/portals/{portalSlug}/feedback/contributors/{authorId}/comments", h.ListWorkspacePublicContributorComments)
	app.Get("/workspaces/{workspaceSlug}/feedback/portals", h.ListPortals, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/feedback/team-summaries", h.ListTeamSummaries, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/teams/{teamId}/feedback", h.ListTeamItems, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/feedback/items/{itemId}", h.GetItem, auth, workspace)
	app.Get("/workspaces/{workspaceSlug}/feedback/items/{itemId}/private-author", h.GetPrivateAuthor, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/stories/{storyId}/feedback-links", h.GetStoryFeedbackLinks, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/portals/{portalId}", h.UpdatePortal, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/updates", h.ListWorkspaceUpdates, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/updates", h.CreateUpdate, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/updates/{updateId}", h.GetWorkspaceUpdate, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/feedback/updates/{updateId}", h.UpdateUpdate, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/feedback/updates/{updateId}", h.DeleteUpdate, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/updates/{updateId}/publish", h.PublishUpdate, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/updates/{updateId}/unpublish", h.UnpublishUpdate, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/portals/{portalId}/item-candidates", h.ListPortalItemCandidates, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/items/{sourceItemId}/merge-candidates", h.ListMergeCandidates, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{sourceItemId}/merge", h.MergeItem, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/portals/{portalId}/widget-settings", h.GetWidgetSettings, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/feedback/portals/{portalId}/widget-settings", h.UpdateWidgetSettings, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/portals/{portalId}/widget-settings/signing-secret", h.CreateWidgetSigningSecret, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/portals/{portalId}/widget-settings/signing-secret/rotate", h.RotateWidgetSigningSecret, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/boards", h.CreateBoard, auth, workspace, adminOnly)
	app.Delete("/workspaces/{workspaceSlug}/feedback/boards/{boardId}", h.DeleteBoard, auth, workspace, adminOnly)
	app.Get("/workspaces/{workspaceSlug}/feedback/boards/{boardId}/reviewers", h.ListBoardReviewers, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/feedback/boards/{boardId}/reviewers/{userId}", h.SetBoardReviewer, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items", h.CreateItem, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/items/{itemId}/status", h.UpdateItemStatus, auth, workspace, memberAndAdmin)
	app.Delete("/workspaces/{workspaceSlug}/feedback/items/{itemId}", h.TrashItem, auth, workspace, adminOnly)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/restore", h.RestoreItem, auth, workspace, adminOnly)
	app.Put("/workspaces/{workspaceSlug}/feedback/items/{itemId}/read", h.MarkItemRead, auth, workspace)
	app.Put("/workspaces/{workspaceSlug}/feedback/items/{itemId}/unread", h.MarkItemUnread, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/comments", h.CreateComment, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/vote", h.ToggleVote, auth, workspace)
	app.Post("/workspaces/{workspaceSlug}/feedback/items/{itemId}/story", h.CreateStoryFromItem, auth, workspace, memberAndAdmin)
}
