package services

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/2024-dissertation/openmvgo/pkg/mvgoutils"
	"github.com/2024-dissertation/openmvgo/pkg/openmvg"
	"github.com/2024-dissertation/openmvgo/pkg/openmvs"
	models "github.com/Soup666/modelmaker/model"
	repositories "github.com/Soup666/modelmaker/repository"
	"github.com/Soup666/modelmaker/utils"
	"gorm.io/gorm"
)

type TaskServiceImpl struct {
	taskRepo            repositories.TaskRepository
	appFileService      AppFileService
	chatRepository      repositories.ChatRepository
	notificationService NotificationService
	storageService      StorageService
	jobQueue            chan TaskJob
}

func NewTaskService(
	taskRepo repositories.TaskRepository,
	appFileService AppFileService,
	chatRepository repositories.ChatRepository,
	notificationService NotificationService,
	storageService StorageService,
) TaskServiceImpl {
	return TaskServiceImpl{
		taskRepo:            taskRepo,
		appFileService:      appFileService,
		chatRepository:      chatRepository,
		notificationService: notificationService,
		storageService:      storageService,
		jobQueue:            make(chan TaskJob, 100),
	}
}

func (s *TaskServiceImpl) CreateTask(task *models.Task) error {
	err := s.taskRepo.CreateTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskServiceImpl) GetTask(taskID uint) (*models.Task, error) {
	task, err := s.taskRepo.GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskServiceImpl) GetUnarchivedTasks(userID uint) ([]*models.Task, error) {

	tasks, err := s.taskRepo.GetUnarchivedTasks(userID)

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskServiceImpl) GetArchivedTasks(userID uint) ([]*models.Task, error) {

	tasks, err := s.taskRepo.GetArchivedTasks(userID)

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *TaskServiceImpl) UpdateTask(task *models.Task) error {

	err := s.taskRepo.SaveTask(task)

	if err != nil {
		return err
	}
	return nil
}

func (s *TaskServiceImpl) UpdateMeta(task *models.Task, key string, value interface{}) error {
	if task.Metadata == nil {
		task.Metadata = make(map[string]interface{})
	}
	task.Metadata[key] = value
	err := s.UpdateTask(task)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskServiceImpl) ArchiveTask(taskID uint) (*models.Task, error) {

	task, err := s.taskRepo.ArchiveTask(taskID)

	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskServiceImpl) UnarchiveTask(taskID uint) (*models.Task, error) {

	task, err := s.taskRepo.UnarchiveTask(taskID)

	if err != nil {
		return nil, err
	}
	return task, nil
}

func (s *TaskServiceImpl) SaveTask(task *models.Task) error {
	err := s.taskRepo.SaveTask(task)

	if err != nil {
		return err
	}
	return nil
}

func (s *TaskServiceImpl) FailTask(task *models.Task, message string) error {
	task.Status = models.FAILED
	if err := s.UpdateTask(task); err != nil {
		return err
	}

	if err := s.AddLog(task.ID, message); err != nil {
		log.Printf("Failed to add log: %v\n", err)
	}

	log.Printf("Task %d failed: %s\n", task.ID, message)

	s.notificationService.SendMessage(&models.Notification{
		UserID:  task.UserId,
		Message: "Task failed",
		Title:   task.Title,
	})

	return nil
}

func (s *TaskServiceImpl) GetTaskFiles(taskID uint, fileType string) ([]models.AppFile, error) {
	files, err := s.appFileService.GetTaskFiles(taskID, fileType)
	if err != nil {
		return nil, err
	}
	return files, nil
}

func (s *TaskServiceImpl) GetTaskFile(taskID uint, fileType string) (*models.AppFile, error) {
	file, err := s.appFileService.GetTaskFile(taskID, fileType)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func (s *TaskServiceImpl) FullyLoadTask(task *models.Task) (*models.Task, error) {
	files, err := s.GetTaskFiles(task.ID, "upload")
	if err != nil {
		return nil, err
	}
	task.Images = files

	mesh, err := s.GetTaskFile(task.ID, "mesh")
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			task.Mesh = nil
		} else {
			return nil, err
		}
	} else {
		task.Mesh = mesh
	}

	return task, nil
}

