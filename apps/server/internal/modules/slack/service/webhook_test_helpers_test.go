package slack

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/platform/integrations"
	"github.com/complexus-tech/projects-api/internal/platform/webhooks"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
)

func (s *eventStoreStub) GetByExternalKey(
	_ context.Context,
	provider integrations.ProviderKey,
	externalAccountID, deliveryID string,
) (webhooks.Record, error) {
	if provider != slackWebhookProvider {
		return webhooks.Record{}, errors.New("unexpected provider")
	}
	if externalAccountID != "T1" {
		return webhooks.Record{}, errors.New("unexpected external workspace")
	}
	if s.getInboundErr != nil {
		return webhooks.Record{}, s.getInboundErr
	}
	record, ok := s.inboundRecords[deliveryID]
	if !ok {
		return webhooks.Record{}, webhooks.ErrNotFound
	}
	return normalizeTestWebhookRecord(record, externalAccountID, deliveryID), nil
}

func (s *eventStoreStub) GetByID(_ context.Context, id uuid.UUID) (webhooks.Record, error) {
	if s.getInboundErr != nil {
		return webhooks.Record{}, s.getInboundErr
	}
	for eventID, record := range s.inboundRecords {
		record = normalizeTestWebhookRecord(record, "T1", eventID)
		if record.ID == id {
			return record, nil
		}
	}
	return webhooks.Record{}, webhooks.ErrNotFound
}

func (s *eventStoreStub) Start(
	ctx context.Context,
	id uuid.UUID,
	_ time.Time,
	_ time.Duration,
) (webhooks.Record, bool, error) {
	record, err := s.GetByID(ctx, id)
	if err != nil {
		return webhooks.Record{}, false, err
	}
	s.startedEventIDs = append(s.startedEventIDs, record.DeliveryID)
	if s.inboundErr != nil {
		return record, false, s.inboundErr
	}
	if !s.processInbound {
		return record, false, nil
	}
	record.Status = webhooks.StatusProcessing
	record.AttemptCount++
	s.inboundRecords[record.DeliveryID] = record
	return record, true, nil
}

func (s *eventStoreStub) Complete(
	_ context.Context,
	id uuid.UUID,
	status webhooks.Status,
	outcomeCode string,
	_ time.Time,
) error {
	s.completions = append(s.completions, inboundCompletion{id: id, status: string(status), message: outcomeCode})
	for eventID, record := range s.inboundRecords {
		record = normalizeTestWebhookRecord(record, "T1", eventID)
		if record.ID == id {
			record.Status = status
			s.inboundRecords[eventID] = record
			break
		}
	}
	return nil
}

func normalizeTestWebhookRecord(record webhooks.Record, externalAccountID, deliveryID string) webhooks.Record {
	if record.ID == uuid.Nil {
		record.ID = testInboundReceiptID
	}
	if record.Provider == "" {
		record.Provider = slackWebhookProvider
	}
	if record.ExternalAccountID == "" {
		record.ExternalAccountID = externalAccountID
	}
	if record.DeliveryID == "" {
		record.DeliveryID = deliveryID
	}
	if record.Status == "" {
		record.Status = webhooks.StatusPending
	}
	return record
}

type eventQueueStub struct {
	payloads []tasks.SlackEventPayload
	requests []webhooks.SignedRequest
	errors   []error
}

func (q *eventQueueStub) Receive(_ context.Context, provider integrations.ProviderKey, request webhooks.SignedRequest) (webhooks.Receipt, error) {
	inboxID := uuid.New()
	q.requests = append(q.requests, request)
	q.payloads = append(q.payloads, tasks.SlackEventPayload{Provider: string(provider), InboxID: inboxID})
	index := len(q.payloads) - 1
	if index < len(q.errors) && q.errors[index] != nil {
		return webhooks.Receipt{}, q.errors[index]
	}
	return webhooks.Receipt{ID: inboxID, Status: webhooks.StatusPending, Created: true, Queued: true}, nil
}

func (q *eventQueueStub) EnqueueSlackEvent(_ context.Context, payload tasks.SlackEventPayload) error {
	q.payloads = append(q.payloads, payload)
	index := len(q.payloads) - 1
	if index < len(q.errors) {
		return q.errors[index]
	}
	return nil
}

type webhookRecoveryStub struct {
	report   webhooks.RecoveryReport
	err      error
	provider integrations.ProviderKey
	policy   webhooks.RecoveryPolicy
}

func (recovery *webhookRecoveryStub) Recover(
	_ context.Context,
	provider integrations.ProviderKey,
	policy webhooks.RecoveryPolicy,
) (webhooks.RecoveryReport, error) {
	recovery.provider = provider
	recovery.policy = policy
	return recovery.report, recovery.err
}

func processSlackRaw(t *testing.T, processor *EventProcessor, rawBody []byte) error {
	t.Helper()
	store, ok := processor.webhookInbox.(*eventStoreStub)
	if !ok {
		t.Fatalf("test webhook inbox = %T, want *eventStoreStub", processor.webhookInbox)
	}
	envelope, err := decodeSlackEvent(rawBody)
	if err != nil {
		return err
	}
	if _, exists := store.inboundRecords[envelope.EventID]; !exists {
		generation := testInstallGeneration
		if store.installGeneration != nil {
			generation = *store.installGeneration
		}
		record := webhooks.Record{
			ID: testInboundReceiptID,
			Envelope: webhooks.Envelope{
				Version:                webhooks.CurrentEnvelopeVersion,
				Provider:               slackWebhookProvider,
				DeliveryID:             envelope.EventID,
				EventType:              envelope.Event.Type,
				ExternalAccountID:      envelope.TeamID,
				WorkspaceID:            testWorkspaceID,
				InstallationID:         testSlackWorkspaceID,
				InstallationGeneration: generation,
				ReceivedAt:             time.Now().UTC(),
			},
			Status: webhooks.StatusPending,
		}
		encrypted, sealErr := processor.webhookPayloads.Seal(
			context.Background(),
			slackWebhookPayloadBinding(record),
			rawBody,
		)
		if sealErr != nil {
			t.Fatalf("seal test Slack webhook: %v", sealErr)
		}
		record.EncryptedPayload = &encrypted
		store.inboundRecords[envelope.EventID] = record
	}
	return processor.Process(context.Background(), rawBody)
}
