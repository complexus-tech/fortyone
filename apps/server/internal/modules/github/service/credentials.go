package github

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	githubsdk "github.com/google/go-github/v72/github"
	"github.com/google/uuid"
)

const (
	userLinkStateTTL                  = 15 * time.Minute
	githubCredentialBackfillBatchSize = 100
)

var ErrGitHubOAuthRevocationUnavailable = errors.New("github oauth credential revocation is unavailable")

type githubCredentialMaintenanceStore interface {
	ListGitHubUserCredentialsForRewrap(ctx context.Context, after *uuid.UUID, limit int) ([]githubshared.CredentialRecord, error)
	RewrapGitHubUserCredential(ctx context.Context, record githubshared.CredentialRecord, rewrapped string) (bool, error)
}

func (s *Service) CreateUserLinkSession(ctx context.Context, userID uuid.UUID, returnTo string) (CoreCreateUserLinkSession, error) {
	state, err := s.createUserLinkState(ctx, userID, returnTo)
	if err != nil {
		return CoreCreateUserLinkSession{}, err
	}
	return CoreCreateUserLinkSession{State: state}, nil
}

func (s *Service) LinkGitHubUser(ctx context.Context, userID uuid.UUID, code, state string) error {
	if _, err := s.consumeUserLinkState(ctx, state, userID); err != nil {
		return err
	}
	token, err := s.exchangeOAuthCode(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to exchange github oauth code: %w", err)
	}
	cleanupToken := func(cause error) error {
		return errors.Join(cause, s.revokeUserOAuthToken(ctx, token))
	}
	ghClient := githubsdk.NewClient(s.httpClient).WithAuthToken(token)
	user, _, err := ghClient.Users.Get(ctx, "")
	if err != nil {
		return cleanupToken(fmt.Errorf("failed to get github user: %w", err))
	}
	if s.credentials == nil {
		return cleanupToken(credentialvault.ErrNotConfigured)
	}
	generation := uuid.New()
	encrypted, err := s.sealUserCredential(userID, generation, token)
	if err != nil {
		return cleanupToken(fmt.Errorf("encrypt github user credential: %w", err))
	}
	if err := s.repo.LinkGitHubUser(
		ctx,
		userID,
		user.GetID(),
		user.GetLogin(),
		encrypted,
		credentialvault.CurrentVersion,
		generation,
	); err != nil {
		return cleanupToken(fmt.Errorf("persist github user credential: %w", err))
	}
	return nil
}

func (s *Service) UnlinkGitHubUser(ctx context.Context, userID uuid.UUID) error {
	token, err := s.userGitHubToken(ctx, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return s.repo.UnlinkGitHubUser(ctx, userID)
	}
	if err != nil {
		return fmt.Errorf("load github user credential for revocation: %w", err)
	}
	if err := s.revokeUserOAuthToken(ctx, token); err != nil {
		return err
	}
	return s.repo.UnlinkGitHubUser(ctx, userID)
}

func (s *Service) revokeUserOAuthToken(ctx context.Context, token string) error {
	clientID := strings.TrimSpace(s.cfg.ClientID)
	clientSecret := strings.TrimSpace(s.cfg.ClientSecret)
	token = strings.TrimSpace(token)
	if clientID == "" || clientSecret == "" || token == "" {
		return ErrGitHubOAuthRevocationUnavailable
	}

	transport := http.DefaultTransport
	timeout := 20 * time.Second
	var checkRedirect func(*http.Request, []*http.Request) error
	if s.httpClient != nil {
		if s.httpClient.Transport != nil {
			transport = s.httpClient.Transport
		}
		if s.httpClient.Timeout > 0 {
			timeout = s.httpClient.Timeout
		}
		checkRedirect = s.httpClient.CheckRedirect
	}
	oauthClient := githubsdk.NewClient(&http.Client{
		Transport: &githubsdk.BasicAuthTransport{
			Username:  clientID,
			Password:  clientSecret,
			Transport: transport,
		},
		Timeout:       timeout,
		CheckRedirect: checkRedirect,
	})
	response, err := oauthClient.Authorizations.Revoke(ctx, clientID, token)
	if err == nil || (response != nil && response.StatusCode == http.StatusNotFound) {
		return nil
	}
	return ErrGitHubOAuthRevocationUnavailable
}

