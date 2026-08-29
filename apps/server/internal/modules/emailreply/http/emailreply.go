package emailreplyhttp

import (
	"context"
	"errors"
	"mime"
	"net/http"
	"strings"
	"time"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	brevoRetryAfterSeconds     = "600"
	inboundEmailWebhookTimeout = 15 * time.Second
)

type Ingress interface {
	VerifyWebhookToken(provided string) bool
	Ingest(ctx context.Context, rawBody []byte) (emailreply.IngestResult, error)
}

type Handlers struct {
	service Ingress
	timeout time.Duration
}

func New(service Ingress) *Handlers {
	return newWithTimeout(service, inboundEmailWebhookTimeout)
}

func newWithTimeout(service Ingress, timeout time.Duration) *Handlers {
	if timeout <= 0 {
		timeout = inboundEmailWebhookTimeout
	}
	return &Handlers{service: service, timeout: timeout}
}

func (h *Handlers) HandleInboundEmailProcessed(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	if h == nil || h.service == nil {
		w.Header().Set("Retry-After", brevoRetryAfterSeconds)
		return web.RespondError(ctx, w, nil, http.StatusTooManyRequests)
	}
	if !h.service.VerifyWebhookToken(r.Header.Get(emailreply.WebhookTokenHeader)) {
		return web.RespondError(ctx, w, nil, http.StatusUnauthorized)
	}
	if !isJSONContentType(r.Header.Get("Content-Type")) {
		return web.RespondError(ctx, w, nil, http.StatusUnsupportedMediaType)
	}

	rawBody, err := web.ReadBoundedBody(w, r, emailreply.MaximumInboundWebhookBytes)
	if err != nil {
		if errors.Is(err, web.ErrRequestBodyTooLarge) {
			return web.RespondError(ctx, w, nil, http.StatusRequestEntityTooLarge)
		}
		w.Header().Set("Retry-After", brevoRetryAfterSeconds)
		return web.RespondError(ctx, w, nil, http.StatusTooManyRequests)
	}

	ingestCtx, cancel := context.WithTimeout(ctx, h.timeout)
	defer cancel()
	if _, err := h.service.Ingest(ingestCtx, rawBody); err != nil {
		switch {
		case errors.Is(err, emailreply.ErrInvalidPayload):
			return web.RespondError(ctx, w, nil, http.StatusBadRequest)
		case errors.Is(err, emailreply.ErrReplyNotAuthorized), errors.Is(err, emailreply.ErrInvalidReplyToken):
			return web.RespondError(ctx, w, nil, http.StatusUnauthorized)
		default:
			// Brevo discards webhooks after all 5xx and most 4xx responses. A
			// 429 is the only documented response that triggers its retry policy.
			w.Header().Set("Retry-After", brevoRetryAfterSeconds)
			return web.RespondError(ctx, w, nil, http.StatusTooManyRequests)
		}
	}
	return web.Respond(ctx, w, nil, http.StatusNoContent)
}

func isJSONContentType(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" {
		return false
	}
	mediaType, _, err := mime.ParseMediaType(value)
	return err == nil && strings.EqualFold(mediaType, "application/json")
}
