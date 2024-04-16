package domain_chat_model

type MessageType int32
const (
    TEXT MessageType = 0
	IMAGE MessageType = 1
	FILE  MessageType = 2
)
