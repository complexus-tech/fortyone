package webhooks

import (
	"context"
	"errors"
	"fmt"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
)

type Gateway struct {
	inbox    Inbox
	runtimes RuntimeRegistry
	config   Config
}

func NewGateway(inbox Inbox, runtimes RuntimeRegistry, config Config) (*Gateway, error) {
	if inbox == nil {
		return nil, ErrNotConfigured
	}
	normalized, err := normalizeConfig(config)
	if err != nil {
		return nil, err
	}
	return &Gateway{inbox: inbox, runtimes: runtimes, config: normalized}, nil
}

// Receive verifies and durably records one provider delivery before dispatch.
// The dispatcher receives only the inbox identity; raw payloads and provider
// credentials never enter the queue backend.
func (gateway *Gateway) Receive(
	ctx context.Context,
	provider integrations.ProviderKey,
	request SignedRequest,
) (Receipt, error) {
	if gateway == nil || gateway.inbox == nil {
		return Receipt{}, ErrNotConfigured
	}
	runtime, err := gateway.runtimes.require(provider)
	if err != nil {
		return Receipt{}, err
	}
	if err := validateRequest(request, gateway.config); err != nil {
		return Receipt{}, err
	}

	now := gateway.config.Now().UTC()
	request.ReceivedAt = now
	request.Headers = cloneHeaders(request.Headers)
	signedBody := append([]byte(nil), request.Body...)
	request.Body = append([]byte(nil), signedBody...)
	verified, err := runtime.Verifier.Verify(ctx, request)
	if err != nil {
		classified := verificationError(err)
		if errors.Is(classified, ErrDeliveryIgnored) {
			return Receipt{Ignored: true}, nil
		}
		return Receipt{}, classified
	}

	envelope := Envelope{
		Version:                CurrentEnvelopeVersion,
		Provider:               provider,
		DeliveryID:             verified.DeliveryID,
		EventType:              verified.EventType,
		ExternalAccountID:      verified.ExternalAccountID,
		WorkspaceID:            verified.WorkspaceID,
		InstallationID:         verified.InstallationID,
		InstallationGeneration: verified.InstallationGeneration,
		TraceID:                verified.TraceID,
		ReceivedAt:             now,
	}
	if err := ValidateEnvelope(envelope); err != nil {
		return Receipt{}, err
	}

	encrypted, err := runtime.Protector.Seal(ctx, PayloadBinding{
		Provider:               provider,
		DeliveryID:             envelope.DeliveryID,
		WorkspaceID:            envelope.WorkspaceID,
		InstallationID:         envelope.InstallationID,
		InstallationGeneration: envelope.InstallationGeneration,
	}, signedBody)
	if err != nil {
		if errors.Is(err, context.Canceled) {
			return Receipt{}, context.Canceled
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return Receipt{}, context.DeadlineExceeded
		}
		return Receipt{}, ErrPayloadProtection
	}
	if encrypted == "" {
		return Receipt{}, ErrPayloadProtection
	}

	record, created, err := gateway.inbox.Register(
		ctx,
		envelope,
		encrypted,
		now.Add(gateway.config.PayloadRetention),
	)
	if err != nil {
		return Receipt{}, fmt.Errorf("persist webhook delivery: %w", err)
	}
	if !sameDelivery(record.Envelope, envelope) {
		return Receipt{}, ErrDeliveryConflict
	}
	if !created && record.Status.Terminal() {
		return Receipt{ID: record.ID, Status: record.Status, Created: false}, nil
	}

	if err := runtime.Dispatcher.Enqueue(ctx, Task{InboxID: record.ID, Provider: provider}); err != nil {
		return Receipt{}, ErrDispatchUnavailable
	}
	if err := gateway.inbox.MarkQueued(ctx, record.ID, now); err != nil {
		return Receipt{}, fmt.Errorf("record webhook queue handoff: %w", err)
	}
	return Receipt{
		ID:      record.ID,
		Status:  record.Status,
		Created: created,
		Queued:  true,
	}, nil
}

func sameDelivery(stored, incoming Envelope) bool {
	return stored.Version == incoming.Version &&
		stored.Provider == incoming.Provider &&
		stored.DeliveryID == incoming.DeliveryID &&
		stored.EventType == incoming.EventType &&
		stored.ExternalAccountID == incoming.ExternalAccountID &&
		stored.WorkspaceID == incoming.WorkspaceID &&
		stored.InstallationID == incoming.InstallationID &&
		stored.InstallationGeneration == incoming.InstallationGeneration
}

func cloneHeaders(headers Headers) Headers {
	cloned := make(Headers, len(headers))
	for name, values := range headers {
		cloned[name] = append([]string(nil), values...)
	}
	return cloned
}
