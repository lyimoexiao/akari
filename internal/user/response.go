package user

import (
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/pagination"
)

type Item struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ListResp is the paginated list response for users.
// Deprecated: use pagination.Paged[Item] directly.
type ListResp = pagination.Paged[Item]

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
