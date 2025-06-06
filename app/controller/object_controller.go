package controller

import (
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/Soup666/modelmaker/services"
	"github.com/gin-gonic/gin"
)

// AuthController is the controller for handling authentication requests
type ObjectController struct {
	storageService services.StorageService
}

func NewObjectController(storageService services.StorageService) *ObjectController {
	return &ObjectController{
		storageService: storageService,
	}
}

func (c *ObjectController) GetObject(ctx *gin.Context) {
	taskId := ctx.Param("taskID")

	// Convert taskId to uint
	taskIdInt, err := strconv.ParseUint(taskId, 10, 32)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	// Get file from object storage
	file, err := c.storageService.GetFile(fmt.Sprintf("objects/%d/final.glb", taskIdInt))
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}
	defer file.Close()

	// Add headers
	ctx.Header("Content-Type", "model/gltf-binary")
	ctx.Header("Content-Disposition", "attachment; filename=final.glb")

	// // Stream the file to the response
	// ctx.Stream(func(w io.Writer) bool {
	// 	_, err := io.Copy(w, file)
	// 	return err == nil
	// })
	// Copy file contents to response writer
	_, err = io.Copy(ctx.Writer, file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to stream file"})
		return
	}
}
