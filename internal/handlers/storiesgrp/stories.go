package storiesgrp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidStoryID     = errors.New("story id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	stories *stories.Service
	log     *logger.Logger
	// audit  *audit.Service
}

// NewStoriesHandlers returns a new storiesHandlers instance.
func New(stories *stories.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		stories: stories,
	}
}

// Get returns the story with the specified ID, including sub-stories.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	storyIdParam := web.Params(r, "id")
	workspaceIdParam := web.Params(r, "workspaceId")
	storyId, err := uuid.Parse(storyIdParam)
	if err != nil {
		return ErrInvalidStoryID
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	story, err := h.stories.Get(ctx, storyId, workspaceId)
	if err != nil {
		return err
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
		return ErrInvalidStoryID
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	if err := h.stories.Delete(ctx, storyId, workspaceId); err != nil {
		return err
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// BulkDelete removes the stories with the specified IDs.
func (h *Handlers) BulkDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}
	var req AppBulkDeleteRequest
	if err := web.Decode(r, &req); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkDelete(ctx, req.StoryIDs, workspaceId); err != nil {
		return err
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
		return ErrInvalidStoryID
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	if err := h.stories.Restore(ctx, storyId, workspaceId); err != nil {
		return err
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
		return ErrInvalidWorkspaceID
	}
	var req AppBulkRestoreRequest
	if err := web.Decode(r, &req); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkRestore(ctx, req.StoryIDs, workspaceId); err != nil {
		return err
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
		return ErrInvalidWorkspaceID
	}
	var ns AppNewStory
	if err := web.Decode(r, &ns); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	story, err := h.stories.Create(ctx, toCoreNewStory(ns), workspaceId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusCreated)
	return nil
}

// MyStories returns a list of stories.
func (h *Handlers) MyStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}
	stories, err := h.stories.MyStories(ctx, workspaceId)
	if err != nil {
		return err
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
		return ErrInvalidStoryID
	}

	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	var requestData map[string]json.RawMessage
	if err := web.Decode(r, &requestData); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	// Get the user ID from the context (assuming it's set by authentication middleware)
	userID, ok := ctx.Value("userID").(uuid.UUID)
	if !ok {
		return errors.New("user ID not found in context")
	}

	updates, err := getUpdates(requestData)
	if err != nil {
		return err
	}
	if err := h.stories.Update(ctx, storyId, workspaceId, updates, userID); err != nil {
		return err
	}
	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

// List returns a list of stories for a team.
func (h *Handlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceIdParam := web.Params(r, "workspaceId")
	workspaceId, err := uuid.Parse(workspaceIdParam)
	if err != nil {
		return ErrInvalidWorkspaceID
	}

	var af AppFilters
	filters, err := web.GetFilters(r.URL.Query(), &af)
	if err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	stories, err := h.stories.List(ctx, workspaceId, filters)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}
