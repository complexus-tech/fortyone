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

// AzureStorageService provides operations with Azure Blob Storage.
type AzureStorageService struct {
	config Config
	client *azblob.Client
	log    *logger.Logger
}

// NewStorageService creates a new Azure storage service.
func NewStorageService(config Config, log *logger.Logger) (*AzureStorageService, error) {
	client, err := azblob.NewClientFromConnectionString(config.ConnectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure client: %w", err)
	}

	return &AzureStorageService{
		config: config,
		client: client,
		log:    log,
	}, nil
}

// UploadFile implements storage.StorageService.
func (s *AzureStorageService) UploadFile(ctx context.Context, containerName string, blobName string, data io.Reader, contentType string) (string, error) {
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

	return s.GetPublicURL(ctx, containerName, blobName)
}

// GenerateAccessURL implements storage.StorageService.
func (s *AzureStorageService) GenerateAccessURL(ctx context.Context, containerName string, blobName string, expiry time.Duration) (string, error) {
	sasToken, err := s.generateSASToken(containerName, blobName, expiry)
	if err != nil {
		return "", err
	}

	baseURL, err := s.GetPublicURL(ctx, containerName, blobName)
	if err != nil {
		return "", err
	}

	return baseURL + "?" + sasToken, nil
}

// DeleteFile implements storage.StorageService.
func (s *AzureStorageService) DeleteFile(ctx context.Context, containerName string, blobName string) error {
	_, err := s.client.DeleteBlob(ctx, containerName, blobName, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob: %w", err)
	}

	return nil
}

// GetPublicURL implements storage.StorageService.
func (s *AzureStorageService) GetPublicURL(ctx context.Context, containerName string, blobName string) (string, error) {
	return fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s",
		s.config.StorageAccountName,
		containerName,
		blobName), nil
}

func (s *AzureStorageService) generateSASToken(containerName string, blobName string, expiry time.Duration) (string, error) {
	permissions := sas.BlobPermissions{
		Read: true,
	}

	credential, err := azblob.NewSharedKeyCredential(s.config.StorageAccountName, s.config.AccountKey)
	if err != nil {
		return "", fmt.Errorf("failed to create shared key credential: %w", err)
	}

	sasValues := sas.BlobSignatureValues{
		Protocol:      sas.ProtocolHTTPS,
		StartTime:     time.Now().UTC().Add(-1 * time.Minute),
		ExpiryTime:    time.Now().UTC().Add(expiry),
		Permissions:   permissions.String(),
		ContainerName: containerName,
		BlobName:      blobName,
	}

	sasToken, err := sasValues.SignWithSharedKey(credential)
	if err != nil {
		return "", fmt.Errorf("failed to generate SAS token: %w", err)
	}

	return sasToken.Encode(), nil
}
