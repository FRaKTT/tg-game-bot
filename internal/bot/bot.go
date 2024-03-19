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
		logrus.Panic(fmt.Errorf("botAPI init: %w", err))
	}

	botAPI.Debug = true
	logrus.Printf("Authorized on account %s", botAPI.Self.UserName)

	game, err := gamePkg.New(storage)
	if err != nil {
		logrus.Panic(fmt.Errorf("create game: %w", err))
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
		switch {
		case update.Message != nil:
			results := b.game.ProcessMessage(gamePkg.UserDTO{
				ID:        update.Message.From.ID,
				FirstName: update.Message.From.FirstName,
				LastName:  update.Message.From.LastName,
				UserName:  update.Message.From.UserName,
			}, update.Message.Text)
			results = collapseMessages(results)

			for _, r := range results {
				responseMsg := tgbotapi.NewMessage(update.Message.Chat.ID, "<здесь должен быть текст>")
				if r.Text != "" {
					responseMsg.Text = r.Text
				}

				if r.Buttons != nil {
					var buttons [][]tgbotapi.KeyboardButton
					for _, b := range r.Buttons {
						buttons = append(buttons, []tgbotapi.KeyboardButton{tgbotapi.NewKeyboardButton(b)})
					}
					responseMsg.ReplyMarkup = tgbotapi.NewReplyKeyboard(buttons...)
				} else {
					responseMsg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
				}
				if _, err := b.botAPI.Send(responseMsg); err != nil {
					return fmt.Errorf("send tg msg: %w", err)
				}
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
