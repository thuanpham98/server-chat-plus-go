package domain_chat_model

type MessageDTO struct {
	Id   string `json:"id"`
	Sender string `json:"sender"`
	Receiver string `json:"receiver"`
	Group GroupDTO `json:"group"`
	CreateAt string `json:"create_at"`
	Content string `json:content`
	Type MessageType
}