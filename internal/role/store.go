package role

import (
	"context"
	"slices"
	"strings"

	"github.com/lyimoexiao/akari/internal/model"
)

func itemByID(ctx context.Context, repository Repository, roleID uint) (Item, error) {
	storedRole, err := repository.RoleByID(ctx, roleID)
	if err != nil {
		return Item{}, err
	}
	return toItem(storedRole), nil
}

func toItem(storedRole model.Role) Item {
	permissions := make([]string, len(storedRole.Permissions))
	for index := range storedRole.Permissions {
		permissions[index] = storedRole.Permissions[index].Name
	}
	slices.Sort(permissions)
	return Item{
		ID: storedRole.ID, Name: storedRole.Name, Description: storedRole.Description,
		IsDefault: storedRole.IsDefault, Permissions: permissions,
	}
}

func requireSuperAdmin(ctx context.Context, repository Repository, callerID uint) error {
	allowed, err := repository.HasUserRole(ctx, callerID, model.RoleSuperAdmin)
	if err != nil {
		return err
	}
	if !allowed {
		return ErrCannotModifySuperAdmin
	}
	return nil
}

func resolvePermissions(ctx context.Context, repository Repository, identifiers []string) ([]model.Permission, error) {
	unique := make(map[string]struct{}, len(identifiers))
	for _, identifier := range identifiers {
		identifier = strings.TrimSpace(identifier)
		if identifier == "" {
			return nil, ErrInvalidPermission
		}
		unique[identifier] = struct{}{}
	}
	if len(unique) == 0 {
		return []model.Permission{}, nil
	}
	names := make([]string, 0, len(unique))
	for name := range unique {
		names = append(names, name)
	}
	permissions, err := repository.PermissionsByName(ctx, names)
	if err != nil {
		return nil, err
	}
	if len(permissions) != len(names) {
		return nil, ErrInvalidPermission
	}
	return permissions, nil
}
