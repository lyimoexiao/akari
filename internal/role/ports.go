package role

import (
	"context"

	"github.com/lyimoexiao/akari/internal/model"
)

type Repository interface {
	HasUserRole(context.Context, uint, string) (bool, error)
	RoleNameExists(context.Context, string, uint) (bool, error)
	PermissionsByName(context.Context, []string) ([]model.Permission, error)
	CreateRole(context.Context, *model.Role) error
	ReplaceRolePermissions(context.Context, uint, []model.Permission) error
	RoleByID(context.Context, uint) (model.Role, error)
	UpdateRole(context.Context, uint, RoleChanges) error
	RoleAssignmentCount(context.Context, uint) (int64, error)
	ClearChildParents(context.Context, uint) error
	DeleteRole(context.Context, uint) error
	ReplaceDefaultRole(context.Context, uint) error
	RoleByName(context.Context, string) (model.Role, error)
	UserExists(context.Context, uint) (bool, error)
	ReplaceUserRole(context.Context, uint, uint) error
	RegistrationRole(context.Context) (*model.Role, error)
	ListRoles(context.Context) ([]model.Role, error)
}

type RoleChanges struct {
	Name        string
	Description string
}

type PolicyUpdater interface {
	UpdatePolicy(context.Context, func(Repository) error) error
}

type Dependencies struct {
	Repository Repository
	Policies   PolicyUpdater
}
