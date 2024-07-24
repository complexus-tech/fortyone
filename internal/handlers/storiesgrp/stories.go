package storiesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

type Handlers struct {
	stories *stories.Service
	// audit  *audit.Service
}

// NewStoriesHandlers returns a new storiesHandlers instance.
func New(stories *stories.Service) *Handlers {
	return &Handlers{
		stories: stories,
	}
}

// Get returns the story with the specified ID.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "id")
	id, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	story, err := h.stories.Get(ctx, id)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusOK)
	return nil
}

// Delete removes the story with the specified ID.
func (h *Handlers) Delete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "id")
	id, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	if err := h.stories.Delete(ctx, id); err != nil {
		return err
	}
	data := map[string]uuid.UUID{"id": id}
	web.Respond(ctx, w, data, http.StatusNoContent)
	return nil
}

// BulkDelete removes the stories with the specified IDs.
func (h *Handlers) BulkDelete(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req AppBulkDeleteRequest
	if err := web.Decode(r, &req); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkDelete(ctx, req.StoryIDs); err != nil {
		return err
	}
	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusNoContent)
	return nil

}

// Restore restores the story with the specified ID.
func (h *Handlers) Restore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "id")
	id, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	if err := h.stories.Restore(ctx, id); err != nil {
		return err
	}
	data := map[string]uuid.UUID{"id": id}
	web.Respond(ctx, w, data, http.StatusNoContent)
	return nil
}

// BulkRestore restores the stories with the specified IDs.
func (h *Handlers) BulkRestore(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var req AppBulkRestoreRequest
	if err := web.Decode(r, &req); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}
	if err := h.stories.BulkRestore(ctx, req.StoryIDs); err != nil {
		return err
	}
	data := map[string][]uuid.UUID{"storyIds": req.StoryIDs}
	web.Respond(ctx, w, data, http.StatusNoContent)
	return nil
}

// Create creates a new story.
func (h *Handlers) Create(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	var ns AppNewStory
	if err := web.Decode(r, &ns); err != nil {
		web.Respond(ctx, w, err.Error(), http.StatusBadRequest)
		return nil
	}

	story, err := h.stories.Create(ctx, toCoreNewStory(ns))
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusCreated)
	return nil
}

// MyStories returns a list of stories.
func (h *Handlers) MyStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	stories, err := h.stories.MyStories(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// TeamStories returns a list of stories for a team.
func (h *Handlers) TeamStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "teamId")
	teamId, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	stories, err := h.stories.TeamStories(ctx, teamId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// ObjectiveStories returns a list of stories for an objective.
func (h *Handlers) ObjectiveStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "objectiveId")
	objectiveId, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	stories, err := h.stories.ObjectiveStories(ctx, objectiveId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// EpicStories returns a list of stories for an epic.
func (h *Handlers) EpicStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "epicId")
	epicId, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}

	stories, err := h.stories.EpicStories(ctx, epicId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}

// SprintStories returns a list of stories for a sprint.
func (h *Handlers) SprintStories(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	param := web.Params(r, "sprintId")
	sprintId, err := uuid.Parse(param)
	if err != nil {
		return ErrInvalidID
	}
	stories, err := h.stories.SprintStories(ctx, sprintId)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStories(stories), http.StatusOK)
	return nil
}
