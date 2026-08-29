package slackhttp

import (
	"context"
	"errors"
	"net/http"
	"time"

	slack "github.com/complexus-tech/projects-api/internal/modules/slack/service"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/web"
)

func (h *Handlers) HandleEvents(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	statusCode := http.StatusOK
	outcome := "received"
	errorMessage := ""
	headers := captureSlackHeaders(r.Header)
	var rawBody []byte
	verified := false
	defer func() {
		if !verified {
			return
		}
		h.recordRequestLogAsync(ctx, slack.CoreRequestLogInput{
			RequestType:  "events",
			Endpoint:     r.URL.Path,
			RawBody:      rawBody,
			Headers:      headers,
			ResponseCode: statusCode,
			Outcome:      outcome,
			ErrorMessage: errorMessage,
		})
	}()

	rawBody, err := readSlackBody(w, r)
	if err != nil {
		statusCode = http.StatusBadRequest
		outcome = "body_read_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusBadRequest)
	}
	if err := h.service.VerifyRequest(rawBody, r.Header); err != nil {
		statusCode = http.StatusUnauthorized
		outcome = "signature_verification_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, http.StatusUnauthorized)
	}
	verified = true
	eventCtx, cancel := context.WithTimeout(ctx, 2500*time.Millisecond)
	defer cancel()
	response, err := h.service.HandleSignedEvents(eventCtx, webhooks.SignedRequest{
		Method:        r.Method,
		RequestTarget: r.URL.RequestURI(),
		Headers:       webhooks.Headers(r.Header),
		Body:          rawBody,
	})
	if err != nil {
		statusCode = http.StatusServiceUnavailable
		if errors.Is(err, slack.ErrSlackInvalidEventPayload) {
			statusCode = http.StatusBadRequest
		}
		outcome = "event_handler_failed"
		errorMessage = err.Error()
		return web.RespondError(ctx, w, err, statusCode)
	}
	if response.Challenge != "" {
		outcome = "url_verification_ack"
		return writeRawJSON(w, http.StatusOK, response)
	}
	outcome = "processed"
	w.WriteHeader(http.StatusOK)
	return nil
}
