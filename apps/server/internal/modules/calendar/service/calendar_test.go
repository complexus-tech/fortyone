package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"google.golang.org/api/googleapi"
)

type fakeProvider struct {
	authURL          string
	token            ProviderToken
	events           []CoreCalendarEvent
	windows          []CoreBusyWindow
	syncErr          error
	exchanged        bool
	syncChangesCalls int
	syncChangesToken string
	delta            CalendarSyncDelta
	watchInput       CalendarWatchInput
	watchCalls       int
	watchErr         error
	upsertedEvents   []ExternalScheduleEventInput
	deletedEventIDs  []string
	writeErr         error
	timezone         string
}

type fakeCalendarTasks struct {
	connectionIDs   []uuid.UUID
	scheduleUserIDs []uuid.UUID
	scheduleErr     error
}

type fakeCalendarUpdates struct {
	workspaceID  uuid.UUID
	userID       uuid.UUID
	connectionID uuid.UUID
	syncedAt     time.Time
	calls        int
	err          error
}

func (p *fakeCalendarUpdates) PublishCalendarUpdated(_ context.Context, workspaceID, userID, connectionID uuid.UUID, syncedAt time.Time) error {
	p.calls++
	p.workspaceID = workspaceID
	p.userID = userID
	p.connectionID = connectionID
	p.syncedAt = syncedAt
	return p.err
}

func (q *fakeCalendarTasks) EnqueueCalendarSync(_ context.Context, connectionID uuid.UUID) error {
	q.connectionIDs = append(q.connectionIDs, connectionID)
	return nil
}

func (q *fakeCalendarTasks) EnqueueCalendarScheduleReconcile(_ context.Context, userID uuid.UUID) error {
	q.scheduleUserIDs = append(q.scheduleUserIDs, userID)
	return q.scheduleErr
}

func (p *fakeProvider) AuthCodeURL(state string) (string, error) {
	return p.authURL + "?state=" + state, nil
}

func (p *fakeProvider) ExchangeCode(ctx context.Context, code string) (ProviderToken, error) {
	p.exchanged = true
	return p.token, nil
}

func (p *fakeProvider) SyncCalendar(ctx context.Context, token ProviderToken, input BusyWindowInput) (CalendarSyncSnapshot, error) {
	if p.syncErr != nil {
		return CalendarSyncSnapshot{}, p.syncErr
	}
	return CalendarSyncSnapshot{Events: p.events, BusyWindows: p.windows, Timezone: p.timezone}, nil
}

func (p *fakeProvider) SyncCalendarChanges(_ context.Context, _ ProviderToken, syncToken string) (CalendarSyncDelta, error) {
	p.syncChangesCalls++
	p.syncChangesToken = syncToken
	return p.delta, p.syncErr
}

func (p *fakeProvider) WatchCalendar(_ context.Context, _ ProviderToken, input CalendarWatchInput) (CalendarWatchChannel, error) {
	p.watchInput = input
	p.watchCalls++
	if p.watchErr != nil {
		return CalendarWatchChannel{}, p.watchErr
	}
	return CalendarWatchChannel{
		ChannelID:  input.ChannelID,
		ResourceID: "google-resource",
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	}, nil
}

func (p *fakeProvider) StopCalendarWatch(context.Context, ProviderToken, CalendarWatchChannel) error {
	return nil
}

func (p *fakeProvider) UpsertScheduleEvent(_ context.Context, _ ProviderToken, input ExternalScheduleEventInput) error {
	if p.writeErr != nil {
		return p.writeErr
	}
	p.upsertedEvents = append(p.upsertedEvents, input)
	return nil
}

func (p *fakeProvider) DeleteScheduleEvent(_ context.Context, _ ProviderToken, _, eventID string) error {
	if p.writeErr != nil {
		return p.writeErr
	}
	p.deletedEventIDs = append(p.deletedEventIDs, eventID)
	return nil
}

type fakeRepo struct {
	connection                CoreConnection
	upserted                  CoreConnectionUpsert
	windows                   []CoreBusyWindow
	events                    []CoreCalendarEventSummary
	event                     CoreCalendarEvent
	blocks                    []CoreScheduleBlock
	scheduleBlocksCalls       int
	accountScheduleBlockCalls int
	revoked                   uuid.UUID
	member                    *bool
	storyAllowed              *bool
	storyUserID               uuid.UUID
	reconcileResult           CoreScheduleReconcileResult
	replacements              int
	replaceErr                error
	markSyncedErr             error
	markFailedErr             error
	markedGeneration          uuid.UUID
	appliedDelta              CalendarSyncDelta
	pendingOutboxBatches      [][]CoreScheduleEventOutbox
	outboxClaimCalls          int
	processedOutbox           []uuid.UUID
	processedOutboxOperations []ScheduleEventOperation
	failedOutbox              []uuid.UUID
	failedOutboxPermanent     []bool
	releasedOutbox            []uuid.UUID
	dispatchLockCalls         int
	readyOutboxUsers          []uuid.UUID
	cleanupPending            bool
	cleanupFinalizeCalls      int
	upsertCurrent             *bool
	upsertCurrentChecks       int
}

func (r *fakeRepo) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error) {
	if r.connection.ID == uuid.Nil {
		return []CoreConnection{}, nil
	}
	return []CoreConnection{r.connection}, nil
}

func (r *fakeRepo) GetOwnedConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error) {
	return r.connection, nil
}

func (r *fakeRepo) GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error) {
	if r.connection.ID == uuid.Nil {
		return CoreConnection{}, ErrCalendarNotFound
	}
	return r.connection, nil
}

func (r *fakeRepo) GetScheduleEventDispatchConnection(_ context.Context, _ uuid.UUID) (CoreConnection, bool, error) {
	if r.connection.ID == uuid.Nil {
		return CoreConnection{}, false, ErrCalendarNotFound
	}
	return r.connection, r.cleanupPending, nil
}

func (r *fakeRepo) GetConnection(context.Context, uuid.UUID) (CoreConnection, error) {
	return r.connection, nil
}

func (r *fakeRepo) ListConnectionsNeedingWatch(context.Context, time.Time) ([]CoreConnection, error) {
	return nil, nil
}

func (r *fakeRepo) WorkspaceMemberExists(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	if r.member == nil {
		return true, nil
	}
	return *r.member, nil
}

