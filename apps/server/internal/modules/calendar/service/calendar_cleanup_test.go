package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/googleapi"
)

func TestCleanupPendingWithoutOwnedScopeDeadLettersAndPurgesCredentials(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID,
		Operation: ScheduleEventOperationDelete, Provider: ProviderGoogle, CalendarID: "primary", ProviderEventID: "never-created",
	}
	repo := &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
			Scopes: []string{GoogleCalendarEventsReadonlyScope},
		},
		cleanupPending: true, pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}},
	}
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})

	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("impossible cleanup returned error: %v", err)
	}
	if len(provider.deletedEventIDs) != 0 || len(repo.failedOutboxPermanent) != 1 || !repo.failedOutboxPermanent[0] || repo.cleanupFinalizeCalls != 1 {
		t.Fatalf("no-owned cleanup must dead-letter with audit then purge: deletes=%v terminal=%v finalizers=%d", provider.deletedEventIDs, repo.failedOutboxPermanent, repo.cleanupFinalizeCalls)
	}
}

func TestCleanupPendingWithUnreadableCredentialsDeadLettersAndPurges(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID,
		Operation: ScheduleEventOperationDelete, Provider: ProviderGoogle, CalendarID: "primary", ProviderEventID: "cleanup-event",
	}
	repo := &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
			Scopes: []string{GoogleCalendarEventsOwnedScope}, TokenPayload: "not-encrypted",
		},
		cleanupPending: true, pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}},
	}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: &fakeProvider{}}})

	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("unreadable cleanup credentials returned error: %v", err)
	}
	if len(repo.failedOutboxPermanent) != 1 || !repo.failedOutboxPermanent[0] || repo.cleanupFinalizeCalls != 1 {
		t.Fatalf("unreadable cleanup credentials must terminally audit then purge: terminal=%v finalizers=%d", repo.failedOutboxPermanent, repo.cleanupFinalizeCalls)
	}
}

func TestStaleScheduleUpsertDeletesProviderEventInsteadOfPublishing(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	event := ExternalScheduleEventInput{
		CalendarID: "primary", EventID: StableGoogleScheduleEventID(blockID), BlockID: blockID,
		StoryID: uuid.New(), WorkspaceID: workspaceID, Title: "Stale owner",
		StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour),
	}
	payload, _ := json.Marshal(event)
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
		Operation: ScheduleEventOperationUpsert, Provider: ProviderGoogle, CalendarID: "primary",
		ProviderEventID: event.EventID, Payload: payload,
	}
	current := false
	repo := &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
			Scopes: []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope},
		},
		upsertCurrent: &current, pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}},
	}
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})

	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("stale upsert dispatch returned error: %v", err)
	}
	if repo.upsertCurrentChecks != 1 || len(provider.upsertedEvents) != 0 || len(provider.deletedEventIDs) != 1 {
		t.Fatalf("stale provider state must be deleted, checks=%d upserts=%v deletes=%v", repo.upsertCurrentChecks, provider.upsertedEvents, provider.deletedEventIDs)
	}
	if len(repo.processedOutboxOperations) != 1 || repo.processedOutboxOperations[0] != ScheduleEventOperationDelete {
		t.Fatalf("stale upsert completion must not mark a block mirrored: %v", repo.processedOutboxOperations)
	}
}

func TestTerminalCleanupFailureDeadLettersBeforeCredentialPurge(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	event := ExternalScheduleEventInput{
		CalendarID: "primary", EventID: StableGoogleScheduleEventID(blockID), BlockID: blockID,
		StoryID: uuid.New(), WorkspaceID: workspaceID,
	}
	payload, _ := json.Marshal(event)
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
		Operation: ScheduleEventOperationDelete, Provider: ProviderGoogle, CalendarID: "primary",
		ProviderEventID: event.EventID, Payload: payload, AttemptCount: 1,
	}
	repo := &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
			Scopes: []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope},
		},
		cleanupPending:       true,
		pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}},
	}
	providerErr := &googleapi.Error{Code: 403, Errors: []googleapi.ErrorItem{{Reason: "forbidden"}}}
	provider := &fakeProvider{writeErr: providerErr}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})
	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); !errors.Is(err, providerErr) {
		t.Fatalf("expected provider failure, got %v", err)
	}
	if len(repo.failedOutboxPermanent) != 1 || !repo.failedOutboxPermanent[0] || repo.cleanupFinalizeCalls != 1 {
		t.Fatalf("terminal teardown failure must be dead-lettered before credentials are purged: terminal=%v finalizers=%d", repo.failedOutboxPermanent, repo.cleanupFinalizeCalls)
	}
}

