package chathttp

import (
	"context"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	chatservice "github.com/complexus-tech/projects-api/internal/modules/chat/service"
	platformauth "github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Service           *chatservice.Service
	ProfileImages     ProfileImageResolver
}

type handlers struct {
	service       *chatservice.Service
	log           *logger.Logger
	profileImages ProfileImageResolver
}

type endpoint func(context.Context, http.ResponseWriter, *http.Request, domain.ActorScope) error

func scoped(handler endpoint) web.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		workspace, err := mid.GetWorkspace(ctx)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		}
		actor, err := mid.GetActor(ctx)
		if err != nil {
			return web.RespondError(ctx, w, err, http.StatusUnauthorized)
		}
		return handler(ctx, w, r, domain.ActorScope{WorkspaceID: workspace.ID, Actor: actor})
	}
}

func Routes(cfg Config, app *web.App) {
	h := &handlers{service: cfg.Service, log: cfg.Log, profileImages: cfg.ProfileImages}
	auth := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	rollout := mid.RequireChatEnabled()
	app.Get("/workspaces/{workspaceSlug}/chat/capabilities", scoped(func(ctx context.Context, w http.ResponseWriter, _ *http.Request, scope domain.ActorScope) error {
		workspace, _ := mid.GetWorkspace(ctx)
		w.Header().Set("Cache-Control", "private, no-store")
		return web.Respond(ctx, w, struct {
			Enabled            bool `json:"enabled"`
			AttachmentsEnabled bool `json:"attachmentsEnabled"`
		}{Enabled: workspace.ChatEnabled, AttachmentsEnabled: workspace.ChatEnabled}, http.StatusOK)
	}), auth, workspace)
	read := mid.RequireScopes(platformauth.ScopeChatRead)
	write := mid.RequireScopes(platformauth.ScopeChatWrite)
	app.Get("/workspaces/{workspaceSlug}/chat/unread", scoped(h.unread), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/threads/unread", scoped(h.unreadThreads), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations", scoped(h.listConversations), auth, read, workspace, rollout)
	app.Post("/workspaces/{workspaceSlug}/chat/conversations/direct", scoped(h.createDirect), auth, write, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}", scoped(h.getConversation), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/members", scoped(h.listMembers), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/messages", scoped(h.listMessages), auth, read, workspace, rollout)
	app.Post("/workspaces/{workspaceSlug}/chat/conversations/{id}/messages", scoped(h.createMessage), auth, write, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/messages/{id}", scoped(h.getMessage), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/messages/{id}/replies", scoped(h.listReplies), auth, read, workspace, rollout)
	app.Patch("/workspaces/{workspaceSlug}/chat/messages/{id}", scoped(h.updateMessage), auth, write, workspace, rollout)
	app.Delete("/workspaces/{workspaceSlug}/chat/messages/{id}", scoped(h.deleteMessage), auth, write, workspace, rollout)
	app.Put("/workspaces/{workspaceSlug}/chat/messages/{id}/reactions/{emoji}", scoped(h.setReaction), auth, write, workspace, rollout)
	app.Delete("/workspaces/{workspaceSlug}/chat/messages/{id}/reactions/{emoji}", scoped(h.setReaction), auth, write, workspace, rollout)
	app.Post("/workspaces/{workspaceSlug}/chat/conversations/{id}/read", scoped(h.markRead), auth, write, workspace, rollout)
	app.Post("/workspaces/{workspaceSlug}/chat/conversations/{id}/typing", scoped(h.typing), auth, write, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/search", scoped(h.search), auth, read, workspace, rollout)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/resources", scoped(h.resources), auth, read, workspace, rollout)
}
