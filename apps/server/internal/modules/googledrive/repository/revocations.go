package googledriverepository

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/safecast"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const maximumRevocationListSize = 100

func (repository *Repository) EnqueueRevocation(
	ctx context.Context,
	revocation domain.Revocation,
) (domain.RevocationCandidate, error) {
	if revocation.UserID == uuid.Nil || revocation.GoogleSubject == "" ||
		revocation.InstallationGeneration == uuid.Nil ||
		revocation.CredentialPayload == "" || revocation.CredentialVersion <= 0 {
		return domain.RevocationCandidate{}, domain.ErrInvalidInput
	}

	var candidate domain.RevocationCandidate
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		if err := queries.LockGoogleDriveUserLifecycle(ctx, googledrivesql.LockGoogleDriveUserLifecycleParams{
			UserID: revocation.UserID,
		}); err != nil {
			return err
		}
		if err := queries.LockGoogleDriveSubjectLifecycle(ctx, googledrivesql.LockGoogleDriveSubjectLifecycleParams{
			GoogleSubject: &revocation.GoogleSubject,
		}); err != nil {
			return err
		}
		_, err := queries.GetActiveGoogleDriveAccountBySubject(
			ctx,
			googledrivesql.GetActiveGoogleDriveAccountBySubjectParams{GoogleSubject: revocation.GoogleSubject},
		)
		if err == nil {
			return domain.ErrConflict
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		credentialPayload := revocation.CredentialPayload
		row, err := queries.EnqueueGoogleDriveRevocation(
			ctx,
			googledrivesql.EnqueueGoogleDriveRevocationParams{
				SourceAccountID: revocation.SourceAccountID,
				UserID:          revocation.UserID, GoogleSubject: revocation.GoogleSubject,
				InstallationGeneration: revocation.InstallationGeneration,
				CredentialPayload:      &credentialPayload,
				CredentialKeyVersion:   revocation.CredentialVersion,
			},
		)
		if err != nil {
			return err
		}
		candidate = domain.RevocationCandidate{
			ID: row.RevocationID, UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		}
		return nil
	})
	return candidate, err
}

func (repository *Repository) ListReadyRevocations(
	ctx context.Context,
	readyAt time.Time,
	limit int,
) ([]domain.RevocationCandidate, error) {
	if limit <= 0 || limit > maximumRevocationListSize {
		return nil, domain.ErrInvalidInput
	}
	rowLimit, err := safecast.Int32(limit)
	if err != nil {
		return nil, domain.ErrInvalidInput
	}
	rows, err := repository.queriesForContext(ctx).ListReadyGoogleDriveRevocations(
		ctx,
		googledrivesql.ListReadyGoogleDriveRevocationsParams{
			ReadyAt:  readyAt.UTC(),
			RowLimit: rowLimit,
		},
	)
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	candidates := make([]domain.RevocationCandidate, len(rows))
	for index, row := range rows {
		candidates[index] = domain.RevocationCandidate{
			ID: row.RevocationID, UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		}
	}
	return candidates, nil
}

func (repository *Repository) ClaimRevocation(
	ctx context.Context,
	candidate domain.RevocationCandidate,
	claimToken uuid.UUID,
	claimedAt, leaseExpiresAt time.Time,
) (domain.Revocation, bool, error) {
	if candidate.ID == uuid.Nil || candidate.UserID == uuid.Nil || candidate.GoogleSubject == "" ||
		claimToken == uuid.Nil || claimedAt.IsZero() || !leaseExpiresAt.After(claimedAt) {
		return domain.Revocation{}, false, domain.ErrInvalidInput
	}

	var claimed domain.Revocation
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		identity, err := queries.GetGoogleDriveRevocationIdentityForUpdate(
			ctx,
			googledrivesql.GetGoogleDriveRevocationIdentityForUpdateParams{RevocationID: candidate.ID},
		)
		if err != nil {
			return err
		}
		if identity.UserID != candidate.UserID || identity.GoogleSubject != candidate.GoogleSubject {
			return domain.ErrConflict
		}
		if err := queries.LockGoogleDriveUserLifecycle(ctx, googledrivesql.LockGoogleDriveUserLifecycleParams{
			UserID: identity.UserID,
		}); err != nil {
			return err
		}
		if err := queries.LockGoogleDriveSubjectLifecycle(ctx, googledrivesql.LockGoogleDriveSubjectLifecycleParams{
			GoogleSubject: &identity.GoogleSubject,
		}); err != nil {
			return err
		}
		superseded, err := queries.SupersedeGoogleDriveRevocationIfSubjectActive(
			ctx,
			googledrivesql.SupersedeGoogleDriveRevocationIfSubjectActiveParams{
				SupersededAt: timePointer(claimedAt.UTC()), RevocationID: candidate.ID,
			},
		)
		if err != nil {
			return err
		}
		if superseded == 1 {
			return nil
		}
		row, err := queries.ClaimGoogleDriveRevocation(
			ctx,
			googledrivesql.ClaimGoogleDriveRevocationParams{
				ClaimToken: uuidPointer(claimToken), LeaseExpiresAt: timePointer(leaseExpiresAt.UTC()),
				ClaimedAt: claimedAt.UTC(), RevocationID: candidate.ID,
			},
		)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		claimed, err = revocationFromClaim(row)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return domain.Revocation{}, false, err
	}
	return claimed, claimed.ID != uuid.Nil, nil
}

func (repository *Repository) CompleteRevocation(
	ctx context.Context,
	revocationID, claimToken uuid.UUID,
	completedAt time.Time,
) error {
	rows, err := repository.queriesForContext(ctx).CompleteGoogleDriveRevocation(
		ctx,
		googledrivesql.CompleteGoogleDriveRevocationParams{
			CompletedAt: timePointer(completedAt.UTC()), RevocationID: revocationID, ClaimToken: uuidPointer(claimToken),
		},
	)
	return requireAffected(rows, err, domain.ErrConflict)
}

func (repository *Repository) RetryRevocation(
	ctx context.Context,
	revocationID, claimToken uuid.UUID,
	lastError string,
	availableAt, releasedAt time.Time,
	terminal bool,
) error {
	rows, err := repository.queriesForContext(ctx).RetryGoogleDriveRevocation(
		ctx,
		googledrivesql.RetryGoogleDriveRevocationParams{
			Terminal: terminal, LastError: lastError, AvailableAt: availableAt.UTC(),
			ReleasedAt: releasedAt.UTC(), RevocationID: revocationID, ClaimToken: uuidPointer(claimToken),
		},
	)
	return requireAffected(rows, err, domain.ErrConflict)
}

func revocationFromClaim(row googledrivesql.ClaimGoogleDriveRevocationRow) (domain.Revocation, error) {
	if row.CredentialPayload == nil || row.ClaimToken == nil || row.LeaseExpiresAt == nil {
		return domain.Revocation{}, domain.ErrConflict
	}
	return domain.Revocation{
		ID: row.RevocationID, SourceAccountID: row.SourceAccountID,
		UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		InstallationGeneration: row.InstallationGeneration,
		CredentialPayload:      *row.CredentialPayload, CredentialVersion: row.CredentialKeyVersion,
		AttemptCount: int(row.AttemptCount), ClaimToken: *row.ClaimToken,
		LeaseExpiresAt: *row.LeaseExpiresAt, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func timePointer(value time.Time) *time.Time { return &value }

func uuidPointer(value uuid.UUID) *uuid.UUID { return &value }
