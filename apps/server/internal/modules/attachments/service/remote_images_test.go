package attachments

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/complexus-tech/projects-api/internal/platform/safehttp"
	"github.com/complexus-tech/projects-api/pkg/logger"
	"github.com/complexus-tech/projects-api/pkg/storage"
	"github.com/complexus-tech/projects-api/pkg/validate"
	"github.com/google/uuid"
)

func TestUploadProfileImageFromURLUsesBoundedSafeDownloader(t *testing.T) {
	downloader := &remoteImageDownloaderStub{download: safehttp.Download{
		Body: []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n', 0, 0, 0, 0},
	}}
	storageStub := &attachmentStorageStub{}
	service := New(
		logger.NewWithText(io.Discard, slog.LevelError, "remote-image-test"),
		&attachmentRepositoryStub{},
		storageStub,
		storage.Config{ProfilesBucket: "profiles"},
		nil,
		WithRemoteImageDownloader(downloader),
	)

	imageURL := "https://images.example.com/avatar.png"
	blobName, err := service.UploadProfileImageFromURL(t.Context(), imageURL, uuid.New())
	if err != nil {
		t.Fatalf("UploadProfileImageFromURL() error = %v", err)
	}
	if downloader.requestedURL != imageURL {
		t.Fatalf("downloaded URL = %q", downloader.requestedURL)
	}
	if blobName == "" || storageStub.uploadCount != 1 || storageStub.contentType != "image/png" {
		t.Fatalf("uploaded image = blob %q, count %d, type %q", blobName, storageStub.uploadCount, storageStub.contentType)
	}
}

func TestUploadProfileImageFromURLDoesNotStoreRejectedEgress(t *testing.T) {
	for name, downloadError := range map[string]error{
		"unsafe target": safehttp.ErrUnsafeAddress,
		"oversized":     safehttp.ErrResponseTooLarge,
	} {
		t.Run(name, func(t *testing.T) {
			downloader := &remoteImageDownloaderStub{err: downloadError}
			storageStub := &attachmentStorageStub{}
			service := New(
				logger.NewWithText(io.Discard, slog.LevelError, "remote-image-test"),
				&attachmentRepositoryStub{},
				storageStub,
				storage.Config{ProfilesBucket: "profiles"},
				nil,
				WithRemoteImageDownloader(downloader),
			)

			_, err := service.UploadProfileImageFromURL(t.Context(), "https://images.example.com/avatar.png", uuid.New())
			if downloadError == safehttp.ErrResponseTooLarge {
				if !errors.Is(err, validate.ErrFileTooLarge) {
					t.Fatalf("error = %v, want file-too-large", err)
				}
			} else if !errors.Is(err, downloadError) {
				t.Fatalf("error = %v, want %v", err, downloadError)
			}
			if storageStub.uploadCount != 0 {
				t.Fatalf("rejected remote image uploaded %d times", storageStub.uploadCount)
			}
		})
	}
}

type remoteImageDownloaderStub struct {
	download     safehttp.Download
	err          error
	requestedURL string
}

func (downloader *remoteImageDownloaderStub) Download(_ context.Context, rawURL string) (safehttp.Download, error) {
	downloader.requestedURL = rawURL
	return downloader.download, downloader.err
}
