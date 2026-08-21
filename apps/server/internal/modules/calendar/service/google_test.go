package calendar

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	googleauth "github.com/complexus-tech/projects-api/pkg/google"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

type googleTestOAuth struct {
	client *http.Client
}

func (g googleTestOAuth) CalendarAuthCodeURL(string) (string, error) { return "", nil }
func (g googleTestOAuth) ExchangeCalendarCode(context.Context, string) (googleauth.CalendarToken, error) {
	return googleauth.CalendarToken{}, nil
}
func (g googleTestOAuth) CalendarHTTPClient(context.Context, *oauth2.Token) (*http.Client, error) {
	return g.client, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (fn roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return fn(request) }

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
	if snapshot.Timezone != "Asia/Tokyo" {
		t.Fatalf("expected response timezone to be persisted, got %q", snapshot.Timezone)
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

func TestGoogleProviderUpsertCreatesWithStablePrivateProvenance(t *testing.T) {
	t.Parallel()

	var methods []string
	var sendUpdates []string
	var inserted calendarapi.Event
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		methods = append(methods, request.Method)
		sendUpdates = append(sendUpdates, request.URL.Query().Get("sendUpdates"))
		w.Header().Set("Content-Type", "application/json")
		if request.Method == http.MethodPut {
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
			return
		}
		if err := json.NewDecoder(request.Body).Decode(&inserted); err != nil {
			t.Fatalf("decode inserted event: %v", err)
		}
		_, _ = w.Write([]byte(`{}`))
	}))
	t.Cleanup(server.Close)

	provider := NewGoogleProvider(googleTestOAuth{client: googleCalendarTestClient(t, server.URL)})
	blockID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	eventID := StableGoogleScheduleEventID(blockID)
	_, err := provider.UpsertScheduleEvent(context.Background(), ProviderToken{Scopes: []string{GoogleCalendarEventsOwnedScope}}, ExternalScheduleEventInput{
		CalendarID: "primary", EventID: eventID, BlockID: blockID, StoryID: storyID, WorkspaceID: workspaceID,
		Title: "Private project work", StartAt: time.Date(2026, 6, 15, 9, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC),
		PrivateProperties: map[string]string{"fortyone_block_id": blockID.String()},
	})
	if err != nil {
		t.Fatalf("UpsertScheduleEvent returned error: %v", err)
	}
	if strings.Join(methods, ",") != "PUT,POST" {
		t.Fatalf("expected update-then-insert, got %v", methods)
	}
	if strings.Join(sendUpdates, ",") != "none,none" {
		t.Fatalf("expected every schedule write to suppress Google attendee updates, got %v", sendUpdates)
	}
	if inserted.Id != eventID || inserted.Visibility != "private" || inserted.ExtendedProperties == nil {
		t.Fatalf("unexpected inserted event: %#v", inserted)
	}
	if inserted.Description != fortyOneMayaDescription {
		t.Fatalf("expected Maya provenance description, got %q", inserted.Description)
	}
	if len(inserted.Attendees) != 0 {
		t.Fatalf("Maya schedule events must not invite attendees: %#v", inserted.Attendees)
	}
	if inserted.ExtendedProperties.Private[fortyOneGoogleSourceKey] != fortyOneGoogleSourceValue {
		t.Fatalf("expected private FortyOne provenance: %#v", inserted.ExtendedProperties.Private)
	}
}

