package main

import (
	"connectly-interview/internal/bot/app"
	"fmt"
)

func main() {
	var err error
	fmt.Printf("initializing bot...\n")
	bot, err := bot_app.New(
		bot_app.WithHttpServer("localhost:8080"),
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
