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
	ID                      uuid.UUID  `json:"id"`
	Provider                string     `json:"provider"`
	ConnectedEmail          string     `json:"connectedEmail"`
	Timezone                string     `json:"timezone"`
	Scopes                  []string   `json:"scopes"`
	CanReadEventDetails     bool       `json:"canReadEventDetails"`
	CanWriteEvents          bool       `json:"canWriteEvents"`
	RequiresReauthorization bool       `json:"requiresReauthorization"`
	ReauthorizationReason   *string    `json:"reauthorizationReason,omitempty"`
	SyncStatus              string     `json:"syncStatus"`
	SyncError               *string    `json:"syncError,omitempty"`
	LastSyncedAt            *time.Time `json:"lastSyncedAt,omitempty"`
	CreatedAt               time.Time  `json:"createdAt"`
	UpdatedAt               time.Time  `json:"updatedAt"`
}

type AppCreateConnectSession struct {
	AuthURL string `json:"authUrl"`
}

type AppSchedule struct {
	StartAt        time.Time                 `json:"startAt"`
	EndAt          time.Time                 `json:"endAt"`
	Events         []AppCalendarEventSummary `json:"events"`
	BusyWindows    []AppBusyWindow           `json:"busyWindows"`
	Blocks         []AppScheduleBlock        `json:"blocks"`
	ScheduleIssues []AppScheduleIssue        `json:"scheduleIssues"`
}

