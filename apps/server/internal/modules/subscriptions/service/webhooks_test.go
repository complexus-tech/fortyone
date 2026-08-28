package subscriptions

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v82"
	"github.com/stripe/stripe-go/v82/webhook"
)

const testWebhookSecret = "whsec_stripe_webhook_reliability_test"

type webhookProcessorStub struct {
	process func(context.Context, stripe.Event) (WebhookOutcome, error)
}

func (s webhookProcessorStub) ProcessWebhookEvent(
	ctx context.Context,
	event stripe.Event,
) (WebhookOutcome, error) {
	return s.process(ctx, event)
}

type webhookRecord struct {
	state          string
	result         WebhookProcessingResult
	leaseToken     uuid.UUID
	leaseExpiresAt time.Time
	attempts       int
	failureCode    WebhookFailureCode
}

type webhookRepositoryStub struct {
	mu sync.Mutex

	records map[string]webhookRecord

	claimErr         error
	markProcessedErr error
	markFailedErr    error
	claimCalls       int
}

func newWebhookRepositoryStub() *webhookRepositoryStub {
	return &webhookRepositoryStub{records: make(map[string]webhookRecord)}
}

func (r *webhookRepositoryStub) ClaimWebhookEvent(
	_ context.Context,
	eventID string,
	_ string,
	attemptedAt time.Time,
	leaseExpiresAt time.Time,
) (WebhookClaim, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.claimCalls++

	if r.claimErr != nil {
		return WebhookClaim{}, r.claimErr
	}

	record, exists := r.records[eventID]
	if exists {
		switch {
		case record.state == "processed":
			return WebhookClaim{Disposition: WebhookClaimAlreadyProcessed, Attempt: record.attempts}, nil
		case record.state == "processing" && record.leaseExpiresAt.After(attemptedAt):
			return WebhookClaim{Disposition: WebhookClaimInProgress, Attempt: record.attempts}, nil
		}
	}

	record.state = "processing"
	record.result = ""
	record.failureCode = ""
	record.leaseToken = uuid.New()
	record.leaseExpiresAt = leaseExpiresAt
	record.attempts++
	r.records[eventID] = record
	return WebhookClaim{
		Disposition: WebhookClaimAcquired,
		LeaseToken:  record.leaseToken,
		Attempt:     record.attempts,
	}, nil
}

func (r *webhookRepositoryStub) MarkWebhookEventProcessed(
	_ context.Context,
	eventID string,
	leaseToken uuid.UUID,
	outcome WebhookOutcome,
	_ time.Time,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.markProcessedErr != nil {
		return r.markProcessedErr
	}
	record := r.records[eventID]
	if record.state != "processing" || record.leaseToken != leaseToken {
		return ErrWebhookEventClaimLost
	}
	record.state = "processed"
	record.result = outcome.Result
	record.leaseToken = uuid.Nil
	record.leaseExpiresAt = time.Time{}
	r.records[eventID] = record
	return nil
}

func (r *webhookRepositoryStub) MarkWebhookEventFailed(
	_ context.Context,
	eventID string,
	leaseToken uuid.UUID,
	failureCode WebhookFailureCode,
	_ time.Time,
) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.markFailedErr != nil {
		return r.markFailedErr
	}
	record := r.records[eventID]
	if record.state != "processing" || record.leaseToken != leaseToken {
		return ErrWebhookEventClaimLost
	}
	record.state = "failed"
	record.failureCode = failureCode
	record.leaseToken = uuid.Nil
	record.leaseExpiresAt = time.Time{}
	r.records[eventID] = record
	return nil
}

func (r *webhookRepositoryStub) record(eventID string) webhookRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.records[eventID]
}

func (r *webhookRepositoryStub) claims() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.claimCalls
}

func TestHandleWebhookEventRejectsInvalidSignatureBeforeClaim(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			t.Fatal("invalid signature reached the event processor")
			return WebhookOutcome{}, nil
		},
	})

	err := service.HandleWebhookEvent(t.Context(), []byte(`{"id":"evt_invalid"}`), "invalid")
	if !errors.Is(err, ErrInvalidWebhookSignature) {
		t.Fatalf("HandleWebhookEvent error = %v, want ErrInvalidWebhookSignature", err)
	}
	if repository.claims() != 0 {
		t.Fatalf("claim calls = %d, want 0", repository.claims())
	}
}

