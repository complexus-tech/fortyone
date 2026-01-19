package aws

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	sdkaws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/complexus-tech/projects-api/pkg/logger"
)

// S3Service provides operations with AWS S3.
type S3Service struct {
	client *s3.Client
	config Config
	log    *logger.Logger
}

// NewS3Service creates a new AWS S3 service.
func NewS3Service(cfg Config, log *logger.Logger) (*S3Service, error) {
	if cfg.AccessKeyID == "" || cfg.SecretAccessKey == "" || cfg.Region == "" {
		return nil, fmt.Errorf("aws access key, secret key, and region are required")
	}

	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(awsCfg, func(options *s3.Options) {
		if cfg.Endpoint != "" {
			options.BaseEndpoint = sdkaws.String(cfg.Endpoint)
			options.UsePathStyle = cfg.ForcePathStyle
		}
	})

	return &S3Service{
		client: client,
		config: cfg,
		log:    log,
	}, nil
}

// UploadFile implements storage.StorageService.
func (s *S3Service) UploadFile(ctx context.Context, bucket, key string, data io.Reader, contentType string) (string, error) {
	input := &s3.PutObjectInput{
		Bucket:      sdkaws.String(bucket),
		Key:         sdkaws.String(key),
		Body:        data,
		ContentType: sdkaws.String(contentType),
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to upload object to S3: %w", err)
	}

	return s.GetPublicURL(ctx, bucket, key)
}

// GenerateAccessURL implements storage.StorageService.
func (s *S3Service) GenerateAccessURL(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: sdkaws.String(bucket),
		Key:    sdkaws.String(key),
	}, func(options *s3.PresignOptions) {
		options.Expires = expiry
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

// DeleteFile implements storage.StorageService.
func (s *S3Service) DeleteFile(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: sdkaws.String(bucket),
		Key:    sdkaws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete object from S3: %w", err)
	}

	return nil
}

// GetPublicURL implements storage.StorageService.
func (s *S3Service) GetPublicURL(ctx context.Context, bucket, key string) (string, error) {
	if s.config.PublicURL != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.config.PublicURL, "/"), bucket, key), nil
	}

	if s.config.Endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", strings.TrimRight(s.config.Endpoint, "/"), bucket, key), nil
	}

	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", bucket, s.config.Region, key), nil
}
