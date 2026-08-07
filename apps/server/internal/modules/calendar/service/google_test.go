package calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

func TestGoogleFreeBusyRangesChunksLongSyncWindow(t *testing.T) {
	t.Parallel()

	timeMin := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	timeMax := timeMin.Add((97 * 24 * time.Hour) + time.Hour)

	ranges := googleFreeBusyRanges(timeMin, timeMax)

	if len(ranges) != 4 {
		t.Fatalf("expected four ranges, got %d: %#v", len(ranges), ranges)
	}
	if ranges[0].start != timeMin {
		t.Fatalf("unexpected first range start: %s", ranges[0].start)
	}
	if ranges[len(ranges)-1].end != timeMax {
		t.Fatalf("unexpected last range end: %s", ranges[len(ranges)-1].end)
	}
	for i, timeRange := range ranges {
		if !timeRange.end.After(timeRange.start) {
			t.Fatalf("range %d is invalid: %#v", i, timeRange)
		}
		if timeRange.end.Sub(timeRange.start) > googleFreeBusyChunkSize {
			t.Fatalf("range %d is too long: %s", i, timeRange.end.Sub(timeRange.start))
		}
		if i > 0 && ranges[i-1].end != timeRange.start {
			t.Fatalf("range %d is not contiguous with previous range", i)
		}
	}
}

func TestGoogleFreeBusyRangesRejectsEmptyWindow(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)

	if ranges := googleFreeBusyRanges(now, now); len(ranges) != 0 {
		t.Fatalf("expected no ranges for empty window, got %#v", ranges)
	}
}

func TestListGoogleCalendarSnapshotPaginatesEvents(t *testing.T) {
	t.Parallel()

	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.Header().Set("Content-Type", "application/json")

		response := &calendarapi.Events{}
		switch r.URL.Query().Get("pageToken") {
		case "":
			response.NextPageToken = "next-page"
			response.Items = []*calendarapi.Event{googleTestEvent("event-one", time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC))}
		case "next-page":
			response.Items = []*calendarapi.Event{googleTestEvent("event-two", time.Date(2026, 6, 15, 11, 0, 0, 0, time.UTC))}
		default:
			http.Error(w, "unexpected page token", http.StatusBadRequest)
			return
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	t.Cleanup(server.Close)

	client, err := calendarapi.NewService(
		context.Background(),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("create calendar client: %v", err)
	}
	timeMin := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	snapshot, err := listGoogleCalendarSnapshot(context.Background(), client, BusyWindowInput{
		TimeMin:  timeMin,
		TimeMax:  timeMin.Add(24 * time.Hour),
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("listGoogleCalendarSnapshot returned error: %v", err)
	}
	if requestCount.Load() != 2 {
		t.Fatalf("expected two pages, got %d requests", requestCount.Load())
	}
	if len(snapshot.Events) != 2 || len(snapshot.BusyWindows) != 2 {
		t.Fatalf("expected both pages in snapshot: %#v", snapshot)
	}
}

func TestGoogleEventToCalendarEventKeepsOwnerVisibleDetails(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:           "event-id",
		Summary:      "Team sync",
		Description:  "Review the release plan",
		Location:     "Harare office",
		HangoutLink:  "http://unsafe.example.com/meeting",
		HtmlLink:     "https://calendar.google.com/calendar/event?eid=event-id",
		Transparency: "opaque",
		Visibility:   "default",
		ConferenceData: &calendarapi.ConferenceData{EntryPoints: []*calendarapi.EntryPoint{
			{EntryPointType: "video", Uri: "https://meet.google.com/abc-defg-hij"},
		}},
		Organizer: &calendarapi.EventOrganizer{DisplayName: "Joseph", Email: "joseph@example.com", Self: true},
		Attendees: []*calendarapi.EventAttendee{
			{DisplayName: "Joseph", Email: "joseph@example.com", ResponseStatus: "accepted", Self: true, Organizer: true},
			{DisplayName: "Tariro", Email: "tariro@example.com", ResponseStatus: "tentative", Optional: true},
		},
		Start: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T10:00:00Z",
		},
		End: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T10:30:00Z",
		},
	}

	calendarEvent, ok, err := googleEventToCalendarEvent("primary", event, "Africa/Harare")
	if err != nil {
		t.Fatalf("googleEventToCalendarEvent returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected event to be normalized")
	}
	if calendarEvent.ProviderEventID != "primary:event-id" {
		t.Fatalf("unexpected provider event id: %s", calendarEvent.ProviderEventID)
	}
	if calendarEvent.Title == nil || *calendarEvent.Title != "Team sync" {
		t.Fatalf("expected event title to be preserved: %#v", calendarEvent)
	}
	if calendarEvent.Description == nil || *calendarEvent.Description != "Review the release plan" {
		t.Fatalf("expected event description to be preserved: %#v", calendarEvent)
	}
	if calendarEvent.MeetingURL == nil || *calendarEvent.MeetingURL != "https://meet.google.com/abc-defg-hij" {
		t.Fatalf("expected safe video entry point to be used: %#v", calendarEvent.MeetingURL)
	}
	if calendarEvent.HTMLLink == nil || calendarEvent.Organizer == nil || len(calendarEvent.Attendees) != 2 {
		t.Fatalf("expected owner event details to be preserved: %#v", calendarEvent)
	}
	if calendarEvent.IsPrivate {
		t.Fatalf("expected visible event to be non-private: %#v", calendarEvent)
	}
	window, blocksTime := calendarEventToBusyWindow(calendarEvent)
	if !blocksTime || window.Title != nil {
		t.Fatalf("expected a title-free blocking window: %#v", window)
	}
}

