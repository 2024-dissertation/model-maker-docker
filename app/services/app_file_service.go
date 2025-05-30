package services

import (
	"github.com/Soup666/modelmaker/model"
)

type AppFileService interface {
	Save(appFile *model.AppFile) (*model.AppFile, error)
	GetTaskFiles(taskID uint, fileType string) ([]model.AppFile, error)
	GetTaskFile(taskID uint, fileType string) (*model.AppFile, error)
}
