package admin

import (
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

// UserItem is the user row returned in list responses.
type UserItem struct {
	ID              uint       `json:"id"`
	Username        string     `json:"username"`
	Email           string     `json:"email"`
	Role            string     `json:"role"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
}

// ListUsersResp wraps a paginated user list.
type ListUsersResp struct {
	Items      []UserItem `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

func toUserItem(u *model.User) UserItem {
	return UserItem{
		ID:              u.ID,
		Username:        u.Username,
		Email:           u.Email,
		Role:            u.Role,
		EmailVerifiedAt: u.EmailVerifiedAt,
		CreatedAt:       u.CreatedAt,
	}
}
