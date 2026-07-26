package rbacadapter

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
)

func (repository *RoleRepository) ClearChildParents(ctx context.Context, roleID uint) error {
	if err := repository.db.WithContext(ctx).Model(&model.Role{}).
		Where("parent_id = ?", roleID).Update("parent_id", nil).Error; err != nil {
		return fmt.Errorf("clear child role parents: %w", err)
	}
	return nil
}

func (repository *RoleRepository) DeleteRole(ctx context.Context, roleID uint) error {
	db := repository.db.WithContext(ctx)
	if err := db.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
		return fmt.Errorf("delete role permissions: %w", err)
	}
	if err := db.Delete(&model.Role{ID: roleID}).Error; err != nil {
		return fmt.Errorf("delete role: %w", err)
	}
	return nil
}

func (repository *RoleRepository) ReplaceDefaultRole(ctx context.Context, roleID uint) error {
	db := repository.db.WithContext(ctx)
	if err := db.Model(&model.Role{}).Where("is_default = ?", true).
		Update("is_default", false).Error; err != nil {
		return fmt.Errorf("clear default role: %w", err)
	}
	if err := db.Model(&model.Role{ID: roleID}).Update("is_default", true).Error; err != nil {
		return fmt.Errorf("set default role: %w", err)
	}
	return nil
}
