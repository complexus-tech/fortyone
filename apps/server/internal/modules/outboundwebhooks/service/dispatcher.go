package outboundwebhooksservice

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	outboundwebhooksdomain "github.com/complexus-tech/projects-api/internal/modules/outboundwebhooks/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/google/uuid"
)

const (
	defaultDeliveryLease       = 30 * time.Second
	defaultMaximumAttempts     = 8
	defaultDisableAfterFailure = 20
	webhookUserAgent           = "FortyOne-Webhooks/1.0"
)

var retrySchedule = []time.Duration{
	time.Minute,
	5 * time.Minute,
	30 * time.Minute,
	2 * time.Hour,
	8 * time.Hour,
	24 * time.Hour,
}

type DeliveryRepository interface {
	ClaimNextDelivery(context.Context, uuid.UUID, time.Time, time.Time) (outboundwebhooksdomain.ClaimedDelivery, error)
	CompleteAttempt(context.Context, outboundwebhooksdomain.DeliveryAttempt, uuid.UUID, uuid.UUID) error
	RecoverExpiredLeases(context.Context, time.Time) (int64, error)
}

type SafeHTTPClient interface {
	Do(context.Context, *http.Request) (safehttp.Result, error)
}

type Dispatcher struct {
	repository           DeliveryRepository
	secrets              *SecretManager
	http                 SafeHTTPClient
	clock                Clock
	ids                  IDGenerator
	leaseDuration        time.Duration
	maximumAttempts      int
	disableAfterFailures int
}

func NewDispatcher(repository DeliveryRepository, secrets *SecretManager, httpClient SafeHTTPClient) (*Dispatcher, error) {
	return newDispatcher(
		repository, secrets, httpClient, systemClock{}, randomIDGenerator{},
		defaultDeliveryLease, defaultMaximumAttempts, defaultDisableAfterFailure,
	)
}

func newDispatcher(
	repository DeliveryRepository,
	secrets *SecretManager,
	httpClient SafeHTTPClient,
	clock Clock,
	ids IDGenerator,
	leaseDuration time.Duration,
	maximumAttempts int,
	disableAfterFailures int,
) (*Dispatcher, error) {
	if repository == nil || secrets == nil || httpClient == nil || clock == nil || ids == nil {
		return nil, errors.New("outbound webhook dispatcher dependencies are required")
	}
	if leaseDuration < 10*time.Second || leaseDuration > 2*time.Minute || maximumAttempts < 1 || maximumAttempts > 32 ||
		disableAfterFailures < 1 || disableAfterFailures > 1000 {
		return nil, errors.New("outbound webhook dispatcher policy is invalid")
	}
	return &Dispatcher{
		repository: repository, secrets: secrets, http: httpClient, clock: clock, ids: ids,
		leaseDuration: leaseDuration, maximumAttempts: maximumAttempts, disableAfterFailures: disableAfterFailures,
	}, nil
}

// DispatchOne claims and completes at most one delivery. A successfully
// persisted retry/terminal outcome is not returned as an infrastructure error;
// callers may immediately claim another endpoint without hot-looping one
// failing destination.
func (dispatcher *Dispatcher) DispatchOne(ctx context.Context) (bool, error) {
	now := dispatcher.clock.Now().UTC()
	leaseToken, err := dispatcher.ids.NewUUID()
	if err != nil {
		return false, fmt.Errorf("generate outbound webhook lease token: %w", err)
	}
	delivery, err := dispatcher.repository.ClaimNextDelivery(ctx, leaseToken, now, now.Add(dispatcher.leaseDuration))
	if errors.Is(err, outboundwebhooksdomain.ErrDeliveryNotFound) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("claim outbound webhook delivery: %w", err)
	}
	attempt := dispatcher.deliver(ctx, delivery, now)
	if err := dispatcher.repository.CompleteAttempt(ctx, attempt, delivery.WorkspaceID, delivery.EndpointID); err != nil {
		return true, fmt.Errorf("complete outbound webhook delivery: %w", err)
	}
	return true, nil
}

func (dispatcher *Dispatcher) RecoverExpiredLeases(ctx context.Context) (int64, error) {
	return dispatcher.repository.RecoverExpiredLeases(ctx, dispatcher.clock.Now().UTC())
}

