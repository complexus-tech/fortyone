package attachmentsgrp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// Handler manages attachment-related endpoints
type Handler struct {
	log               *logger.Logger
	attachmentService *attachments.Service
	validate          *validator.Validate
}

// New creates a new attachments handler
func New(attachmentService *attachments.Service, log *logger.Logger, validate *validator.Validate) *Handler {
	return &Handler{
		log:               log,
		attachmentService: attachmentService,
		validate:          validate,
	}
}

// Upload handles the upload of a new attachment
func (h *Handler) Upload(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// Get team and workspace IDs from query params
	var teamID *uuid.UUID
	teamIDStr := r.URL.Query().Get("teamId")
	if teamIDStr != "" {
		id, err := uuid.Parse(teamIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid team ID"), http.StatusBadRequest)
		}
		teamID = &id
	}

	var workspaceID *uuid.UUID
	workspaceIDStr := r.URL.Query().Get("workspaceId")
	if workspaceIDStr != "" {
		id, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid workspace ID"), http.StatusBadRequest)
		}
		workspaceID = &id
	}

	// Parse multipart form, limit to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	// Upload file
	fileInfo, err := h.attachmentService.UploadAttachment(ctx, file, header, userID, teamID, workspaceID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrFileTooLarge), errors.Is(err, attachments.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, NewAttachmentResponse(fileInfo), http.StatusCreated)
}

// Get handles getting an attachment by ID
func (h *Handler) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get attachment ID from URL
	attachmentIDStr := web.Params(r, "id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	// Get attachment
	fileInfo, err := h.attachmentService.GetAttachment(ctx, attachmentID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return fmt.Errorf("error getting attachment: %w", err)
	}

	return web.Respond(ctx, w, NewAttachmentResponse(fileInfo), http.StatusOK)
}

// GetForStory handles getting all attachments for a story
func (h *Handler) GetForStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get story ID from URL
	storyIDStr := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid story ID"), http.StatusBadRequest)
	}

	// Get attachments for story
	fileInfos, err := h.attachmentService.GetAttachmentsForStory(ctx, storyID)
	if err != nil {
		return fmt.Errorf("error getting attachments for story: %w", err)
	}

	return web.Respond(ctx, w, NewAttachmentsResponse(fileInfos), http.StatusOK)
}

// Delete handles deleting an attachment
func (h *Handler) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// Get attachment ID from URL
	attachmentIDStr := web.Params(r, "id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	// Delete attachment
	err = h.attachmentService.DeleteAttachment(ctx, attachmentID, userID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrNotFound):
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		case errors.Is(err, attachments.ErrUnauthorized):
			return web.RespondError(ctx, w, err, http.StatusForbidden)
		default:
			return fmt.Errorf("error deleting attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// LinkToStory handles linking an attachment to a story
func (h *Handler) LinkToStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get story ID from URL
	storyIDStr := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid story ID"), http.StatusBadRequest)
	}

	// Parse request body
	var req LinkAttachmentRequest
	if err := web.Decode(r, &req); err != nil {
		return fmt.Errorf("error decoding request: %w", err)
	}

	// Validate request
	if err := h.validate.Struct(req); err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("invalid request: %w", err), http.StatusBadRequest)
	}

	// Link attachment to story
	err = h.attachmentService.LinkAttachmentToStory(ctx, storyID, req.AttachmentID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return fmt.Errorf("error linking attachment to story: %w", err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// UnlinkFromStory handles unlinking an attachment from a story
func (h *Handler) UnlinkFromStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get story ID and attachment ID from URL
	storyIDStr := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid story ID"), http.StatusBadRequest)
	}

	attachmentIDStr := web.Params(r, "attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	// Unlink attachment from story
	err = h.attachmentService.UnlinkAttachmentFromStory(ctx, storyID, attachmentID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return fmt.Errorf("error unlinking attachment from story: %w", err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// UploadStoryAttachment handles the upload of a new attachment specifically for a story
func (h *Handler) UploadStoryAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// Get story ID from URL
	storyIDStr := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid story ID"), http.StatusBadRequest)
	}

	// Parse multipart form, limit to 10MB
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	// Get team and workspace IDs from query params
	var teamID *uuid.UUID
	teamIDStr := r.URL.Query().Get("teamId")
	if teamIDStr != "" {
		id, err := uuid.Parse(teamIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid team ID"), http.StatusBadRequest)
		}
		teamID = &id
	}

	var workspaceID *uuid.UUID
	workspaceIDStr := r.URL.Query().Get("workspaceId")
	if workspaceIDStr != "" {
		id, err := uuid.Parse(workspaceIDStr)
		if err != nil {
			return web.RespondError(ctx, w, errors.New("invalid workspace ID"), http.StatusBadRequest)
		}
		workspaceID = &id
	}

	// Upload file and automatically link to story
	fileInfo, err := h.attachmentService.UploadAndLinkToStory(ctx, file, header, userID, storyID, teamID, workspaceID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrFileTooLarge), errors.Is(err, attachments.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, NewAttachmentResponse(fileInfo), http.StatusCreated)
}

// UploadProfileImage handles the upload of a user profile image
func (h *Handler) UploadProfileImage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get user ID from context
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// Parse multipart form, limit to 2MB
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	// Validate file is an image
	if err := validate.ProfileImage(file, header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Reset file position after validation
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Upload profile image
	url, err := h.attachmentService.UploadProfileImage(
		ctx,
		file,
		header.Filename,
		header.Size,
		header.Header.Get("Content-Type"),
		userID,
	)
	if err != nil {
		return fmt.Errorf("error uploading profile image: %w", err)
	}

	return web.Respond(ctx, w, map[string]string{"url": url}, http.StatusCreated)
}

// UploadWorkspaceLogo handles the upload of a workspace logo
func (h *Handler) UploadWorkspaceLogo(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	// Get user ID from context - needed for permission checking
	_, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("unauthorized"), http.StatusUnauthorized)
	}

	// Get workspace ID from URL
	workspaceIDStr := web.Params(r, "id")
	workspaceID, err := uuid.Parse(workspaceIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid workspace ID"), http.StatusBadRequest)
	}

	// Parse multipart form, limit to 2MB
	if err := r.ParseMultipartForm(2 << 20); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Get file from form
	file, header, err := r.FormFile("file")
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("error getting file: %w", err), http.StatusBadRequest)
	}
	defer file.Close()

	// Validate file is an image
	if err := validate.WorkspaceLogo(file, header); err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}

	// Reset file position after validation
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return web.RespondError(ctx, w, err, http.StatusInternalServerError)
	}

	// Upload workspace logo - note: userID is passed to service for permission checking
	url, err := h.attachmentService.UploadWorkspaceLogo(
		ctx,
		file,
		header.Filename,
		header.Size,
		header.Header.Get("Content-Type"),
		workspaceID,
	)
	if err != nil {
		return fmt.Errorf("error uploading workspace logo: %w", err)
	}

	return web.Respond(ctx, w, map[string]string{"url": url}, http.StatusCreated)
}
