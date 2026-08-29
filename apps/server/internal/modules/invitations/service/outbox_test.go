package invitations

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/events"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type invitationOutboxRepositoryStub struct {
	claimed       []CoreInvitationOutboxEvent
	claimErr      error
	completed     []uuid.UUID
	completeErr   error
	retried       []uuid.UUID
	retryTerminal bool
	retryError    string
}

func (r *invitationOutboxRepositoryStub) ClaimInvitationOutboxEvents(context.Context, int, time.Time, time.Time) ([]CoreInvitationOutboxEvent, error) {
	return r.claimed, r.claimErr
}

func (r *invitationOutboxRepositoryStub) CompleteInvitationOutboxEvent(ctx context.Context, outboxID, _ uuid.UUID, _ time.Time) error {
	r.completeErr = ctx.Err()
	r.completed = append(r.completed, outboxID)
	return nil
}

func (r *invitationOutboxRepositoryStub) RetryInvitationOutboxEvent(_ context.Context, outboxID, _ uuid.UUID, lastError string, _, _ time.Time, terminal bool) error {
	r.retried = append(r.retried, outboxID)
	r.retryTerminal = terminal
	r.retryError = lastError
	return nil
}

type invitationEmailSenderSpy struct {
	deliveries  []InvitationEmailDelivery
	accepted    []events.InvitationAcceptedPayload
	acceptedKey []string
	afterSend   func()
	err         error
}

func (s *invitationEmailSenderSpy) SendInvitationEmail(_ context.Context, delivery InvitationEmailDelivery) error {
	s.deliveries = append(s.deliveries, delivery)
	if s.afterSend != nil {
		s.afterSend()
	}
	return s.err
}

func (s *invitationEmailSenderSpy) SendInvitationAccepted(_ context.Context, payload events.InvitationAcceptedPayload, idempotencyKey string) error {
	s.accepted = append(s.accepted, payload)
	s.acceptedKey = append(s.acceptedKey, idempotencyKey)
	return s.err
}

func TestInvitationOutboxEmailKeepsRawTokenAtDeliveryBoundary(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	manager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	}, bytes.NewReader(bytes.Repeat([]byte{2}, invitationTokenRandomBytes)))
	require.NoError(t, err)
	rawToken, stored, err := manager.Issue()
	require.NoError(t, err)
	payload, err := json.Marshal(InvitationEmailOutboxPayload{
		InviterName:   "Ada",
		Email:         "invitee@example.com",
		Role:          InvitationRoleMember,
		ExpiresAt:     now.Add(7 * 24 * time.Hour),
		WorkspaceID:   uuid.New(),
		WorkspaceName: "Compiler Team",
	})
	require.NoError(t, err)
	require.NotContains(t, string(payload), rawToken)

	outboxID := uuid.New()
	expiresAt := now.Add(7 * 24 * time.Hour)
	repo := &invitationOutboxRepositoryStub{claimed: []CoreInvitationOutboxEvent{{
		ID:                  outboxID,
		InvitationID:        uuid.New(),
		WorkspaceID:         uuid.New(),
		ActorID:             uuid.New(),
		EventType:           string(events.InvitationEmail),
		EventPayload:        payload,
		IdempotencyKey:      "invitation-email:" + outboxID.String(),
		ClaimToken:          uuid.New(),
		AttemptCount:        1,
		CreatedAt:           now,
		StoredToken:         &stored,
		InvitationExpiresAt: &expiresAt,
	}}}
	email := &invitationEmailSenderSpy{}
	dispatcher := NewOutboxDispatcher(nil, repo, manager, email)
	dispatcher.now = func() time.Time { return now }

	require.NoError(t, dispatcher.DispatchReady(context.Background()))
	require.Equal(t, []uuid.UUID{outboxID}, repo.completed)
	require.Empty(t, repo.retried)
	require.Len(t, email.deliveries, 1)
	require.Equal(t, rawToken, email.deliveries[0].Token)
	require.Equal(t, repo.claimed[0].IdempotencyKey, email.deliveries[0].IdempotencyKey)
}

