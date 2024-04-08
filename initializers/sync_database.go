package initializers

import (
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
)

func SyncDatabase(){
	DB.AutoMigrate(&domain_auth_model.UserEntity{})
}