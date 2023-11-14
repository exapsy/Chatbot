package bot_app

import (
	"connectly-interview/internal/bot/domain/bot_prompter"
	"connectly-interview/internal/bot/interfaces"
	"context"
	"sync"
)

type Bot struct {
	m          sync.RWMutex
	ctx        context.Context
	interfaces *bot_interfaces.Interfaces
	*bot_prompter.Prompter
}

type Option func(b *Bot) error

func WithHttpServer(addr string) Option {
	return func(b *Bot) error {
		err := b.interfaces.InitHttpServer(addr, nil, nil)
		if err != nil {
			return err
		}

		return nil
	}
}

func New(opts ...Option) (*Bot, error) {
	ctx := context.Background()
	bot := &Bot{
		ctx: ctx,
	}

	for _, o := range opts {
		err := o(bot)
		if err != nil {
			return nil, err
		}
	}

	return bot, nil
}

func (b *Bot) Prompt(prompt string) error {
	return nil
}
