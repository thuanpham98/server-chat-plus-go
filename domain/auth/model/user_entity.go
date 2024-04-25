package domain_auth_model

import (
	"time"

	"gorm.io/gorm"
)

type UserEntity struct{
	Id   string `gorm:"type:uuid;primary_key"`
	Name string
	Email string `gorm:"unique"`
	Phone string `gorm:"unique"`
	Password string
	CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
}