func (r *fakeRepo) UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error) {
	r.upserted = input
	r.connection = CoreConnection{
		ID:                   uuid.New(),
		WorkspaceID:          input.WorkspaceID,
		UserID:               input.UserID,
		CredentialGeneration: uuid.New(),
		ProviderAccountID:    input.ProviderAccountID,
		Provider:             input.Provider,
		ConnectedEmail:       input.ConnectedEmail,
		Timezone:             input.Timezone,
		TokenPayload:         input.TokenPayload,
		Scopes:               input.Scopes,
		SyncStatus:           SyncStatusConnected,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	return r.connection, nil
}

func (r *fakeRepo) BeginConnectionSync(ctx context.Context, connection CoreConnection) (CoreConnection, error) {
	connection.CredentialGeneration = uuid.New()
	r.connection = connection
	return connection, nil
}

func (r *fakeRepo) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	r.revoked = connectionID
	return nil
}

func (r *fakeRepo) ReplaceCalendarSnapshot(ctx context.Context, connection CoreConnection, snapshot CalendarSyncSnapshot) error {
	if r.replaceErr != nil {
		return r.replaceErr
	}
	r.replacements++
	if strings.TrimSpace(snapshot.Timezone) != "" {
		r.connection.Timezone = strings.TrimSpace(snapshot.Timezone)
	}
	r.windows = snapshot.BusyWindows
	r.events = make([]CoreCalendarEventSummary, len(snapshot.Events))
	for i, event := range snapshot.Events {
		r.events[i] = CoreCalendarEventSummary{
			ID:              event.ID,
			ConnectionID:    event.ConnectionID,
			Provider:        event.Provider,
			CalendarID:      event.CalendarID,
			ProviderEventID: event.ProviderEventID,
			Title:           event.Title,
			Location:        event.Location,
			MeetingURL:      event.MeetingURL,
			HTMLLink:        event.HTMLLink,
			StartAt:         event.StartAt,
			EndAt:           event.EndAt,
			IsAllDay:        event.IsAllDay,
			IsPrivate:       event.IsPrivate,
		}
	}
	return nil
}

func (r *fakeRepo) ApplyCalendarChanges(_ context.Context, _ CoreConnection, delta CalendarSyncDelta) error {
	r.appliedDelta = delta
	r.connection.SyncToken = delta.NextSyncToken
	return nil
}

func (r *fakeRepo) SetNotificationChannel(context.Context, CoreConnection, CalendarWatchChannel) error {
	return nil
}

func (r *fakeRepo) ClearNotificationChannel(context.Context, uuid.UUID) error {
	return nil
}

func TestProcessGoogleNotificationEnqueuesAuthenticatedChannelChange(t *testing.T) {
	t.Parallel()

	connectionID := uuid.New()
	channelID := uuid.NewString()
	resourceID := "google-resource"
	tasks := &fakeCalendarTasks{}
	repo := &fakeRepo{connection: CoreConnection{
		ID:                     connectionID,
		NotificationChannelID:  channelID,
		NotificationResourceID: resourceID,
	}}
	service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks})

	err := service.ProcessGoogleNotification(
		context.Background(),
		channelID,
		resourceID,
		"exists",
		service.notificationToken(connectionID, channelID),
	)
	if err != nil {
		t.Fatalf("ProcessGoogleNotification returned error: %v", err)
	}
	if len(tasks.connectionIDs) != 1 || tasks.connectionIDs[0] != connectionID {
		t.Fatalf("expected connection %s to be enqueued, got %v", connectionID, tasks.connectionIDs)
	}
}

func TestProcessGoogleNotificationRejectsInvalidToken(t *testing.T) {
	t.Parallel()

	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret", Tasks: &fakeCalendarTasks{}})
	err := service.ProcessGoogleNotification(context.Background(), uuid.NewString(), "resource", "exists", "invalid")
	if !errors.Is(err, ErrInvalidCalendarNotification) {
		t.Fatalf("expected invalid notification error, got %v", err)
	}
}

func TestSyncConnectionFromNotificationUsesIncrementalChanges(t *testing.T) {
	t.Parallel()

	connectionID := uuid.New()
	provider := &fakeProvider{delta: CalendarSyncDelta{NextSyncToken: "next-sync-token"}}
	repo := &fakeRepo{connection: CoreConnection{
		ID:                   connectionID,
		Provider:             ProviderGoogle,
		TokenPayload:         "encrypted-token",
		SyncToken:            "current-sync-token",
		Scopes:               []string{GoogleCalendarEventsReadonlyScope},
		CredentialGeneration: uuid.New(),
	}}
	service := New(nil, repo, Config{
		SecretKey: "test-secret",
		Providers: map[Provider]CalendarProvider{ProviderGoogle: provider},
	})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{
		AccessToken: "access-token",
	})

	if err := service.SyncConnectionFromNotification(context.Background(), connectionID); err != nil {
		t.Fatalf("SyncConnectionFromNotification returned error: %v", err)
	}
	if provider.syncChangesCalls != 1 {
		t.Fatalf("expected one incremental sync, got %d", provider.syncChangesCalls)
	}
	if provider.syncChangesToken != "current-sync-token" {
		t.Fatalf("expected current sync token, got %q", provider.syncChangesToken)
	}
	if repo.replacements != 0 {
		t.Fatalf("incremental sync must not perform a full snapshot, got %d replacements", repo.replacements)
	}
	if repo.appliedDelta.NextSyncToken != "next-sync-token" {
		t.Fatalf("expected incremental sync token to be applied, got %q", repo.appliedDelta.NextSyncToken)
	}
}

func TestPushSyncEnqueuesReconciliationOnlyAfterRelevantChangesPersist(t *testing.T) {
	t.Parallel()
	connectionID := uuid.New()
	userID := uuid.New()
	provider := &fakeProvider{delta: CalendarSyncDelta{
		BusyWindows:   []CoreBusyWindow{{ProviderEventID: "changed", StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour)}},
		NextSyncToken: "next",
	}}
	repo := &fakeRepo{connection: CoreConnection{
		ID: connectionID, UserID: userID, Provider: ProviderGoogle, SyncToken: "current",
		Scopes: []string{GoogleCalendarEventsReadonlyScope}, CredentialGeneration: uuid.New(),
	}}
	tasks := &fakeCalendarTasks{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks, Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})
	if err := service.SyncConnectionFromNotification(context.Background(), connectionID); err != nil {
		t.Fatalf("SyncConnectionFromNotification returned error: %v", err)
	}
	if len(tasks.scheduleUserIDs) != 1 || tasks.scheduleUserIDs[0] != userID {
		t.Fatalf("expected user reconciliation after successful changed delta: %v", tasks.scheduleUserIDs)
	}
}

