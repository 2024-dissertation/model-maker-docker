package services

import (
	"io"
	"mime/multipart"
)

// StorageService defines the interface for object storage operations
type StorageService interface {
	// UploadFile uploads a file to object storage and returns its public URL
	UploadFile(file *multipart.FileHeader, taskID uint, fileType string) (string, error)

	// UploadFromReader uploads data from an io.Reader to object storage
	UploadFromReader(reader io.Reader, taskID uint, filename string, fileType string) (string, error)

	// GetFile retrieves a file from object storage
	GetFile(filepath string) (io.ReadCloser, error)

	// DeleteFile deletes a file from object storage
	DeleteFile(taskID uint, filename string) error
}
