package bot

import (
	rolesPkg "github.com/fraktt/tg-game-bot/internal/bot/roles"
	gamePkg "github.com/fraktt/tg-game-bot/internal/game"
	"github.com/fraktt/tg-game-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type Bot struct {
	botAPI  *tgbotapi.BotAPI
	storage storage.Interface
	game    *gamePkg.Game //todo: interface

	roles       rolesPkg.UserRoles
	adminChatID int64
}

func MustNew(apiKey string, storage storage.Interface, userRoles rolesPkg.UserRoles, gameSteps []gamePkg.Step) *Bot {
	botAPI, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		logrus.Panicf("botAPI init: %v", err)
	}

	botAPI.Debug = true
	logrus.Infof("Authorized on account %s", botAPI.Self.UserName)

	game, err := gamePkg.New(storage, gameSteps)
	if err != nil {
		logrus.Panicf("create game: %v", err)
	}

	return &Bot{
		botAPI:  botAPI,
		storage: storage,
		game:    game,

		roles: userRoles,
	}
}
