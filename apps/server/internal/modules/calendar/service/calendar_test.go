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
	connectionIDs     []uuid.UUID
	scheduleUserIDs   []uuid.UUID
	storyWorkspaceIDs []uuid.UUID
	storyIDs          []uuid.UUID
	scheduleErr       error
	storyScheduleErr  error
}

type fakeCalendarUpdates struct {
	workspaceID  uuid.UUID
	userID       uuid.UUID
	connectionID uuid.UUID
	syncedAt     time.Time
	calls        int
	err          error
}

func TestWorkspaceCalendarURLUsesWorkspaceRouting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		websiteURL string
		want       string
	}{
		{
			name:       "hosted subdomain",
			websiteURL: "https://cloud.fortyone.app",
			want:       "https://acme.fortyone.app/settings/account/calendar?connected=1&calendar_provider=microsoft",
		},
		{
			name:       "local path",
			websiteURL: "http://localhost:3000",
			want:       "http://localhost:3000/acme/settings/account/calendar?connected=1&calendar_provider=microsoft",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			service := &Service{cfg: Config{WebsiteURL: test.websiteURL}}
			got := service.workspaceCalendarURL("acme", "connected=1&calendar_provider=microsoft")
			if got != test.want {
				t.Fatalf("workspaceCalendarURL() = %q, want %q", got, test.want)
			}
		})
	}
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

func (q *fakeCalendarTasks) EnqueueStoryScheduleReconcile(_ context.Context, workspaceID, storyID uuid.UUID) error {
	q.storyWorkspaceIDs = append(q.storyWorkspaceIDs, workspaceID)
	q.storyIDs = append(q.storyIDs, storyID)
	return q.storyScheduleErr
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

func (p *fakeProvider) UpsertScheduleEvent(_ context.Context, _ ProviderToken, input ExternalScheduleEventInput) (ExternalScheduleEventResult, error) {
	if p.writeErr != nil {
		return ExternalScheduleEventResult{}, p.writeErr
	}
	p.upsertedEvents = append(p.upsertedEvents, input)
	return ExternalScheduleEventResult{EventID: input.EventID}, nil
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
	manualRescheduleInputs    []ManualScheduleBlockInput
	manualRescheduleResult    ManualScheduleBlockResult
	manualRescheduleErr       error
	primaryConnectionID       uuid.UUID
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

func (r *fakeRepo) SetPrimaryConnection(_ context.Context, _, _ uuid.UUID, connectionID uuid.UUID) (CoreConnection, error) {
	r.primaryConnectionID = connectionID
	r.connection.IsPrimary = true
	return r.connection, nil
}

func TestSetPrimaryConnectionUsesAccountOwnedConnection(t *testing.T) {
	t.Parallel()

	connectionID := uuid.New()
	repo := &fakeRepo{connection: CoreConnection{ID: connectionID, Provider: ProviderMicrosoft}}
	service := New(nil, repo, Config{})

	connection, err := service.SetPrimaryConnection(context.Background(), uuid.New(), uuid.New(), connectionID)
	if err != nil {
		t.Fatalf("set primary calendar connection: %v", err)
	}
	if repo.primaryConnectionID != connectionID {
		t.Fatalf("expected primary connection %s, got %s", connectionID, repo.primaryConnectionID)
	}
	if !connection.IsPrimary {
		t.Fatal("expected selected connection to be primary")
	}
}

func (r *fakeRepo) UpdateConnectionToken(_ context.Context, connection CoreConnection, tokenPayload string) error {
	r.connection = connection
	r.connection.TokenPayload = tokenPayload
	return nil
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
