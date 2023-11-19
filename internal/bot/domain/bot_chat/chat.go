package bot_chat

import (
	"connectly-interview/internal/libs/lists"
	"fmt"
	"github.com/google/uuid"
)

var (
	ErrChatsCapacityFull = fmt.Errorf("chats' capacity is full")
)

type ChatId uuid.UUID

func NewChatId() ChatId {
	uuid, err := uuid.NewUUID()
	if err != nil {
		panic(fmt.Errorf("unexpected error: could not create uuid: %s", err))
	}

	return ChatId(uuid)
}

type Chat struct {
	id              ChatId
	history         [][]byte
	historyCapacity uint16
}

type Args struct {
	HistoryCapacity uint16
}

func New(args Args) *Chat {
	chatId := NewChatId()
	history := make([][]byte, args.HistoryCapacity)
	return &Chat{
		id:              chatId,
		history:         history,
		historyCapacity: args.HistoryCapacity,
	}
}

func (c *Chat) Id() ChatId {
	return c.id
}

func (c *Chat) AppendAnswer(answer []byte) {
	if uint16(len(c.history)) > c.historyCapacity {
		curHistorySize := len(c.history)
		c.history = c.history[:curHistorySize-1]
	}
	c.history = append(c.history, answer)
}

type Chats struct {
	items      map[ChatId]*Chat
	itemsQueue *lists.Queue[*Chat]
	capacity   uint16
	length     uint16
}

type ChatsArgs struct {
	Capacity uint16
}

func NewChats(args ChatsArgs) *Chats {
	chatsMap := make(map[ChatId]*Chat, args.Capacity)
	return &Chats{
		items:      chatsMap,
		capacity:   args.Capacity,
		itemsQueue: nil,
		length:     0,
	}
}

func (c *Chats) Get(chatId ChatId) *Chat {
	if c.items == nil {
		c.items = make(map[ChatId]*Chat)
	}
	chat, ok := c.items[chatId]
	if !ok {
		return nil
	}
	return chat
}

func (c *Chats) Delete(chatId ChatId) error {
	if c.length == 0 {
		return fmt.Errorf("no chats exist")
	}

	chat, found := c.items[chatId]
	if !found {
		return fmt.Errorf("chat with id %q not found", chatId)
	}

	// delete from queue
	err := c.itemsQueue.Delete(chat)
	if err != nil {
		return fmt.Errorf("could not delete chat with id %q from the chat queue", chatId)
	}

	// remove from hashmap
	delete(c.items, chatId)

	c.length--

	return nil
}

func (c *Chats) New(historyCapacity uint16) (*Chat, error) {
	// full capacity reached
	if c.length >= c.capacity {
		return nil, ErrChatsCapacityFull
	}

	newChat := New(Args{
		HistoryCapacity: historyCapacity,
	})

	c.itemsQueue.Enqueue(newChat)
	c.items[newChat.id] = newChat
	c.length++

	return newChat, nil
}
