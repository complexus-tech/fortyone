package calendar

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	googleauth "github.com/complexus-tech/projects-api/pkg/google"
	"golang.org/x/oauth2"
	calendarapi "google.golang.org/api/calendar/v3"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

const (
	googleFreeBusyChunkSize     = 30 * 24 * time.Hour
	maxCalendarAttendees        = 100
	fortyOneGoogleEventIDPrefix = "f41sched"
	fortyOneGoogleSourceKey     = "fortyone_source"
	fortyOneGoogleSourceValue   = "maya_schedule"
	fortyOneMayaDescription     = "Scheduled by Maya in FortyOne"
)

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
		AccessToken:       token.Token.AccessToken,
		RefreshToken:      token.Token.RefreshToken,
		TokenType:         token.Token.TokenType,
		Expiry:            token.Token.Expiry,
		ProviderAccountID: token.Identity.Subject,
		ConnectedEmail:    token.Identity.Email,
		Timezone:          "UTC",
		Scopes:            token.Scopes,
	}, nil
}

func (p *GoogleProvider) SyncCalendar(ctx context.Context, token ProviderToken, input BusyWindowInput) (CalendarSyncSnapshot, error) {
	if p.google == nil {
		return CalendarSyncSnapshot{}, ErrCalendarNotConfigured
	}
	httpClient, err := p.google.CalendarHTTPClient(ctx, &oauth2.Token{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
		TokenType:    token.TokenType,
		Expiry:       token.Expiry,
	})
	if err != nil {
		return CalendarSyncSnapshot{}, err
	}
	client, err := calendarapi.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		return CalendarSyncSnapshot{}, err
	}
	if hasProviderScope(token.Scopes, GoogleCalendarEventsReadonlyScope) {
		snapshot, err := listGoogleCalendarSnapshot(ctx, client, input)
		if err == nil {
			snapshot.CanReadEventDetails = true
			return snapshot, nil
		}
		if !isGoogleInsufficientPermissionsError(err) {
			return CalendarSyncSnapshot{}, err
		}
	}
	windows, err := listGoogleFreeBusyWindows(ctx, client, input)
	if err != nil {
		return CalendarSyncSnapshot{}, err
	}
	return CalendarSyncSnapshot{
		Events:              []CoreCalendarEvent{},
		BusyWindows:         windows,
		CanReadEventDetails: false,
		Timezone:            normalizedCalendarTimezone(input.Timezone, "UTC"),
	}, nil
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
				CalendarID:      stringPointer("primary"),
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

func listGoogleCalendarSnapshot(ctx context.Context, client *calendarapi.Service, input BusyWindowInput) (CalendarSyncSnapshot, error) {
	events := []CoreCalendarEvent{}
	windows := []CoreBusyWindow{}
	eventIndexes := make(map[string]int)
	windowIndexes := make(map[string]int)
	snapshotTimezone := normalizedCalendarTimezone(input.Timezone, "UTC")
	call := client.Events.List("primary").
		Context(ctx).
		ShowDeleted(false).
		SingleEvents(true).
		OrderBy("startTime").
		TimeMin(input.TimeMin.Format(time.RFC3339)).
		TimeMax(input.TimeMax.Format(time.RFC3339))
	nextSyncToken := ""
	for {
		response, err := call.Do()
		if err != nil {
			return CalendarSyncSnapshot{}, err
		}
		if responseTimezone := strings.TrimSpace(response.TimeZone); responseTimezone != "" {
			snapshotTimezone = normalizedCalendarTimezone(responseTimezone, snapshotTimezone)
		}
		for _, googleEvent := range response.Items {
			if isFortyOneScheduleEvent(googleEvent) {
				continue
			}
			event, ok, err := googleEventToCalendarEvent("primary", googleEvent, snapshotTimezone)
			if err != nil {
				return CalendarSyncSnapshot{}, err
			}
			if !ok {
				continue
			}
			event.ConnectionID = input.ConnectionID
			event.WorkspaceID = input.WorkspaceID
			event.UserID = input.UserID
			event.Provider = ProviderGoogle
			key := event.CalendarID + ":" + event.ProviderEventID
			if index, exists := eventIndexes[key]; exists {
				events[index] = event
			} else {
				eventIndexes[key] = len(events)
				events = append(events, event)
			}

			window, blocksTime := calendarEventToBusyWindow(event)
			if !blocksTime {
				continue
			}
			windowKey := window.ProviderEventID
			if index, exists := windowIndexes[windowKey]; exists {
				windows[index] = window
			} else {
				windowIndexes[windowKey] = len(windows)
				windows = append(windows, window)
			}
		}
		if strings.TrimSpace(response.NextPageToken) == "" {
			nextSyncToken = strings.TrimSpace(response.NextSyncToken)
			break
		}
		call.PageToken(response.NextPageToken)
	}
	return CalendarSyncSnapshot{
		Events:        events,
		BusyWindows:   windows,
		NextSyncToken: nextSyncToken,
		Timezone:      snapshotTimezone,
	}, nil
}

