package attachments

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/stretchr/testify/require"
)

func TestRetainedObjectDeletionUsesConfiguredCredentialFreeRoute(t *testing.T) {
	storageStub := &retainedObjectStorageStub{}
	service := New(nil, nil, storageStub, storage.Config{
		Provider:          "aws",
		AttachmentsBucket: "attachments",
	}, nil)

	provider, container, err := service.RetainedObjectStorage()
	require.NoError(t, err)
	require.Equal(t, "aws", provider)
	require.Equal(t, "attachments", container)
	require.NoError(t, service.DeleteRetainedObject(
		context.Background(), provider, container, "private/object-name.png",
	))
	require.Equal(t, 1, storageStub.deleteCalls)
	require.Equal(t, "attachments", storageStub.deletedContainer)
	require.Equal(t, "private/object-name.png", storageStub.deletedBlobName)
}

func TestRetainedObjectDeletionRejectsAStaleOrForeignStorageRoute(t *testing.T) {
	storageStub := &retainedObjectStorageStub{}
	service := New(nil, nil, storageStub, storage.Config{
		Provider:          "aws",
		AttachmentsBucket: "attachments",
	}, nil)

	err := service.DeleteRetainedObject(
		context.Background(), "azure", "attachments", "private/object-name.png",
	)
	require.ErrorIs(t, err, ErrRetainedObjectStorageRoute)
	require.Zero(t, storageStub.deleteCalls)
}

func TestRetainedObjectDeletionSuppressesSensitiveProviderErrors(t *testing.T) {
	const blobName = "private/customer-object-name.png"
	storageStub := &retainedObjectStorageStub{
		deleteErr: errors.New("provider failed for " + blobName + "?signed-secret=credential"),
	}
	service := New(nil, nil, storageStub, storage.Config{
		Provider:          "azure",
		AttachmentsBucket: "attachments",
	}, nil)

	err := service.DeleteRetainedObject(context.Background(), "azure", "attachments", blobName)
	require.ErrorIs(t, err, ErrRetainedObjectDeletion)
	require.NotContains(t, err.Error(), blobName)
	require.NotContains(t, strings.ToLower(err.Error()), "credential")
}

type retainedObjectStorageStub struct {
	deleteCalls      int
	deletedContainer string
	deletedBlobName  string
	deleteErr        error
}

func (*retainedObjectStorageStub) UploadFile(context.Context, string, string, io.Reader, string) (string, error) {
	return "", errors.New("unexpected upload")
}

func (*retainedObjectStorageStub) DownloadFile(context.Context, string, string, int64) ([]byte, string, error) {
	return nil, "", errors.New("unexpected download")
}

func (*retainedObjectStorageStub) GenerateAccessURL(context.Context, string, string, time.Duration) (string, error) {
	return "", errors.New("unexpected access URL")
}

func (s *retainedObjectStorageStub) DeleteFile(_ context.Context, container, blobName string) error {
	s.deleteCalls++
	s.deletedContainer = container
	s.deletedBlobName = blobName
	return s.deleteErr
}

func (*retainedObjectStorageStub) GetPublicURL(context.Context, string, string) (string, error) {
	return "", errors.New("unexpected public URL")
}
