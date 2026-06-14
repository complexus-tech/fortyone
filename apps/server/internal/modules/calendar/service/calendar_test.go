package calendar

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fakeProvider struct {
	authURL string
	token   ProviderToken
	windows []CoreBusyWindow
}

func (p *fakeProvider) AuthCodeURL(state string) (string, error) {
	return p.authURL + "?state=" + state, nil
}

func (p *fakeProvider) ExchangeCode(ctx context.Context, code string) (ProviderToken, error) {
	return p.token, nil
}

func (p *fakeProvider) ListBusyWindows(ctx context.Context, token ProviderToken, input BusyWindowInput) ([]CoreBusyWindow, error) {
	return p.windows, nil
}

type fakeRepo struct {
	connection CoreConnection
	upserted   CoreConnectionUpsert
	windows    []CoreBusyWindow
	blocks     []CoreScheduleBlock
	revoked    uuid.UUID
}

func (r *fakeRepo) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error) {
	if r.connection.ID == uuid.Nil {
		return []CoreConnection{}, nil
	}
	return []CoreConnection{r.connection}, nil
}

func (r *fakeRepo) GetConnection(ctx context.Context, workspaceID, connectionID uuid.UUID) (CoreConnection, error) {
	return r.connection, nil
}

func (r *fakeRepo) GetActiveConnection(ctx context.Context, workspaceID, userID uuid.UUID, provider Provider) (CoreConnection, error) {
	return r.connection, nil
}

func (r *fakeRepo) UpsertConnection(ctx context.Context, input CoreConnectionUpsert) (CoreConnection, error) {
	r.upserted = input
	r.connection = CoreConnection{
		ID:             uuid.New(),
		WorkspaceID:    input.WorkspaceID,
		UserID:         input.UserID,
		Provider:       input.Provider,
		ConnectedEmail: input.ConnectedEmail,
		Timezone:       input.Timezone,
		TokenPayload:   input.TokenPayload,
		Scopes:         input.Scopes,
		SyncStatus:     SyncStatusConnected,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	return r.connection, nil
}

func (r *fakeRepo) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	r.revoked = connectionID
	return nil
}

func (r *fakeRepo) ReplaceBusyWindows(ctx context.Context, connection CoreConnection, windows []CoreBusyWindow) error {
	r.windows = windows
	return nil
}

func (r *fakeRepo) MarkConnectionSynced(ctx context.Context, workspaceID, connectionID uuid.UUID, syncedAt time.Time) error {
	r.connection.LastSyncedAt = &syncedAt
	r.connection.SyncStatus = SyncStatusSynced
	return nil
}

func (r *fakeRepo) MarkConnectionSyncFailed(ctx context.Context, workspaceID, connectionID uuid.UUID, message string) error {
	r.connection.SyncStatus = SyncStatusFailed
	r.connection.SyncError = &message
	return nil
}

func (r *fakeRepo) ListBusyWindows(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreBusyWindow, error) {
	return r.windows, nil
}

func (r *fakeRepo) ListScheduleBlocks(ctx context.Context, workspaceID, userID uuid.UUID, startAt, endAt time.Time) ([]CoreScheduleBlock, error) {
	return r.blocks, nil
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

func TestCompleteConnectEncryptsProviderToken(t *testing.T) {
	t.Parallel()

	repo := &fakeRepo{}
	service := New(nil, repo, Config{
		SecretKey:  "test-secret",
		WebsiteURL: "https://app.fortyone.test",
		Providers: map[Provider]CalendarProvider{
			ProviderGoogle: &fakeProvider{
				token: ProviderToken{
					AccessToken:    "access-token",
					RefreshToken:   "refresh-token",
					Expiry:         time.Now().Add(time.Hour),
					ConnectedEmail: "joseph@example.com",
					Timezone:       "Africa/Harare",
					Scopes:         []string{"calendar.freebusy"},
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
		WorkspaceID:   uuid.New(),
		UserID:        uuid.New(),
		WorkspaceSlug: "acme",
		Provider:      ProviderGoogle,
		ExpiresAt:     time.Now().Add(time.Minute).Unix(),
	})
	if err != nil {
		t.Fatalf("signState returned error: %v", err)
	}

	connection, redirectURL, err := service.CompleteConnect(context.Background(), "code", state)
	if err != nil {
		t.Fatalf("CompleteConnect returned error: %v", err)
	}

	if connection.ConnectedEmail != "joseph@example.com" {
		t.Fatalf("unexpected connected email: %s", connection.ConnectedEmail)
	}
	if redirectURL != "https://app.fortyone.test/acme/settings/workspace/integrations/calendar?connected=1" {
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
			ID:           connectionID,
			WorkspaceID:  workspaceID,
			UserID:       userID,
			Provider:     ProviderGoogle,
			TokenPayload: payload,
		},
	}
	service = New(nil, repo, Config{
		SecretKey: "test-secret",
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

func TestCreateScheduleBlockValidatesRangeAndType(t *testing.T) {
	t.Parallel()

	service := New(nil, &fakeRepo{}, Config{SecretKey: "test-secret"})
	workspaceID := uuid.New()
	userID := uuid.New()
	startAt := time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC)

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
