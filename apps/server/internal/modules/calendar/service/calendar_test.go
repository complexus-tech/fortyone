package calendar

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeProvider struct {
	authURL   string
	token     ProviderToken
	events    []CoreCalendarEvent
	windows   []CoreBusyWindow
	syncErr   error
	exchanged bool
}

type fakeCalendarTasks struct {
	connectionIDs []uuid.UUID
}

type fakeCalendarUpdates struct {
	workspaceID  uuid.UUID
	userID       uuid.UUID
	connectionID uuid.UUID
	syncedAt     time.Time
}

func (p *fakeCalendarUpdates) PublishCalendarUpdated(_ context.Context, workspaceID, userID, connectionID uuid.UUID, syncedAt time.Time) error {
	p.workspaceID = workspaceID
	p.userID = userID
	p.connectionID = connectionID
	p.syncedAt = syncedAt
	return nil
}

func (q *fakeCalendarTasks) EnqueueCalendarSync(_ context.Context, connectionID uuid.UUID) error {
	q.connectionIDs = append(q.connectionIDs, connectionID)
	return nil
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
	return CalendarSyncSnapshot{Events: p.events, BusyWindows: p.windows}, nil
}

func (p *fakeProvider) SyncCalendarChanges(context.Context, ProviderToken, string) (CalendarSyncDelta, error) {
	return CalendarSyncDelta{}, p.syncErr
}

func (p *fakeProvider) WatchCalendar(context.Context, ProviderToken, CalendarWatchInput) (CalendarWatchChannel, error) {
	return CalendarWatchChannel{}, nil
}

func (p *fakeProvider) StopCalendarWatch(context.Context, ProviderToken, CalendarWatchChannel) error {
	return nil
}

type fakeRepo struct {
	connection       CoreConnection
	upserted         CoreConnectionUpsert
	windows          []CoreBusyWindow
	events           []CoreCalendarEventSummary
	event            CoreCalendarEvent
	blocks           []CoreScheduleBlock
	revoked          uuid.UUID
	member           *bool
	storyAllowed     *bool
	storyUserID      uuid.UUID
	replacements     int
	replaceErr       error
	markSyncedErr    error
	markFailedErr    error
	markedGeneration uuid.UUID
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

func (r *fakeRepo) ApplyCalendarChanges(context.Context, CoreConnection, CalendarSyncDelta) error {
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
	return r.blocks, nil
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
	if redirectURL != "https://app.fortyone.test/acme%20workspace/settings/integrations/calendar?calendar_error=connection_failed" {
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
	if redirectURL != "https://app.fortyone.test/acme/settings/integrations/calendar?connected=1" {
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
	service = New(nil, repo, Config{
		SecretKey: "test-secret",
		Updates:   updates,
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
