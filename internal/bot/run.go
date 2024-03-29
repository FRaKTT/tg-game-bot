package bot

import (
	"fmt"
	"os"
	"strings"

	gamePkg "github.com/fraktt/tg-game-bot/internal/game"
	"github.com/fraktt/tg-game-bot/internal/logging"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/sirupsen/logrus"
)

func (b *Bot) Run() error {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.botAPI.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			logrus.Errorf("update.Message == nil")
			continue
		}

		switch b.getUserRole(update.Message.From) {
		case participantRole:
			if err := b.handleParticipantMessage(update.Message); err != nil {
				return fmt.Errorf("handle participant message: %w", err)
			}
		case adminRole:
			if err := b.handleAdminMessage(update.Message); err != nil {
				return fmt.Errorf("handle admin message: %w", err)
			}
		default:
			if err := b.handleUnknownUserMessage(update.Message); err != nil {
				return fmt.Errorf("handle unknown user message: %w", err)
			}
		}
	}

	return nil
}

func (b *Bot) handleParticipantMessage(msg *tgbotapi.Message) error {
	logrus.Infof(logRecvMsg(msg.From, msg.Text))

	results := b.game.ProcessMessage(userToDTO(msg.From), msg.Text)
	results = collapseMessages(results)

	for _, r := range results {
		logrus.Infof(logSendMsg(msg.From, r))

		responseMsg := fillResponseMsg(msg.Chat.ID, r.Text, r.Buttons)
		if _, err := b.botAPI.Send(responseMsg); err != nil {
			return fmt.Errorf("send tg msg: %w", err)
		}
	}

	return nil
}

const (
	adminButtonStorage = "data_storage"
	adminButtonLogs    = "logs"
)

var (
	adminChatID  int64
	adminButtons = []string{adminButtonStorage, adminButtonLogs}
)

func (b *Bot) handleAdminMessage(msg *tgbotapi.Message) error {
	adminChatID = msg.Chat.ID // сохраняем для асинхронных уведомлений

	switch msg.Text {
	case "/start": // вывести доступные кнопки
		responseMsg := fillResponseMsg(msg.Chat.ID, "hello", adminButtons)
		if _, err := b.botAPI.Send(responseMsg); err != nil {
			return fmt.Errorf("send tg msg: %w", err)
		}

	case adminButtonStorage: // прочитать содержимое хранилища
		var text string
		text, err := b.storage.ReadAll()
		if err != nil {
			text = fmt.Sprintf("error reading storage: %v", err)
		}
		if text == "" {
			text = "STORAGE IS EMPTY"
		}
		responseMsg := fillResponseMsg(msg.Chat.ID, text, adminButtons)
		if _, err := b.botAPI.Send(responseMsg); err != nil {
			return fmt.Errorf("send tg msg: %w", err)
		}

	case adminButtonLogs: // прочитать логи
		var text string
		data, err := os.ReadFile(logging.LogFile())
		if err != nil {
			text = fmt.Sprintf("read log file %q: %v", logging.LogFile(), err)
		} else {
			text = string(data)
			if text == "" {
				text = "LOGS ARE EMPTY"
			}
		}

		textPieces := textRowsToPieces(text, 20)
		for _, tp := range textPieces {
			msg := fillResponseMsg(msg.Chat.ID, tp, adminButtons)
			if _, err := b.botAPI.Send(msg); err != nil {
				if strings.Contains(err.Error(), "Bad Request: message is too long") {
					logrus.Errorf("msg too long: %v", err)
					continue
				}
				return fmt.Errorf("send tg msg: %w", err)
			}
		}
	}

	return nil
}

func (b *Bot) handleUnknownUserMessage(msg *tgbotapi.Message) error {
	unknownUserMsg := "[MSG FROM UNKNOWN USER] " + logRecvMsg(msg.From, msg.Text)
	logrus.Warn(unknownUserMsg)

	responseMsg := fillResponseMsg(msg.Chat.ID, "Sorry, I don't know you(", nil)
	if _, err := b.botAPI.Send(responseMsg); err != nil {
		return fmt.Errorf("send tg msg: %w", err)
	}

	alertToAdminMsg := fillResponseMsg(adminChatID, unknownUserMsg, nil)
	if _, err := b.botAPI.Send(alertToAdminMsg); err != nil {
		return fmt.Errorf("send tg msg: %w", err)
	}

	return nil
}

// textRowsToPieces разбивает строку на блоки с ограниченным кол-вом строк
// нужно, т.к. телеграм имеет лимит на размер отправляемого сообщения
// todo: правильнее разбивать по кол-ву символов (4096, check) и/или уменьшать размер при получении ошибки MESSAGE_TOO_LONG (https://core.telegram.org/method/messages.sendMessage)
func textRowsToPieces(text string, nRowsInPiece int) []string {
	out := []string{}

	rows := strings.Split(text, "\n")
	for i := 0; i < len(rows); {
		upperLimit := min(i+nRowsInPiece, len(rows))
		out = append(out, strings.Join(rows[i:upperLimit], "\n"))
		i = upperLimit
	}

	return out
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
		text = `ERROR: text == ""`
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

func userToDTO(u *tgbotapi.User) gamePkg.UserDTO {
	return gamePkg.UserDTO{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		UserName:  u.UserName,
	}
}

func logRecvMsg(u *tgbotapi.User, text string) string {
	return fmt.Sprintf("FROM user: %s - msg: %q", userLogString(u), text)
}

func logSendMsg(u *tgbotapi.User, r gamePkg.ProcessMessageResult) string {
	return fmt.Sprintf("TO user: %s - response: %q, buttons: %v", userLogString(u), r.Text, r.Buttons)
}

func userLogString(u *tgbotapi.User) string {
	return fmt.Sprintf("[%d,%s,%s,%s]", u.ID, u.FirstName, u.LastName, u.UserName)
}
