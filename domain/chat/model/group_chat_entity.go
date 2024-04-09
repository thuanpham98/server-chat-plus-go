package domain_chat_model

import (
	"time"

	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	"gorm.io/gorm"
)

type GroupEntity struct {
	Id string `gorm:"primaryKey"`
	CreatedAt time.Time
    UpdatedAt time.Time
    DeletedAt gorm.DeletedAt `gorm:"index"`
	Name string
	Member []domain_auth_model.UserEntity `gorm:"foreignkey:Id"`
}