func TestPushSyncRetryEnqueuesAfterCommittedDeltaBecomesEmpty(t *testing.T) {
	t.Parallel()
	connectionID := uuid.New()
	userID := uuid.New()
	enqueueErr := errors.New("queue unavailable")
	provider := &fakeProvider{delta: CalendarSyncDelta{
		BusyWindows:   []CoreBusyWindow{{ProviderEventID: "changed", StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour)}},
		NextSyncToken: "committed-next-token",
	}}
	repo := &fakeRepo{connection: CoreConnection{
		ID: connectionID, UserID: userID, Provider: ProviderGoogle, SyncToken: "current-token",
		Scopes: []string{GoogleCalendarEventsReadonlyScope}, CredentialGeneration: uuid.New(),
	}}
	tasks := &fakeCalendarTasks{scheduleErr: enqueueErr}
	service := New(nil, repo, Config{SecretKey: "test-secret", Tasks: tasks, Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})

	if err := service.SyncConnectionFromNotification(context.Background(), connectionID); !errors.Is(err, enqueueErr) {
		t.Fatalf("expected first enqueue failure, got %v", err)
	}
	if repo.connection.SyncToken != "committed-next-token" {
		t.Fatalf("expected provider delta to commit before enqueue failure, got %q", repo.connection.SyncToken)
	}

	provider.delta = CalendarSyncDelta{NextSyncToken: "empty-retry-token"}
	tasks.scheduleErr = nil
	if err := service.SyncConnectionFromNotification(context.Background(), connectionID); err != nil {
		t.Fatalf("empty retry returned error: %v", err)
	}
	if len(tasks.scheduleUserIDs) != 2 || tasks.scheduleUserIDs[1] != userID {
		t.Fatalf("empty retry must preserve the reconciliation handoff: %v", tasks.scheduleUserIDs)
	}
}

func TestConnectionWriteCapabilityRequiresOwnedEventsScope(t *testing.T) {
	t.Parallel()
	legacy := CoreConnection{Provider: ProviderGoogle, Scopes: []string{GoogleCalendarEventsReadonlyScope}}
	if legacy.CanWriteEvents() || !legacy.RequiresReauthorization() {
		t.Fatalf("legacy read-only grant must require reauthorization: %#v", legacy)
	}
	upgraded := CoreConnection{Provider: ProviderGoogle, Scopes: []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope}}
	if !upgraded.CanWriteEvents() || upgraded.RequiresReauthorization() {
		t.Fatalf("owned-events grant must be write-capable: %#v", upgraded)
	}
	partial := CoreConnection{Provider: ProviderGoogle, Scopes: []string{GoogleCalendarEventsOwnedScope}}
	if partial.CanWriteEvents() || !partial.CanDeleteOwnedEvents() || !partial.RequiresReauthorization() {
		t.Fatalf("write-only grant cannot safely filter Maya's mirrored events: %#v", partial)
	}
}

func TestManagedMayaScheduleBlocksRejectGenericMutation(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	repo := &fakeRepo{blocks: []CoreScheduleBlock{{
		ID: blockID, WorkspaceID: workspaceID, UserID: userID, Source: ScheduleBlockSourceMaya,
	}}}
	service := New(nil, repo, Config{})
	startAt := time.Now().UTC().Add(time.Hour)
	_, err := service.UpdateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		ID: blockID, WorkspaceID: workspaceID, UserID: userID, BlockType: ScheduleBlockTypeWork,
		Title: "Detached edit", StartAt: startAt, EndAt: startAt.Add(time.Hour), Source: ScheduleBlockSourceUser,
	})
	if !errors.Is(err, ErrManagedScheduleBlock) {
		t.Fatalf("expected managed-block update rejection, got %v", err)
	}
	if err := service.DeleteScheduleBlock(context.Background(), workspaceID, userID, blockID); !errors.Is(err, ErrManagedScheduleBlock) {
		t.Fatalf("expected managed-block delete rejection, got %v", err)
	}
}

func TestScheduleOutboxWaitsForReauthorizationThenDrains(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	event := ExternalScheduleEventInput{
		CalendarID: "primary", EventID: StableGoogleScheduleEventID(blockID), BlockID: blockID,
		StoryID: uuid.New(), WorkspaceID: workspaceID, Title: "Scheduled work",
		StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour),
	}
	payload, _ := json.Marshal(event)
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
		Operation: ScheduleEventOperationUpsert, Provider: ProviderGoogle, CalendarID: "primary",
		ProviderEventID: event.EventID, Payload: payload,
	}
	repo := &fakeRepo{connection: CoreConnection{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
		Scopes: []string{GoogleCalendarEventsReadonlyScope},
	}, pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}}}
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token", Scopes: []string{GoogleCalendarEventsOwnedScope}})

	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("read-only dispatch returned error: %v", err)
	}
	if repo.outboxClaimCalls != 0 || len(provider.upsertedEvents) != 0 {
		t.Fatalf("read-only connection must leave pending rows unclaimed: claims=%d writes=%d", repo.outboxClaimCalls, len(provider.upsertedEvents))
	}
	repo.connection.Scopes = []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope}
	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("upgraded dispatch returned error: %v", err)
	}
	if len(provider.upsertedEvents) != 1 || len(repo.processedOutbox) != 1 {
		t.Fatalf("expected pending event to deliver after reauthorization: writes=%d processed=%d", len(provider.upsertedEvents), len(repo.processedOutbox))
	}
	if repo.dispatchLockCalls != 2 {
		t.Fatalf("expected both dispatch attempts to use per-user serialization, got %d", repo.dispatchLockCalls)
	}
}

func TestScheduleOutboxDrainsEveryBatch(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	batches := make([][]CoreScheduleEventOutbox, 2)
	batches[0] = make([]CoreScheduleEventOutbox, 100)
	batches[1] = make([]CoreScheduleEventOutbox, 1)
	for batchIndex := range batches {
		for itemIndex := range batches[batchIndex] {
			blockID := uuid.New()
			event := ExternalScheduleEventInput{CalendarID: "primary", EventID: StableGoogleScheduleEventID(blockID), BlockID: blockID, StoryID: uuid.New(), WorkspaceID: workspaceID, StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour)}
			payload, _ := json.Marshal(event)
			batches[batchIndex][itemIndex] = CoreScheduleEventOutbox{
				ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
				Operation: ScheduleEventOperationUpsert, Provider: ProviderGoogle, CalendarID: "primary", ProviderEventID: event.EventID, Payload: payload,
			}
		}
	}
	repo := &fakeRepo{connection: CoreConnection{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
		Scopes: []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope},
	}, pendingOutboxBatches: batches}
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token", Scopes: []string{GoogleCalendarEventsReadonlyScope, GoogleCalendarEventsOwnedScope}})
	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("DispatchScheduleEventOutbox returned error: %v", err)
	}
	if len(provider.upsertedEvents) != 101 || len(repo.processedOutbox) != 101 || repo.outboxClaimCalls != 3 {
		t.Fatalf("expected all batches plus empty continuation: writes=%d processed=%d claims=%d", len(provider.upsertedEvents), len(repo.processedOutbox), repo.outboxClaimCalls)
	}
}

