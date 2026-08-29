package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func (s *Service) ListConnections(ctx context.Context, workspaceID uuid.UUID, userID *uuid.UUID) ([]CoreConnection, error) {
	if s.repo == nil {
		return nil, ErrCalendarNotConfigured
	}
	return s.repo.ListConnections(ctx, workspaceID, userID)
}

func (s *Service) SetPrimaryConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) (CoreConnection, error) {
	if s.repo == nil {
		return CoreConnection{}, ErrCalendarNotConfigured
	}
	return s.repo.SetPrimaryConnection(ctx, workspaceID, userID, connectionID)
}

func (s *Service) CreateConnectSession(ctx context.Context, workspaceID, userID uuid.UUID, workspaceSlug string, providerName ...Provider) (CoreConnectSession, error) {
	providerID := ProviderGoogle
	if len(providerName) > 0 {
		providerID = providerName[0]
	}
	provider, err := s.provider(providerID)
	if err != nil {
		return CoreConnectSession{}, err
	}
	state, err := s.signState(stateClaims{
		WorkspaceID:   workspaceID,
		UserID:        userID,
		WorkspaceSlug: strings.TrimSpace(workspaceSlug),
		Provider:      providerID,
		ExpiresAt:     s.now().Add(connectStateTTL).Unix(),
	})
	if err != nil {
		return CoreConnectSession{}, err
	}
	authURL, err := provider.AuthCodeURL(state)
	if err != nil {
		return CoreConnectSession{}, err
	}
	return CoreConnectSession{AuthURL: authURL}, nil
}

func (s *Service) CompleteConnect(
	ctx context.Context,
	callbackUserID uuid.UUID,
	code, state string,
) (CoreConnection, string, error) {
	if s.repo == nil {
		return CoreConnection{}, "", ErrCalendarNotConfigured
	}
	claims, err := s.verifyActiveState(state)
	if err != nil {
		return CoreConnection{}, "", err
	}
	if callbackUserID == uuid.Nil || claims.UserID != callbackUserID {
		return CoreConnection{}, "", ErrCalendarAccessDenied
	}
	isMember, err := s.repo.WorkspaceMemberExists(ctx, claims.WorkspaceID, claims.UserID)
	if err != nil {
		return CoreConnection{}, "", err
	}
	if !isMember {
		return CoreConnection{}, "", ErrCalendarAccessDenied
	}
	provider, err := s.provider(claims.Provider)
	if err != nil {
		return CoreConnection{}, "", err
	}
	token, err := provider.ExchangeCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return CoreConnection{}, "", err
	}
	token, err = s.withRetainedRefreshToken(ctx, claims, token)
	if err != nil {
		return CoreConnection{}, "", err
	}
	payload, err := s.encryptTokenPayload(token)
	if err != nil {
		return CoreConnection{}, "", err
	}
	connection, err := s.repo.UpsertConnection(ctx, CoreConnectionUpsert{
		WorkspaceID:       claims.WorkspaceID,
		UserID:            claims.UserID,
		Provider:          claims.Provider,
		ProviderAccountID: strings.TrimSpace(token.ProviderAccountID),
		ConnectedEmail:    strings.TrimSpace(token.ConnectedEmail),
		Timezone:          fallbackTimezone(token.Timezone),
		TokenPayload:      payload,
		Scopes:            token.Scopes,
	})
	if err != nil {
		return CoreConnection{}, "", err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return CoreConnection{}, "", fmt.Errorf("sync calendar connection after connect: %w", err)
	}
	current, err := s.repo.GetConnection(ctx, connection.ID)
	if err != nil {
		return CoreConnection{}, "", fmt.Errorf("load calendar connection after initial sync: %w", err)
	}
	if err := s.replaceNotificationChannel(ctx, current); err != nil {
		return CoreConnection{}, "", fmt.Errorf("start calendar notification channel: %w", err)
	}
	if scheduleTasks, ok := s.cfg.Tasks.(CalendarScheduleTaskQueue); ok {
		if err := scheduleTasks.EnqueueCalendarScheduleReconcile(ctx, current.UserID); err != nil {
			return CoreConnection{}, "", fmt.Errorf("enqueue calendar schedule reconciliation after connect: %w", err)
		}
	}
	return connection, s.workspaceCalendarURL(claims.WorkspaceSlug, "connected=1&calendar_provider="+url.QueryEscape(string(claims.Provider))), nil
}

