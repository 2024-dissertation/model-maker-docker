package controller

import (
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/Soup666/modelmaker/database"
	"github.com/Soup666/modelmaker/model"
	services "github.com/Soup666/modelmaker/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TaskController struct {
	TaskService    services.TaskService
	AppFileService services.AppFileService
	VisionService  services.VisionService
	StorageService services.StorageService
}

func NewTaskController(taskService services.TaskService, appFileService services.AppFileService, visionService services.VisionService, storageService services.StorageService) *TaskController {
	return &TaskController{
		TaskService:    taskService,
		AppFileService: appFileService,
		VisionService:  visionService,
		StorageService: storageService,
	}
}

func (c *TaskController) GetUnarchivedTasks(ctx *gin.Context) {

	user := ctx.MustGet("user")
	userId := user.(*model.User).Model.ID

	tasks, err := c.TaskService.GetUnarchivedTasks(userId)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	for i := range tasks {
		c.TaskService.FullyLoadTask(tasks[i])
	}

	ctx.JSON(200, gin.H{"tasks": tasks})
}

func (c *TaskController) GetArchivedTasks(ctx *gin.Context) {

	user := ctx.MustGet("user")
	userId := user.(*model.User).Model.ID

	tasks, err := c.TaskService.GetArchivedTasks(userId)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	for i := range tasks {
		c.TaskService.FullyLoadTask(tasks[i])
	}

	ctx.JSON(200, gin.H{"tasks": tasks})
}

func (c *TaskController) GetTask(ctx *gin.Context) {

	// Get the Task ID from the route
	taskIDParam := ctx.Param("taskID")
	taskID, err := strconv.Atoi(taskIDParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := c.TaskService.GetTask(uint(taskID))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	// Load relations
	c.TaskService.FullyLoadTask(task)

	ctx.JSON(200, gin.H{"task": task})
}

// CreateTask handles task creation
// @Summary Create a new task
// @Description Creates a new task for the authenticated user
// @Tags tasks
// @Accept json
// @Produce json
// @Param request body CreateTaskRequest true "Task data"
// @Security BearerAuth
// @Success 201 {object} model.Task
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks [post]
func (c *TaskController) CreateTask(ctx *gin.Context) {
	user := ctx.MustGet("user").(*model.User)

	task := &model.Task{
		Title:       "",
		Description: "", // Overriden by ai-description
		UserId:      user.Model.ID,
		Completed:   false,
		Status:      "INITIAL",
	}

	err := c.TaskService.CreateTask(task)

	if err != nil {
		log.Printf("Error creating task: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create task"})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"task": task})
}

// UploadFileToTask handles file uploads for a task
// @Summary Upload files to a task
// @Description Uploads files to a task
// @Tags tasks
// @Accept json
// @Produce json
// Security BearerAuth
// @Param taskID path string true "Task ID"
// @Param files formData file true "Files to upload"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{taskID}/upload [post]
func (c *TaskController) UploadFileToTask(ctx *gin.Context) {
	// Get the Task ID from the route
	taskIdParam := ctx.Param("taskID")
	taskId, err := strconv.Atoi(taskIdParam)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Check if the Task exists
	var task model.Task
	if err := database.DB.First(&task, taskId).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	// Retrieve files from the request
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse form data"})
		return
	}
	files := form.File["files"]
	if len(files) == 0 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "No files uploaded"})
		return
	}

	var uploadedImages []model.AppFile
	var wg sync.WaitGroup
	var mu sync.Mutex
	var hasError bool

	// Use transactions for better consistency
	tx := database.DB.Begin()

	for index, file := range files {
		wg.Add(1)
		go func(index int, file *multipart.FileHeader) {
			defer wg.Done()

			// Validate file extension
			fileExt := strings.ToLower(filepath.Ext(file.Filename))
			if fileExt != ".jpg" && fileExt != ".jpeg" && fileExt != ".png" {
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file type"})
				hasError = true
				return
			}

			// Upload to object storage
			url, err := c.StorageService.UploadFile(file, uint(taskId), "upload")
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to save file %s", file.Filename)})
				hasError = true
				return
			}

			// Save metadata to DB
			image := model.AppFile{
				Filename: file.Filename,
				Url:      url,
				TaskId:   uint(taskId),
				FileType: "upload",
			}

			mu.Lock()
			if err := tx.Create(&image).Error; err != nil {
				hasError = true
			} else {
				uploadedImages = append(uploadedImages, image)
			}
			mu.Unlock()
		}(index, file)
	}

	wg.Wait()

	if hasError {
		tx.Rollback()
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload some files"})
		return
	}

	tx.Commit()

	go func() {
		// Get the first image URL from object storage
		file, err := c.StorageService.GetFile(fmt.Sprintf("uploads/%d/%s", taskId, uploadedImages[0].Filename))
		if err != nil {
			log.Printf("Unable to get image for analysis: %v", err)
			return
		}
		defer file.Close()

		// Create a temporary file
		tempFile, err := os.CreateTemp("", "analysis-*.jpg")
		if err != nil {
			log.Printf("Unable to create temp file: %v", err)
			return
		}
		defer os.Remove(tempFile.Name())
		defer tempFile.Close()

		// Copy the file content
		if _, err := io.Copy(tempFile, file); err != nil {
			log.Printf("Unable to copy file content: %v", err)
			return
		}

		// Generate caption
		result, err := c.VisionService.AnalyseImage(tempFile.Name(), "")
		if err != nil {
			log.Printf("Unable to analyze the image: %v", err)
			return
		}

		if err := c.TaskService.UpdateMeta(&task, "ai-description", result); err != nil {
			log.Printf("Failed to update task metadata: %v", err)
		}
	}()

	ctx.JSON(http.StatusOK, gin.H{"message": "Files uploaded successfully", "images": uploadedImages})
}

