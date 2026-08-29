package calendar

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

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
	if redirectURL != "https://app.fortyone.test/acme%20workspace/settings/account/calendar?calendar_error=connection_failed&calendar_provider=google" {
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
	if redirectURL != "https://app.fortyone.test/acme/settings/account/calendar?connected=1&calendar_provider=google" {
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
