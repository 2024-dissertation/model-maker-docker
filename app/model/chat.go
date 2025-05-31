package model

import (
	"time"

	"gorm.io/gorm"
)

var CHAT_MESSAGE_JSON string = `{
				"ID":0,
				"TaskId":0,
				"Sender":"",
				"Message":"",
				"CreatedAt":"0001-01-01T00:00:00Z"
			}`

type ChatMessage struct {
	gorm.Model
	TaskId    uint      `gorm:"not null"` // Foreign key
	Sender    string    `gorm:"type:text;not null;check:sender IN ('USER','AI')"`
	Message   string    `gorm:"type:text;not null"`
	CreatedAt time.Time `gorm:"not null;default:now()"`
}
