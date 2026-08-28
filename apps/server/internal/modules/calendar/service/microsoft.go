package calendar

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/pkg/microsoft"
	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

const (
	microsoftGraphBaseURL       = "https://graph.microsoft.com/v1.0"
	microsoftManagedCategory    = "FortyOne"
	microsoftManagedDescription = "Managed by FortyOne"
	microsoftWatchTTL           = 6 * 24 * time.Hour
)

type MicrosoftCalendarOAuth interface {
	CalendarAuthCodeURL(state string) (string, error)
	ExchangeCalendarCode(ctx context.Context, code string) (microsoft.CalendarToken, error)
	CalendarHTTPClient(ctx context.Context, token *oauth2.Token) (*http.Client, error)
	RefreshCalendarToken(ctx context.Context, token *oauth2.Token) (*oauth2.Token, error)
}

type MicrosoftProvider struct {
	microsoft MicrosoftCalendarOAuth
	graphURL  string
}

func NewMicrosoftProvider(service MicrosoftCalendarOAuth) *MicrosoftProvider {
	return &MicrosoftProvider{microsoft: service, graphURL: microsoftGraphBaseURL}
}

func (p *MicrosoftProvider) AuthCodeURL(state string) (string, error) {
	if p.microsoft == nil {
		return "", ErrCalendarNotConfigured
	}
	return p.microsoft.CalendarAuthCodeURL(state)
}

