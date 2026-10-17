package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	chat "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/internal/platform/http/mux"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type chatAttachmentResponse struct {
	ID         uuid.UUID `json:"id"`
	FileName   string    `json:"fileName"`
	MediaType  string    `json:"mediaType"`
	SizeBytes  int64     `json:"sizeBytes"`
	URL        string    `json:"url"`
	Kind       string    `json:"kind,omitempty"`
	DurationMS *int32    `json:"durationMs,omitempty"`
}

// The composition root joins chat authorization with the shared, scanned media
// lifecycle. A storage URL is never used as proof of conversation membership.
func registerChatAttachments(cfg mux.Config, resolver mid.WorkspaceResolver, sessions mid.SessionResolver, svcs services, app *web.App) {
	authenticate := mid.Auth(cfg.Log, cfg.SecretKey, sessions)
	workspace := mid.Workspace(cfg.Log, resolver)
	rollout := mid.RequireChatEnabled()
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	app.Post("/workspaces/{workspaceSlug}/chat/conversations/{id}/attachments", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		scope, err := chatAttachmentScope(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		if !scope.Actor.Scopes.Has(auth.ScopeChatWrite) {
			return chatAttachmentError(ctx, w, chat.ErrForbidden)
		}
		if _, err := svcs.chat.GetConversation(ctx, scope.ActorScope, scope.ConversationID); err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		if err := web.ParseMultipartForm(w, r, validate.MaxAttachmentSize+(1<<20)); err != nil {
			return web.RespondError(ctx, w, errors.New("invalid upload request"), http.StatusBadRequest)
		}
		defer func() {
			if err := web.RemoveMultipartForm(r); err != nil {
				cfg.Log.Warn(ctx, "remove chat upload temporary files", "error", err)
			}
		}()
		file, header, err := r.FormFile("file")
		if err != nil {
			return web.RespondError(ctx, w, errors.New("attachment file is required"), http.StatusBadRequest)
		}
		defer file.Close()
		metadata, err := parseChatAttachmentMetadata(r.FormValue("kind"), r.FormValue("durationMs"))
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		var uploaded attachments.FileInfo
		if metadata.Kind == "voice-note" {
			uploaded, err = svcs.attachments.UploadChatVoiceNote(ctx, file, header, scope.Actor.PrincipalID, scope.WorkspaceID, int64(*metadata.DurationMS))
		} else {
			uploaded, err = svcs.attachments.UploadChatAttachment(ctx, file, header, scope.Actor.PrincipalID, scope.WorkspaceID)
		}
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		if err := svcs.chat.LinkPendingAttachment(ctx, scope, uploaded.ID, metadata); err != nil {
			cleanup, cancel := context.WithTimeout(context.WithoutCancel(ctx), 10*time.Second)
			defer cancel()
			if cleanupErr := svcs.chatAttachments.RetireUnlinkedUpload(cleanup, scope.WorkspaceID, uploaded.ID); cleanupErr != nil {
				cfg.Log.Error(ctx, "clean up unlinked chat attachment", "error", cleanupErr, "attachment_id", uploaded.ID)
			}
			return chatAttachmentError(ctx, w, err)
		}
		stableURL := fmt.Sprintf("/workspaces/%s/chat/conversations/%s/attachments/%s", url.PathEscape(web.Params(r, "workspaceSlug")), scope.ConversationID, uploaded.ID)
		return web.Respond(ctx, w, chatAttachmentResponse{ID: uploaded.ID, FileName: uploaded.Filename, MediaType: uploaded.MimeType, SizeBytes: uploaded.Size, URL: stableURL, Kind: metadata.Kind, DurationMS: metadata.DurationMS}, http.StatusCreated)
	}, authenticate, workspace, rollout, member)
	app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/attachments/{attachmentId}", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		scope, id, err := chatAttachmentRequest(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		if err := svcs.chat.AuthorizeAttachment(ctx, scope, id); err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		file, err := svcs.attachments.ResolveAttachmentAccessURL(ctx, id, scope.WorkspaceID, time.Minute)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		w.Header().Set("Cache-Control", "private, no-store")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		http.Redirect(w, r, file.URL, http.StatusTemporaryRedirect)
		return nil
	}, authenticate, workspace, rollout)
	app.Delete("/workspaces/{workspaceSlug}/chat/conversations/{id}/attachments/{attachmentId}", func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		scope, id, err := chatAttachmentRequest(ctx, r)
		if err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		if err := svcs.chat.UnlinkPendingAttachment(ctx, scope, id); err != nil {
			return chatAttachmentError(ctx, w, err)
		}
		return web.Respond(ctx, w, nil, http.StatusNoContent)
	}, authenticate, workspace, rollout, member)
}

func parseChatAttachmentMetadata(kind, duration string) (chat.AttachmentMetadata, error) {
	if kind == "" && duration == "" {
		return chat.AttachmentMetadata{}, nil
	}
	if kind != "voice-note" {
		return chat.AttachmentMetadata{}, chat.ErrInvalid
	}
	milliseconds, err := strconv.ParseInt(duration, 10, 32)
	if err != nil || milliseconds < 1 || milliseconds > validate.MaxVoiceNoteDurationMS {
		return chat.AttachmentMetadata{}, fmt.Errorf("%w: voice notes must be five minutes or shorter", chat.ErrInvalid)
	}
	value := int32(milliseconds)
	return chat.AttachmentMetadata{Kind: kind, DurationMS: &value}, nil
}
func chatAttachmentScope(ctx context.Context, r *http.Request) (chat.AttachmentScope, error) {
	actor, err := auth.GetActor(ctx)
	if err != nil {
		return chat.AttachmentScope{}, chat.ErrForbidden
	}
	id, err := uuid.Parse(web.Params(r, "id"))
	if err != nil || id == uuid.Nil {
		return chat.AttachmentScope{}, chat.ErrInvalid
	}
	return chat.AttachmentScope{ActorScope: chat.ActorScope{WorkspaceID: actor.WorkspaceID, Actor: actor}, ConversationID: id}, nil
}
func chatAttachmentRequest(ctx context.Context, r *http.Request) (chat.AttachmentScope, uuid.UUID, error) {
	scope, err := chatAttachmentScope(ctx, r)
	if err != nil {
		return scope, uuid.Nil, err
	}
	id, err := uuid.Parse(web.Params(r, "attachmentId"))
	if err != nil || id == uuid.Nil {
		return scope, uuid.Nil, chat.ErrInvalid
	}
	return scope, id, nil
}
func chatAttachmentError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	switch {
	case errors.Is(err, chat.ErrNotFound), errors.Is(err, attachments.ErrNotFound):
		status = http.StatusNotFound
	case errors.Is(err, chat.ErrForbidden), errors.Is(err, attachments.ErrUnauthorized):
		status = http.StatusForbidden
	case errors.Is(err, chat.ErrTranscriptionAccess):
		status = http.StatusForbidden
	case errors.Is(err, chat.ErrTranscriptionUnavailable):
		status = http.StatusServiceUnavailable
	case errors.Is(err, chat.ErrInvalid), errors.Is(err, attachments.ErrInvalidFile), errors.Is(err, attachments.ErrInvalidFileType):
		status = http.StatusBadRequest
	case errors.Is(err, chat.ErrConflict):
		status = http.StatusConflict
	case errors.Is(err, attachments.ErrFileTooLarge):
		status = http.StatusRequestEntityTooLarge
	}
	return web.RespondError(ctx, w, err, status)
}
