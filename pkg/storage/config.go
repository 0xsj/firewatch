package storage

import "time"

// Config contains the configuration for storage backends
type Config struct {
	// Provider is the storage backend type (e.g., "s3", "minio")
	Provider string `yaml:"provider" env:"STORAGE_PROVIDER" default:"s3"`

	// Bucket is the storage bucket name
	Bucket string `yaml:"bucket" env:"STORAGE_BUCKET" required:"true"`

	// Region is the bucket region
	Region string `yaml:"region" env:"STORAGE_REGION" default:"us-east-1"`

	// Endpoint is the storage service endpoint (required for MinIO, optional for S3)
	Endpoint string `yaml:"endpoint" env:"STORAGE_ENDPOINT"`

	// AccessKeyID is the access key for authentication
	AccessKeyID string `yaml:"accessKeyId" env:"STORAGE_ACCESS_KEY_ID" required:"true"`

	// SecretAccessKey is the secret key for authentication
	SecretAccessKey string `yaml:"secretAccessKey" env:"STORAGE_SECRET_ACCESS_KEY" required:"true"`

	// UsePathStyle forces path-style addressing (required for MinIO)
	UsePathStyle bool `yaml:"usePathStyle" env:"STORAGE_USE_PATH_STYLE" default:"false"`

	// DefaultPresignExpiry is the default expiry duration for presigned URLs
	DefaultPresignExpiry time.Duration `yaml:"defaultPresignExpiry" env:"STORAGE_DEFAULT_PRESIGN_EXPIRY" default:"15m"`

	// MaxFileSize is the maximum allowed file size in bytes (0 = unlimited)
	MaxFileSize int64 `yaml:"maxFileSize" env:"STORAGE_MAX_FILE_SIZE" default:"104857600"` // 100MB

	// NOTE: For tenant isolation, add TenantBucketStrategy field later
	// Options: "prefix" (tenant-id/path), "bucket-per-tenant" (tenant-id-bucket)
}

// DefaultConfig returns sensible defaults for local development with MinIO
func DefaultConfig() Config {
	return Config{
		Provider:             "s3",
		Bucket:               "uploads",
		Region:               "us-east-1",
		Endpoint:             "http://localhost:9000",
		AccessKeyID:          "minioadmin",
		SecretAccessKey:      "minioadmin",
		UsePathStyle:         true,
		DefaultPresignExpiry: 15 * time.Minute,
		MaxFileSize:          104857600, // 100MB
	}
}