func (s *TaskServiceImpl) SendMessage(taskID uint, message string, sender string) (*models.ChatMessage, error) {
	chatMessage := &models.ChatMessage{
		Message: message,
		TaskId:  taskID,
		Sender:  sender,
	}

	err := s.chatRepository.CreateChat(chatMessage)
	if err != nil {
		return chatMessage, err
	}

	return chatMessage, nil
}

func (s *TaskServiceImpl) AddLog(taskID uint, log string) error {

	err := s.taskRepo.AddLog(taskID, log)
	if err != nil {
		return err
	}
	return nil
}

func (s *TaskServiceImpl) EnqueueJob(job TaskJob) bool {

	select {
	case s.jobQueue <- job:
		return true
	default:
		return false
	}
}

func (s *TaskServiceImpl) initializeTask(task *models.Task) error {
	task.Status = models.INPROGRESS
	if err := s.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}

	if err := s.UpdateMeta(task, "opensfm-process", 0.0); err != nil {
		return fmt.Errorf("failed to update meta: %v", err)
	}

	s.notificationService.SendMessage(&models.Notification{
		UserID:  task.UserId,
		Message: "Scan started",
		Title:   task.Title,
	})

	return nil
}

func (s *TaskServiceImpl) setupDirectories(timestamp int64) (string, string, string, error) {
	buildDir, err := os.MkdirTemp("", fmt.Sprintf("%d-build", timestamp))
	if err != nil {
		return "", "", "", fmt.Errorf("failed to create build directory: %v", err)
	}

	outputDir, err := os.MkdirTemp("", fmt.Sprintf("%d-convert", timestamp))
	if err != nil {
		os.RemoveAll(buildDir)
		return "", "", "", fmt.Errorf("failed to create output directory: %v", err)
	}

	inputDir, err := os.MkdirTemp("", fmt.Sprintf("%d-input", timestamp))
	if err != nil {
		os.RemoveAll(buildDir)
		os.RemoveAll(outputDir)
		return "", "", "", fmt.Errorf("failed to create input directory: %v", err)
	}

	return buildDir, outputDir, inputDir, nil
}

func (s *TaskServiceImpl) downloadTaskFiles(task *models.Task, inputDir string) error {
	for _, taskFile := range task.Images {
		fmt.Printf("Downloading file: %s to %s\n", taskFile.Url, filepath.Join(inputDir, taskFile.Filename))
		file, err := s.storageService.GetFile(taskFile.Url)
		if err != nil {
			return fmt.Errorf("failed to get file: %v", err)
		}
		defer file.Close()

		dstFile, err := os.Create(filepath.Join(inputDir, taskFile.Filename))
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dstFile.Close()

		if _, err = io.Copy(dstFile, file); err != nil {
			return fmt.Errorf("failed to copy file: %v", err)
		}
	}
	return nil
}

func (s *TaskServiceImpl) runPipelines(utils mvgoutils.OpenmvgoUtilsInterface, inputDir, buildDir, outputDir string) error {
	cameraDBFile := filepath.Join("bin", "sensor_width_camera_database.txt")

	openmvgService := openmvg.NewOpenMVGService(
		openmvg.NewOpenMVGConfig(
			inputDir,
			buildDir,
			&cameraDBFile,
		),
		utils,
	)

	openmvsService := openmvs.NewOpenMVSService(
		openmvs.NewOpenMVSConfig(
			outputDir,
			buildDir,
			1,
		),
		utils,
	)

	openmvgService.PopulateTmpDir()
	defer os.RemoveAll(openmvgService.Config.MatchesDir)
	defer os.RemoveAll(openmvgService.Config.ReconstructionDir)

	openmvgService.SfMSequentialPipeline()

	// Run OpenMVS pipeline with panic recovery
	if err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("OpenMVS pipeline failed: %v", r)
			}
		}()
		openmvsService.RunPipeline()
		return nil
	}(); err != nil {
		return err
	}

	return nil
}

