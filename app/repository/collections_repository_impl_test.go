package repositories_test

import (
	"testing"

	database "github.com/Soup666/modelmaker/database"
	"github.com/Soup666/modelmaker/model"
	repositories "github.com/Soup666/modelmaker/repository"
	"github.com/stretchr/testify/assert"
)

func TestCollectionsRepository(t *testing.T) {

	err := database.SetupTestDB(t)

	if err != nil {
		t.Fatalf("Failed to setup test DB: %v", err)
		return
	}

	repo := repositories.NewCollectionsRepository(database.DB)
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
	collection := &model.Collection{
		Name:   "Test Collection",
		Tasks:  []model.Task{},
		UserID: user.Model.ID,
	}

	err = repo.CreateCollection(collection)
	assert.NoError(t, err)

	// Test GetTaskByID
	// fetchedCollection, err := repo.GetCollectionByID(1)
	// assert.NoError(t, err)
	// assert.NotNil(t, fetchedCollection)
	// assert.Equal(t, collection.Name, fetchedCollection.Name)

	// Test GetTaskByID with non-existent UID
	nonExistentUser, err := repo.GetCollectionByID(2)
	assert.Error(t, err)
	assert.Nil(t, nonExistentUser)

	// Test UpdateUser
	collection.Name = "Test Collection 2"
	err = repo.SaveCollection(collection)
	assert.NoError(t, err)

	updatedTask, _ := repo.GetCollectionByID(1)
	assert.Equal(t, "Test Collection 2", updatedTask.Name)
}
