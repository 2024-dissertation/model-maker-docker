package repositories

import (
	"github.com/Soup666/modelmaker/database"
	models "github.com/Soup666/modelmaker/model"
	"gorm.io/gorm"
)

type TaskRepositoryImpl struct {
	DB *gorm.DB
}

func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &TaskRepositoryImpl{DB: db}
}

func (repo *TaskRepositoryImpl) GetUnarchivedTasks(userID uint) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := database.DB.
		Where("user_id = ?", userID).
		Where("archived = ?", false).
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (repo *TaskRepositoryImpl) GetArchivedTasks(userID uint) ([]*models.Task, error) {
	var tasks []*models.Task
	if err := database.DB.
		Where("user_id = ?", userID).
		Where("archived = ?", true).
		Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (repo *TaskRepositoryImpl) GetTaskByID(taskID uint) (*models.Task, error) {
	var task models.Task
	if err := database.DB.Where("id = ?", taskID).
		Preload("ChatMessages").
		Preload("Images", "file_type = ?", "upload").
		Preload("Mesh", "file_type = ?", "mesh").
		Preload("Logs").
		First(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask creates a new task in the database
func (repo *TaskRepositoryImpl) CreateTask(task *models.Task) error {
	if err := database.DB.Create(task).Error; err != nil {
		return err
	}
	return nil
}

func (repo *TaskRepositoryImpl) SaveTask(task *models.Task) error {
	if err := database.DB.Save(task).Error; err != nil {
		return err
	}
	return nil
}

func (repo *TaskRepositoryImpl) ArchiveTask(taskID uint) (*models.Task, error) {
	task, err := repo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	task.Archived = true
	if err := database.DB.Save(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

func (repo *TaskRepositoryImpl) UnarchiveTask(taskID uint) (*models.Task, error) {
	task, err := repo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}

	task.Archived = false
	if err := database.DB.Save(task).Error; err != nil {
		return nil, err
	}

	return task, nil
}

func (repo *TaskRepositoryImpl) AddLog(taskID uint, log string) error {
	task, err := repo.GetTaskByID(taskID)
	if err != nil {
		return err
	}

	newLog := models.TaskLog{
		TaskId:  task.ID,
		Message: log,
	}

	task.Logs = append(task.Logs, newLog)

	if err := database.DB.Save(task).Error; err != nil {
		return err
	}
	return nil
}
