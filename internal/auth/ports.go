package auth

import (
	"context"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

type UserRepository interface {
	Count(context.Context) (int64, error)
	PrepareRegistration(context.Context, string, string) error
	Create(context.Context, *model.User, *model.Role) error
	FindByLogin(context.Context, string) (model.User, error)
	FindByID(context.Context, uint) (model.User, error)
	FindByEmail(context.Context, string) (model.User, error)
	MarkEmailVerified(context.Context, uint, time.Time) error
	UpdatePassword(context.Context, uint, string) error
}

type RoleFinder interface {
	FindByName(context.Context, string) (model.Role, error)
	FindRegistrationRole(context.Context) (*model.Role, error)
}

type TokenClaims struct {
	UserID    uint
	Username  string
	Role      string
	ExpiresAt time.Time
}

type TokenManager interface {
	GenerateToken(uint, string, string) (string, error)
	ValidateToken(string) (TokenClaims, error)
}

type TokenStore interface {
	SaveVerificationToken(context.Context, uint, string, time.Duration) error
	VerificationUser(context.Context, uint, string) (uint, error)
	DeleteVerificationToken(context.Context, uint, string) error
	SavePasswordResetToken(context.Context, string, uint, time.Duration) error
	PasswordResetUser(context.Context, string) (uint, error)
	DeletePasswordResetToken(context.Context, string) error
	BlacklistToken(context.Context, string, time.Duration) error
	IsTokenBlacklisted(context.Context, string) bool
}

type EmailSender interface {
	Send([]string, string, string) error
}

type Settings struct {
	RegistrationEnabled      bool
	EmailVerificationEnabled bool
	PasswordResetEnabled     bool
	VerifyEmailTokenTTL      time.Duration
	PasswordResetTokenTTL    time.Duration
}

type Dependencies struct {
	Users    UserRepository
	Roles    RoleFinder
	Tokens   TokenManager
	Store    TokenStore
	Mailer   EmailSender
	Settings Settings
}
