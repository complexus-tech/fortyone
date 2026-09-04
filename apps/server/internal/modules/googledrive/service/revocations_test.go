package googledrive

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRevocationDispatcherOpensStoredRefreshTokenWithoutRefreshing(t *testing.T) {
	t.Parallel()

	dispatcher, repo, provider := newRevocationDispatcherFixture(t)

	processed, err := dispatcher.DispatchPendingRevocations(t.Context())

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, "refresh-token", provider.revokedToken)
	require.Equal(t, []string{
		"list", "user_lock_begin", "subject_lock_begin", "claim", "revoke",
		"complete", "subject_lock_end", "user_lock_end",
	}, repo.events)
	require.Zero(t, repo.retryCalls)
}

func TestRevocationDispatcherRevokesAccessOnlyFailedOAuthEnvelope(t *testing.T) {
	t.Parallel()

	dispatcher, repo, provider := newRevocationDispatcherFixture(t)
	revocation := repo.revocation
	revocation.SourceAccountID = nil
	revocation.CredentialPayload = ""
	payload, err := (&Service{config: Config{Credentials: dispatcher.credentials}}).sealToken(
		revocation.Account(),
		domain.OAuthToken{AccessToken: "access-only-token"},
	)
	require.NoError(t, err)
	repo.revocation.SourceAccountID = nil
	repo.revocation.CredentialPayload = payload

	processed, err := dispatcher.DispatchPendingRevocations(t.Context())

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Equal(t, "access-only-token", provider.revokedToken)
	require.Equal(t, 1, repo.completeCalls)
	require.Zero(t, repo.retryCalls)
}

func TestRevocationDispatcherSkipsProviderWhenReconnectSupersedesClaim(t *testing.T) {
	t.Parallel()

	dispatcher, repo, provider := newRevocationDispatcherFixture(t)
	repo.claimed = false

	processed, err := dispatcher.DispatchPendingRevocations(t.Context())

	require.NoError(t, err)
	require.Equal(t, 1, processed)
	require.Zero(t, provider.revokeCalls)
	require.Zero(t, repo.completeCalls)
	require.Zero(t, repo.retryCalls)
}

func TestRevocationDispatcherTreatsAlreadyInvalidTokenAsCompleted(t *testing.T) {
	t.Parallel()

	dispatcher, repo, provider := newRevocationDispatcherFixture(t)
	provider.revokeErr = &APIError{StatusCode: http.StatusBadRequest, Code: "invalid_token"}

	_, err := dispatcher.DispatchPendingRevocations(t.Context())

	require.NoError(t, err)
	require.Equal(t, 1, repo.completeCalls)
	require.Zero(t, repo.retryCalls)
}

func TestRevocationDispatcherRetriesTransientProviderFailure(t *testing.T) {
	t.Parallel()

	dispatcher, repo, provider := newRevocationDispatcherFixture(t)
	provider.revokeErr = errors.New("provider unavailable")

	_, err := dispatcher.DispatchPendingRevocations(t.Context())

	require.ErrorContains(t, err, "remote revocation request failed")
	require.Zero(t, repo.completeCalls)
	require.Equal(t, 1, repo.retryCalls)
	require.False(t, repo.retryTerminal)
	require.Equal(t, "Google Drive remote revocation request failed", repo.retryError)
}

func TestRevocationDispatcherClassifiesRetryableAndTerminalProviderFailures(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		name       string
		revokeErr  error
		attempts   int
		wantFailed bool
	}{
		{
			name: "rate limit retries",
			revokeErr: &APIError{
				StatusCode: http.StatusTooManyRequests, Code: "resource_exhausted",
				Message: "provider response body must not be persisted",
			},
		},
		{
			name: "Google 403 rate-limit reason retries",
			revokeErr: &APIError{
				StatusCode: http.StatusForbidden, Reasons: []string{"userRateLimitExceeded"},
				Message: "provider response body must not be persisted",
			},
		},
		{name: "server failure retries", revokeErr: &APIError{StatusCode: http.StatusServiceUnavailable}},
		{name: "permanent client failure stops", revokeErr: &APIError{StatusCode: http.StatusForbidden}, wantFailed: true},
		{name: "retry budget stops", revokeErr: errors.New("provider unavailable"), attempts: revocationMaximumAttempts, wantFailed: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			dispatcher, repo, provider := newRevocationDispatcherFixture(t)
			provider.revokeErr = test.revokeErr
			if test.attempts > 0 {
				repo.revocation.AttemptCount = test.attempts
			}

			_, err := dispatcher.DispatchPendingRevocations(t.Context())

			require.Error(t, err)
			require.Equal(t, 1, repo.retryCalls)
			require.Equal(t, test.wantFailed, repo.retryTerminal)
			require.NotContains(t, repo.retryError, "provider response body")
			require.NotContains(t, err.Error(), "provider response body")
		})
	}
}

