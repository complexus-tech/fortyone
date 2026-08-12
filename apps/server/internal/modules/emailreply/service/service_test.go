package emailreply

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"testing"
	"time"

	messagingrepository "github.com/complexus-tech/projects-api/internal/modules/messaging/repository"
	messaging "github.com/complexus-tech/projects-api/internal/modules/messaging/service"
	"github.com/complexus-tech/projects-api/pkg/tasks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type storeStub struct {
	lookup          messaging.EmailThreadLookup
	lookups         []messaging.EmailThreadLookup
	resolveErr      error
	receipt         messagingrepository.InboundEventRecord
	created         bool
	registerErr     error
	markQueuedErr   error
	recoveryRecords []messagingrepository.InboundEventRecord
	claimErr        error
	releaseErr      error
	lookupInput     messaging.EmailReplyTokenLookup
	input           messagingrepository.InboundEventInput
	markedID        uuid.UUID
	claimProvider   string
	claimLimit      int
	releases        []recoveryRelease
}

type recoveryRelease struct {
	id         uuid.UUID
	generation int
}

func (s *storeStub) FindEmailThreadByReplyToken(_ context.Context, input messaging.EmailReplyTokenLookup) (messaging.EmailThreadLookup, error) {
	s.lookupInput = input
	if len(s.lookups) > 0 {
		lookup := s.lookups[0]
		s.lookups = s.lookups[1:]
		return lookup, s.resolveErr
	}
	return s.lookup, s.resolveErr
}

func (s *storeStub) RegisterInboundEvent(_ context.Context, input messagingrepository.InboundEventInput) (messagingrepository.InboundEventRecord, bool, error) {
	s.input = input
	return s.receipt, s.created, s.registerErr
}

func (s *storeStub) MarkInboundEventQueued(_ context.Context, id uuid.UUID) error {
	s.markedID = id
	return s.markQueuedErr
}

func (s *storeStub) ClaimRecoverableInboundEvents(_ context.Context, provider string, limit int) ([]messagingrepository.InboundEventRecord, error) {
	s.claimProvider = provider
	s.claimLimit = limit
	return s.recoveryRecords, s.claimErr
}

func (s *storeStub) ReleaseInboundEventRecovery(_ context.Context, id uuid.UUID, generation int) error {
	s.releases = append(s.releases, recoveryRelease{id: id, generation: generation})
	return s.releaseErr
}

type queueStub struct {
	payload tasks.BrevoEmailReplyPayload
	err     error
	calls   int
}

func (q *queueStub) EnqueueBrevoEmailReply(_ context.Context, payload tasks.BrevoEmailReplyPayload) error {
	q.calls++
	q.payload = payload
	return q.err
}

func TestIngestPersistsEncryptedItemBeforeQueueHandoff(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	threadID := uuid.New()
	receiptID := uuid.New()
	store := &storeStub{
		lookup: messaging.EmailThreadLookup{
			Thread: messaging.EmailThreadRecord{
				ID:             threadID,
				WorkspaceID:    workspaceID,
				UserID:         userID,
				RecipientEmail: "Joseph+fortyone@example.com",
			},
		},
		receipt: messagingrepository.InboundEventRecord{ID: receiptID, Status: "pending"},
		created: true,
	}
	queue := &queueStub{}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)
	now := time.Date(2026, time.August, 12, 10, 30, 0, 0, time.UTC)
	service.now = func() time.Time { return now }
	raw := inboundWebhookFixture(t, "<message-1@example.com>", "joseph+fortyone@example.com", "Update this objective to on track.")

	result, err := service.Ingest(context.Background(), raw)

	require.NoError(t, err)
	require.Equal(t, IngestResult{Accepted: 1}, result)
	expectedTokenHash := sha256.Sum256([]byte("abcdefghijklmnop"))
	require.Equal(t, Provider, store.lookupInput.Provider)
	require.Equal(t, expectedTokenHash[:], store.lookupInput.TokenHash)
	require.Equal(t, now, store.lookupInput.Now)
	require.Equal(t, Provider, store.input.Provider)
	require.Equal(t, InboundProcessedEvent, store.input.EventType)
	eventScope := workspaceID.String() + ":" + threadID.String()
	require.Equal(t, eventScope, store.input.ExternalWorkspaceID)
	require.Equal(t, "<message-1@example.com>", store.input.ExternalEventID)
	require.NotNil(t, store.input.WorkspaceID)
	require.Equal(t, workspaceID, *store.input.WorkspaceID)
	require.NotContains(t, store.input.PayloadEncrypted, "Update this objective")
	require.Equal(t, tasks.BrevoEmailReplyPayload{
		ExternalWorkspaceID: eventScope,
		EventID:             "<message-1@example.com>",
	}, queue.payload)
	require.Equal(t, receiptID, store.markedID)

	opened, err := service.codec.Open(store.input.PayloadEncrypted)
	require.NoError(t, err)
	var stored storedInboundEmail
	require.NoError(t, json.Unmarshal(opened, &stored))
	require.Equal(t, threadID, stored.ThreadID)
	require.Equal(t, workspaceID, stored.WorkspaceID)
	require.Equal(t, userID, stored.UserID)
	require.Contains(t, string(stored.Email), "Update this objective to on track.")
}

