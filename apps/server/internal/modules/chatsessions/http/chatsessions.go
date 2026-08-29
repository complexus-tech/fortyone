package chatsessionshttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	chatsessions "github.com/complexus-tech/projects-api/internal/modules/chatsessions/service"
	mid "github.com/complexus-tech/projects-api/internal/platform/http/middleware"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidSessionID     = errors.New("session id is not in its proper form")
	ErrInvalidWorkspaceID   = errors.New("workspace id is not in its proper form")
	ErrInvalidToolCallID    = errors.New("tool call id is not in its proper form")
	ErrMutationUnauthorized = errors.New("mutation approval authentication is required")
)

type Handlers struct {
	chatsessions *chatsessions.Service
	log          *logger.Logger
}

func New(chatsessions *chatsessions.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		chatsessions: chatsessions,
		log:          log,
	}
}

func (h *Handlers) CreateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.CreateSession")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppNewChatSession
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	ncs := chatsessions.CoreNewChatSession{
		ID:          req.ID,
		UserID:      userID,
		WorkspaceID: workspace.ID,
		Title:       req.Title,
		Messages:    req.Messages,
	}

	session, err := h.chatsessions.CreateSession(ctx, ncs)
	if err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppChatSession(session), http.StatusCreated)
	return nil
}

func (h *Handlers) GetSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.GetSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	session, err := h.chatsessions.GetSession(ctx, sessionID, userID, workspace.ID)
	if err != nil {
		if errors.Is(err, chatsessions.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppChatSession(session), http.StatusOK)
	return nil
}

func (h *Handlers) ListSessions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.ListSessions")
	defer span.End()

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	sessions, err := h.chatsessions.ListSessions(ctx, userID, workspace.ID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppChatSessions(sessions), http.StatusOK)
	return nil
}

func (h *Handlers) UpdateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.UpdateSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppUpdateSessionRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.chatsessions.UpdateSession(ctx, sessionID, userID, workspace.ID, req.Title); err != nil {
		if errors.Is(err, chatsessions.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

func (h *Handlers) DeleteSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.DeleteSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	if err := h.chatsessions.DeleteSession(ctx, sessionID, userID, workspace.ID); err != nil {
		if errors.Is(err, chatsessions.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusNoContent)
	return nil
}

func (h *Handlers) SaveMessages(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.SaveMessages")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppSaveMessagesRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	if err := h.chatsessions.SaveMessages(ctx, sessionID, userID, workspace.ID, req.Messages); err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

func (h *Handlers) BeginMessageWrite(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.BeginMessageWrite")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppBeginMessageWriteRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	reservation, err := h.chatsessions.BeginMessageWrite(ctx, chatsessions.BeginMessageWriteParams{
		Session: chatsessions.CoreChatSession{
			ID:          sessionID,
			UserID:      userID,
			WorkspaceID: workspace.ID,
			Title:       req.Title,
		},
		Messages:        req.Messages,
		Operation:       req.Operation,
		TargetMessageID: req.MessageID,
	})
	if err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMessageWriteReservation(reservation), http.StatusOK)
	return nil
}

func (h *Handlers) FinalizeMessageWrite(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.FinalizeMessageWrite")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	var req AppFinalizeMessageWriteRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	result, err := h.chatsessions.FinalizeMessageWrite(ctx, chatsessions.FinalizeMessageWriteParams{
		SessionID:   sessionID,
		UserID:      userID,
		WorkspaceID: workspace.ID,
		Messages:    req.Messages,
		Generation:  req.Generation,
		Token:       req.Token,
	})
	if err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMessageWriteResult(result), http.StatusOK)
	return nil
}

func (h *Handlers) RecoverMutationApprovalOutput(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.RecoverMutationApprovalOutput")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}
	var req AppRecoverMutationApprovalOutputRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	result, err := h.chatsessions.RecoverMutationApprovalOutput(ctx, chatsessions.RecoverMutationApprovalOutputParams{
		SessionID:   params.SessionID,
		UserID:      params.UserID,
		WorkspaceID: params.WorkspaceID,
		ToolCallID:  params.ToolCallID,
		Fingerprint: req.Fingerprint,
	})
	if err != nil {
		web.RespondError(ctx, w, err, messageWriteRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMessageWriteResult(result), http.StatusOK)
	return nil
}

func (h *Handlers) GetMessages(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.GetMessages")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	messages, err := h.chatsessions.GetMessages(ctx, sessionID, userID, workspace.ID)
	if err != nil {
		if errors.Is(err, chatsessions.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, messages, http.StatusOK)
	return nil
}

func (h *Handlers) GetLatestAssistantMessage(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.GetLatestAssistantMessage")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}
	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusUnauthorized)
		return nil
	}

	message, err := h.chatsessions.GetLatestAssistantMessage(ctx, sessionID, userID, workspace.ID)
	if err != nil {
		if errors.Is(err, chatsessions.ErrNotFound) {
			web.RespondError(ctx, w, err, http.StatusNotFound)
			return nil
		}
		web.RespondError(ctx, w, err, http.StatusInternalServerError)
		return nil
	}

	web.Respond(ctx, w, message, http.StatusOK)
	return nil
}

