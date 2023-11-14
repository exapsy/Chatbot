package bot_prompter

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"context"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultChatsCapacity   uint16 = 256
	DefaultHistoryCapacity uint16 = 1024
)

type Prompt struct {
	ChatId bot_chat.ChatId
}

// Prompter is the engine that compiles the strings into answers
type Prompter struct {
	ctx context.Context
	// queue is used to store
	// the int that is hash for the user that prompts a string,
	// string is the prompt
	queue         chan *Prompt
	chats         *bot_chat.Chats
	m             sync.RWMutex
	workers       Workers
	promptTimeout time.Duration
}

type Args struct {
	Context           context.Context
	PromptTimeout     time.Duration
	WorkersAmount     uint8
	HistoryCapacity   uint16
	PromptQueueBuffer uint8
	ChatsCapacity     uint16
}

func New(args Args) *Prompter {
	if args.ChatsCapacity == 0 {
		args.ChatsCapacity = DefaultChatsCapacity
	}

	if args.HistoryCapacity == 0 {
		args.HistoryCapacity = DefaultHistoryCapacity
	}

	chats := bot_chat.NewChats(bot_chat.ChatsArgs{
		Capacity: args.ChatsCapacity,
	})

	return &Prompter{
		ctx:           args.Context,
		queue:         make(chan *Prompt, args.PromptQueueBuffer),
		chats:         chats,
		workers:       NewWorkers(args.WorkersAmount),
		promptTimeout: args.PromptTimeout,
	}
}

func (p *Prompter) Start() error {
	go func() {
	outerloop:
		for {
			select {
			case prompt := <-p.queue:
				answerChan := p.waitWorkerToPrompt(prompt)
				timeoutTicker := time.NewTicker(p.promptTimeout)
			answerLoop:
				for {
					select {
					case <-timeoutTicker.C:
						break answerLoop
					case <-p.ctx.Done():
						break outerloop
					case answer := <-answerChan:
						chat := p.chats.Get(prompt.ChatId)
						if chat == nil {
							fmt.Printf("error: could not find chat with id %q", prompt.ChatId)
							continue
						}

						chat.AppendAnswer(answer)
					}
				}
			}
		}
	}()
	return nil
}

func (p *Prompter) waitWorkerToPrompt(prompt *Prompt) (answer <-chan string) {
	answer = make(<-chan string)
	// wait for any worker to be free
	go func() {

	}()
	return answer
}

func (p *Prompter) Prompt(chatId bot_chat.ChatId, prompt *Prompt) (answer <-chan string, err error) {
	answer = make(chan string)
	p.queue <- prompt
	return answer, err
}

type Worker struct {
	id uint
	m  sync.Mutex
}

type Workers []Worker

func NewWorkers(total_workers uint8) Workers {
	w := make(Workers, total_workers)
	return w
}

func (w *Worker) Id() uint {
	return w.id
}

func (w *Worker) Compile(prompt string) (string, error) {
	w.m.Lock()
	defer w.m.Unlock()

	// TODO: Compile and do not return a constant string
	return fmt.Sprintf("hello from the worker %d for prompt %q", w.id, prompt), nil
}

type PromptUser struct {
	Id bot_chat.ChatId
}
