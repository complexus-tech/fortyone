package calendar

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

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
