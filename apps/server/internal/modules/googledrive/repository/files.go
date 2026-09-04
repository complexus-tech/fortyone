package googledriverepository

import (
	"context"
	"errors"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) TargetAccessible(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	targetType domain.TargetType,
	targetID uuid.UUID,
) (bool, error) {
	accessible, err := repository.queries.GoogleDriveTargetAccessible(ctx, googledrivesql.GoogleDriveTargetAccessibleParams{
		TargetType: string(targetType), UserID: userID, TargetID: targetID, WorkspaceID: workspaceID,
	})
	return accessible, mapDatabaseError(err)
}

func (repository *Repository) TargetMutable(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	targetType domain.TargetType,
	targetID uuid.UUID,
) (bool, error) {
	mutable, err := repository.queries.GoogleDriveTargetMutable(ctx, googledrivesql.GoogleDriveTargetMutableParams{
		TargetType: string(targetType), UserID: userID, TargetID: targetID, WorkspaceID: workspaceID,
	})
	return mutable, mapDatabaseError(err)
}

func (repository *Repository) AttachFile(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	accountID uuid.UUID,
	targetType domain.TargetType,
	targetID uuid.UUID,
	providerFile domain.ProviderFile,
) (domain.FileReference, error) {
	references, err := repository.AttachFiles(
		ctx, workspaceID, userID, accountID, targetType, targetID,
		[]domain.ProviderFile{providerFile},
	)
	if err != nil {
		return domain.FileReference{}, err
	}
	if len(references) != 1 {
		return domain.FileReference{}, domain.ErrConflict
	}
	return references[0], nil
}

