package infrastructure

import (
	domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"
	domain_chat_model "github.com/thuanpham98/go-websocker-server/domain/chat/model"
)

func SyncDatabase(){
	DB.AutoMigrate(&domain_auth_model.UserEntity{})
	DB.AutoMigrate(&domain_chat_model.GroupEntity{})
	DB.AutoMigrate(&domain_chat_model.MessageEntity{})
}