func (p *MicrosoftProvider) ExchangeCode(ctx context.Context, code string) (ProviderToken, error) {
	if p.microsoft == nil {
		return ProviderToken{}, ErrCalendarNotConfigured
	}
	calendarToken, err := p.microsoft.ExchangeCalendarCode(ctx, code)
	if err != nil {
		return ProviderToken{}, err
	}
	if calendarToken.Token == nil {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	return ProviderToken{
		AccessToken:       calendarToken.Token.AccessToken,
		RefreshToken:      calendarToken.Token.RefreshToken,
		TokenType:         calendarToken.Token.TokenType,
		Expiry:            calendarToken.Token.Expiry,
		ProviderAccountID: strings.TrimSpace(calendarToken.AccountID),
		ConnectedEmail:    strings.TrimSpace(calendarToken.Email),
		Timezone:          "UTC",
		Scopes:            append([]string(nil), calendarToken.Scopes...),
	}, nil
}

func (p *MicrosoftProvider) RefreshToken(ctx context.Context, token ProviderToken) (ProviderToken, error) {
	if p.microsoft == nil {
		return ProviderToken{}, ErrCalendarNotConfigured
	}
	refreshed, err := p.microsoft.RefreshCalendarToken(ctx, &oauth2.Token{
		AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.TokenType, Expiry: token.Expiry,
	})
	if err != nil {
		return ProviderToken{}, err
	}
	token.AccessToken = refreshed.AccessToken
	token.RefreshToken = refreshed.RefreshToken
	token.TokenType = refreshed.TokenType
	token.Expiry = refreshed.Expiry
	return token, nil
}

func (p *MicrosoftProvider) SyncCalendar(ctx context.Context, token ProviderToken, input BusyWindowInput) (CalendarSyncSnapshot, error) {
	startURL, err := url.Parse(strings.TrimRight(p.graphURL, "/") + "/me/calendarView/delta")
	if err != nil {
		return CalendarSyncSnapshot{}, err
	}
	query := startURL.Query()
	query.Set("startDateTime", input.TimeMin.UTC().Format(time.RFC3339))
	query.Set("endDateTime", input.TimeMax.UTC().Format(time.RFC3339))
	startURL.RawQuery = query.Encode()

	events, windows, _, _, nextToken, err := p.readCalendarDelta(ctx, token, startURL.String())
	if err != nil {
		return CalendarSyncSnapshot{}, err
	}
	return CalendarSyncSnapshot{
		Events: events, BusyWindows: windows, CanReadEventDetails: true,
		NextSyncToken: nextToken, Timezone: "UTC",
	}, nil
}

func (p *MicrosoftProvider) SyncCalendarChanges(ctx context.Context, token ProviderToken, syncToken string) (CalendarSyncDelta, error) {
	syncToken = strings.TrimSpace(syncToken)
	if syncToken == "" {
		return CalendarSyncDelta{}, ErrCalendarFullSyncRequired
	}
	events, windows, deleted, managed, nextToken, err := p.readCalendarDelta(ctx, token, syncToken)
	if err != nil {
		var graphErr *MicrosoftGraphError
		if errors.As(err, &graphErr) && (graphErr.StatusCode == http.StatusGone || graphErr.Code == "SyncStateNotFound") {
			return CalendarSyncDelta{}, ErrCalendarFullSyncRequired
		}
		return CalendarSyncDelta{}, err
	}
	return CalendarSyncDelta{
		Events: events, BusyWindows: windows, DeletedEventIDs: deleted,
		ManagedScheduleEventChanges: managed, NextSyncToken: nextToken,
	}, nil
}

func (p *MicrosoftProvider) readCalendarDelta(ctx context.Context, token ProviderToken, requestURL string) ([]CoreCalendarEvent, []CoreBusyWindow, []string, []ManagedScheduleEventChange, string, error) {
	events := []CoreCalendarEvent{}
	windows := []CoreBusyWindow{}
	deleted := []string{}
	managed := []ManagedScheduleEventChange{}
	nextURL := strings.TrimSpace(requestURL)
	for nextURL != "" {
		var page microsoftEventPage
		if err := p.doJSON(ctx, token, http.MethodGet, nextURL, nil, &page); err != nil {
			return nil, nil, nil, nil, "", err
		}
		for index := range page.Value {
			item := page.Value[index]
			if strings.TrimSpace(item.ID) == "" {
				continue
			}
			providerEventID := microsoftProviderEventID(item.ID)
			if item.Removed != nil || item.IsCancelled {
				deleted = append(deleted, providerEventID)
				managed = append(managed, ManagedScheduleEventChange{EventID: item.ID, Deleted: true})
				continue
			}
			if metadata := microsoftManagedMetadata(item.Body.Content); metadata != nil {
				change, changeErr := microsoftManagedScheduleEventChange(item, metadata)
				if changeErr != nil {
					return nil, nil, nil, nil, "", changeErr
				}
				managed = append(managed, change)
				continue
			}
			event, ok, err := microsoftEventToCalendarEvent(item)
			if err != nil {
				return nil, nil, nil, nil, "", err
			}
			if !ok {
				continue
			}
			events = append(events, event)
			if window, blocksTime := calendarEventToBusyWindow(event); blocksTime {
				windows = append(windows, window)
			}
		}
		if strings.TrimSpace(page.NextLink) != "" {
			nextURL = page.NextLink
			continue
		}
		nextURL = ""
		if strings.TrimSpace(page.DeltaLink) == "" {
			return nil, nil, nil, nil, "", errors.New("microsoft calendar delta response did not include a delta link")
		}
		return events, windows, deleted, managed, page.DeltaLink, nil
	}
	return events, windows, deleted, managed, "", errors.New("microsoft calendar delta URL is empty")
}

func (p *MicrosoftProvider) WatchCalendar(ctx context.Context, token ProviderToken, input CalendarWatchInput) (CalendarWatchChannel, error) {
	expiresAt := time.Now().UTC().Add(microsoftWatchTTL)
	if input.TTL > 0 && input.TTL < microsoftWatchTTL {
		expiresAt = time.Now().UTC().Add(input.TTL)
	}
	payload := microsoftSubscriptionRequest{
		ChangeType: "created,updated,deleted", NotificationURL: strings.TrimSpace(input.Address),
		Resource: "me/events", ExpirationDateTime: expiresAt, ClientState: strings.TrimSpace(input.Token),
	}
	var subscription microsoftSubscription
	if err := p.doJSON(ctx, token, http.MethodPost, strings.TrimRight(p.graphURL, "/")+"/subscriptions", payload, &subscription); err != nil {
		return CalendarWatchChannel{}, err
	}
	return microsoftSubscriptionChannel(subscription)
}

func (p *MicrosoftProvider) RenewCalendarWatch(ctx context.Context, token ProviderToken, channel CalendarWatchChannel, input CalendarWatchInput) (CalendarWatchChannel, error) {
	if strings.TrimSpace(channel.ChannelID) == "" {
		return p.WatchCalendar(ctx, token, input)
	}
	payload := struct {
		ExpirationDateTime time.Time `json:"expirationDateTime"`
	}{ExpirationDateTime: time.Now().UTC().Add(microsoftWatchTTL)}
	var subscription microsoftSubscription
	requestURL := strings.TrimRight(p.graphURL, "/") + "/subscriptions/" + url.PathEscape(channel.ChannelID)
	if err := p.doJSON(ctx, token, http.MethodPatch, requestURL, payload, &subscription); err != nil {
		var graphErr *MicrosoftGraphError
		if errors.As(err, &graphErr) && graphErr.StatusCode == http.StatusNotFound {
			return p.WatchCalendar(ctx, token, input)
		}
		return CalendarWatchChannel{}, err
	}
	return microsoftSubscriptionChannel(subscription)
}

func (p *MicrosoftProvider) StopCalendarWatch(ctx context.Context, token ProviderToken, channel CalendarWatchChannel) error {
	if strings.TrimSpace(channel.ChannelID) == "" {
		return nil
	}
	requestURL := strings.TrimRight(p.graphURL, "/") + "/subscriptions/" + url.PathEscape(channel.ChannelID)
	err := p.doJSON(ctx, token, http.MethodDelete, requestURL, nil, nil)
	var graphErr *MicrosoftGraphError
	if errors.As(err, &graphErr) && graphErr.StatusCode == http.StatusNotFound {
		return nil
	}
	return err
}

func (p *MicrosoftProvider) UpsertScheduleEvent(ctx context.Context, token ProviderToken, input ExternalScheduleEventInput) (ExternalScheduleEventResult, error) {
	if !hasProviderScope(token.Scopes, MicrosoftCalendarReadWriteScope) {
		return ExternalScheduleEventResult{}, ErrCalendarReauthorizationRequired
	}
	if input.BlockID == uuid.Nil || strings.TrimSpace(input.Title) == "" || !input.EndAt.After(input.StartAt) {
		return ExternalScheduleEventResult{}, ErrInvalidScheduleBlock
	}
	payload := microsoftScheduleEvent{
		Subject:       strings.TrimSpace(input.Title),
		Body:          microsoftItemBody{ContentType: "text", Content: microsoftScheduleDescription(input)},
		Start:         microsoftDateTime{DateTime: input.StartAt.UTC().Format("2006-01-02T15:04:05"), TimeZone: "UTC"},
		End:           microsoftDateTime{DateTime: input.EndAt.UTC().Format("2006-01-02T15:04:05"), TimeZone: "UTC"},
		ShowAs:        "busy",
		Sensitivity:   "private",
		Categories:    []string{microsoftManagedCategory},
		TransactionID: input.BlockID.String(),
	}
	eventID := strings.TrimSpace(input.EventID)
	var response microsoftEvent
	if eventID != "" && !strings.HasPrefix(eventID, "pending:") {
		requestURL := strings.TrimRight(p.graphURL, "/") + "/me/events/" + url.PathEscape(eventID)
		err := p.doJSON(ctx, token, http.MethodPatch, requestURL, payload, &response)
		if err == nil {
			if strings.TrimSpace(response.ID) == "" {
				response.ID = eventID
			}
			return ExternalScheduleEventResult{EventID: response.ID}, nil
		}
		var graphErr *MicrosoftGraphError
		if !errors.As(err, &graphErr) || graphErr.StatusCode != http.StatusNotFound {
			return ExternalScheduleEventResult{}, err
		}
	}
	if err := p.doJSON(ctx, token, http.MethodPost, strings.TrimRight(p.graphURL, "/")+"/me/events", payload, &response); err != nil {
		return ExternalScheduleEventResult{}, err
	}
	if strings.TrimSpace(response.ID) == "" {
		return ExternalScheduleEventResult{}, errors.New("microsoft calendar event response is missing an ID")
	}
	return ExternalScheduleEventResult{EventID: response.ID}, nil
}

func (p *MicrosoftProvider) DeleteScheduleEvent(ctx context.Context, token ProviderToken, _, eventID string) error {
	if !hasProviderScope(token.Scopes, MicrosoftCalendarReadWriteScope) {
		return ErrCalendarReauthorizationRequired
	}
	eventID = strings.TrimSpace(eventID)
	if eventID == "" || strings.HasPrefix(eventID, "pending:") {
		return nil
	}
	err := p.doJSON(ctx, token, http.MethodDelete, strings.TrimRight(p.graphURL, "/")+"/me/events/"+url.PathEscape(eventID), nil, nil)
	var graphErr *MicrosoftGraphError
	if errors.As(err, &graphErr) && (graphErr.StatusCode == http.StatusNotFound || graphErr.StatusCode == http.StatusGone) {
		return nil
	}
	return err
}

func (p *MicrosoftProvider) doJSON(ctx context.Context, token ProviderToken, method, requestURL string, body, destination any) error {
	if p.microsoft == nil {
		return ErrCalendarNotConfigured
	}
	client, err := p.microsoft.CalendarHTTPClient(ctx, &oauth2.Token{
		AccessToken: token.AccessToken, RefreshToken: token.RefreshToken, TokenType: token.TokenType, Expiry: token.Expiry,
	})
	if err != nil {
		return err
	}
	var requestBody io.Reader
	if body != nil {
		encoded, encodeErr := json.Marshal(body)
		if encodeErr != nil {
			return encodeErr
		}
		requestBody = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, requestURL, requestBody)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Prefer", `IdType="ImmutableId", outlook.timezone="UTC", outlook.body-content-type="text"`)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		payload, _ := io.ReadAll(io.LimitReader(response.Body, 64<<10))
		return decodeMicrosoftGraphError(response.StatusCode, payload)
	}
	if destination == nil || response.StatusCode == http.StatusNoContent {
		return nil
	}
	if err := json.NewDecoder(response.Body).Decode(destination); err != nil {
		return fmt.Errorf("decode Microsoft Graph response: %w", err)
	}
	return nil
}

