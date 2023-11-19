package bot_app

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"connectly-interview/internal/bot/domain/bot_prompter"
	"connectly-interview/internal/bot/infrastructure/kafka"
	"connectly-interview/internal/bot/infrastructure/kafka/nokafka"
	"connectly-interview/internal/bot/infrastructure/kafka/segmentio"
	"connectly-interview/internal/bot/infrastructure/openai"
	"connectly-interview/internal/bot/interfaces"
	"connectly-interview/internal/bot/types"
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultPromptTimeout       = time.Second * 30
	DefaultWorkersAmount       = 30
	DefaultQueueBuffer   uint8 = 125
	DefaultChatsCapacity       = 256
)

type Bot struct {
	m              sync.RWMutex
	ctx            context.Context
	interfaces     *bot_interfaces.Interfaces
	prompter       bot_prompter.Prompter
	chats          *bot_chat.Chats
	bus            bot_infrastructure_kafka.Kafka
	openaiApiKey   string
	newChatChan    <-chan struct{}
	newChatMsgChan <-chan []byte
}

type Option func(b *Bot) error

func WithOpenAiKey(key string) Option {
	return func(b *Bot) error {
		b.openaiApiKey = key
		return nil
	}
}

func WithHttpServer(addr string) Option {
	return func(b *Bot) error {
		err := b.interfaces.InitHttpServer(addr, nil, nil)
		if err != nil {
			return err
		}

		return nil
	}
}

func WithKafka(addr string) Option {
	return func(b *Bot) error {
		kafkaDialer := bot_infastructure_kafka_segmentio.New(b.ctx, addr)
		b.bus = kafkaDialer
		return nil
	}
}

func WithNoKafka() Option {
	return func(b *Bot) error {
		kafkaDialer := bot_infrastructure_kafka_nokafka.NewNoKafka()
		b.bus = kafkaDialer
		return nil
	}
}

func New(opts ...Option) (*Bot, error) {
	// bot vars
	ctx := context.Background()
	bot := &Bot{}

	newChatChan := make(chan struct{})
	newChatMsgChan := make(chan []byte)

	prompter := bot_prompter.New(bot_prompter.Args{
		Context:           ctx,
		PromptTimeout:     DefaultPromptTimeout,
		WorkersAmount:     DefaultWorkersAmount,
		PromptQueueBuffer: DefaultQueueBuffer,
		ChatsCapacity:     DefaultChatsCapacity,
	})
	chats := bot_chat.NewChats(bot_chat.ChatsArgs{
		Capacity: 24,
	})
	comm_interfaces := bot_interfaces.New(
		ctx,
		bot_interfaces.WithMessageQueueCapacity(24),
		bot_interfaces.WithNewChatHandler(func() {
			if bot.bus == nil {
				fmt.Printf("no bus found\n")
				return
			}

			_, err := chats.New(0)
			if err != nil {
				fmt.Printf("could not create new chat: %s", err)
				return
			}
			newChatBusMsg := types.Communication_interface_incoming_new_chat{
				FromUser: "exapsy", // TODO: hard-written ...... we want this to be from the specific user that the chat was made from
			}
			err = bot.bus.Send(bot_infrastructure_kafka.TopicPrompt, []byte(newChatBusMsg.Json()))
			if err != nil {
				fmt.Printf("could not send : %s\n", err)
				return
			}

		}),
		bot_interfaces.WithNewChatMessageHandler(func(chatId bot_chat.ChatId, msg []byte) error {
			var err error

			if bot.bus == nil {
				return fmt.Errorf("no bus found")
			}

			//newChatMsgBusMsg := types.Communication_interface_incoming_new_chat_reply{
			//	ChatId: chatId,
			//	Prompt: string(msg),
			//}

			prompt, err := openai.Prompt(bot.openaiApiKey, string(msg))
			if err != nil {
				return err
			}

			err = bot.bus.Send(bot_infrastructure_kafka.TopicPrompt, []byte(prompt))
			if err != nil {
				return fmt.Errorf("could not send message %q to bus: %w", string(msg), err)
			}

			return nil
		}),
	)

	bot.chats = chats
	bot.ctx = ctx
	bot.interfaces = comm_interfaces
	bot.prompter = prompter
	bot.newChatMsgChan = newChatMsgChan
	bot.newChatChan = newChatChan

	// apply options
	for _, o := range opts {
		err := o(bot)
		if err != nil {
			return nil, err
		}
	}

	return bot, nil
}

func (b *Bot) Start() error {
	var wg sync.WaitGroup

	var err error

	err = openai.Ping(b.openaiApiKey)
	if err != nil {
		return fmt.Errorf("could not ping to openai - make sure to provide the OpenAI key and that it is a valid one:\n%w", err)
	}

	err = b.prompter.Start()
	if err != nil {
		return fmt.Errorf("failed to start promtper: %w", err)
	}

	reqsChan := b.interfaces.Listen()

	// Loop looking over prompt requests
	for {
		select {
		case req, closed := <-reqsChan:
			if closed {
				fmt.Printf("interfaces channel is closed, returning ...")
				return nil
			}
			// unmarshall msg prompt
			msg := types.Communication_interface_incoming_msg{}
			err = json.Unmarshal(req, &msg)
			if err != nil {
				continue
			}

			fmt.Printf("got message: %+v\n", msg)

			// compile an answer
			prompt := &bot_prompter.Prompt{}
			responseChan, err := b.prompter.Prompt(prompt)
			if err != nil {
				fmt.Printf("error on prompt %+v: %s", prompt, err)
			}

			fmt.Printf("got answer for message %q: %q\n", msg, prompt.Msg)

			// Wait for response from the prompt
			go func() {
				wg.Add(1)
				defer wg.Done()
				// TODO:
				//   should it be configurable?
				//   Like "ReadPromptTimeout"?
				//   And shouldn't we cancel the job then to put the prompt's worker at rest?
				//   Things to think ... second part about the job could change the whole architecture of how workers work, with jobs
				timeout := time.Second * 10
				timeoutTicker := time.NewTicker(timeout)
				for {
					select {
					case response := <-responseChan:
						// Send response to bus
						err = b.bus.Send(bot_infrastructure_kafka.TopicPrompt, response)
						if err != nil {
							fmt.Printf("error sending to msg %q to the bus: %s", err)
							return
						}
					case <-timeoutTicker.C:
						return
					}
				}
			}()
		}
	}
}

func (b *Bot) Prompt(prompt string) error {
	return nil
}
