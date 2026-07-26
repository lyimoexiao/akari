package rbacadapter

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/role"
)

func (repository *RoleRepository) CreateRole(ctx context.Context, storedRole *model.Role) error {
	if err := repository.db.WithContext(ctx).Create(storedRole).Error; err != nil {
		return fmt.Errorf("create role: %w", err)
	}
	return nil
}

func (repository *RoleRepository) ReplaceRolePermissions(ctx context.Context, roleID uint, permissions []model.Permission) error {
	db := repository.db.WithContext(ctx)
	if err := db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return fmt.Errorf("clear role permissions: %w", err)
	}
	for _, storedPermission := range permissions {
		relation := model.RolePermission{RoleID: roleID, PermissionID: storedPermission.ID}
		if err := db.Create(&relation).Error; err != nil {
			return fmt.Errorf("assign role permission: %w", err)
		}
	}
	return nil
}

func (repository *RoleRepository) UpdateRole(ctx context.Context, roleID uint, changes role.RoleChanges) error {
	if err := repository.db.WithContext(ctx).Model(&model.Role{ID: roleID}).Updates(map[string]any{
		"name": changes.Name, "description": changes.Description,
	}).Error; err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}
