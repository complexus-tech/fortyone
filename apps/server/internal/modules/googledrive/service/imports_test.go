package googledrive

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	"github.com/complexus-tech/projects-api/internal/platform/credentialvault"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestImportFileRequiresAnIdempotencyKey(t *testing.T) {
	t.Parallel()

	repo := &importRepositoryStub{}
	service := newImportTestService(t, repo, &providerClientStub{})

	_, err := service.ImportFile(t.Context(), ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: uuid.New(),
		Visibility: "private",
	})

	require.ErrorIs(t, err, domain.ErrInvalidInput)
	require.Zero(t, repo.getReferenceCalls)
}

func TestImportFileReplaysCompletedOperationWithoutProviderIO(t *testing.T) {
	t.Parallel()

	documentID := uuid.New()
	repo := &importRepositoryStub{
		createImport: func(operation domain.ImportOperation) (domain.ImportOperation, bool, error) {
			operation.ID = uuid.New()
			operation.DocumentID = documentID
			operation.Status = domain.ImportOperationCompleted
			return operation, false, nil
		},
	}
	client := &providerClientStub{}
	service := newImportTestService(t, repo, client)
	referenceID := uuid.New()

	result, err := service.ImportFile(t.Context(), ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: referenceID,
		Visibility: "private", IdempotencyKey: "stable-import-key",
	})

	require.NoError(t, err)
	require.Equal(t, documentID, result.DocumentID)
	require.Equal(t, referenceID, result.SourceReferenceID)
	require.Zero(t, repo.getReferenceCalls)
	require.Zero(t, client.readFileCalls)
	require.Zero(t, repo.finalizeCalls)
}

func TestImportFileRejectsIdempotencyKeyReuseForDifferentRequest(t *testing.T) {
	t.Parallel()

	repo := &importRepositoryStub{
		createImport: func(operation domain.ImportOperation) (domain.ImportOperation, bool, error) {
			operation.ID = uuid.New()
			operation.RequestHash = strings.Repeat("f", 64)
			operation.Status = domain.ImportOperationCompleted
			return operation, false, nil
		},
	}
	service := newImportTestService(t, repo, &providerClientStub{})

	_, err := service.ImportFile(t.Context(), ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: uuid.New(),
		Visibility: "workspace", IdempotencyKey: "reused-key",
	})

	require.ErrorIs(t, err, domain.ErrConflict)
	require.Zero(t, repo.getReferenceCalls)
}

func TestImportFileKeepsConcurrentPendingAttemptSingleFlight(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	repo := &importRepositoryStub{
		createImport: func(operation domain.ImportOperation) (domain.ImportOperation, bool, error) {
			operation.ID = uuid.New()
			operation.Status = domain.ImportOperationPending
			operation.UpdatedAt = now.Add(-time.Minute)
			return operation, false, nil
		},
	}
	service := newImportTestService(t, repo, &providerClientStub{})
	service.now = func() time.Time { return now }

	_, err := service.ImportFile(t.Context(), ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: uuid.New(),
		Visibility: "workspace", IdempotencyKey: "pending-key",
	})

	require.ErrorIs(t, err, domain.ErrOperationInProgress)
	require.Zero(t, repo.claimCalls)
	require.Zero(t, repo.getReferenceCalls)
}

