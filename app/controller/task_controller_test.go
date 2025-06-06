package controller_test

import (
	"net/http"
	"testing"

	"github.com/Soup666/modelmaker/controller"
	"github.com/Soup666/modelmaker/mocks"
	models "github.com/Soup666/modelmaker/model"
	utils "github.com/Soup666/modelmaker/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTaskController(t *testing.T) {

	mockTaskService := new(mocks.MockTaskService)
	mockAppFileService := new(mocks.MockAppFileService)
	mockVisionService := new(mocks.MockVisionService)
	mockStorageService := new(mocks.MockStorageService)

	taskController := controller.NewTaskController(mockTaskService, mockAppFileService, mockVisionService, mockStorageService)

	t.Run("GetUnarchivedTasks", func(t *testing.T) {
		recorder, c := utils.SetupRecorder()

		mockTaskService.On("GetUnarchivedTasks", uint(1)).Return([]*models.Task{}, nil)

		taskController.GetUnarchivedTasks(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
		assert.JSONEq(t, `{"tasks":[]}`, recorder.Body.String())
	})

	t.Run("GetTask", func(t *testing.T) {
		recorder, c := utils.SetupRecorder()

		mockTaskService.On("GetTask", uint(1)).Return(&models.Task{}, nil)
		mockTaskService.On("FullyLoadTask", &models.Task{}).Return(nil)

		c.AddParam("taskID", "1")

		taskController.GetTask(c)

		assert.Equal(t, http.StatusOK, recorder.Code)
	})

	t.Run("CreateTask", func(t *testing.T) {
		recorder, c := utils.SetupRecorder()

		mockTaskService.On("CreateTask", mock.Anything).Return(nil)

		taskController.CreateTask(c)

		assert.Equal(t, http.StatusCreated, recorder.Code)
	})

	// t.Run("SendMessage", func(t *testing.T) {
	// 	recorder, c := utils.SetupRecorder()

	// 	// Mock the request body to parse into a message
	// 	c.Request = &http.Request{
	// 		Header: make(http.Header),
	// 	}

	// 	utils.MockJsonPost(c, map[string]interface{}{
	// 		"Id":        0,
	// 		"TaskId":    0,
	// 		"Sender":    "",
	// 		"Message":   "",
	// 		"CreatedAt": "0001-01-01T00:00:00Z",
	// 	})

	// 	c.AddParam("taskID", "1")

	// 	mockTaskService.On("SendMessage", mock.Anything, mock.Anything, "USER").Return(&models.ChatMessage{}, nil)

	// 	taskController.SendMessage(c)

	// 	assert.Equal(t, http.StatusOK, recorder.Code)
	// })
}
