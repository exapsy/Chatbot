package types

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"fmt"
)

type Communication_interface_incoming_msg_topic int

const (
	Communication_interface_incoming_msg_topic_new_chat = iota + 1
	Communication_interface_incoming_msg_topic_new_chat_msg
	Communication_interface_incoming_msg_topic_err
)

type Communication_interface_incoming_msg struct {
	Topic Communication_interface_incoming_msg_topic
	Msg   []byte
}

type Communication_interface_incoming_new_chat_reply struct {
	ChatId bot_chat.ChatId `json:"chat_id,omitempty"`
	Prompt string          `json:"prompt"`
}

func (msg *Communication_interface_incoming_new_chat_reply) Json() string {
	jsonMsg := fmt.Sprintf(`{"chat_id": %q", "prompt": %q}`, msg.ChatId, msg.Prompt)
	return jsonMsg
}

type Communication_interface_incoming_new_chat struct {
	FromUser string
}

func (msg *Communication_interface_incoming_new_chat) Json() string {
	jsonMsg := fmt.Sprintf(`{"from_user": %q"}`, msg.FromUser)
	return jsonMsg
}
