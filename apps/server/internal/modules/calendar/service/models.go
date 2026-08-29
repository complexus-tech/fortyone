package calendar

import (
	"context"
	"strings"
	"time"

	calendardomain "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	"github.com/google/uuid"
)

// Compatibility aliases keep the existing service and HTTP contracts stable
// while persistence depends only on transport-neutral domain values.
type Provider = calendardomain.Provider

const (
	ProviderGoogle    = calendardomain.ProviderGoogle
	ProviderMicrosoft = calendardomain.ProviderMicrosoft
)

const (
	GoogleCalendarEventsReadonlyScope = calendardomain.GoogleCalendarEventsReadonlyScope
	GoogleCalendarEventsOwnedScope    = calendardomain.GoogleCalendarEventsOwnedScope
	GoogleCalendarWriteScopeReason    = calendardomain.GoogleCalendarWriteScopeReason
	MicrosoftCalendarReadWriteScope   = calendardomain.MicrosoftCalendarReadWriteScope
	MicrosoftCalendarWriteScopeReason = calendardomain.MicrosoftCalendarWriteScopeReason
)

type SyncStatus = calendardomain.SyncStatus

const (
	SyncStatusConnected = calendardomain.SyncStatusConnected
	SyncStatusSynced    = calendardomain.SyncStatusSynced
	SyncStatusFailed    = calendardomain.SyncStatusFailed
	SyncStatusRevoked   = calendardomain.SyncStatusRevoked
)

type BusyStatus = calendardomain.BusyStatus

const BusyStatusBusy = calendardomain.BusyStatusBusy

type BusyTransparency = calendardomain.BusyTransparency

const BusyTransparencyOpaque = calendardomain.BusyTransparencyOpaque

type ScheduleBlockType = calendardomain.ScheduleBlockType

const (
	ScheduleBlockTypeWork  = calendardomain.ScheduleBlockTypeWork
	ScheduleBlockTypeFocus = calendardomain.ScheduleBlockTypeFocus
)

type ScheduleBlockSource = calendardomain.ScheduleBlockSource

const (
	ScheduleBlockSourceUser = calendardomain.ScheduleBlockSourceUser
	ScheduleBlockSourceMaya = calendardomain.ScheduleBlockSourceMaya
)

type CoreConnection = calendardomain.CoreConnection
type CoreScheduleRescheduleEvent = calendardomain.CoreScheduleRescheduleEvent
type CoreScheduleIssue = calendardomain.CoreScheduleIssue
type CoreScheduleBlock = calendardomain.CoreScheduleBlock
type ManualScheduleBlockChange = calendardomain.ManualScheduleBlockChange

const (
	ManualScheduleBlockChangeMove   = calendardomain.ManualScheduleBlockChangeMove
	ManualScheduleBlockChangeResize = calendardomain.ManualScheduleBlockChangeResize
)

type ManualScheduleBlockInput = calendardomain.ManualScheduleBlockInput
type ManualScheduleBlockResult = calendardomain.ManualScheduleBlockResult
type CoreScheduleBlockInput = calendardomain.CoreScheduleBlockInput
type MayaScheduleSegmentInput = calendardomain.MayaScheduleSegmentInput
type MayaScheduleReconcileInput = calendardomain.MayaScheduleReconcileInput
type ScheduleReconcileAction = calendardomain.ScheduleReconcileAction

const (
	ScheduleReconcileActionCreated   = calendardomain.ScheduleReconcileActionCreated
	ScheduleReconcileActionUpdated   = calendardomain.ScheduleReconcileActionUpdated
	ScheduleReconcileActionDeleted   = calendardomain.ScheduleReconcileActionDeleted
	ScheduleReconcileActionUnchanged = calendardomain.ScheduleReconcileActionUnchanged
)

type CoreScheduleReconcileResult = calendardomain.CoreScheduleReconcileResult
type ScheduleEventOperation = calendardomain.ScheduleEventOperation

const (
	ScheduleEventOperationUpsert = calendardomain.ScheduleEventOperationUpsert
	ScheduleEventOperationDelete = calendardomain.ScheduleEventOperationDelete
)

type ExternalScheduleEventInput = calendardomain.ExternalScheduleEventInput
type CoreScheduleEventOutbox = calendardomain.CoreScheduleEventOutbox
type CoreConnectionUpsert = calendardomain.CoreConnectionUpsert
type CoreBusyWindow = calendardomain.CoreBusyWindow
type CoreCalendarParticipant = calendardomain.CoreCalendarParticipant
type CoreCalendarEvent = calendardomain.CoreCalendarEvent
type CoreCalendarEventSummary = calendardomain.CoreCalendarEventSummary
type CalendarSyncSnapshot = calendardomain.CalendarSyncSnapshot
type CalendarSyncDelta = calendardomain.CalendarSyncDelta
type ManagedScheduleEventChange = calendardomain.ManagedScheduleEventChange
type CalendarWatchChannel = calendardomain.CalendarWatchChannel

var StableGoogleScheduleEventID = calendardomain.StableGoogleScheduleEventID
var ScheduleEventSyncHash = calendardomain.ScheduleEventSyncHash

type CoreSchedule = calendardomain.CoreSchedule
type CoreSchedulePreference = calendardomain.CoreSchedulePreference

type CoreCalendarView struct {
	StartAt        time.Time                  `json:"startAt"`
	EndAt          time.Time                  `json:"endAt"`
	Events         []CoreCalendarEventSummary `json:"events"`
	BusyWindows    []CoreBusyWindow           `json:"busyWindows"`
	Blocks         []CoreScheduleBlock        `json:"blocks"`
	ScheduleIssues []CoreScheduleIssue        `json:"scheduleIssues"`
}

type ManualScheduleBlockRepository interface {
	ManuallyRescheduleScheduleBlock(context.Context, ManualScheduleBlockInput) (ManualScheduleBlockResult, error)
}

type ScheduleFeedbackRepository interface {
	ListManualScheduleRescheduleEvents(context.Context, uuid.UUID, uuid.UUID, time.Time) ([]CoreScheduleRescheduleEvent, error)
}

type CoreConnectSession struct {
	AuthURL string `json:"authUrl"`
}

type ExternalScheduleEventResult struct {
	EventID string
}

type CalendarWatchInput struct {
	ChannelID string
	Address   string
	Token     string
	TTL       time.Duration
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
		if strings.EqualFold(strings.TrimSpace(scope), strings.TrimSpace(required)) {
			return true
		}
	}
	return false
}
