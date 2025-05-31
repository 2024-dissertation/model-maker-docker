package database

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/Soup666/modelmaker/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() error {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_TIMEZONE"))

	fmt.Println("Connecting to database with DSN:", dsn)
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Error connecting to the database: ", err)
		return err
	}

	err = MigrateScheme()

	if err != nil {
		log.Fatal("Failed to migrate test DB: ", err)
		return err
	}

	log.Println("Connected to database", dsn)

	return nil
}

func SetupTestDB(t *testing.T) error {

	log.Println("Connecting to database...")

	err := ConnectDatabase()

	ResetTestDB()

	if err != nil {
		t.Fatalf("Failed to connect to test DB: %v", err)
		return err
	}

	err = MigrateScheme()

	if err != nil {
		t.Fatalf("Failed to migrate test DB: %v", err)
		return err
	}
	return nil
}

func ResetTestDB() {
	log.Println("Resetting test database...")

	tables := []string{"users", "tasks", "reports", "collections"}
	for _, table := range tables {
		truncateTableCommand := fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", table)
		DB.Exec(truncateTableCommand)
	}

}

func MigrateScheme() error {
	createEnumCommand := `
	--create types
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'taskstatus') THEN
        CREATE TYPE TaskStatus AS ENUM
        (
            'SUCCESS', 'INPROGRESS', 'FAILED', 'INITIAL', 'QUEUED'
        );
    END IF;
	IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'reporttype') THEN
        CREATE TYPE ReportType AS ENUM
        (
            'BUG', 'FEEDBACK'
        );
    END IF;
END$$;
`

	DB.Exec(createEnumCommand)

	err := DB.AutoMigrate(&model.User{}, &model.Task{}, &model.Report{}, &model.Collection{}, &model.ChatMessage{}, &model.AppFile{}, &model.TaskLog{})
	if err != nil {
		log.Fatalf("Failed to migrate test DB: %v", err)
		return err
	}

	log.Println("Database migrated successfully")
	return nil
}
