package handlers

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/complexus-tech/projects-api/internal/core/issues"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/jmoiron/sqlx"
)

var (
	ErrInvalidID = errors.New("ID is not in its proper form")
)

type IssuesHandler interface {
	Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error
	List(ctx context.Context, w http.ResponseWriter, r *http.Request) error
}

type issuesHandlers struct {
	issues issues.Service
	log    *logger.Logger
}

// NewIssuesHandlers returns a new issuesHandlers instance.
func NewIssuesHandlers(log *logger.Logger, db *sqlx.DB) IssuesHandler {
	i := issues.NewService(log, db)
	return &issuesHandlers{
		issues: i,
		log:    log,
	}
}

// Get returns the issue with the specified ID.
func (h *issuesHandlers) Get(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(web.Params(r, "id"))
	if err != nil {
		h.log.Error(ctx, "Error")
		return ErrInvalidID
	}
	issue, err := h.issues.Get(ctx, id)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, issue, http.StatusOK)
	return nil
}

// List returns a list of issues.
func (h *issuesHandlers) List(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	issues, err := h.issues.List(ctx)
	if err != nil {
		return err
	}
	web.Respond(ctx, w, issues, http.StatusOK)
	return nil
}