func TestProcessorEmailPayloadDropsUnneededSensitiveProviderFields(t *testing.T) {
	t.Parallel()

	rawText := "full raw body"
	rawHTML := "<p>full raw body</p>"
	signature := "private signature"
	payload, err := processorEmailPayload(InboundEmail{
		MessageID:   "<message@example.com>",
		From:        Mailbox{Address: "joseph@example.com"},
		To:          Mailboxes{{Address: "maya+abcdefghijklmnop@reply.fortyone.app"}},
		Cc:          Mailboxes{{Address: "observer@example.com"}},
		RawTextBody: &rawText, RawHTMLBody: &rawHTML,
		ExtractedMarkdownMessage:   "Current reply",
		ExtractedMarkdownSignature: &signature,
		Attachments: []struct {
			Name          string `json:"Name"`
			ContentType   string `json:"ContentType"`
			ContentLength int64  `json:"ContentLength"`
			ContentID     string `json:"ContentID"`
			DownloadToken string `json:"DownloadToken"`
		}{{Name: "private.pdf", DownloadToken: "download-capability"}},
	})

	require.NoError(t, err)
	require.Contains(t, string(payload), "Current reply")
	require.NotContains(t, string(payload), "full raw body")
	require.NotContains(t, string(payload), "private signature")
	require.NotContains(t, string(payload), "observer@example.com")
	require.NotContains(t, string(payload), "download-capability")
}

func TestIngestIgnoresSenderMismatchBeforePersistence(t *testing.T) {
	t.Parallel()

	store := &storeStub{lookup: messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
		ID:             uuid.New(),
		WorkspaceID:    uuid.New(),
		UserID:         uuid.New(),
		RecipientEmail: "expected@example.com",
	}}}
	queue := &queueStub{}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), inboundWebhookFixture(t, "<message-2@example.com>", "attacker@example.com", "Change the objective."))

	require.NoError(t, err)
	require.Equal(t, IngestResult{Ignored: 1}, result)
	require.Empty(t, store.input.Provider)
	require.Zero(t, queue.calls)
}

func TestIngestAcknowledgesTerminalDuplicateWithoutQueueing(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	store := &storeStub{
		lookup: messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			UserID:         uuid.New(),
			RecipientEmail: "joseph@example.com",
		}},
		receipt: messagingrepository.InboundEventRecord{ID: uuid.New(), Status: "completed"},
	}
	queue := &queueStub{}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), inboundWebhookFixture(t, "<message-3@example.com>", "joseph@example.com", "Already handled."))

	require.NoError(t, err)
	require.Equal(t, IngestResult{Duplicates: 1}, result)
	require.Zero(t, queue.calls)
	require.Equal(t, uuid.Nil, store.markedID)
}

func TestIngestPersistsReceiptWhenQueueHandoffNeedsBrevoRetry(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	store := &storeStub{
		lookup: messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			UserID:         uuid.New(),
			RecipientEmail: "joseph@example.com",
		}},
		receipt: messagingrepository.InboundEventRecord{ID: uuid.New(), Status: "pending"},
		created: true,
	}
	queue := &queueStub{err: errors.New("redis unavailable")}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), inboundWebhookFixture(t, "<message-4@example.com>", "joseph@example.com", "Please update it."))

	require.ErrorIs(t, err, ErrWebhookRetry)
	require.Equal(t, IngestResult{}, result)
	require.Equal(t, Provider, store.input.Provider)
	require.Equal(t, 1, queue.calls)
	require.Equal(t, uuid.Nil, store.markedID)
}

func TestIngestReturnsTransientFailureAfterSkippingPermanentReject(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	store := &storeStub{
		lookup: messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
			ID:             uuid.New(),
			WorkspaceID:    workspaceID,
			UserID:         uuid.New(),
			RecipientEmail: "joseph@example.com",
		}},
		receipt: messagingrepository.InboundEventRecord{ID: uuid.New(), Status: "pending"},
		created: true,
	}
	queue := &queueStub{err: errors.New("redis unavailable")}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	valid := inboundWebhookFixture(t, "<message-after-reject@example.com>", "joseph@example.com", "Please update it.")
	var validPayload InboundWebhook
	require.NoError(t, json.Unmarshal(valid, &validPayload))
	raw, err := json.Marshal(InboundWebhook{Items: []json.RawMessage{
		json.RawMessage(`{}`),
		validPayload.Items[0],
	}})
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), raw)

	require.ErrorIs(t, err, ErrWebhookRetry)
	require.Equal(t, IngestResult{Ignored: 1}, result)
	require.Equal(t, 1, queue.calls)
}

