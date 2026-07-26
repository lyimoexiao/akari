package permission

import "errors"

var ErrUserNotFound = errors.New("用户不存在")

type Rule struct {
	Role   string `json:"role"`
	Object string `json:"object"`
	Action string `json:"action"`
}

type RoleInheritance struct {
	Role   string `json:"role"`
	Parent string `json:"parent"`
}

type Snapshot struct {
	Roles       []string          `json:"roles"`
	Rules       []Rule            `json:"rules"`
	Inheritance []RoleInheritance `json:"inheritance"`
}

type Check struct {
	UserID uint
	Object string
	Action string
}

type IdentifierList struct {
	Permissions []string `json:"permissions"`
}