func TestCalendarWriteErrorClassifierKeepsRateLimitsRetryable(t *testing.T) {
	t.Parallel()
	for _, reason := range []string{"rateLimitExceeded", "userRateLimitExceeded", "backendError"} {
		err := &googleapi.Error{Code: 403, Errors: []googleapi.ErrorItem{{Reason: reason}}}
		if isPermanentCalendarWriteError(err) {
			t.Fatalf("expected Google 403 reason %q to remain retryable", reason)
		}
	}
	if !isPermanentCalendarWriteError(&googleapi.Error{Code: 403, Errors: []googleapi.ErrorItem{{Reason: "forbidden"}}}) {
		t.Fatal("expected a non-rate-limit permission failure to be terminal")
	}
	for _, code := range []int{408, 409, 429, 500, 503} {
		if isPermanentCalendarWriteError(&googleapi.Error{Code: code}) {
			t.Fatalf("expected HTTP %d to remain retryable", code)
		}
	}
}

func TestCleanupPendingDispatchDeletesInsteadOfUpsertingAndPurgesCredentials(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	event := ExternalScheduleEventInput{
		CalendarID: "primary", EventID: StableGoogleScheduleEventID(blockID), BlockID: blockID,
		StoryID: uuid.New(), WorkspaceID: workspaceID, Title: "Sensitive focus title",
		StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour),
	}
	payload, _ := json.Marshal(event)
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
		Operation: ScheduleEventOperationUpsert, Provider: ProviderGoogle, CalendarID: "primary",
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
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})
	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("DispatchScheduleEventOutbox returned error: %v", err)
	}
	if len(provider.upsertedEvents) != 0 || len(provider.deletedEventIDs) != 1 || provider.deletedEventIDs[0] != event.EventID {
		t.Fatalf("cleanup-pending credentials must be delete-only: upserts=%#v deletes=%#v", provider.upsertedEvents, provider.deletedEventIDs)
	}
	if len(repo.processedOutbox) != 1 || repo.cleanupFinalizeCalls != 1 {
		t.Fatalf("expected durable delete completion followed by credential purge: processed=%v finalizers=%d", repo.processedOutbox, repo.cleanupFinalizeCalls)
	}
}

func TestCleanupPendingDispatchNeedsOnlyOwnedEventScope(t *testing.T) {
	t.Parallel()
	workspaceID := uuid.New()
	userID := uuid.New()
	blockID := uuid.New()
	eventID := StableGoogleScheduleEventID(blockID)
	item := CoreScheduleEventOutbox{
		ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, ScheduleBlockID: &blockID,
		Operation: ScheduleEventOperationDelete, Provider: ProviderGoogle, CalendarID: "primary", ProviderEventID: eventID,
	}
	repo := &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle,
			Scopes: []string{GoogleCalendarEventsOwnedScope},
		},
		cleanupPending: true, pendingOutboxBatches: [][]CoreScheduleEventOutbox{{item}},
	}
	provider := &fakeProvider{}
	service := New(nil, repo, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderGoogle: provider}})
	repo.connection.TokenPayload, _ = service.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})

	if err := service.DispatchScheduleEventOutbox(context.Background(), userID); err != nil {
		t.Fatalf("owned-only cleanup returned error: %v", err)
	}
	if len(provider.deletedEventIDs) != 1 || provider.deletedEventIDs[0] != eventID || repo.cleanupFinalizeCalls != 1 {
		t.Fatalf("owned-only grant must delete then purge: deletes=%v finalizers=%d", provider.deletedEventIDs, repo.cleanupFinalizeCalls)
	}
}

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

func (r *fakeRepo) MarkConnectionSynced(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	syncedAt time.Time,
) error {
	r.markedGeneration = credentialGeneration
	if r.markSyncedErr != nil {
		return r.markSyncedErr
	}
	r.connection.LastSyncedAt = &syncedAt
	r.connection.SyncStatus = SyncStatusSynced
	return nil
}

func (r *fakeRepo) MarkConnectionSyncFailed(
	ctx context.Context,
	workspaceID, connectionID, credentialGeneration uuid.UUID,
	message string,
) error {
	r.markedGeneration = credentialGeneration
	if r.markFailedErr != nil {
		return r.markFailedErr
	}
	r.connection.SyncStatus = SyncStatusFailed
	r.connection.SyncError = &message
	return nil
}

func (r *fakeRepo) ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error) {
	return r.windows, nil
}

func (r *fakeRepo) ListCalendarEvents(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreCalendarEventSummary, error) {
	return r.events, nil
}

func (r *fakeRepo) GetCalendarEvent(ctx context.Context, workspaceID, userID, eventID uuid.UUID) (CoreCalendarEvent, error) {
	if r.event.ID == uuid.Nil || r.event.ID != eventID {
		return CoreCalendarEvent{}, ErrCalendarEventNotFound
	}
	return r.event, nil
}

func (r *fakeRepo) ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error) {
	r.scheduleBlocksCalls++
	return r.blocks, nil
}

func (r *fakeRepo) GetScheduleBlock(_ context.Context, _, _ uuid.UUID, blockID uuid.UUID) (CoreScheduleBlock, error) {
	for _, block := range r.blocks {
		if block.ID == blockID {
			return block, nil
		}
	}
	return CoreScheduleBlock{}, ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) ScheduleStoryExists(ctx context.Context, workspaceID, userID, storyID uuid.UUID) (bool, error) {
	r.storyUserID = userID
	if r.storyAllowed == nil {
		return true, nil
	}
	return *r.storyAllowed, nil
}

func (r *fakeRepo) CreateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	block := CoreScheduleBlock{
		ID:          uuid.New(),
		WorkspaceID: input.WorkspaceID,
		UserID:      input.UserID,
		StoryID:     input.StoryID,
		BlockType:   input.BlockType,
		Title:       input.Title,
		StartAt:     input.StartAt,
		EndAt:       input.EndAt,
		IsLocked:    input.IsLocked,
		Source:      input.Source,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	r.blocks = append(r.blocks, block)
	return block, nil
}

