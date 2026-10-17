package chathttp

import (
	"context"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/google/uuid"
)

func (h *handlers) unread(ctx context.Context, w http.ResponseWriter, _ *http.Request, scope domain.ActorScope) error {
	w.Header().Set("Cache-Control", "private, no-store")
	value, err := h.service.Unread(ctx, scope)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) unreadThreads(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	limit, err := pageLimit(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	query := domain.UnreadThreadQuery{Scope: scope, Cursor: r.URL.Query().Get("cursor"), Limit: limit}
	if raw := r.URL.Query().Get("conversationId"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil || id == uuid.Nil {
			return badRequest(ctx, w, domain.ErrInvalid)
		}
		query.ConversationID = &id
	}
	w.Header().Set("Cache-Control", "private, no-store")
	value, err := h.service.UnreadThreads(ctx, query)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) markAllThreadsRead(ctx context.Context, w http.ResponseWriter, _ *http.Request, scope domain.ActorScope) error {
	err := h.service.MarkAllThreadsRead(ctx, scope)
	return h.respond(ctx, w, nil, err, http.StatusNoContent)
}
