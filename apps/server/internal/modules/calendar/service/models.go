package calendar

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderGoogle Provider = "google"
)

const googleCalendarEventsReadonlyScope = "https://www.googleapis.com/auth/calendar.events.readonly"

type SyncStatus string

const (
	SyncStatusConnected SyncStatus = "connected"
	SyncStatusSynced    SyncStatus = "synced"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusRevoked   SyncStatus = "revoked"
)

type BusyStatus string

const (
	BusyStatusBusy BusyStatus = "busy"
)

type BusyTransparency string

const (
	BusyTransparencyOpaque BusyTransparency = "opaque"
)

type ScheduleBlockType string

const (
	ScheduleBlockTypeWork  ScheduleBlockType = "work"
	ScheduleBlockTypeFocus ScheduleBlockType = "focus"
)

type ScheduleBlockSource string

const (
	ScheduleBlockSourceUser ScheduleBlockSource = "user"
	ScheduleBlockSourceMaya ScheduleBlockSource = "maya"
)

type CoreConnection struct {
	ID             uuid.UUID  `json:"id"`
	WorkspaceID    uuid.UUID  `json:"workspaceId"`
	UserID         uuid.UUID  `json:"userId"`
	Provider       Provider   `json:"provider"`
	ConnectedEmail string     `json:"connectedEmail"`
	Timezone       string     `json:"timezone"`
	TokenPayload   string     `json:"-"`
	Scopes         []string   `json:"scopes"`
	SyncStatus     SyncStatus `json:"syncStatus"`
	SyncError      *string    `json:"syncError,omitempty"`
	LastSyncedAt   *time.Time `json:"lastSyncedAt,omitempty"`
	RevokedAt      *time.Time `json:"revokedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (connection CoreConnection) CanReadEventDetails() bool {
	if connection.Provider != ProviderGoogle {
		return false
	}
	return hasProviderScope(connection.Scopes, googleCalendarEventsReadonlyScope)
}

type CoreSchedule struct {
	StartAt     time.Time           `json:"startAt"`
	EndAt       time.Time           `json:"endAt"`
	BusyWindows []CoreBusyWindow    `json:"busyWindows"`
	Blocks      []CoreScheduleBlock `json:"blocks"`
}

type CoreScheduleBlock struct {
	ID          uuid.UUID           `json:"id"`
	WorkspaceID uuid.UUID           `json:"workspaceId"`
	UserID      uuid.UUID           `json:"userId"`
	StoryID     *uuid.UUID          `json:"storyId,omitempty"`
	StoryTitle  *string             `json:"storyTitle,omitempty"`
	StoryCode   *string             `json:"storyCode,omitempty"`
	TeamID      *uuid.UUID          `json:"teamId,omitempty"`
	TeamName    *string             `json:"teamName,omitempty"`
	TeamCode    *string             `json:"teamCode,omitempty"`
	BlockType   ScheduleBlockType   `json:"blockType"`
	Title       string              `json:"title"`
	StartAt     time.Time           `json:"startAt"`
	EndAt       time.Time           `json:"endAt"`
	IsLocked    bool                `json:"isLocked"`
	Source      ScheduleBlockSource `json:"source"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type CoreScheduleBlockInput struct {
	ID          uuid.UUID
	WorkspaceID uuid.UUID
	UserID      uuid.UUID
	StoryID     *uuid.UUID
	BlockType   ScheduleBlockType
	Title       string
	StartAt     time.Time
	EndAt       time.Time
	IsLocked    bool
	Source      ScheduleBlockSource
}

type CoreConnectionUpsert struct {
	WorkspaceID    uuid.UUID
	UserID         uuid.UUID
	Provider       Provider
	ConnectedEmail string
	Timezone       string
	TokenPayload   string
	Scopes         []string
}

type CoreConnectSession struct {
	AuthURL string `json:"authUrl"`
}

type CoreBusyWindow struct {
	ID              uuid.UUID        `json:"id"`
	ConnectionID    uuid.UUID        `json:"connectionId"`
	WorkspaceID     uuid.UUID        `json:"workspaceId"`
	UserID          uuid.UUID        `json:"userId"`
	Provider        Provider         `json:"provider"`
	ProviderEventID string           `json:"providerEventId"`
	CalendarID      *string          `json:"calendarId,omitempty"`
	Title           *string          `json:"title,omitempty"`
	StartAt         time.Time        `json:"startAt"`
	EndAt           time.Time        `json:"endAt"`
	Status          BusyStatus       `json:"status"`
	Transparency    BusyTransparency `json:"transparency"`
	IsPrivate       bool             `json:"isPrivate"`
	SourceHash      string           `json:"sourceHash"`
	CreatedAt       time.Time        `json:"createdAt"`
	UpdatedAt       time.Time        `json:"updatedAt"`
}

type ProviderToken struct {
	AccessToken    string    `json:"accessToken"`
	RefreshToken   string    `json:"refreshToken"`
	TokenType      string    `json:"tokenType"`
	Expiry         time.Time `json:"expiry"`
	ConnectedEmail string    `json:"connectedEmail"`
	Timezone       string    `json:"timezone"`
	Scopes         []string  `json:"scopes"`
}

type BusyWindowInput struct {
	ConnectionID uuid.UUID
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	TimeMin      time.Time
	TimeMax      time.Time
	Timezone     string
}

type stateClaims struct {
	WorkspaceID   uuid.UUID `json:"workspaceId"`
	UserID        uuid.UUID `json:"userId"`
	WorkspaceSlug string    `json:"workspaceSlug"`
	Provider      Provider  `json:"provider"`
	ExpiresAt     int64     `json:"expiresAt"`
}

func hasProviderScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if strings.TrimSpace(scope) == required {
			return true
		}
	}
	return false
}
