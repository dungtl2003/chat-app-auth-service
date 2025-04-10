package password

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type PasswordManager interface {
	Hash(password string) (string, error)
	IsCorrectPassword(password, hashedPassword string) bool
}

type BcryptPasswordManager struct {
	cost int
}

func NewBcryptPasswordManager(cost int) (BcryptPasswordManager, error) {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		return BcryptPasswordManager{}, fmt.Errorf("invalid bcrypt cost: %d", cost)
	}
	return BcryptPasswordManager{
		cost: cost,
	}, nil
}

func (b BcryptPasswordManager) Hash(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", err
	}

	return string(hashedPassword), nil
}

func (b BcryptPasswordManager) IsCorrectPassword(password, hashedPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