type MicrosoftGraphError struct {
	StatusCode int
	Code       string
	Message    string
}

func (e *MicrosoftGraphError) Error() string {
	return fmt.Sprintf("Microsoft Graph returned %d (%s): %s", e.StatusCode, e.Code, e.Message)
}

func decodeMicrosoftGraphError(status int, payload []byte) error {
	var response struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.Unmarshal(payload, &response)
	return &MicrosoftGraphError{StatusCode: status, Code: response.Error.Code, Message: response.Error.Message}
}

type microsoftEventPage struct {
	Value     []microsoftEvent `json:"value"`
	NextLink  string           `json:"@odata.nextLink"`
	DeltaLink string           `json:"@odata.deltaLink"`
}

type microsoftEvent struct {
	ID               string            `json:"id"`
	Subject          string            `json:"subject"`
	Body             microsoftItemBody `json:"body"`
	BodyPreview      string            `json:"bodyPreview"`
	Start            microsoftDateTime `json:"start"`
	End              microsoftDateTime `json:"end"`
	IsAllDay         bool              `json:"isAllDay"`
	IsCancelled      bool              `json:"isCancelled"`
	ShowAs           string            `json:"showAs"`
	Sensitivity      string            `json:"sensitivity"`
	WebLink          string            `json:"webLink"`
	OnlineMeetingURL string            `json:"onlineMeetingUrl"`
	OnlineMeeting    *struct {
		JoinURL string `json:"joinUrl"`
	} `json:"onlineMeeting"`
	Location struct {
		DisplayName string `json:"displayName"`
	} `json:"location"`
	Organizer  microsoftRecipient  `json:"organizer"`
	Attendees  []microsoftAttendee `json:"attendees"`
	Categories []string            `json:"categories"`
	Type       string              `json:"type"`
	Removed    map[string]any      `json:"@removed"`
}

