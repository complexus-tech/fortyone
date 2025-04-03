package storiesgrp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/attachments"
	"github.com/complexus-tech/projects-api/internal/core/comments"
	"github.com/complexus-tech/projects-api/internal/core/links"
	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/internal/handlers/linksgrp"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidStoryID     = errors.New("story id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	stories     *stories.Service
	comments    *comments.Service
	links       *links.Service
	attachments *attachments.Service
	log         *logger.Logger
}

// NewStoriesHandlers returns a new storiesHandlers instance.
func New(stories *stories.Service, comments *comments.Service, links *links.Service, attachments *attachments.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		stories:     stories,
		comments:    comments,
		links:       links,
		attachments: attachments,
		log:         log,
	}
}

// Get returns the story with the specified ID, including sub-stories.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		h.log.Error(ctx, "invalid story id", "error", err)
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	story, err := h.stories.Get(ctx, storyId, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusOK)
	return nil
}

// Delete removes the story with the specified ID.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.Delete(ctx, storyId, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// BulkDelete removes the stories with the specified IDs.
func (h *Handlers) BulkDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var req AppBulkDeleteRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkDelete(ctx, req.StoryIDs, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil

}

// Restore restores the story with the specified ID.
func (h *Handlers) Restore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.Restore(ctx, storyId, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	data := map[string]uuid.UUID{"id": storyId}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (h *Handlers) BulkRestore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var req AppBulkRestoreRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkRestore(ctx, req.StoryIDs, workspaceId); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusOK)
	return nil
}

// Create creates a new story.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	var ns AppNewStory
	if err := web.Decode(r, &ns); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	story, err := h.stories.Create(ctx, toCoreNewStory(ns), workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusCreated)
	return nil
}

// UpdateLabels replaces the labels for a story.
func (h *Handlers) UpdateLabels(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var req AppNewLabels
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.stories.UpdateLabels(ctx, storyId, workspaceId, req.Labels); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// GetStoryLinks returns the links for a story.
func (h *Handlers) GetStoryLinks(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyID, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	links, err := h.stories.GetStoryLinks(ctx, storyID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	return web.Respond(ctx, w, linksgrp.ToLinks(links), http.StatusOK)
}

// MyStories returns a list of stories.
func (h *Handlers) MyStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	stories, err := h.stories.MyStories(ctx, workspaceId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// Update updates the story with the specified ID.
func (h *Handlers) Update(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var requestData map[string]json.RawMessage
	if err := web.Decode(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	updates, err := getUpdates(requestData)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	if err := h.stories.Update(ctx, storyId, workspaceId, updates); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// List returns a list of stories for a team.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	stories, err := h.stories.List(ctx, workspaceId, filters)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// GetActivities returns the activities for a story.
func (h *Handlers) GetActivities(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	activities, err := h.stories.GetActivities(ctx, storyId)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppActivities(activities), http.StatusOK)
	return nil
}

// CreateComment creates a comment for a story.
func (h *Handlers) CreateComment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	var requestData AppNewComment
	if err := web.Decode(r, &requestData); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	ca := stories.CoreNewComment{
		StoryID: storyId,
		Parent:  requestData.Parent,
		UserID:  userID,
		Comment: requestData.Comment,
	}

	comment, err := h.stories.CreateComment(ctx, ca)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppComment(comment), http.StatusCreated)

	return nil
}

// GetComments returns the comments for a story.
func (h *Handlers) GetComments(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		h.log.Error(ctx, "invalid story id", "error", err)
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	comments, err := h.stories.GetComments(ctx, storyId)
	if err != nil {
		h.log.Error(ctx, "failed to get comments", "error", err)
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	web.Respond(ctx, w, toAppComments(comments), http.StatusOK)
	return nil

}

// DuplicateStory creates a copy of an existing story.
func (h *Handlers) DuplicateStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	// Get the user ID from context
	userID, _ := mid.GetUserID(ctx)

	duplicatedStory, err := h.stories.DuplicateStory(ctx, storyId, workspaceId, userID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppStory(duplicatedStory), http.StatusCreated)
	return nil
}

// GetAttachmentsForStory returns the attachments for a story.
func (h *Handlers) GetAttachmentsForStory(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	fileInfos, err := h.attachments.GetAttachmentsForStory(ctx, storyId)
	if err != nil {
		return fmt.Errorf("error getting attachments for story: %w", err)
	}

	return web.Respond(ctx, w, attachments.ToAppAttachments(fileInfos), http.StatusOK)
}

// DeleteAttachment deletes an attachment from a story.
func (h *Handlers) DeleteAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	userID, _ := mid.GetUserID(ctx)

	attachmentIDStr := web.Params(r, "attachmentId")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return web.RespondError(ctx, w, errors.New("invalid attachment ID"), http.StatusBadRequest)
	}

	// delete attachment from Azure
	err = h.attachments.DeleteAttachment(ctx, attachmentID, userID)
	if err != nil {
		return fmt.Errorf("error deleting attachment: %w", err)
	}

	// Unlink attachment from story
	err = h.attachments.UnlinkAttachmentFromStory(ctx, storyId, attachmentID)
	if err != nil {
		if errors.Is(err, attachments.ErrNotFound) {
			return web.RespondError(ctx, w, err, http.StatusNotFound)
		}
		return fmt.Errorf("error unlinking attachment from story: %w", err)
	}

	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

// UploadStoryAttachment handles the upload of a new attachment specifically for a story
func (h *Handlers) UploadStoryAttachment(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	userID, _ := mid.GetUserID(ctx)
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")

	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidStoryID, http.StatusBadRequest)
		return nil
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
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

	// Upload file and automatically link to story
	fileInfo, err := h.attachments.UploadAndLinkToStory(ctx, file, header, userID, storyId, workspaceId)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrFileTooLarge), errors.Is(err, attachments.ErrInvalidFileType):
			return web.RespondError(ctx, w, err, http.StatusBadRequest)
		default:
			return fmt.Errorf("error uploading attachment: %w", err)
		}
	}

	return web.Respond(ctx, w, attachments.ToAppAttachment(fileInfo), http.StatusCreated)
}
