package calendar

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

func (s *Service) validateScheduleStory(ctx context.Context, input CoreScheduleBlockInput) error {
	if input.StoryID == nil {
		return nil
	}
	exists, err := s.repo.ScheduleStoryExists(ctx, input.WorkspaceID, input.UserID, *input.StoryID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrInvalidScheduleBlock
	}
	return nil
}

func (s *Service) syncConnection(ctx context.Context, connection CoreConnection) error {
	connection, err := s.repo.BeginConnectionSync(ctx, connection)
	if err != nil {
		return err
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	token, err := s.tokenForConnection(ctx, connection, provider)
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	timeMin := s.now().Add(defaultSyncLookback)
	timeMax := s.now().Add(defaultSyncLookahead)
	snapshot, err := provider.SyncCalendar(ctx, token, BusyWindowInput{
		ConnectionID: connection.ID,
		WorkspaceID:  connection.WorkspaceID,
		UserID:       connection.UserID,
		TimeMin:      timeMin,
		TimeMax:      timeMax,
		Timezone:     fallbackTimezone(connection.Timezone),
	})
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	for i := range snapshot.Events {
		snapshot.Events[i].ConnectionID = connection.ID
		snapshot.Events[i].WorkspaceID = connection.WorkspaceID
		snapshot.Events[i].UserID = connection.UserID
		snapshot.Events[i].Provider = connection.Provider
		if strings.TrimSpace(snapshot.Events[i].CalendarID) == "" {
			snapshot.Events[i].CalendarID = "primary"
		}
		if strings.TrimSpace(snapshot.Events[i].Visibility) == "" {
			snapshot.Events[i].Visibility = "default"
		}
		if snapshot.Events[i].Attendees == nil {
			snapshot.Events[i].Attendees = []CoreCalendarParticipant{}
		}
	}
	for i := range snapshot.BusyWindows {
		snapshot.BusyWindows[i].ConnectionID = connection.ID
		snapshot.BusyWindows[i].WorkspaceID = connection.WorkspaceID
		snapshot.BusyWindows[i].UserID = connection.UserID
		snapshot.BusyWindows[i].Provider = connection.Provider
	}
	if err := s.repo.ReplaceCalendarSnapshot(ctx, connection, snapshot); err != nil {
		if errors.Is(err, ErrCalendarSyncSuperseded) {
			return err
		}
		return s.failConnectionSync(ctx, connection, err)
	}
	return s.completeConnectionSync(ctx, connection)
}

func (s *Service) completeConnectionSync(ctx context.Context, connection CoreConnection) error {
	syncedAt := s.now().UTC()
	if err := s.repo.MarkConnectionSynced(
		ctx,
		connection.WorkspaceID,
		connection.ID,
		connection.CredentialGeneration,
		syncedAt,
	); err != nil {
		return err
	}
	if s.cfg.Updates == nil {
		return nil
	}
	if err := s.cfg.Updates.PublishCalendarUpdated(
		ctx,
		connection.WorkspaceID,
		connection.UserID,
		connection.ID,
		syncedAt,
	); err != nil {
		return fmt.Errorf("publish calendar update: %w", err)
	}
	return nil
}

func (s *Service) replaceNotificationChannel(ctx context.Context, connection CoreConnection) error {
	address := s.webhookURL(connection.Provider)
	if address == "" {
		if !s.cfg.RequireWebhook {
			return nil
		}
		return ErrCalendarWebhookNotConfigured
	}
	parsedAddress, err := url.Parse(address)
	if err != nil || parsedAddress.Scheme != "https" || parsedAddress.Host == "" {
		return fmt.Errorf("calendar webhook must be a public HTTPS URL")
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return err
	}
	token, err := s.tokenForConnection(ctx, connection, provider)
	if err != nil {
		return err
	}
	channelID := uuid.NewString()
	watchInput := CalendarWatchInput{
		ChannelID: channelID,
		Address:   parsedAddress.String(),
		Token:     s.notificationToken(connection.ID, channelID),
		TTL:       s.watchTTL(connection.Provider),
	}
	var channel CalendarWatchChannel
	if connection.NotificationChannelID != "" {
		if renewer, ok := provider.(CalendarWatchRenewer); ok {
			channel, err = renewer.RenewCalendarWatch(ctx, token, CalendarWatchChannel{
				ChannelID: connection.NotificationChannelID, ResourceID: connection.NotificationResourceID,
				ExpiresAt: valueOrZeroTime(connection.NotificationExpiresAt),
			}, watchInput)
		} else {
			channel, err = provider.WatchCalendar(ctx, token, watchInput)
		}
	} else {
		channel, err = provider.WatchCalendar(ctx, token, watchInput)
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(channel.ChannelID) == "" ||
		strings.TrimSpace(channel.ResourceID) == "" ||
		channel.ExpiresAt.IsZero() {
		return errors.New("calendar provider returned an incomplete notification channel")
	}
	if err := s.repo.SetNotificationChannel(ctx, connection, channel); err != nil {
		_ = provider.StopCalendarWatch(ctx, token, channel)
		return err
	}
	_, renewedInPlace := provider.(CalendarWatchRenewer)
	if !renewedInPlace && connection.NotificationChannelID != "" && connection.NotificationResourceID != "" {
		old := CalendarWatchChannel{
			ChannelID:  connection.NotificationChannelID,
			ResourceID: connection.NotificationResourceID,
		}
		if err := provider.StopCalendarWatch(ctx, token, old); err != nil && s.log != nil {
			s.log.Error(ctx, "failed to stop replaced calendar notification channel", "err", err, "connection_id", connection.ID)
		}
	}
	return nil
}

func (s *Service) webhookURL(provider Provider) string {
	if value := strings.TrimSpace(s.cfg.WebhookURLs[provider]); value != "" {
		return value
	}
	if provider == ProviderGoogle {
		return strings.TrimSpace(s.cfg.WebhookURL)
	}
	return ""
}

func (s *Service) watchTTL(provider Provider) time.Duration {
	if provider == ProviderMicrosoft {
		return microsoftWatchTTL
	}
	return googleWatchTTL
}

func valueOrZeroTime(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func (s *Service) stopNotificationChannel(ctx context.Context, connection CoreConnection) {
	if connection.NotificationChannelID == "" || connection.NotificationResourceID == "" {
		return
	}
	provider, err := s.provider(connection.Provider)
	if err != nil {
		return
	}
	token, err := s.tokenForConnection(ctx, connection, provider)
	if err != nil {
		return
	}
	err = provider.StopCalendarWatch(ctx, token, CalendarWatchChannel{
		ChannelID:  connection.NotificationChannelID,
		ResourceID: connection.NotificationResourceID,
	})
	if err != nil && s.log != nil {
		s.log.Error(ctx, "failed to stop calendar notification channel", "err", err, "connection_id", connection.ID)
	}
}

func (s *Service) notificationToken(connectionID uuid.UUID, channelID string) string {
	payload := connectionID.String() + "." + channelID
	mac := hmac.New(sha256.New, []byte(s.cfg.SecretKey))
	_, _ = mac.Write([]byte(payload))
	return payload + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *Service) verifyNotificationToken(value string) (uuid.UUID, string, error) {
	parts := strings.Split(strings.TrimSpace(value), ".")
	if len(parts) != 3 {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	connectionID, err := uuid.Parse(parts[0])
	if err != nil {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	expected := s.notificationToken(connectionID, parts[1])
	if !hmac.Equal([]byte(expected), []byte(strings.TrimSpace(value))) {
		return uuid.Nil, "", ErrInvalidCalendarNotification
	}
	return connectionID, parts[1], nil
}

func (s *Service) normalizeDelta(connection CoreConnection, delta *CalendarSyncDelta) {
	for i := range delta.Events {
		delta.Events[i].ConnectionID = connection.ID
		delta.Events[i].WorkspaceID = connection.WorkspaceID
		delta.Events[i].UserID = connection.UserID
		delta.Events[i].Provider = connection.Provider
		if strings.TrimSpace(delta.Events[i].CalendarID) == "" {
			delta.Events[i].CalendarID = "primary"
		}
		if strings.TrimSpace(delta.Events[i].Visibility) == "" {
			delta.Events[i].Visibility = "default"
		}
		if delta.Events[i].Attendees == nil {
			delta.Events[i].Attendees = []CoreCalendarParticipant{}
		}
	}
	for i := range delta.BusyWindows {
		delta.BusyWindows[i].ConnectionID = connection.ID
		delta.BusyWindows[i].WorkspaceID = connection.WorkspaceID
		delta.BusyWindows[i].UserID = connection.UserID
		delta.BusyWindows[i].Provider = connection.Provider
	}
}

func (s *Service) failConnectionSync(ctx context.Context, connection CoreConnection, syncErr error) error {
	if s.log != nil {
		s.log.Error(ctx, "calendar sync failed", "err", syncErr, "connection_id", connection.ID)
	}
	markErr := s.repo.MarkConnectionSyncFailed(
		ctx,
		connection.WorkspaceID,
		connection.ID,
		connection.CredentialGeneration,
		"Calendar could not be refreshed.",
	)
	if markErr == nil {
		return syncErr
	}
	if errors.Is(markErr, ErrCalendarSyncSuperseded) {
		return markErr
	}
	return errors.Join(syncErr, markErr)
}