// StartProcess handles the process of starting the photogrammetry process
// @Summary Upload files to a task
// @Description Uploads files to a task
// @Tags tasks
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body CreateTaskRequest true "Task data"
// @Param taskID path string true "Task ID"
// @Success 201 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tasks/{taskID}/start [post]
func (c *TaskController) StartProcess(ctx *gin.Context) {

	taskId := ctx.Param("taskID")
	taskIdInt, err := strconv.Atoi(taskId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := c.TaskService.GetTask(uint(taskIdInt))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		} else {
			log.Printf("Error retrieving task: %v", err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}

	task.Completed = false
	task.Status = "QUEUED"

	if err := c.TaskService.SaveTask(task); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update task"})
		return
	}

	// Add job to queue
	if !c.TaskService.EnqueueJob(
		services.TaskJob{ID: task.ID},
	) {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Task queue is full"})
		return
	}

	ctx.JSON(http.StatusAccepted, gin.H{"message": "Process started."})
}

func (c *TaskController) UpdateTask(ctx *gin.Context) {
	task := &model.Task{}

	if err := ctx.ShouldBindJSON(task); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	user := ctx.MustGet("user").(*model.User)
	task.UserId = user.Model.ID

	err := c.TaskService.UpdateTask(task)
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"task": task})
}

func (c *TaskController) SendMessage(ctx *gin.Context) {

	taskId := ctx.Param("taskID")
	taskIdInt, err := strconv.Atoi(taskId)

	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	type MessageBody struct {
		Message string
	}

	chatMessage := &MessageBody{}

	if err := ctx.ShouldBindJSON(chatMessage); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	message, err := c.TaskService.SendMessage(uint(taskIdInt), chatMessage.Message, "USER")

	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	// Call visionService to generate a message
	go func(taskId uint) {
		task, err := c.TaskService.GetTask(taskId)

		if err != nil {
			log.Printf("Failed to get task: %v\n", err)
			return
		}

		imagePath := fmt.Sprintf("uploads/%d/%s", taskId, task.Images[0].Filename)

		if _, err := os.Stat(imagePath); os.IsNotExist(err) {
			log.Printf("Image file does not exist: %v\n", err)

			aiMessage, err := c.VisionService.GenerateMessage(fmt.Sprintf("You are a help bot for photogrammetry software. The user has asked you: %s. Please answer in a friendly and helpful manner. Keep the answer short and to the point. Do not use any technical terms. If you don't know the answer, say 'I don't know'.", message.Message))

			if err != nil {
				log.Printf("Failed to handle vision message: %v\n", err)
			}
			c.TaskService.SendMessage(uint(taskIdInt), aiMessage, "AI")
			return
		} else {
			aiMessage, err := c.VisionService.AnalyseImage(imagePath, fmt.Sprintf("You are a help bot for photogrammetry software. The user has asked you: %s. Please answer in a friendly and helpful manner. Keep the answer short and to the point. Do not use any technical terms. If you don't know the answer, say 'I don't know'. Also sent is a screenshot of the object the user is scanning.", message.Message))

			if err != nil {
				log.Printf("Failed to handle vision message: %v\n", err)
			}
			c.TaskService.SendMessage(uint(taskIdInt), aiMessage, "AI")
		}
	}(uint(taskIdInt))

	ctx.JSON(http.StatusOK, gin.H{"message": message})

}

func (c *TaskController) ArchiveTask(ctx *gin.Context) {
	taskId := ctx.Param("taskID")
	taskIdInt, err := strconv.Atoi(taskId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := c.TaskService.ArchiveTask(uint(taskIdInt))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": task})
}

func (c *TaskController) UnarchiveTask(ctx *gin.Context) {
	taskId := ctx.Param("taskID")
	taskIdInt, err := strconv.Atoi(taskId)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := c.TaskService.UnarchiveTask(uint(taskIdInt))
	if err != nil {
		ctx.AbortWithStatusJSON(400, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": task})
}