func (s *TaskServiceImpl) convertAndStoreMesh(task *models.Task, outputDir string) error {
	fileName := filepath.Join(outputDir, "scene_dense_mesh_refine_texture.obj")
	convertedFileName := filepath.Join(outputDir, "scene_dense_mesh_refine_texture.glb")

	cmd := exec.Command("blender", "-b", "-P", "./bin/convert_obj_to_glb.py", "--", fileName, convertedFileName)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("mesh conversion failed: %v", err)
	}

	if err := s.UpdateMeta(task, "log", "MeshConversion started"); err != nil {
		log.Printf("Failed to update meta: log: %v\n", err)
	}

	mesh, err := s.appFileService.Save(&models.AppFile{
		Url:      fileName + ".glb",
		Filename: "final.glb",
		TaskId:   task.ID,
		FileType: "mesh",
	})
	if err != nil {
		return fmt.Errorf("failed to save mesh: %v", err)
	}

	file, err := os.Open(convertedFileName)
	if err != nil {
		return fmt.Errorf("failed to open converted file: %v", err)
	}
	defer file.Close()

	meshUrl, err := s.storageService.UploadFromReader(file, task.ID, "final.glb", "mesh")
	if err != nil {
		return fmt.Errorf("failed to upload mesh: %v", err)
	}

	mesh.Url = meshUrl
	task.Mesh = mesh
	task.Completed = true
	task.Status = models.SUCCESS

	if err := s.UpdateTask(task); err != nil {
		return fmt.Errorf("failed to update task: %v", err)
	}

	s.notificationService.SendMessage(&models.Notification{
		UserID:  task.UserId,
		Message: "Scan finished",
		Title:   task.Title,
	})

	return nil
}

func (s *TaskServiceImpl) processTask(job TaskJob) {
	fmt.Printf("Processing job: %+v\n", job)

	task, err := s.taskRepo.GetTaskByID(job.ID)
	if err != nil {
		log.Printf("Failed to get task: %v\n", err)
		return
	}

	utils.PrettyPrint(task)

	if err := s.initializeTask(task); err != nil {
		s.FailTask(task, err.Error())
		return
	}

	timestamp := time.Now().Unix()

	buildDir, outputDir, inputDir, err := s.setupDirectories(timestamp)
	if err != nil {
		s.FailTask(task, err.Error())
		return
	}
	// defer os.RemoveAll(buildDir)
	// defer os.RemoveAll(outputDir)
	// defer os.RemoveAll(inputDir)

	if err := s.downloadTaskFiles(task, inputDir); err != nil {
		s.FailTask(task, err.Error())
		return
	}

	if err := s.runPipelines(nil, inputDir, buildDir, outputDir); err != nil {
		s.FailTask(task, err.Error())
		return
	}

	if err := s.convertAndStoreMesh(task, outputDir); err != nil {
		s.FailTask(task, err.Error())
		return
	}

	fmt.Println("OpenMVGO pipeline completed successfully!")
}

func (s *TaskServiceImpl) StartWorker() {
	go func() {
		for job := range s.jobQueue {
			go s.processTask(job)
		}
	}()
}

func (s *TaskServiceImpl) GetJobQueue() chan TaskJob {
	return s.jobQueue
}

// func (s *TaskServiceImpl) RunPhotogrammetryProcess(task *models.Task) error {

// 	if task.Status == models.INPROGRESS {
// 		log.Printf("Task %d is already in progress\n", task.ID)
// 		return nil
// 	}

// 	startTime := time.Now()

