package main

import (
	"connectly-interview/internal/bot/app"
	"fmt"
	"os"
)

func main() {
	var err error

	fmt.Printf("\033[0;30mⓘ Make sure to incorporate OpenAI key on the %q environment key/value\n"+
		"Used for generating responses\n\033[0;0m",
		"OPENAI_API_KEY",
	)

	openaiKey := os.Getenv("OPENAI_API_KEY")

	fmt.Printf("initializing bot...\n")
	bot, err := bot_app.New(
		bot_app.WithOpenAiKey(openaiKey),
		bot_app.WithHttpServer("localhost:8080"),
		bot_app.WithNoKafka(),
	)
	if err != nil {
		panic(err)
	}

	fmt.Printf("starting bot...\n")
	err = bot.Start()
	if err != nil {
		panic(err)
		return
	}
}