func (r *fakeRepo) UpdateScheduleBlock(ctx context.Context, input CoreScheduleBlockInput) (CoreScheduleBlock, error) {
	for i := range r.blocks {
		if r.blocks[i].ID == input.ID {
			r.blocks[i].StoryID = input.StoryID
			r.blocks[i].BlockType = input.BlockType
			r.blocks[i].Title = input.Title
			r.blocks[i].StartAt = input.StartAt
			r.blocks[i].EndAt = input.EndAt
			r.blocks[i].IsLocked = input.IsLocked
			r.blocks[i].Source = input.Source
			return r.blocks[i], nil
		}
	}
	return CoreScheduleBlock{}, ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) DeleteScheduleBlock(ctx context.Context, workspaceID, userID, blockID uuid.UUID) error {
	for i := range r.blocks {
		if r.blocks[i].ID == blockID {
			r.blocks = append(r.blocks[:i], r.blocks[i+1:]...)
			return nil
		}
	}
	return ErrCalendarScheduleBlockNotFound
}

func (r *fakeRepo) ListSchedulingBlocksForUser(_ context.Context, _, _ uuid.UUID, _, _ time.Time) ([]CoreScheduleBlock, error) {
	r.accountScheduleBlockCalls++
	return r.blocks, nil
}

func (r *fakeRepo) ListMayaScheduleBlocksForStory(_ context.Context, workspaceID, userID, storyID uuid.UUID) ([]CoreScheduleBlock, error) {
	return r.blocks, nil
}

func (r *fakeRepo) MayaScheduleOwnershipExists(_ context.Context, _, _, _ uuid.UUID) (bool, error) {
	return false, nil
}

func (r *fakeRepo) ReconcileMayaScheduleBlocks(_ context.Context, _ MayaScheduleReconcileInput) (CoreScheduleReconcileResult, error) {
	return r.reconcileResult, nil
}

func (r *fakeRepo) ListReadyScheduleEventOutboxUsers(_ context.Context, _ int) ([]uuid.UUID, error) {
	return append([]uuid.UUID(nil), r.readyOutboxUsers...), nil
}

func (r *fakeRepo) WithScheduleEventDispatchLock(_ context.Context, _ uuid.UUID, dispatch func(ScheduleEventOutboxStore) error) error {
	r.dispatchLockCalls++
	return dispatch(r)
}

func (r *fakeRepo) ListPendingScheduleEventOutbox(_ context.Context, _ uuid.UUID, _ int) ([]CoreScheduleEventOutbox, error) {
	r.outboxClaimCalls++
	if len(r.pendingOutboxBatches) == 0 {
		return nil, nil
	}
	batch := r.pendingOutboxBatches[0]
	r.pendingOutboxBatches = r.pendingOutboxBatches[1:]
	return batch, nil
}

func (r *fakeRepo) ScheduleEventUpsertIsCurrent(_ context.Context, _ CoreScheduleEventOutbox, _ ExternalScheduleEventInput) (bool, error) {
	r.upsertCurrentChecks++
	if r.upsertCurrent == nil {
		return true, nil
	}
	return *r.upsertCurrent, nil
}

func (r *fakeRepo) MarkScheduleEventOutboxProcessed(_ context.Context, item CoreScheduleEventOutbox, _ string) error {
	r.processedOutbox = append(r.processedOutbox, item.ID)
	r.processedOutboxOperations = append(r.processedOutboxOperations, item.Operation)
	return nil
}

func (r *fakeRepo) MarkScheduleEventOutboxFailed(_ context.Context, item CoreScheduleEventOutbox, _ string, permanent bool) error {
	r.failedOutbox = append(r.failedOutbox, item.ID)
	r.failedOutboxPermanent = append(r.failedOutboxPermanent, permanent)
	return nil
}

func (r *fakeRepo) ReleaseScheduleEventOutbox(_ context.Context, outboxIDs []uuid.UUID) error {
	r.releasedOutbox = append(r.releasedOutbox, outboxIDs...)
	return nil
}

func (r *fakeRepo) DeleteCleanupPendingConnectionIfDrained(_ context.Context, _ uuid.UUID) error {
	r.cleanupFinalizeCalls++
	return nil
}

func TestReconcileMayaScheduleBlocksPublishesCalendarInvalidationAfterChange(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	storyID := uuid.New()
	updates := &fakeCalendarUpdates{}
	repo := &fakeRepo{reconcileResult: CoreScheduleReconcileResult{
		Actions: []ScheduleReconcileAction{ScheduleReconcileActionUpdated},
	}}
	service := New(nil, repo, Config{Updates: updates})

	result, err := service.ReconcileMayaScheduleBlocks(context.Background(), MayaScheduleReconcileInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     storyID,
	})
	if err != nil {
		t.Fatalf("ReconcileMayaScheduleBlocks returned error: %v", err)
	}
	if len(result.Actions) != 1 || result.Actions[0] != ScheduleReconcileActionUpdated {
		t.Fatalf("unexpected reconciliation result: %#v", result)
	}
	if updates.calls != 1 || updates.workspaceID != workspaceID || updates.userID != userID || updates.connectionID != uuid.Nil || updates.syncedAt.IsZero() {
		t.Fatalf("changed reconciliation did not publish a scoped calendar invalidation: %#v", updates)
	}
}

func TestReconcileMayaScheduleBlocksSkipsInvalidationForUnchangedBlocks(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{reconcileResult: CoreScheduleReconcileResult{
		Actions: []ScheduleReconcileAction{ScheduleReconcileActionUnchanged},
	}}
	updates := &fakeCalendarUpdates{}
	service := New(nil, repo, Config{Updates: updates})

	_, err := service.ReconcileMayaScheduleBlocks(context.Background(), MayaScheduleReconcileInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		StoryID:     uuid.New(),
	})
	if err != nil {
		t.Fatalf("ReconcileMayaScheduleBlocks returned error: %v", err)
	}
	if updates.calls != 0 {
		t.Fatalf("unchanged reconciliation should not publish calendar invalidation: %#v", updates)
	}
}

func TestReconcileMayaScheduleBlocksDoesNotFailCommittedChangeWhenInvalidationFails(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{reconcileResult: CoreScheduleReconcileResult{
		Actions: []ScheduleReconcileAction{ScheduleReconcileActionCreated},
	}}
	updates := &fakeCalendarUpdates{err: errors.New("redis unavailable")}
	service := New(nil, repo, Config{Updates: updates})

	_, err := service.ReconcileMayaScheduleBlocks(context.Background(), MayaScheduleReconcileInput{
		WorkspaceID: uuid.New(),
		UserID:      uuid.New(),
		StoryID:     uuid.New(),
	})
	if err != nil {
		t.Fatalf("a best-effort invalidation must not roll back a committed schedule: %v", err)
	}
	if updates.calls != 1 {
		t.Fatalf("expected one invalidation attempt, got %d", updates.calls)
	}
}

