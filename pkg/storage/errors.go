package storage

import "errors"

var (
	// ErrNotFound indicates the requested file does not exist
	ErrNotFound = errors.New("file not found")

	// ErrAccessDenied indicates insufficient permissions to access the file
	ErrAccessDenied = errors.New("access denied")

	// ErrInvalidKey indicates the provided key is malformed or empty
	ErrInvalidKey = errors.New("invalid file key")

	// ErrFileTooLarge indicates the file exceeds the maximum allowed size
	ErrFileTooLarge = errors.New("file too large")

	// ErrInvalidContentType indicates the content type is not allowed
	ErrInvalidContentType = errors.New("invalid content type")

	// ErrUploadFailed indicates the upload operation failed
	ErrUploadFailed = errors.New("upload failed")

	// ErrDownloadFailed indicates the download operation failed
	ErrDownloadFailed = errors.New("download failed")

	// ErrPresignFailed indicates presigned URL generation failed
	ErrPresignFailed = errors.New("presigned URL generation failed")
)
