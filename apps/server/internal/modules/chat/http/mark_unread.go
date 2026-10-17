package chathttp

import (
	"context"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *handlers) markUnread(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	var input struct {
		FromMessageID uuid.UUID  `json:"fromMessageId"`
		ThreadRootID  *uuid.UUID `json:"threadRootId"`
	}
	if err = web.Decode(r, &input); err != nil {
		return badRequest(ctx, w, err)
	}
	w.Header().Set("Cache-Control", "private, no-store")
	value, err := h.service.MarkUnread(ctx, domain.MarkUnreadCommand{
		Scope: scope, ConversationID: id, FromMessageID: input.FromMessageID, ThreadRootID: input.ThreadRootID,
	})
	return h.respond(ctx, w, value, err, http.StatusOK)
}