func (s *Service) withRetainedRefreshToken(
	ctx context.Context,
	claims stateClaims,
	token ProviderToken,
) (ProviderToken, error) {
	if strings.TrimSpace(token.AccessToken) == "" ||
		strings.TrimSpace(token.ProviderAccountID) == "" ||
		strings.TrimSpace(token.ConnectedEmail) == "" {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	if strings.TrimSpace(token.RefreshToken) != "" {
		return token, nil
	}

	existing, err := s.repo.GetActiveConnection(ctx, claims.WorkspaceID, claims.UserID, claims.Provider)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return ProviderToken{}, ErrCalendarCredentialsIncomplete
		}
		return ProviderToken{}, err
	}
	if strings.TrimSpace(existing.ProviderAccountID) == "" ||
		existing.ProviderAccountID != token.ProviderAccountID {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	previousToken, err := s.decryptTokenPayload(existing.TokenPayload)
	if err != nil || strings.TrimSpace(previousToken.RefreshToken) == "" {
		return ProviderToken{}, ErrCalendarCredentialsIncomplete
	}
	token.RefreshToken = previousToken.RefreshToken
	return token, nil
}

// CalendarCallbackErrorURL turns a valid signed OAuth state into a safe
// account-settings redirect without reflecting provider error text.
func (s *Service) CalendarCallbackErrorURL(state string, callbackUserID uuid.UUID, code string) (string, error) {
	claims, err := s.verifyActiveState(state)
	if err != nil {
		return "", err
	}
	if callbackUserID == uuid.Nil || claims.UserID != callbackUserID {
		return "", ErrCalendarAccessDenied
	}
	switch code {
	case "access_denied", "connection_failed":
	default:
		code = "connection_failed"
	}
	return s.workspaceCalendarURL(
		claims.WorkspaceSlug,
		"calendar_error="+url.QueryEscape(code)+"&calendar_provider="+url.QueryEscape(string(claims.Provider)),
	), nil
}

func (s *Service) RevokeConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetOwnedConnection(ctx, workspaceID, userID, connectionID)
	if err != nil {
		return err
	}
	s.stopNotificationChannel(ctx, connection)
	if err := s.repo.RevokeConnection(ctx, workspaceID, userID, connectionID); err != nil {
		return err
	}
	// Provider deletion is backed by the durable outbox. Try immediately for a
	// responsive disconnect, while the periodic dispatcher remains the retry
	// contract if the calendar provider is temporarily unavailable.
	_ = s.DispatchScheduleEventOutbox(ctx, userID)
	return nil
}

func (s *Service) ProcessGoogleNotification(ctx context.Context, channelID, resourceID, state, token string) error {
	connectionID, tokenChannelID, err := s.verifyNotificationToken(token)
	if err != nil || strings.TrimSpace(channelID) != tokenChannelID {
		return ErrInvalidCalendarNotification
	}
	if state == "sync" {
		return nil
	}
	if state != "exists" && state != "not_exists" {
		return ErrInvalidCalendarNotification
	}
	connection, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return nil
		}
		return err
	}
	if connection.NotificationChannelID != channelID || connection.NotificationResourceID != strings.TrimSpace(resourceID) {
		return nil
	}
	if s.cfg.Tasks == nil {
		return ErrCalendarNotConfigured
	}
	return s.cfg.Tasks.EnqueueCalendarSync(ctx, connectionID)
}