func TestGoogleProviderUpsertRecoversFromInsertRaceAndDeleteIsIdempotent(t *testing.T) {
	t.Parallel()

	var putCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPut:
			if putCalls.Add(1) == 1 {
				http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
				return
			}
			_, _ = w.Write([]byte(`{}`))
		case http.MethodPost:
			http.Error(w, `{"error":{"code":409,"message":"already exists"}}`, http.StatusConflict)
		case http.MethodDelete:
			http.Error(w, `{"error":{"code":404,"message":"not found"}}`, http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	provider := NewGoogleProvider(googleTestOAuth{client: googleCalendarTestClient(t, server.URL)})
	token := ProviderToken{Scopes: []string{GoogleCalendarEventsOwnedScope}}
	eventID := StableGoogleScheduleEventID(uuid.New())
	input := ExternalScheduleEventInput{CalendarID: "primary", EventID: eventID, StartAt: time.Now().UTC(), EndAt: time.Now().UTC().Add(time.Hour)}
	if _, err := provider.UpsertScheduleEvent(context.Background(), token, input); err != nil {
		t.Fatalf("expected insert race to retry update: %v", err)
	}
	if putCalls.Load() != 2 {
		t.Fatalf("expected two update attempts, got %d", putCalls.Load())
	}
	if err := provider.DeleteScheduleEvent(context.Background(), token, "primary", eventID); err != nil {
		t.Fatalf("expected missing delete to be idempotent: %v", err)
	}
}

func TestGoogleProviderRequiresOwnedEventsScope(t *testing.T) {
	t.Parallel()
	provider := NewGoogleProvider(googleTestOAuth{client: http.DefaultClient})
	_, err := provider.UpsertScheduleEvent(context.Background(), ProviderToken{Scopes: []string{GoogleCalendarEventsReadonlyScope}}, ExternalScheduleEventInput{})
	if !errors.Is(err, ErrCalendarReauthorizationRequired) {
		t.Fatalf("expected reauthorization error, got %v", err)
	}
}

func TestFortyOneScheduleEventsNeverBecomeExternalBusyWindows(t *testing.T) {
	t.Parallel()
	event := googleTestEvent(StableGoogleScheduleEventID(uuid.New()), time.Now().UTC())
	event.ExtendedProperties = &calendarapi.EventExtendedProperties{Private: map[string]string{
		fortyOneGoogleSourceKey: fortyOneGoogleSourceValue,
	}}
	if normalized, ok, err := googleEventToCalendarEvent("primary", event, "UTC"); err != nil || ok || normalized.ProviderEventID != "" {
		t.Fatalf("expected FortyOne event to be filtered before busy-window creation: %#v, %v", normalized, err)
	}
}

func TestGoogleDeltaReportsManagedEventDriftWithoutCreatingBusyTime(t *testing.T) {
	t.Parallel()
	startAt := time.Date(2026, 8, 17, 11, 0, 0, 0, time.UTC)
	eventID := StableGoogleScheduleEventID(uuid.New())
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		event := googleTestEvent(eventID, startAt)
		event.Summary = "User moved Maya work"
		event.Status = "confirmed"
		event.Visibility = "public"
		event.Transparency = "transparent"
		event.Attendees = []*calendarapi.EventAttendee{{Email: "guest@example.com"}}
		event.ExtendedProperties = &calendarapi.EventExtendedProperties{Private: map[string]string{
			fortyOneGoogleSourceKey: fortyOneGoogleSourceValue,
		}}
		_ = json.NewEncoder(w).Encode(&calendarapi.Events{Items: []*calendarapi.Event{event}, NextSyncToken: "next"})
	}))
	t.Cleanup(server.Close)
	provider := NewGoogleProvider(googleTestOAuth{client: googleCalendarTestClient(t, server.URL)})
	delta, err := provider.SyncCalendarChanges(context.Background(), ProviderToken{AccessToken: "token"}, "sync-token")
	if err != nil {
		t.Fatalf("SyncCalendarChanges returned error: %v", err)
	}
	if len(delta.ManagedScheduleEventChanges) != 1 {
		t.Fatalf("expected managed event drift signal: %#v", delta)
	}
	change := delta.ManagedScheduleEventChanges[0]
	if change.EventID != eventID || change.Deleted || change.Title != "User moved Maya work" || !change.StartAt.Equal(startAt) {
		t.Fatalf("unexpected managed event change: %#v", change)
	}
	if change.Visibility != "public" || change.Transparency != "transparent" || !change.HasAttendees || change.Source != fortyOneGoogleSourceValue || change.BlockID != "" {
		t.Fatalf("expected managed privacy/provenance drift to be preserved for reconciliation: %#v", change)
	}
	if len(delta.Events) != 0 || len(delta.BusyWindows) != 0 {
		t.Fatalf("managed event drift must not become external availability: %#v", delta)
	}
}

func TestGoogleManagedEventChangePreservesCanonicalPrivacyAndProvenance(t *testing.T) {
	t.Parallel()
	blockID := uuid.New()
	storyID := uuid.New()
	workspaceID := uuid.New()
	startAt := time.Date(2026, 8, 17, 11, 0, 0, 0, time.UTC)
	event := googleTestEvent(StableGoogleScheduleEventID(blockID), startAt)
	event.Status = "confirmed"
	event.Visibility = "private"
	event.Transparency = "opaque"
	event.ExtendedProperties = &calendarapi.EventExtendedProperties{Private: map[string]string{
		fortyOneGoogleSourceKey: fortyOneGoogleSourceValue,
		"fortyone_block_id":     blockID.String(),
		"fortyone_story_id":     storyID.String(),
		"fortyone_workspace_id": workspaceID.String(),
	}}
	change, err := googleManagedScheduleEventChange(event, "UTC")
	if err != nil {
		t.Fatalf("googleManagedScheduleEventChange returned error: %v", err)
	}
	if change.Visibility != "private" || change.Transparency != "opaque" || change.Status != "confirmed" || change.HasAttendees || change.Recurring ||
		change.Source != fortyOneGoogleSourceValue || change.BlockID != blockID.String() || change.StoryID != storyID.String() || change.WorkspaceID != workspaceID.String() {
		t.Fatalf("unexpected canonical managed event metadata: %#v", change)
	}
}

func googleCalendarTestClient(t *testing.T, serverURL string) *http.Client {
	t.Helper()
	base, err := url.Parse(serverURL)
	if err != nil {
		t.Fatalf("parse test server URL: %v", err)
	}
	return &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		clone := request.Clone(request.Context())
		clone.URL.Scheme = base.Scheme
		clone.URL.Host = base.Host
		return http.DefaultTransport.RoundTrip(clone)
	})}
}

func googleTestEvent(id string, startAt time.Time) *calendarapi.Event {
	return &calendarapi.Event{
		Id:      id,
		Summary: id,
		Start:   &calendarapi.EventDateTime{DateTime: startAt.Format(time.RFC3339)},
		End:     &calendarapi.EventDateTime{DateTime: startAt.Add(time.Hour).Format(time.RFC3339)},
	}
}
