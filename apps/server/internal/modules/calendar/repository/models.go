package calendarrepository

import (
	"encoding/json"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type dbConnection struct {
	ID                     uuid.UUID      `db:"connection_id"`
	WorkspaceID            uuid.UUID      `db:"workspace_id"`
	UserID                 uuid.UUID      `db:"user_id"`
	CredentialGeneration   uuid.UUID      `db:"credential_generation"`
	ProviderAccountID      string         `db:"provider_account_id"`
	Provider               string         `db:"provider"`
	ConnectedEmail         string         `db:"connected_email"`
	Timezone               string         `db:"timezone"`
	TokenPayload           string         `db:"token_payload"`
	Scopes                 pq.StringArray `db:"scopes"`
	SyncStatus             string         `db:"sync_status"`
	SyncError              *string        `db:"sync_error"`
	LastSyncedAt           *time.Time     `db:"last_synced_at"`
	SyncToken              *string        `db:"sync_token"`
	NotificationChannelID  *string        `db:"notification_channel_id"`
	NotificationResourceID *string        `db:"notification_resource_id"`
	NotificationExpiresAt  *time.Time     `db:"notification_expires_at"`
	RevokedAt              *time.Time     `db:"revoked_at"`
	CreatedAt              time.Time      `db:"created_at"`
	UpdatedAt              time.Time      `db:"updated_at"`
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

type dbCalendarEvent struct {
	ID               uuid.UUID  `db:"event_id"`
	ConnectionID     uuid.UUID  `db:"connection_id"`
	WorkspaceID      uuid.UUID  `db:"workspace_id"`
	UserID           uuid.UUID  `db:"user_id"`
	Provider         string     `db:"provider"`
	CalendarID       string     `db:"calendar_id"`
	ProviderEventID  string     `db:"provider_event_id"`
	Title            *string    `db:"title"`
	Description      *string    `db:"description"`
	Location         *string    `db:"location"`
	MeetingURL       *string    `db:"meeting_url"`
	HTMLLink         *string    `db:"html_link"`
	Organizer        []byte     `db:"organizer"`
	Attendees        []byte     `db:"attendees"`
	AttendeesOmitted bool       `db:"attendees_omitted"`
	IsAllDay         bool       `db:"is_all_day"`
	StartDate        *time.Time `db:"start_date"`
	EndDate          *time.Time `db:"end_date"`
	StartAt          time.Time  `db:"start_at"`
	EndAt            time.Time  `db:"end_at"`
	Visibility       string     `db:"visibility"`
	IsPrivate        bool       `db:"is_private"`
	SourceHash       string     `db:"source_hash"`
	CreatedAt        time.Time  `db:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at"`
}

type dbCalendarEventSummary struct {
	ID              uuid.UUID  `db:"event_id"`
	ConnectionID    uuid.UUID  `db:"connection_id"`
	Provider        string     `db:"provider"`
	CalendarID      string     `db:"calendar_id"`
	ProviderEventID string     `db:"provider_event_id"`
	Title           *string    `db:"title"`
	Location        *string    `db:"location"`
	MeetingURL      *string    `db:"meeting_url"`
	HTMLLink        *string    `db:"html_link"`
	StartAt         time.Time  `db:"start_at"`
	EndAt           time.Time  `db:"end_at"`
	IsAllDay        bool       `db:"is_all_day"`
	StartDate       *time.Time `db:"start_date"`
	EndDate         *time.Time `db:"end_date"`
	IsPrivate       bool       `db:"is_private"`
	CreatedAt       time.Time  `db:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type dbScheduleBlock struct {
	ID                   uuid.UUID  `db:"block_id"`
	WorkspaceID          uuid.UUID  `db:"workspace_id"`
	UserID               uuid.UUID  `db:"user_id"`
	StoryID              *uuid.UUID `db:"story_id"`
	StoryTitle           *string    `db:"story_title"`
	StoryCode            *string    `db:"story_code"`
	StoryPriority        string     `db:"story_priority"`
	StoryEndDate         *time.Time `db:"story_end_date"`
	TeamID               *uuid.UUID `db:"team_id"`
	TeamName             *string    `db:"team_name"`
	TeamCode             *string    `db:"team_code"`
	BlockType            string     `db:"block_type"`
	Title                string     `db:"title"`
	StartAt              time.Time  `db:"start_at"`
	EndAt                time.Time  `db:"end_at"`
	HasConflict          bool       `db:"has_conflict"`
	IsLocked             bool       `db:"is_locked"`
	AutoSchedulingStatus *string    `db:"auto_scheduling_status"`
	AutoSchedulingReason *string    `db:"auto_scheduling_reason"`
	Source               string     `db:"source"`
	SegmentIndex         int        `db:"segment_index"`
	ExternalProvider     *string    `db:"external_provider"`
	ExternalCalendarID   *string    `db:"external_calendar_id"`
	ExternalEventID      *string    `db:"external_event_id"`
	ExternalSyncHash     *string    `db:"external_sync_hash"`
	ExternalSyncedAt     *time.Time `db:"external_synced_at"`
	CreatedAt            time.Time  `db:"created_at"`
	UpdatedAt            time.Time  `db:"updated_at"`
	ManualOverrideAt     *time.Time `db:"manual_override_at"`
	ManualOverrideBy     *uuid.UUID `db:"manual_override_by"`
}

func toCoreConnection(row dbConnection) calendar.CoreConnection {
	return calendar.CoreConnection{
		ID:                     row.ID,
		WorkspaceID:            row.WorkspaceID,
		UserID:                 row.UserID,
		CredentialGeneration:   row.CredentialGeneration,
		ProviderAccountID:      row.ProviderAccountID,
		Provider:               calendar.Provider(row.Provider),
		ConnectedEmail:         row.ConnectedEmail,
		Timezone:               row.Timezone,
		TokenPayload:           row.TokenPayload,
		Scopes:                 []string(row.Scopes),
		SyncStatus:             calendar.SyncStatus(row.SyncStatus),
		SyncError:              row.SyncError,
		LastSyncedAt:           row.LastSyncedAt,
		SyncToken:              valueOrEmpty(row.SyncToken),
		NotificationChannelID:  valueOrEmpty(row.NotificationChannelID),
		NotificationResourceID: valueOrEmpty(row.NotificationResourceID),
		NotificationExpiresAt:  row.NotificationExpiresAt,
		RevokedAt:              row.RevokedAt,
		CreatedAt:              row.CreatedAt,
		UpdatedAt:              row.UpdatedAt,
	}
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
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
		Title:           nil,
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

func toCoreCalendarEvent(row dbCalendarEvent) (calendar.CoreCalendarEvent, error) {
	attendees := []calendar.CoreCalendarParticipant{}
	if len(row.Attendees) > 0 {
		if err := json.Unmarshal(row.Attendees, &attendees); err != nil {
			return calendar.CoreCalendarEvent{}, fmt.Errorf("decode calendar event attendees: %w", err)
		}
	}
	var organizer *calendar.CoreCalendarParticipant
	if len(row.Organizer) > 0 && string(row.Organizer) != "null" {
		var value calendar.CoreCalendarParticipant
		if err := json.Unmarshal(row.Organizer, &value); err != nil {
			return calendar.CoreCalendarEvent{}, fmt.Errorf("decode calendar event organizer: %w", err)
		}
		organizer = &value
	}
	return calendar.CoreCalendarEvent{
		ID:               row.ID,
		ConnectionID:     row.ConnectionID,
		WorkspaceID:      row.WorkspaceID,
		UserID:           row.UserID,
		Provider:         calendar.Provider(row.Provider),
		CalendarID:       row.CalendarID,
		ProviderEventID:  row.ProviderEventID,
		Title:            row.Title,
		Description:      row.Description,
		Location:         row.Location,
		MeetingURL:       row.MeetingURL,
		HTMLLink:         row.HTMLLink,
		Organizer:        organizer,
		Attendees:        attendees,
		AttendeesOmitted: row.AttendeesOmitted,
		IsAllDay:         row.IsAllDay,
		StartDate:        calendarDateString(row.StartDate),
		EndDate:          calendarDateString(row.EndDate),
		StartAt:          row.StartAt,
		EndAt:            row.EndAt,
		Visibility:       row.Visibility,
		IsPrivate:        row.IsPrivate,
		SourceHash:       row.SourceHash,
		CreatedAt:        row.CreatedAt,
		UpdatedAt:        row.UpdatedAt,
	}, nil
}

func toCoreCalendarEventSummaries(rows []dbCalendarEventSummary) []calendar.CoreCalendarEventSummary {
	events := make([]calendar.CoreCalendarEventSummary, len(rows))
	for i, row := range rows {
		events[i] = calendar.CoreCalendarEventSummary{
			ID:              row.ID,
			ConnectionID:    row.ConnectionID,
			Provider:        calendar.Provider(row.Provider),
			CalendarID:      row.CalendarID,
			ProviderEventID: row.ProviderEventID,
			Title:           row.Title,
			Location:        row.Location,
			MeetingURL:      row.MeetingURL,
			HTMLLink:        row.HTMLLink,
			StartAt:         row.StartAt,
			EndAt:           row.EndAt,
			IsAllDay:        row.IsAllDay,
			StartDate:       calendarDateString(row.StartDate),
			EndDate:         calendarDateString(row.EndDate),
			IsPrivate:       row.IsPrivate,
			CreatedAt:       row.CreatedAt,
			UpdatedAt:       row.UpdatedAt,
		}
	}
	return events
}

func calendarDateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format("2006-01-02")
	return &formatted
}

func toCoreScheduleBlock(row dbScheduleBlock) calendar.CoreScheduleBlock {
	block := calendar.CoreScheduleBlock{
		ID:                   row.ID,
		WorkspaceID:          row.WorkspaceID,
		UserID:               row.UserID,
		StoryID:              row.StoryID,
		StoryTitle:           row.StoryTitle,
		StoryCode:            row.StoryCode,
		StoryPriority:        row.StoryPriority,
		StoryEndDate:         row.StoryEndDate,
		TeamID:               row.TeamID,
		TeamName:             row.TeamName,
		TeamCode:             row.TeamCode,
		BlockType:            calendar.ScheduleBlockType(row.BlockType),
		Title:                row.Title,
		StartAt:              row.StartAt,
		EndAt:                row.EndAt,
		HasConflict:          row.HasConflict,
		IsLocked:             row.IsLocked,
		AutoSchedulingStatus: row.AutoSchedulingStatus,
		AutoSchedulingReason: row.AutoSchedulingReason,
		Source:               calendar.ScheduleBlockSource(row.Source),
		SegmentIndex:         row.SegmentIndex,
		ExternalCalendarID:   row.ExternalCalendarID,
		ExternalEventID:      row.ExternalEventID,
		ExternalSyncHash:     row.ExternalSyncHash,
		ExternalSyncedAt:     row.ExternalSyncedAt,
		CreatedAt:            row.CreatedAt,
		UpdatedAt:            row.UpdatedAt,
		ManualOverrideAt:     row.ManualOverrideAt,
		ManualOverrideBy:     row.ManualOverrideBy,
	}
	if row.ExternalProvider != nil {
		provider := calendar.Provider(*row.ExternalProvider)
		block.ExternalProvider = &provider
	}
	return block
}

func toCoreScheduleBlocks(rows []dbScheduleBlock) []calendar.CoreScheduleBlock {
	blocks := make([]calendar.CoreScheduleBlock, len(rows))
	for i, row := range rows {
		blocks[i] = toCoreScheduleBlock(row)
	}
	return blocks
}
