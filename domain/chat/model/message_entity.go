package domain_chat_model

import (
	"time"

	"gorm.io/gorm"
)


type MessageEntity struct{
	ID string `gorm:"type:uuid;primary_key"`
	CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt 
	Sender string
	Receiver string
	Type MessageType
	Group GroupEntity `gorm:"foreignkey:Id"`
	Content string
}