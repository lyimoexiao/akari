package rbacadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/permission"
	"gorm.io/gorm"
)

func (manager *Manager) EnforceUser(ctx context.Context, check permission.Check) (bool, string, error) {
	manager.mutationMu.RLock()
	defer manager.mutationMu.RUnlock()

	roles, err := manager.userRoles(ctx, check.UserID)
	if err != nil {
		return false, "", err
	}
	primaryRole := model.PrimaryRole(roles)
	for _, storedRole := range roles {
		allowed, err := manager.enforcer.Load().Enforce(storedRole.Name, check.Object, check.Action)
		if err != nil {
			return false, primaryRole, fmt.Errorf("enforce casbin policy: %w", err)
		}
		if allowed {
			return true, primaryRole, nil
		}
	}
	return false, primaryRole, nil
}

func (manager *Manager) IdentifiersForUser(ctx context.Context, userID uint) ([]string, error) {
	manager.mutationMu.RLock()
	defer manager.mutationMu.RUnlock()

	roles, err := manager.userRoles(ctx, userID)
	if err != nil {
		return nil, err
	}
	effectiveRoles := make(map[string]struct{}, len(roles))
	for _, storedRole := range roles {
		effectiveRoles[storedRole.Name] = struct{}{}
		inherited, err := manager.enforcer.Load().GetImplicitRolesForUser(storedRole.Name)
		if err != nil {
			return nil, fmt.Errorf("resolve inherited roles: %w", err)
		}
		for _, name := range inherited {
			effectiveRoles[name] = struct{}{}
		}
	}
	roleNames := make([]string, 0, len(effectiveRoles))
	for name := range effectiveRoles {
		roleNames = append(roleNames, name)
	}
	identifiers := make([]string, 0)
	if len(roleNames) == 0 {
		return identifiers, nil
	}
	if err := manager.db.WithContext(ctx).Table("role_permissions").
		Distinct("permissions.name").
		Joins("JOIN roles ON roles.id = role_permissions.role_id").
		Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
		Where("roles.name IN ?", roleNames).
		Order("permissions.name").
		Pluck("permissions.name", &identifiers).Error; err != nil {
		return nil, fmt.Errorf("load permission identifiers: %w", err)
	}
	return identifiers, nil
}

func (manager *Manager) userRoles(ctx context.Context, userID uint) ([]model.Role, error) {
	var storedUser model.User
	if err := manager.db.WithContext(ctx).Preload("Roles").First(&storedUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, permission.ErrUserNotFound
		}
		return nil, fmt.Errorf("load user roles: %w", err)
	}
	return storedUser.Roles, nil
}