func (s *Service) ProcessMicrosoftNotification(ctx context.Context, subscriptionID, clientState string) error {
	connectionID, _, err := s.verifyNotificationToken(clientState)
	if err != nil {
		return ErrInvalidCalendarNotification
	}
	connection, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return nil
		}
		return err
	}
	if connection.Provider != ProviderMicrosoft || connection.NotificationChannelID != strings.TrimSpace(subscriptionID) {
		return nil
	}
	if s.cfg.Tasks == nil {
		return ErrCalendarNotConfigured
	}
	return s.cfg.Tasks.EnqueueCalendarSync(ctx, connectionID)
}

func (s *Service) SyncConnectionFromNotification(ctx context.Context, connectionID uuid.UUID) error {
	connection, err := s.repo.GetConnection(ctx, connectionID)
	if err != nil {
		if errors.Is(err, ErrCalendarNotFound) {
			return nil
		}
		return err
	}
	if strings.TrimSpace(connection.SyncToken) == "" || !connection.CanReadEventDetails() {
		if err := s.syncConnection(ctx, connection); err != nil {
			return err
		}
		return s.enqueueScheduleReconciliation(ctx, connection.UserID)
	}
	connection, err = s.repo.BeginConnectionSync(ctx, connection)
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
	delta, err := provider.SyncCalendarChanges(ctx, token, connection.SyncToken)
	if errors.Is(err, ErrCalendarFullSyncRequired) {
		current, getErr := s.repo.GetConnection(ctx, connection.ID)
		if getErr != nil {
			return getErr
		}
		if err := s.syncConnection(ctx, current); err != nil {
			return err
		}
		return s.enqueueScheduleReconciliation(ctx, current.UserID)
	}
	if err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	s.normalizeDelta(connection, &delta)
	if err := s.repo.ApplyCalendarChanges(ctx, connection, delta); err != nil {
		return s.failConnectionSync(ctx, connection, err)
	}
	if err := s.completeConnectionSync(ctx, connection); err != nil {
		return err
	}
	// Always enqueue after a successful incremental sync, including an empty
	// delta. If the previous attempt committed the provider's next sync token but the
	// enqueue failed, Asynq retries with an empty delta; this trailing enqueue is
	// what makes that handoff retry-safe.
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}

func (s *Service) enqueueScheduleReconciliation(ctx context.Context, userID uuid.UUID) error {
	scheduleTasks, ok := s.cfg.Tasks.(CalendarScheduleTaskQueue)
	if !ok {
		return nil
	}
	return scheduleTasks.EnqueueCalendarScheduleReconcile(ctx, userID)
}

func (s *Service) RenewExpiringNotificationChannels(ctx context.Context) (int, error) {
	if strings.TrimSpace(s.cfg.WebhookURL) == "" && len(s.cfg.WebhookURLs) == 0 {
		return 0, nil
	}
	connections, err := s.repo.ListConnectionsNeedingWatch(ctx, s.now().Add(googleWatchRenewalWindow))
	if err != nil {
		return 0, err
	}
	renewed := 0
	var renewalErr error
	for _, connection := range connections {
		if err := s.replaceNotificationChannel(ctx, connection); err != nil {
			renewalErr = errors.Join(renewalErr, err)
			continue
		}
		renewed++
	}
	return renewed, renewalErr
}

func (s *Service) SyncConnection(ctx context.Context, workspaceID, userID, connectionID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetOwnedConnection(ctx, workspaceID, userID, connectionID)
	if err != nil {
		return err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return err
	}
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}

func (s *Service) SyncActiveGoogleConnection(ctx context.Context, workspaceID, userID uuid.UUID) error {
	if s.repo == nil {
		return ErrCalendarNotConfigured
	}
	connection, err := s.repo.GetActiveConnection(ctx, workspaceID, userID, ProviderGoogle)
	if err != nil {
		return err
	}
	if err := s.syncConnection(ctx, connection); err != nil {
		return err
	}
	return s.enqueueScheduleReconciliation(ctx, connection.UserID)
}