func TestPrepareImportOperationRetriesWithTheSameDocumentAndANewAttemptFence(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 9, 4, 8, 0, 0, 0, time.UTC)
	documentID := uuid.New()
	previousAttempt := uuid.New()
	repo := &importRepositoryStub{}
	repo.createImport = func(operation domain.ImportOperation) (domain.ImportOperation, bool, error) {
		operation.ID = uuid.New()
		operation.DocumentID = documentID
		operation.AttemptGeneration = previousAttempt
		operation.Status = domain.ImportOperationFailed
		operation.UpdatedAt = now.Add(-time.Minute)
		return operation, false, nil
	}
	repo.claimImport = func(_ uuid.UUID, attemptGeneration uuid.UUID, _, _ time.Time) (domain.ImportOperation, bool, error) {
		operation := repo.createdOperation
		operation.AttemptGeneration = attemptGeneration
		operation.Status = domain.ImportOperationPending
		operation.UpdatedAt = now
		return operation, true, nil
	}
	service := newImportTestService(t, repo, &providerClientStub{})
	service.now = func() time.Time { return now }
	input := ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: uuid.New(),
		Visibility: "private", IdempotencyKey: "retry-key",
	}

	operation, shouldRun, err := service.prepareImportOperation(t.Context(), input, importRequestHash(input))

	require.NoError(t, err)
	require.True(t, shouldRun)
	require.Equal(t, documentID, operation.DocumentID)
	require.NotEqual(t, previousAttempt, operation.AttemptGeneration)
	require.Equal(t, 1, repo.claimCalls)
}

func TestImportFileFinalizesDocumentAndProvenanceThroughRepository(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	referenceID := uuid.New()
	account := domain.Account{
		ID: uuid.New(), UserID: userID, GoogleSubject: "google-subject",
		CredentialVersion:      int16(credentialvault.CurrentVersion),
		InstallationGeneration: uuid.New(), ExpiresAt: time.Now().Add(time.Hour),
	}
	reference := domain.FileReference{
		ID: referenceID, FileID: "provider-file", MimeType: googleDocumentMimeType,
		TargetType: domain.TargetStory, TargetID: uuid.New(), Account: &account,
	}
	grantGeneration := uuid.New()
	repo := &importRepositoryStub{
		reference: reference, targetMutable: true, grantGeneration: grantGeneration,
		createImport: newPendingImportOperation,
	}
	client := &providerClientStub{
		getFileResult: domain.ProviderFile{
			ID: "provider-file", Name: " Launch <brief> ", MimeType: googleDocumentMimeType,
			WebViewLink: "https://docs.google.com/document/d/provider-file/edit",
			Version:     pointer("version-7"),
		},
		readFileResult: ProviderContent{Text: "First & foremost\n\nSecond <line>"},
	}
	service := newImportTestService(t, repo, client)
	sealed, err := service.sealToken(account, domain.OAuthToken{
		AccessToken: "access-token", RefreshToken: "refresh-token", Expiry: account.ExpiresAt,
	})
	require.NoError(t, err)
	account.CredentialPayload = sealed
	reference.Account = &account
	repo.reference = reference

	result, err := service.ImportFile(t.Context(), ImportInput{
		WorkspaceID: workspaceID, UserID: userID, ReferenceID: referenceID,
		Visibility: "private", IdempotencyKey: "successful-import",
	})

	require.NoError(t, err)
	require.Equal(t, repo.createdOperation.DocumentID, result.DocumentID)
	require.Equal(t, referenceID, result.SourceReferenceID)
	require.Equal(t, 1, repo.finalizeCalls)
	require.Equal(t, repo.createdOperation, repo.finalization.Operation)
	require.Equal(t, account.ID, repo.finalization.AccountID)
	require.Equal(t, grantGeneration, repo.finalization.GrantGeneration)
	require.Equal(t, "Launch <brief>", repo.finalization.Title)
	require.Equal(t, "<p>First &amp; foremost</p><p>Second &lt;line&gt;</p>", repo.finalization.ContentHTML)
	require.Equal(t, "First & foremost\n\nSecond <line>", repo.finalization.ContentText)
	require.Zero(t, repo.failCalls)
}

func TestImportFileRecordsFailureAfterRequestCancellation(t *testing.T) {
	t.Parallel()

	repo := &importRepositoryStub{createImport: newPendingImportOperation}
	service := newImportTestService(t, repo, &providerClientStub{})
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	_, err := service.ImportFile(ctx, ImportInput{
		WorkspaceID: uuid.New(), UserID: uuid.New(), ReferenceID: uuid.New(),
		Visibility: "workspace", IdempotencyKey: "cancelled-import",
	})

	require.ErrorIs(t, err, context.Canceled)
	require.Equal(t, 1, repo.failCalls)
	require.NoError(t, repo.failContextError)
	require.True(t, repo.failContextHasDeadline)
	require.Equal(t, repo.createdOperation.AttemptGeneration, repo.failedAttemptGeneration)
}