func newRevocationDispatcherFixture(
	t *testing.T,
) (*RevocationDispatcher, *revocationRepositoryStub, *revocationProviderStub) {
	t.Helper()

	vault, err := credentialvault.NewFromEncodedKeyring(
		credentialvault.DevelopmentKeyID,
		credentialvault.DevelopmentKeyVersion,
		credentialvault.DevelopmentEncodedKeys,
	)
	require.NoError(t, err)
	userID := uuid.New()
	generation := uuid.New()
	accountID := uuid.New()
	account := domain.Account{
		ID: accountID, UserID: userID, GoogleSubject: "google-subject",
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: generation,
	}
	token := domain.OAuthToken{
		AccessToken: "expired-access-token", RefreshToken: "refresh-token",
		Expiry: time.Now().Add(-time.Hour),
	}
	payload, err := (&Service{config: Config{Credentials: vault}}).sealToken(account, token)
	require.NoError(t, err)

	candidate := domain.RevocationCandidate{ID: uuid.New(), UserID: userID, GoogleSubject: account.GoogleSubject}
	repo := &revocationRepositoryStub{
		candidates: []domain.RevocationCandidate{candidate},
		claimed:    true,
		revocation: domain.Revocation{
			ID: candidate.ID, SourceAccountID: &accountID,
			UserID: userID, GoogleSubject: account.GoogleSubject,
			InstallationGeneration: generation,
			CredentialPayload:      payload, CredentialVersion: int16(credentialvault.CurrentVersion),
			AttemptCount: 1, ClaimToken: uuid.New(), LeaseExpiresAt: time.Now().Add(time.Minute),
		},
	}
	provider := &revocationProviderStub{events: &repo.events}
	dispatcher := NewRevocationDispatcher(nil, repo, vault)
	dispatcher.client = provider
	dispatcher.now = func() time.Time { return time.Date(2026, 9, 4, 0, 0, 0, 0, time.UTC) }
	return dispatcher, repo, provider
}

type revocationRepositoryStub struct {
	events        []string
	candidates    []domain.RevocationCandidate
	revocation    domain.Revocation
	claimed       bool
	completeCalls int
	retryCalls    int
	retryTerminal bool
	retryError    string
}

func (repo *revocationRepositoryStub) WithinProviderUserLifecycle(
	ctx context.Context,
	_ uuid.UUID,
	operation func(context.Context) error,
) error {
	repo.events = append(repo.events, "user_lock_begin")
	err := operation(ctx)
	repo.events = append(repo.events, "user_lock_end")
	return err
}

func (repo *revocationRepositoryStub) WithinProviderSubjectLifecycle(
	ctx context.Context,
	_ string,
	operation func(context.Context) error,
) error {
	repo.events = append(repo.events, "subject_lock_begin")
	err := operation(ctx)
	repo.events = append(repo.events, "subject_lock_end")
	return err
}

func (repo *revocationRepositoryStub) ListReadyRevocations(
	context.Context,
	time.Time,
	int,
) ([]domain.RevocationCandidate, error) {
	repo.events = append(repo.events, "list")
	return repo.candidates, nil
}

func (repo *revocationRepositoryStub) ClaimRevocation(
	context.Context,
	domain.RevocationCandidate,
	uuid.UUID,
	time.Time,
	time.Time,
) (domain.Revocation, bool, error) {
	repo.events = append(repo.events, "claim")
	return repo.revocation, repo.claimed, nil
}

func (repo *revocationRepositoryStub) CompleteRevocation(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	time.Time,
) error {
	repo.events = append(repo.events, "complete")
	repo.completeCalls++
	return nil
}

func (repo *revocationRepositoryStub) RetryRevocation(
	_ context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	lastError string,
	_ time.Time,
	_ time.Time,
	terminal bool,
) error {
	repo.events = append(repo.events, "retry")
	repo.retryCalls++
	repo.retryTerminal = terminal
	repo.retryError = lastError
	return nil
}

type revocationProviderStub struct {
	ProviderClient
	events       *[]string
	revokedToken string
	revokeErr    error
	revokeCalls  int
}

func (provider *revocationProviderStub) Revoke(_ context.Context, token string) error {
	provider.revokeCalls++
	provider.revokedToken = token
	if provider.events != nil {
		*provider.events = append(*provider.events, "revoke")
	}
	return provider.revokeErr
}