func TestCreateConnectURLSignsWorkspaceAndUserState(t *testing.T) {
	t.Parallel()

	service := New(nil, &fakeRepo{}, Config{
		SecretKey:  "test-secret",
		WebsiteURL: "https://app.fortyone.test",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{authURL: "https://accounts.google.test/oauth"},
		},
	})
	workspaceID := uuid.New()
	userID := uuid.New()

	session, err := service.CreateConnectSession(context.Background(), workspaceID, userID, "acme")
	if err != nil {
		t.Fatalf("CreateConnectSession returned error: %v", err)
	}

	if !strings.HasPrefix(session.AuthURL, "https://accounts.google.test/oauth?state=") {
		t.Fatalf("unexpected auth url: %s", session.AuthURL)
	}
	state := strings.TrimPrefix(session.AuthURL, "https://accounts.google.test/oauth?state=")
	claims, err := service.verifyState(state)
	if err != nil {
		t.Fatalf("verifyState returned error: %v", err)
	}
	if claims.WorkspaceID != workspaceID || claims.UserID != userID || claims.WorkspaceSlug != "acme" {
		t.Fatalf("state claims mismatch: %#v", claims)
	}
}

func TestCalendarCallbackErrorURLUsesSignedStateAndSafeCode(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	service := New(nil, nil, Config{
		SecretKey:  "test-secret",
		WebsiteURL: "https://app.fortyone.test",
	})
	service.now = func() time.Time { return now }
	userID := uuid.New()
	state, err := service.signState(stateClaims{
		WorkspaceID:   uuid.New(),
		UserID:        userID,
		WorkspaceSlug: "acme workspace",
		Provider:      ProviderGoogle,
		ExpiresAt:     now.Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	redirectURL, err := service.CalendarCallbackErrorURL(state, userID, "raw provider secret")
	if err != nil {
		t.Fatalf("CalendarCallbackErrorURL returned error: %v", err)
	}
	if redirectURL != "https://app.fortyone.test/acme%20workspace/settings/account/calendar?calendar_error=connection_failed" {
		t.Fatalf("unexpected callback error URL: %s", redirectURL)
	}
	if strings.Contains(redirectURL, "raw") || strings.Contains(redirectURL, "secret") {
		t.Fatalf("provider error text leaked into redirect: %s", redirectURL)
	}
	if _, err := service.CalendarCallbackErrorURL(state, uuid.New(), "access_denied"); !errors.Is(err, ErrCalendarAccessDenied) {
		t.Fatalf("expected cross-user error redirect to be denied, got %v", err)
	}
}

func TestCompleteConnectEncryptsProviderToken(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &fakeRepo{}
	service := New(nil, repo, Config{
		SecretKey:  "test-secret",
		WebsiteURL: "https://app.fortyone.test",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{
				token: ProviderToken{
					AccessToken:       "access-token",
					RefreshToken:      "refresh-token",
					Expiry:            time.Now().Add(time.Hour),
					ProviderAccountID: "google-account-1",
					ConnectedEmail:    "joseph@example.com",
					Timezone:          "Africa/Harare",
					Scopes:            []string{"calendar.freebusy"},
				},
				windows: []CoreBusyWindow{
					{
						ProviderEventID: "first-sync-window",
						StartAt:         time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC),
						EndAt:           time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC),
						Status:          BusyStatusBusy,
						Transparency:    BusyTransparencyOpaque,
						SourceHash:      "first-sync-window",
					},
				},
			},
		},
	})
	state, err := service.signState(stateClaims{
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	connection, redirectURL, err := service.CompleteConnect(context.Background(), userID, "code", state)
	if err != nil {
		t.Fatalf("CompleteConnect returned error: %v", err)
	}

	if connection.ConnectedEmail != "joseph@example.com" {
		t.Fatalf("unexpected connected email: %s", connection.ConnectedEmail)
	}
	if redirectURL != "https://app.fortyone.test/acme/settings/account/calendar?connected=1" {
		t.Fatalf("unexpected redirect url: %s", redirectURL)
	}
	if strings.Contains(repo.upserted.TokenPayload, "refresh-token") || strings.Contains(repo.upserted.TokenPayload, "access-token") {
		t.Fatalf("token payload was not encrypted: %s", repo.upserted.TokenPayload)
	}
	if len(repo.windows) != 1 || repo.windows[0].ProviderEventID != "first-sync-window" {
		t.Fatalf("calendar was not synced after connect: %#v", repo.windows)
	}

	token, err := service.decryptTokenPayload(repo.upserted.TokenPayload)
	if err != nil {
		t.Fatalf("decryptTokenPayload returned error: %v", err)
	}
	if token.RefreshToken != "refresh-token" || token.AccessToken != "access-token" {
		t.Fatalf("decrypted token mismatch: %#v", token)
	}
}

func TestCompleteConnectRetainsRefreshTokenForSameGoogleAccount(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	cryptoService := New(nil, nil, Config{SecretKey: "test-secret"})
	previousPayload, err := cryptoService.encryptTokenPayload(ProviderToken{
		AccessToken:       "old-access-token",
		RefreshToken:      "durable-refresh-token",
		ProviderAccountID: "google-account-1",
		ConnectedEmail:    "joseph@example.com",
	})
	if err != nil {
		t.Fatalf("encryptTokenPayload returned error: %v", err)
	}
	repo := &fakeRepo{connection: CoreConnection{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		UserID:            userID,
		Provider:          ProviderGoogle,
		ProviderAccountID: "google-account-1",
		ConnectedEmail:    "Joseph@example.com",
		TokenPayload:      previousPayload,
	}}
	service := New(nil, repo, Config{
		SecretKey:  "test-secret",
		WebsiteURL: "https://app.fortyone.test",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{token: ProviderToken{
				AccessToken:       "new-access-token",
				ProviderAccountID: "google-account-1",
				ConnectedEmail:    "joseph@example.com",
			}},
		},
	})
	state, err := service.signState(stateClaims{
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	if _, _, err := service.CompleteConnect(context.Background(), userID, "code", state); err != nil {
		t.Fatalf("CompleteConnect returned error: %v", err)
	}
	storedToken, err := service.decryptTokenPayload(repo.upserted.TokenPayload)
	if err != nil {
		t.Fatalf("decryptTokenPayload returned error: %v", err)
	}
	if storedToken.AccessToken != "new-access-token" || storedToken.RefreshToken != "durable-refresh-token" {
		t.Fatalf("unexpected retained credentials: %#v", storedToken)
	}
}

func TestCompleteConnectDoesNotRetainRefreshTokenByEmailAlone(t *testing.T) {
	t.Parallel()

	cryptoService := New(nil, nil, Config{SecretKey: "test-secret"})
	previousPayload, err := cryptoService.encryptTokenPayload(ProviderToken{
		RefreshToken:      "former-owner-refresh-token",
		ProviderAccountID: "google-account-1",
		ConnectedEmail:    "shared@example.com",
	})
	if err != nil {
		t.Fatalf("encryptTokenPayload returned error: %v", err)
	}
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &fakeRepo{connection: CoreConnection{
		ID:                uuid.New(),
		WorkspaceID:       workspaceID,
		UserID:            userID,
		Provider:          ProviderGoogle,
		ProviderAccountID: "google-account-1",
		ConnectedEmail:    "shared@example.com",
		TokenPayload:      previousPayload,
	}}
	service := New(nil, repo, Config{SecretKey: "test-secret"})

	_, err = service.withRetainedRefreshToken(
		context.Background(),
		stateClaims{WorkspaceID: workspaceID, UserID: userID, Provider: ProviderGoogle},
		ProviderToken{
			AccessToken:       "new-access-token",
			ProviderAccountID: "google-account-2",
			ConnectedEmail:    "shared@example.com",
		},
	)
	if !errors.Is(err, ErrCalendarCredentialsIncomplete) {
		t.Fatalf("expected credentials incomplete for a different provider subject, got %v", err)
	}
}

func TestSyncConnectionStoresOnlyBusyWindows(t *testing.T) {
	t.Parallel()

	connectionID := uuid.New()
	workspaceID := uuid.New()
	userID := uuid.New()
	service := New(nil, nil, Config{SecretKey: "test-secret"})
	payload, err := service.encryptTokenPayload(ProviderToken{
		AccessToken:    "access-token",
		RefreshToken:   "refresh-token",
		ConnectedEmail: "joseph@example.com",
		Timezone:       "Africa/Harare",
	})
	if err != nil {
		t.Fatalf("encryptTokenPayload returned error: %v", err)
	}

	repo := &fakeRepo{
		connection: CoreConnection{
			ID:                   connectionID,
			WorkspaceID:          workspaceID,
			UserID:               userID,
			CredentialGeneration: uuid.New(),
			Provider:             ProviderGoogle,
			TokenPayload:         payload,
		},
	}
	updates := &fakeCalendarUpdates{}
	tasks := &fakeCalendarTasks{}
	service = New(nil, repo, Config{
		SecretKey: "test-secret",
		Updates:   updates,
		Tasks:     tasks,
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{
				windows: []CoreBusyWindow{
					{
						WorkspaceID:     workspaceID,
						UserID:          userID,
						ProviderEventID: "opaque-event-id",
						StartAt:         time.Date(2026, 6, 13, 9, 0, 0, 0, time.UTC),
						EndAt:           time.Date(2026, 6, 13, 10, 0, 0, 0, time.UTC),
						Status:          BusyStatusBusy,
						Transparency:    BusyTransparencyOpaque,
						SourceHash:      "hash",
					},
				},
			},
		},
	})

	if err := service.SyncConnection(context.Background(), workspaceID, userID, connectionID); err != nil {
		t.Fatalf("SyncConnection returned error: %v", err)
	}

	if len(repo.windows) != 1 {
		t.Fatalf("expected one busy window, got %d", len(repo.windows))
	}
	window := repo.windows[0]
	if window.ProviderEventID != "opaque-event-id" || window.Status != BusyStatusBusy {
		t.Fatalf("unexpected busy window: %#v", window)
	}
	if window.ConnectionID != connectionID || window.WorkspaceID != workspaceID || window.UserID != userID {
		t.Fatalf("busy window was not scoped to connection/workspace/user: %#v", window)
	}
	if repo.markedGeneration != repo.connection.CredentialGeneration {
		t.Fatalf("sync status used the wrong credential generation: got %s want %s", repo.markedGeneration, repo.connection.CredentialGeneration)
	}
	if updates.workspaceID != workspaceID || updates.userID != userID || updates.connectionID != connectionID || updates.syncedAt.IsZero() {
		t.Fatalf("calendar update was not published with the synced connection scope: %#v", updates)
	}
	if len(tasks.scheduleUserIDs) != 1 || tasks.scheduleUserIDs[0] != userID {
		t.Fatalf("manual full sync must enqueue schedule reconciliation: %v", tasks.scheduleUserIDs)
	}
}

func TestListScheduleCombinesBusyWindowsAndScheduleBlocks(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC)
	endAt := startAt.Add(8 * time.Hour)
	blockID := uuid.New()
	repo := &fakeRepo{
		windows: []CoreBusyWindow{
			{
				ID:          uuid.New(),
				WorkspaceID: workspaceID,
				UserID:      userID,
				StartAt:     startAt.Add(time.Hour),
				EndAt:       startAt.Add(2 * time.Hour),
				Status:      BusyStatusBusy,
			},
		},
		blocks: []CoreScheduleBlock{
			{
				ID:          blockID,
				WorkspaceID: workspaceID,
				UserID:      userID,
				BlockType:   ScheduleBlockTypeWork,
				Title:       "Review checkout flow",
				StartAt:     startAt.Add(3 * time.Hour),
				EndAt:       startAt.Add(4 * time.Hour),
				IsLocked:    true,
				Source:      ScheduleBlockSourceUser,
			},
		},
	}
	service := New(nil, repo, Config{SecretKey: "test-secret"})

	schedule, err := service.ListSchedule(context.Background(), workspaceID, userID, startAt, endAt)
	if err != nil {
		t.Fatalf("ListSchedule returned error: %v", err)
	}

	if len(schedule.BusyWindows) != 1 || len(schedule.Blocks) != 1 {
		t.Fatalf("expected one busy window and one block, got %#v", schedule)
	}
	if schedule.Blocks[0].ID != blockID || schedule.Blocks[0].Title != "Review checkout flow" {
		t.Fatalf("unexpected schedule block: %#v", schedule.Blocks[0])
	}
}

