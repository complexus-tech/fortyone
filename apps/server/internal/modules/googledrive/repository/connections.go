package googledriverepository

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) SaveOAuthState(ctx context.Context, state domain.OAuthState) error {
	queries := repository.queriesForContext(ctx)
	if _, err := queries.DeleteExpiredGoogleDriveOAuthStates(ctx); err != nil {
		return mapDatabaseError(err)
	}
	rows, err := queries.SaveGoogleDriveOAuthState(ctx, googledrivesql.SaveGoogleDriveOAuthStateParams{
		StateHash: state.StateHash, WorkspaceID: state.WorkspaceID, UserID: state.UserID,
		WorkspaceSlug: state.WorkspaceSlug, ReturnURL: state.ReturnURL,
		CodeVerifier: state.CodeVerifier, ExpiresAt: state.ExpiresAt.UTC(),
	})
	return requireAffected(rows, err, domain.ErrForbidden)
}

func (repository *Repository) ConsumeOAuthState(ctx context.Context, stateHash string, now time.Time) (domain.OAuthState, error) {
	consumedAt := now.UTC()
	row, err := repository.queriesForContext(ctx).ConsumeGoogleDriveOAuthState(ctx, googledrivesql.ConsumeGoogleDriveOAuthStateParams{
		StateHash: stateHash, ConsumedAt: consumedAt,
	})
	if err != nil {
		return domain.OAuthState{}, mapDatabaseError(err)
	}
	return domain.OAuthState{
		StateHash: row.StateHash, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		WorkspaceSlug: row.WorkspaceSlug, ReturnURL: row.ReturnURL,
		CodeVerifier: row.CodeVerifier, ExpiresAt: row.ExpiresAt,
	}, nil
}

func (repository *Repository) UpsertConnection(ctx context.Context, workspaceID uuid.UUID, account domain.Account) (domain.Connection, error) {
	var result domain.Connection
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		if err := queries.LockGoogleDriveUserLifecycle(ctx, googledrivesql.LockGoogleDriveUserLifecycleParams{
			UserID: account.UserID,
		}); err != nil {
			return err
		}
		if err := queries.LockGoogleDriveSubjectLifecycle(ctx, googledrivesql.LockGoogleDriveSubjectLifecycleParams{
			GoogleSubject: &account.GoogleSubject,
		}); err != nil {
			return err
		}
		existing, err := queries.GetActiveGoogleDriveAccountBySubject(
			ctx,
			googledrivesql.GetActiveGoogleDriveAccountBySubjectParams{GoogleSubject: account.GoogleSubject},
		)
		if err == nil && existing.UserID != account.UserID {
			return domain.ErrAccountOwned
		}
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return err
		}
		row, err := queries.UpsertGoogleDriveAccount(ctx, googledrivesql.UpsertGoogleDriveAccountParams{
			UserID: account.UserID, GoogleSubject: account.GoogleSubject, Email: account.Email,
			DisplayName: account.DisplayName, CredentialPayload: account.CredentialPayload,
			CredentialKeyVersion:   account.CredentialVersion,
			InstallationGeneration: account.InstallationGeneration,
			Scopes:                 account.Scopes, ExpiresAt: account.ExpiresAt.UTC(),
		})
		if err != nil {
			return err
		}
		rows, err := queries.UpsertGoogleDriveWorkspaceConnection(ctx, googledrivesql.UpsertGoogleDriveWorkspaceConnectionParams{
			UserID: account.UserID, AccountID: row.AccountID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		if rows != 1 {
			return domain.ErrForbidden
		}
		if _, err := queries.SupersedeGoogleDriveRevocationsForGeneration(
			ctx,
			googledrivesql.SupersedeGoogleDriveRevocationsForGenerationParams{
				GoogleSubject:    account.GoogleSubject,
				ActiveGeneration: account.InstallationGeneration,
				SupersededAt:     timePointer(repository.currentTime()),
			},
		); err != nil {
			return err
		}
		result = domain.Connection{WorkspaceID: workspaceID, Account: accountFromUpsert(row)}
		return nil
	})
	return result, err
}

