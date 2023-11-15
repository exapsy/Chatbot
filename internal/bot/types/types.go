package types

import (
	"connectly-interview/internal/bot/domain/bot_chat"
)

type MessageBotPrompt struct {
	ChatId bot_chat.ChatId
	Prompt string
}
