package githubhttp

import (
	"context"
	"net/http"

	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/web"
	"github.com/google/uuid"
)

func (h *Handlers) HandleWebhook(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	body, err := web.ReadBoundedBody(w, r, maxGitHubWebhookBodyBytes)
	if err != nil {
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	receipt, err := h.service.ReceiveWebhook(ctx, webhooks.SignedRequest{
		Method:        r.Method,
		RequestTarget: r.URL.RequestURI(),
		Headers: webhooks.Headers{
			"X-GitHub-Delivery":   append([]string(nil), r.Header.Values("X-GitHub-Delivery")...),
			"X-GitHub-Event":      append([]string(nil), r.Header.Values("X-GitHub-Event")...),
			"X-Hub-Signature-256": append([]string(nil), r.Header.Values("X-Hub-Signature-256")...),
		},
		Body: body,
	})
	if err != nil {
		return web.RespondError(ctx, w, err, webhooks.IngressHTTPStatus(err))
	}
	return web.Respond(ctx, w, map[string]any{
		"accepted":  true,
		"duplicate": !receipt.Created && receipt.ID != uuid.Nil,
	}, http.StatusAccepted)
}
