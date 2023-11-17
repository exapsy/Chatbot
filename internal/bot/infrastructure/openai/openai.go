package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GPTResponse struct {
	Choices []struct {
		Text string `json:"text"`
	} `json:"choices"`
}

func Ping(key string) error {
	// Prepare the request body
	requestBody, err := json.Marshal(map[string]string{
		"model":  "text-davinci-004", // or any other GPT-4 model you want to use
		"prompt": "I'm testing to see if you work:)",
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/engines/text-davinci-004/completions", bytes.NewBuffer(requestBody))
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

		return fmt.Errorf("got %q response, body: \n%s\n", resp.Status, string(body))
	}

	return nil
}
