package googledriverepository

import (
	"context"
	"testing"

	"github.com/complexus-tech/projects-api/internal/modules/googledrive/domain"
	googledrivesql "github.com/complexus-tech/projects-api/internal/modules/googledrive/repository/sqlc"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestRevalidateReferenceReturnsThePersistedGrantGeneration(t *testing.T) {
	t.Parallel()

	workspaceID := uuid.New()
	userID := uuid.New()
	accountID := uuid.New()
	fileID := uuid.New()
	queries := &fileRevalidationQueries{fileID: fileID, targetAccessible: true, grantRows: 1}
	repository := fileRevalidationRepository(queries)

	grantGeneration, err := repository.RevalidateReference(
		t.Context(), workspaceID, userID, accountID,
		domain.FileReference{
			ID:         uuid.New(),
			TargetType: domain.TargetStory,
			TargetID:   uuid.New(),
		},
		domain.ProviderFile{ID: "provider-file", Name: "Launch plan", MimeType: "application/vnd.google-apps.document"},
	)

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, grantGeneration)
	require.Equal(t, grantGeneration, queries.grantGeneration)
	require.Equal(t, fileID, queries.grantFileID)
	require.Equal(t, accountID, queries.grantAccountID)
}

func TestRevalidateReferenceDoesNotReturnAnUnpersistedGrantGeneration(t *testing.T) {
	t.Parallel()

	queries := &fileRevalidationQueries{
		fileID:           uuid.New(),
		targetAccessible: true,
		grantRows:        0,
	}
	repository := fileRevalidationRepository(queries)

	grantGeneration, err := repository.RevalidateReference(
		t.Context(), uuid.New(), uuid.New(), uuid.New(),
		domain.FileReference{
			ID:         uuid.New(),
			TargetType: domain.TargetStory,
			TargetID:   uuid.New(),
		},
		domain.ProviderFile{ID: "provider-file", Name: "Launch plan", MimeType: "application/vnd.google-apps.document"},
	)

	require.ErrorIs(t, err, domain.ErrForbidden)
	require.Equal(t, uuid.Nil, grantGeneration)
}

type fileRevalidationQueries struct {
	googledrivesql.Querier

	fileID           uuid.UUID
	targetAccessible bool
	grantRows        int64
	grantGeneration  uuid.UUID
	grantFileID      uuid.UUID
	grantAccountID   uuid.UUID
}

func (queries *fileRevalidationQueries) GoogleDriveTargetAccessible(
	context.Context,
	googledrivesql.GoogleDriveTargetAccessibleParams,
) (bool, error) {
	return queries.targetAccessible, nil
}

func (queries *fileRevalidationQueries) RevalidateGoogleDriveFileReference(
	context.Context,
	googledrivesql.RevalidateGoogleDriveFileReferenceParams,
) (uuid.UUID, error) {
	return queries.fileID, nil
}

func (queries *fileRevalidationQueries) UpsertGoogleDriveFileGrant(
	_ context.Context,
	params googledrivesql.UpsertGoogleDriveFileGrantParams,
) (int64, error) {
	queries.grantGeneration = params.GrantGeneration
	queries.grantFileID = params.FileID
	queries.grantAccountID = params.AccountID
	return queries.grantRows, nil
}

func fileRevalidationRepository(queries googledrivesql.Querier) *Repository {
	return &Repository{
		queries: queries,
		runTransaction: func(ctx context.Context, operation func(googledrivesql.Querier) error) error {
			return operation(queries)
		},
	}
}
