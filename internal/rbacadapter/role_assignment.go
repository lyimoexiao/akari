package rbacadapter

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
)

func (repository *RoleRepository) ReplaceUserRole(ctx context.Context, userID, roleID uint) error {
	db := repository.db.WithContext(ctx)
	if err := db.Where("user_id = ?", userID).Delete(&model.UserRole{}).Error; err != nil {
		return fmt.Errorf("clear user roles: %w", err)
	}
	if err := db.Create(&model.UserRole{UserID: userID, RoleID: roleID}).Error; err != nil {
		return fmt.Errorf("assign user role: %w", err)
	}
	return nil
}
