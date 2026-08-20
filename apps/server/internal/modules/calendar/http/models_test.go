package calendarhttp

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	calendar "github.com/complexus-tech/projects-api/internal/modules/calendar/service"
	"github.com/google/uuid"
)

func TestToAppConnectionReportsEventDetailCapability(t *testing.T) {
	t.Parallel()

	connection := toAppConnection(calendar.CoreConnection{
		ID:       uuid.New(),
		Provider: calendar.ProviderGoogle,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar.events.readonly",
		},
	})

	if !connection.CanReadEventDetails {
		t.Fatal("expected Google event detail scope to enable event details")
	}
	if connection.CanWriteEvents || !connection.RequiresReauthorization || connection.ReauthorizationReason == nil || *connection.ReauthorizationReason != calendar.GoogleCalendarWriteScopeReason {
		t.Fatalf("expected deterministic write-scope reconnect state: %#v", connection)
	}

	upgraded := toAppConnection(calendar.CoreConnection{
		ID: uuid.New(), Provider: calendar.ProviderGoogle,
		Scopes: []string{calendar.GoogleCalendarEventsReadonlyScope, calendar.GoogleCalendarEventsOwnedScope},
	})
	if !upgraded.CanWriteEvents || upgraded.RequiresReauthorization || upgraded.ReauthorizationReason != nil {
		t.Fatalf("expected upgraded grant to be write-capable: %#v", upgraded)
	}
}

func TestCalendarSummaryDoesNotLeakEventDetailsOrProviderIdentity(t *testing.T) {
	t.Parallel()

	title := "Customer review"
	summary := toAppSchedule(calendar.CoreCalendarView{
		StartAt: time.Date(2026, 6, 15, 8, 0, 0, 0, time.UTC),
		EndAt:   time.Date(2026, 6, 15, 18, 0, 0, 0, time.UTC),
		Events: []calendar.CoreCalendarEventSummary{{
			ID:              uuid.New(),
			ConnectionID:    uuid.New(),
			Provider:        calendar.ProviderGoogle,
			CalendarID:      "primary",
			ProviderEventID: "primary:provider-secret",
			Title:           &title,
			StartAt:         time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
			EndAt:           time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
		}},
	})

	payload, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("marshal calendar summary: %v", err)
	}
	serialized := string(payload)
	for _, forbidden := range []string{"providerEventId", "provider-secret", "description", "attendees", "sourceHash", "connectionId"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("calendar summary leaked %q: %s", forbidden, serialized)
		}
	}
}

func TestCalendarEventDetailUsesOwnerSafeContract(t *testing.T) {
	t.Parallel()

	title := "Customer review"
	description := "Review the launch plan"
	meetingURL := "https://meet.google.com/abc-defg-hij"
	htmlLink := "https://calendar.google.com/calendar/event?eid=event-id"
	event := toAppCalendarEvent(calendar.CoreCalendarEvent{
		ID:              uuid.New(),
		ConnectionID:    uuid.New(),
		WorkspaceID:     uuid.New(),
		UserID:          uuid.New(),
		Provider:        calendar.ProviderGoogle,
		CalendarID:      "primary",
		ProviderEventID: "primary:provider-secret",
		Title:           &title,
		Description:     &description,
		MeetingURL:      &meetingURL,
		HTMLLink:        &htmlLink,
		Organizer: &calendar.CoreCalendarParticipant{
			DisplayName: "Joseph",
			Email:       "joseph@example.com",
			Self:        true,
		},
		Attendees: []calendar.CoreCalendarParticipant{{
			DisplayName:    "Tariro",
			Email:          "tariro@example.com",
			ResponseStatus: "accepted",
			Optional:       true,
		}},
		StartAt:    time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		EndAt:      time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
		SourceHash: "secret-source-hash",
	})

	payload, err := json.Marshal(event)
	if err != nil {
		t.Fatalf("marshal calendar event: %v", err)
	}
	serialized := string(payload)
	for _, required := range []string{"description", "organizer", "attendees", "optional", "self", "htmlLink", "isAllDay"} {
		if !strings.Contains(serialized, required) {
			t.Fatalf("calendar detail omitted %q: %s", required, serialized)
		}
	}
	for _, forbidden := range []string{"providerEventId", "provider-secret", "sourceHash", "secret-source-hash", "workspaceId", "userId", "connectionId"} {
		if strings.Contains(serialized, forbidden) {
			t.Fatalf("calendar detail leaked %q: %s", forbidden, serialized)
		}
	}
}

