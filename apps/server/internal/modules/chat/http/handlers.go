package chathttp

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func status(err error) int {
	switch {
	case errors.Is(err, domain.ErrInvalid):
		return http.StatusBadRequest
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

func badRequest(ctx context.Context, w http.ResponseWriter, err error) error {
	return web.RespondError(ctx, w, err, http.StatusBadRequest)
}

func pageLimit(r *http.Request) (int, error) {
	value := r.URL.Query().Get("limit")
	if value == "" {
		return 0, nil
	}
	limit, err := strconv.Atoi(value)
	if err != nil || limit < 1 || limit > 200 {
		return 0, domain.ErrInvalid
	}
	return limit, nil
}

func requestID(r *http.Request) (uuid.UUID, error) { return web.UUIDPathParameter(r, "id") }

func messageQuery(r *http.Request, scope domain.ActorScope, id uuid.UUID) (domain.MessageQuery, error) {
	limit, err := pageLimit(r)
	if err != nil {
		return domain.MessageQuery{}, err
	}
	query := r.URL.Query()
	result := domain.MessageQuery{Scope: scope, ConversationID: id, Before: query.Get("before"), After: query.Get("after"), Limit: limit}
	if result.Before == "" {
		result.Before = query.Get("cursor")
	}
	if around := query.Get("around"); around != "" {
		parsed, err := uuid.Parse(around)
		if err != nil || parsed == uuid.Nil {
			return result, domain.ErrInvalid
		}
		result.Around = &parsed
	}
	return result, nil
}

func (h *handlers) listConversations(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	limit, err := pageLimit(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.ListConversations(ctx, domain.ListQuery{Scope: scope, Cursor: r.URL.Query().Get("cursor"), Search: r.URL.Query().Get("search"), Limit: limit})
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) getConversation(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.GetConversation(ctx, scope, id)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) createDirect(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	var input struct {
		UserID uuid.UUID `json:"userId"`
	}
	if err := web.Decode(r, &input); err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.CreateDirect(ctx, scope, input.UserID)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) getMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.GetMessage(ctx, scope, id)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) listMessages(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	query, err := messageQuery(r, scope, id)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.ListMessages(ctx, query)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) listReplies(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	root, err := h.service.GetMessage(ctx, scope, id)
	if err != nil {
		return h.respond(ctx, w, nil, err, http.StatusOK)
	}
	if root.ThreadRootID != nil {
		return h.respond(ctx, w, nil, domain.ErrNotFound, http.StatusOK)
	}
	query, err := messageQuery(r, scope, root.ConversationID)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	query.ThreadRootID = &id
	value, err := h.service.ListMessages(ctx, query)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

type messageInput struct {
	CreationID       uuid.UUID       `json:"creationId"`
	ThreadRootID     *uuid.UUID      `json:"threadRootId"`
	Version          int             `json:"version"`
	Text             string          `json:"text"`
	RichContent      json.RawMessage `json:"richContent"`
	MentionedUserIDs []uuid.UUID     `json:"mentionedUserIds"`
	AgentMentions    []string        `json:"agentMentions"`
	AttachmentIDs    []uuid.UUID     `json:"attachmentIds"`
}

func decodeMessage(w http.ResponseWriter, r *http.Request) (messageInput, error) {
	var input messageInput
	r.Body = http.MaxBytesReader(w, r.Body, 96*1024)
	err := web.Decode(r, &input)
	return input, err
}

func (h *handlers) createMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	input, err := decodeMessage(w, r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.CreateMessage(ctx, domain.CreateMessageCommand{Scope: scope, ConversationID: id, CreationID: input.CreationID, ThreadRootID: input.ThreadRootID, Text: input.Text, RichContent: input.RichContent, MentionedUserIDs: input.MentionedUserIDs, AgentMentions: input.AgentMentions, AttachmentIDs: input.AttachmentIDs})
	code := http.StatusCreated
	if value.Replayed {
		code = http.StatusOK
	}
	return h.respond(ctx, w, value, err, code)
}

func (h *handlers) updateMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	input, err := decodeMessage(w, r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.UpdateMessage(ctx, domain.UpdateMessageCommand{Scope: scope, MessageID: id, Version: input.Version, Text: input.Text, RichContent: input.RichContent, MentionedUserIDs: input.MentionedUserIDs, AgentMentions: input.AgentMentions, AttachmentIDs: input.AttachmentIDs})
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) deleteMessage(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	version, err := strconv.Atoi(r.URL.Query().Get("version"))
	if err != nil {
		return badRequest(ctx, w, domain.ErrInvalid)
	}
	value, err := h.service.DeleteMessage(ctx, domain.DeleteMessageCommand{Scope: scope, MessageID: id, Version: version})
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) setReaction(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	emoji, err := web.RequiredPathParameter(r, "emoji", 64)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.SetReaction(ctx, scope, id, emoji, r.Method == http.MethodPut)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) markRead(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	var input struct {
		ThroughMessageID uuid.UUID  `json:"throughMessageId"`
		ThreadRootID     *uuid.UUID `json:"threadRootId"`
	}
	if err = web.Decode(r, &input); err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.MarkRead(ctx, domain.ReadCommand{Scope: scope, ConversationID: id, ThroughMessageID: input.ThroughMessageID, ThreadRootID: input.ThreadRootID})
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func resourceQuery(r *http.Request, scope domain.ActorScope) (domain.SearchQuery, error) {
	id, err := requestID(r)
	if err != nil {
		return domain.SearchQuery{}, err
	}
	limit, err := pageLimit(r)
	if err != nil {
		return domain.SearchQuery{}, err
	}
	q := r.URL.Query()
	return domain.SearchQuery{Scope: scope, ConversationID: id, Query: q.Get("q"), Cursor: q.Get("cursor"), Kind: q.Get("kind"), Sort: q.Get("sort"), Limit: limit}, nil
}

func (h *handlers) search(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	query, err := resourceQuery(r, scope)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.Search(ctx, query)
	return h.respond(ctx, w, value, err, http.StatusOK)
}

func (h *handlers) resources(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	query, err := resourceQuery(r, scope)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	value, err := h.service.Resources(ctx, query)
	return h.respond(ctx, w, value, err, http.StatusOK)
}
