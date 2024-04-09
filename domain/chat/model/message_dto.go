package domain_chat_model

import domain_auth_model "github.com/thuanpham98/go-websocker-server/domain/auth/model"

type MessageDTO struct {
	Id   string `json:"id"`
	Sender domain_auth_model.UserDTO `json:"sender"`
	Receiver domain_auth_model.UserDTO `json:"receiver"`
	Group string `json:"phone"`
	CreateAt string `json:"create_at"`
	Content string `json:content`
}