func TestCalendarContractsRedactPrivateEventDetailsDefensively(t *testing.T) {
	t.Parallel()

	title := "Private appointment"
	description := "Sensitive notes"
	location := "Private clinic"
	meetingURL := "https://meet.google.com/private"
	htmlLink := "https://calendar.google.com/private"
	event := calendar.CoreCalendarEvent{
		ID:          uuid.New(),
		Provider:    calendar.ProviderGoogle,
		CalendarID:  "primary",
		Title:       &title,
		Description: &description,
		Location:    &location,
		MeetingURL:  &meetingURL,
		HTMLLink:    &htmlLink,
		Organizer:   &calendar.CoreCalendarParticipant{Email: "owner@example.com"},
		Attendees:   []calendar.CoreCalendarParticipant{{Email: "guest@example.com"}},
		StartAt:     time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		EndAt:       time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC),
		IsPrivate:   true,
	}

	payloads := []any{
		toAppCalendarEvent(event),
		toAppCalendarEventSummaries([]calendar.CoreCalendarEventSummary{{
			ID:         event.ID,
			Provider:   event.Provider,
			CalendarID: event.CalendarID,
			Title:      event.Title,
			Location:   event.Location,
			MeetingURL: event.MeetingURL,
			HTMLLink:   event.HTMLLink,
			StartAt:    event.StartAt,
			EndAt:      event.EndAt,
			IsPrivate:  true,
		}}),
	}
	for _, payload := range payloads {
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal private calendar payload: %v", err)
		}
		serialized := string(encoded)
		for _, forbidden := range []string{title, description, location, meetingURL, htmlLink, "owner@example.com", "guest@example.com"} {
			if strings.Contains(serialized, forbidden) {
				t.Fatalf("private calendar payload leaked %q: %s", forbidden, serialized)
			}
		}
	}
}

func TestScheduleBlockResponsePropagatesAutoSchedulingMetadata(t *testing.T) {
	t.Parallel()

	status := "cannot_fit"
	reason := "Maya could not fit this story into the current planning window."
	block := toAppScheduleBlock(calendar.CoreScheduleBlock{
		AutoSchedulingStatus: &status,
		AutoSchedulingReason: &reason,
		Source:               calendar.ScheduleBlockSourceMaya,
	})
	if block.AutoSchedulingStatus == nil || *block.AutoSchedulingStatus != status {
		t.Fatalf("auto-scheduling status = %v, want %q", block.AutoSchedulingStatus, status)
	}
	if block.AutoSchedulingReason == nil || *block.AutoSchedulingReason != reason {
		t.Fatalf("auto-scheduling reason = %v, want %q", block.AutoSchedulingReason, reason)
	}

	payload, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal schedule block: %v", err)
	}
	serialized := string(payload)
	for _, contract := range []string{
		`"autoSchedulingStatus":"cannot_fit"`,
		`"autoSchedulingReason":"Maya could not fit this story into the current planning window."`,
	} {
		if !strings.Contains(serialized, contract) {
			t.Errorf("schedule block response is missing %s: %s", contract, serialized)
		}
	}

	withoutMetadata, err := json.Marshal(toAppScheduleBlock(calendar.CoreScheduleBlock{
		Source: calendar.ScheduleBlockSourceUser,
	}))
	if err != nil {
		t.Fatalf("marshal user schedule block: %v", err)
	}
	for _, omitted := range []string{"autoSchedulingStatus", "autoSchedulingReason"} {
		if strings.Contains(string(withoutMetadata), omitted) {
			t.Errorf("schedule block response must omit absent %q metadata: %s", omitted, withoutMetadata)
		}
	}
}

func TestScheduleBlockResponsePropagatesStoryStatusColor(t *testing.T) {
	t.Parallel()

	color := "#3c90ff"
	block := toAppScheduleBlock(calendar.CoreScheduleBlock{
		StoryStatusColor: &color,
		Source:           calendar.ScheduleBlockSourceUser,
	})
	if block.StoryStatusColor == nil || *block.StoryStatusColor != color {
		t.Fatalf("story status color = %v, want %q", block.StoryStatusColor, color)
	}

	payload, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal schedule block: %v", err)
	}
	if !strings.Contains(string(payload), `"storyStatusColor":"#3c90ff"`) {
		t.Fatalf("schedule block response is missing story status color: %s", payload)
	}
}

func TestCrossWorkspaceScheduleBlockResponseUsesGenericProvenance(t *testing.T) {
	t.Parallel()

	block := toAppScheduleBlock(calendar.CoreScheduleBlock{
		ID:               uuid.New(),
		IsCrossWorkspace: true,
		Source:           calendar.ScheduleBlockSourceMaya,
		Title:            "Scheduled elsewhere",
	})
	if !block.IsCrossWorkspace || block.Source != "other_workspace" {
		t.Fatalf("unexpected cross-workspace response: %#v", block)
	}

	payload, err := json.Marshal(block)
	if err != nil {
		t.Fatalf("marshal cross-workspace schedule block: %v", err)
	}
	serialized := string(payload)
	for _, contract := range []string{
		`"isCrossWorkspace":true`,
		`"source":"other_workspace"`,
		`"title":"Scheduled elsewhere"`,
	} {
		if !strings.Contains(serialized, contract) {
			t.Errorf("cross-workspace response is missing %s: %s", contract, serialized)
		}
	}
	if strings.Contains(serialized, `"source":"maya"`) {
		t.Fatalf("cross-workspace response leaked scheduling provenance: %s", serialized)
	}
}
