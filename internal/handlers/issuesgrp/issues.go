package issuesgrp

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

type Handlers struct {
	issues *issues.Service
	// audit  *audit.Service
}

// NewIssuesHandlers returns a new issuesHandlers instance.
func New(issues *issues.Service) *Handlers {
	return &Handlers{
		issues: issues,
	}
}

// Get returns the issue with the specified ID.
func (h *Handlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(web.Params(r, "id"))
	if err != nil {
		return ErrInvalidID
	}
	issue, err := h.issues.Get(ctx, id)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, issue, http.StatusOK)
	return nil
}

// MyIssues returns a list of issues.
func (h *Handlers) MyIssues(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	issues, err := h.issues.MyIssues(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, issues, http.StatusOK)
	return nil
}
