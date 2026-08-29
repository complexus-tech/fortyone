package domain

import (
	"context"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Provider string

const (
	ProviderGoogle    Provider = "google"
	ProviderMicrosoft Provider = "microsoft"
)

const (
	GoogleCalendarEventsReadonlyScope = "https://www.googleapis.com/auth/calendar.events.readonly"
	GoogleCalendarEventsOwnedScope    = "https://www.googleapis.com/auth/calendar.events.owned"
	GoogleCalendarWriteScopeReason    = "google_calendar_write_scope_required"
	MicrosoftCalendarReadWriteScope   = "Calendars.ReadWrite"
	MicrosoftCalendarWriteScopeReason = "microsoft_calendar_write_scope_required"
)

type SyncStatus string

const (
	SyncStatusConnected SyncStatus = "connected"
	SyncStatusSynced    SyncStatus = "synced"
	SyncStatusFailed    SyncStatus = "failed"
	SyncStatusRevoked   SyncStatus = "revoked"
)

type BusyStatus string

const BusyStatusBusy BusyStatus = "busy"

type BusyTransparency string

const BusyTransparencyOpaque BusyTransparency = "opaque"

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
	ID                     uuid.UUID  `json:"id"`
	WorkspaceID            uuid.UUID  `json:"workspaceId"`
	UserID                 uuid.UUID  `json:"userId"`
	CredentialGeneration   uuid.UUID  `json:"-"`
	ProviderAccountID      string     `json:"-"`
	Provider               Provider   `json:"provider"`
	IsPrimary              bool       `json:"isPrimary"`
	ConnectedEmail         string     `json:"connectedEmail"`
	Timezone               string     `json:"timezone"`
	TokenPayload           string     `json:"-"`
	Scopes                 []string   `json:"scopes"`
	SyncStatus             SyncStatus `json:"syncStatus"`
	SyncError              *string    `json:"syncError,omitempty"`
	LastSyncedAt           *time.Time `json:"lastSyncedAt,omitempty"`
	SyncToken              string     `json:"-"`
	NotificationChannelID  string     `json:"-"`
	NotificationResourceID string     `json:"-"`
	NotificationExpiresAt  *time.Time `json:"-"`
	RevokedAt              *time.Time `json:"revokedAt,omitempty"`
	CreatedAt              time.Time  `json:"createdAt"`
	UpdatedAt              time.Time  `json:"updatedAt"`
}

func (connection CoreConnection) CanReadEventDetails() bool {
	switch connection.Provider {
	case ProviderGoogle:
		return hasProviderScope(connection.Scopes, GoogleCalendarEventsReadonlyScope)
	case ProviderMicrosoft:
		return hasProviderScope(connection.Scopes, MicrosoftCalendarReadWriteScope)
	default:
		return false
	}
}

func (connection CoreConnection) CanWriteEvents() bool {
	switch connection.Provider {
	case ProviderGoogle:
		return connection.CanDeleteOwnedEvents() && hasProviderScope(connection.Scopes, GoogleCalendarEventsReadonlyScope)
	case ProviderMicrosoft:
		return hasProviderScope(connection.Scopes, MicrosoftCalendarReadWriteScope)
	default:
		return false
	}
}

// CanDeleteOwnedEvents is the narrower cleanup capability. Read access is
// required for mirroring, while deleting a known stable event only needs the
// provider's owned-event permission.
func (connection CoreConnection) CanDeleteOwnedEvents() bool {
	switch connection.Provider {
	case ProviderGoogle:
		return hasProviderScope(connection.Scopes, GoogleCalendarEventsOwnedScope)
	case ProviderMicrosoft:
		return hasProviderScope(connection.Scopes, MicrosoftCalendarReadWriteScope)
	default:
		return false
	}
}

func (connection CoreConnection) RequiresReauthorization() bool {
	return (connection.Provider == ProviderGoogle || connection.Provider == ProviderMicrosoft) && !connection.CanWriteEvents()
}

type CoreScheduleRescheduleEvent struct {
	NextStartAt time.Time `json:"nextStartAt"`
	Timezone    string    `json:"timezone"`
	CreatedAt   time.Time `json:"createdAt"`
}

