package rbacadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/role"
	"gorm.io/gorm"
)

func (repository *RoleRepository) HasUserRole(ctx context.Context, userID uint, roleName string) (bool, error) {
	var count int64
	if err := repository.db.WithContext(ctx).Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, roleName).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check user role: %w", err)
	}
	return count > 0, nil
}

func (repository *RoleRepository) RoleNameExists(ctx context.Context, name string, excludeID uint) (bool, error) {
	query := repository.db.WithContext(ctx).Model(&model.Role{}).Where("name = ?", name)
	if excludeID != 0 {
		query = query.Where("id <> ?", excludeID)
	}
	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, fmt.Errorf("check role name: %w", err)
	}
	return count > 0, nil
}

func (repository *RoleRepository) PermissionsByName(ctx context.Context, names []string) ([]model.Permission, error) {
	var permissions []model.Permission
	if err := repository.db.WithContext(ctx).Where("name IN ?", names).Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("load permissions: %w", err)
	}
	return permissions, nil
}

func (repository *RoleRepository) RoleByID(ctx context.Context, roleID uint) (model.Role, error) {
	var storedRole model.Role
	if err := repository.db.WithContext(ctx).Preload("Permissions").First(&storedRole, roleID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Role{}, role.ErrRoleNotFound
		}
		return model.Role{}, fmt.Errorf("find role: %w", err)
	}
	return storedRole, nil
}

func (repository *RoleRepository) RoleAssignmentCount(ctx context.Context, roleID uint) (int64, error) {
	var count int64
	if err := repository.db.WithContext(ctx).Model(&model.UserRole{}).
		Where("role_id = ?", roleID).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count role assignments: %w", err)
	}
	return count, nil
}

func (repository *RoleRepository) RoleByName(ctx context.Context, name string) (model.Role, error) {
	var storedRole model.Role
	if err := repository.db.WithContext(ctx).Where("name = ?", name).First(&storedRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.Role{}, role.ErrInvalidRole
		}
		return model.Role{}, fmt.Errorf("find role: %w", err)
	}
	return storedRole, nil
}

func (repository *RoleRepository) UserExists(ctx context.Context, userID uint) (bool, error) {
	var count int64
	if err := repository.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("find user: %w", err)
	}
	return count > 0, nil
}

func (repository *RoleRepository) RegistrationRole(ctx context.Context) (*model.Role, error) {
	var storedRole model.Role
	if err := repository.db.WithContext(ctx).Where("is_default = ?", true).
		Order("id").First(&storedRole).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("find registration role: %w", err)
	}
	return &storedRole, nil
}

func (repository *RoleRepository) ListRoles(ctx context.Context) ([]model.Role, error) {
	var roles []model.Role
	if err := repository.db.WithContext(ctx).Preload("Permissions").Order("id").Find(&roles).Error; err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}
