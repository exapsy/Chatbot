package bot_prompter

import (
	"bytes"
	"connectly-interview/internal/bot/domain/bot_chat"
	"connectly-interview/internal/bot/infrastructure/openai"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	answer = make(chan []byte)
	p.queue <- prompt
	return answer, err
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

	// Prepare the request body
	requestBody, err := json.Marshal(map[string]string{
		"model":  "text-davinci-004", // or any other GPT-4 model you want to use
		"prompt": prompt,
	})
	if err != nil {
		return "", err
	}

	// Make the HTTP request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/text-davinci-004/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.openaiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var gptResponse openai.GPTResponse
	err = json.Unmarshal(body, &gptResponse)
	if err != nil {
		return "", err
	}

	// Assuming the first choice is the one we need
	if len(gptResponse.Choices) > 0 {
		return gptResponse.Choices[0].Text, nil
	}

	w.isBusy = false

	return "", fmt.Errorf("no response from GPT-4")
}

type PromptUser struct {
	Id bot_chat.ChatId
}
