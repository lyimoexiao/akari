package model

import (
	"time"

	"gorm.io/gorm"
)

// User represents a registered user in the system.
type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Username        string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email           string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password        string         `gorm:"size:255;not null" json:"-"` // never expose in JSON
	Roles           []Role         `gorm:"many2many:user_roles" json:"roles,omitempty"`
	EmailVerifiedAt *time.Time     `gorm:"index" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// IsEmailVerified returns true if the user has verified their email.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}