// BackfillLegacyUserCredentials is the only GitHub path allowed to read a
// legacy plaintext token. Each conversion is an atomic compare-and-swap that
// replaces the original value with a context-bound vault envelope.
func (s *Service) BackfillLegacyUserCredentials(ctx context.Context) (int, error) {
	if s.credentials == nil {
		return 0, credentialvault.ErrNotConfigured
	}
	updated := 0
	for {
		records, err := s.repo.ListLegacyGitHubUserCredentials(ctx, githubCredentialBackfillBatchSize)
		if err != nil {
			return updated, fmt.Errorf("list legacy github user credentials: %w", err)
		}
		for _, record := range records {
			generation := uuid.New()
			encrypted, err := s.sealUserCredential(record.UserID, generation, record.Plaintext)
			if err != nil {
				return updated, fmt.Errorf("encrypt legacy github user credential for user %s: %w", record.UserID, err)
			}
			err = s.repo.UpgradeLegacyGitHubUserCredential(
				ctx,
				record.UserID,
				record.Plaintext,
				encrypted,
				credentialvault.CurrentVersion,
				generation,
			)
			if errors.Is(err, sql.ErrNoRows) {
				continue
			}
			if err != nil {
				return updated, fmt.Errorf("upgrade legacy github user credential for user %s: %w", record.UserID, err)
			}
			updated++
		}
		if len(records) < githubCredentialBackfillBatchSize {
			return updated, nil
		}
	}
}

// RewrapUserCredentials authenticates and rewraps every retained GitHub user
// credential under the active vault key. Pages are stable by user ID and every
// write compares the original envelope plus provider generation, so a
// concurrent relink or unlink always wins. Re-running after success is a no-op.
func (s *Service) RewrapUserCredentials(ctx context.Context) (credentialvault.RotationReport, error) {
	if s.credentials == nil {
		return credentialvault.RotationReport{}, credentialvault.ErrNotConfigured
	}
	if s.repo == nil {
		return credentialvault.RotationReport{}, errors.New("github credential store does not support key rewrap")
	}
	var store githubCredentialMaintenanceStore = s.repo
	activeKey, err := s.credentials.ActiveKeyRef()
	if err != nil {
		return credentialvault.RotationReport{}, err
	}
	report := credentialvault.RotationReport{ActiveKey: activeKey}
	var after *uuid.UUID
	for {
		records, err := store.ListGitHubUserCredentialsForRewrap(
			ctx,
			after,
			credentialvault.DefaultMaintenanceBatchSize,
		)
		if err != nil {
			return report, fmt.Errorf("list github user credentials for rewrap: %w", err)
		}
		for _, record := range records {
			report.Scanned++
			result, err := s.credentials.Rewrap(
				githubUserCredentialContext(record.UserID, record.Generation),
				record.Payload,
			)
			if err != nil {
				return report, fmt.Errorf("rewrap github user credential for user %s: %w", record.UserID, err)
			}
			if !result.Changed {
				report.Current++
				continue
			}
			replaced, err := store.RewrapGitHubUserCredential(ctx, record, result.Envelope)
			if err != nil {
				return report, fmt.Errorf("persist github user credential rewrap for user %s: %w", record.UserID, err)
			}
			if replaced {
				report.Rewrapped++
			} else {
				report.Stale++
			}
		}
		if len(records) < credentialvault.DefaultMaintenanceBatchSize {
			return report, nil
		}
		cursor := records[len(records)-1].UserID
		after = &cursor
	}
}

func (s *Service) userGitHubToken(ctx context.Context, userID uuid.UUID) (string, error) {
	credential, err := s.repo.GetUserGitHubCredential(ctx, userID)
	if err != nil {
		return "", err
	}
	if s.credentials == nil {
		return "", credentialvault.ErrNotConfigured
	}
	if credential.EnvelopeVersion != credentialvault.CurrentVersion || credential.Generation == uuid.Nil {
		return "", errors.New("github user credential requires vault migration")
	}
	opened, err := s.credentials.Open(
		githubUserCredentialContext(credential.UserID, credential.Generation),
		credential.Payload,
	)
	if err != nil {
		return "", fmt.Errorf("decrypt github user credential: %w", err)
	}
	defer opened.Destroy()
	tokenBytes := opened.Reveal()
	defer clear(tokenBytes)
	token := strings.TrimSpace(string(tokenBytes))
	if token == "" {
		return "", errors.New("github user credential is empty")
	}
	return token, nil
}

func githubUserCredentialContext(userID, generation uuid.UUID) credentialvault.Context {
	// #nosec G101 -- these literals are public AAD domain identifiers, not credentials.
	return credentialvault.Context{
		Provider:       "github",
		TenantID:       "account:" + userID.String(),
		SubjectID:      userID.String(),
		CredentialType: "user-oauth-access-token",
		Generation:     generation.String(),
	}
}

func (s *Service) sealUserCredential(userID, generation uuid.UUID, token string) (string, error) {
	plaintext := []byte(token)
	defer clear(plaintext)
	return s.credentials.Seal(githubUserCredentialContext(userID, generation), plaintext)
}
