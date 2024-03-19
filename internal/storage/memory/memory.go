package memory

import (
	"fmt"
	"sync"

	storagePkg "github.com/fraktt/tg-game-bot/internal/storage"
)

type storage struct {
	usersSteps map[int]int // map userID -> stepID
	mu         sync.RWMutex
}

func New() storagePkg.Interface {
	return &storage{
		usersSteps: make(map[int]int),
		mu:         sync.RWMutex{},
	}
}

func (s *storage) GetUserStep(userID int) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stepID, ok := s.usersSteps[userID]
	if !ok {
		return 0, fmt.Errorf("пользователь %d не найден: %w", userID, storagePkg.ErrUserNotFound)
	}
	return stepID, nil
}

func (s *storage) SaveUserStep(userID int, stepID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.usersSteps[userID] = stepID

	return nil
}
