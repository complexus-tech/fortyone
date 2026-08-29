package githubrepository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	githubsql "github.com/complexus-tech/projects-api/internal/modules/github/repository/sqlc"
	githubshared "github.com/complexus-tech/projects-api/internal/modules/github/shared"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
)

type GitHubUserCredentialRecord = githubshared.CredentialRecord
type LegacyGitHubUserCredentialRecord = githubshared.LegacyCredentialRecord

func (r *Repo) LinkGitHubUser(
	ctx context.Context,
	userID uuid.UUID,
	githubUserID int64,
	githubUsername, credentialPayload string,
	envelopeVersion int,
	generation uuid.UUID,
) error {
	if envelopeVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(credentialPayload) || generation == uuid.Nil {
		return errors.New("github encrypted credential metadata is required")
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.LinkGitHubUser(ctx, githubsql.LinkGitHubUserParams{
		GithubUserID:      &githubUserID,
		GithubUsername:    &githubUsername,
		CredentialPayload: &credentialPayload,
		EnvelopeVersion:   int16(envelopeVersion),
		Generation:        &generation,
		UserID:            userID,
	})
}

func (r *Repo) UnlinkGitHubUser(ctx context.Context, userID uuid.UUID) error {
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	return queries.UnlinkGitHubUser(ctx, githubsql.UnlinkGitHubUserParams{UserID: userID})
}

func (r *Repo) GetUserGitHubCredential(
	ctx context.Context,
	userID uuid.UUID,
) (GitHubUserCredentialRecord, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return GitHubUserCredentialRecord{}, err
	}
	row, err := queries.GetGitHubUserCredential(ctx, githubsql.GetGitHubUserCredentialParams{UserID: userID})
	if err != nil {
		return GitHubUserCredentialRecord{}, mapDatabaseError(err)
	}
	return mapCredentialRecord(row.UserID, row.CredentialPayload, row.EnvelopeVersion, row.Generation)
}

func (r *Repo) ListGitHubUserCredentialsForRewrap(
	ctx context.Context,
	after *uuid.UUID,
	limit int,
) ([]GitHubUserCredentialRecord, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	limit, pageLimit, err := normalizedCredentialPageLimit(limit)
	if err != nil {
		return nil, err
	}

	credentials := make([]GitHubUserCredentialRecord, 0, limit)
	if after == nil {
		rows, err := queries.ListGitHubUserCredentialsForRewrap(ctx, githubsql.ListGitHubUserCredentialsForRewrapParams{
			EnvelopeVersion: int16(credentialvault.CurrentVersion),
			PageLimit:       pageLimit,
		})
		if err != nil {
			return nil, err
		}
		for _, row := range rows {
			credential, err := mapCredentialRecord(row.UserID, row.CredentialPayload, row.EnvelopeVersion, row.Generation)
			if err != nil {
				return nil, err
			}
			credentials = append(credentials, credential)
		}
		return credentials, nil
	}

	rows, err := queries.ListGitHubUserCredentialsForRewrapAfter(ctx, githubsql.ListGitHubUserCredentialsForRewrapAfterParams{
		EnvelopeVersion: int16(credentialvault.CurrentVersion),
		AfterUserID:     *after,
		PageLimit:       pageLimit,
	})
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		credential, err := mapCredentialRecord(row.UserID, row.CredentialPayload, row.EnvelopeVersion, row.Generation)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	return credentials, nil
}

func (r *Repo) RewrapGitHubUserCredential(
	ctx context.Context,
	record GitHubUserCredentialRecord,
	rewrapped string,
) (bool, error) {
	if record.UserID == uuid.Nil || record.Generation == uuid.Nil ||
		record.EnvelopeVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(record.Payload) || !credentialvault.IsEnvelope(rewrapped) {
		return false, errors.New("github credential rewrap metadata is required")
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return false, err
	}
	affected, err := queries.RewrapGitHubUserCredential(ctx, githubsql.RewrapGitHubUserCredentialParams{
		RewrappedPayload: &rewrapped,
		UserID:           record.UserID,
		Generation:       &record.Generation,
		EnvelopeVersion:  int16(record.EnvelopeVersion),
		ExpectedPayload:  &record.Payload,
	})
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}

func (r *Repo) ListLegacyGitHubUserCredentials(
	ctx context.Context,
	limit int,
) ([]LegacyGitHubUserCredentialRecord, error) {
	queries, err := r.configuredQueries()
	if err != nil {
		return nil, err
	}
	_, pageLimit, err := normalizedCredentialPageLimit(limit)
	if err != nil {
		return nil, err
	}
	rows, err := queries.ListLegacyGitHubUserCredentials(ctx, githubsql.ListLegacyGitHubUserCredentialsParams{
		PageLimit: pageLimit,
	})
	if err != nil {
		return nil, err
	}
	credentials := make([]LegacyGitHubUserCredentialRecord, 0, len(rows))
	for _, row := range rows {
		if row.CredentialPayload == nil || strings.TrimSpace(*row.CredentialPayload) == "" {
			return nil, errors.New("GitHub legacy credential row is missing its payload")
		}
		credentials = append(credentials, LegacyGitHubUserCredentialRecord{
			UserID:    row.UserID,
			Plaintext: *row.CredentialPayload,
		})
	}
	return credentials, nil
}

func normalizedCredentialPageLimit(requested int) (int, int32, error) {
	normalized := credentialvault.MaintenanceBatchSize(requested)
	pageLimit, err := safecast.Int32(normalized)
	if err != nil {
		return 0, 0, fmt.Errorf("convert GitHub credential page limit: %w", err)
	}
	return normalized, pageLimit, nil
}

func (r *Repo) UpgradeLegacyGitHubUserCredential(
	ctx context.Context,
	userID uuid.UUID,
	expectedPlaintext, encrypted string,
	envelopeVersion int,
	generation uuid.UUID,
) error {
	if strings.TrimSpace(expectedPlaintext) == "" ||
		envelopeVersion != credentialvault.CurrentVersion ||
		!credentialvault.IsEnvelope(encrypted) || generation == uuid.Nil {
		return errors.New("github credential upgrade metadata is required")
	}
	queries, err := r.configuredQueries()
	if err != nil {
		return err
	}
	affected, err := queries.UpgradeLegacyGitHubUserCredential(ctx, githubsql.UpgradeLegacyGitHubUserCredentialParams{
		EncryptedPayload:  &encrypted,
		EnvelopeVersion:   int16(envelopeVersion),
		Generation:        &generation,
		UserID:            userID,
		ExpectedPlaintext: &expectedPlaintext,
	})
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func mapCredentialRecord(
	userID uuid.UUID,
	payload *string,
	envelopeVersion int32,
	generation *uuid.UUID,
) (GitHubUserCredentialRecord, error) {
	if payload == nil || generation == nil || *generation == uuid.Nil {
		return GitHubUserCredentialRecord{}, fmt.Errorf("GitHub credential metadata is incomplete for user %s", userID)
	}
	return GitHubUserCredentialRecord{
		UserID:          userID,
		Payload:         *payload,
		EnvelopeVersion: int(envelopeVersion),
		Generation:      *generation,
	}, nil
}
