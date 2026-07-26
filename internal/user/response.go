package user

import (
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

type Item struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

type ListResp struct {
	Items      []Item `json:"items"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

func toItem(user *model.User) Item {
	return Item{
		ID:              user.ID,
		Username:        user.Username,
		Email:           user.Email,
		Role:            model.PrimaryRole(user.Roles),
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
	}
}
