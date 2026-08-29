package emailreplyhttp

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	emailreply "github.com/complexus-tech/projects-api/internal/modules/emailreply/service"
	"github.com/stretchr/testify/require"
)

type ingressStub struct {
	authorized          bool
	err                 error
	body                []byte
	calls               int
	deadline            time.Time
	waitForCancellation bool
}

func (s *ingressStub) VerifyWebhookToken(_ string) bool {
	return s.authorized
}

func (s *ingressStub) Ingest(ctx context.Context, rawBody []byte) (emailreply.IngestResult, error) {
	s.calls++
	s.body = append([]byte(nil), rawBody...)
	s.deadline, _ = ctx.Deadline()
	if s.waitForCancellation {
		<-ctx.Done()
		return emailreply.IngestResult{}, ctx.Err()
	}
	return emailreply.IngestResult{Accepted: 1}, s.err
}

func TestHandleInboundEmailProcessedPersistsBeforeAcknowledgement(t *testing.T) {
	t.Parallel()

	service := &ingressStub{authorized: true}
	h := New(service)
	request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewBufferString(`{"items":[{}]}`))
	request.Header.Set("Content-Type", "application/json; charset=utf-8")
	request.Header.Set(emailreply.WebhookTokenHeader, "derived-token")
	response := httptest.NewRecorder()

	err := h.HandleInboundEmailProcessed(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusNoContent, response.Code)
	require.Equal(t, 1, service.calls)
	require.JSONEq(t, `{"items":[{}]}`, string(service.body))
}

func TestHandleInboundEmailProcessedRejectsUnauthenticatedRequestBeforeReadingBody(t *testing.T) {
	t.Parallel()

	service := &ingressStub{}
	h := New(service)
	request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewBufferString(`{"items":[{}]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	err := h.HandleInboundEmailProcessed(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusUnauthorized, response.Code)
	require.Zero(t, service.calls)
}

func TestHandleInboundEmailProcessedUsesBrevoRetryableStatus(t *testing.T) {
	t.Parallel()

	service := &ingressStub{authorized: true, err: errors.New("queue unavailable")}
	h := New(service)
	request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewBufferString(`{"items":[{}]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	err := h.HandleInboundEmailProcessed(context.Background(), response, request)

	require.NoError(t, err)
	require.Equal(t, http.StatusTooManyRequests, response.Code)
	require.Equal(t, brevoRetryAfterSeconds, response.Header().Get("Retry-After"))
}

func TestHandleInboundEmailProcessedRejectsMalformedAndOversizedBodies(t *testing.T) {
	t.Parallel()

	t.Run("malformed", func(t *testing.T) {
		service := &ingressStub{authorized: true, err: emailreply.ErrInvalidPayload}
		h := New(service)
		request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewBufferString(`{"items":`))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		require.NoError(t, h.HandleInboundEmailProcessed(context.Background(), response, request))
		require.Equal(t, http.StatusBadRequest, response.Code)
	})

	t.Run("oversized", func(t *testing.T) {
		service := &ingressStub{authorized: true}
		h := New(service)
		request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewReader(make([]byte, emailreply.MaximumInboundWebhookBytes+1)))
		request.Header.Set("Content-Type", "application/json")
		response := httptest.NewRecorder()

		require.NoError(t, h.HandleInboundEmailProcessed(context.Background(), response, request))
		require.Equal(t, http.StatusRequestEntityTooLarge, response.Code)
		require.Zero(t, service.calls)
	})
}

func TestHandleInboundEmailProcessedBoundsIngressWork(t *testing.T) {
	t.Parallel()

	service := &ingressStub{authorized: true, waitForCancellation: true}
	h := newWithTimeout(service, 5*time.Millisecond)
	request := httptest.NewRequest(http.MethodPost, "/webhooks/brevo/inbound-email-processed", bytes.NewBufferString(`{"items":[{}]}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	require.NoError(t, h.HandleInboundEmailProcessed(context.Background(), response, request))
	require.Equal(t, http.StatusTooManyRequests, response.Code)
	require.False(t, service.deadline.IsZero())
}
