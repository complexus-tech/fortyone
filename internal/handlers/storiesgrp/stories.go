package storiesgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/stories"
	"github.com/complexus-tech/projects-api/pkg/web"
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
	id := web.Params(r, "id")
	// if err != nil {
	// 	return ErrInvalidID
	// }
	story, err := h.stories.Get(ctx, id)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, toAppStory(story), http.StatusOK)
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
