package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Soup666/modelmaker/database"
	models "github.com/Soup666/modelmaker/model"
	"github.com/Soup666/modelmaker/services"

	"github.com/gin-gonic/gin"
)

// AuthController is the controller for handling authentication requests
type UploadController struct {
	storageService services.StorageService
}

func NewUploadController(storageService services.StorageService) *UploadController {
	return &UploadController{
		storageService: storageService,
	}
}

func (c *UploadController) UploadFile(ctx *gin.Context) {
	file, header, err := ctx.Request.FormFile("file")
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "File upload failed"})
		return
	}
	defer file.Close()

	// Upload file to object storage
	url, err := c.storageService.UploadFile(header, 0, "upload")
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
		return
	}

	// Save file metadata in the database
	image := models.AppFile{
		Filename: header.Filename,
		Url:      url,
	}
	database.DB.Create(&image)

	ctx.JSON(http.StatusOK, gin.H{"message": "File uploaded successfully", "image": image})
}

func (c *UploadController) GetFile(ctx *gin.Context) {
	taskId := ctx.Param("taskId")
	filename := ctx.Param("filename")

	// Convert taskId to uint
	taskIdInt, err := strconv.ParseUint(taskId, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get file from object storage
	file, err := c.storageService.GetFile(fmt.Sprintf("uploads/%d/%s", taskIdInt, filename))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}
	defer file.Close()

	// Stream the file directly to the response writer
	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream file"})
		return
	}
}

func (c *UploadController) GetObject(ctx *gin.Context) {
	filename := ctx.Param("filename")
	taskId := ctx.Param("taskID")

	// Convert taskId to uint
	taskIdInt, err := strconv.ParseUint(taskId, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get file from object storage
	file, err := c.storageService.GetFile(fmt.Sprintf("objects/%d/%s", taskIdInt, filename))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}
	defer file.Close()

	// Stream the file to the response
	ctx.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, file)
		return err == nil
	})
}
