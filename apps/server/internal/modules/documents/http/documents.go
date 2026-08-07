package documentshttp

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	attachments "github.com/complexus-tech/projects-api/internal/modules/attachments/service"
	documents "github.com/complexus-tech/projects-api/internal/modules/documents/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

const (
	documentMediaAccessTTL           = 5 * time.Minute
	multipartOverheadAllowance int64 = 1 << 20
)

type attachmentMediaService interface {
	UploadDocumentMedia(context.Context, multipart.File, *multipart.FileHeader, uuid.UUID, uuid.UUID) (attachments.FileInfo, error)
	ResolveAttachmentAccessURL(context.Context, uuid.UUID, uuid.UUID, time.Duration) (attachments.FileInfo, error)
	DeleteAttachment(context.Context, uuid.UUID, uuid.UUID) error
	DeleteDocumentMedia(context.Context, uuid.UUID, uuid.UUID) error
}

type Handlers struct {
	documents   *documents.Service
	attachments attachmentMediaService
	log         *logger.Logger
}

func New(service *documents.Service, attachmentService attachmentMediaService, log *logger.Logger) *Handlers {
	return &Handlers{documents: service, attachments: attachmentService, log: log}
}

func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	limit, err := documentListLimit(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	result, err := h.documents.List(ctx, documents.CoreListInput{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Search:      r.URL.Query().Get("search"),
		Scope:       r.URL.Query().Get("scope"),
		Limit:       limit,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocumentSummaries(result, canMutateDocuments(workspace)), http.StatusOK)
}

func documentListLimit(r *http.Request) (*int, error) {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return nil, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit <= 0 {
		return nil, documents.ErrInvalidInput
	}
	return &limit, nil
}

func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	documentID, err := documentIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	document, err := h.documents.Get(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(document, canMutateDocuments(workspace)), http.StatusOK)
}

func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	input := AppCreateDocument{Visibility: string(documents.VisibilityWorkspace)}
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	document, err := h.documents.Create(ctx, documents.CoreCreateInput{
		WorkspaceID: workspace.ID,
		UserID:      userID,
		Title:       input.Title,
		Visibility:  documents.Visibility(input.Visibility),
		ContentHTML: input.ContentHTML,
		ContentText: input.ContentText,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(document, canMutateDocuments(workspace)), http.StatusCreated)
}

func (h *Handlers) Duplicate(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	document, err := h.documents.Duplicate(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(document, canMutateDocuments(workspace)), http.StatusCreated)
}

func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input AppUpdateDocument
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	document, err := h.documents.Update(ctx, documents.CoreUpdateInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		Title: input.Title, ContentHTML: input.ContentHTML, ContentText: input.ContentText,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(document, canMutateDocuments(workspace)), http.StatusOK)
}

func (h *Handlers) UploadMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	document, err := h.documents.Get(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	if !document.CanEdit {
		return web.RespondError(ctx, w, documents.ErrForbidden, http.StatusForbidden)
	}

	r.Body = http.MaxBytesReader(w, r.Body, validate.MaxAttachmentSize+multipartOverheadAllowance)
	if err := r.ParseMultipartForm(validate.MaxAttachmentSize); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			return web.RespondError(
				ctx,
				w,
				fmt.Errorf("%w: maximum document media size is 25 MB", attachments.ErrFileTooLarge),
				http.StatusRequestEntityTooLarge,
			)
		}
		return web.RespondError(ctx, w, fmt.Errorf("invalid upload request: %w", err), http.StatusBadRequest)
	}
	if r.MultipartForm != nil {
		defer r.MultipartForm.RemoveAll()
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("document media file is required: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	fileInfo, err := h.attachments.UploadDocumentMedia(ctx, file, header, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, err, attachmentHTTPStatus(err))
	}

	mediaInput := documents.CoreMediaInput{
		WorkspaceID:  workspace.ID,
		UserID:       userID,
		DocumentID:   documentID,
		AttachmentID: fileInfo.ID,
	}
	if err := h.documents.LinkMedia(ctx, mediaInput); err != nil {
		if cleanupErr := h.attachments.DeleteAttachment(ctx, fileInfo.ID, userID); cleanupErr != nil && h.log != nil {
			h.log.Error(ctx, "failed to clean up unlinked document media", "error", cleanupErr, "attachment_id", fileInfo.ID)
		}
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}

	stableURL := documentMediaURL(workspace.Slug, documentID, fileInfo.ID)
	return web.Respond(ctx, w, toAppDocumentMedia(fileInfo, stableURL), http.StatusCreated)
}

func (h *Handlers) ResolveMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	attachmentID, err := attachmentIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	mediaInput := documents.CoreMediaInput{
		WorkspaceID:  workspace.ID,
		UserID:       userID,
		DocumentID:   documentID,
		AttachmentID: attachmentID,
	}
	if err := h.documents.AuthorizeMedia(ctx, mediaInput); err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	fileInfo, err := h.attachments.ResolveAttachmentAccessURL(ctx, attachmentID, workspace.ID, documentMediaAccessTTL)
	if err != nil {
		return web.RespondError(ctx, w, err, attachmentHTTPStatus(err))
	}
	if !strings.HasPrefix(fileInfo.MimeType, "image/") && fileInfo.MimeType != "video/mp4" {
		return web.RespondError(ctx, w, attachments.ErrNotFound, http.StatusNotFound)
	}

	w.Header().Set("Cache-Control", "private, no-store")
	w.Header().Set("Referrer-Policy", "no-referrer")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	http.Redirect(w, r, fileInfo.URL, http.StatusTemporaryRedirect)
	return nil
}