func TestIngestScopesSameMessageIDToEachThread(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	messageID := "<forwarded-message@example.com>"
	raw := inboundWebhookFixture(t, messageID, "joseph@example.com", "Forwarded separately.")

	ingest := func(threadID uuid.UUID) (messagingrepository.InboundEventInput, tasks.BrevoEmailReplyPayload) {
		store := &storeStub{
			lookup: messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
				ID:             threadID,
				WorkspaceID:    workspaceID,
				UserID:         userID,
				RecipientEmail: "joseph@example.com",
			}},
			receipt: messagingrepository.InboundEventRecord{ID: uuid.New(), Status: "pending"},
			created: true,
		}
		queue := &queueStub{}
		service, err := New("application-secret", store, queue)
		require.NoError(t, err)
		_, err = service.Ingest(context.Background(), raw)
		require.NoError(t, err)
		return store.input, queue.payload
	}

	firstInput, firstTask := ingest(uuid.New())
	secondInput, secondTask := ingest(uuid.New())

	require.Equal(t, firstInput.ExternalEventID, secondInput.ExternalEventID)
	require.NotEqual(t, firstInput.ExternalWorkspaceID, secondInput.ExternalWorkspaceID)
	require.Equal(t, firstInput.ExternalWorkspaceID, firstTask.ExternalWorkspaceID)
	require.Equal(t, secondInput.ExternalWorkspaceID, secondTask.ExternalWorkspaceID)
}

func TestIngestProcessesEveryItemInBrevoBatch(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	threadID := uuid.New()
	thread := messaging.EmailThreadLookup{Thread: messaging.EmailThreadRecord{
		ID:             threadID,
		WorkspaceID:    workspaceID,
		UserID:         uuid.New(),
		RecipientEmail: "joseph@example.com",
	}}
	store := &batchStoreStub{
		storeStub: storeStub{lookup: thread, created: true},
	}
	queue := &batchQueueStub{}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	first := inboundWebhookFixture(t, "<batch-1@example.com>", "joseph@example.com", "First reply")
	second := inboundWebhookFixture(t, "<batch-2@example.com>", "joseph@example.com", "Second reply")
	var firstPayload, secondPayload InboundWebhook
	require.NoError(t, json.Unmarshal(first, &firstPayload))
	require.NoError(t, json.Unmarshal(second, &secondPayload))
	raw, err := json.Marshal(InboundWebhook{Items: []json.RawMessage{firstPayload.Items[0], secondPayload.Items[0]}})
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), raw)

	require.NoError(t, err)
	require.Equal(t, IngestResult{Accepted: 2}, result)
	require.Len(t, store.inputs, 2)
	require.Len(t, store.markedIDs, 2)
	require.Len(t, queue.payloads, 2)
	require.Equal(t, "<batch-1@example.com>", store.inputs[0].ExternalEventID)
	require.Equal(t, "<batch-2@example.com>", store.inputs[1].ExternalEventID)
}

func TestIngestSkipsPermanentItemFailuresAndContinuesBatch(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	threadID := uuid.New()
	store := &batchStoreStub{storeStub: storeStub{
		lookups: []messaging.EmailThreadLookup{
			{Thread: messaging.EmailThreadRecord{
				ID:             threadID,
				WorkspaceID:    workspaceID,
				UserID:         uuid.New(),
				RecipientEmail: "expected@example.com",
			}},
			{Thread: messaging.EmailThreadRecord{
				ID:             threadID,
				WorkspaceID:    workspaceID,
				UserID:         uuid.New(),
				RecipientEmail: "joseph@example.com",
			}},
		},
		created: true,
	}}
	queue := &batchQueueStub{}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	unauthorized := inboundWebhookFixture(t, "<batch-unauthorized@example.com>", "attacker@example.com", "Ignore me")
	valid := inboundWebhookFixture(t, "<batch-valid@example.com>", "joseph@example.com", "Process me")
	var unauthorizedPayload, validPayload InboundWebhook
	require.NoError(t, json.Unmarshal(unauthorized, &unauthorizedPayload))
	require.NoError(t, json.Unmarshal(valid, &validPayload))
	raw, err := json.Marshal(InboundWebhook{Items: []json.RawMessage{
		json.RawMessage(`{}`),
		unauthorizedPayload.Items[0],
		validPayload.Items[0],
	}})
	require.NoError(t, err)

	result, err := service.Ingest(context.Background(), raw)

	require.NoError(t, err)
	require.Equal(t, IngestResult{Accepted: 1, Ignored: 2}, result)
	require.Len(t, store.inputs, 1)
	require.Equal(t, "<batch-valid@example.com>", store.inputs[0].ExternalEventID)
	require.Len(t, queue.payloads, 1)
}