func (h *Handlers) ClaimMutationApproval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.ClaimMutationApproval")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	var req AppClaimMutationApprovalRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	params.Fingerprint = req.Fingerprint

	execution, err := h.chatsessions.ClaimMutationApproval(ctx, params)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMutationApprovalExecution(execution), http.StatusOK)
	return nil
}

func (h *Handlers) CompleteMutationApproval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.CompleteMutationApproval")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	var req AppCompleteMutationApprovalRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	params.Fingerprint = req.Fingerprint
	params.LeaseToken = req.LeaseToken

	execution, err := h.chatsessions.CompleteMutationApproval(ctx, params, req.Output)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMutationApprovalExecution(execution), http.StatusOK)
	return nil
}

func (h *Handlers) StartMutationApproval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.StartMutationApproval")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	var req AppStartMutationApprovalRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	params.Fingerprint = req.Fingerprint
	params.LeaseToken = req.LeaseToken

	execution, err := h.chatsessions.StartMutationApproval(ctx, params)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMutationApprovalExecution(execution), http.StatusOK)
	return nil
}

func (h *Handlers) FailMutationApproval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.FailMutationApproval")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	var req AppFailMutationApprovalRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	params.Fingerprint = req.Fingerprint
	params.LeaseToken = req.LeaseToken

	execution, err := h.chatsessions.FailMutationApproval(ctx, params, req.FailureCode)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMutationApprovalExecution(execution), http.StatusOK)
	return nil
}

func (h *Handlers) ReconcileMutationApproval(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionshttp.handlers.ReconcileMutationApproval")
	defer span.End()

	params, err := mutationApprovalExecutionParams(ctx, r)
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	var req AppReconcileMutationApprovalRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}
	params.Fingerprint = req.Fingerprint

	execution, err := h.chatsessions.ReconcileMutationApproval(ctx, params, chatsessions.MutationApprovalReconciliation{
		Resolution: req.Resolution,
		Evidence:   req.Evidence,
		Output:     req.Output,
	})
	if err != nil {
		web.RespondError(ctx, w, err, mutationApprovalRequestStatus(err))
		return nil
	}

	web.Respond(ctx, w, toAppMutationApprovalExecution(execution), http.StatusOK)
	return nil
}

func mutationApprovalExecutionParams(ctx context.Context, r *http.Request) (chatsessions.MutationApprovalExecutionParams, error) {
	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		return chatsessions.MutationApprovalExecutionParams{}, ErrInvalidSessionID
	}

	toolCallID := web.Params(r, "toolCallId")
	if len(toolCallID) == 0 || len(toolCallID) > 255 {
		return chatsessions.MutationApprovalExecutionParams{}, ErrInvalidToolCallID
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		return chatsessions.MutationApprovalExecutionParams{}, ErrInvalidWorkspaceID
	}
	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return chatsessions.MutationApprovalExecutionParams{}, fmt.Errorf("%w: %v", ErrMutationUnauthorized, err)
	}

	return chatsessions.MutationApprovalExecutionParams{
		SessionID:   sessionID,
		ToolCallID:  toolCallID,
		UserID:      userID,
		WorkspaceID: workspace.ID,
	}, nil
}

func mutationApprovalRequestStatus(err error) int {
	switch {
	case errors.Is(err, chatsessions.ErrMutationApprovalConflict),
		errors.Is(err, chatsessions.ErrMutationApprovalUncertain),
		errors.Is(err, chatsessions.ErrMutationApprovalLease):
		return http.StatusConflict
	case errors.Is(err, chatsessions.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrMutationUnauthorized):
		return http.StatusUnauthorized
	case errors.Is(err, ErrInvalidSessionID), errors.Is(err, ErrInvalidWorkspaceID), errors.Is(err, ErrInvalidToolCallID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func messageWriteRequestStatus(err error) int {
	switch {
	case errors.Is(err, chatsessions.ErrMessageWriteConflict),
		errors.Is(err, chatsessions.ErrMessageWriteApprovalOpen):
		return http.StatusConflict
	case errors.Is(err, chatsessions.ErrMessageWriteInvalid):
		return http.StatusUnprocessableEntity
	case errors.Is(err, chatsessions.ErrNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}

func (h *Handlers) GetUserMessageCount(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "handlers.chatsessions.GetUserMessageCount")
	defer span.End()

	userID, err := mid.GetUserID(ctx)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}

	workspace, err := mid.GetWorkspace(ctx)
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	count, err := h.chatsessions.CountUserMessagesCurrentMonth(ctx, userID, workspace.ID)
	if err != nil {
		return web.RespondError(ctx, w, fmt.Errorf("failed to count user messages: %w", err), http.StatusInternalServerError)
	}

	return web.Respond(ctx, w, GetUserMessageCountResponse{Count: count}, http.StatusOK)
}