func (repository *Repository) AttachFiles(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	accountID uuid.UUID,
	targetType domain.TargetType,
	targetID uuid.UUID,
	providerFiles []domain.ProviderFile,
) ([]domain.FileReference, error) {
	referenceIDs := make([]uuid.UUID, 0, len(providerFiles))
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		accessible, err := queries.GoogleDriveTargetMutable(ctx, googledrivesql.GoogleDriveTargetMutableParams{
			TargetType: string(targetType), UserID: userID, TargetID: targetID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		if !accessible {
			return domain.ErrForbidden
		}
		for _, providerFile := range providerFiles {
			file, err := queries.UpsertGoogleDriveFile(ctx, googledrivesql.UpsertGoogleDriveFileParams{
				WorkspaceID: workspaceID, GoogleFileID: providerFile.ID,
				ResourceKey: providerFile.ResourceKey, Name: providerFile.Name,
				MimeType: providerFile.MimeType, WebViewLink: providerFile.WebViewLink,
				DriveID: providerFile.DriveID, Version: providerFile.Version,
				SizeBytes: providerFile.SizeBytes, ModifiedAt: providerFile.ModifiedAt,
				Metadata: providerFile.Metadata,
			})
			if err != nil {
				return err
			}
			rows, err := queries.UpsertGoogleDriveFileGrant(ctx, googledrivesql.UpsertGoogleDriveFileGrantParams{
				FileID: file.FileID, UserID: userID, AccountID: accountID, WorkspaceID: workspaceID,
				GrantGeneration: uuid.New(),
			})
			if err := requireAffected(rows, err, domain.ErrForbidden); err != nil {
				return err
			}
			params := googledrivesql.UpsertGoogleDriveFileReferenceParams{
				WorkspaceID: workspaceID, FileID: file.FileID,
				TargetType: string(targetType), CreatedByUserID: userID,
			}
			setTargetColumns(&params, targetType, targetID)
			referenceID, err := queries.UpsertGoogleDriveFileReference(ctx, params)
			if errors.Is(err, pgx.ErrNoRows) {
				referenceID, err = queries.FindGoogleDriveFileReference(ctx, googledrivesql.FindGoogleDriveFileReferenceParams{
					FileID: file.FileID, TargetType: string(targetType), TargetID: &targetID,
				})
			}
			if err != nil {
				return err
			}
			referenceIDs = append(referenceIDs, referenceID)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	references := make([]domain.FileReference, 0, len(referenceIDs))
	for _, referenceID := range referenceIDs {
		reference, err := repository.GetReference(ctx, workspaceID, userID, referenceID)
		if err != nil {
			return nil, err
		}
		references = append(references, reference)
	}
	return references, nil
}

// RevalidateReference refreshes provider metadata and restores only the
// requesting actor's grant after the existing target authorization is checked
// again inside the transaction. It never creates a new target reference.
func (repository *Repository) RevalidateReference(
	ctx context.Context,
	workspaceID, userID, accountID uuid.UUID,
	reference domain.FileReference,
	providerFile domain.ProviderFile,
) (uuid.UUID, error) {
	grantGeneration := uuid.New()
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		accessible, err := queries.GoogleDriveTargetAccessible(ctx, googledrivesql.GoogleDriveTargetAccessibleParams{
			TargetType: string(reference.TargetType), UserID: userID,
			TargetID: reference.TargetID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		if !accessible {
			return domain.ErrForbidden
		}
		fileID, err := queries.RevalidateGoogleDriveFileReference(ctx, googledrivesql.RevalidateGoogleDriveFileReferenceParams{
			ResourceKey: providerFile.ResourceKey, Name: providerFile.Name,
			MimeType: providerFile.MimeType, WebViewLink: providerFile.WebViewLink,
			DriveID: providerFile.DriveID, Version: providerFile.Version,
			SizeBytes: providerFile.SizeBytes, ModifiedAt: providerFile.ModifiedAt,
			Metadata: providerFile.Metadata, ReferenceID: reference.ID,
			WorkspaceID: workspaceID, GoogleFileID: providerFile.ID,
		})
		if err != nil {
			return err
		}
		rows, err := queries.UpsertGoogleDriveFileGrant(ctx, googledrivesql.UpsertGoogleDriveFileGrantParams{
			FileID: fileID, UserID: userID, AccountID: accountID, WorkspaceID: workspaceID,
			GrantGeneration: grantGeneration,
		})
		return requireAffected(rows, err, domain.ErrForbidden)
	})
	if err != nil {
		return uuid.Nil, err
	}
	return grantGeneration, nil
}

func (repository *Repository) DeleteReferenceGrant(
	ctx context.Context,
	workspaceID, userID, accountID, referenceID, grantGeneration uuid.UUID,
) error {
	rows, err := repository.queries.DeleteGoogleDriveFileGrantForActor(ctx, googledrivesql.DeleteGoogleDriveFileGrantForActorParams{
		ReferenceID: referenceID, WorkspaceID: workspaceID,
		UserID: userID, AccountID: accountID, GrantGeneration: grantGeneration,
	})
	if err != nil {
		return mapDatabaseError(err)
	}
	if rows > 1 {
		return domain.ErrConflict
	}
	return nil
}

func (repository *Repository) MarkReferenceUnavailable(
	ctx context.Context,
	workspaceID, referenceID uuid.UUID,
) error {
	rows, err := repository.queries.MarkGoogleDriveFileUnavailable(ctx, googledrivesql.MarkGoogleDriveFileUnavailableParams{
		ReferenceID: referenceID, WorkspaceID: workspaceID,
	})
	return requireAffected(rows, err, domain.ErrNotFound)
}

func (repository *Repository) ListReferences(
	ctx context.Context,
	workspaceID, userID uuid.UUID,
	targetType domain.TargetType,
	targetID uuid.UUID,
) ([]domain.FileReference, error) {
	accessible, err := repository.TargetAccessible(ctx, workspaceID, userID, targetType, targetID)
	if err != nil {
		return nil, err
	}
	if !accessible {
		return nil, domain.ErrForbidden
	}
	rows, err := repository.queries.ListGoogleDriveFileReferences(ctx, googledrivesql.ListGoogleDriveFileReferencesParams{
		UserID: userID, WorkspaceID: workspaceID, TargetType: string(targetType), TargetID: &targetID,
	})
	if err != nil {
		return nil, mapDatabaseError(err)
	}
	result := make([]domain.FileReference, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapListReference(row))
	}
	return result, nil
}

func (repository *Repository) GetReference(
	ctx context.Context,
	workspaceID, userID, referenceID uuid.UUID,
) (domain.FileReference, error) {
	row, err := repository.queries.GetGoogleDriveFileReference(ctx, googledrivesql.GetGoogleDriveFileReferenceParams{
		UserID: userID, ReferenceID: referenceID, WorkspaceID: workspaceID,
	})
	if err != nil {
		return domain.FileReference{}, mapDatabaseError(err)
	}
	if row.TargetID == nil {
		return domain.FileReference{}, domain.ErrNotFound
	}
	accessible, err := repository.TargetAccessible(ctx, workspaceID, userID, domain.TargetType(row.TargetType), *row.TargetID)
	if err != nil {
		return domain.FileReference{}, err
	}
	if !accessible {
		return domain.FileReference{}, domain.ErrForbidden
	}
	return mapReference(row), nil
}

func (repository *Repository) DeleteReference(
	ctx context.Context,
	workspaceID, userID, referenceID uuid.UUID,
) error {
	reference, err := repository.GetReference(ctx, workspaceID, userID, referenceID)
	if err != nil {
		return err
	}
	accessible, err := repository.TargetMutable(ctx, workspaceID, userID, reference.TargetType, reference.TargetID)
	if err != nil {
		return err
	}
	if !accessible {
		return domain.ErrForbidden
	}
	return repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		mutable, err := queries.GoogleDriveTargetMutable(ctx, googledrivesql.GoogleDriveTargetMutableParams{
			TargetType: string(reference.TargetType), UserID: userID,
			TargetID: reference.TargetID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		if !mutable {
			return domain.ErrForbidden
		}
		fileID, err := queries.DeleteGoogleDriveFileReference(ctx, googledrivesql.DeleteGoogleDriveFileReferenceParams{
			ReferenceID: referenceID, WorkspaceID: workspaceID,
		})
		if err != nil {
			return err
		}
		_, err = queries.DeleteOrphanGoogleDriveFile(ctx, googledrivesql.DeleteOrphanGoogleDriveFileParams{
			FileID: fileID, WorkspaceID: workspaceID,
		})
		return err
	})
}

func (repository *Repository) CreateOperation(ctx context.Context, operation domain.CreateOperation) (domain.CreateOperation, bool, error) {
	row, err := repository.queries.CreateGoogleDriveOperation(ctx, googledrivesql.CreateGoogleDriveOperationParams{
		WorkspaceID: operation.WorkspaceID, UserID: operation.UserID,
		IdempotencyKey: operation.IdempotencyKey, RequestHash: operation.RequestHash,
		TargetType: string(operation.TargetType), TargetID: operation.TargetID,
		FileType: string(operation.FileType), Title: operation.Title,
	})
	if err == nil {
		return mapCreateOperation(row), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.CreateOperation{}, false, mapDatabaseError(err)
	}
	existing, err := repository.queries.GetGoogleDriveOperation(ctx, googledrivesql.GetGoogleDriveOperationParams{
		WorkspaceID: operation.WorkspaceID, UserID: operation.UserID,
		IdempotencyKey: operation.IdempotencyKey,
	})
	if err != nil {
		return domain.CreateOperation{}, false, mapDatabaseError(err)
	}
	return mapGetOperation(existing), false, nil
}

func (repository *Repository) ClaimOperation(
	ctx context.Context,
	operationID uuid.UUID,
	previousUpdatedAt, staleBefore time.Time,
) (domain.CreateOperation, bool, error) {
	row, err := repository.queries.ClaimGoogleDriveOperation(ctx, googledrivesql.ClaimGoogleDriveOperationParams{
		OperationID: operationID, PreviousUpdatedAt: previousUpdatedAt, StaleBefore: staleBefore,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.CreateOperation{}, false, nil
	}
	if err != nil {
		return domain.CreateOperation{}, false, mapDatabaseError(err)
	}
	return mapClaimOperation(row), true, nil
}

func (repository *Repository) CompleteOperation(ctx context.Context, operationID uuid.UUID, googleFileID string, referenceID uuid.UUID) error {
	rows, err := repository.queries.CompleteGoogleDriveOperation(ctx, googledrivesql.CompleteGoogleDriveOperationParams{
		GoogleFileID: &googleFileID, ReferenceID: &referenceID, OperationID: operationID,
	})
	return requireAffected(rows, err, domain.ErrConflict)
}

func (repository *Repository) FailOperation(ctx context.Context, operationID uuid.UUID, errorCode string) error {
	rows, err := repository.queries.FailGoogleDriveOperation(ctx, googledrivesql.FailGoogleDriveOperationParams{
		ErrorCode: &errorCode, OperationID: operationID,
	})
	return requireAffected(rows, err, domain.ErrConflict)
}

func setTargetColumns(params *googledrivesql.UpsertGoogleDriveFileReferenceParams, targetType domain.TargetType, targetID uuid.UUID) {
	switch targetType {
	case domain.TargetStory:
		params.StoryID = &targetID
	case domain.TargetObjective:
		params.ObjectiveID = &targetID
	case domain.TargetDocument:
		params.DocumentID = &targetID
	case domain.TargetComment:
		params.CommentID = &targetID
	}
}

func mapListReference(row googledrivesql.ListGoogleDriveFileReferencesRow) domain.FileReference {
	targetID := uuid.Nil
	if row.TargetID != nil {
		targetID = *row.TargetID
	}
	availability := availability(row.UnavailableAt, row.ConnectionEmail, row.RequiresReauthorization)
	return domain.FileReference{
		ID: row.ReferenceID, InternalFileID: row.InternalFileID, FileID: row.GoogleFileID,
		ResourceKey: row.ResourceKey, Name: row.Name, MimeType: row.MimeType,
		WebViewLink: row.WebViewLink, Version: row.Version, ModifiedTime: row.ModifiedAt,
		ConnectionEmail: row.ConnectionEmail, Availability: availability,
		TargetType: domain.TargetType(row.TargetType), TargetID: targetID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapReference(row googledrivesql.GetGoogleDriveFileReferenceRow) domain.FileReference {
	targetID := uuid.Nil
	if row.TargetID != nil {
		targetID = *row.TargetID
	}
	result := domain.FileReference{
		ID: row.ReferenceID, InternalFileID: row.InternalFileID, FileID: row.GoogleFileID,
		ResourceKey: row.ResourceKey, Name: row.Name, MimeType: row.MimeType,
		WebViewLink: row.WebViewLink, Version: row.Version, ModifiedTime: row.ModifiedAt,
		ConnectionEmail: row.ConnectionEmail,
		Availability:    availability(row.UnavailableAt, row.ConnectionEmail, row.RequiresReauthorization),
		TargetType:      domain.TargetType(row.TargetType), TargetID: targetID,
		GrantGeneration: row.GrantGeneration,
		CreatedAt:       row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
	if row.AccountID == nil || row.AccountUserID == nil || row.GoogleSubject == nil ||
		row.CredentialPayload == nil || row.CredentialKeyVersion == nil ||
		row.InstallationGeneration == nil || row.ExpiresAt == nil ||
		row.RequiresReauthorization == nil || row.AccountCreatedAt == nil || row.AccountUpdatedAt == nil {
		return result
	}
	result.Account = &domain.Account{
		ID: *row.AccountID, UserID: *row.AccountUserID, GoogleSubject: *row.GoogleSubject,
		Email: pointerValue(row.ConnectionEmail), DisplayName: row.DisplayName,
		CredentialPayload: *row.CredentialPayload, CredentialVersion: *row.CredentialKeyVersion,
		InstallationGeneration: *row.InstallationGeneration, Scopes: row.Scopes,
		ExpiresAt: *row.ExpiresAt, RequiresReauthorization: *row.RequiresReauthorization,
		CreatedAt: *row.AccountCreatedAt, UpdatedAt: *row.AccountUpdatedAt,
	}
	return result
}

func availability(unavailableAt *time.Time, email *string, requiresReauthorization *bool) string {
	if unavailableAt != nil {
		return "deleted"
	}
	if email == nil {
		return "access_required"
	}
	if requiresReauthorization != nil && *requiresReauthorization {
		return "reauthorization_required"
	}
	return "available"
}

func pointerValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func mapCreateOperation(row googledrivesql.CreateGoogleDriveOperationRow) domain.CreateOperation {
	return domain.CreateOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		TargetType: domain.TargetType(row.TargetType), TargetID: row.TargetID,
		FileType: domain.FileType(row.FileType), Title: row.Title, Status: row.Status,
		GoogleFileID: row.GoogleFileID, ReferenceID: row.ReferenceID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapGetOperation(row googledrivesql.GetGoogleDriveOperationRow) domain.CreateOperation {
	return domain.CreateOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		TargetType: domain.TargetType(row.TargetType), TargetID: row.TargetID,
		FileType: domain.FileType(row.FileType), Title: row.Title, Status: row.Status,
		GoogleFileID: row.GoogleFileID, ReferenceID: row.ReferenceID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}

func mapClaimOperation(row googledrivesql.ClaimGoogleDriveOperationRow) domain.CreateOperation {
	return domain.CreateOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		TargetType: domain.TargetType(row.TargetType), TargetID: row.TargetID,
		FileType: domain.FileType(row.FileType), Title: row.Title, Status: row.Status,
		GoogleFileID: row.GoogleFileID, ReferenceID: row.ReferenceID,
		CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}
}
