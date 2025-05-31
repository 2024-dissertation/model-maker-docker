package repositories_test

import (
	"testing"

	database "github.com/Soup666/modelmaker/database"
	"github.com/Soup666/modelmaker/model"
	repositories "github.com/Soup666/modelmaker/repository"
	"github.com/stretchr/testify/assert"
)

func TestReportsRepository(t *testing.T) {

	err := database.SetupTestDB(t)

	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
		return
	}

	repo := repositories.NewReportsRepository(database.DB)
	userRepo := repositories.NewUserRepository(database.DB)

	// Create dummy user
	user := &model.User{
		FirebaseUid: "test_firebase_uid",
		Email:       "test@example.com",
	}

	err = userRepo.Create(user)
	assert.NoError(t, err)
	assert.NotZero(t, user.Model.ID)

	// Test Create
	report := &model.Report{
		Title:      "test_report",
		Rating:     5,
		Body:       "test_body",
		UserID:     user.Model.ID,
		ReportType: "BUG",
	}

	err = repo.CreateReport(report)
	assert.NoError(t, err)

	// Test GetTaskByID
	fetchedReport, err := repo.GetReportByID(1)
	assert.NoError(t, err)
	assert.NotNil(t, fetchedReport)
	assert.Equal(t, report.Title, fetchedReport.Title)

	// Test GetTaskByID with non-existent UID
	nonExistentUser, err := repo.GetReportByID(2)
	assert.Error(t, err)
	assert.Nil(t, nonExistentUser)

	// Test UpdateUser
	report.Title = "test_report2"
	err = repo.SaveReport(report)
	assert.NoError(t, err)

	updatedReport, _ := repo.GetReportByID(1)
	assert.Equal(t, "test_report2", updatedReport.Title)
}
