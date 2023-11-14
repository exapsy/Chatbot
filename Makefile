BOT_CMD=./cmd/bot/main.go

.PHONY: start-browser-ui
start-browser-ui:
	@echo "Opening browser UI"

.PHONY: build-bot
build-bot:
	@go build $(BOT_CMD)

.PHONY: build-and-run-executable
build-and-run-executable: build
	@echo "Running bot executable"

.PHONY: run-bot
run-bot:
	@go run $(BOT_CMD)

.PHONY: start
start: start-browser-ui build-and-run-executable
	@echo "Starting bot and UI"