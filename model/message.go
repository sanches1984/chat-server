package model

import "encoding/json"

type MessageType string

const (
	MessageEnter   MessageType = "enter"
	MessageExit    MessageType = "exit"
	MessagePublic  MessageType = "public"
	MessagePrivate MessageType = "private"
	MessageStat    MessageType = "stat"
	MessageList    MessageType = "list"
)

type Message struct {
	UserName string      `json:"username"`
	To       string      `json:"to"`
	Type     MessageType `json:"type"`
	Message  string      `json:"message"`
}

func (m Message) ToJSON() []byte {
	bytes, _ := json.Marshal(&m)
	return bytes
}
