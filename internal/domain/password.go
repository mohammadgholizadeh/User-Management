package domain

import (
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

const minLength = 8

type Password string

func (p Password) Validate() error {
	if len(p) < minLength {
		zap.L().Warn("Password is too short",
			zap.String("entity", "password"),
			zap.Int("min_required_length", minLength),
			zap.Int("actual_length", len(p)),
		)
		return fmt.Errorf("password must be at least %d characters long", minLength)
	}
	return nil
}

func (p Password) Hash() ([]byte, error) {
	return bcrypt.GenerateFromPassword([]byte(p), bcrypt.DefaultCost)
}

func (p Password) CompareWithHashedPassword(hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}
