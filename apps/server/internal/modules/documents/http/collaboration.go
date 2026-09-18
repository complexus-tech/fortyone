package documentshttp

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func privateDocumentResponse(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Robots-Tag", "noindex, nofollow, noarchive")
}

func (h *Handlers) ListRevisions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var before int64
	if value := r.URL.Query().Get("before"); value != "" {
		before, err = strconv.ParseInt(value, 10, 64)
		if err != nil || before < 1 {
			return web.RespondError(ctx, w, documents.ErrInvalidInput, http.StatusBadRequest)
		}
	}
	result, err := h.documents.ListRevisions(ctx, workspace.ID, userID, documentID, before)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) GetRevision(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	revision, err := strconv.ParseInt(web.Params(r, "revision"), 10, 64)
	if err != nil {
		return web.RespondError(ctx, w, documents.ErrInvalidInput, http.StatusBadRequest)
	}
	result, err := h.documents.GetRevision(ctx, workspace.ID, userID, documentID, revision)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) RestoreRevision(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input struct {
		Revision         int64 `json:"revision"`
		ExpectedRevision int64 `json:"expectedRevision"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.documents.RestoreRevision(ctx, workspace.ID, userID, documentID, input.Revision, input.ExpectedRevision)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(result, canMutateDocuments(workspace)), http.StatusOK)
}

func (h *Handlers) SetPublicLink(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input struct {
		Enabled bool `json:"enabled"`
	}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.documents.SetPublicLink(ctx, workspace.ID, userID, documentID, input.Enabled)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(result, canMutateDocuments(workspace)), http.StatusOK)
}

func (h *Handlers) PublicDocument(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	result, err := h.documents.GetPublicDocument(ctx, web.Params(r, "token"))
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, result, http.StatusOK)
}

func (h *Handlers) PublicMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	attachmentID, err := attachmentIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, documents.ErrNotFound, http.StatusNotFound)
	}
	document, err := h.documents.AuthorizePublicMedia(ctx, web.Params(r, "token"), attachmentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	file, err := h.attachments.ResolveAttachmentAccessURL(ctx, attachmentID, document.WorkspaceID, time.Minute)
	if err != nil {
		return web.RespondError(ctx, w, err, attachmentHTTPStatus(err))
	}
	if !strings.HasPrefix(file.MimeType, "image/") && file.MimeType != "video/mp4" {
		return web.RespondError(ctx, w, documents.ErrNotFound, http.StatusNotFound)
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Redirect(w, r, file.URL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) CollaborationSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	privateDocumentResponse(w)
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	result, err := h.documents.CreateCollaborationSession(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, result, http.StatusCreated)
}
