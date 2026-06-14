package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	googleauth "github.com/complexus-tech/projects-api/pkg/google"
	"golang.org/x/oauth2"
	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const googleFreeBusyChunkSize = 30 * 24 * time.Hour

type GoogleOAuth interface {
	CalendarAuthCodeURL(state string) (string, error)
	ExchangeCalendarCode(ctx context.Context, code string) (googleauth.CalendarToken, error)
	CalendarHTTPClient(ctx context.Context, token *oauth2.Token) (*http.Client, error)
}

type GoogleProvider struct {
	google GoogleOAuth
}

func NewGoogleProvider(googleService GoogleOAuth) *GoogleProvider {
	return &GoogleProvider{google: googleService}
}

func (p *GoogleProvider) AuthCodeURL(state string) (string, error) {
	if p.google == nil {
		return "", ErrCalendarNotConfigured
	}
	return p.google.CalendarAuthCodeURL(state)
}

func (p *GoogleProvider) ExchangeCode(ctx context.Context, code string) (ProviderToken, error) {
	if p.google == nil {
		return ProviderToken{}, ErrCalendarNotConfigured
	}
	token, err := p.google.ExchangeCalendarCode(ctx, code)
	if err != nil {
		return ProviderToken{}, err
	}
	return ProviderToken{
		AccessToken:    token.Token.AccessToken,
		RefreshToken:   token.Token.RefreshToken,
		TokenType:      token.Token.TokenType,
		Expiry:         token.Token.Expiry,
		ConnectedEmail: token.Identity.Email,
		Timezone:       "UTC",
		Scopes:         token.Scopes,
	}, nil
}

func (p *GoogleProvider) ListBusyWindows(ctx context.Context, token ProviderToken, input BusyWindowInput) ([]CoreBusyWindow, error) {
	if p.google == nil {
		return nil, ErrCalendarNotConfigured
	}
	httpClient, err := p.google.CalendarHTTPClient(ctx, &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return nil, err
	}
	client, err := calendarapi.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return nil, err
	}
	if hasProviderScope(token.Scopes, googleCalendarEventsReadonlyScope) {
		windows, err := listGoogleEventWindows(ctx, client, input)
		if err == nil {
			return windows, nil
		}
		if !isGoogleInsufficientPermissionsError(err) {
			return nil, err
		}
	}
	return listGoogleFreeBusyWindows(ctx, client, input)
}

func listGoogleFreeBusyWindows(ctx context.Context, client *calendarapi.Service, input BusyWindowInput) ([]CoreBusyWindow, error) {
	timezone := fallbackTimezone(input.Timezone)
	ranges := googleFreeBusyRanges(input.TimeMin, input.TimeMax)
	windows := []CoreBusyWindow{}
	for _, timeRange := range ranges {
		response, err := client.Freebusy.Query(&calendarapi.FreeBusyRequest{
			TimeMin:  timeRange.start.Format(time.RFC3339),
			TimeMax:  timeRange.end.Format(time.RFC3339),
			TimeZone: timezone,
			Items: []*calendarapi.FreeBusyRequestItem{
				{Id: "primary"},
			},
		}).Context(ctx).Do()
		if err != nil {
			return nil, err
		}
		primary, ok := response.Calendars["primary"]
		if !ok {
			continue
		}
		for index, busy := range primary.Busy {
			startAt, err := time.Parse(time.RFC3339, busy.Start)
			if err != nil {
				return nil, err
			}
			endAt, err := time.Parse(time.RFC3339, busy.End)
			if err != nil {
				return nil, err
			}
			eventID := googleFreeBusyEventID(startAt, endAt, index)
			windows = append(windows, CoreBusyWindow{
				ConnectionID:    input.ConnectionID,
				WorkspaceID:     input.WorkspaceID,
				UserID:          input.UserID,
				Provider:        ProviderGoogle,
				ProviderEventID: eventID,
				StartAt:         startAt,
				EndAt:           endAt,
				Status:          BusyStatusBusy,
				Transparency:    BusyTransparencyOpaque,
				IsPrivate:       true,
				SourceHash:      eventID,
			})
		}
	}
	return windows, nil
}

