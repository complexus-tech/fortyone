package agentinvocationshttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	agentruns "github.com/complexus-tech/projects-api/internal/modules/agentinvocations/service"
	"github.com/complexus-tech/projects-api/internal/platform/auth"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

type StoryAccess interface {
	CanRead(context.Context, uuid.UUID, uuid.UUID) error
	Enabled(context.Context, uuid.UUID) (bool, error)
}
type Wakeup interface{ EnqueueAgentInvocation(uuid.UUID) error }

type Config struct {
	Log               *logger.Logger
	SecretKey         string
	BrowserSessions   mid.SessionResolver
	WorkspaceResolver mid.WorkspaceResolver
	Runs              *agentruns.Service
	Access            StoryAccess
	Tasks             Wakeup
}

type runResponse struct {
	ID             uuid.UUID        `json:"id"`
	SourceID       uuid.UUID        `json:"sourceId"`
	RequesterID    uuid.UUID        `json:"requesterId"`
	Status         agentruns.Status `json:"status"`
	Summary        string           `json:"summary"`
	ProposalReview string           `json:"proposalReview,omitempty"`
	ReplyCommentID *uuid.UUID       `json:"replyCommentId"`
	Version        int64            `json:"version"`
	UpdatedAt      time.Time        `json:"updatedAt"`
}

func present(run agentruns.Run) runResponse {
	response := runResponse{ID: run.ID, SourceID: run.SourceID, RequesterID: run.RequesterID, Status: run.Status,
		Summary: run.Summary, ReplyCommentID: publishedReplyID(run), Version: run.Version, UpdatedAt: run.UpdatedAt}
	if run.Status == agentruns.StatusAwaitingApproval && run.Proposal != nil {
		response.ProposalReview = run.Proposal.Review
	}
	return response
}

func publishedReplyID(run agentruns.Run) *uuid.UUID {
	if run.ReplyCommentID != nil {
		return run.ReplyCommentID
	}
	active := run.Status == agentruns.StatusQueued || run.Status == agentruns.StatusRunning
	if run.FailureReplyID != nil && (run.Status == agentruns.StatusFailed || (active && run.FailurePending)) {
		return run.FailureReplyID
	}
	// Publication commits before notification delivery. Expose that reply while
	// delivery retries, but do not reuse an old question after approval or retry.
	if run.ProposalReplyID != nil && (run.Status == agentruns.StatusAwaitingApproval ||
		(active && run.ApprovedAt == nil && !run.FailurePending && run.Proposal != nil && run.Proposal.RequiresApproval)) {
		return run.ProposalReplyID
	}
	return nil
}

func Routes(cfg Config, app *web.App) {
	h := &handler{cfg: cfg}
	authenticate := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/maya-runs", h.list, authenticate, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/maya-runs/{runId}/{operation}", h.control, authenticate, workspace, member)
}

type handler struct{ cfg Config }

func (h *handler) scope(ctx context.Context, r *http.Request) (uuid.UUID, uuid.UUID, error) {
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return uuid.Nil, uuid.Nil, err
	}
	storyID, err := uuid.Parse(web.Params(r, "id"))
	if err != nil {
		return uuid.Nil, uuid.Nil, agentruns.ErrInvalidInput
	}
	if h.cfg.Access == nil || h.cfg.Runs == nil {
		return uuid.Nil, uuid.Nil, agentruns.ErrUnavailable
	}
	if err := h.cfg.Access.CanRead(ctx, storyID, workspace.ID); err != nil {
		return uuid.Nil, uuid.Nil, agentruns.ErrNotFound
	}
	return workspace.ID, storyID, nil
}

func (h *handler) list(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, storyID, err := h.scope(ctx, r)
	if err != nil {
		return respondError(ctx, w, err)
	}
	enabled, err := h.cfg.Access.Enabled(ctx, workspaceID)
	if err != nil {
		return respondError(ctx, w, err)
	}
	runs, err := h.cfg.Runs.List(ctx, workspaceID, storyID)
	if err != nil {
		return respondError(ctx, w, err)
	}
	result := make([]runResponse, 0, len(runs))
	for _, run := range runs {
		if run.Surface == "story_comment" {
			result = append(result, present(run))
		}
	}
	return web.Respond(ctx, w, struct {
		Enabled bool          `json:"enabled"`
		Runs    []runResponse `json:"runs"`
	}{enabled, result}, http.StatusOK)
}

func (h *handler) control(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	workspaceID, storyID, err := h.scope(ctx, r)
	if err != nil {
		return respondError(ctx, w, err)
	}
	actor, err := auth.GetActor(ctx)
	if err != nil || actor.Kind != auth.PrincipalHumanUser {
		return respondError(ctx, w, agentruns.ErrPermissionDenied)
	}
	id, err := uuid.Parse(web.Params(r, "runId"))
	if err != nil {
		return respondError(ctx, w, agentruns.ErrInvalidInput)
	}
	run, err := h.cfg.Runs.Get(ctx, workspaceID, id)
	if err != nil {
		return respondError(ctx, w, err)
	}
	if run.EntityID != storyID || run.Surface != "story_comment" {
		return respondError(ctx, w, agentruns.ErrNotFound)
	}
	var input struct {
		Version int64 `json:"version"`
	}
	if err := web.Decode(r, &input); err != nil {
		return respondError(ctx, w, agentruns.ErrInvalidInput)
	}
	control := agentruns.Control{WorkspaceID: workspaceID, RequesterID: actor.PrincipalID, RunID: id, ExpectedVersion: input.Version}
	switch web.Params(r, "operation") {
	case "cancel":
		run, err = h.cfg.Runs.Cancel(ctx, control)
	case "retry":
		run, err = h.cfg.Runs.Retry(ctx, control)
	case "approve":
		run, err = h.cfg.Runs.Approve(ctx, control)
	default:
		err = agentruns.ErrInvalidInput
	}
	if err != nil {
		return respondError(ctx, w, err)
	}
	if run.Status == agentruns.StatusQueued && h.cfg.Tasks != nil {
		if err := h.cfg.Tasks.EnqueueAgentInvocation(run.ID); err != nil && h.cfg.Log != nil {
			h.cfg.Log.Error(ctx, "Maya invocation will be recovered by worker", "run_id", run.ID, "error", err)
		}
	}
	return web.Respond(ctx, w, present(run), http.StatusOK)
}

func respondError(ctx context.Context, w http.ResponseWriter, err error) error {
	status := http.StatusInternalServerError
	safe := errors.New("Maya request could not be loaded. Try again shortly.")
	switch {
	case errors.Is(err, agentruns.ErrInvalidInput):
		status = http.StatusBadRequest
		safe = agentruns.ErrInvalidInput
	case errors.Is(err, agentruns.ErrNotFound):
		status = http.StatusNotFound
		safe = agentruns.ErrNotFound
	case errors.Is(err, agentruns.ErrPermissionDenied):
		status = http.StatusForbidden
		safe = agentruns.ErrPermissionDenied
	case errors.Is(err, agentruns.ErrConflict):
		status = http.StatusConflict
		safe = errors.New("This request has changed or its action has already been applied. Refresh its status.")
	}
	return web.RespondError(ctx, w, safe, status)
}
