package game

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"

	"github.com/fraktt/tg-game-bot/internal/storage"
	"github.com/sirupsen/logrus"
)

var ErrStepNotFound = fmt.Errorf("step not found")

type Game struct {
	storage   storage.Interface
	steps     []Step
	startStep int
}

func New(storage storage.Interface, steps []Step) (*Game, error) {
	if len(steps) == 0 {
		return nil, fmt.Errorf("пустой список шагов")
	}
	startStep := steps[0].GetID()

	if err := validateSteps(steps); err != nil {
		return nil, fmt.Errorf("валидация шагов: %w", err)
	}

	return &Game{
		storage:   storage,
		steps:     steps,
		startStep: startStep,
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

func (g *Game) ProcessMessage(user UserDTO, userMsg string) (res []ProcessMessageResult) { //nolint:gocognit,nonamedreturns,lll //todo
	var errMsg string
	defer func() {
		if errMsg != "" {
			logrus.Error(errMsg)
			res = append(res, ProcessMessageResult{
				Text: fmt.Sprintf("что-то пошло не так: %s", errMsg),
			})
		}
	}()

	// получаем текущий шаг пользователя
	stepID, err := g.storage.GetUserStep(int(user.ID))
	if err != nil {
		if errors.Is(err, storage.ErrUserNotFound) {
			stepID = g.startStep
		} else {
			errMsg = fmt.Sprintf("GetUserStep(userID=%v): %v", user.ID, err)
			return res
		}
	}

	// если пользователь отправляет команду /start, нужно начать обрабатывать его заново
	if userMsg == "/start" {
		stepID = g.startStep
	}

	step, err := g.getStepByID(stepID)
	if err != nil {
		if errors.Is(err, ErrStepNotFound) {
			stepID = g.startStep
			step, _ = g.getStepByID(stepID) // получаем step заново
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

	logrus.Infof("ProcessMessage: [user %d,%s,%s,%s] - (step %v(%d)) msg: %q",
		user.ID, user.FirstName, user.LastName, user.UserName, step.GetName(), step.GetID(), userMsg)

	// проходим по линейным шагам
	linearRes, stopStepID, err := g.processStepsChain(stepID)
	res = append(res, linearRes...)
	if err != nil {
		errMsg = err.Error()
		return res
	}

	//todo: сделать через коллбек, чтоб переходить в другой стейт уже после успешной отправки сообщения
	if err = g.storage.SaveUserStep(int(user.ID), stopStepID); err != nil {
		errMsg = err.Error()
		return res
	}

	return res
}

func normalize(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// processStepsChain идёт по непрерывной последовательности шагов
func (g *Game) processStepsChain(stepID int) (res []ProcessMessageResult, stopStepID int, _ error) { //nolint:nonamedreturns,lll // todo
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
			randomIdx := rand.Intn(len(v.RandomMessages)) //nolint:gosec // no security requirements
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