func TestGoogleEventToCalendarEventRedactsPrivateDetails(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:           "private-event-id",
		Summary:      "Private appointment",
		Description:  "Sensitive notes",
		Location:     "Private clinic",
		HangoutLink:  "https://meet.google.com/private",
		HtmlLink:     "https://calendar.google.com/private",
		Transparency: "opaque",
		Visibility:   "private",
		Organizer:    &calendarapi.EventOrganizer{Email: "owner@example.com"},
		Attendees:    []*calendarapi.EventAttendee{{Email: "guest@example.com"}},
		Start: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T11:00:00Z",
		},
		End: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T12:00:00Z",
		},
	}

	calendarEvent, ok, err := googleEventToCalendarEvent("primary", event, "UTC")
	if err != nil {
		t.Fatalf("googleEventToCalendarEvent returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected private event to be normalized")
	}
	if calendarEvent.Title != nil || calendarEvent.Description != nil || calendarEvent.Location != nil || calendarEvent.MeetingURL != nil || calendarEvent.HTMLLink != nil {
		t.Fatalf("expected private event details to be hidden: %#v", calendarEvent)
	}
	if calendarEvent.Organizer != nil || len(calendarEvent.Attendees) != 0 {
		t.Fatalf("expected private participants to be hidden: %#v", calendarEvent)
	}
	if !calendarEvent.IsPrivate {
		t.Fatalf("expected private event to be marked private: %#v", calendarEvent)
	}
}

func TestGoogleEventToCalendarEventKeepsTransparentEventsOutOfBusyWindows(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:           "available-event",
		Summary:      "Optional office hours",
		Transparency: "transparent",
		Start:        &calendarapi.EventDateTime{DateTime: "2026-06-15T11:00:00Z"},
		End:          &calendarapi.EventDateTime{DateTime: "2026-06-15T12:00:00Z"},
	}

	calendarEvent, ok, err := googleEventToCalendarEvent("primary", event, "UTC")
	if err != nil || !ok {
		t.Fatalf("expected transparent event to be normalized: %#v, %v", calendarEvent, err)
	}
	if calendarEvent.BlocksTime {
		t.Fatalf("expected transparent event not to block time: %#v", calendarEvent)
	}
	if _, blocksTime := calendarEventToBusyWindow(calendarEvent); blocksTime {
		t.Fatal("expected transparent event to stay out of availability windows")
	}
}

