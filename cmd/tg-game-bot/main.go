package main

import (
	"log"

	"github.com/fraktt/tg-game-bot/internal/config"
	botPkg "github.com/fraktt/tg-game-bot/internal/pkg/bot"
	fileStorage "github.com/fraktt/tg-game-bot/internal/pkg/storage/file"
	memoryStorage "github.com/fraktt/tg-game-bot/internal/pkg/storage/memory"
)

func main() {
	apiKey := config.GetAPIKey()

	_ = memoryStorage.New() // todo: выбор хранилища можно реализовать через аргументы или переменные среды
	storage, err := fileStorage.New("")
	if err != nil {
		log.Fatal(err)
	}

	bot := botPkg.MustNew(apiKey, storage)

	if err := bot.Run(); err != nil {
		log.Fatal(err)
	}
}
