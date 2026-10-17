package api

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func registerChatAttachmentThumbnails(cfg mux.Config, resolver mid.WorkspaceResolver, sessions mid.SessionResolver, svcs services, app *web.App) {
	authenticate := mid.Auth(cfg.Log, cfg.SecretKey, sessions)
	workspace := mid.Workspace(cfg.Log, resolver)
	rollout := mid.RequireChatEnabled()
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/attachments/{attachmentId}/thumbnail", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		scope, id, err := chatAttachmentRequest(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		// Authorization is deliberately before conditional-cache handling. A 304
		// must never allow a removed member to reuse previously cached pixels.
		if err := svcs.chat.AuthorizeAttachment(ctx, scope, id); err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		thumbnail, err := svcs.attachments.ResolveThumbnail(ctx, id, scope.WorkspaceID, r.URL.Query().Get("refresh") == "1")
		if errors.Is(err, attachments.ErrThumbnailPending) {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusAccepted)
			return nil
		}
		if errors.Is(err, attachments.ErrThumbnailUnavailable) {
			return web.RespondError(ctx, w, errors.New("Preview unavailable; open the original file to view it."), http.StatusUnprocessableEntity)
		}
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		etag := `"` + thumbnail.Version + `"`
		w.Header().Set("ETag", etag)
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
		w.Header().Set("X-Thumbnail-Bytes", strconv.Itoa(int(thumbnail.Size)))
		http.Redirect(w, r, thumbnail.URL, http.StatusTemporaryRedirect)
		return nil
	}, authenticate, workspace, rollout)
}