func newImportTestService(t *testing.T, repo Repository, client ProviderClient) *Service {
	t.Helper()
	vault, err := credentialvault.NewFromEncodedKeyring(
		credentialvault.DevelopmentKeyID,
		credentialvault.DevelopmentKeyVersion,
		credentialvault.DevelopmentEncodedKeys,
	)
	require.NoError(t, err)
	service := New(nil, repo, Config{
		ClientID: "client", ClientSecret: "secret", RedirectURL: "https://example.com/callback",
		PickerAPIKey: "picker", AppID: "123", Credentials: vault,
	})
	service.client = client
	return service
}

func newPendingImportOperation(operation domain.ImportOperation) (domain.ImportOperation, bool, error) {
	now := time.Now().UTC()
	operation.ID = uuid.New()
	operation.Status = domain.ImportOperationPending
	operation.CreatedAt = now
	operation.UpdatedAt = now
	return operation, true, nil
}

type importRepositoryStub struct {
	Repository

	createImport            func(domain.ImportOperation) (domain.ImportOperation, bool, error)
	claimImport             func(uuid.UUID, uuid.UUID, time.Time, time.Time) (domain.ImportOperation, bool, error)
	createdOperation        domain.ImportOperation
	reference               domain.FileReference
	targetMutable           bool
	grantGeneration         uuid.UUID
	finalization            domain.ImportFinalization
	failedAttemptGeneration uuid.UUID
	failContextError        error
	failContextHasDeadline  bool
	claimCalls              int
	getReferenceCalls       int
	finalizeCalls           int
	failCalls               int
}

func (repo *importRepositoryStub) CreateImportOperation(
	_ context.Context,
	operation domain.ImportOperation,
) (domain.ImportOperation, bool, error) {
	result, created, err := repo.createImport(operation)
	repo.createdOperation = result
	return result, created, err
}

func (repo *importRepositoryStub) ClaimImportOperation(
	_ context.Context,
	operationID uuid.UUID,
	attemptGeneration uuid.UUID,
	previousUpdatedAt time.Time,
	staleBefore time.Time,
) (domain.ImportOperation, bool, error) {
	repo.claimCalls++
	if repo.claimImport != nil {
		return repo.claimImport(operationID, attemptGeneration, previousUpdatedAt, staleBefore)
	}
	return domain.ImportOperation{}, false, nil
}

func (repo *importRepositoryStub) GetReference(
	ctx context.Context,
	_ uuid.UUID,
	_ uuid.UUID,
	_ uuid.UUID,
) (domain.FileReference, error) {
	repo.getReferenceCalls++
	if err := ctx.Err(); err != nil {
		return domain.FileReference{}, err
	}
	return repo.reference, nil
}

func (repo *importRepositoryStub) TargetMutable(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	domain.TargetType,
	uuid.UUID,
) (bool, error) {
	return repo.targetMutable, nil
}

func (repo *importRepositoryStub) RevalidateReference(
	context.Context,
	uuid.UUID,
	uuid.UUID,
	uuid.UUID,
	domain.FileReference,
	domain.ProviderFile,
) (uuid.UUID, error) {
	return repo.grantGeneration, nil
}

func (repo *importRepositoryStub) FinalizeDocumentImport(
	_ context.Context,
	input domain.ImportFinalization,
) (uuid.UUID, error) {
	repo.finalizeCalls++
	repo.finalization = input
	return input.Operation.DocumentID, nil
}

func (repo *importRepositoryStub) FailImportOperation(
	ctx context.Context,
	_ uuid.UUID,
	attemptGeneration uuid.UUID,
	_ string,
) error {
	repo.failCalls++
	repo.failContextError = ctx.Err()
	_, repo.failContextHasDeadline = ctx.Deadline()
	repo.failedAttemptGeneration = attemptGeneration
	return nil
}
