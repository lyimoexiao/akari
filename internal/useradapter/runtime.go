package useradapter

import (
	"time"

	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

type Clock struct{}

func (Clock) Now() time.Time {
	return time.Now()
}

type PasswordHasher struct{}

func (PasswordHasher) Hash(password string) (string, error) {
	return bcrypt.HashPassword(password)
}