func TestCompleteConnectRejectsRemovedWorkspaceMember(t *testing.T) {
	t.Parallel()

	allowed := false
	repo := &fakeRepo{member: &allowed}
	provider := &fakeProvider{token: ProviderToken{ConnectedEmail: "joseph@example.com"}}
	userID := uuid.New()
	service := New(nil, repo, Config{
		SecretKey: "test-secret",
		Providers: map[Provider]CalendarProvider{ProviderGoogle: provider},
	})
	state, err := service.signState(stateClaims{
		WorkspaceID:   uuid.New(),
		UserID:        userID,
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	_, _, err = service.CompleteConnect(context.Background(), userID, "code", state)
	if !errors.Is(err, ErrCalendarAccessDenied) {
		t.Fatalf("expected access denied, got %v", err)
	}
	if provider.exchanged {
		t.Fatal("expected membership to be checked before exchanging provider credentials")
	}
}

func TestCompleteConnectRejectsOAuthStateFromAnotherUser(t *testing.T) {
	t.Parallel()

	stateUserID := uuid.New()
	provider := &fakeProvider{token: ProviderToken{
		AccessToken:    "access-token",
		RefreshToken:   "refresh-token",
		ConnectedEmail: "victim@example.com",
	}}
	service := New(nil, &fakeRepo{}, Config{
		SecretKey: "test-secret",
		Providers: map[Provider]CalendarProvider{ProviderGoogle: provider},
	})
	state, err := service.signState(stateClaims{
		WorkspaceID:   uuid.New(),
		UserID:        stateUserID,
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	_, _, err = service.CompleteConnect(context.Background(), uuid.New(), "code", state)
	if !errors.Is(err, ErrCalendarAccessDenied) {
		t.Fatalf("expected cross-user callback to be denied, got %v", err)
	}
	if provider.exchanged {
		t.Fatal("provider code must not be exchanged for a cross-user callback")
	}
}

func TestCompleteConnectSynchronizesAndStartsNotificationChannel(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	provider := &fakeProvider{token: ProviderToken{
		AccessToken:       "access-token",
		RefreshToken:      "refresh-token",
		ProviderAccountID: "google-account",
		ConnectedEmail:    "person@example.com",
	}}
	repo := &fakeRepo{}
	service := New(nil, repo, Config{
		SecretKey:  "test-secret",
		WebhookURL: "https://api.example.com/webhooks/google/calendar",
		Providers:  map[Provider]CalendarProvider{ProviderGoogle: provider},
	})
	state, err := service.signState(stateClaims{
		WorkspaceID:   uuid.New(),
		UserID:        userID,
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	_, _, err = service.CompleteConnect(context.Background(), userID, "code", state)
	if err != nil {
		t.Fatalf("CompleteConnect returned error: %v", err)
	}
	if repo.replacements != 1 {
		t.Fatalf("expected one initial calendar sync, got %d", repo.replacements)
	}
	if provider.watchCalls != 1 {
		t.Fatalf("expected one notification channel, got %d", provider.watchCalls)
	}
	if provider.watchInput.Address != "https://api.example.com/webhooks/google/calendar" {
		t.Fatalf("unexpected webhook address %q", provider.watchInput.Address)
	}
	if provider.watchInput.TTL != googleWatchTTL {
		t.Fatalf("expected a seven-day watch TTL, got %s", provider.watchInput.TTL)
	}
	if provider.watchInput.Token == "" {
		t.Fatal("expected a signed notification token")
	}
}