// 	TASK_COUNT := 7.0
// 	CURRENT_TASK := 0.0

// 	inputPath := filepath.Join("uploads", fmt.Sprintf("%d", task.ID))
// 	outputPath := filepath.Join("objects", fmt.Sprintf("%d", task.ID))
// 	mvsPath := filepath.Join(outputPath, "mvs")

// 	task.Status = models.INPROGRESS
// 	if err := s.UpdateTask(task); err != nil {
// 		log.Printf("Failed to update task status to INPROGRESS: %v\n", err)
// 		return err
// 	}

// 	// Clear the build directory
// 	if err := os.RemoveAll(outputPath); err != nil {
// 		s.FailTask(task, fmt.Sprintf("Failed to clear directory %s: %v", outputPath, err))
// 		return err
// 	}

// 	if err := os.MkdirAll(outputPath, os.ModePerm); err != nil {
// 		s.FailTask(task, fmt.Sprintf("Failed to create directory %s: %v", outputPath, err))
// 		return err
// 	}

// 	if err := os.MkdirAll(mvsPath, os.ModePerm); err != nil {
// 		s.FailTask(task, fmt.Sprintf("Failed to create directory %s: %v", mvsPath, err))
// 		return err
// 	}

// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	// 1
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}

// 	log.Println("# 1 ./bin/SfM_SequentialPipeline.py", inputPath, outputPath, "--opensfm-processes", "8")
// 	cmd := exec.Command("./bin/SfM_SequentialPipeline.py", inputPath, outputPath, "--opensfm-processes", "8")

// 	var stdoutBuffer saveOutput
// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err := cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("SfM_SequentialPipeline failed: %s", err))
// 		return err
// 	}

// 	// 2
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}
// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: task.Title + " - Processing started",
// 		Title:   task.Title,
// 	})

// 	log.Println("# 2 openMVG_main_openMVG2openMVS", "-i", filepath.Join(outputPath, "reconstruction_sequential/sfm_data.bin"), "-o", filepath.Join(mvsPath, "scene.mvs"), inputPath, outputPath, "-d", mvsPath)
// 	cmd = exec.Command("openMVG_main_openMVG2openMVS", "-i", filepath.Join(outputPath, "reconstruction_sequential/sfm_data.bin"), "-o", filepath.Join(mvsPath, "scene.mvs"), inputPath, outputPath, "-d", mvsPath)

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		log.Println()
// 		s.FailTask(task, fmt.Sprintf("openMVG_main_openMVG2openMVS failed: %s", err))
// 		return err
// 	}

// 	// 3
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}
// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step 2/7 started",
// 		Title:   task.Title,
// 	})

// 	log.Println("# 3 DensifyPointCloud", "scene.mvs", "-o", "scene_dense.mvs", "-w", mvsPath, "--max-threads", "1")
// 	cmd = exec.Command("DensifyPointCloud", "scene.mvs", "-o", "scene_dense.mvs", "-w", mvsPath, "--max-threads", "1")

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("DensifyPointCloud failed: %s", err))
// 		return err
// 	}

// 	// 4
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}
// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step 3/7 started",
// 		Title:   task.Title,
// 	})

// 	log.Println("# 4 ReconstructMesh", "scene_dense.mvs", "-o", "scene_mesh.ply", "-w", mvsPath)
// 	cmd = exec.Command("ReconstructMesh", "scene_dense.mvs", "-o", "scene_mesh.ply", "-w", mvsPath)

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("ReconstructMesh failed: %v", err))
// 		return err
// 	}

// 	// 5
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}
// 	if err := s.AddLog(task.ID, "ReconstructMesh started"); err != nil {
// 		log.Printf("Failed to add log: %v\n", err)
// 	}
// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step 4/7 started",
// 		Title:   task.Title,
// 	})

