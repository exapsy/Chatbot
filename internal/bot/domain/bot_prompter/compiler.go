package bot_prompter

import (
	"connectly-interview/internal/bot/domain/bot_chat"
	"connectly-interview/internal/bot/infrastructure/openai"
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
	queue          chan *Prompt
	m              sync.RWMutex
	workers        Workers
	promptTimeout  time.Duration
	answersMapChan map[bot_chat.ChatId]chan []byte
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

// waitWorkerToPrompt sends the prompt to an available worker with round-robin,
// and returns a channel through the worker will respond to.
//
// This approach may not scale well later, it would probably be better to have a completely separate executable for workers,
// listening to a daemon or getting HTTP requests whenever there's work to do for a worker, it works for now, but to take care later.
func (p *prompter) waitWorkerToPrompt(prompt *Prompt) <-chan []byte {
	answerChan := make(chan []byte)

	go func() {
		// Select a worker (simple example, consider a more sophisticated method for production)
		var selectedWorker *Worker
		for _, worker := range p.workers {
			if !worker.IsBusy() {
				selectedWorker = &worker
				break
			}
		}

		if selectedWorker == nil {
			answerChan <- []byte("No available workers")
			return
		}

		// Use the selected worker to compile the response
		response, err := selectedWorker.Compile(prompt.Msg)
		if err != nil {
			answerChan <- []byte(fmt.Sprintf("Error compiling prompt: %s", err))
			return
		}

		answerChan <- []byte(response)
	}()

	return answerChan
}

func (p *prompter) Prompt(prompt *Prompt) (answer <-chan []byte, err error) {
	p.queue <- prompt
	chatId := prompt.Chat.Id()
	promptChan := make(chan []byte)
	p.answersMapChan[chatId] = promptChan
	return promptChan, err
}

type Worker struct {
	id        uint
	m         sync.Mutex
	openaiKey string
	isBusy    bool
}

type Workers []Worker

func NewWorkers(total_workers uint8) Workers {
	w := make(Workers, total_workers)
	return w
}

func (w *Worker) Id() uint {
	return w.id
}

func (w *Worker) IsBusy() bool {
	return w.isBusy
}

func (w *Worker) Compile(prompt string) (string, error) {
	w.m.Lock()
	defer w.m.Unlock()

	if w.isBusy {
		return "", fmt.Errorf("worker %d is busy", w.id)
	}
	w.isBusy = true
	defer func() {
		w.isBusy = false
	}()

	openAiAnswer, err := openai.Prompt(w.openaiKey, prompt)
	if err != nil {
		return "", err
	}

	if openAiAnswer == "" {
		return "", fmt.Errorf("no response from GPT-4")
	}

	return openAiAnswer, nil
}

type PromptUser struct {
	Id bot_chat.ChatId
}