func TestGoogleEventToCalendarEventParsesAllDayInCalendarTimezone(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:      "all-day-event",
		Summary: "Company offsite",
		Start:   &calendarapi.EventDateTime{Date: "2026-06-15"},
		End:     &calendarapi.EventDateTime{Date: "2026-06-16"},
	}

	calendarEvent, ok, err := googleEventToCalendarEvent("primary", event, "Africa/Harare")
	if err != nil || !ok {
		t.Fatalf("expected all-day event to be normalized: %#v, %v", calendarEvent, err)
	}
	location, err := time.LoadLocation("Africa/Harare")
	if err != nil {
		t.Fatalf("load test timezone: %v", err)
	}
	expectedStart := time.Date(2026, 6, 15, 0, 0, 0, 0, location)
	if !calendarEvent.IsAllDay || !calendarEvent.StartAt.Equal(expectedStart) {
		t.Fatalf("unexpected all-day start: %#v", calendarEvent)
	}
	if calendarEvent.StartDate == nil || *calendarEvent.StartDate != "2026-06-15" || calendarEvent.EndDate == nil || *calendarEvent.EndDate != "2026-06-16" {
		t.Fatalf("expected provider date-only values: %#v", calendarEvent)
	}
}

func TestListGoogleCalendarSnapshotUsesResponseTimezoneForAllDayAvailability(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(&calendarapi.Events{
			TimeZone: "Asia/Tokyo",
			Items: []*calendarapi.Event{{
				Id:      "all-day-event",
				Summary: "Tokyo holiday",
				Start:   &calendarapi.EventDateTime{Date: "2026-06-15"},
				End:     &calendarapi.EventDateTime{Date: "2026-06-16"},
			}},
		})
	}))
	t.Cleanup(server.Close)

	client, err := calendarapi.NewService(
		context.Background(),
		option.WithEndpoint(server.URL+"/"),
		option.WithoutAuthentication(),
	)
	if err != nil {
		t.Fatalf("create calendar client: %v", err)
	}
	timeMin := time.Date(2026, 6, 15, 0, 0, 0, 0, time.UTC)
	snapshot, err := listGoogleCalendarSnapshot(context.Background(), client, BusyWindowInput{
		TimeMin:  timeMin,
		TimeMax:  timeMin.Add(24 * time.Hour),
		Timezone: "UTC",
	})
	if err != nil {
		t.Fatalf("listGoogleCalendarSnapshot returned error: %v", err)
	}
	location, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		t.Fatalf("load test timezone: %v", err)
	}
	expectedStart := time.Date(2026, 6, 15, 0, 0, 0, 0, location)
	if len(snapshot.BusyWindows) != 1 || !snapshot.BusyWindows[0].StartAt.Equal(expectedStart) {
		t.Fatalf("expected response timezone to anchor all-day availability: %#v", snapshot)
	}
}

func TestIsGoogleInsufficientPermissionsError(t *testing.T) {
	t.Parallel()

	err := &googleapi.Error{
		Code:    http.StatusForbidden,
		Message: "Insufficient Permission",
		Errors: []googleapi.ErrorItem{
			{Reason: "insufficientPermissions"},
		},
	}

	if !isGoogleInsufficientPermissionsError(err) {
		t.Fatal("expected insufficient permissions error to be detected")
	}
}

func googleTestEvent(id string, startAt time.Time) *calendarapi.Event {
	return &calendarapi.Event{
		Id:      id,
		Summary: id,
		Start:   &calendarapi.EventDateTime{DateTime: startAt.Format(time.RFC3339)},
		End:     &calendarapi.EventDateTime{DateTime: startAt.Add(time.Hour).Format(time.RFC3339)},
	}
}
