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
	Chat *bot_chat.Chat
	Msg  string
}

type Prompter interface {
	Start() error
	Prompt(prompt *Prompt) (answer <-chan []byte, err error)
}

// Prompter is the engine that compiles the strings into answers
type prompter struct {
	ctx context.Context
	// queue is used to store
	// the int that is hash for the user that prompts a string,
	// string is the prompt
	queue         chan *Prompt
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

func New(args Args) Prompter {
	if args.ChatsCapacity == 0 {
		args.ChatsCapacity = DefaultChatsCapacity
	}

	if args.HistoryCapacity == 0 {
		args.HistoryCapacity = DefaultHistoryCapacity
	}

	return &prompter{
		ctx:           args.Context,
		queue:         make(chan *Prompt, args.PromptQueueBuffer),
		workers:       NewWorkers(args.WorkersAmount),
		promptTimeout: args.PromptTimeout,
	}
}

func (p *prompter) Start() error {
	go p.accept_prompts()

	return nil
}

func (p *prompter) accept_prompts() {
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
					prompt.Chat.AppendAnswer(answer)
				}
			}
		}
	}
}

func (p *prompter) waitWorkerToPrompt(prompt *Prompt) (answer <-chan []byte) {
	answer = make(<-chan []byte)
	// wait for any worker to be free
	go func() {

	}()
	return answer
}

func (p *prompter) Prompt(prompt *Prompt) (answer <-chan []byte, err error) {
	answer = make(chan []byte)
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