type CoreScheduleIssue struct {
	StoryID                  uuid.UUID `json:"storyId"`
	StoryTitle               string    `json:"storyTitle"`
	StoryCode                string    `json:"storyCode"`
	TeamID                   uuid.UUID `json:"teamId"`
	TeamName                 string    `json:"teamName"`
	TeamCode                 string    `json:"teamCode"`
	EstimatedDurationMinutes *int      `json:"estimatedDurationMinutes"`
	ScheduledDurationMinutes int       `json:"scheduledDurationMinutes"`
	RemainingDurationMinutes int       `json:"remainingDurationMinutes"`
	AutoSchedulingStatus     string    `json:"autoSchedulingStatus"`
	AutoSchedulingReason     *string   `json:"autoSchedulingReason,omitempty"`
	UpdatedAt                time.Time `json:"updatedAt"`
}

type CoreScheduleBlock struct {
	ID                   uuid.UUID           `json:"id"`
	WorkspaceID          uuid.UUID           `json:"workspaceId"`
	UserID               uuid.UUID           `json:"userId"`
	StoryID              *uuid.UUID          `json:"storyId,omitempty"`
	StoryTitle           *string             `json:"storyTitle,omitempty"`
	StoryCode            *string             `json:"storyCode,omitempty"`
	StoryStatusColor     *string             `json:"storyStatusColor,omitempty"`
	TeamID               *uuid.UUID          `json:"teamId,omitempty"`
	TeamName             *string             `json:"teamName,omitempty"`
	TeamCode             *string             `json:"teamCode,omitempty"`
	BlockType            ScheduleBlockType   `json:"blockType"`
	Title                string              `json:"title"`
	StartAt              time.Time           `json:"startAt"`
	EndAt                time.Time           `json:"endAt"`
	CompletedAt          *time.Time          `json:"completedAt,omitempty"`
	HasConflict          bool                `json:"hasConflict"`
	IsLocked             bool                `json:"isLocked"`
	IsCrossWorkspace     bool                `json:"isCrossWorkspace,omitempty"`
	AutoSchedulingStatus *string             `json:"autoSchedulingStatus,omitempty"`
	AutoSchedulingReason *string             `json:"autoSchedulingReason,omitempty"`
	Source               ScheduleBlockSource `json:"source"`
	SegmentIndex         int                 `json:"segmentIndex"`
	ExternalProvider     *Provider           `json:"-"`
	ExternalCalendarID   *string             `json:"-"`
	ExternalEventID      *string             `json:"-"`
	ExternalSyncHash     *string             `json:"-"`
	ExternalSyncedAt     *time.Time          `json:"-"`
	CreatedAt            time.Time           `json:"createdAt"`
	UpdatedAt            time.Time           `json:"updatedAt"`
	ManualOverrideAt     *time.Time          `json:"manualOverrideAt,omitempty"`
	ManualOverrideBy     *uuid.UUID          `json:"manualOverrideBy,omitempty"`
	StoryPriority        string              `json:"-"`
	StoryEndDate         *time.Time          `json:"-"`
}

// CoreSchedule is the transport-neutral availability view consumed by
// scheduling use cases such as Maya. Keeping it in the calendar domain avoids
// coupling consumers to the calendar service implementation.
type CoreSchedule struct {
	StartAt     time.Time           `json:"startAt"`
	EndAt       time.Time           `json:"endAt"`
	Timezone    string              `json:"timezone"`
	BusyWindows []CoreBusyWindow    `json:"busyWindows"`
	Blocks      []CoreScheduleBlock `json:"blocks"`
}

// CoreSchedulePreference describes an inferred working-time preference. It is
// a domain value rather than a calendar service concern so scheduling clients
// can depend on the stable contract directly.
type CoreSchedulePreference struct {
	PreferredStartMinute *int
	SampleCount          int
	Confidence           float64
}

type ManualScheduleBlockChange string

const (
	ManualScheduleBlockChangeMove   ManualScheduleBlockChange = "move"
	ManualScheduleBlockChangeResize ManualScheduleBlockChange = "resize"
)

type ManualScheduleBlockInput struct {
	WorkspaceID       uuid.UUID
	UserID            uuid.UUID
	ActorID           uuid.UUID
	BlockID           uuid.UUID
	StartAt           time.Time
	EndAt             time.Time
	ExpectedUpdatedAt *time.Time
	Timezone          string
	Change            ManualScheduleBlockChange
	ClientMutationID  uuid.UUID
}

type ManualScheduleBlockResult struct {
	Block                    CoreScheduleBlock
	StoryScheduleReconcileID *uuid.UUID
}

