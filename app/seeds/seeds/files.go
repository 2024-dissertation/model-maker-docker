package seeds

import (
	"fmt"

	"github.com/Soup666/modelmaker/model"
	"gorm.io/gorm"
)

func CreateFile(db *gorm.DB, appFile *model.AppFile) error {
	return db.Create(appFile).Error
}

var filenames = []string{
	"3-0.png",
	"3-1.png",
	"3-2.png",
	"3-3.png",
	"3-4.png",
	"3-5.png",
	"3-6.png",
	"3-7.png",
	"3-8.png",
	"3-9.png",
	"3-10.png",
	"3-11.png",
	"3-12.png",
	"3-13.png",
	"3-14.png",
	"3-15.png",
	"3-16.png",
	"3-17.png",
	"3-18.png",
	"3-19.png",
	"3-20.png",
	"3-21.png",
	"3-22.png",
	"3-23.png",
	"3-24.png",
	"3-25.png",
	"3-26.png",
	"3-27.png",
	"3-28.png",
	"3-29.png",
	"3-30.png",
	"3-31.png",
	"3-32.png",
	"3-33.png",
	"3-34.png",
	"3-35.png",
	"3-36.png",
	"3-37.png",
	"3-38.png",
	"3-39.png",
	"3-40.png",
	"3-41.png",
	"3-42.png",
	"3-43.png",
	"3-44.png",
	"3-45.png",
	"3-46.png",
	"3-47.png",
	"3-48.png",
	"3-49.png",
	"3-50.png",
	"3-51.png",
	"3-52.png",
	"3-53.png",
	"3-54.png",
	"3-55.png",
	"3-56.png",
	"3-57.png",
	"3-58.png",
	"3-59.png",
	"3-60.png",
	"3-61.png",
	"3-62.png",
	"3-63.png",
	"3-64.png",
	"3-65.png",
	"3-66.png",
	"3-67.png",
	"3-68.png",
	"3-69.png",
	"3-70.png",
	"3-71.png",
	"3-72.png",
	"3-73.png",
	"3-74.png",
	"3-75.png",
	"3-76.png",
	"3-77.png",
	"3-78.png",
	"3-79.png",
	"3-80.png",
	"3-81.png",
	"3-82.png",
}

// CreateDummyFiles creates dummy files in the database for testing purposes.
func CreateDummyFiles(db *gorm.DB, taskId uint) ([]model.AppFile, error) {
	files := make([]model.AppFile, len(filenames))

	for i, filename := range filenames {
		files[i] = model.AppFile{
			TaskId:   taskId,
			Url:      fmt.Sprintf("uploads/%d/%s", taskId, filename),
			Filename: filename,
			FileType: "upload",
		}
	}

	for _, f := range files {
		if err := CreateFile(db, &f); err != nil {
			return nil, err
		}
	}

	return files, nil
}

func CreateDummyMesh(db *gorm.DB) (*model.AppFile, error) {

	mesh := &model.AppFile{
		TaskId:   1,
		Url:      "/objects/1/mvs/final_model.glb",
		Filename: "final_model.glb",
		FileType: "mesh",
	}

	if err := CreateFile(db, mesh); err != nil {
		return nil, err
	}

	return mesh, nil
}