func (repository *Repository) GetConnection(ctx context.Context, workspaceID, userID uuid.UUID) (domain.Connection, error) {
	row, err := repository.queriesForContext(ctx).GetGoogleDriveConnection(ctx, googledrivesql.GetGoogleDriveConnectionParams{
		WorkspaceID: workspaceID, UserID: userID,
	})
	if err != nil {
		return domain.Connection{}, mapDatabaseError(err)
	}
	return domain.Connection{WorkspaceID: row.WorkspaceID, Account: domain.Account{
		ID: row.AccountID, UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		Email: row.Email, DisplayName: row.DisplayName,
		CredentialPayload: row.CredentialPayload, CredentialVersion: row.CredentialKeyVersion,
		InstallationGeneration: row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: row.ExpiresAt, RequiresReauthorization: row.RequiresReauthorization,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}}, nil
}

func (repository *Repository) GetActiveAccountBySubject(ctx context.Context, googleSubject string) (domain.Account, error) {
	row, err := repository.queriesForContext(ctx).GetActiveGoogleDriveAccountBySubject(ctx, googledrivesql.GetActiveGoogleDriveAccountBySubjectParams{
		GoogleSubject: googleSubject,
	})
	if err != nil {
		return domain.Account{}, mapDatabaseError(err)
	}
	return domain.Account{
		ID: row.AccountID, UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		Email: row.Email, DisplayName: row.DisplayName,
		CredentialPayload: row.CredentialPayload, CredentialVersion: row.CredentialKeyVersion,
		InstallationGeneration: row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: row.ExpiresAt, RequiresReauthorization: row.RequiresReauthorization,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (repository *Repository) CompareAndSwapCredential(
	ctx context.Context,
	account domain.Account,
	nextPayload string,
	expiresAt time.Time,
) (bool, error) {
	rows, err := repository.queriesForContext(ctx).CompareAndSwapGoogleDriveCredential(ctx, googledrivesql.CompareAndSwapGoogleDriveCredentialParams{
		NextPayload: nextPayload, CredentialKeyVersion: int16(credentialvault.CurrentVersion),
		ExpiresAt: expiresAt.UTC(), AccountID: account.ID,
		InstallationGeneration: account.InstallationGeneration,
		PreviousPayload:        account.CredentialPayload,
	})
	if err != nil {
		return false, mapDatabaseError(err)
	}
	return rows == 1, nil
}

func (repository *Repository) MarkReauthorizationRequired(ctx context.Context, account domain.Account, errorCode string) error {
	rows, err := repository.queriesForContext(ctx).MarkGoogleDriveReauthorizationRequired(ctx, googledrivesql.MarkGoogleDriveReauthorizationRequiredParams{
		ErrorCode: &errorCode, AccountID: account.ID,
		InstallationGeneration: account.InstallationGeneration,
	})
	return requireAffected(rows, err, domain.ErrConflict)
}

// Disconnect removes one workspace binding. The final-binding database trigger
// atomically stages the existing sealed credential for asynchronous provider
// revocation before destroying local account state. No provider call runs in
// this transaction.
func (repository *Repository) Disconnect(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
) (bool, error) {
	orphaned := false
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		if err := queries.LockGoogleDriveUserLifecycle(ctx, googledrivesql.LockGoogleDriveUserLifecycleParams{
			UserID: userID,
		}); err != nil {
			return err
		}
		accountID, err := queries.DeleteGoogleDriveWorkspaceConnection(ctx, googledrivesql.DeleteGoogleDriveWorkspaceConnectionParams{
			WorkspaceID: workspaceID, UserID: userID,
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil
			}
			return err
		}
		count, err := queries.CountGoogleDriveAccountConnections(ctx, googledrivesql.CountGoogleDriveAccountConnectionsParams{AccountID: accountID})
		if err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		_, err = queries.RevokeUnusedGoogleDriveAccount(ctx, googledrivesql.RevokeUnusedGoogleDriveAccountParams{
			AccountID: accountID, UserID: userID,
		})
		if err != nil {
			return err
		}
		orphaned = true
		return nil
	})
	return orphaned, err
}

func accountFromUpsert(row googledrivesql.UpsertGoogleDriveAccountRow) domain.Account {
	return domain.Account{
		ID: row.AccountID, UserID: row.UserID, GoogleSubject: row.GoogleSubject,
		Email: row.Email, DisplayName: row.DisplayName,
		CredentialPayload: row.CredentialPayload, CredentialVersion: row.CredentialKeyVersion,
		InstallationGeneration: row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: row.ExpiresAt, RequiresReauthorization: row.RequiresReauthorization,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