func (dispatcher *Dispatcher) deliver(ctx context.Context, delivery outboundwebhooksdomain.ClaimedDelivery, startedAt time.Time) outboundwebhooksdomain.DeliveryAttempt {
	attemptID, idErr := dispatcher.ids.NewUUID()
	if idErr != nil {
		// A deterministic non-zero fallback preserves the durable attempt/lease
		// transition even if the random source fails after the network claim.
		attemptID = fallbackAttemptID(delivery)
	}
	attempt := outboundwebhooksdomain.DeliveryAttempt{
		ID: attemptID, DeliveryID: delivery.ID, LeaseToken: delivery.LeaseToken,
		AttemptNumber: delivery.AttemptNumber, StartedAt: startedAt,
		DisableAfterFailures: dispatcher.disableAfterFailures,
	}

	secret, err := dispatcher.secrets.Open(
		delivery.WorkspaceID, delivery.EndpointID, delivery.SecretGeneration, delivery.SigningSecretEnvelope,
	)
	if err != nil {
		if errors.Is(err, credentialvault.ErrNotConfigured) || errors.Is(err, credentialvault.ErrUnknownKey) {
			return dispatcher.retryWithoutResponse(attempt, delivery, "secret_vault_unavailable")
		}
		return dispatcher.finishWithoutResponse(attempt, "secret_invalid", true, true)
	}
	defer clear(secret)
	headers, err := Sign(delivery.ID, startedAt, delivery.PayloadBody, secret)
	if err != nil {
		return dispatcher.finishWithoutResponse(attempt, "secret_invalid", true, true)
	}
	headers.WebhookSignature = dispatcher.previousSignature(delivery, startedAt, headers.WebhookSignature)

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, delivery.EndpointURL, bytes.NewReader(delivery.PayloadBody))
	if err != nil {
		return dispatcher.finishWithoutResponse(attempt, "endpoint_invalid", true, true)
	}
	defer request.Body.Close()
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("User-Agent", webhookUserAgent)
	request.Header.Set("Webhook-Id", headers.WebhookID)
	request.Header.Set("Webhook-Timestamp", headers.WebhookTimestamp)
	request.Header.Set("Webhook-Signature", headers.WebhookSignature)

	result, requestErr := dispatcher.http.Do(ctx, request)
	finishedAt := dispatcher.clock.Now().UTC()
	if finishedAt.Before(startedAt) {
		finishedAt = startedAt
	}
	attempt.FinishedAt = finishedAt
	attempt.Duration = result.Duration
	if attempt.Duration <= 0 {
		attempt.Duration = finishedAt.Sub(startedAt)
	}
	if attempt.Duration > 30*time.Second {
		attempt.Duration = 30 * time.Second
	}
	if result.ResolvedIP.IsValid() {
		resolved := result.ResolvedIP.Unmap()
		attempt.ResolvedIP = &resolved
	}
	if result.StatusCode > 0 {
		status := result.StatusCode
		bytesRead := int(result.ResponseBytes)
		attempt.HTTPStatus = &status
		attempt.ResponseBytes = &bytesRead
		attempt.ResponseDigest = append([]byte(nil), result.ResponseDigest[:]...)
	}

	classification := classifyDelivery(result, requestErr, delivery.AttemptNumber, dispatcher.maximumAttempts)
	attempt.Outcome = classification.outcome
	attempt.ErrorCode = classification.errorCode
	attempt.DisableEndpoint = classification.disableEndpoint
	attempt.CountEndpointFailure = classification.countEndpointFailure
	if classification.outcome == outboundwebhooksdomain.AttemptRetryScheduled {
		delay := retryDelay(delivery.ID, delivery.AttemptNumber)
		if result.RetryAfter > delay {
			delay = result.RetryAfter
		}
		next := finishedAt.Add(delay)
		attempt.NextAttemptAt = &next
	}
	return attempt
}

func (dispatcher *Dispatcher) previousSignature(
	delivery outboundwebhooksdomain.ClaimedDelivery,
	timestamp time.Time,
	current string,
) string {
	if delivery.PreviousSecretEnvelope == nil || delivery.PreviousSecretGeneration == nil ||
		delivery.PreviousSecretExpiresAt == nil || !delivery.PreviousSecretExpiresAt.After(timestamp) {
		return current
	}
	secret, err := dispatcher.secrets.Open(
		delivery.WorkspaceID, delivery.EndpointID, *delivery.PreviousSecretGeneration, *delivery.PreviousSecretEnvelope,
	)
	if err != nil {
		return current
	}
	defer clear(secret)
	previous, err := Sign(delivery.ID, timestamp, delivery.PayloadBody, secret)
	if err != nil {
		return current
	}
	return current + " " + previous.WebhookSignature
}

func (dispatcher *Dispatcher) finishWithoutResponse(
	attempt outboundwebhooksdomain.DeliveryAttempt,
	errorCode string,
	disable bool,
	countEndpointFailure bool,
) outboundwebhooksdomain.DeliveryAttempt {
	finishedAt := dispatcher.clock.Now().UTC()
	if finishedAt.Before(attempt.StartedAt) {
		finishedAt = attempt.StartedAt
	}
	attempt.Outcome = outboundwebhooksdomain.AttemptFailed
	attempt.ErrorCode = errorCode
	attempt.DisableEndpoint = disable
	attempt.CountEndpointFailure = countEndpointFailure
	attempt.FinishedAt = finishedAt
	attempt.Duration = finishedAt.Sub(attempt.StartedAt)
	return attempt
}

