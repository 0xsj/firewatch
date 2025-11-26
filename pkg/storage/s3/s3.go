package s3

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"

	"github.com/0xsj/hexagonal-go/pkg/storage"
)

// Storage implements the storage.Storage interface using S3/MinIO
type Storage struct {
	client        *s3.Client
	presignClient *s3.PresignClient
	bucket        string
	config        storage.Config
}

// New creates a new S3 storage instance
func New(ctx context.Context, cfg storage.Config) (*Storage, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOpts := func(o *s3.Options) {
		if cfg.Endpoint != "" {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
		}
		o.UsePathStyle = cfg.UsePathStyle
	}

	client := s3.NewFromConfig(awsCfg, clientOpts)
	presignClient := s3.NewPresignClient(client)

	return &Storage{
		client:        client,
		presignClient: presignClient,
		bucket:        cfg.Bucket,
		config:        cfg,
	}, nil
}

// Upload stores a file and returns the result
func (s *Storage) Upload(ctx context.Context, input storage.UploadInput) (*storage.UploadOutput, error) {
	if input.Key == "" {
		return nil, storage.ErrInvalidKey
	}

	if s.config.MaxFileSize > 0 && input.Size > s.config.MaxFileSize {
		return nil, storage.ErrFileTooLarge
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(input.Key),
		Body:        input.Body,
		ContentType: aws.String(input.ContentType),
	}

	if input.Size > 0 {
		putInput.ContentLength = aws.Int64(input.Size)
	}

	if len(input.Metadata) > 0 {
		putInput.Metadata = input.Metadata
	}

	result, err := s.client.PutObject(ctx, putInput)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrUploadFailed, err)
	}

	output := &storage.UploadOutput{
		Key: input.Key,
	}

	if result.ETag != nil {
		output.ETag = *result.ETag
	}
	if result.VersionId != nil {
		output.VersionID = *result.VersionId
	}

	return output, nil
}

// Download retrieves a file by key
func (s *Storage) Download(ctx context.Context, key string) (*storage.DownloadOutput, error) {
	if key == "" {
		return nil, storage.ErrInvalidKey
	}

	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return nil, storage.ErrNotFound
		}
		if isAccessDeniedError(err) {
			return nil, storage.ErrAccessDenied
		}
		return nil, fmt.Errorf("%w: %v", storage.ErrDownloadFailed, err)
	}

	output := &storage.DownloadOutput{
		Body: result.Body,
	}

	if result.ContentType != nil {
		output.ContentType = *result.ContentType
	}
	if result.ContentLength != nil {
		output.Size = *result.ContentLength
	}
	if result.ETag != nil {
		output.ETag = *result.ETag
	}

	return output, nil
}

// Delete removes a file by key
func (s *Storage) Delete(ctx context.Context, key string) error {
	if key == "" {
		return storage.ErrInvalidKey
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return storage.ErrNotFound
		}
		if isAccessDeniedError(err) {
			return storage.ErrAccessDenied
		}
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// Exists checks if a file exists
func (s *Storage) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, storage.ErrInvalidKey
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isNotFoundError(err) {
			return false, nil
		}
		if isAccessDeniedError(err) {
			return false, storage.ErrAccessDenied
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}

	return true, nil
}

// GenerateUploadURL creates a presigned URL for direct client upload
func (s *Storage) GenerateUploadURL(ctx context.Context, input storage.PresignedUploadInput) (*storage.PresignedURL, error) {
	if input.Key == "" {
		return nil, storage.ErrInvalidKey
	}

	expiry := input.Expiry
	if expiry == 0 {
		expiry = s.config.DefaultPresignExpiry
	}

	putInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(input.Key),
		ContentType: aws.String(input.ContentType),
	}

	result, err := s.presignClient.PresignPutObject(ctx, putInput, s3.WithPresignExpires(expiry))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrPresignFailed, err)
	}

	return &storage.PresignedURL{
		URL:       result.URL,
		Method:    result.Method,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// GenerateDownloadURL creates a presigned URL for direct client download
func (s *Storage) GenerateDownloadURL(ctx context.Context, key string, expiry time.Duration) (*storage.PresignedURL, error) {
	if key == "" {
		return nil, storage.ErrInvalidKey
	}

	if expiry == 0 {
		expiry = s.config.DefaultPresignExpiry
	}

	result, err := s.presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", storage.ErrPresignFailed, err)
	}

	return &storage.PresignedURL{
		URL:       result.URL,
		Method:    result.Method,
		ExpiresAt: time.Now().Add(expiry),
	}, nil
}

// isNotFoundError checks if the error is a not found error
func isNotFoundError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		code := apiErr.ErrorCode()
		return code == "NotFound" || code == "NoSuchKey"
	}
	return false
}

// isAccessDeniedError checks if the error is an access denied error
func isAccessDeniedError(err error) bool {
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		return apiErr.ErrorCode() == "AccessDenied"
	}
	return false
}
