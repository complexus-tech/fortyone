package chatsessionsgrp

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/chatsessions"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
)

var (
	ErrInvalidSessionID   = errors.New("session id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
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
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.CreateSession")
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
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppChatSession(session), http.StatusCreated)
	return nil
}

func (h *Handlers) GetSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.GetSession")
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

	session, err := h.chatsessions.GetSession(ctx, sessionID, workspace.ID)
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
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.ListSessions")
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
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.UpdateSession")
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

	var req AppUpdateSessionRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.chatsessions.UpdateSession(ctx, sessionID, workspace.ID, req.Title); err != nil {
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
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.DeleteSession")
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

	if err := h.chatsessions.DeleteSession(ctx, sessionID, workspace.ID); err != nil {
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
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.SaveMessages")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	var req AppSaveMessagesRequest
	if err := web.Decode(r, &req); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	if err := h.chatsessions.SaveMessages(ctx, sessionID, req.Messages); err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, nil, http.StatusOK)
	return nil
}

func (h *Handlers) GetMessages(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.GetMessages")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	messages, err := h.chatsessions.GetMessages(ctx, sessionID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, messages, http.StatusOK)
	return nil
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
