package model

import (
	"time"

	"gorm.io/gorm"
)

// Role constants
const (
	RoleSuperAdmin = "super_admin"
	RoleStaff      = "staff"
	RoleUser       = "user"
)

// ValidRoles returns all valid role strings.
func ValidRoles() []string {
	return []string{RoleSuperAdmin, RoleStaff, RoleUser}
}

// IsValidRole checks whether the given role is valid.
func IsValidRole(role string) bool {
	for _, r := range ValidRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// User represents a registered user in the system.
type User struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Username        string         `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email           string         `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password        string         `gorm:"size:255;not null" json:"-"` // never expose in JSON
	Role            string         `gorm:"size:32;not null;default:user" json:"role"`
	EmailVerifiedAt *time.Time     `gorm:"index" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// IsEmailVerified returns true if the user has verified their email.
func (u *User) IsEmailVerified() bool {
	return u.EmailVerifiedAt != nil
}

// IsSuperAdmin returns true if the user is a super admin.
func (u *User) IsSuperAdmin() bool {
	return u.Role == RoleSuperAdmin
}

// IsStaff returns true if the user has staff role or above.
func (u *User) IsStaff() bool {
	return u.Role == RoleStaff || u.Role == RoleSuperAdmin
}
