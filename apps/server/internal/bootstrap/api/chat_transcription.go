package api

import (
	"context"
	"net/http"
	"time"

	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func registerChatTranscriptions(cfg mux.Config, resolver mid.WorkspaceResolver, sessions mid.SessionResolver, svcs services, app *web.App) {
	authenticate := mid.Auth(cfg.Log, cfg.SecretKey, sessions)
	workspace := mid.Workspace(cfg.Log, resolver)
	rollout := mid.RequireChatEnabled()
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	limit := mid.AuthenticatedUserRateLimit(cfg.Log, cfg.Cache, mid.AuthenticatedUserRateLimitConfig{Scope: "chat-transcription", Limit: 12, Window: time.Minute})
	const path = "/workspaces/{workspaceSlug}/chat/conversations/{id}/attachments/{attachmentId}/transcription"
	app.Get(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		scope, id, err := chatAttachmentRequest(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		result, err := svcs.chatTranscriptions.Get(ctx, scope, id)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		w.Header().Set("Cache-Control", "private, no-store")
		return web.Respond(ctx, w, result, http.StatusOK)
	}, authenticate, workspace, rollout)
	app.Post(path, func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		scope, id, err := chatAttachmentRequest(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		var input struct {
			Retry bool `json:"retry"`
		}
		if err := web.DecodeWithLimit(r, &input, 1024); err != nil {
			return chatAttachmentError(ctx, w, chat.ErrInvalid)
		}
		result, err := svcs.chatTranscriptions.Request(ctx, scope, id, input.Retry)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		status := http.StatusOK
		if result.Status == "pending" || result.Status == "processing" {
			status = http.StatusAccepted
		}
		w.Header().Set("Cache-Control", "private, no-store")
		return web.Respond(ctx, w, result, status)
	}, authenticate, workspace, rollout, member, limit)
}
