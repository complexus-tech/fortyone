package chathttp

import (
	"context"
	"net/http"

	domain "github.com/complexus-tech/projects-api/internal/modules/chat/domain"
)

func (h *handlers) listMembers(ctx context.Context, w http.ResponseWriter, r *http.Request, scope domain.ActorScope) error {
	id, err := requestID(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	limit, err := pageLimit(r)
	if err != nil {
		return badRequest(ctx, w, err)
	}
	result, err := h.service.ListMembers(ctx, domain.MemberSearch{Scope: scope, ConversationID: id, Query: r.URL.Query().Get("q"), Limit: limit})
	return h.respond(ctx, w, result, err, http.StatusOK)
}
