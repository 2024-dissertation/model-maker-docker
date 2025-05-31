package main

import (
	"context"
	"log"
	"os"

	"github.com/Soup666/modelmaker/controller"
	db "github.com/Soup666/modelmaker/database"
	repositories "github.com/Soup666/modelmaker/repository"
	"github.com/Soup666/modelmaker/router"
	"github.com/Soup666/modelmaker/services"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	_ "github.com/joho/godotenv/autoload"
)

func main() {
	// Set up the database connection
	log.Println("Connecting to database...")

	db.ConnectDatabase()

	// Create a Firebase app instance
	opt := option.WithCredentialsFile(os.Getenv("GOOGLE_CREDENTIALS_FILE"))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("Failed to create Firebase app: %v", err)
	}

	// Create a Firebase auth client instance
	authClient, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("Failed to create Firebase auth client: %v", err)
	}

	// Set up the repositories
	userRepo := repositories.NewUserRepository(db.DB)
	taskRepo := repositories.NewTaskRepository(db.DB)
	appFileRepo := repositories.NewAppFileRepository(db.DB)
	reportsRepo := repositories.NewReportsRepository(db.DB)
	collectionsRepo := repositories.NewCollectionsRepository(db.DB)
	chatRepo := repositories.NewChatRepository(db.DB)
	userAnalyticsRepo := repositories.NewUserAnalyticsRepository(db.DB)

	// Set up the authentication service
	authService := services.NewAuthService(authClient, db.DB, userRepo)
	userService := services.NewUserService(userRepo)
	notificationService := services.NewNotificationService()
	appFileService := services.NewAppFileServiceFile(appFileRepo)
	taskService := services.NewTaskService(taskRepo, appFileService, chatRepo, notificationService)
	visionService := services.NewVisionService()
	reportsService := services.NewReportsService(reportsRepo)
	collectionsService := services.NewCollectionsService(collectionsRepo)
	userAnalyticsService := services.NewUserAnalyticsService(userAnalyticsRepo)

	// Initialise Job Queue
	taskService.StartWorker()

	authController := controller.NewAuthController(authService, userService)
	taskController := controller.NewTaskController(&taskService, appFileService, visionService)
	uploadController := controller.NewUploadController()
	objectController := controller.NewObjectController()
	visionController := controller.NewVisionController(visionService, taskRepo, &taskService)
	reportsController := controller.NewReportsController(reportsService)
	collectionsController := controller.NewCollectionsController(collectionsService)
	userAnalyticsController := controller.NewUserAnalyticsController(userAnalyticsService)
	notificationController := controller.NewNotificationController(notificationService)

	// Set up the HTTP router
	r := router.NewRouter(authController, taskController, uploadController, objectController, visionController, authService, reportsController, collectionsController, userAnalyticsController, notificationController)

	// Start the server
	if r.Run(":"+os.Getenv("PORT")) != nil {
		panic("[Error] failed to start Gin server due to: " + err.Error())
	}

}
