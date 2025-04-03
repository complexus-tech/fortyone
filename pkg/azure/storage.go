package azure

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

// BlobService provides operations with Azure Blob Storage
type BlobService struct {
	config Config
	client *azblob.Client
	log    *logger.Logger
}

// New creates a new Azure blob service
func New(config Config, log *logger.Logger) (*BlobService, error) {
	client, err := azblob.NewClientFromConnectionString(config.ConnectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &BlobService{
		config: config,
		client: client,
		log:    log,
	}, nil
}

// UploadBlob uploads a file to Azure Blob Storage
func (s *BlobService) UploadBlob(ctx context.Context, containerName string, blobName string, data io.Reader, contentType string) (string, error) {
	// Ensure container exists
	_, err := s.client.CreateContainer(ctx, containerName, nil)
	if err != nil {
		// Ignore error if container already exists
		var respErr *azcore.ResponseError
		if !(errors.As(err, &respErr) && respErr.ErrorCode == "ContainerAlreadyExists") {
			return "", fmt.Errorf("failed to create container: %w", err)
		}
	}

	// Set blob options with content type
	options := &azblob.UploadStreamOptions{
		BlockSize: 4 * 1024 * 1024, // 4MB block size
	}

	if contentType != "" {
		options.HTTPHeaders = &blob.HTTPHeaders{
			BlobContentType: &contentType,
		}
	}

	// Upload the blob
	_, err = s.client.UploadStream(ctx, containerName, blobName, data, options)
	if err != nil {
		return "", fmt.Errorf("failed to upload blob: %w", err)
	}

	// Construct the blob URL
	blobURL := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.config.StorageAccountName,
		containerName,
		blobName)

	return blobURL, nil
}

// GenerateSASToken generates a SAS token for temporary access to a blob
func (s *BlobService) GenerateSASToken(ctx context.Context, containerName string, blobName string, expiry time.Duration) (string, error) {
	// Create a BlobSASPermissions object with desired permissions and expiry time
	permissions := sas.BlobPermissions{
		Read: true,
	}

	// Convert permissions to string
	permStr := permissions.String()

	// Get account key credential
	credential, err := azblob.NewSharedKeyCredential(s.config.StorageAccountName, s.config.AccountKey)
	if err != nil {
		return "", fmt.Errorf("failed to create shared key credential: %w", err)
	}

	// Setup SAS token generation parameters
	sasValues := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().UTC().Add(-1 * time.Minute),
		ExpiryTime:    time.Now().UTC().Add(expiry),
		Permissions:   permStr,
		ContainerName: containerName,
		BlobName:      blobName,
	}

	// Generate SAS token
	sasToken, err := sasValues.SignWithSharedKey(credential)
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS token: %w", err)
	}

	return sasToken.Encode(), nil
}

// DeleteBlob deletes a blob from Azure Blob Storage
func (s *BlobService) DeleteBlob(ctx context.Context, containerName string, blobName string) error {
	_, err := s.client.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}