func TestHandleWebhookEventMarksHandlerFailureRetryable(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			return WebhookOutcome{Result: WebhookResultHandled}, errors.New("transient handler failure")
		},
	})
	payload, signature := signedWebhook(t, "evt_handler_failure", "invoice.paid")

	err := service.HandleWebhookEvent(t.Context(), payload, signature)
	if !errors.Is(err, ErrWebhookEventProcessingFailed) {
		t.Fatalf("HandleWebhookEvent error = %v, want ErrWebhookEventProcessingFailed", err)
	}
	record := repository.record("evt_handler_failure")
	if record.state != "failed" || record.failureCode != WebhookFailureHandler || record.result != "" {
		t.Fatalf("failure record = %+v", record)
	}
}

func TestHandleWebhookEventReturnsPersistenceFailureWhenFailureStateCannotBeRecorded(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	repository.markFailedErr = errors.New("database unavailable")
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			return WebhookOutcome{Result: WebhookResultHandled}, errors.New("handler failure")
		},
	})
	payload, signature := signedWebhook(t, "evt_failure_record", "invoice.paid")

	err := service.HandleWebhookEvent(t.Context(), payload, signature)
	if !errors.Is(err, ErrWebhookEventPersistenceFailed) || !errors.Is(err, ErrWebhookEventProcessingFailed) {
		t.Fatalf("HandleWebhookEvent error = %v, want processing and persistence failures", err)
	}
	if record := repository.record("evt_failure_record"); record.state != "processing" {
		t.Fatalf("record state = %q, want processing until lease expiry", record.state)
	}
}

func TestHandleWebhookEventDoesNotAcknowledgeCompletionPersistenceFailure(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	repository.markProcessedErr = errors.New("database unavailable")
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			return WebhookOutcome{Result: WebhookResultHandled}, nil
		},
	})
	payload, signature := signedWebhook(t, "evt_completion_failure", "invoice.paid")

	err := service.HandleWebhookEvent(t.Context(), payload, signature)
	if !errors.Is(err, ErrWebhookEventPersistenceFailed) {
		t.Fatalf("HandleWebhookEvent error = %v, want ErrWebhookEventPersistenceFailed", err)
	}
	if record := repository.record("evt_completion_failure"); record.state != "processing" {
		t.Fatalf("record state = %q, want processing until lease expiry", record.state)
	}
}

func TestHandleWebhookEventProcessesSuccessfulDuplicateOnce(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	var calls atomic.Int32
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			calls.Add(1)
			return WebhookOutcome{Result: WebhookResultHandled}, nil
		},
	})
	payload, signature := signedWebhook(t, "evt_duplicate", "invoice.paid")

	if err := service.HandleWebhookEvent(t.Context(), payload, signature); err != nil {
		t.Fatalf("first HandleWebhookEvent: %v", err)
	}
	if err := service.HandleWebhookEvent(t.Context(), payload, signature); err != nil {
		t.Fatalf("duplicate HandleWebhookEvent: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("processor calls = %d, want 1", calls.Load())
	}
}

