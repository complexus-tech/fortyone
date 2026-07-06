package calendar

import (
	"net/http"
	"testing"
	"time"

	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
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

func TestGoogleEventToBusyWindowKeepsVisibleEventTitle(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:           "event-id",
		Summary:      "Team sync",
		Transparency: "opaque",
		Visibility:   "default",
		Start: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T10:00:00Z",
		},
		End: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T10:30:00Z",
		},
	}

	window, ok, err := googleEventToBusyWindow("primary", event)
	if err != nil {
		t.Fatalf("googleEventToBusyWindow returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected event to become a busy window")
	}
	if window.ProviderEventID != "primary:event-id" {
		t.Fatalf("unexpected provider event id: %s", window.ProviderEventID)
	}
	if window.Title == nil || *window.Title != "Team sync" {
		t.Fatalf("expected event title to be preserved: %#v", window)
	}
	if window.IsPrivate {
		t.Fatalf("expected visible event to be non-private: %#v", window)
	}
}

func TestGoogleEventToBusyWindowFallsBackToBusyForPrivateEvents(t *testing.T) {
	t.Parallel()

	event := &calendarapi.Event{
		Id:           "private-event-id",
		Summary:      "Private appointment",
		Transparency: "opaque",
		Visibility:   "private",
		Start: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T11:00:00Z",
		},
		End: &calendarapi.EventDateTime{
			DateTime: "2026-06-15T12:00:00Z",
		},
	}

	window, ok, err := googleEventToBusyWindow("primary", event)
	if err != nil {
		t.Fatalf("googleEventToBusyWindow returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected private event to become a busy window")
	}
	if window.Title != nil {
		t.Fatalf("expected private event title to be hidden: %#v", window.Title)
	}
	if !window.IsPrivate {
		t.Fatalf("expected private event to be marked private: %#v", window)
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
