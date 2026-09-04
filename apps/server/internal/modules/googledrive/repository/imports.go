package googledriverepository

import (
	"context"
	"errors"
	"time"
	"unicode/utf8"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (repository *Repository) CreateImportOperation(
	ctx context.Context,
	operation domain.ImportOperation,
) (domain.ImportOperation, bool, error) {
	row, err := repository.queries.CreateGoogleDriveImportOperation(ctx, googledrivesql.CreateGoogleDriveImportOperationParams{
		WorkspaceID: operation.WorkspaceID, UserID: operation.UserID,
		SourceReferenceID: operation.SourceReferenceID, DocumentID: operation.DocumentID,
		IdempotencyKey: operation.IdempotencyKey, RequestHash: operation.RequestHash,
		Visibility: operation.Visibility, AttemptGeneration: operation.AttemptGeneration,
	})
	if err == nil {
		return mapCreatedImportOperation(row), true, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.ImportOperation{}, false, mapDatabaseError(err)
	}
	existing, err := repository.queries.GetGoogleDriveImportOperation(ctx, googledrivesql.GetGoogleDriveImportOperationParams{
		WorkspaceID: operation.WorkspaceID, UserID: operation.UserID,
		IdempotencyKey: operation.IdempotencyKey,
	})
	if err != nil {
		return domain.ImportOperation{}, false, mapDatabaseError(err)
	}
	return mapExistingImportOperation(existing), false, nil
}

func (repository *Repository) ClaimImportOperation(
	ctx context.Context,
	operationID, attemptGeneration uuid.UUID,
	previousUpdatedAt, staleBefore time.Time,
) (domain.ImportOperation, bool, error) {
	row, err := repository.queries.ClaimGoogleDriveImportOperation(ctx, googledrivesql.ClaimGoogleDriveImportOperationParams{
		OperationID: operationID, AttemptGeneration: attemptGeneration,
		PreviousUpdatedAt: previousUpdatedAt, StaleBefore: staleBefore,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ImportOperation{}, false, nil
	}
	if err != nil {
		return domain.ImportOperation{}, false, mapDatabaseError(err)
	}
	return mapClaimedImportOperation(row), true, nil
}

func (repository *Repository) FailImportOperation(
	ctx context.Context,
	operationID, attemptGeneration uuid.UUID,
	errorCode string,
) error {
	rows, err := repository.queries.FailGoogleDriveImportOperation(ctx, googledrivesql.FailGoogleDriveImportOperationParams{
		ErrorCode: &errorCode, OperationID: operationID, AttemptGeneration: attemptGeneration,
	})
	return requireAffected(rows, err, domain.ErrConflict)
}

// FinalizeDocumentImport creates the native document, records its Google
// provenance, and marks the idempotency operation complete in one transaction.
// A lost HTTP response can therefore only observe all three writes or none.
func (repository *Repository) FinalizeDocumentImport(
	ctx context.Context,
	input domain.ImportFinalization,
) (uuid.UUID, error) {
	if !validImportFinalization(input) {
		return uuid.Nil, domain.ErrInvalidInput
	}

	documentID := uuid.Nil
	err := repository.withinTransaction(ctx, func(queries googledrivesql.Querier) error {
		row, err := queries.LockGoogleDriveImportOperation(ctx, googledrivesql.LockGoogleDriveImportOperationParams{
			OperationID: input.Operation.ID, WorkspaceID: input.Operation.WorkspaceID,
			UserID: input.Operation.UserID,
		})
		if err != nil {
			return err
		}
		operation := mapLockedImportOperation(row)
		if !sameImportOperation(operation, input.Operation) {
			return domain.ErrConflict
		}
		if operation.Status == domain.ImportOperationCompleted {
			documentID = operation.DocumentID
			return nil
		}
		if operation.Status != domain.ImportOperationPending {
			return domain.ErrConflict
		}
		if operation.AttemptGeneration != input.Operation.AttemptGeneration {
			return domain.ErrOperationInProgress
		}

		targetID := input.TargetID
		importable, err := queries.GoogleDriveReferenceImportable(ctx, googledrivesql.GoogleDriveReferenceImportableParams{
			UserID: input.Operation.UserID, AccountID: input.AccountID,
			GrantGeneration: input.GrantGeneration, ReferenceID: operation.SourceReferenceID,
			WorkspaceID: input.Operation.WorkspaceID, TargetType: string(input.TargetType),
			TargetID: &targetID, GoogleFileID: input.GoogleFileID,
		})
		if err != nil {
			return err
		}
		if !importable {
			return domain.ErrForbidden
		}
		mutable, err := queries.GoogleDriveTargetMutable(ctx, googledrivesql.GoogleDriveTargetMutableParams{
			TargetType: string(input.TargetType), UserID: input.Operation.UserID,
			TargetID: input.TargetID, WorkspaceID: input.Operation.WorkspaceID,
		})
		if err != nil {
			return err
		}
		if !mutable {
			return domain.ErrForbidden
		}

		createdID, err := queries.CreateGoogleDriveImportedDocument(ctx, googledrivesql.CreateGoogleDriveImportedDocumentParams{
			DocumentID: operation.DocumentID, WorkspaceID: operation.WorkspaceID,
			UserID: operation.UserID, Title: input.Title, Visibility: operation.Visibility,
			ContentHtml: input.ContentHTML, ContentText: input.ContentText,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrForbidden
		}
		if err != nil {
			return err
		}
		if createdID != operation.DocumentID {
			return domain.ErrConflict
		}
		rows, err := queries.SaveGoogleDriveDocumentImport(ctx, googledrivesql.SaveGoogleDriveDocumentImportParams{
			WorkspaceID: operation.WorkspaceID, DocumentID: operation.DocumentID,
			ReferenceID: &operation.SourceReferenceID, GoogleFileID: input.GoogleFileID,
			SourceVersion: input.SourceVersion, ImportedByUserID: operation.UserID,
		})
		if err := requireAffected(rows, err, domain.ErrConflict); err != nil {
			return err
		}
		rows, err = queries.CompleteGoogleDriveImportOperation(ctx, googledrivesql.CompleteGoogleDriveImportOperationParams{
			OperationID: operation.ID, AttemptGeneration: operation.AttemptGeneration,
		})
		if err := requireAffected(rows, err, domain.ErrOperationInProgress); err != nil {
			return err
		}
		documentID = operation.DocumentID
		return nil
	})
	if err != nil {
		return uuid.Nil, err
	}
	return documentID, nil
}

func validImportFinalization(input domain.ImportFinalization) bool {
	operation := input.Operation
	return operation.ID != uuid.Nil && operation.WorkspaceID != uuid.Nil && operation.UserID != uuid.Nil &&
		operation.SourceReferenceID != uuid.Nil && operation.DocumentID != uuid.Nil &&
		operation.AttemptGeneration != uuid.Nil && len(operation.RequestHash) == 64 &&
		(operation.Visibility == "workspace" || operation.Visibility == "private") &&
		input.AccountID != uuid.Nil && input.GrantGeneration != uuid.Nil && input.TargetType.Valid() &&
		input.TargetID != uuid.Nil && input.GoogleFileID != "" && input.Title != "" &&
		utf8.RuneCountInString(input.Title) <= 255
}

func sameImportOperation(stored, expected domain.ImportOperation) bool {
	return stored.ID == expected.ID && stored.WorkspaceID == expected.WorkspaceID &&
		stored.UserID == expected.UserID && stored.SourceReferenceID == expected.SourceReferenceID &&
		stored.DocumentID == expected.DocumentID && stored.IdempotencyKey == expected.IdempotencyKey &&
		stored.RequestHash == expected.RequestHash && stored.Visibility == expected.Visibility
}

func mapCreatedImportOperation(row googledrivesql.CreateGoogleDriveImportOperationRow) domain.ImportOperation {
	return domain.ImportOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		SourceReferenceID: row.SourceReferenceID, DocumentID: row.DocumentID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		Visibility: row.Visibility, AttemptGeneration: row.AttemptGeneration,
		Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		CompletedAt: row.CompletedAt,
	}
}

func mapExistingImportOperation(row googledrivesql.GetGoogleDriveImportOperationRow) domain.ImportOperation {
	return domain.ImportOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		SourceReferenceID: row.SourceReferenceID, DocumentID: row.DocumentID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		Visibility: row.Visibility, AttemptGeneration: row.AttemptGeneration,
		Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		CompletedAt: row.CompletedAt,
	}
}

func mapClaimedImportOperation(row googledrivesql.ClaimGoogleDriveImportOperationRow) domain.ImportOperation {
	return domain.ImportOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		SourceReferenceID: row.SourceReferenceID, DocumentID: row.DocumentID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		Visibility: row.Visibility, AttemptGeneration: row.AttemptGeneration,
		Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		CompletedAt: row.CompletedAt,
	}
}

func mapLockedImportOperation(row googledrivesql.LockGoogleDriveImportOperationRow) domain.ImportOperation {
	return domain.ImportOperation{
		ID: row.OperationID, WorkspaceID: row.WorkspaceID, UserID: row.UserID,
		SourceReferenceID: row.SourceReferenceID, DocumentID: row.DocumentID,
		IdempotencyKey: row.IdempotencyKey, RequestHash: row.RequestHash,
		Visibility: row.Visibility, AttemptGeneration: row.AttemptGeneration,
		Status: row.Status, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		CompletedAt: row.CompletedAt,
	}
}
