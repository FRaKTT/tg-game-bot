package storage

import "fmt"

var ErrUserNotFound = fmt.Errorf("user not found")

type Interface interface {
	GetUserStep(userID int) (int, error)
	SaveUserStep(userID int, stepID int) error

	ReadAll() (string, error)
}
