package googledriverepository

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestFinalizeDocumentImportCommitsOneAtomicUnit(t *testing.T) {
	t.Parallel()

	input := newValidImportFinalization()
	queries := &importFinalizationQueries{
		operation:  operationRow(input.Operation),
		importable: true, targetMutable: true,
		createdDocumentID: input.Operation.DocumentID,
		provenanceRows:    1, completionRows: 1,
	}
	repository := importFinalizationRepository(queries)

	documentID, err := repository.FinalizeDocumentImport(t.Context(), input)

	require.NoError(t, err)
	require.Equal(t, input.Operation.DocumentID, documentID)
	require.Equal(t, 1, queries.transactionCalls)
	require.Equal(t, []string{"lock", "reference", "target", "document", "provenance", "complete"}, queries.calls)
	require.Equal(t, input.Operation.DocumentID, queries.provenance.DocumentID)
	require.Equal(t, input.Operation.SourceReferenceID, *queries.provenance.ReferenceID)
	require.Equal(t, input.Operation.AttemptGeneration, queries.completion.AttemptGeneration)
}

func TestFinalizeDocumentImportDoesNotCompleteWhenProvenanceFails(t *testing.T) {
	t.Parallel()

	input := newValidImportFinalization()
	provenanceErr := errors.New("provenance unavailable")
	queries := &importFinalizationQueries{
		operation:  operationRow(input.Operation),
		importable: true, targetMutable: true,
		createdDocumentID: input.Operation.DocumentID,
		provenanceErr:     provenanceErr,
	}
	repository := importFinalizationRepository(queries)

	documentID, err := repository.FinalizeDocumentImport(t.Context(), input)

	require.ErrorIs(t, err, provenanceErr)
	require.Equal(t, uuid.Nil, documentID)
	require.Equal(t, 1, queries.transactionCalls)
	require.Equal(t, []string{"lock", "reference", "target", "document", "provenance"}, queries.calls)
	require.Zero(t, queries.completionCalls)
}

func TestFinalizeDocumentImportRejectsSupersededAttemptBeforeCreatingDocument(t *testing.T) {
	t.Parallel()

	input := newValidImportFinalization()
	stored := operationRow(input.Operation)
	stored.AttemptGeneration = uuid.New()
	queries := &importFinalizationQueries{operation: stored}
	repository := importFinalizationRepository(queries)

	_, err := repository.FinalizeDocumentImport(t.Context(), input)

	require.ErrorIs(t, err, domain.ErrOperationInProgress)
	require.Equal(t, []string{"lock"}, queries.calls)
	require.Zero(t, queries.documentCalls)
}

func TestFinalizeDocumentImportReplaysAlreadyCompletedIdentity(t *testing.T) {
	t.Parallel()

	input := newValidImportFinalization()
	stored := operationRow(input.Operation)
	stored.Status = domain.ImportOperationCompleted
	stored.AttemptGeneration = uuid.New()
	completedAt := time.Now().UTC()
	stored.CompletedAt = &completedAt
	queries := &importFinalizationQueries{operation: stored}
	repository := importFinalizationRepository(queries)

	documentID, err := repository.FinalizeDocumentImport(t.Context(), input)

	require.NoError(t, err)
	require.Equal(t, input.Operation.DocumentID, documentID)
	require.Equal(t, []string{"lock"}, queries.calls)
	require.Zero(t, queries.documentCalls)
}

func newValidImportFinalization() domain.ImportFinalization {
	now := time.Now().UTC()
	operation := domain.ImportOperation{
		ID: uuid.New(), WorkspaceID: uuid.New(), UserID: uuid.New(),
		SourceReferenceID: uuid.New(), DocumentID: uuid.New(),
		IdempotencyKey: "import-operation", RequestHash: strings.Repeat("a", 64),
		Visibility: "private", AttemptGeneration: uuid.New(),
		Status: domain.ImportOperationPending, CreatedAt: now, UpdatedAt: now,
	}
	return domain.ImportFinalization{
		Operation: operation, AccountID: uuid.New(), GrantGeneration: uuid.New(),
		TargetType: domain.TargetStory, TargetID: uuid.New(),
		GoogleFileID: "provider-file", SourceVersion: stringPointer("version-3"),
		Title: "Launch brief", ContentHTML: "<p>Launch brief</p>", ContentText: "Launch brief",
	}
}

