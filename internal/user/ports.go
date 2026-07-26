package user

import (
	"context"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

type ListQuery struct {
	Search   string
	Page     int
	PageSize int
}

type Repository interface {
	List(context.Context, ListQuery) ([]model.User, int64, error)
	FindByID(context.Context, uint) (model.User, error)
	HasRole(context.Context, uint, string) (bool, error)
	VerifyEmail(context.Context, uint, time.Time) error
	UpdatePassword(context.Context, uint, string) error
	Delete(context.Context, uint) error
}

type Clock interface {
	Now() time.Time
}

type PasswordHasher interface {
	Hash(string) (string, error)
}

type Dependencies struct {
	Repository Repository
	Clock      Clock
	Hasher     PasswordHasher
}
