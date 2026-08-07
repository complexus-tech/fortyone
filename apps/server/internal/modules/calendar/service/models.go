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

const GoogleCalendarEventsReadonlyScope = "https://www.googleapis.com/auth/calendar.events.readonly"

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
	ID                   uuid.UUID  `json:"id"`
	WorkspaceID          uuid.UUID  `json:"workspaceId"`
	UserID               uuid.UUID  `json:"userId"`
	CredentialGeneration uuid.UUID  `json:"-"`
	ProviderAccountID    string     `json:"-"`
	Provider             Provider   `json:"provider"`
	ConnectedEmail       string     `json:"connectedEmail"`
	Timezone             string     `json:"timezone"`
	TokenPayload         string     `json:"-"`
	Scopes               []string   `json:"scopes"`
	SyncStatus           SyncStatus `json:"syncStatus"`
	SyncError            *string    `json:"syncError,omitempty"`
	LastSyncedAt         *time.Time `json:"lastSyncedAt,omitempty"`
	RevokedAt            *time.Time `json:"revokedAt,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
}

func (connection CoreConnection) CanReadEventDetails() bool {
	if connection.Provider != ProviderGoogle {
		return false
	}
	return hasProviderScope(connection.Scopes, GoogleCalendarEventsReadonlyScope)
}

type CoreSchedule struct {
	StartAt     time.Time           `json:"startAt"`
	EndAt       time.Time           `json:"endAt"`
	BusyWindows []CoreBusyWindow    `json:"busyWindows"`
	Blocks      []CoreScheduleBlock `json:"blocks"`
}

type CoreCalendarView struct {
	StartAt     time.Time                  `json:"startAt"`
	EndAt       time.Time                  `json:"endAt"`
	Events      []CoreCalendarEventSummary `json:"events"`
	BusyWindows []CoreBusyWindow           `json:"busyWindows"`
	Blocks      []CoreScheduleBlock        `json:"blocks"`
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
	HasConflict bool                `json:"hasConflict"`
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
	WorkspaceID       uuid.UUID
	UserID            uuid.UUID
	Provider          Provider
	ProviderAccountID string
	ConnectedEmail    string
	Timezone          string
	TokenPayload      string
	Scopes            []string
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

type CoreCalendarParticipant struct {
	DisplayName    string `json:"displayName,omitempty"`
	Email          string `json:"email,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
	Optional       bool   `json:"optional"`
	Organizer      bool   `json:"organizer"`
	Self           bool   `json:"self"`
}

type CoreCalendarEvent struct {
	ID               uuid.UUID                 `json:"id"`
	ConnectionID     uuid.UUID                 `json:"-"`
	WorkspaceID      uuid.UUID                 `json:"-"`
	UserID           uuid.UUID                 `json:"-"`
	Provider         Provider                  `json:"provider"`
	CalendarID       string                    `json:"calendarId"`
	ProviderEventID  string                    `json:"-"`
	Title            *string                   `json:"title,omitempty"`
	Description      *string                   `json:"description,omitempty"`
	Location         *string                   `json:"location,omitempty"`
	MeetingURL       *string                   `json:"meetingUrl,omitempty"`
	HTMLLink         *string                   `json:"htmlLink,omitempty"`
	Organizer        *CoreCalendarParticipant  `json:"organizer,omitempty"`
	Attendees        []CoreCalendarParticipant `json:"attendees"`
	AttendeesOmitted bool                      `json:"attendeesOmitted"`
	IsAllDay         bool                      `json:"isAllDay"`
	StartDate        *string                   `json:"startDate,omitempty"`
	EndDate          *string                   `json:"endDate,omitempty"`
	StartAt          time.Time                 `json:"startAt"`
	EndAt            time.Time                 `json:"endAt"`
	Visibility       string                    `json:"visibility"`
	IsPrivate        bool                      `json:"isPrivate"`
	BlocksTime       bool                      `json:"-"`
	SourceHash       string                    `json:"-"`
	CreatedAt        time.Time                 `json:"createdAt"`
	UpdatedAt        time.Time                 `json:"updatedAt"`
}

type CoreCalendarEventSummary struct {
	ID              uuid.UUID `json:"id"`
	ConnectionID    uuid.UUID `json:"-"`
	Provider        Provider  `json:"provider"`
	CalendarID      string    `json:"calendarId"`
	ProviderEventID string    `json:"-"`
	Title           *string   `json:"title,omitempty"`
	Location        *string   `json:"location,omitempty"`
	MeetingURL      *string   `json:"meetingUrl,omitempty"`
	HTMLLink        *string   `json:"htmlLink,omitempty"`
	StartAt         time.Time `json:"startAt"`
	EndAt           time.Time `json:"endAt"`
	IsAllDay        bool      `json:"isAllDay"`
	StartDate       *string   `json:"startDate,omitempty"`
	EndDate         *string   `json:"endDate,omitempty"`
	IsPrivate       bool      `json:"isPrivate"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type CalendarSyncSnapshot struct {
	Events              []CoreCalendarEvent
	BusyWindows         []CoreBusyWindow
	CanReadEventDetails bool
}

type ProviderToken struct {
	AccessToken       string    `json:"accessToken"`
	RefreshToken      string    `json:"refreshToken"`
	TokenType         string    `json:"tokenType"`
	Expiry            time.Time `json:"expiry"`
	ProviderAccountID string    `json:"providerAccountId"`
	ConnectedEmail    string    `json:"connectedEmail"`
	Timezone          string    `json:"timezone"`
	Scopes            []string  `json:"scopes"`
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