func TestCalendarViewKeepsBusyWindowsForBackwardCompatibleAvailability(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	connectionID := uuid.New()
	eventID := uuid.New()
	startAt := time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC)
	endAt := startAt.Add(8 * time.Hour)
	providerEventID := "primary:event-id"
	title := "Customer review"
	repo := &fakeRepo{
		events: []CoreCalendarEventSummary{{
			ID:              eventID,
			ConnectionID:    connectionID,
			Provider:        ProviderGoogle,
			CalendarID:      "primary",
			ProviderEventID: providerEventID,
			Title:           &title,
			StartAt:         startAt.Add(time.Hour),
			EndAt:           startAt.Add(2 * time.Hour),
		}},
		windows: []CoreBusyWindow{{
			ID:              uuid.New(),
			ConnectionID:    connectionID,
			WorkspaceID:     workspaceID,
			UserID:          userID,
			ProviderEventID: providerEventID,
			StartAt:         startAt.Add(time.Hour),
			EndAt:           startAt.Add(2 * time.Hour),
			Status:          BusyStatusBusy,
		}},
	}
	service := New(nil, repo, Config{SecretKey: "test-secret"})

	view, err := service.ListCalendarView(context.Background(), workspaceID, userID, startAt, endAt)
	if err != nil {
		t.Fatalf("ListCalendarView returned error: %v", err)
	}
	if len(view.Events) != 1 || len(view.BusyWindows) != 1 {
		t.Fatalf("expected owner event and backward-compatible busy window: %#v", view)
	}
	if repo.accountScheduleBlockCalls != 1 || repo.scheduleBlocksCalls != 0 {
		t.Fatalf("calendar view must use privacy-redacted account-wide blocks: account=%d workspace=%d", repo.accountScheduleBlockCalls, repo.scheduleBlocksCalls)
	}

	schedule, err := service.ListSchedule(context.Background(), workspaceID, userID, startAt, endAt)
	if err != nil {
		t.Fatalf("ListSchedule returned error: %v", err)
	}
	if len(schedule.BusyWindows) != 1 {
		t.Fatalf("expected availability schedule to retain blocking window: %#v", schedule)
	}
}

