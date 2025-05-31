package services

import (
	"github.com/Soup666/modelmaker/model"
)

type TaskJob struct {
	ID uint
}

type TaskService interface {
	CreateTask(task *model.Task) error
	GetTask(taskID uint) (*model.Task, error)
	GetUnarchivedTasks(userID uint) ([]*model.Task, error)
	GetArchivedTasks(userID uint) ([]*model.Task, error)
	UpdateTask(task *model.Task) error
	ArchiveTask(taskID uint) (*model.Task, error)
	UnarchiveTask(taskID uint) (*model.Task, error)
	SaveTask(task *model.Task) error
	FailTask(task *model.Task, message string) error
	RunPhotogrammetryProcess(task *model.Task) error
	UpdateMeta(task *model.Task, key string, value interface{}) error
	FullyLoadTask(task *model.Task) (*model.Task, error)
	SendMessage(taskID uint, message string, sender string) (*model.ChatMessage, error)
	AddLog(taskID uint, log string) error
	EnqueueJob(job TaskJob) bool
	StartWorker()
	GetJobQueue() chan TaskJob
}
