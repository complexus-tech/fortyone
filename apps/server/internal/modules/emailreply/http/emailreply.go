package emailreplyhttp

import (
	"context"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	"github.com/complexus-tech/projects-api/pkg/web"
)

const (
	maxInboundEmailWebhookBytes = 5 << 20
	brevoRetryAfterSeconds      = "600"
)

type Ingress interface {
	VerifyWebhookToken(provided string) bool
	Ingest(ctx context.Context, rawBody []byte) (emailreply.IngestResult, error)
}

type Handlers struct {
	service Ingress
}

func New(service Ingress) *Handlers {
	return &Handlers{service: service}
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

	r.Body = http.MaxBytesReader(w, r.Body, maxInboundEmailWebhookBytes)
	rawBody, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return web.RespondError(ctx, w, nil, http.StatusRequestEntityTooLarge)
		}
		w.Header().Set("Retry-After", brevoRetryAfterSeconds)
		return web.RespondError(ctx, w, nil, http.StatusTooManyRequests)
	}

	if _, err := h.service.Ingest(ctx, rawBody); err != nil {
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
