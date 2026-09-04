package googledrive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"strings"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/google/uuid"
)

const (
	maxImportBytes             int64 = 512 << 10
	importOperationLease             = 2 * time.Minute
	importFailureRecordTimeout       = 3 * time.Second
)

func (service *Service) ImportFile(ctx context.Context, input ImportInput) (ImportResult, error) {
	if !service.configured() {
		return ImportResult{}, domain.ErrNotConfigured
	}
	input.IdempotencyKey = strings.TrimSpace(input.IdempotencyKey)
	if input.WorkspaceID == uuid.Nil || input.UserID == uuid.Nil || input.ReferenceID == uuid.Nil ||
		(input.Visibility != "workspace" && input.Visibility != "private") ||
		input.IdempotencyKey == "" || len(input.IdempotencyKey) > 200 {
		return ImportResult{}, domain.ErrInvalidInput
	}

	requestHash := importRequestHash(input)
	operation, shouldRun, err := service.prepareImportOperation(ctx, input, requestHash)
	if err != nil {
		return ImportResult{}, err
	}
	if !shouldRun {
		return importResult(operation), nil
	}

	result, err := service.executeImport(ctx, input, operation)
	if err != nil {
		service.failImportOperation(ctx, operation, "import_failed")
		return ImportResult{}, err
	}
	return result, nil
}

func (service *Service) prepareImportOperation(
	ctx context.Context,
	input ImportInput,
	requestHash string,
) (domain.ImportOperation, bool, error) {
	operation, created, err := service.repo.CreateImportOperation(ctx, domain.ImportOperation{
		WorkspaceID: input.WorkspaceID, UserID: input.UserID,
		SourceReferenceID: input.ReferenceID, DocumentID: uuid.New(),
		IdempotencyKey: input.IdempotencyKey, RequestHash: requestHash,
		Visibility: input.Visibility, AttemptGeneration: uuid.New(),
	})
	if err != nil {
		return domain.ImportOperation{}, false, err
	}
	if operation.RequestHash != requestHash || operation.SourceReferenceID != input.ReferenceID ||
		operation.Visibility != input.Visibility {
		return domain.ImportOperation{}, false, domain.ErrConflict
	}
	if operation.Status == domain.ImportOperationCompleted {
		return operation, false, nil
	}
	if created {
		return operation, true, nil
	}

	now := service.currentImportTime()
	switch operation.Status {
	case domain.ImportOperationPending:
		if now.Before(operation.UpdatedAt.Add(importOperationLease)) {
			return domain.ImportOperation{}, false, domain.ErrOperationInProgress
		}
	case domain.ImportOperationFailed:
		// A failed attempt has no native document: document creation, provenance,
		// and completion roll back as a unit. It is safe to retry immediately.
	default:
		return domain.ImportOperation{}, false, domain.ErrConflict
	}

	operation, claimed, err := service.repo.ClaimImportOperation(
		ctx, operation.ID, uuid.New(), operation.UpdatedAt, now.Add(-importOperationLease),
	)
	if err != nil {
		return domain.ImportOperation{}, false, err
	}
	if !claimed {
		return domain.ImportOperation{}, false, domain.ErrOperationInProgress
	}
	return operation, true, nil
}

