package domain_chat_model

import "strings"

type MessageType string
const (
    TEXT MessageType = "TEXT"
	IMAGE MessageType = "IMAGE";
	FILE  MessageType = "FILE";
)

var (
    capabilitiesMessageType = map[string]MessageType{
        "text":   TEXT,
        "image": IMAGE,
        "file": FILE,
    }
)
func ParseString(str string) (MessageType, bool) {
    c, ok := capabilitiesMessageType[strings.ToLower(str)]
    return c, ok
}