func listGoogleEventWindows(ctx context.Context, client *calendarapi.Service, input BusyWindowInput) ([]CoreBusyWindow, error) {
	ranges := googleFreeBusyRanges(input.TimeMin, input.TimeMax)
	windows := []CoreBusyWindow{}
	for _, timeRange := range ranges {
		response, err := client.Events.List("primary").
			Context(ctx).
			ShowDeleted(false).
			SingleEvents(true).
			OrderBy("startTime").
			TimeMin(timeRange.start.Format(time.RFC3339)).
			TimeMax(timeRange.end.Format(time.RFC3339)).
			Do()
		if err != nil {
			return nil, err
		}
		for _, event := range response.Items {
			window, ok, err := googleEventToBusyWindow("primary", event)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			window.ConnectionID = input.ConnectionID
			window.WorkspaceID = input.WorkspaceID
			window.UserID = input.UserID
			window.Provider = ProviderGoogle
			windows = append(windows, window)
		}
	}
	return windows, nil
}

func googleEventToBusyWindow(calendarID string, event *calendarapi.Event) (CoreBusyWindow, bool, error) {
	if event == nil || strings.TrimSpace(event.Id) == "" || event.Status == "cancelled" {
		return CoreBusyWindow{}, false, nil
	}
	if event.Transparency == "transparent" {
		return CoreBusyWindow{}, false, nil
	}
	startAt, err := googleEventTime(event.Start)
	if err != nil {
		return CoreBusyWindow{}, false, err
	}
	endAt, err := googleEventTime(event.End)
	if err != nil {
		return CoreBusyWindow{}, false, err
	}
	if !endAt.After(startAt) {
		return CoreBusyWindow{}, false, nil
	}

	calendarID = strings.TrimSpace(calendarID)
	if calendarID == "" {
		calendarID = "primary"
	}
	providerEventID := calendarID + ":" + event.Id
	isPrivate := event.Visibility == "private" || event.Visibility == "confidential"
	var title *string
	if !isPrivate {
		summary := strings.TrimSpace(event.Summary)
		if summary != "" {
			title = &summary
		}
	}
	sourceHash := googleEventSourceHash(providerEventID, startAt, endAt, title, isPrivate)
	return CoreBusyWindow{
		Provider:        ProviderGoogle,
		ProviderEventID: providerEventID,
		CalendarID:      &calendarID,
		Title:           title,
		StartAt:         startAt,
		EndAt:           endAt,
		Status:          BusyStatusBusy,
		Transparency:    BusyTransparencyOpaque,
		IsPrivate:       isPrivate,
		SourceHash:      sourceHash,
	}, true, nil
}

func googleEventTime(value *calendarapi.EventDateTime) (time.Time, error) {
	if value == nil {
		return time.Time{}, fmt.Errorf("google calendar event is missing time")
	}
	if strings.TrimSpace(value.DateTime) != "" {
		return time.Parse(time.RFC3339, value.DateTime)
	}
	if strings.TrimSpace(value.Date) != "" {
		return time.Parse("2006-01-02", value.Date)
	}
	return time.Time{}, fmt.Errorf("google calendar event is missing time")
}

type googleFreeBusyRange struct {
	start time.Time
	end   time.Time
}

func googleFreeBusyRanges(timeMin, timeMax time.Time) []googleFreeBusyRange {
	if !timeMax.After(timeMin) {
		return []googleFreeBusyRange{}
	}
	ranges := []googleFreeBusyRange{}
	for start := timeMin; start.Before(timeMax); start = start.Add(googleFreeBusyChunkSize) {
		end := start.Add(googleFreeBusyChunkSize)
		if end.After(timeMax) {
			end = timeMax
		}
		ranges = append(ranges, googleFreeBusyRange{start: start, end: end})
	}
	return ranges
}

func googleFreeBusyEventID(startAt, endAt time.Time, index int) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%d", startAt.UTC().Format(time.RFC3339Nano), endAt.UTC().Format(time.RFC3339Nano), index)))
	return hex.EncodeToString(hash[:])
}

func googleEventSourceHash(providerEventID string, startAt, endAt time.Time, title *string, isPrivate bool) string {
	titleValue := ""
	if title != nil {
		titleValue = *title
	}
	hash := sha256.Sum256([]byte(fmt.Sprintf("%s:%s:%s:%s:%t", providerEventID, startAt.UTC().Format(time.RFC3339Nano), endAt.UTC().Format(time.RFC3339Nano), titleValue, isPrivate)))
	return hex.EncodeToString(hash[:])
}

func isGoogleInsufficientPermissionsError(err error) bool {
	var googleErr *googleapi.Error
	if !errors.As(err, &googleErr) {
		return false
	}
	if googleErr.Code != http.StatusForbidden {
		return false
	}
	for _, item := range googleErr.Errors {
		if item.Reason == "insufficientPermissions" {
			return true
		}
	}
	return strings.Contains(googleErr.Message, "Insufficient Permission")
}
