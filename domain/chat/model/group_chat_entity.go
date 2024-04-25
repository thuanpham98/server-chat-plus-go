package domain_chat_model

import (
	"time"

	"gorm.io/gorm"
)

type GroupEntity struct {
	Id string `gorm:"primaryKey"`
	CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
	Name string
}