package model

import (
	"slices"
	"time"
)

const (
	RoleSuperAdmin = "super_admin"
	RoleStaff      = "staff"
	RoleUser       = "user"
)

type Role struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	Name        string       `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Description string       `gorm:"size:255;not null;default:''" json:"description"`
	IsDefault   bool         `gorm:"index;not null;default:false" json:"is_default"`
	ParentID    *uint        `gorm:"index" json:"parent_id,omitempty"`
	Parent      *Role        `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Permissions []Permission `gorm:"many2many:role_permissions" json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Permission struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:128;not null" json:"name"`
	Object      string    `gorm:"uniqueIndex:idx_permission_rule;size:255;not null" json:"object"`
	Action      string    `gorm:"uniqueIndex:idx_permission_rule;size:16;not null" json:"action"`
	Description string    `gorm:"size:255;not null;default:''" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UserRole struct {
	UserID    uint      `gorm:"primaryKey" json:"user_id"`
	RoleID    uint      `gorm:"primaryKey" json:"role_id"`
	CreatedAt time.Time `json:"created_at"`
}

type RolePermission struct {
	RoleID       uint      `gorm:"primaryKey" json:"role_id"`
	PermissionID uint      `gorm:"primaryKey" json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}

func PrimaryRole(roles []Role) string {
	if len(roles) == 0 {
		return ""
	}
	names := make([]string, 0, len(roles))
	for _, role := range roles {
		if role.Name == RoleSuperAdmin {
			return RoleSuperAdmin
		}
		names = append(names, role.Name)
	}
	slices.Sort(names)
	return names[0]
}