func TestFailedSyncPreservesLastCalendarSnapshot(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	connectionID := uuid.New()
	cryptoService := New(nil, nil, Config{SecretKey: "test-secret"})
	payload, err := cryptoService.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})
	if err != nil {
		t.Fatalf("encryptTokenPayload returned error: %v", err)
	}
	repo := &fakeRepo{connection: CoreConnection{
		ID:           connectionID,
		WorkspaceID:  workspaceID,
		UserID:       userID,
		Provider:     ProviderGoogle,
		TokenPayload: payload,
	}}
	providerErr := errors.New("provider unavailable")
	service := New(nil, repo, Config{
		SecretKey: "test-secret",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{syncErr: providerErr},
		},
	})

	err = service.SyncConnection(context.Background(), workspaceID, userID, connectionID)
	if !errors.Is(err, providerErr) {
		t.Fatalf("expected provider failure, got %v", err)
	}
	if repo.replacements != 0 {
		t.Fatal("expected failed provider sync not to replace the last good snapshot")
	}
}

func TestFailedSyncCannotMarkNewerCredentialsFailed(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	connectionID := uuid.New()
	credentialGeneration := uuid.New()
	cryptoService := New(nil, nil, Config{SecretKey: "test-secret"})
	payload, err := cryptoService.encryptTokenPayload(ProviderToken{AccessToken: "access-token"})
	if err != nil {
		t.Fatalf("encryptTokenPayload returned error: %v", err)
	}
	repo := &fakeRepo{
		connection: CoreConnection{
			ID:                   connectionID,
			WorkspaceID:          workspaceID,
			UserID:               userID,
			CredentialGeneration: credentialGeneration,
			Provider:             ProviderGoogle,
			TokenPayload:         payload,
		},
		markFailedErr: ErrCalendarSyncSuperseded,
	}
	service := New(nil, repo, Config{
		SecretKey: "test-secret",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{syncErr: errors.New("old credential failed")},
		},
	})

	err = service.SyncConnection(context.Background(), workspaceID, userID, connectionID)
	if !errors.Is(err, ErrCalendarSyncSuperseded) {
		t.Fatalf("expected superseded sync, got %v", err)
	}
	if repo.markedGeneration == uuid.Nil || repo.markedGeneration == credentialGeneration {
		t.Fatalf("failure status did not use the rotated sync generation: got %s", repo.markedGeneration)
	}
}

func TestCreateScheduleBlockValidatesRangeAndType(t *testing.T) {
	t.Parallel()

	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret"})
	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)
	service.now = func() time.Time { return startAt.Add(-time.Hour) }

	if _, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		BlockType:   ScheduleBlockTypeWork,
		Title:       "Invalid",
		StartAt:     startAt,
		EndAt:       startAt,
		IsLocked:    true,
		Source:      ScheduleBlockSourceUser,
	}); err == nil {
		t.Fatal("expected invalid range error")
	}

	block, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		BlockType:   ScheduleBlockTypeFocus,
		Title:       "Deep work",
		StartAt:     startAt,
		EndAt:       startAt.Add(90 * time.Minute),
		IsLocked:    true,
		Source:      ScheduleBlockSourceUser,
	})
	if err != nil {
		t.Fatalf("CreateScheduleBlock returned error: %v", err)
	}
	if block.Title != "Deep work" || block.BlockType != ScheduleBlockTypeFocus || !block.IsLocked {
		t.Fatalf("unexpected schedule block: %#v", block)
	}
}

func TestCreateScheduleBlockRejectsTimesOutsideSyncedCoverage(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 7, 7, 12, 0, 0, 0, time.UTC)
	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret"})
	service.now = func() time.Time { return now }
	workspaceID := uuid.New()
	userID := uuid.New()

	for _, startAt := range []time.Time{
		now.Add(defaultSyncLookback).Add(-time.Minute),
		now.Add(defaultSyncLookahead).Add(-30 * time.Minute),
	} {
		_, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
			WorkspaceID: workspaceID,
			UserID:      userID,
			BlockType:   ScheduleBlockTypeFocus,
			Title:       "Outside coverage",
			StartAt:     startAt,
			EndAt:       startAt.Add(time.Hour),
		})
		if !errors.Is(err, ErrInvalidScheduleRange) {
			t.Fatalf("expected invalid range for %s, got %v", startAt, err)
		}
	}
}

func TestCreateScheduleBlockRejectsStoryOutsideWorkspace(t *testing.T) {
	t.Parallel()

	allowed := false
	storyID := uuid.New()
	workspaceID := uuid.New()
	userID := uuid.New()
	repo := &fakeRepo{storyAllowed: &allowed}
	service := New(nil, repo, Config{SecretKey: "test-secret"})
	_, err := service.CreateScheduleBlock(context.Background(), CoreScheduleBlockInput{
		WorkspaceID: workspaceID,
		UserID:      userID,
		StoryID:     &storyID,
		BlockType:   ScheduleBlockTypeWork,
		Title:       "Cross-workspace story",
		StartAt:     time.Now().UTC(),
		EndAt:       time.Now().UTC().Add(time.Hour),
	})
	if !errors.Is(err, ErrInvalidScheduleBlock) {
		t.Fatalf("expected invalid schedule block, got %v", err)
	}
	if repo.storyUserID != userID {
		t.Fatalf("story authorization checked the wrong user: got %s want %s", repo.storyUserID, userID)
	}
}