func TestInvitationOutboxCompletesAfterDeliveryContextCancellation(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	manager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	}, bytes.NewReader(bytes.Repeat([]byte{4}, invitationTokenRandomBytes)))
	require.NoError(t, err)
	_, stored, err := manager.Issue()
	require.NoError(t, err)
	payload, err := json.Marshal(InvitationEmailOutboxPayload{ExpiresAt: now.Add(time.Hour)})
	require.NoError(t, err)
	expiresAt := now.Add(time.Hour)
	outboxID := uuid.New()
	repo := &invitationOutboxRepositoryStub{claimed: []CoreInvitationOutboxEvent{{
		ID:                  outboxID,
		EventType:           string(events.InvitationEmail),
		EventPayload:        payload,
		IdempotencyKey:      "invitation-email:" + outboxID.String(),
		ClaimToken:          uuid.New(),
		AttemptCount:        1,
		StoredToken:         &stored,
		InvitationExpiresAt: &expiresAt,
	}}}
	ctx, cancel := context.WithCancel(context.Background())
	email := &invitationEmailSenderSpy{afterSend: cancel}
	dispatcher := NewOutboxDispatcher(nil, repo, manager, email)
	dispatcher.now = func() time.Time { return now }

	require.NoError(t, dispatcher.DispatchReady(ctx))
	require.Equal(t, []uuid.UUID{outboxID}, repo.completed)
	require.NoError(t, repo.completeErr)
}

func TestInvitationOutboxSkipsStaleCredentialAndPublishesAcceptedEvent(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, time.August, 28, 9, 0, 0, 0, time.UTC)
	manager, err := NewInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	})
	require.NoError(t, err)
	usedAt := now.Add(-time.Minute)
	expiresAt := now.Add(time.Hour)
	acceptedPayload, err := json.Marshal(events.InvitationAcceptedPayload{
		InviterEmail: "admin@example.com",
		InviteeEmail: "invitee@example.com",
		WorkspaceID:  uuid.New(),
	})
	require.NoError(t, err)
	emailPayload, err := json.Marshal(InvitationEmailOutboxPayload{})
	require.NoError(t, err)
	repo := &invitationOutboxRepositoryStub{claimed: []CoreInvitationOutboxEvent{{
		ID:                  uuid.New(),
		EventType:           string(events.InvitationEmail),
		EventPayload:        emailPayload,
		ClaimToken:          uuid.New(),
		InvitationExpiresAt: &expiresAt,
		InvitationUsedAt:    &usedAt,
	}, {
		ID:             uuid.New(),
		ActorID:        uuid.New(),
		EventType:      string(events.InvitationAccepted),
		EventPayload:   acceptedPayload,
		IdempotencyKey: "invitation-accepted:test",
		ClaimToken:     uuid.New(),
		CreatedAt:      now,
	}}}
	email := &invitationEmailSenderSpy{}
	dispatcher := NewOutboxDispatcher(nil, repo, manager, email)
	dispatcher.now = func() time.Time { return now }

	require.NoError(t, dispatcher.DispatchReady(context.Background()))
	require.Len(t, repo.completed, 2)
	require.Empty(t, email.deliveries)
	require.Len(t, email.accepted, 1)
	require.Equal(t, "invitation-accepted:test", email.acceptedKey[0])
}

func TestInvitationOutboxRetriesDeliveryFailureWithoutLeakingToken(t *testing.T) {
	t.Parallel()

	now := time.Now().UTC()
	manager, err := newInvitationTokenManager(InvitationTokenConfig{
		Current: InvitationTokenKey{ID: "v1", Secret: invitationTokenTestSecret},
	}, bytes.NewReader(bytes.Repeat([]byte{3}, invitationTokenRandomBytes)))
	require.NoError(t, err)
	rawToken, stored, err := manager.Issue()
	require.NoError(t, err)
	payload, err := json.Marshal(InvitationEmailOutboxPayload{ExpiresAt: now.Add(time.Hour)})
	require.NoError(t, err)
	expiresAt := now.Add(time.Hour)
	outboxID := uuid.New()
	repo := &invitationOutboxRepositoryStub{claimed: []CoreInvitationOutboxEvent{{
		ID:                  outboxID,
		EventType:           string(events.InvitationEmail),
		EventPayload:        payload,
		IdempotencyKey:      "invitation-email:" + outboxID.String(),
		ClaimToken:          uuid.New(),
		AttemptCount:        1,
		StoredToken:         &stored,
		InvitationExpiresAt: &expiresAt,
	}}}
	email := &invitationEmailSenderSpy{err: errors.New("SMTP unavailable")}
	dispatcher := NewOutboxDispatcher(nil, repo, manager, email)
	dispatcher.now = func() time.Time { return now }

	require.NoError(t, dispatcher.DispatchReady(context.Background()))
	require.Empty(t, repo.completed)
	require.Equal(t, []uuid.UUID{outboxID}, repo.retried)
	require.False(t, repo.retryTerminal)
	require.NotContains(t, repo.retryError, rawToken)
}
