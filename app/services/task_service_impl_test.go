package services_test

import (
	"errors"
	"testing"

	"github.com/Soup666/modelmaker/mocks"
	models "github.com/Soup666/modelmaker/model"
	"github.com/Soup666/modelmaker/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestTaskService(t *testing.T) {
	mockTaskRepository := new(mocks.MockTaskRepository)
	mockChatRepository := new(mocks.MockChatRepository)

	mockAppFileService := new(mocks.MockAppFileService)
	mockNotificationService := new(mocks.MockNotificationService)
	mockStorageService := new(mocks.MockStorageService)

	taskService := services.NewTaskService(mockTaskRepository, mockAppFileService, mockChatRepository, mockNotificationService, mockStorageService)

	t.Run("CreateTask", func(t *testing.T) {

		var userId = uint(1)

		task := &models.Task{
			Title:       "Test Task",
			Description: "This is a test task",
			UserId:      userId,
		}

		mockTaskRepository.On("CreateTask", task).Return(nil)
		mockTaskRepository.On("CreateTask", nil).Return(errors.New("error"))
		mockTaskRepository.On("CreateTask", &models.Task{}).Return(errors.New("error"))
		mockTaskRepository.On("CreateTask", &models.Task{Model: gorm.Model{ID: 1}}).Return(nil)

		err := taskService.CreateTask(task)

		mockTaskRepository.AssertCalled(t, "CreateTask", task)
		assert.NoError(t, err)
		assert.NotNil(t, task)
		assert.NotNil(t, task.ID)
		assert.Equal(t, task.Title, "Test Task")
	})

	t.Run("GetTask", func(t *testing.T) {
		task := &models.Task{
			Title:       "Test Task",
			Description: "This is a test task",
		}

		mockTaskRepository.On("GetTaskByID", uint(1)).Return(task, nil)
		mockTaskRepository.On("GetTaskByID", uint(2)).Return(nil, errors.New("error"))

		fetchedTask, err := taskService.GetTask(1)

		mockTaskRepository.AssertCalled(t, "GetTaskByID", uint(1))
		assert.NoError(t, err)
		assert.NotNil(t, fetchedTask)
		assert.Equal(t, fetchedTask.Title, "Test Task")
	})

	t.Run("GetTaskByID with non-existent ID", func(t *testing.T) {
		mockTaskRepository.On("GetTaskByID", uint(2)).Return(nil, errors.New("error"))

		fetchedTask, err := taskService.GetTask(2)

		mockTaskRepository.AssertCalled(t, "GetTaskByID", uint(2))

		assert.Error(t, err)
		assert.Nil(t, fetchedTask)
		assert.Equal(t, err.Error(), "error")
	})

	t.Run("GetUnarchivedTasks", func(t *testing.T) {

		var userId = uint(1)

		tasks := []*models.Task{
			{
				Model:       gorm.Model{ID: 1},
				Title:       "Test Task",
				Description: "This is a test task",
				UserId:      userId,
			},
			{
				Model:       gorm.Model{ID: 2},
				Title:       "Test Task 2",
				Description: "This is a test task 2",
				UserId:      userId,
			},
		}

		mockTaskRepository.On("GetUnarchivedTasks", userId).Return(tasks, nil)
		mockTaskRepository.On("GetTasksByUser", uint(2)).Return(nil, errors.New("error"))

		fetchedTasks, err := taskService.GetUnarchivedTasks(userId)

		mockTaskRepository.AssertCalled(t, "GetUnarchivedTasks", userId)
		assert.NoError(t, err)
		assert.NotNil(t, fetchedTasks)
		assert.Equal(t, len(fetchedTasks), 2)
	})

	t.Run("UpdateTask", func(t *testing.T) {
		task := &models.Task{
			Title:       "Test Task",
			Description: "This is a test task",
		}

		updatedTask := &models.Task{
			Model:       gorm.Model{ID: 1},
			Title:       "Test Task",
			Description: "This is a test task",
		}

		mockTaskRepository.On("SaveTask", task).Return(nil)

		err := taskService.UpdateTask(task)

		mockTaskRepository.AssertCalled(t, "SaveTask", task)
		assert.NoError(t, err)
		assert.NotNil(t, updatedTask)
		assert.Equal(t, updatedTask.ID, uint(1))
	})

	t.Run("DeleteTask", func(t *testing.T) {

		mockTaskRepository.On("ArchiveTask", uint(1)).Return(nil)

		_, err := taskService.ArchiveTask(uint(1))

		mockTaskRepository.AssertCalled(t, "ArchiveTask", uint(1))
		assert.NoError(t, err)
	})

	t.Run("SaveTask", func(t *testing.T) {
		task := &models.Task{
			Model:       gorm.Model{ID: 1},
			Title:       "Test Task",
			Description: "This is a test task",
		}

		mockTaskRepository.On("SaveTask", task).Return(nil)

		err := taskService.SaveTask(task)

		mockTaskRepository.AssertCalled(t, "SaveTask", task)
		assert.NoError(t, err)
	})

	// t.Run("FailTask", func(t *testing.T) {
	// 	task := &models.Task{
	// 		Model:       gorm.Model{ID: 1},
	// 		Title:       "Test Task",
	// 		Description: "This is a test task",
	// 	}

	// 	mockTaskRepository.On("UpdateTask", task).Return(nil)

	// 	err := taskService.FailTask(task, "Failed due to an error")

	// 	assert.NoError(t, err)
	// 	assert.Equal(t, task.Status, models.FAILED)
	// })

	t.Run("SendMessage", func(t *testing.T) {

		mockChatRepository.On("CreateChat", mock.Anything).Return(nil)

		chat, err := taskService.SendMessage(uint(1), "Hello World", "USER")

		mockChatRepository.AssertCalled(t, "CreateChat", mock.Anything)
		assert.NoError(t, err)
		assert.NotNil(t, chat)
		assert.NotNil(t, chat.ID)
		assert.Equal(t, chat.Message, "Hello World")
	})

	t.Run("EnqueueJob", func(t *testing.T) {
		job := services.TaskJob{ID: 1}

		// Enqueue the job
		taskService.EnqueueJob(job)

		// Read from the queue to verify
		select {
		case queuedJob := <-taskService.GetJobQueue():
			assert.Equal(t, job.ID, queuedJob.ID)
		default:
			t.Fatal("Job was not enqueued")
		}

	})
}
