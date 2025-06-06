package router

import (
	"fmt"
	"io"

	"github.com/Soup666/modelmaker/controller"
	"github.com/Soup666/modelmaker/middleware"
	"github.com/Soup666/modelmaker/services"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	authController *controller.AuthController,
	taskController *controller.TaskController,
	uploadController *controller.UploadController,
	objectController *controller.ObjectController,
	visionController *controller.VisionController,
	authService services.AuthService,
	reportsController *controller.ReportsController,
	collectionsController *controller.CollectionsController,
	userAnalyticsController *controller.UserAnalyticsController,
	notificationController *controller.NotificationController,
) *gin.Engine {

	r := gin.Default()

	// Global middlewares
	r.Use(middleware.CORSMiddleware())

	// Swagger route
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Authenticated routes
	authRequired := r.Group("/")
	authRequired.Use(middleware.AuthMiddleware(authService))

	// Authentication routes
	authRequired.POST("/verify", authController.Verify)
	authRequired.PATCH("/verify", authController.Verify)

	authRequired.POST("/unverify", authController.Unverify)

	// Tasks (protected by AuthMiddleware)
	authRequired.GET("/tasks", taskController.GetUnarchivedTasks)
	authRequired.GET("/archived/tasks", taskController.GetArchivedTasks)
	authRequired.POST("/tasks", taskController.CreateTask)
	authRequired.PUT("/tasks", taskController.UpdateTask)
	authRequired.GET("/tasks/:taskID", taskController.GetTask)
	authRequired.POST("/tasks/:taskID/upload", taskController.UploadFileToTask)
	authRequired.POST("/tasks/:taskID/start", taskController.StartProcess)
	authRequired.POST("/tasks/:taskID/message", taskController.SendMessage)
	authRequired.POST("/tasks/:taskID/archive", taskController.ArchiveTask)
	authRequired.POST("/tasks/:taskID/unarchive", taskController.UnarchiveTask)

	// Anlytics
	authRequired.GET("/analytics", userAnalyticsController.GetAnalytics)

	// Reports
	authRequired.GET("/reports", reportsController.GetReports)
	authRequired.POST("/reports", reportsController.CreateReport)
	authRequired.GET("/reports/:reportID", reportsController.GetReportByID)
	authRequired.PUT("/reports", reportsController.SaveReport)

	// Collections
	authRequired.GET("/collections", collectionsController.GetCollections)
	authRequired.POST("/collections", collectionsController.CreateCollection)
	authRequired.GET("/collections/:collectionID", collectionsController.GetCollection)
	authRequired.PUT("/collections", collectionsController.SaveCollection)
	authRequired.DELETE("/collections/:collectionID", collectionsController.ArchiveCollection)

	// Image analysis
	authRequired.POST("/analyze", visionController.AnalyzeImage)
	authRequired.POST("/analyze/:taskID", visionController.AnalyzeTask)

	// Debug
	authRequired.POST("/debug/notification", notificationController.SendMessage)
	authRequired.POST("/debug/storage", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(400, gin.H{"error": "No file provided"})
			return
		}
		storageService := services.NewKatapultStorageService()
		url, err := storageService.UploadFile(file, 1, "test")
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to upload file: %v", err)})
			return
		}
		c.JSON(200, gin.H{"status": "ok", "url": url})
	})
	authRequired.GET("/debug/storage", func(c *gin.Context) {
		storageService := services.NewKatapultStorageService()
		reader, err := storageService.GetFile("uploads/1/00006._c.png")
		if err != nil {
			c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to get file: %v", err)})
			return
		}
		io.Copy(c.Writer, reader)
		reader.Close()
	})

	// Unauthenticated routes
	r.POST("/uploads", uploadController.UploadFile)
	r.GET("/uploads/:taskId/:filename", uploadController.GetFile)
	r.GET("/objects/:taskID/model", objectController.GetObject)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	return r
}