func (h *Handlers) DeleteMedia(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	attachmentID, err := attachmentIDFromRequest(r)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	isOrphaned, err := h.documents.UnlinkMedia(ctx, documents.CoreMediaInput{
		WorkspaceID:  workspace.ID,
		UserID:       userID,
		DocumentID:   documentID,
		AttachmentID: attachmentID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	if isOrphaned {
		if err := h.attachments.DeleteDocumentMedia(ctx, attachmentID, workspace.ID); err != nil && h.log != nil {
			h.log.Error(ctx, "failed to clean up unlinked document media", "error", err, "attachment_id", attachmentID)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Archive(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	if err := h.documents.Archive(ctx, workspace.ID, userID, documentID); err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	orphanedAttachmentIDs, err := h.documents.Delete(ctx, workspace.ID, userID, documentID)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	for _, attachmentID := range orphanedAttachmentIDs {
		if err := h.attachments.DeleteDocumentMedia(ctx, attachmentID, workspace.ID); err != nil && h.log != nil {
			h.log.Error(ctx, "failed to clean up deleted document media", "error", err, "attachment_id", attachmentID)
		}
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) SetAccess(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input AppDocumentAccess
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	members := make([]documents.CoreDocumentMember, len(input.Members))
	for i, member := range input.Members {
		members[i] = documents.CoreDocumentMember{UserID: member.UserID, Role: member.Role}
	}
	document, err := h.documents.SetAccess(ctx, documents.CoreAccessInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		Visibility: documents.Visibility(input.Visibility), Members: members,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocument(document, canMutateDocuments(workspace)), http.StatusOK)
}

func (h *Handlers) AddRelationship(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	var input AppDocumentRelationship
	if err := web.Decode(r, &input); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	related, err := h.documents.AddRelationship(ctx, documents.CoreRelationshipInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		EntityType: documents.RelationshipType(input.EntityType), EntityID: input.EntityID,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppRelatedWork(related), http.StatusCreated)
}

func (h *Handlers) RemoveRelationship(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, documentID, err := mutationContext(ctx, r)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	entityID, err := uuid.Parse(web.Params(r, "entityId"))
	if err != nil {
		return web.RespondError(ctx, w, documents.ErrInvalidInput, http.StatusBadRequest)
	}
	if err := h.documents.RemoveRelationship(ctx, documents.CoreRelationshipInput{
		WorkspaceID: workspace.ID, UserID: userID, DocumentID: documentID,
		EntityType: documents.RelationshipType(web.Params(r, "entityType")), EntityID: entityID,
	}); err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func (h *Handlers) ListRelatedDocuments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	entityID, err := uuid.Parse(r.URL.Query().Get("entityId"))
	if err != nil {
		return web.RespondError(ctx, w, documents.ErrInvalidInput, http.StatusBadRequest)
	}
	result, err := h.documents.ListRelatedDocuments(
		ctx, workspace.ID, userID,
		documents.RelationshipType(r.URL.Query().Get("entityType")), entityID,
	)
	if err != nil {
		return web.RespondError(ctx, w, err, documentHTTPStatus(err))
	}
	return web.Respond(ctx, w, toAppDocumentSummaries(result, canMutateDocuments(workspace)), http.StatusOK)
}

func requestContext(ctx context.Context) (mid.WorkspaceInfo, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return mid.WorkspaceInfo{}, uuid.Nil, err
	}
	userID, err := mid.GetUserID(ctx)
	return workspace, userID, err
}

func canMutateDocuments(workspace mid.WorkspaceInfo) bool {
	role := mid.Role(workspace.UserRole)
	return role == mid.RoleMember || role == mid.RoleAdmin
}

func mutationContext(ctx context.Context, r *http.Request) (mid.WorkspaceInfo, uuid.UUID, uuid.UUID, error) {
	workspace, userID, err := requestContext(ctx)
	if err != nil {
		return mid.WorkspaceInfo{}, uuid.Nil, uuid.Nil, err
	}
	documentID, err := documentIDFromRequest(r)
	return workspace, userID, documentID, err
}

func documentIDFromRequest(r *http.Request) (uuid.UUID, error) {
	documentID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil || documentID == uuid.Nil {
		return uuid.Nil, documents.ErrInvalidInput
	}
	return documentID, nil
}

func attachmentIDFromRequest(r *http.Request) (uuid.UUID, error) {
	attachmentID, err := uuid.Parse(web.Params(r, "attachmentId"))
	if err != nil || attachmentID == uuid.Nil {
		return uuid.Nil, documents.ErrInvalidInput
	}
	return attachmentID, nil
}

func documentMediaURL(workspaceSlug string, documentID, attachmentID uuid.UUID) string {
	return fmt.Sprintf(
		"/workspaces/%s/documents/%s/media/%s",
		url.PathEscape(workspaceSlug),
		documentID,
		attachmentID,
	)
}

func documentHTTPStatus(err error) int {
	switch {
	case errors.Is(err, documents.ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, documents.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, documents.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func attachmentHTTPStatus(err error) int {
	switch {
	case errors.Is(err, attachments.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, attachments.ErrInvalidFileType), errors.Is(err, attachments.ErrInvalidFile):
		return http.StatusBadRequest
	case errors.Is(err, attachments.ErrUnauthorized):
		return http.StatusForbidden
	case errors.Is(err, attachments.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