type CoreScheduleBlockInput struct {
	ID           uuid.UUID
	WorkspaceID  uuid.UUID
	UserID       uuid.UUID
	StoryID      *uuid.UUID
	BlockType    ScheduleBlockType
	Title        string
	StartAt      time.Time
	EndAt        time.Time
	IsLocked     bool
	Source       ScheduleBlockSource
	SegmentIndex int
}

type MayaScheduleSegmentInput struct {
	SegmentIndex int
	Title        string
	StartAt      time.Time
	EndAt        time.Time
}

type MayaScheduleReconcileInput struct {
	WorkspaceID            uuid.UUID
	UserID                 uuid.UUID
	StoryID                uuid.UUID
	ExpectedStoryUpdatedAt *time.Time
	Segments               []MayaScheduleSegmentInput
	PreemptBlockIDs        []uuid.UUID
	KeepOwnership          bool
	Locked                 bool
	AllowConflicts         bool
}

type ScheduleReconcileAction string

const (
	ScheduleReconcileActionCreated   ScheduleReconcileAction = "created"
	ScheduleReconcileActionUpdated   ScheduleReconcileAction = "updated"
	ScheduleReconcileActionDeleted   ScheduleReconcileAction = "deleted"
	ScheduleReconcileActionUnchanged ScheduleReconcileAction = "unchanged"
)

type CoreScheduleReconcileResult struct {
	Blocks  []CoreScheduleBlock
	Actions []ScheduleReconcileAction
}

type ScheduleEventOperation string

const (
	ScheduleEventOperationUpsert ScheduleEventOperation = "upsert"
	ScheduleEventOperationDelete ScheduleEventOperation = "delete"
)

type ExternalScheduleEventInput struct {
	CalendarID        string
	EventID           string
	BlockID           uuid.UUID
	StoryID           uuid.UUID
	WorkspaceID       uuid.UUID
	Title             string
	StartAt           time.Time
	EndAt             time.Time
	PrivateProperties map[string]string
}

type CoreScheduleEventOutbox struct {
	ID              uuid.UUID
	WorkspaceID     uuid.UUID
	UserID          uuid.UUID
	ScheduleBlockID *uuid.UUID
	Operation       ScheduleEventOperation
	Provider        Provider
	CalendarID      string
	ProviderEventID string
	Payload         json.RawMessage
	DedupeKey       string
	AttemptCount    int
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
	NextSyncToken       string
	Timezone            string
}

type CalendarSyncDelta struct {
	Events                      []CoreCalendarEvent
	BusyWindows                 []CoreBusyWindow
	DeletedEventIDs             []string
	ManagedScheduleEventChanges []ManagedScheduleEventChange
	NextSyncToken               string
}

type ManagedScheduleEventChange struct {
	EventID      string
	Deleted      bool
	Title        string
	StartAt      time.Time
	EndAt        time.Time
	Visibility   string
	Transparency string
	Status       string
	Source       string
	BlockID      string
	StoryID      string
	WorkspaceID  string
	HasAttendees bool
	Recurring    bool
}

type CalendarWatchChannel struct {
	ChannelID  string
	ResourceID string
	ExpiresAt  time.Time
}

type ScheduleEventOutboxStore interface {
	ListPendingScheduleEventOutbox(context.Context, uuid.UUID, Provider, int) ([]CoreScheduleEventOutbox, error)
	ScheduleEventUpsertIsCurrent(context.Context, CoreScheduleEventOutbox, ExternalScheduleEventInput) (bool, error)
	MarkScheduleEventOutboxProcessed(context.Context, CoreScheduleEventOutbox, string) error
	MarkScheduleEventOutboxFailed(context.Context, CoreScheduleEventOutbox, string, bool) error
	ReleaseScheduleEventOutbox(context.Context, []uuid.UUID) error
	DeleteCleanupPendingConnectionIfDrained(context.Context, uuid.UUID, Provider) error
}

func StableGoogleScheduleEventID(blockID uuid.UUID) string {
	digest := sha256.Sum256(blockID[:])
	return "f41sched" + strings.ToLower(base32.HexEncoding.WithPadding(base32.NoPadding).EncodeToString(digest[:20]))
}

func ScheduleEventSyncHash(input ExternalScheduleEventInput) string {
	payload, _ := json.Marshal(input)
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func hasProviderScope(scopes []string, required string) bool {
	for _, scope := range scopes {
		if strings.EqualFold(strings.TrimSpace(scope), strings.TrimSpace(required)) {
			return true
		}
	}
	return false
}