func TestRecoverPendingEventsUsesRecoveryGenerationAndReleasesFailedClaims(t *testing.T) {
	t.Parallel()

	firstID := uuid.New()
	secondID := uuid.New()
	store := &storeStub{recoveryRecords: []messagingrepository.InboundEventRecord{
		{
			ID:                  firstID,
			ExternalWorkspaceID: "workspace-1:thread-1",
			ExternalEventID:     "<recovery-1@example.com>",
			RecoveryGeneration:  3,
		},
		{
			ID:                  secondID,
			ExternalWorkspaceID: "workspace-2:thread-2",
			ExternalEventID:     "<recovery-2@example.com>",
			RecoveryGeneration:  4,
		},
	}}
	queue := &recoveryQueueStub{errs: []error{nil, errors.New("redis unavailable")}}
	service, err := New("application-secret", store, queue)
	require.NoError(t, err)

	recovered, err := service.RecoverPendingEvents(context.Background())

	require.Equal(t, 1, recovered)
	require.ErrorContains(t, err, "re-enqueue Brevo email reply <recovery-2@example.com>")
	require.Equal(t, Provider, store.claimProvider)
	require.Equal(t, 500, store.claimLimit)
	require.Equal(t, []tasks.BrevoEmailReplyPayload{
		{
			ExternalWorkspaceID: "workspace-1:thread-1",
			EventID:             "<recovery-1@example.com>",
			RecoveryAttempt:     3,
		},
		{
			ExternalWorkspaceID: "workspace-2:thread-2",
			EventID:             "<recovery-2@example.com>",
			RecoveryAttempt:     4,
		},
	}, queue.payloads)
	require.Equal(t, []recoveryRelease{{id: secondID, generation: 4}}, store.releases)
}

func TestRecoverPendingEventsReturnsClaimFailure(t *testing.T) {
	t.Parallel()

	store := &storeStub{claimErr: errors.New("database unavailable")}
	service, err := New("application-secret", store, &queueStub{})
	require.NoError(t, err)

	recovered, err := service.RecoverPendingEvents(context.Background())

	require.Zero(t, recovered)
	require.ErrorContains(t, err, "claim recoverable Brevo email replies")
}

type batchStoreStub struct {
	storeStub
	inputs    []messagingrepository.InboundEventInput
	markedIDs []uuid.UUID
}

func (s *batchStoreStub) RegisterInboundEvent(_ context.Context, input messagingrepository.InboundEventInput) (messagingrepository.InboundEventRecord, bool, error) {
	s.inputs = append(s.inputs, input)
	return messagingrepository.InboundEventRecord{ID: uuid.New(), Status: "pending"}, true, nil
}

func (s *batchStoreStub) MarkInboundEventQueued(_ context.Context, id uuid.UUID) error {
	s.markedIDs = append(s.markedIDs, id)
	return nil
}

type batchQueueStub struct {
	payloads []tasks.BrevoEmailReplyPayload
}

func (q *batchQueueStub) EnqueueBrevoEmailReply(_ context.Context, payload tasks.BrevoEmailReplyPayload) error {
	q.payloads = append(q.payloads, payload)
	return nil
}

type recoveryQueueStub struct {
	payloads []tasks.BrevoEmailReplyPayload
	errs     []error
}

func (q *recoveryQueueStub) EnqueueBrevoEmailReply(_ context.Context, payload tasks.BrevoEmailReplyPayload) error {
	q.payloads = append(q.payloads, payload)
	if len(q.errs) == 0 {
		return nil
	}
	err := q.errs[0]
	q.errs = q.errs[1:]
	return err
}

func inboundWebhookFixture(t *testing.T, messageID, sender, message string) []byte {
	t.Helper()
	payload := map[string]any{
		"items": []any{
			map[string]any{
				"Uuid":      []string{"brevo-receipt-1"},
				"MessageId": messageID,
				"From": map[string]any{
					"Name":    "Joseph",
					"Address": sender,
				},
				"To": []any{
					map[string]any{"Name": "Maya", "Address": "maya+abcdefghijklmnop@reply.fortyone.app"},
				},
				"Recipients":               []string{"maya+abcdefghijklmnop@reply.fortyone.app"},
				"Subject":                  "Re: Weekly strategy check-in",
				"ExtractedMarkdownMessage": message,
			},
		},
	}
	raw, err := json.Marshal(payload)
	require.NoError(t, err)
	return raw
}
