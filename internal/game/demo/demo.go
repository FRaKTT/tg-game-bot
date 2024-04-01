package demo

import gamePkg "github.com/fraktt/tg-game-bot/internal/game"

const (
	stepStart = iota + 1
	stepGreeting
	stepIntro

	stepCosmonautQuestion
	stepCosmonautVariants
	stepCosmonautBadAnswer

	stepSpiderQuestion
	stepSpiderVariants
	stepSpiderBadAnswer
	stepSpiderRightAnswer

	stepSnowQuestion
	stepSnowVariants
	stepSnowBadAnswer
	stepSnowRightAnswer

	stepVaultOpened
	stepCongrats
	stepFinish
)

// DemoGameSteps - демо-игра
var DemoGameSteps = []gamePkg.Step{ //nolint:gochecknoglobals // demo
	// Intro
	gamePkg.LinearStep{
		ID:         stepStart,
		Name:       "Start",
		Message:    "Привет!",
		NextStepID: stepGreeting,
	},
	gamePkg.ForkStep{ // остановка перед началом
		ID:       stepGreeting,
		Name:     "Greeting",
		Variants: []string{"Привет!"},
		NextSteps: map[string]int{
			"Привет!": stepIntro,
		},
		DefaultNextStep: stepIntro,
	},
	gamePkg.LinearStep{
		ID:         stepIntro,
		Name:       "Intro",
		Message:    "ДЕМО ИГРА",
		NextStepID: stepCosmonautQuestion,
	},

	gamePkg.LinearStep{
		ID:         stepCosmonautQuestion,
		Name:       "CosmonautQuestion",
		Message:    "Первый вопрос: назови фамилию первого космонавта",
		NextStepID: stepCosmonautVariants,
	},
	gamePkg.ForkStep{
		ID:       stepCosmonautVariants,
		Name:     "CosmonautVariants",
		Variants: []string{"Терешкова", "Гагарин", "Леонов", "Титов"},
		NextSteps: map[string]int{
			"Терешкова": stepCosmonautBadAnswer,
			"Гагарин":   stepSpiderQuestion,
			"Леонов":    stepCosmonautBadAnswer,
			"Титов":     stepCosmonautBadAnswer,
		},
	},
	gamePkg.LinearStep{
		ID:         stepCosmonautBadAnswer,
		Name:       "CosmonautBadAnswer",
		Message:    "Попробуй ещё раз",
		NextStepID: stepCosmonautVariants,
	},

	gamePkg.LinearStep{
		ID:         stepSpiderQuestion,
		Name:       "SpiderQuestion",
		Message:    "Сколько лап паука?",
		NextStepID: stepSpiderVariants,
	},
	gamePkg.ForkStep{
		ID:       stepSpiderVariants,
		Name:     "SpiderVariants",
		Variants: []string{"Четыре", "Шесть", "Восемь", "Шестнадцать"},
		NextSteps: map[string]int{
			"Четыре":      stepSpiderBadAnswer,
			"Шесть":       stepSpiderBadAnswer,
			"Восемь":      stepSpiderRightAnswer,
			"Шестнадцать": stepSpiderBadAnswer,
		},
	},
	gamePkg.LinearStep{
		ID:         stepSpiderBadAnswer,
		Name:       "SpiderBadAnswer",
		Message:    "Неверно, попробуй ещё раз",
		NextStepID: stepSpiderVariants,
	},
	gamePkg.LinearStep{
		ID:         stepSpiderRightAnswer,
		Name:       "SpiderRightAnswer",
		Message:    "🕸",
		NextStepID: stepSnowQuestion,
	},

	gamePkg.LinearStep{
		ID:         stepSnowQuestion,
		Name:       "WaistQuestion",
		Message:    "На дворе горой, а в избе водой. Что это?",
		NextStepID: stepSnowVariants,
	},
	gamePkg.ForkStepOpenQuestion{
		ID:            stepSnowVariants,
		Name:          "WaistVariants",
		RightAnswers:  []string{"Снег", "Cytu"}, // для удобства ;)
		NextStepRight: stepSnowRightAnswer,
		NextStepWrong: stepSnowBadAnswer,
	},
	gamePkg.LinearStep{
		ID:         stepSnowBadAnswer,
		Name:       "WaistBadAnswer",
		Message:    "Неа",
		NextStepID: stepSnowVariants,
	},
	gamePkg.LinearStep{
		ID:         stepSnowRightAnswer,
		Name:       "WaistRightAnswer",
		Message:    "Точно!",
		NextStepID: stepCongrats,
	},

	gamePkg.LinearStepRandomMsg{
		ID:   stepCongrats,
		Name: "Congrats",
		RandomMessages: []string{
			"Ура! Игра пройдена",
			"Поздравляю! Ты ответил(а) на все вопросы",
		},
		NextStepID: stepFinish,
	},
	gamePkg.ForkStep{
		ID:       stepFinish,
		Name:     "Finish",
		Variants: []string{"Начать заново 🔁"},
		NextSteps: map[string]int{
			"Начать заново 🔁": stepIntro,
		},
		DefaultDoNothing: true,
	},
}
