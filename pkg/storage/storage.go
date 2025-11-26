package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the port for file storage operations.
// NOTE: For tenant/user isolation, implement key prefixing at the application layer
// using a pattern like: {tenant-id}/{user-id}/{filename}
type Storage interface {
	// Upload stores a file and returns the result
	Upload(ctx context.Context, input UploadInput) (*UploadOutput, error)

	// Download retrieves a file by key
	Download(ctx context.Context, key string) (*DownloadOutput, error)

	// Delete removes a file by key
	Delete(ctx context.Context, key string) error

	// Exists checks if a file exists
	Exists(ctx context.Context, key string) (bool, error)

	// GenerateUploadURL creates a presigned URL for direct client upload
	GenerateUploadURL(ctx context.Context, input PresignedUploadInput) (*PresignedURL, error)

	// GenerateDownloadURL creates a presigned URL for direct client download
	GenerateDownloadURL(ctx context.Context, key string, expiry time.Duration) (*PresignedURL, error)
}

// UploadInput contains the parameters for uploading a file
type UploadInput struct {
	// Key is the unique identifier/path for the file
	Key string

	// Body is the file content
	Body io.Reader

	// ContentType is the MIME type of the file
	ContentType string

	// Size is the file size in bytes (optional, improves upload efficiency)
	Size int64

	// Metadata contains optional key-value pairs stored with the file
	Metadata map[string]string
}

// UploadOutput contains the result of an upload operation
type UploadOutput struct {
	Key       string
	ETag      string
	VersionID string
}

// DownloadOutput contains the result of a download operation
type DownloadOutput struct {
	// Body is the file content - caller must close
	Body io.ReadCloser

	// ContentType is the MIME type of the file
	ContentType string

	// Size is the file size in bytes
	Size int64

	// ETag is the entity tag for cache validation
	ETag string
}

// PresignedUploadInput contains parameters for generating an upload URL
type PresignedUploadInput struct {
	// Key is the unique identifier/path for the file
	Key string

	// ContentType is the required MIME type (client must match)
	ContentType string

	// MaxSize is the maximum allowed file size in bytes (optional)
	MaxSize int64

	// Expiry is how long the URL remains valid
	Expiry time.Duration
}

// PresignedURL contains a presigned URL and its metadata
type PresignedURL struct {
	// URL is the presigned URL
	URL string

	// Method is the HTTP method to use (PUT for upload, GET for download)
	Method string

	// ExpiresAt is when the URL becomes invalid
	ExpiresAt time.Time
}
