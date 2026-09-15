package calendar

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2"
)

type cleanupRefreshProvider struct {
	*fakeProvider
	refreshErr error
}

func (provider *cleanupRefreshProvider) RefreshToken(_ context.Context, token ProviderToken) (ProviderToken, error) {
	if provider.refreshErr != nil {
		return ProviderToken{}, provider.refreshErr
	}
	token.AccessToken += "-refreshed"
	return token, nil
}

type cleanupCredentialRepository struct {
	*fakeRepo
	storageErr error
}

func (repository *cleanupCredentialRepository) UpdateConnectionToken(ctx context.Context, connection CoreConnection, payload string) error {
	if repository.storageErr != nil {
		return repository.storageErr
	}
	return repository.fakeRepo.UpdateConnectionToken(ctx, connection, payload)
}

func newAccountCleanupFixture(t *testing.T) (*Service, *cleanupCredentialRepository, *cleanupRefreshProvider) {
	t.Helper()
	userID, workspaceID := uuid.New(), uuid.New()
	repository := &cleanupCredentialRepository{fakeRepo: &fakeRepo{
		connection: CoreConnection{
			ID: uuid.New(), UserID: userID, WorkspaceID: workspaceID, Provider: ProviderMicrosoft,
			Scopes: []string{MicrosoftCalendarReadWriteScope},
		},
		cleanupPending: true,
		pendingOutboxBatches: [][]CoreScheduleEventOutbox{{{
			ID: uuid.New(), UserID: userID, WorkspaceID: workspaceID,
			Provider: ProviderMicrosoft, Operation: ScheduleEventOperationDelete,
			CalendarID: "primary", ProviderEventID: "owned-event",
		}}},
	}}
	provider := &cleanupRefreshProvider{fakeProvider: &fakeProvider{}}
	service := New(nil, repository, Config{SecretKey: "test-secret", Providers: map[Provider]CalendarProvider{ProviderMicrosoft: provider}})
	payload, err := service.encryptTokenPayload(ProviderToken{AccessToken: "access", RefreshToken: "refresh"})
	require.NoError(t, err)
	repository.connection.TokenPayload = payload
	return service, repository, provider
}

func TestAccountCleanupRetriesRefreshAndStorageFailuresWithoutLosingDeletion(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name    string
		failure error
		storage bool
	}{
		{name: "refresh timeout", failure: context.DeadlineExceeded},
		{name: "refresh outage", failure: &oauth2.RetrieveError{Response: &http.Response{StatusCode: 503}, ErrorCode: "temporarily_unavailable"}},
		{name: "application configuration", failure: &oauth2.RetrieveError{Response: &http.Response{StatusCode: 401}, ErrorCode: "invalid_client"}},
		{name: "credential persistence", failure: errors.New("database unavailable"), storage: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			service, repository, provider := newAccountCleanupFixture(t)
			if scenario.storage {
				repository.storageErr = scenario.failure
			} else {
				provider.refreshErr = scenario.failure
			}
			require.ErrorIs(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID), scenario.failure)
			require.Empty(t, repository.failedOutboxPermanent)
			require.Zero(t, repository.cleanupFinalizeCalls, "transient failures must retain the only cleanup credential")
			require.Zero(t, repository.outboxClaimCalls, "retry must retain the pending delete")

			repository.storageErr, provider.refreshErr = nil, nil
			require.NoError(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID))
			require.Equal(t, []string{"owned-event"}, provider.deletedEventIDs)
			require.Len(t, repository.processedOutbox, 1)
			require.Equal(t, 1, repository.cleanupFinalizeCalls)
		})
	}
}

func TestAccountCleanupInvalidGrantCanFinishWithoutRetainingRevokedCredentials(t *testing.T) {
	t.Parallel()
	service, repository, provider := newAccountCleanupFixture(t)
	provider.refreshErr = fmt.Errorf("refresh: %w", &oauth2.RetrieveError{Response: &http.Response{StatusCode: 400}, ErrorCode: "invalid_grant"})
	require.NoError(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID))
	require.Equal(t, []bool{true}, repository.failedOutboxPermanent)
	require.Equal(t, 1, repository.cleanupFinalizeCalls)
	require.Empty(t, provider.deletedEventIDs)
}

func TestAccountCleanupRecoversAfterRestoringEncryptionKey(t *testing.T) {
	t.Parallel()
	service, repository, provider := newAccountCleanupFixture(t)
	originalKey, originalPayload := service.cfg.SecretKey, repository.connection.TokenPayload
	service.cfg.SecretKey = "wrong-deployment-secret"
	require.Error(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID))
	require.Equal(t, originalPayload, repository.connection.TokenPayload)
	require.Empty(t, repository.failedOutboxPermanent)
	require.Zero(t, repository.cleanupFinalizeCalls)
	require.Zero(t, repository.outboxClaimCalls)
	require.Empty(t, provider.deletedEventIDs)

	service.cfg.SecretKey = originalKey
	require.NoError(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID))
	require.Equal(t, []string{"owned-event"}, provider.deletedEventIDs)
	require.Len(t, repository.processedOutbox, 1)
	require.Equal(t, 1, repository.cleanupFinalizeCalls)
}

func TestAccountCleanupWaitsForMissingProviderConfiguration(t *testing.T) {
	t.Parallel()
	service, repository, _ := newAccountCleanupFixture(t)
	service.cfg.Providers = nil
	require.ErrorIs(t, service.DispatchScheduleEventOutbox(t.Context(), repository.connection.UserID), ErrCalendarNotConfigured)
	require.Zero(t, repository.outboxClaimCalls)
	require.Zero(t, repository.cleanupFinalizeCalls)
	require.Empty(t, repository.failedOutboxPermanent)
}
