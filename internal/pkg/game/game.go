package game

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	"github.com/fraktt/tg-game-bot/internal/pkg/storage"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var ErrStepNotFound = fmt.Errorf("step not found")

type Game struct {
	storage storage.Interface
	steps   []Step
}

func New(storage storage.Interface) (*Game, error) {
	steps := hardcodedSteps
	if err := validateSteps(steps); err != nil {
		return nil, fmt.Errorf("ошибка валидации шагов: %v", err)
	}

	return &Game{
		storage: storage,
		steps:   steps,
	}, nil
}

type (
	UserDTO struct {
		ID        int64
		FirstName string
		LastName  string
		UserName  string
	}

	ProcessMessageResult struct {
		Text    string
		Buttons []string
	}
)

func (g *Game) ProcessMessage(user UserDTO, userMsg string) (res []ProcessMessageResult) {
	var errMsg string
	defer func() {
		if errMsg != "" {
			log.Printf(errMsg)
			res = append(res, ProcessMessageResult{
				Text: fmt.Sprintf("что-то пошло не так: %s", errMsg),
			})
		}
	}()

	// получаем текущий шаг пользователя
	stepID, err := g.storage.GetUserStep(int(user.ID))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			stepID = stepStart
		} else {
			errMsg = fmt.Sprintf("GetUserStep(userID=%v): %v", user.ID, err)
			return res
		}
	}
	step, err := g.getStepByID(stepID)
	if err != nil {
		if errors.Is(err, ErrStepNotFound) {
			stepID = stepStart
		} else {
			errMsg = fmt.Sprintf("getStepByID(stepID=%v): %v", stepID, err)
			return res
		}
	}

	// если это форк, то чекаем ответ пользователя
	switch s := step.(type) {
	case ForkStep:
		nextStepID, nextStepFound := s.NextSteps[userMsg]
		if !nextStepFound { // от пользователя пришло сообщение, не подходящее ни под один вариант
			if s.DefaultDoNothing { // задано ничего не делать в таких случаях
				return res
			}
			if s.DefaultNextStep == 0 { // дефолтное поведение не задано - остаёмся на текущем шаге
				res = append(res, ProcessMessageResult{
					Text:    "неизвестный вариант",
					Buttons: s.Variants, // текущую клавиатуру оставляем
				})
				return res
			}
			nextStepID = s.DefaultNextStep
		}
		stepID = nextStepID // переходим на следующий шаг

	case ForkStepOpenQuestion:
		var answerIsRight bool
		for _, rightAnswer := range s.RightAnswers {
			if normalize(userMsg) == normalize(rightAnswer) {
				stepID = s.NextStepRight
				answerIsRight = true
			}
		}
		if !answerIsRight {
			stepID = s.NextStepWrong
		}
	}

	log.Printf("ProcessMessage: [user %d,%s,%s,%s] - (step %v(%d)) msg: %q",
		user.ID, user.FirstName, user.LastName, user.UserName, step.GetName(), step.GetID(), userMsg)

	// проходим по линейным шагам
	linearRes, stopStepID, err := g.processStepsChain(stepID)
	res = append(res, linearRes...)
	if err != nil {
		errMsg = err.Error()
		return res
	}

	g.storage.SaveUserStep(int(user.ID), stopStepID) //todo: сделать через коллбек, чтоб переходить в другой стейт уже после успешной отправки сообщения

	return res
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// processStepsChain идёт по непрерывной последовательности шагов
func (g *Game) processStepsChain(stepID int) (res []ProcessMessageResult, stopStepID int, _ error) {
	for {
		step, err := g.getStepByID(stepID)
		if err != nil {
			return res, stepID, fmt.Errorf("getStepByID(stepID=%v): %w", stepID, err)
		}

		switch v := step.(type) {
		case LinearStep:
			res = append(res, ProcessMessageResult{
				Text: v.Message,
			})
			stepID = v.NextStepID
			if stepID == 0 { // конец игры
				return res, stepID, nil
			}
			continue

		case LinearStepRandomMsg:
			randomIdx := rand.Intn(len(v.RandomMessages))
			res = append(res, ProcessMessageResult{
				Text: v.RandomMessages[randomIdx],
			})
			stepID = v.NextStepID
			if stepID == 0 { // конец игры
				return res, stepID, nil
			}
			continue

		case ForkStep:
			res = append(res, ProcessMessageResult{
				Buttons: v.Variants,
			})
			return res, stepID, nil // выходим, дойдя до форка

		case ForkStepOpenQuestion:
			return res, stepID, nil // выходим, дойдя до форка

		default:
			return res, stepID, fmt.Errorf("неизвестный тип следующего шага (stepID=%d): %T", stepID, v)
		}
	}
}

func (g *Game) getStepByID(stepID int) (Step, error) {
	for _, s := range g.steps {
		if s.GetID() == stepID {
			return s, nil
		}
	}
	return nil, fmt.Errorf("шаг с id=%v не найден: %w", stepID, ErrStepNotFound)
}
