package chathttp

import (
	"context"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *handlers) typing(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	var input struct {
		ThreadRootID *uuid.UUID `json:"threadRootId"`
		IsTyping     *bool      `json:"isTyping"`
	}
	r.Body = http.MaxBytesReader(w, r.Body, 1024)
	if err := web.Decode(r, &input); err != nil {
		return badRequest(ctx, w, err)
	}
	if input.IsTyping == nil {
		return badRequest(ctx, w, domain.ErrInvalid)
	}
	err = h.service.Typing(ctx, domain.TypingCommand{Scope: scope, ConversationID: id, ThreadRootID: input.ThreadRootID, IsTyping: *input.IsTyping})
	return h.respond(ctx, w, nil, err, http.StatusNoContent)
}