// 	log.Println("# 5 RefineMesh", "scene.mvs", "-m", "scene_mesh.ply", "-o", "scene_dense_mesh_refine.mvs", "-w", mvsPath, "--scales", "1", "--max-face-area", "16", "--max-threads", "1")
// 	cmd = exec.Command("RefineMesh", "scene.mvs", "-m", "scene_mesh.ply", "-o", "scene_dense_mesh_refine.mvs", "-w", mvsPath, "--scales", "1", "--max-face-area", "16", "--max-threads", "1")

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("RefineMesh failed: %v", err))
// 		return err
// 	}

// 	// 6
// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}

// 	if err := s.AddLog(task.ID, "TextureMesh started"); err != nil {
// 		log.Printf("Failed to add log: %v\n", err)
// 	}
// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step 5/7 started",
// 		Title:   task.Title,
// 	})

// 	log.Println("# 6 TextureMesh", "scene_dense.mvs", "-m", "scene_dense_mesh_refine.ply", "-o", "scene_dense_mesh_refine_texture.mvs", "-w", mvsPath, "--export-type", "obj")
// 	cmd = exec.Command("TextureMesh", "scene_dense.mvs", "-m", "scene_dense_mesh_refine.ply", "-o", "scene_dense_mesh_refine_texture.mvs", "-w", mvsPath, "--export-type", "obj")

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer.savedOutput); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("TextureMesh failed: %v", err))
// 		return err
// 	}

// 	// 7

// 	CURRENT_TASK = CURRENT_TASK + 1.0

// 	log.Println("Updating meta for task:", task.ID, " - ", CURRENT_TASK, "/", TASK_COUNT)
// 	if err := s.UpdateMeta(task, "opensfm-process", (CURRENT_TASK/TASK_COUNT)*100.0); err != nil {
// 		log.Printf("Failed to update meta: %f: %v\n", CURRENT_TASK, err)
// 		return err
// 	}

// 	if err := s.AddLog(task.ID, "MeshConversion started"); err != nil {
// 		log.Printf("Failed to add log: %v\n", err)
// 	}

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step 6/7 started",
// 		Title:   task.Title,
// 	})

// 	fileName := filepath.Join(mvsPath, "final_model")
// 	fmt.Println("blender", "-b", "-P", "./bin/convert_obj_to_glb.py", "--", filepath.Join(mvsPath, "scene_dense_mesh_refine_texture.obj"), fileName)
// 	cmd = exec.Command("blender", "-b", "-P", "./bin/convert_obj_to_glb.py", "--", filepath.Join(mvsPath, "scene_dense_mesh_refine_texture.obj"), fileName)

// 	cmd.Stdout = &stdoutBuffer
// 	cmd.Stderr = &stdoutBuffer
// 	err = cmd.Run()

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("MeshConversion failed: %v", err))
// 		return err
// 	}

// 	if err := s.UpdateMeta(task, "log", stdoutBuffer); err != nil {
// 		log.Printf("Failed to update meta: log: %v\n", err)
// 		return err
// 	}

// 	mesh, err := s.appFileService.Save(&models.AppFile{
// 		Url:      fileName + ".glb",
// 		Filename: "final_model.glb",
// 		TaskId:   task.ID,
// 		FileType: "mesh",
// 	})

// 	if err != nil {
// 		s.FailTask(task, fmt.Sprintf("Failed to Save mesh: %v", err))
// 		return err
// 	}

// 	task.Mesh = mesh
// 	task.Completed = true
// 	task.Status = models.SUCCESS

// 	if err := s.UpdateTask(task); err != nil {
// 		s.FailTask(task, fmt.Sprintf("Failed to update task: %v", err))
// 		return err
// 	}

// 	s.notificationService.SendMessage(&models.Notification{
// 		UserID:  task.UserId,
// 		Message: "Step Scan finished",
// 		Title:   task.Title,
// 	})

// 	log.Println("Task updated successfully.")
// 	log.Printf("Processing completed in %s\n", time.Since(startTime))
// 	return nil
// }
