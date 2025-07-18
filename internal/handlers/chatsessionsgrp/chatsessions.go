package chatsessionsgrp

import (
	"context"
	"errors"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/core/chatsessions"
	"github.com/complexus-tech/projects-api/internal/web/mid"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

var (
	ErrInvalidSessionID   = errors.New("session id is not in its proper form")
	ErrInvalidWorkspaceID = errors.New("workspace id is not in its proper form")
)

type Handlers struct {
	chatsessions *chatsessions.Service
	log          *logger.Logger
}

// New returns a new chat sessions handlers instance.
func New(chatsessions *chatsessions.Service, log *logger.Logger) *Handlers {
	return &Handlers{
		chatsessions: chatsessions,
		log:          log,
	}
}

// CreateSession creates a new chat session with initial messages.
func (h *Handlers) CreateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.CreateSession")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
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
		WorkspaceID: workspaceID,
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

// GetSession returns the chat session with the specified ID.
func (h *Handlers) GetSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.GetSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	session, err := h.chatsessions.GetSession(ctx, sessionID, workspaceID)
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

// ListSessions returns a list of chat sessions for the current user in the workspace.
func (h *Handlers) ListSessions(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.ListSessions")
	defer span.End()

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
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

	sessions, err := h.chatsessions.ListSessions(ctx, userID, workspaceID)
	if err != nil {
		web.RespondError(ctx, w, err, http.StatusBadRequest)
		return nil
	}

	web.Respond(ctx, w, toAppChatSessions(sessions), http.StatusOK)
	return nil
}

// UpdateSession updates the title of a chat session.
func (h *Handlers) UpdateSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.UpdateSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
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

	if err := h.chatsessions.UpdateSession(ctx, sessionID, workspaceID, req.Title); err != nil {
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

// DeleteSession deletes the chat session with the specified ID.
func (h *Handlers) DeleteSession(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	ctx, span := web.AddSpan(ctx, "chatsessionsgrp.handlers.DeleteSession")
	defer span.End()

	sessionID := web.Params(r, "sessionId")
	if len(sessionID) != 16 {
		h.log.Error(ctx, "invalid session id length", "session_id", sessionID)
		web.RespondError(ctx, w, ErrInvalidSessionID, http.StatusBadRequest)
		return nil
	}

	workspaceID, err := uuid.Parse(web.Params(r, "workspaceId"))
	if err != nil {
		h.log.Error(ctx, "invalid workspace id", "error", err)
		web.RespondError(ctx, w, ErrInvalidWorkspaceID, http.StatusBadRequest)
		return nil
	}

	if err := h.chatsessions.DeleteSession(ctx, sessionID, workspaceID); err != nil {
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

// SaveMessages saves messages for a chat session.
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

// GetMessages returns the messages for a chat session.
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