func (p *GoogleProvider) SyncCalendarChanges(ctx context.Context, token ProviderToken, syncToken string) (CalendarSyncDelta, error) {
	client, err := p.calendarClient(ctx, token)
	if err != nil {
		return CalendarSyncDelta{}, err
	}
	call := client.Events.List("primary").Context(ctx).ShowDeleted(true).SingleEvents(true).SyncToken(strings.TrimSpace(syncToken))
	delta := CalendarSyncDelta{
		Events: []CoreCalendarEvent{}, BusyWindows: []CoreBusyWindow{}, DeletedEventIDs: []string{},
		ManagedScheduleEventChanges: []ManagedScheduleEventChange{},
	}
	for {
		response, err := call.Do()
		if err != nil {
			var apiErr *googleapi.Error
			if errors.As(err, &apiErr) && apiErr.Code == http.StatusGone {
				return CalendarSyncDelta{}, ErrCalendarFullSyncRequired
			}
			return CalendarSyncDelta{}, err
		}
		for _, googleEvent := range response.Items {
			if googleEvent == nil || strings.TrimSpace(googleEvent.Id) == "" {
				continue
			}
			if isFortyOneScheduleEvent(googleEvent) {
				change, err := googleManagedScheduleEventChange(googleEvent, response.TimeZone)
				if err != nil {
					return CalendarSyncDelta{}, err
				}
				delta.ManagedScheduleEventChanges = append(delta.ManagedScheduleEventChanges, change)
				continue
			}
			if googleEvent.Status == "cancelled" {
				delta.DeletedEventIDs = append(delta.DeletedEventIDs, "primary:"+googleEvent.Id)
				continue
			}
			event, ok, err := googleEventToCalendarEvent("primary", googleEvent, response.TimeZone)
			if err != nil {
				return CalendarSyncDelta{}, err
			}
			if !ok {
				continue
			}
			delta.Events = append(delta.Events, event)
			if window, blocksTime := calendarEventToBusyWindow(event); blocksTime {
				delta.BusyWindows = append(delta.BusyWindows, window)
			}
		}
		if strings.TrimSpace(response.NextPageToken) == "" {
			delta.NextSyncToken = strings.TrimSpace(response.NextSyncToken)
			break
		}
		call.PageToken(response.NextPageToken)
	}
	return delta, nil
}

func googleManagedScheduleEventChange(event *calendarapi.Event, fallbackTimezone string) (ManagedScheduleEventChange, error) {
	change := ManagedScheduleEventChange{
		EventID:      strings.TrimSpace(event.Id),
		Deleted:      event.Status == "cancelled",
		Visibility:   strings.TrimSpace(event.Visibility),
		Transparency: strings.TrimSpace(event.Transparency),
		Status:       strings.TrimSpace(event.Status),
		HasAttendees: len(event.Attendees) > 0 || event.AttendeesOmitted,
		Recurring:    strings.TrimSpace(event.RecurringEventId) != "" || len(event.Recurrence) > 0,
	}
	if event.ExtendedProperties != nil {
		change.Source = strings.TrimSpace(event.ExtendedProperties.Private[fortyOneGoogleSourceKey])
		change.BlockID = strings.TrimSpace(event.ExtendedProperties.Private["fortyone_block_id"])
		change.StoryID = strings.TrimSpace(event.ExtendedProperties.Private["fortyone_story_id"])
		change.WorkspaceID = strings.TrimSpace(event.ExtendedProperties.Private["fortyone_workspace_id"])
	}
	if change.Deleted {
		return change, nil
	}
	startAt, _, err := googleEventTime(event.Start, fallbackTimezone)
	if err != nil {
		return ManagedScheduleEventChange{}, err
	}
	endAt, _, err := googleEventTime(event.End, fallbackTimezone)
	if err != nil {
		return ManagedScheduleEventChange{}, err
	}
	if !endAt.After(startAt) {
		return ManagedScheduleEventChange{}, ErrInvalidScheduleRange
	}
	change.Title = strings.TrimSpace(event.Summary)
	change.StartAt = startAt.UTC()
	change.EndAt = endAt.UTC()
	return change, nil
}

func (p *GoogleProvider) WatchCalendar(ctx context.Context, token ProviderToken, input CalendarWatchInput) (CalendarWatchChannel, error) {
	client, err := p.calendarClient(ctx, token)
	if err != nil {
		return CalendarWatchChannel{}, err
	}
	ttlSeconds := int64(input.TTL / time.Second)
	channel := &calendarapi.Channel{
		Id:      strings.TrimSpace(input.ChannelID),
		Type:    "web_hook",
		Address: strings.TrimSpace(input.Address),
		Token:   strings.TrimSpace(input.Token),
	}
	if ttlSeconds > 0 {
		channel.Params = map[string]string{"ttl": strconv.FormatInt(ttlSeconds, 10)}
	}
	created, err := client.Events.Watch("primary", channel).Context(ctx).Do()
	if err != nil {
		return CalendarWatchChannel{}, err
	}
	return CalendarWatchChannel{
		ChannelID:  strings.TrimSpace(created.Id),
		ResourceID: strings.TrimSpace(created.ResourceId),
		ExpiresAt:  time.UnixMilli(created.Expiration).UTC(),
	}, nil
}

