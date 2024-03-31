package main

import (
	"os"

	botPkg "github.com/fraktt/tg-game-bot/internal/bot"
	rolesPkg "github.com/fraktt/tg-game-bot/internal/bot/roles"
	"github.com/fraktt/tg-game-bot/internal/logging"
	fileStorage "github.com/fraktt/tg-game-bot/internal/storage/file"
	memoryStorage "github.com/fraktt/tg-game-bot/internal/storage/memory"
	"github.com/sirupsen/logrus"
)

func main() {
	logging.SetupLogging()

	apiKey := os.Getenv("TGBOTAPIKEY")

	_ = memoryStorage.New() // todo: select storage via arguments or environment variables
	storage, err := fileStorage.New("")
	if err != nil {
		logrus.Fatal(err)
	}

	ur := rolesPkg.CreateUserRoles(
	// todo: fill admin and participants IDs with CreateUserRoles options
	// rolesPkg.WithAdminID(),
	// rolesPkg.WithParticipantIDs(),
	)

	bot := botPkg.MustNew(apiKey, storage, ur)

	logrus.Info("Launching bot")
	logrus.Fatal(bot.Run())
}
