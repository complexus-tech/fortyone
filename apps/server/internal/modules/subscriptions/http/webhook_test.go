package subscriptionshttp

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	subscriptions "github.com/complexus-tech/projects-api/internal/modules/subscriptions/service"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

type webhookEventServiceStub struct {
	err       error
	calls     int
	payload   []byte
	signature string
}

func (s *webhookEventServiceStub) HandleWebhookEvent(
	_ context.Context,
	payload []byte,
	signature string,
) error {
	s.calls++
	s.payload = append([]byte(nil), payload...)
	s.signature = signature
	return s.err
}

func TestHandleWebhookMapsInvalidSignatureToBadRequest(t *testing.T) {
	t.Parallel()

	service := &webhookEventServiceStub{err: subscriptions.ErrInvalidWebhookSignature}
	response := executeWebhookRequest(t, service, []byte(`{"id":"evt_invalid"}`), "bad-signature")

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
	if !strings.Contains(response.Body.String(), "invalid Stripe webhook signature") {
		t.Fatalf("response body = %s", response.Body.String())
	}
}

func TestHandleWebhookReturnsRetryableStatusForHandlerFailure(t *testing.T) {
	t.Parallel()

	service := &webhookEventServiceStub{
		err: fmtWebhookError(subscriptions.ErrWebhookEventProcessingFailed, "transient provider failure"),
	}
	response := executeWebhookRequest(t, service, []byte(`{"id":"evt_handler"}`), "valid-signature")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "transient provider failure") {
		t.Fatalf("internal failure leaked in response body: %s", response.Body.String())
	}
}

func TestHandleWebhookReturnsRetryableStatusForCompletionRecordFailure(t *testing.T) {
	t.Parallel()

	service := &webhookEventServiceStub{
		err: fmtWebhookError(subscriptions.ErrWebhookEventPersistenceFailed, "database unavailable"),
	}
	response := executeWebhookRequest(t, service, []byte(`{"id":"evt_record"}`), "valid-signature")

	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusInternalServerError)
	}
	if strings.Contains(response.Body.String(), "database unavailable") {
		t.Fatalf("persistence failure leaked in response body: %s", response.Body.String())
	}
}

func TestHandleWebhookAcceptsDurableTerminalResult(t *testing.T) {
	t.Parallel()

	service := &webhookEventServiceStub{}
	payload := []byte(`{"id":"evt_success"}`)
	response := executeWebhookRequest(t, service, payload, "valid-signature")

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if !bytes.Equal(service.payload, payload) || service.signature != "valid-signature" {
		t.Fatalf("service received payload %q and signature %q", service.payload, service.signature)
	}
}

func TestHandleWebhookRejectsOversizedSignedBodyBeforeService(t *testing.T) {
	t.Parallel()

	service := &webhookEventServiceStub{}
	response := executeWebhookRequest(t, service, make([]byte, 64*1024+1), "valid-signature")

	if response.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusRequestEntityTooLarge)
	}
	if service.calls != 0 {
		t.Fatalf("service calls = %d, want 0", service.calls)
	}
}

func executeWebhookRequest(
	t *testing.T,
	service webhookEventService,
	payload []byte,
	signature string,
) *httptest.ResponseRecorder {
	t.Helper()

	handler := &Handlers{
		webhookEvents: service,
		log:           logger.NewWithText(io.Discard, slog.LevelError, "stripe-webhook-http-test"),
	}
	request := httptest.NewRequest(http.MethodPost, "/webhooks/stripe", bytes.NewReader(payload))
	request.Header.Set("Stripe-Signature", signature)
	response := httptest.NewRecorder()

	if err := handler.HandleWebhook(t.Context(), response, request); err != nil {
		t.Fatalf("HandleWebhook returned error: %v", err)
	}
	return response
}

func fmtWebhookError(kind error, detail string) error {
	return errors.Join(kind, errors.New(detail))
}
