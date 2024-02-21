# lint:
# 	golangci-lint run

test:
	go test ./...

build:
	go build -o bin/tg-game-bot ./cmd/tg-game-bot

run:
	go run ./cmd/tg-game-bot