func (dispatcher *Dispatcher) retryWithoutResponse(
	attempt outboundwebhooksdomain.DeliveryAttempt,
	delivery outboundwebhooksdomain.ClaimedDelivery,
	errorCode string,
) outboundwebhooksdomain.DeliveryAttempt {
	finishedAt := dispatcher.clock.Now().UTC()
	if finishedAt.Before(attempt.StartedAt) {
		finishedAt = attempt.StartedAt
	}
	attempt.FinishedAt = finishedAt
	attempt.Duration = finishedAt.Sub(attempt.StartedAt)
	attempt.ErrorCode = errorCode
	if delivery.AttemptNumber >= dispatcher.maximumAttempts {
		attempt.Outcome = outboundwebhooksdomain.AttemptFailed
		attempt.ErrorCode = "attempts_exhausted"
		return attempt
	}
	attempt.Outcome = outboundwebhooksdomain.AttemptRetryScheduled
	next := finishedAt.Add(retryDelay(delivery.ID, delivery.AttemptNumber))
	attempt.NextAttemptAt = &next
	return attempt
}

func fallbackAttemptID(delivery outboundwebhooksdomain.ClaimedDelivery) uuid.UUID {
	digest := sha256.Sum256([]byte(delivery.ID.String() + ":" + strconv.Itoa(delivery.AttemptNumber)))
	id, err := uuid.FromBytes(digest[:16])
	if err != nil || id == uuid.Nil {
		return uuid.MustParse("00000000-0000-4000-8000-000000000001")
	}
	return id
}

type deliveryClassification struct {
	outcome              outboundwebhooksdomain.AttemptOutcome
	errorCode            string
	disableEndpoint      bool
	countEndpointFailure bool
}

func classifyDelivery(result safehttp.Result, requestErr error, attemptNumber, maximumAttempts int) deliveryClassification {
	if result.StatusCode >= 200 && result.StatusCode <= 299 {
		return deliveryClassification{outcome: outboundwebhooksdomain.AttemptSucceeded}
	}
	terminal := func(code string, disable bool) deliveryClassification {
		return deliveryClassification{
			outcome: outboundwebhooksdomain.AttemptFailed, errorCode: code,
			disableEndpoint: disable, countEndpointFailure: true,
		}
	}
	retry := func(code string) deliveryClassification {
		if attemptNumber >= maximumAttempts {
			return terminal("attempts_exhausted", false)
		}
		return deliveryClassification{
			outcome: outboundwebhooksdomain.AttemptRetryScheduled, errorCode: code,
			countEndpointFailure: true,
		}
	}

	if errors.Is(requestErr, safehttp.ErrUnsafeAddress) || errors.Is(requestErr, safehttp.ErrInsecureScheme) ||
		errors.Is(requestErr, safehttp.ErrIPAddressHost) || errors.Is(requestErr, safehttp.ErrUnsupportedPort) ||
		errors.Is(requestErr, safehttp.ErrCredentialsInURL) || errors.Is(requestErr, safehttp.ErrFragmentInURL) {
		return terminal("unsafe_endpoint", true)
	}
	if errors.Is(requestErr, safehttp.ErrRedirectDenied) {
		return terminal("redirect_denied", true)
	}
	if result.StatusCode == http.StatusGone {
		return terminal("http_410", true)
	}
	if result.StatusCode == http.StatusRequestTimeout || result.StatusCode == http.StatusTooEarly ||
		result.StatusCode == http.StatusTooManyRequests {
		return retry("http_" + strconv.Itoa(result.StatusCode))
	}
	if result.StatusCode >= 500 {
		return retry("http_5xx")
	}
	if result.StatusCode >= 300 {
		return terminal("http_"+strconv.Itoa(result.StatusCode), false)
	}
	if requestErr != nil {
		return retry("network_error")
	}
	return retry("empty_response")
}

func retryDelay(deliveryID uuid.UUID, attemptNumber int) time.Duration {
	index := attemptNumber - 1
	if index < 0 {
		index = 0
	}
	if index >= len(retrySchedule) {
		index = len(retrySchedule) - 1
	}
	base := retrySchedule[index]
	digest := sha256.Sum256([]byte(deliveryID.String() + ":" + strconv.Itoa(attemptNumber)))
	// Stable +/-10% jitter prevents synchronized retries and remains exactly
	// reproducible after a worker restart.
	partsPerThousand := int64(900 + int(digest[0])%201)
	return time.Duration(int64(base) * partsPerThousand / 1000)
}
