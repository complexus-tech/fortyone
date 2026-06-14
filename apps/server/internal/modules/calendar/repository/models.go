package calendarrepository

import (
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type dbConnection struct {
	ID             uuid.UUID      `db:"connection_id"`
	WorkspaceID    uuid.UUID      `db:"workspace_id"`
	UserID         uuid.UUID      `db:"user_id"`
	Provider       string         `db:"provider"`
	ConnectedEmail string         `db:"connected_email"`
	Timezone       string         `db:"timezone"`
	TokenPayload   string         `db:"token_payload"`
	Scopes         pq.StringArray `db:"scopes"`
	SyncStatus     string         `db:"sync_status"`
	SyncError      *string        `db:"sync_error"`
	LastSyncedAt   *time.Time     `db:"last_synced_at"`
	RevokedAt      *time.Time     `db:"revoked_at"`
	CreatedAt      time.Time      `db:"created_at"`
	UpdatedAt      time.Time      `db:"updated_at"`
}

type dbBusyWindow struct {
	ID              uuid.UUID `db:"window_id"`
	ConnectionID    uuid.UUID `db:"connection_id"`
	WorkspaceID     uuid.UUID `db:"workspace_id"`
	UserID          uuid.UUID `db:"user_id"`
	Provider        string    `db:"provider"`
	ProviderEventID string    `db:"provider_event_id"`
	CalendarID      *string   `db:"calendar_id"`
	Title           *string   `db:"title"`
	StartAt         time.Time `db:"start_at"`
	EndAt           time.Time `db:"end_at"`
	Status          string    `db:"status"`
	Transparency    string    `db:"transparency"`
	IsPrivate       bool      `db:"is_private"`
	SourceHash      string    `db:"source_hash"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

type dbScheduleBlock struct {
	ID          uuid.UUID  `db:"block_id"`
	WorkspaceID uuid.UUID  `db:"workspace_id"`
	UserID      uuid.UUID  `db:"user_id"`
	StoryID     *uuid.UUID `db:"story_id"`
	StoryTitle  *string    `db:"story_title"`
	StoryCode   *string    `db:"story_code"`
	TeamID      *uuid.UUID `db:"team_id"`
	TeamName    *string    `db:"team_name"`
	TeamCode    *string    `db:"team_code"`
	BlockType   string     `db:"block_type"`
	Title       string     `db:"title"`
	StartAt     time.Time  `db:"start_at"`
	EndAt       time.Time  `db:"end_at"`
	IsLocked    bool       `db:"is_locked"`
	Source      string     `db:"source"`
	CreatedAt   time.Time  `db:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at"`
}

func toCoreConnection(row dbConnection) calendar.CoreConnection {
	return calendar.CoreConnection{
		ID:             row.ID,
		WorkspaceID:    row.WorkspaceID,
		UserID:         row.UserID,
		Provider:       calendar.Provider(row.Provider),
		ConnectedEmail: row.ConnectedEmail,
		Timezone:       row.Timezone,
		TokenPayload:   row.TokenPayload,
		Scopes:         []string(row.Scopes),
		SyncStatus:     calendar.SyncStatus(row.SyncStatus),
		SyncError:      row.SyncError,
		LastSyncedAt:   row.LastSyncedAt,
		RevokedAt:      row.RevokedAt,
		CreatedAt:      row.CreatedAt,
		UpdatedAt:      row.UpdatedAt,
	}
}

func toCoreConnections(rows []dbConnection) []calendar.CoreConnection {
	connections := make([]calendar.CoreConnection, len(rows))
	for i, row := range rows {
		connections[i] = toCoreConnection(row)
	}
	return connections
}

func toCoreBusyWindow(row dbBusyWindow) calendar.CoreBusyWindow {
	return calendar.CoreBusyWindow{
		ID:              row.ID,
		ConnectionID:    row.ConnectionID,
		WorkspaceID:     row.WorkspaceID,
		UserID:          row.UserID,
		Provider:        calendar.Provider(row.Provider),
		ProviderEventID: row.ProviderEventID,
		CalendarID:      row.CalendarID,
		Title:           row.Title,
		StartAt:         row.StartAt,
		EndAt:           row.EndAt,
		Status:          calendar.BusyStatus(row.Status),
		Transparency:    calendar.BusyTransparency(row.Transparency),
		IsPrivate:       row.IsPrivate,
		SourceHash:      row.SourceHash,
		CreatedAt:       row.CreatedAt,
		UpdatedAt:       row.UpdatedAt,
	}
}

func toCoreBusyWindows(rows []dbBusyWindow) []calendar.CoreBusyWindow {
	windows := make([]calendar.CoreBusyWindow, len(rows))
	for i, row := range rows {
		windows[i] = toCoreBusyWindow(row)
	}
	return windows
}

func toCoreScheduleBlock(row dbScheduleBlock) calendar.CoreScheduleBlock {
	return calendar.CoreScheduleBlock{
		ID:          row.ID,
		WorkspaceID: row.WorkspaceID,
		UserID:      row.UserID,
		StoryID:     row.StoryID,
		StoryTitle:  row.StoryTitle,
		StoryCode:   row.StoryCode,
		TeamID:      row.TeamID,
		TeamName:    row.TeamName,
		TeamCode:    row.TeamCode,
		BlockType:   calendar.ScheduleBlockType(row.BlockType),
		Title:       row.Title,
		StartAt:     row.StartAt,
		EndAt:       row.EndAt,
		IsLocked:    row.IsLocked,
		Source:      calendar.ScheduleBlockSource(row.Source),
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

func toCoreScheduleBlocks(rows []dbScheduleBlock) []calendar.CoreScheduleBlock {
	blocks := make([]calendar.CoreScheduleBlock, len(rows))
	for i, row := range rows {
		blocks[i] = toCoreScheduleBlock(row)
	}
	return blocks
}