func (p *GoogleProvider) StopCalendarWatch(ctx context.Context, token ProviderToken, channel CalendarWatchChannel) error {
	client, err := p.calendarClient(ctx, token)
	if err != nil {
		return err
	}
	return client.Channels.Stop(&calendarapi.Channel{Id: channel.ChannelID, ResourceId: channel.ResourceID}).Context(ctx).Do()
}

func (p *GoogleProvider) UpsertScheduleEvent(ctx context.Context, token ProviderToken, input ExternalScheduleEventInput) (ExternalScheduleEventResult, error) {
	if !hasProviderScope(token.Scopes, GoogleCalendarEventsOwnedScope) {
		return ExternalScheduleEventResult{}, ErrCalendarReauthorizationRequired
	}
	client, err := p.calendarClient(ctx, token)
	if err != nil {
		return ExternalScheduleEventResult{}, err
	}
	calendarID := strings.TrimSpace(input.CalendarID)
	if calendarID == "" {
		calendarID = "primary"
	}
	eventID := strings.TrimSpace(input.EventID)
	if eventID == "" || !strings.HasPrefix(eventID, fortyOneGoogleEventIDPrefix) || !input.EndAt.After(input.StartAt) {
		return ExternalScheduleEventResult{}, ErrInvalidScheduleBlock
	}
	privateProperties := make(map[string]string, len(input.PrivateProperties)+1)
	for key, value := range input.PrivateProperties {
		privateProperties[key] = value
	}
	privateProperties[fortyOneGoogleSourceKey] = fortyOneGoogleSourceValue
	event := &calendarapi.Event{
		Id:                 eventID,
		Summary:            strings.TrimSpace(input.Title),
		Description:        fortyOneMayaDescription,
		Status:             "confirmed",
		Transparency:       "opaque",
		Visibility:         "private",
		Start:              &calendarapi.EventDateTime{DateTime: input.StartAt.UTC().Format(time.RFC3339)},
		End:                &calendarapi.EventDateTime{DateTime: input.EndAt.UTC().Format(time.RFC3339)},
		ExtendedProperties: &calendarapi.EventExtendedProperties{Private: privateProperties},
	}
	_, err = client.Events.Update(calendarID, eventID, event).SendUpdates("none").Context(ctx).Do()
	if err == nil {
		return ExternalScheduleEventResult{EventID: eventID}, nil
	}
	var apiErr *googleapi.Error
	if !errors.As(err, &apiErr) || apiErr.Code != http.StatusNotFound {
		return ExternalScheduleEventResult{}, err
	}
	_, err = client.Events.Insert(calendarID, event).SendUpdates("none").Context(ctx).Do()
	if err == nil {
		return ExternalScheduleEventResult{EventID: eventID}, nil
	}
	if errors.As(err, &apiErr) && apiErr.Code == http.StatusConflict {
		_, err = client.Events.Update(calendarID, eventID, event).SendUpdates("none").Context(ctx).Do()
	}
	if err != nil {
		return ExternalScheduleEventResult{}, err
	}
	return ExternalScheduleEventResult{EventID: eventID}, nil
}

func (p *GoogleProvider) DeleteScheduleEvent(ctx context.Context, token ProviderToken, calendarID, eventID string) error {
	if !hasProviderScope(token.Scopes, GoogleCalendarEventsOwnedScope) {
		return ErrCalendarReauthorizationRequired
	}
	client, err := p.calendarClient(ctx, token)
	if err != nil {
		return err
	}
	calendarID = strings.TrimSpace(calendarID)
	if calendarID == "" {
		calendarID = "primary"
	}
	err = client.Events.Delete(calendarID, strings.TrimSpace(eventID)).SendUpdates("none").Context(ctx).Do()
	if err == nil {
		return nil
	}
	var apiErr *googleapi.Error
	if errors.As(err, &apiErr) && (apiErr.Code == http.StatusNotFound || apiErr.Code == http.StatusGone) {
		return nil
	}
	return err
}

func (p *GoogleProvider) calendarClient(ctx context.Context, token ProviderToken) (*calendarapi.Service, error) {
	if p.google == nil {
		return nil, ErrCalendarNotConfigured
	}
	httpClient, err := p.google.CalendarHTTPClient(ctx, &oauth2.Token{
		AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.TokenType, Expiry: token.Expiry,
	})
	if err != nil {
		return nil, err
	}
	return calendarapi.NewService(ctx, option.WithHTTPClient(httpClient))
}
