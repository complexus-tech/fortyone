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
	DocumentAccess    StoryAccess
	ChatAccess        StoryAccess
	Tasks             Wakeup
}

type runResponse struct {
	ID             uuid.UUID          `json:"id"`
	SourceID       uuid.UUID          `json:"sourceId"`
	ThreadRootID   *uuid.UUID         `json:"threadRootId,omitempty"`
	RequesterID    uuid.UUID          `json:"requesterId"`
	Status         agentruns.Status   `json:"status"`
	Summary        string             `json:"summary"`
	ProposalReview string             `json:"proposalReview,omitempty"`
	ReplyCommentID *uuid.UUID         `json:"replyCommentId"`
	Version        int64              `json:"version"`
	UpdatedAt      time.Time          `json:"updatedAt"`
	ReplyMessageID *uuid.UUID         `json:"replyMessageId,omitempty"`
	Tasks          []chatTaskResponse `json:"tasks,omitempty"`
}

func present(run agentruns.Run) runResponse {
	response := runResponse{ID: run.ID, SourceID: run.SourceID, RequesterID: run.RequesterID, Status: run.Status,
		Summary: run.Summary, ReplyCommentID: publishedReplyID(run), Version: run.Version, UpdatedAt: run.UpdatedAt}
	if run.Status == agentruns.StatusAwaitingApproval && run.Proposal != nil {
		response.ProposalReview = run.Proposal.Review
	}
	if run.Surface == "chat_message" {
		response.ThreadRootID = &run.ConversationID
		response.ReplyMessageID = response.ReplyCommentID
		response.Tasks = presentChatTasks(run)
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

func presentRuns(runs []agentruns.Run, surface string) []runResponse {
	result := make([]runResponse, 0, len(runs))
	for _, run := range runs {
		if run.Surface == surface && (run.Proposal == nil || !run.Proposal.Ignore) {
			result = append(result, present(run))
		}
	}
	return result
}

func Routes(cfg Config, app *web.App) {
	h := &handler{cfg: cfg, surface: "story_comment"}
	authenticate := mid.Auth(cfg.Log, cfg.SecretKey, cfg.BrowserSessions)
	workspace := mid.Workspace(cfg.Log, cfg.WorkspaceResolver)
	member := mid.RequireMinimumRole(cfg.Log, mid.RoleMember)
	app.Get("/workspaces/{workspaceSlug}/stories/{id}/maya-runs", h.list, authenticate, workspace)
	app.Post("/workspaces/{workspaceSlug}/stories/{id}/maya-runs/{runId}/{operation}", h.control, authenticate, workspace, member)
	if cfg.DocumentAccess != nil {
		documentConfig := cfg
		documentConfig.Access = cfg.DocumentAccess
		documents := &handler{cfg: documentConfig, surface: "document_comment"}
		app.Get("/workspaces/{workspaceSlug}/documents/{id}/maya-runs", documents.list, authenticate, workspace)
		app.Post("/workspaces/{workspaceSlug}/documents/{id}/maya-runs/{runId}/{operation}", documents.control, authenticate, workspace, member)
	}
	if cfg.ChatAccess != nil {
		chatConfig := cfg
		chatConfig.Access = cfg.ChatAccess
		chat := &handler{cfg: chatConfig, surface: "chat_message"}
		app.Get("/workspaces/{workspaceSlug}/chat/conversations/{id}/maya-runs", chat.list, authenticate, workspace)
		app.Post("/workspaces/{workspaceSlug}/chat/conversations/{id}/maya-runs/{runId}/{operation}", chat.control, authenticate, workspace, member)
	}
}

type handler struct {
	cfg     Config
	surface string
}

func (h *handler) sourceSurface() string {
	if h.surface == "" {
		return "story_comment"
	}
	return h.surface
}

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
	var enabled bool
	if access, ok := h.cfg.Access.(interface {
		EnabledForEntity(context.Context, uuid.UUID, uuid.UUID) (bool, error)
	}); ok {
		enabled, err = access.EnabledForEntity(ctx, storyID, workspaceID)
	} else {
		enabled, err = h.cfg.Access.Enabled(ctx, workspaceID)
	}
	if err != nil {
		return respondError(ctx, w, err)
	}
	runs, err := h.cfg.Runs.List(ctx, workspaceID, storyID)
	if err != nil {
		return respondError(ctx, w, err)
	}
	return web.Respond(ctx, w, struct {
		Enabled bool          `json:"enabled"`
		Runs    []runResponse `json:"runs"`
	}{enabled, presentRuns(runs, h.sourceSurface())}, http.StatusOK)
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
	if run.EntityID != storyID || run.Surface != h.sourceSurface() {
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
