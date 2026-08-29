package attachments

import (
	"context"
	"strings"
)

const (
	maximumRetainedObjectProviderLength  = 32
	maximumRetainedObjectContainerLength = 255
	maximumRetainedObjectBlobNameLength  = 1024
)

// RetainedObjectStorage returns the credential-free storage route captured by
// the story-retention transaction. Credentials remain exclusively in runtime
// configuration and are never persisted in the deletion outbox.
func (s *Service) RetainedObjectStorage() (string, string, error) {
	if s == nil || s.storage == nil {
		return "", "", ErrRetainedObjectStorageRoute
	}
	provider := strings.TrimSpace(s.config.Provider)
	container := strings.TrimSpace(s.config.AttachmentsBucket)
	if provider == "" || len(provider) > maximumRetainedObjectProviderLength ||
		container == "" || len(container) > maximumRetainedObjectContainerLength {
		return "", "", ErrRetainedObjectStorageRoute
	}
	return provider, container, nil
}

// DeleteRetainedObject performs an idempotent object-store deletion using only
// a route captured from this service. It deliberately suppresses provider
// errors because they may contain object names or signed request details.
func (s *Service) DeleteRetainedObject(
	ctx context.Context,
	provider string,
	container string,
	blobName string,
) error {
	if ctx == nil || s == nil || s.storage == nil {
		return ErrRetainedObjectStorageRoute
	}
	configuredProvider, configuredContainer, err := s.RetainedObjectStorage()
	if err != nil || strings.TrimSpace(provider) != configuredProvider ||
		strings.TrimSpace(container) != configuredContainer ||
		strings.TrimSpace(blobName) == "" || len(blobName) > maximumRetainedObjectBlobNameLength {
		return ErrRetainedObjectStorageRoute
	}
	if err := s.storage.DeleteFile(ctx, configuredContainer, blobName); err != nil {
		return ErrRetainedObjectDeletion
	}
	return nil
}