type microsoftItemBody struct {
	ContentType string `json:"contentType"`
	Content     string `json:"content"`
}

type microsoftDateTime struct {
	DateTime string `json:"dateTime"`
	TimeZone string `json:"timeZone"`
}

type microsoftRecipient struct {
	EmailAddress struct {
		Name    string `json:"name"`
		Address string `json:"address"`
	} `json:"emailAddress"`
}

type microsoftAttendee struct {
	microsoftRecipient
	Type   string `json:"type"`
	Status struct {
		Response string `json:"response"`
	} `json:"status"`
}

type microsoftScheduleEvent struct {
	Subject       string            `json:"subject"`
	Body          microsoftItemBody `json:"body"`
	Start         microsoftDateTime `json:"start"`
	End           microsoftDateTime `json:"end"`
	ShowAs        string            `json:"showAs"`
	Sensitivity   string            `json:"sensitivity"`
	Categories    []string          `json:"categories"`
	TransactionID string            `json:"transactionId,omitempty"`
}

type microsoftSubscriptionRequest struct {
	ChangeType         string    `json:"changeType"`
	NotificationURL    string    `json:"notificationUrl"`
	Resource           string    `json:"resource"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
	ClientState        string    `json:"clientState"`
}

type microsoftSubscription struct {
	ID                 string    `json:"id"`
	Resource           string    `json:"resource"`
	ExpirationDateTime time.Time `json:"expirationDateTime"`
}

func microsoftSubscriptionChannel(subscription microsoftSubscription) (CalendarWatchChannel, error) {
	channel := CalendarWatchChannel{ChannelID: strings.TrimSpace(subscription.ID), ResourceID: strings.TrimSpace(subscription.Resource), ExpiresAt: subscription.ExpirationDateTime.UTC()}
	if channel.ChannelID == "" || channel.ResourceID == "" || channel.ExpiresAt.IsZero() {
		return CalendarWatchChannel{}, errors.New("microsoft Graph returned an incomplete subscription")
	}
	return channel, nil
}

func microsoftEventToCalendarEvent(item microsoftEvent) (CoreCalendarEvent, bool, error) {
	startAt, err := microsoftEventTime(item.Start)
	if err != nil {
		return CoreCalendarEvent{}, false, err
	}
	endAt, err := microsoftEventTime(item.End)
	if err != nil {
		return CoreCalendarEvent{}, false, err
	}
	if !endAt.After(startAt) {
		return CoreCalendarEvent{}, false, nil
	}
	visibility := normalizeMicrosoftSensitivity(item.Sensitivity)
	isPrivate := visibility == "private" || visibility == "confidential"
	event := CoreCalendarEvent{
		Provider: ProviderMicrosoft, CalendarID: "primary", ProviderEventID: microsoftProviderEventID(item.ID),
		Attendees: []CoreCalendarParticipant{}, IsAllDay: item.IsAllDay, StartAt: startAt, EndAt: endAt,
		Visibility: visibility, IsPrivate: isPrivate, BlocksTime: strings.ToLower(strings.TrimSpace(item.ShowAs)) != "free",
	}
	if item.IsAllDay {
		startDate := startAt.Format("2006-01-02")
		endDate := endAt.Format("2006-01-02")
		event.StartDate = &startDate
		event.EndDate = &endDate
	}
	if !isPrivate {
		event.Title = optionalString(item.Subject)
		description := item.Body.Content
		if strings.TrimSpace(description) == "" {
			description = item.BodyPreview
		}
		event.Description = optionalString(description)
		event.Location = optionalString(item.Location.DisplayName)
		event.HTMLLink = safeHTTPSURL(item.WebLink)
		meetingURL := item.OnlineMeetingURL
		if item.OnlineMeeting != nil && strings.TrimSpace(item.OnlineMeeting.JoinURL) != "" {
			meetingURL = item.OnlineMeeting.JoinURL
		}
		event.MeetingURL = safeHTTPSURL(meetingURL)
		event.Organizer = microsoftParticipant(item.Organizer, true)
		for _, attendee := range item.Attendees {
			participant := microsoftParticipant(attendee.microsoftRecipient, false)
			if participant == nil {
				continue
			}
			participant.Optional = strings.EqualFold(attendee.Type, "optional")
			participant.ResponseStatus = strings.TrimSpace(attendee.Status.Response)
			event.Attendees = append(event.Attendees, *participant)
		}
	}
	event.SourceHash = microsoftEventSourceHash(event)
	return event, true, nil
}

func microsoftParticipant(value microsoftRecipient, organizer bool) *CoreCalendarParticipant {
	name := strings.TrimSpace(value.EmailAddress.Name)
	email := strings.TrimSpace(value.EmailAddress.Address)
	if name == "" && email == "" {
		return nil
	}
	return &CoreCalendarParticipant{DisplayName: name, Email: email, Organizer: organizer}
}

func microsoftEventTime(value microsoftDateTime) (time.Time, error) {
	raw := strings.TrimSpace(value.DateTime)
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02T15:04:05.9999999", "2006-01-02T15:04:05"} {
		if parsed, err := time.Parse(layout, raw); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("microsoft calendar event has invalid time %q", raw)
}

func normalizeMicrosoftSensitivity(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "private", "personal":
		return "private"
	case "confidential":
		return "confidential"
	default:
		return "default"
	}
}

func microsoftProviderEventID(eventID string) string {
	return "primary:" + strings.TrimSpace(eventID)
}

func microsoftEventSourceHash(event CoreCalendarEvent) string {
	payload, _ := json.Marshal(event)
	digest := sha256.Sum256(payload)
	return hex.EncodeToString(digest[:])
}

func microsoftScheduleDescription(input ExternalScheduleEventInput) string {
	return fmt.Sprintf("%s\nfortyone_source:maya_schedule\nfortyone_block_id:%s\nfortyone_story_id:%s\nfortyone_workspace_id:%s", microsoftManagedDescription, input.BlockID, input.StoryID, input.WorkspaceID)
}

func microsoftManagedMetadata(description string) map[string]string {
	if !strings.Contains(description, microsoftManagedDescription) || !strings.Contains(description, "fortyone_source:maya_schedule") {
		return nil
	}
	metadata := map[string]string{"fortyone_source": "maya_schedule"}
	for _, line := range strings.Split(description, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if ok && strings.HasPrefix(key, "fortyone_") {
			metadata[strings.TrimSpace(key)] = strings.TrimSpace(value)
		}
	}
	return metadata
}

func microsoftManagedScheduleEventChange(item microsoftEvent, metadata map[string]string) (ManagedScheduleEventChange, error) {
	startAt, err := microsoftEventTime(item.Start)
	if err != nil {
		return ManagedScheduleEventChange{}, err
	}
	endAt, err := microsoftEventTime(item.End)
	if err != nil {
		return ManagedScheduleEventChange{}, err
	}
	return ManagedScheduleEventChange{
		EventID: item.ID, Title: strings.TrimSpace(item.Subject), StartAt: startAt, EndAt: endAt,
		Visibility: normalizeMicrosoftSensitivity(item.Sensitivity), Transparency: microsoftTransparency(item.ShowAs),
		Status: "confirmed", Source: metadata["fortyone_source"], BlockID: metadata["fortyone_block_id"],
		StoryID: metadata["fortyone_story_id"], WorkspaceID: metadata["fortyone_workspace_id"],
		HasAttendees: len(item.Attendees) > 0, Recurring: item.Type == "seriesMaster" || item.Type == "occurrence" || item.Type == "exception",
	}, nil
}

func microsoftTransparency(showAs string) string {
	if strings.EqualFold(strings.TrimSpace(showAs), "free") {
		return "transparent"
	}
	return "opaque"
}
