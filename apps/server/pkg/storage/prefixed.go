package storage

import (
	"context"
	"io"
	"strings"
	"time"
)

type prefixedStorageService struct {
	base   StorageService
	bucket string
}

func newPrefixedStorageService(base StorageService, bucket string) StorageService {
	return &prefixedStorageService{
		base:   base,
		bucket: bucket,
	}
}

func (s *prefixedStorageService) UploadFile(ctx context.Context, container, filename string, data io.Reader, contentType string) (string, error) {
	key := s.prefixedKey(container, filename)
	return s.base.UploadFile(ctx, s.bucket, key, data, contentType)
}

func (s *prefixedStorageService) GenerateAccessURL(ctx context.Context, container, filename string, expiry time.Duration) (string, error) {
	key := s.prefixedKey(container, filename)
	return s.base.GenerateAccessURL(ctx, s.bucket, key, expiry)
}

func (s *prefixedStorageService) DeleteFile(ctx context.Context, container, filename string) error {
	key := s.prefixedKey(container, filename)
	return s.base.DeleteFile(ctx, s.bucket, key)
}

func (s *prefixedStorageService) GetPublicURL(ctx context.Context, container, filename string) (string, error) {
	key := s.prefixedKey(container, filename)
	return s.base.GetPublicURL(ctx, s.bucket, key)
}

func (s *prefixedStorageService) prefixedKey(prefix, key string) string {
	cleanPrefix := strings.Trim(prefix, "/")
	cleanBucket := strings.Trim(s.bucket, "/")
	cleanKey := strings.TrimPrefix(key, "/")

	if cleanPrefix == "" || cleanPrefix == cleanBucket {
		return cleanKey
	}
	if cleanKey == "" {
		return cleanPrefix
	}
	if strings.HasPrefix(cleanKey, cleanPrefix+"/") || cleanKey == cleanPrefix {
		return cleanKey
	}

	return cleanPrefix + "/" + cleanKey
}
