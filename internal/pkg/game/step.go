package game

import "fmt"

type Step interface {
	GetID() int
	GetName() string
}

type LinearStep struct {
	ID         int
	Name       string
	NextStepID int
	Message    string
}

func (s LinearStep) GetName() string {
	return s.Name
}
func (s LinearStep) GetID() int {
	return s.ID
}

type LinearStepRandomMsg struct {
	ID             int
	Name           string
	NextStepID     int
	RandomMessages []string
}

func (s LinearStepRandomMsg) GetName() string {
	return s.Name
}
func (s LinearStepRandomMsg) GetID() int {
	return s.ID
}

type ForkStep struct {
	ID               int //todo: мб использовать строковые названия?
	Name             string
	Variants         []string       // упорядоченные варианты ответа
	NextSteps        map[string]int // map variant -> next step id
	DefaultNextStep  int            // если ответ не соответствует ни одному из вариантов - переход на следующий шаг
	DefaultDoNothing bool           // если ответ не соответствует ни одному из вариантов - ничего не делаем
}

func (s ForkStep) GetName() string {
	return s.Name
}
func (s ForkStep) GetID() int {
	return s.ID
}

type ForkStepOpenQuestion struct {
	ID            int
	Name          string
	RightAnswers  []string // список - на случай, если возможно несколько правильных вариантов
	NextStepRight int
	NextStepWrong int
}

func (s ForkStepOpenQuestion) GetName() string {
	return s.Name
}
func (s ForkStepOpenQuestion) GetID() int {
	return s.ID
}

func validateSteps(steps []Step) error {
	for _, s := range steps {
		switch fs := s.(type) {
		case ForkStep:
			if err := validateForkStep(fs); err != nil {
				return fmt.Errorf("ошибка валидации шага %v: %w", fs.ID, err)
			}
		}
	}
	return nil
}

func validateForkStep(fs ForkStep) error {
	// проверка уникальности вариантов в списке
	tmpMap := map[string]struct{}{}
	for _, v := range fs.Variants {
		if _, ok := tmpMap[v]; ok {
			return fmt.Errorf("найден повторящийся элемент %v", v)
		}
		tmpMap[v] = struct{}{}
	}

	// проверка, что варианты в списке соответствуют вариантам в мапе
	for _, v := range fs.Variants {
		if _, ok := fs.NextSteps[v]; !ok {
			return fmt.Errorf("в мапе нет ключа %v", v)
		}
	}

	return nil
}
