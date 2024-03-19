package main

import (
	"os"

	botPkg "github.com/fraktt/tg-game-bot/internal/bot"
	"github.com/fraktt/tg-game-bot/internal/logging"
	fileStorage "github.com/fraktt/tg-game-bot/internal/storage/file"
	memoryStorage "github.com/fraktt/tg-game-bot/internal/storage/memory"
	"github.com/sirupsen/logrus"
)

func main() {
	logging.SetupLogging()

	apiKey := os.Getenv("TGBOTAPIKEY")

	_ = memoryStorage.New() // todo: выбор хранилища можно реализовать через аргументы или переменные среды
	storage, err := fileStorage.New("")
	if err != nil {
		logrus.Fatal(err)
	}

	bot := botPkg.MustNew(apiKey, storage)
	logrus.Info("Launching bot")
	logrus.Fatal(bot.Run())
}
