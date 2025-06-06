package mocks

import (
	"io"
	"mime/multipart"

	"github.com/stretchr/testify/mock"
)

// MockVisionService is a mock implementation of VisionService.
type MockStorageService struct {
	mock.Mock
}

func (m *MockStorageService) UploadFile(file *multipart.FileHeader, taskID uint, fileType string) (string, error) {
	args := m.Called(file, taskID, fileType)
	if args.Get(0) != nil {
		return args.Get(0).(string), args.Error(1)
	}
	return "", args.Error(1)
}
func (m *MockStorageService) UploadFromReader(reader io.Reader, taskID uint, filename string, fileType string) (string, error) {
	args := m.Called(reader, taskID, filename, fileType)
	if args.Get(0) != nil {
		return args.Get(0).(string), args.Error(1)
	}
	return "", args.Error(1)
}

func (m *MockStorageService) GetFile(filepath string) (io.ReadCloser, error) {
	args := m.Called(filepath)
	if args.Get(0) != nil {
		return args.Get(0).(io.ReadCloser), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockStorageService) DeleteFile(taskID uint, filename string) error {
	args := m.Called(taskID, filename)
	return args.Error(0)
}