func (service *Service) executeImport(
	ctx context.Context,
	input ImportInput,
	operation domain.ImportOperation,
) (ImportResult, error) {
	reference, err := service.repo.GetReference(ctx, input.WorkspaceID, input.UserID, input.ReferenceID)
	if err != nil {
		return ImportResult{}, err
	}
	if reference.MimeType != googleDocumentMimeType {
		return ImportResult{}, fmt.Errorf("%w: only Google Docs can be imported as FortyOne documents", domain.ErrInvalidInput)
	}
	mutable, err := service.repo.TargetMutable(ctx, input.WorkspaceID, input.UserID, reference.TargetType, reference.TargetID)
	if err != nil {
		return ImportResult{}, err
	}
	if !mutable {
		return ImportResult{}, domain.ErrForbidden
	}
	if reference.Account == nil {
		return ImportResult{}, domain.ErrForbidden
	}
	_, token, err := service.accountToken(ctx, input.WorkspaceID, input.UserID, *reference.Account)
	if err != nil {
		return ImportResult{}, err
	}
	providerFile, err := service.client.GetFile(ctx, token.AccessToken, reference.FileID, reference.ResourceKey)
	if err != nil {
		return ImportResult{}, service.handleReferenceProviderError(
			ctx, input.WorkspaceID, input.UserID, reference, *reference.Account, err,
		)
	}
	if providerFile.Trashed {
		return ImportResult{}, service.markReferenceUnavailable(
			ctx, input.WorkspaceID, input.UserID, reference.ID,
		)
	}
	if providerFile.MimeType != googleDocumentMimeType {
		if _, err := service.repo.RevalidateReference(
			ctx, input.WorkspaceID, input.UserID, reference.Account.ID, reference, providerFile,
		); err != nil {
			return ImportResult{}, err
		}
		return ImportResult{}, fmt.Errorf("%w: only Google Docs can be imported as FortyOne documents", domain.ErrInvalidInput)
	}
	grantGeneration, err := service.repo.RevalidateReference(
		ctx, input.WorkspaceID, input.UserID, reference.Account.ID, reference, providerFile,
	)
	if err != nil {
		return ImportResult{}, err
	}
	reference.GrantGeneration = &grantGeneration
	content, err := service.client.ReadFile(ctx, token.AccessToken, providerFile, maxImportBytes)
	if err != nil {
		return ImportResult{}, service.handleReferenceProviderError(
			ctx, input.WorkspaceID, input.UserID, reference, *reference.Account, err,
		)
	}
	if content.Truncated {
		return ImportResult{}, fmt.Errorf("%w: Google Doc is larger than the supported import limit", domain.ErrContentTooLarge)
	}
	documentID, err := service.repo.FinalizeDocumentImport(ctx, domain.ImportFinalization{
		Operation: operation, AccountID: reference.Account.ID, GrantGeneration: grantGeneration,
		TargetType: reference.TargetType, TargetID: reference.TargetID,
		GoogleFileID: providerFile.ID, SourceVersion: providerFile.Version,
		Title: normalizedImportTitle(providerFile.Name), ContentHTML: textToHTML(content.Text),
		ContentText: content.Text,
	})
	if err != nil {
		return ImportResult{}, err
	}
	return ImportResult{DocumentID: documentID, SourceReferenceID: operation.SourceReferenceID}, nil
}

func (service *Service) failImportOperation(ctx context.Context, operation domain.ImportOperation, errorCode string) {
	writeCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), importFailureRecordTimeout)
	defer cancel()
	err := service.repo.FailImportOperation(writeCtx, operation.ID, operation.AttemptGeneration, errorCode)
	if err != nil && !errors.Is(err, domain.ErrConflict) && service.log != nil {
		service.log.Error(writeCtx, "failed recording Google Drive import operation failure", "error", err, "operation_id", operation.ID)
	}
}

func (service *Service) currentImportTime() time.Time {
	if service.now == nil {
		return time.Now().UTC()
	}
	return service.now().UTC()
}

func importRequestHash(input ImportInput) string {
	payload := input.ReferenceID.String() + "\x00" + input.Visibility
	digest := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(digest[:])
}

func importResult(operation domain.ImportOperation) ImportResult {
	return ImportResult{
		DocumentID:        operation.DocumentID,
		SourceReferenceID: operation.SourceReferenceID,
	}
}

func normalizedImportTitle(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "Imported Google Doc"
	}
	runes := []rune(value)
	if len(runes) <= maxTitleRunes {
		return value
	}
	return strings.TrimSpace(string(runes[:maxTitleRunes]))
}

func textToHTML(value string) string {
	lines := strings.Split(strings.ReplaceAll(value, "\r\n", "\n"), "\n")
	paragraphs := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		paragraphs = append(paragraphs, "<p>"+html.EscapeString(line)+"</p>")
	}
	return strings.Join(paragraphs, "")
}
