package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

var (
	ErrNoResponse = fmt.Errorf("no response")
)

type Model string

func (m Model) String() string {
	return string(m)
}

const (
	Model35 Model = "gpt-3.5-turbo"
)

type GptRole string

func (m GptRole) String() string {
	return string(m)
}

const (
	GptRoleUser = "user"
)

type GPTRequestMessage struct {
	Role    GptRole `json:"role"`
	Content string  `json:"content"`
}

type GPTRequest struct {
	Model    Model `json:"model"`
	Messages []GPTRequestMessage
}

type GPTResponse struct {
	Id      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Usage   struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Choices []struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
		Index        int    `json:"index"`
	} `json:"choices"`
}

func Ping(key string) error {
	// Prepare the request body
	requestBody, err := json.Marshal(GPTRequest{
		Model: Model35,
		Messages: []GPTRequestMessage{
			{
				Role:    GptRoleUser,
				Content: "testing",
			},
		},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+key)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	// Ping accepts an authorized "bad-request", but if it's not an "Okay" or a "Bad request"
	// there's an issue
	if resp.StatusCode != 200 && resp.StatusCode != 400 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("got %q response status code", resp.Status)
		}

		return fmt.Errorf("Probably you did not insert the `OPENAI_API_KEY` key - got %q response, body: \n%s\n", resp.Status, string(body))
	}

	// Too many requests - no quota
	if resp.StatusCode != 433 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("got %q response status code", resp.Status)
		}

		return fmt.Errorf("Probably not a paid plan - got %q response, body: \n%s\n", resp.Status, string(body))
	}

	return nil
}

func Prompt(apikey string, prompt string) (response string, err error) {
	requestBody, err := json.Marshal(map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []interface{}{
			map[string]string{
				"role":    "user",
				"content": prompt,
			},
		},
		"temperature": 0.7,
	})

	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apikey)

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

	var gptResponse GPTResponse
	err = json.Unmarshal(body, &gptResponse)
	if err != nil {
		return "", err
	}

	// Assuming the first choice is the one we need
	if len(gptResponse.Choices) > 0 {
		return gptResponse.Choices[0].Message.Content, nil
	}

	return "", ErrNoResponse
}