func operationRow(operation domain.ImportOperation) googledrivesql.LockGoogleDriveImportOperationRow {
	return googledrivesql.LockGoogleDriveImportOperationRow{
		OperationID: operation.ID, WorkspaceID: operation.WorkspaceID, UserID: operation.UserID,
		SourceReferenceID: operation.SourceReferenceID, DocumentID: operation.DocumentID,
		IdempotencyKey: operation.IdempotencyKey, RequestHash: operation.RequestHash,
		Visibility: operation.Visibility, AttemptGeneration: operation.AttemptGeneration,
		Status: operation.Status, CreatedAt: operation.CreatedAt, UpdatedAt: operation.UpdatedAt,
		CompletedAt: operation.CompletedAt,
	}
}

func stringPointer(value string) *string { return &value }

type importFinalizationQueries struct {
	googledrivesql.Querier

	operation         googledrivesql.LockGoogleDriveImportOperationRow
	importable        bool
	targetMutable     bool
	createdDocumentID uuid.UUID
	provenanceRows    int64
	provenanceErr     error
	completionRows    int64
	provenance        googledrivesql.SaveGoogleDriveDocumentImportParams
	completion        googledrivesql.CompleteGoogleDriveImportOperationParams
	calls             []string
	transactionCalls  int
	documentCalls     int
	completionCalls   int
}

func (queries *importFinalizationQueries) LockGoogleDriveImportOperation(
	context.Context,
	googledrivesql.LockGoogleDriveImportOperationParams,
) (googledrivesql.LockGoogleDriveImportOperationRow, error) {
	queries.calls = append(queries.calls, "lock")
	return queries.operation, nil
}

func (queries *importFinalizationQueries) GoogleDriveReferenceImportable(
	context.Context,
	googledrivesql.GoogleDriveReferenceImportableParams,
) (bool, error) {
	queries.calls = append(queries.calls, "reference")
	return queries.importable, nil
}

func (queries *importFinalizationQueries) GoogleDriveTargetMutable(
	context.Context,
	googledrivesql.GoogleDriveTargetMutableParams,
) (bool, error) {
	queries.calls = append(queries.calls, "target")
	return queries.targetMutable, nil
}

func (queries *importFinalizationQueries) CreateGoogleDriveImportedDocument(
	context.Context,
	googledrivesql.CreateGoogleDriveImportedDocumentParams,
) (uuid.UUID, error) {
	queries.calls = append(queries.calls, "document")
	queries.documentCalls++
	return queries.createdDocumentID, nil
}

func (queries *importFinalizationQueries) SaveGoogleDriveDocumentImport(
	_ context.Context,
	params googledrivesql.SaveGoogleDriveDocumentImportParams,
) (int64, error) {
	queries.calls = append(queries.calls, "provenance")
	queries.provenance = params
	return queries.provenanceRows, queries.provenanceErr
}

func (queries *importFinalizationQueries) CompleteGoogleDriveImportOperation(
	_ context.Context,
	params googledrivesql.CompleteGoogleDriveImportOperationParams,
) (int64, error) {
	queries.calls = append(queries.calls, "complete")
	queries.completion = params
	queries.completionCalls++
	return queries.completionRows, nil
}

func importFinalizationRepository(queries *importFinalizationQueries) *Repository {
	return &Repository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(googledrivesql.Querier) error) error {
			queries.transactionCalls++
			return operation(queries)
		},
	}
}
