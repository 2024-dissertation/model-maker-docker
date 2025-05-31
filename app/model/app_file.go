package model

import (
	"time"

	"gorm.io/gorm"
)

type AppFile struct {
	gorm.Model
	Filename  string    `gorm:"not null"`
	Url       string    `gorm:"not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	TaskId    uint      `gorm:"not null"` // Foreign key
	FileType  string    `gorm:"not null"`
}