func TestHandleWebhookEventDoesNotProcessConcurrentDuplicate(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	entered := make(chan struct{})
	release := make(chan struct{})
	var calls atomic.Int32
	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			if calls.Add(1) == 1 {
				close(entered)
			}
			<-release
			return WebhookOutcome{Result: WebhookResultHandled}, nil
		},
	})
	payload, signature := signedWebhook(t, "evt_concurrent", "invoice.paid")

	firstResult := make(chan error, 1)
	go func() {
		firstResult <- service.HandleWebhookEvent(context.Background(), payload, signature)
	}()

	select {
	case <-entered:
	case <-time.After(2 * time.Second):
		t.Fatal("first event did not reach processor")
	}

	if err := service.HandleWebhookEvent(t.Context(), payload, signature); !errors.Is(err, ErrAlreadyProcessingEvent) {
		t.Fatalf("concurrent duplicate error = %v, want ErrAlreadyProcessingEvent", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("processor calls during concurrent delivery = %d, want 1", calls.Load())
	}

	close(release)
	if err := <-firstResult; err != nil {
		t.Fatalf("first HandleWebhookEvent: %v", err)
	}
	if err := service.HandleWebhookEvent(t.Context(), payload, signature); err != nil {
		t.Fatalf("terminal duplicate HandleWebhookEvent: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("processor calls after terminal duplicate = %d, want 1", calls.Load())
	}
}

func TestHandleWebhookEventReclaimsStaleLeaseAndRejectsOldOwner(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	firstAttemptAt := time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
	staleClaim, err := repository.ClaimWebhookEvent(
		t.Context(),
		"evt_stale",
		"invoice.paid",
		firstAttemptAt,
		firstAttemptAt.Add(defaultWebhookEventLease),
	)
	if err != nil {
		t.Fatalf("create stale claim: %v", err)
	}

	service := newWebhookTestService(repository, webhookProcessorStub{
		process: func(context.Context, stripe.Event) (WebhookOutcome, error) {
			return WebhookOutcome{Result: WebhookResultHandled}, nil
		},
	})
	service.webhookClock = func() time.Time {
		return firstAttemptAt.Add(defaultWebhookEventLease + time.Second)
	}
	payload, signature := signedWebhook(t, "evt_stale", "invoice.paid")

	if err := service.HandleWebhookEvent(t.Context(), payload, signature); err != nil {
		t.Fatalf("retry stale HandleWebhookEvent: %v", err)
	}
	record := repository.record("evt_stale")
	if record.state != "processed" || record.attempts != 2 {
		t.Fatalf("reclaimed record = %+v, want processed attempt 2", record)
	}
	if err := repository.MarkWebhookEventProcessed(
		t.Context(),
		"evt_stale",
		staleClaim.LeaseToken,
		WebhookOutcome{Result: WebhookResultHandled},
		firstAttemptAt,
	); !errors.Is(err, ErrWebhookEventClaimLost) {
		t.Fatalf("stale owner completion error = %v, want ErrWebhookEventClaimLost", err)
	}
}

func TestHandleWebhookEventDurablyIgnoresUnsupportedType(t *testing.T) {
	t.Parallel()

	repository := newWebhookRepositoryStub()
	service := newWebhookTestService(repository, nil)
	service.webhookEvents = serviceWebhookEventProcessor{service: service}
	payload, signature := signedWebhook(t, "evt_ignored", "product.updated")

	if err := service.HandleWebhookEvent(t.Context(), payload, signature); err != nil {
		t.Fatalf("HandleWebhookEvent: %v", err)
	}
	record := repository.record("evt_ignored")
	if record.state != "processed" || record.result != WebhookResultIgnored {
		t.Fatalf("ignored record = %+v", record)
	}
}

func newWebhookTestService(repository WebhookRepository, processor webhookEventProcessor) *Service {
	return &Service{
		webhookRepo:   repository,
		webhookEvents: processor,
		log:           logger.NewWithText(io.Discard, slog.LevelError, "stripe-webhook-test"),
		webhookSecret: testWebhookSecret,
		webhookClock: func() time.Time {
			return time.Date(2026, time.August, 28, 12, 0, 0, 0, time.UTC)
		},
		webhookLease: defaultWebhookEventLease,
	}
}

func signedWebhook(t *testing.T, eventID string, eventType stripe.EventType) ([]byte, string) {
	t.Helper()

	payload := []byte(
		`{"id":"` + eventID +
			`","object":"event","api_version":"` + stripe.APIVersion +
			`","type":"` + string(eventType) + `","data":{"object":{}}}`,
	)
	signed := webhook.GenerateTestSignedPayload(&webhook.UnsignedPayload{
		Payload:   payload,
		Secret:    testWebhookSecret,
		Timestamp: time.Now(),
	})
	return signed.Payload, signed.Header
}