type AppScheduleIssue struct {
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

type AppCalendarEventSummary struct {
	ID         uuid.UUID `json:"id"`
	Provider   string    `json:"provider"`
	CalendarID string    `json:"calendarId"`
	Title      *string   `json:"title,omitempty"`
	Location   *string   `json:"location,omitempty"`
	MeetingURL *string   `json:"meetingUrl,omitempty"`
	HTMLLink   *string   `json:"htmlLink,omitempty"`
	StartAt    time.Time `json:"startAt"`
	EndAt      time.Time `json:"endAt"`
	IsAllDay   bool      `json:"isAllDay"`
	StartDate  *string   `json:"startDate,omitempty"`
	EndDate    *string   `json:"endDate,omitempty"`
	IsPrivate  bool      `json:"isPrivate"`
}

type AppCalendarOrganizer struct {
	DisplayName string `json:"displayName,omitempty"`
	Email       string `json:"email,omitempty"`
}

type AppCalendarAttendee struct {
	DisplayName    string `json:"displayName,omitempty"`
	Email          string `json:"email,omitempty"`
	ResponseStatus string `json:"responseStatus,omitempty"`
	Optional       bool   `json:"optional"`
	Organizer      bool   `json:"organizer"`
	Self           bool   `json:"self"`
}

type AppCalendarEvent struct {
	ID               uuid.UUID             `json:"id"`
	Provider         string                `json:"provider"`
	CalendarID       string                `json:"calendarId"`
	Title            *string               `json:"title,omitempty"`
	Description      *string               `json:"description,omitempty"`
	Location         *string               `json:"location,omitempty"`
	MeetingURL       *string               `json:"meetingUrl,omitempty"`
	HTMLLink         *string               `json:"htmlLink,omitempty"`
	Organizer        *AppCalendarOrganizer `json:"organizer,omitempty"`
	Attendees        []AppCalendarAttendee `json:"attendees"`
	AttendeesOmitted bool                  `json:"attendeesOmitted"`
	StartAt          time.Time             `json:"startAt"`
	EndAt            time.Time             `json:"endAt"`
	IsAllDay         bool                  `json:"isAllDay"`
	StartDate        *string               `json:"startDate,omitempty"`
	EndDate          *string               `json:"endDate,omitempty"`
	IsPrivate        bool                  `json:"isPrivate"`
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
	ID                   uuid.UUID  `json:"id"`
	StoryID              *uuid.UUID `json:"storyId,omitempty"`
	StoryTitle           *string    `json:"storyTitle,omitempty"`
	StoryCode            *string    `json:"storyCode,omitempty"`
	StoryStatusColor     *string    `json:"storyStatusColor,omitempty"`
	TeamID               *uuid.UUID `json:"teamId,omitempty"`
	TeamName             *string    `json:"teamName,omitempty"`
	TeamCode             *string    `json:"teamCode,omitempty"`
	BlockType            string     `json:"blockType"`
	Title                string     `json:"title"`
	StartAt              time.Time  `json:"startAt"`
	EndAt                time.Time  `json:"endAt"`
	HasConflict          bool       `json:"hasConflict"`
	IsLocked             bool       `json:"isLocked"`
	IsCrossWorkspace     bool       `json:"isCrossWorkspace,omitempty"`
	AutoSchedulingStatus *string    `json:"autoSchedulingStatus,omitempty"`
	AutoSchedulingReason *string    `json:"autoSchedulingReason,omitempty"`
	Source               string     `json:"source"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            time.Time  `json:"updatedAt"`
	ManualOverrideAt     *time.Time `json:"manualOverrideAt,omitempty"`
	ManualOverrideBy     *uuid.UUID `json:"manualOverrideBy,omitempty"`
}

type AppScheduleBlockRequest struct {
	StoryID   *uuid.UUID `json:"storyId"`
	BlockType string     `json:"blockType"`
	Title     string     `json:"title"`
	StartAt   time.Time  `json:"startAt"`
	EndAt     time.Time  `json:"endAt"`
	IsLocked  *bool      `json:"isLocked"`
}

type AppManualScheduleBlockRequest struct {
	StartAt           time.Time  `json:"startAt"`
	EndAt             time.Time  `json:"endAt"`
	ExpectedUpdatedAt *time.Time `json:"expectedUpdatedAt"`
	Timezone          string     `json:"timezone"`
	Change            string     `json:"change"`
	ClientMutationID  uuid.UUID  `json:"clientMutationId"`
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
	result := AppConnection{
		ID:                      connection.ID,
		Provider:                string(connection.Provider),
		ConnectedEmail:          connection.ConnectedEmail,
		Timezone:                connection.Timezone,
		Scopes:                  connection.Scopes,
		CanReadEventDetails:     connection.CanReadEventDetails(),
		CanWriteEvents:          connection.CanWriteEvents(),
		RequiresReauthorization: connection.RequiresReauthorization(),
		SyncStatus:              string(connection.SyncStatus),
		SyncError:               connection.SyncError,
		LastSyncedAt:            connection.LastSyncedAt,
		CreatedAt:               connection.CreatedAt,
		UpdatedAt:               connection.UpdatedAt,
	}
	if result.RequiresReauthorization {
		reason := calendar.GoogleCalendarWriteScopeReason
		result.ReauthorizationReason = &reason
	}
	return result
}

func toAppSchedule(schedule calendar.CoreCalendarView) AppSchedule {
	return AppSchedule{
		StartAt:        schedule.StartAt,
		EndAt:          schedule.EndAt,
		Events:         toAppCalendarEventSummaries(schedule.Events),
		BusyWindows:    toAppBusyWindows(schedule.BusyWindows),
		Blocks:         toAppScheduleBlocks(schedule.Blocks),
		ScheduleIssues: toAppScheduleIssues(schedule.ScheduleIssues),
	}
}

func toAppScheduleIssues(issues []calendar.CoreScheduleIssue) []AppScheduleIssue {
	out := make([]AppScheduleIssue, len(issues))
	for index, issue := range issues {
		out[index] = AppScheduleIssue{
			StoryID:                  issue.StoryID,
			StoryTitle:               issue.StoryTitle,
			StoryCode:                issue.StoryCode,
			TeamID:                   issue.TeamID,
			TeamName:                 issue.TeamName,
			TeamCode:                 issue.TeamCode,
			EstimatedDurationMinutes: issue.EstimatedDurationMinutes,
			ScheduledDurationMinutes: issue.ScheduledDurationMinutes,
			RemainingDurationMinutes: issue.RemainingDurationMinutes,
			AutoSchedulingStatus:     issue.AutoSchedulingStatus,
			AutoSchedulingReason:     issue.AutoSchedulingReason,
			UpdatedAt:                issue.UpdatedAt,
		}
	}
	return out
}

func toAppCalendarEventSummaries(events []calendar.CoreCalendarEventSummary) []AppCalendarEventSummary {
	out := make([]AppCalendarEventSummary, len(events))
	for i, event := range events {
		if event.IsPrivate {
			out[i] = AppCalendarEventSummary{
				ID:         event.ID,
				Provider:   string(event.Provider),
				CalendarID: event.CalendarID,
				StartAt:    event.StartAt,
				EndAt:      event.EndAt,
				IsAllDay:   event.IsAllDay,
				StartDate:  event.StartDate,
				EndDate:    event.EndDate,
				IsPrivate:  true,
			}
			continue
		}
		out[i] = AppCalendarEventSummary{
			ID:         event.ID,
			Provider:   string(event.Provider),
			CalendarID: event.CalendarID,
			Title:      event.Title,
			Location:   event.Location,
			MeetingURL: event.MeetingURL,
			HTMLLink:   event.HTMLLink,
			StartAt:    event.StartAt,
			EndAt:      event.EndAt,
			IsAllDay:   event.IsAllDay,
			StartDate:  event.StartDate,
			EndDate:    event.EndDate,
			IsPrivate:  event.IsPrivate,
		}
	}
	return out
}

func toAppCalendarEvent(event calendar.CoreCalendarEvent) AppCalendarEvent {
	if event.IsPrivate {
		return AppCalendarEvent{
			ID:         event.ID,
			Provider:   string(event.Provider),
			CalendarID: event.CalendarID,
			Attendees:  []AppCalendarAttendee{},
			StartAt:    event.StartAt,
			EndAt:      event.EndAt,
			IsAllDay:   event.IsAllDay,
			StartDate:  event.StartDate,
			EndDate:    event.EndDate,
			IsPrivate:  true,
		}
	}

	attendees := make([]AppCalendarAttendee, len(event.Attendees))
	for i, attendee := range event.Attendees {
		attendees[i] = AppCalendarAttendee{
			DisplayName:    attendee.DisplayName,
			Email:          attendee.Email,
			ResponseStatus: attendee.ResponseStatus,
			Optional:       attendee.Optional,
			Organizer:      attendee.Organizer,
			Self:           attendee.Self,
		}
	}
	var organizer *AppCalendarOrganizer
	if event.Organizer != nil {
		organizer = &AppCalendarOrganizer{
			DisplayName: event.Organizer.DisplayName,
			Email:       event.Organizer.Email,
		}
	}
	return AppCalendarEvent{
		ID:               event.ID,
		Provider:         string(event.Provider),
		CalendarID:       event.CalendarID,
		Title:            event.Title,
		Description:      event.Description,
		Location:         event.Location,
		MeetingURL:       event.MeetingURL,
		HTMLLink:         event.HTMLLink,
		Organizer:        organizer,
		Attendees:        attendees,
		AttendeesOmitted: event.AttendeesOmitted,
		StartAt:          event.StartAt,
		EndAt:            event.EndAt,
		IsAllDay:         event.IsAllDay,
		StartDate:        event.StartDate,
		EndDate:          event.EndDate,
		IsPrivate:        event.IsPrivate,
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
	source := string(block.Source)
	if block.IsCrossWorkspace {
		source = "other_workspace"
	}
	return AppScheduleBlock{
		ID:                   block.ID,
		StoryID:              block.StoryID,
		StoryTitle:           block.StoryTitle,
		StoryCode:            block.StoryCode,
		StoryStatusColor:     block.StoryStatusColor,
		TeamID:               block.TeamID,
		TeamName:             block.TeamName,
		TeamCode:             block.TeamCode,
		BlockType:            string(block.BlockType),
		Title:                block.Title,
		StartAt:              block.StartAt,
		EndAt:                block.EndAt,
		HasConflict:          block.HasConflict,
		IsLocked:             block.IsLocked,
		IsCrossWorkspace:     block.IsCrossWorkspace,
		AutoSchedulingStatus: block.AutoSchedulingStatus,
		AutoSchedulingReason: block.AutoSchedulingReason,
		Source:               source,
		CreatedAt:            block.CreatedAt,
		UpdatedAt:            block.UpdatedAt,
		ManualOverrideAt:     block.ManualOverrideAt,
		ManualOverrideBy:     block.ManualOverrideBy,
	}
}
