package bot

import (
	gamePkg "github.com/fraktt/tg-game-bot/internal/game"
	"github.com/fraktt/tg-game-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	storage storage.Interface
	game    *gamePkg.Game //todo: interface

	roles userRoles
}

func MustNew(apiKey string, storage storage.Interface) *Bot {
	botAPI, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		logrus.Panicf("botAPI init: %v", err)
	}

	botAPI.Debug = true
	logrus.Infof("Authorized on account %s", botAPI.Self.UserName)

	game, err := gamePkg.New(storage, gamePkg.DemoGameSteps)
	if err != nil {
		logrus.Panicf("create game: %v", err)
	}

	return &Bot{
		botAPI:  botAPI,
		storage: storage,
		game:    game,
	}
}
