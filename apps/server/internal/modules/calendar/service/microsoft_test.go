package calendar

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type microsoftTestOAuth struct {
	client *http.Client
}

func (m microsoftTestOAuth) CalendarAuthCodeURL(string) (string, error) {
	return "https://login.microsoftonline.com/test", nil
}

func (m microsoftTestOAuth) ExchangeCalendarCode(context.Context, string) (microsoft.CalendarToken, error) {
	return microsoft.CalendarToken{}, nil
}

func (m microsoftTestOAuth) CalendarHTTPClient(context.Context, *oauth2.Token) (*http.Client, error) {
	return m.client, nil
}

func (m microsoftTestOAuth) RefreshCalendarToken(_ context.Context, token *oauth2.Token) (*oauth2.Token, error) {
	return token, nil
}

func TestMicrosoftProviderSyncMapsDetailsPrivacyAndAvailability(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/me/calendarView/delta" {
			t.Fatalf("unexpected Graph path: %s", request.URL.Path)
		}
		if request.URL.Query().Get("startDateTime") == "" || request.URL.Query().Get("endDateTime") == "" {
			t.Fatal("calendar delta request is missing its sync window")
		}
		if !strings.Contains(request.Header.Get("Prefer"), `IdType="ImmutableId"`) {
			t.Fatalf("immutable IDs were not requested: %s", request.Header.Get("Prefer"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"value": [
				{"id":"public-event","subject":"Planning","body":{"contentType":"text","content":"Roadmap"},"start":{"dateTime":"2026-08-21T09:00:00","timeZone":"UTC"},"end":{"dateTime":"2026-08-21T10:00:00","timeZone":"UTC"},"showAs":"busy","sensitivity":"normal","webLink":"https://outlook.office.com/calendar/item","organizer":{"emailAddress":{"name":"Ada","address":"ada@example.com"}}},
				{"id":"private-event","subject":"Secret","start":{"dateTime":"2026-08-21T11:00:00","timeZone":"UTC"},"end":{"dateTime":"2026-08-21T12:00:00","timeZone":"UTC"},"showAs":"busy","sensitivity":"private"},
				{"id":"free-event","subject":"Optional","start":{"dateTime":"2026-08-21T13:00:00","timeZone":"UTC"},"end":{"dateTime":"2026-08-21T14:00:00","timeZone":"UTC"},"showAs":"free","sensitivity":"normal"},
				{"id":"managed-event","subject":"Focus","body":{"contentType":"text","content":"Managed by FortyOne\nfortyone_source:maya_schedule\nfortyone_block_id:00000000-0000-0000-0000-000000000001"},"start":{"dateTime":"2026-08-21T15:00:00","timeZone":"UTC"},"end":{"dateTime":"2026-08-21T16:00:00","timeZone":"UTC"},"showAs":"busy","sensitivity":"private"}
			],
			"@odata.deltaLink":"https://graph.microsoft.com/v1.0/me/calendarView/delta?$deltatoken=next"
		}`))
	}))
	t.Cleanup(server.Close)

	provider := NewMicrosoftProvider(microsoftTestOAuth{client: server.Client()})
	provider.graphURL = server.URL
	snapshot, err := provider.SyncCalendar(context.Background(), ProviderToken{}, BusyWindowInput{
		TimeMin: time.Date(2026, 8, 21, 0, 0, 0, 0, time.UTC),
		TimeMax: time.Date(2026, 8, 22, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("SyncCalendar returned error: %v", err)
	}
	if len(snapshot.Events) != 3 || len(snapshot.BusyWindows) != 2 {
		t.Fatalf("unexpected imported event counts: events=%d busy=%d", len(snapshot.Events), len(snapshot.BusyWindows))
	}
	if snapshot.Events[0].Provider != ProviderMicrosoft || snapshot.Events[0].ProviderEventID != "primary:public-event" {
		t.Fatalf("unexpected public event mapping: %#v", snapshot.Events[0])
	}
	if snapshot.Events[1].Title != nil || !snapshot.Events[1].IsPrivate {
		t.Fatalf("private event details were not redacted: %#v", snapshot.Events[1])
	}
	if snapshot.Events[2].BlocksTime {
		t.Fatalf("a free Outlook event must not block availability: %#v", snapshot.Events[2])
	}
	if !strings.Contains(snapshot.NextSyncToken, "$deltatoken=next") {
		t.Fatalf("unexpected delta link: %s", snapshot.NextSyncToken)
	}
}

func TestMicrosoftProviderCreatesPrivateManagedScheduleEvent(t *testing.T) {
	t.Parallel()

	var inserted microsoftScheduleEvent
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodPost || request.URL.Path != "/me/events" {
			t.Fatalf("unexpected Graph request: %s %s", request.Method, request.URL.Path)
		}
		if err := json.NewDecoder(request.Body).Decode(&inserted); err != nil {
			t.Fatalf("decode schedule event: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"id":"immutable-outlook-event"}`))
	}))
	t.Cleanup(server.Close)

	provider := NewMicrosoftProvider(microsoftTestOAuth{client: server.Client()})
	provider.graphURL = server.URL
	blockID := uuid.New()
	result, err := provider.UpsertScheduleEvent(context.Background(), ProviderToken{
		Scopes: []string{MicrosoftCalendarReadWriteScope},
	}, ExternalScheduleEventInput{
		EventID: "pending:" + blockID.String(), BlockID: blockID, StoryID: uuid.New(), WorkspaceID: uuid.New(),
		Title: "Project work", StartAt: time.Date(2026, 8, 21, 9, 0, 0, 0, time.UTC), EndAt: time.Date(2026, 8, 21, 10, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("UpsertScheduleEvent returned error: %v", err)
	}
	if result.EventID != "immutable-outlook-event" {
		t.Fatalf("unexpected provider event ID: %s", result.EventID)
	}
	if inserted.Sensitivity != "private" || inserted.ShowAs != "busy" || inserted.TransactionID != blockID.String() {
		t.Fatalf("unexpected schedule event privacy or idempotency fields: %#v", inserted)
	}
	if len(inserted.Categories) != 1 || inserted.Categories[0] != microsoftManagedCategory || !strings.Contains(inserted.Body.Content, blockID.String()) {
		t.Fatalf("managed event provenance is incomplete: %#v", inserted)
	}
}
