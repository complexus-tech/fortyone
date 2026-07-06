package calendarhttp

import (
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

type AppIntegration struct {
	Connections []AppConnection `json:"connections"`
}

type AppConnection struct {
	ID                  uuid.UUID  `json:"id"`
	Provider            string     `json:"provider"`
	ConnectedEmail      string     `json:"connectedEmail"`
	Timezone            string     `json:"timezone"`
	Scopes              []string   `json:"scopes"`
	CanReadEventDetails bool       `json:"canReadEventDetails"`
	SyncStatus          string     `json:"syncStatus"`
	SyncError           *string    `json:"syncError,omitempty"`
	LastSyncedAt        *time.Time `json:"lastSyncedAt,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

type AppCreateConnectSession struct {
	AuthURL string `json:"authUrl"`
}

type AppSchedule struct {
	StartAt     time.Time          `json:"startAt"`
	EndAt       time.Time          `json:"endAt"`
	BusyWindows []AppBusyWindow    `json:"busyWindows"`
	Blocks      []AppScheduleBlock `json:"blocks"`
}

type AppBusyWindow struct {
	ID        uuid.UUID `json:"id"`
	Provider  string    `json:"provider"`
	Title     *string   `json:"title,omitempty"`
	StartAt   time.Time `json:"startAt"`
	EndAt     time.Time `json:"endAt"`
	Status    string    `json:"status"`
	IsPrivate bool      `json:"isPrivate"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type AppScheduleBlock struct {
	ID         uuid.UUID  `json:"id"`
	StoryID    *uuid.UUID `json:"storyId,omitempty"`
	StoryTitle *string    `json:"storyTitle,omitempty"`
	StoryCode  *string    `json:"storyCode,omitempty"`
	TeamID     *uuid.UUID `json:"teamId,omitempty"`
	TeamName   *string    `json:"teamName,omitempty"`
	TeamCode   *string    `json:"teamCode,omitempty"`
	BlockType  string     `json:"blockType"`
	Title      string     `json:"title"`
	StartAt    time.Time  `json:"startAt"`
	EndAt      time.Time  `json:"endAt"`
	IsLocked   bool       `json:"isLocked"`
	Source     string     `json:"source"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

type AppScheduleBlockRequest struct {
	StoryID   *uuid.UUID `json:"storyId"`
	BlockType string     `json:"blockType"`
	Title     string     `json:"title"`
	StartAt   time.Time  `json:"startAt"`
	EndAt     time.Time  `json:"endAt"`
	IsLocked  *bool      `json:"isLocked"`
}

func toAppIntegration(connections []calendar.CoreConnection) AppIntegration {
	return AppIntegration{Connections: toAppConnections(connections)}
}

func toAppConnections(connections []calendar.CoreConnection) []AppConnection {
	out := make([]AppConnection, len(connections))
	for i, connection := range connections {
		out[i] = toAppConnection(connection)
	}
	return out
}

func toAppConnection(connection calendar.CoreConnection) AppConnection {
	return AppConnection{
		ID:                  connection.ID,
		Provider:            string(connection.Provider),
		ConnectedEmail:      connection.ConnectedEmail,
		Timezone:            connection.Timezone,
		Scopes:              connection.Scopes,
		CanReadEventDetails: connection.CanReadEventDetails(),
		SyncStatus:          string(connection.SyncStatus),
		SyncError:           connection.SyncError,
		LastSyncedAt:        connection.LastSyncedAt,
		CreatedAt:           connection.CreatedAt,
		UpdatedAt:           connection.UpdatedAt,
	}
}

func toAppSchedule(schedule calendar.CoreSchedule) AppSchedule {
	return AppSchedule{
		StartAt:     schedule.StartAt,
		EndAt:       schedule.EndAt,
		BusyWindows: toAppBusyWindows(schedule.BusyWindows),
		Blocks:      toAppScheduleBlocks(schedule.Blocks),
	}
}

func toAppBusyWindows(windows []calendar.CoreBusyWindow) []AppBusyWindow {
	out := make([]AppBusyWindow, len(windows))
	for i, window := range windows {
		out[i] = AppBusyWindow{
			ID:        window.ID,
			Provider:  string(window.Provider),
			Title:     window.Title,
			StartAt:   window.StartAt,
			EndAt:     window.EndAt,
			Status:    string(window.Status),
			IsPrivate: window.IsPrivate,
			CreatedAt: window.CreatedAt,
			UpdatedAt: window.UpdatedAt,
		}
	}
	return out
}

func toAppScheduleBlocks(blocks []calendar.CoreScheduleBlock) []AppScheduleBlock {
	out := make([]AppScheduleBlock, len(blocks))
	for i, block := range blocks {
		out[i] = toAppScheduleBlock(block)
	}
	return out
}

func toAppScheduleBlock(block calendar.CoreScheduleBlock) AppScheduleBlock {
	return AppScheduleBlock{
		ID:         block.ID,
		StoryID:    block.StoryID,
		StoryTitle: block.StoryTitle,
		StoryCode:  block.StoryCode,
		TeamID:     block.TeamID,
		TeamName:   block.TeamName,
		TeamCode:   block.TeamCode,
		BlockType:  string(block.BlockType),
		Title:      block.Title,
		StartAt:    block.StartAt,
		EndAt:      block.EndAt,
		IsLocked:   block.IsLocked,
		Source:     string(block.Source),
		CreatedAt:  block.CreatedAt,
		UpdatedAt:  block.UpdatedAt,
	}
}
