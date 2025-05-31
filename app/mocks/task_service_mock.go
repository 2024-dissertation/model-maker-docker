package mocks

import (
	models "github.com/Soup666/modelmaker/model"
	"github.com/Soup666/modelmaker/services"
	"github.com/stretchr/testify/mock"
)

// MockTaskService is a mock implementation of TaskService.
type MockTaskService struct {
	mock.Mock
	jobQueue chan services.TaskJob
}

func (m *MockTaskService) CreateTask(task *models.Task) error {
	args := m.Called(task)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) UnarchiveTask(taskID uint) (*models.Task, error) {
	args := m.Called(taskID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetUnarchivedTasks(taskID uint) ([]*models.Task, error) {
	args := m.Called(taskID)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetTask(taskID uint) (*models.Task, error) {
	args := m.Called(taskID)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) GetArchivedTasks(taskID uint) ([]*models.Task, error) {
	args := m.Called(taskID)
	if args.Get(0) != nil {
		return args.Get(0).([]*models.Task), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) UpdateTask(task *models.Task) error {
	args := m.Called(task)
	if args.Get(0) != nil {
		return args.Error(1)
	}
	return nil
}

func (m *MockTaskService) ArchiveTask(taskID uint) (*models.Task, error) {
	args := m.Called(taskID)
	if args.Get(0) != nil {
		return nil, args.Get(0).(error)
	}
	return nil, args.Get(0).(error)
}

func (m *MockTaskService) SaveTask(task *models.Task) error {
	args := m.Called(task)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) FailTask(task *models.Task, message string) error {
	args := m.Called(task, message)
	if args.Get(0) != nil {

		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) RunPhotogrammetryProcess(task *models.Task) error {
	args := m.Called(task)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) UpdateMeta(task *models.Task, key string, value interface{}) error {
	args := m.Called(task, key, value)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) FullyLoadTask(task *models.Task) (*models.Task, error) {
	args := m.Called(task)
	if args.Get(0) != nil {
		return args.Get(0).(*models.Task), args.Error(1)
	}
	return nil, nil
}

func (m *MockTaskService) SendMessage(taskID uint, message string, sender string) (*models.ChatMessage, error) {
	args := m.Called(taskID, message, sender)
	if args.Get(0) != nil {
		return args.Get(0).(*models.ChatMessage), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTaskService) AddLog(taskID uint, log string) error {
	args := m.Called(taskID, log)
	if args.Get(0) != nil {
		return args.Get(0).(error)
	}
	return nil
}

func (m *MockTaskService) EnqueueJob(job services.TaskJob) bool {
	m.jobQueue <- job
	m.Called(job)
	return true
}

func (m *MockTaskService) StartWorker() {
	m.Called()
}

func (m *MockTaskService) GetJobQueue() chan services.TaskJob {
	return m.jobQueue
}
