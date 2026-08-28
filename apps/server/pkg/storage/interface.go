package storage

import (
	"context"
	"io"
	"time"

	"github.com/complexus-tech/projects-api/pkg/safeio"
)

var (
	ErrInvalidDownloadLimit = safeio.ErrInvalidLimit
	ErrDownloadTooLarge     = safeio.ErrLimitExceeded
)

// StorageService defines the interface for storage operations.
type StorageService interface {
	// UploadFile uploads a file to the specified container/bucket.
	UploadFile(ctx context.Context, container, filename string, data io.Reader, contentType string) (string, error)

	// DownloadFile downloads at most maxBytes and returns the bytes and content
	// type. Callers own the resource-specific limit; adapters must enforce it
	// even when provider metadata claims a smaller object.
	DownloadFile(ctx context.Context, container, filename string, maxBytes int64) ([]byte, string, error)

	// GenerateAccessURL generates a temporary access URL for a file.
	GenerateAccessURL(ctx context.Context, container, filename string, expiry time.Duration) (string, error)

	// DeleteFile removes a file from storage.
	DeleteFile(ctx context.Context, container, filename string) error

	// GetPublicURL returns a permanent public URL for a file.
	GetPublicURL(ctx context.Context, container, filename string) (string, error)
}
