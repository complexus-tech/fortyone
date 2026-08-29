package calendarrepository

import (
	"encoding/json"
	"fmt"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/domain"
	calendarsql "github.com/complexus-tech/projects-api/internal/modules/calendar/repository/sqlc"
	"github.com/google/uuid"
)

func toCoreConnection(row calendarsql.CalendarConnection) calendar.CoreConnection {
	return calendar.CoreConnection{
		ID:                     row.ConnectionID,
		WorkspaceID:            row.WorkspaceID,
		UserID:                 row.UserID,
		CredentialGeneration:   row.CredentialGeneration,
		ProviderAccountID:      row.ProviderAccountID,
		Provider:               calendar.Provider(row.Provider),
		IsPrimary:              row.IsPrimary,
		ConnectedEmail:         row.ConnectedEmail,
		Timezone:               row.Timezone,
		TokenPayload:           row.TokenPayload,
		Scopes:                 row.Scopes,
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

func toCoreConnections(rows []calendarsql.CalendarConnection) []calendar.CoreConnection {
	connections := make([]calendar.CoreConnection, len(rows))
	for index, row := range rows {
		connections[index] = toCoreConnection(row)
	}
	return connections
}

func toCoreBusyWindow(row calendarsql.CalendarBusyWindow) calendar.CoreBusyWindow {
	return calendar.CoreBusyWindow{
		ID:              row.WindowID,
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

func toCoreBusyWindows(rows []calendarsql.CalendarBusyWindow) []calendar.CoreBusyWindow {
	windows := make([]calendar.CoreBusyWindow, len(rows))
	for index, row := range rows {
		windows[index] = toCoreBusyWindow(row)
	}
	return windows
}

func toCoreCalendarEvent(row calendarsql.CalendarEvent) (calendar.CoreCalendarEvent, error) {
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
		ID:               row.EventID,
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
		HTMLLink:         row.HtmlLink,
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

func toCoreCalendarEventSummaries(rows []calendarsql.ListCalendarEventsRow) []calendar.CoreCalendarEventSummary {
	events := make([]calendar.CoreCalendarEventSummary, len(rows))
	for index, row := range rows {
		events[index] = calendar.CoreCalendarEventSummary{
			ID: row.EventID, ConnectionID: row.ConnectionID, Provider: calendar.Provider(row.Provider),
			CalendarID: row.CalendarID, ProviderEventID: row.ProviderEventID, Title: row.Title,
			Location: row.Location, MeetingURL: row.MeetingURL, HTMLLink: row.HtmlLink,
			StartAt: row.StartAt, EndAt: row.EndAt, IsAllDay: row.IsAllDay,
			StartDate: calendarDateString(row.StartDate), EndDate: calendarDateString(row.EndDate),
			IsPrivate: row.IsPrivate, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
	}
	return events
}

func calendarDateString(value *time.Time) *string {
	if value == nil {
		return nil
	}
	formatted := value.Format(time.DateOnly)
	return &formatted
}

type scheduleBlockRecord struct {
	ID                   uuid.UUID
	WorkspaceID          uuid.UUID
	UserID               uuid.UUID
	StoryID              *uuid.UUID
	StoryTitle           *string
	StoryCode            string
	StoryStatusColor     *string
	StoryPriority        string
	StoryEndDate         *time.Time
	TeamID               *uuid.UUID
	TeamName             *string
	TeamCode             *string
	BlockType            string
	Title                string
	StartAt              time.Time
	EndAt                time.Time
	CompletedAt          *time.Time
	HasConflict          bool
	IsLocked             bool
	AutoSchedulingStatus string
	AutoSchedulingReason string
	Source               string
	SegmentIndex         int32
	ExternalProvider     *string
	ExternalCalendarID   *string
	ExternalEventID      *string
	ExternalSyncHash     *string
	ExternalSyncedAt     *time.Time
	CreatedAt            time.Time
	UpdatedAt            time.Time
	ManualOverrideAt     *time.Time
	ManualOverrideBy     *uuid.UUID
}

func toCoreScheduleBlock(row scheduleBlockRecord) calendar.CoreScheduleBlock {
	block := calendar.CoreScheduleBlock{
		ID: row.ID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		StoryID: row.StoryID, StoryTitle: row.StoryTitle,
		StoryCode: stringPointerUnlessEmpty(row.StoryCode), StoryStatusColor: row.StoryStatusColor,
		StoryPriority: row.StoryPriority, StoryEndDate: row.StoryEndDate,
		TeamID: row.TeamID, TeamName: row.TeamName, TeamCode: row.TeamCode,
		BlockType: calendar.ScheduleBlockType(row.BlockType), Title: row.Title,
		StartAt: row.StartAt, EndAt: row.EndAt, CompletedAt: row.CompletedAt,
		HasConflict: row.HasConflict, IsLocked: row.IsLocked,
		AutoSchedulingStatus: stringPointerUnlessEmpty(row.AutoSchedulingStatus),
		AutoSchedulingReason: stringPointerUnlessEmpty(row.AutoSchedulingReason),
		Source:               calendar.ScheduleBlockSource(row.Source), SegmentIndex: int(row.SegmentIndex),
		ExternalCalendarID: row.ExternalCalendarID, ExternalEventID: row.ExternalEventID,
		ExternalSyncHash: row.ExternalSyncHash, ExternalSyncedAt: row.ExternalSyncedAt,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
	}
	if row.ExternalProvider != nil {
		provider := calendar.Provider(*row.ExternalProvider)
		block.ExternalProvider = &provider
	}
	return block
}

func stringPointerUnlessEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func scheduleBlockFromGet(row calendarsql.GetCalendarScheduleBlockRow) scheduleBlockRecord {
	return scheduleBlockRecord{
		ID: row.BlockID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, StoryID: row.StoryID,
		StoryTitle: row.StoryTitle, StoryCode: row.StoryCode, StoryStatusColor: row.StoryStatusColor,
		StoryPriority: row.StoryPriority, StoryEndDate: row.StoryEndDate, TeamID: row.TeamID,
		TeamName: row.TeamName, TeamCode: row.TeamCode, BlockType: row.BlockType, Title: row.Title,
		StartAt: row.StartAt, EndAt: row.EndAt, CompletedAt: row.CompletedAt, HasConflict: row.HasConflict,
		IsLocked: row.IsLocked, AutoSchedulingStatus: row.AutoSchedulingStatus,
		AutoSchedulingReason: row.AutoSchedulingReason, Source: row.Source, SegmentIndex: row.SegmentIndex,
		ExternalProvider: row.ExternalProvider, ExternalCalendarID: row.ExternalCalendarID,
		ExternalEventID: row.ExternalEventID, ExternalSyncHash: row.ExternalSyncHash,
		ExternalSyncedAt: row.ExternalSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
	}
}

func scheduleBlocksFromList(rows []calendarsql.ListCalendarScheduleBlocksRow) []calendar.CoreScheduleBlock {
	blocks := make([]calendar.CoreScheduleBlock, len(rows))
	for index, row := range rows {
		blocks[index] = toCoreScheduleBlock(scheduleBlockRecord{
			ID: row.BlockID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, StoryID: row.StoryID,
			StoryTitle: row.StoryTitle, StoryCode: row.StoryCode, StoryStatusColor: row.StoryStatusColor,
			StoryPriority: row.StoryPriority, StoryEndDate: row.StoryEndDate, TeamID: row.TeamID,
			TeamName: row.TeamName, TeamCode: row.TeamCode, BlockType: row.BlockType, Title: row.Title,
			StartAt: row.StartAt, EndAt: row.EndAt, CompletedAt: row.CompletedAt, HasConflict: row.HasConflict,
			IsLocked: row.IsLocked, AutoSchedulingStatus: row.AutoSchedulingStatus,
			AutoSchedulingReason: row.AutoSchedulingReason, Source: row.Source, SegmentIndex: row.SegmentIndex,
			ExternalProvider: row.ExternalProvider, ExternalCalendarID: row.ExternalCalendarID,
			ExternalEventID: row.ExternalEventID, ExternalSyncHash: row.ExternalSyncHash,
			ExternalSyncedAt: row.ExternalSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
		})
	}
	return blocks
}

func scheduleBlocksFromSchedulingList(rows []calendarsql.ListSchedulingBlocksForUserRow) []calendar.CoreScheduleBlock {
	blocks := make([]calendar.CoreScheduleBlock, len(rows))
	for index, row := range rows {
		blocks[index] = toCoreScheduleBlock(scheduleBlockRecord{
			ID: row.BlockID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, StoryID: row.StoryID,
			StoryTitle: row.StoryTitle, StoryCode: row.StoryCode, StoryStatusColor: row.StoryStatusColor,
			StoryPriority: row.StoryPriority, StoryEndDate: row.StoryEndDate, TeamID: row.TeamID,
			TeamName: row.TeamName, TeamCode: row.TeamCode, BlockType: row.BlockType, Title: row.Title,
			StartAt: row.StartAt, EndAt: row.EndAt, CompletedAt: row.CompletedAt, HasConflict: row.HasConflict,
			IsLocked: row.IsLocked, AutoSchedulingStatus: row.AutoSchedulingStatus,
			AutoSchedulingReason: row.AutoSchedulingReason, Source: row.Source, SegmentIndex: row.SegmentIndex,
			ExternalProvider: row.ExternalProvider, ExternalCalendarID: row.ExternalCalendarID,
			ExternalEventID: row.ExternalEventID, ExternalSyncHash: row.ExternalSyncHash,
			ExternalSyncedAt: row.ExternalSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
		})
	}
	return blocks
}

func scheduleBlocksFromMayaList(rows []calendarsql.ListMayaScheduleBlocksForStoryRow) []calendar.CoreScheduleBlock {
	blocks := make([]calendar.CoreScheduleBlock, len(rows))
	for index, row := range rows {
		blocks[index] = toCoreScheduleBlock(scheduleBlockRecord{
			ID: row.BlockID, WorkspaceID: row.WorkspaceID, UserID: row.UserID, StoryID: row.StoryID,
			StoryTitle: row.StoryTitle, StoryCode: row.StoryCode, StoryStatusColor: row.StoryStatusColor,
			StoryPriority: row.StoryPriority, StoryEndDate: row.StoryEndDate, TeamID: row.TeamID,
			TeamName: row.TeamName, TeamCode: row.TeamCode, BlockType: row.BlockType, Title: row.Title,
			StartAt: row.StartAt, EndAt: row.EndAt, CompletedAt: row.CompletedAt, HasConflict: row.HasConflict,
			IsLocked: row.IsLocked, AutoSchedulingStatus: row.AutoSchedulingStatus,
			AutoSchedulingReason: row.AutoSchedulingReason, Source: row.Source, SegmentIndex: row.SegmentIndex,
			ExternalProvider: row.ExternalProvider, ExternalCalendarID: row.ExternalCalendarID,
			ExternalEventID: row.ExternalEventID, ExternalSyncHash: row.ExternalSyncHash,
			ExternalSyncedAt: row.ExternalSyncedAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
			ManualOverrideAt: row.ManualOverrideAt, ManualOverrideBy: row.ManualOverrideBy,
		})
	}
	return blocks
}
