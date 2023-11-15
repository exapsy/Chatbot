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

.PHONY: compile-docker-bot
compile-docker-bot:
	docker build -t exapsy/connectly-interview

.PHONY: compile-docker-web
compile-docker-web:

.PHONY: compile-docker
compile-docker: compile-docker-bot compile-docker-web

.PHONY: push-docker-web
push-docker-bot:

.PHONY: push-docker-web
push-docker-web:

.PHONY: push-docker
push-docker:

.PHONY: deploy
deploy: compile-docker push-docker
	@echo "deploying ..."
