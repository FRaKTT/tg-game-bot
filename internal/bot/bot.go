package bot

import (
	"fmt"

	gamePkg "github.com/fraktt/tg-game-bot/internal/game"
	"github.com/fraktt/tg-game-bot/internal/storage"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

type bot struct {
	botAPI *tgbotapi.BotAPI
	game   *gamePkg.Game //todo: interface
}

type Interface interface {
	Run() error
}

func MustNew(apiKey string, storage storage.Interface) Interface {
	botAPI, err := tgbotapi.NewBotAPI(apiKey)
	if err != nil {
		logrus.Panicf("botAPI init: %v", err)
	}

	botAPI.Debug = true
	logrus.Infof("Authorized on account %s", botAPI.Self.UserName)

	game, err := gamePkg.New(storage)
	if err != nil {
		logrus.Panicf("create game: %v", err)
	}

	return &bot{
		botAPI: botAPI,
		game:   game,
	}
}

func (b *bot) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			logrus.Errorf("update.Message == nil")
			continue
		}

		logrus.Infof(logRecvMsg(update.Message.From, update.Message.Text))

		results := b.game.ProcessMessage(gamePkg.UserDTO{
			ID:        update.Message.From.ID,
			FirstName: update.Message.From.FirstName,
			LastName:  update.Message.From.LastName,
			UserName:  update.Message.From.UserName,
		}, update.Message.Text)
		results = collapseMessages(results)

		for _, r := range results {
			logrus.Infof(logSendMsg(update.Message.From, r))

			responseMsg := fillResponseMsg(update.Message.Chat.ID, r.Text, r.Buttons)
			if _, err := b.botAPI.Send(responseMsg); err != nil {
				return fmt.Errorf("send tg msg: %w", err)
			}
		}
	}

	return nil
}

// collapseMessages схлопывает результаты, чтобы они были валидны для отправки пользователям
// например, в сообщении всегда должен быть текст
func collapseMessages(results []gamePkg.ProcessMessageResult) []gamePkg.ProcessMessageResult {
	var newResults []gamePkg.ProcessMessageResult
	var currentResult gamePkg.ProcessMessageResult
	for _, r := range results {
		if r.Text != "" {
			if currentResult.Text != "" { // если в currentResult уже что-то есть - записываем в результат
				newResults = append(newResults, currentResult)
				currentResult = gamePkg.ProcessMessageResult{} // сброс
			}
			currentResult.Text = r.Text
		}
		if len(r.Buttons) != 0 {
			currentResult.Buttons = r.Buttons
		}
	}

	// записываем оставшееся в currentResult после последней итерации
	if currentResult.Text != "" {
		newResults = append(newResults, currentResult)
	}

	return newResults
}

// fillResponseMsg заполняет сообщение для отправки пользователю
func fillResponseMsg(chatID int64, text string, buttons []string) tgbotapi.MessageConfig {
	if text == "" {
		text = `ОШИБКА: text == ""`
	}
	responseMsg := tgbotapi.NewMessage(chatID, text)

	if len(buttons) != 0 {
		var keyboardButtons [][]tgbotapi.KeyboardButton
		for _, b := range buttons {
			keyboardButtons = append(keyboardButtons, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(b)})
		}
		responseMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(keyboardButtons...)

	} else {
		responseMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
	}

	return responseMsg
}

func logRecvMsg(user *tgbotapi.User, text string) string {
	return fmt.Sprintf("FROM user: [%d,%s,%s,%s] - msg: %q",
		user.ID, user.FirstName, user.LastName, user.UserName, text)
}

func logSendMsg(user *tgbotapi.User, r gamePkg.ProcessMessageResult) string {
	return fmt.Sprintf("TO user: [%d,%s,%s,%s] - response: %q, buttons: %v",
		user.ID, user.FirstName, user.LastName, user.UserName, r.Text, r.Buttons)
}
