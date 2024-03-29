package file

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	storagePkg "github.com/fraktt/tg-game-bot/internal/storage"
)

const (
	defaultStorageFilename = "bot_data.json"
	storageFilePermissions = 0666
)

type storage struct {
	filename string
	mu       sync.RWMutex
}

func New(filename string) (storagePkg.Interface, error) {
	if filename == "" {
		filename = defaultStorageFilename
	}

	f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE, storageFilePermissions)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", filename, err)
	}
	defer f.Close()

	return &storage{
		filename: filename,
		mu:       sync.RWMutex{},
	}, nil
}

func (s *storage) GetUserStep(userID int) (int, error) {
	usersSteps, err := s.getMapFromFile()
	if err != nil {
		return 0, err
	}

	stepID, ok := usersSteps[userID]
	if !ok {
		return 0, fmt.Errorf("пользователь %d не найден: %w", userID, storagePkg.ErrUserNotFound)
	}
	return stepID, nil
}

func (s *storage) SaveUserStep(userID int, stepID int) error {
	usersSteps, err := s.getMapFromFile()
	if err != nil {
		return err
	}

	usersSteps[userID] = stepID
	updatedData, err := json.Marshal(usersSteps)
	if err != nil {
		return fmt.Errorf("marshal updated users steps: %w", err)
	}

	s.mu.Lock()
	os.WriteFile(s.filename, updatedData, storageFilePermissions)
	s.mu.Unlock()

	return nil
}

func (s *storage) ReadAll() (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filename)
	if err != nil {
		return "", fmt.Errorf("read file %q: %w", s.filename, err)
	}

	return string(data), nil
}

func (s *storage) getMapFromFile() (map[int]int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filename)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", s.filename, err)
	}

	usersSteps := make(map[int]int)
	if len(data) != 0 {
		if err := json.Unmarshal(data, &usersSteps); err != nil {
			return nil, fmt.Errorf("unmarshal file: %w", err)
		}
	}

	return usersSteps, nil
}
