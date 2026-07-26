package auth

import (
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

// AuthResp is returned on successful login or registration.
type AuthResp struct {
	Token string   `json:"token"`
	User  UserResp `json:"user"`
}

// UserResp is the public user info returned to clients.
type UserResp struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

func toUserResp(u *model.User) UserResp {
	return UserResp{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		Role:            model.PrimaryRole(u.Roles),
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}
