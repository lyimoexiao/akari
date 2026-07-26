package useradapter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/pagination"
	"github.com/lyimoexiao/akari/internal/user"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) List(ctx context.Context, query user.ListQuery) ([]model.User, int64, error) {
	dbQuery := repository.db.WithContext(ctx).Model(&model.User{})
	if query.Search != "" {
		like := "%" + query.Search + "%"
		dbQuery = dbQuery.Where("username LIKE ? OR email LIKE ?", like, like)
	}
	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count users: %w", err)
	}
	var users []model.User
	if err := dbQuery.Preload("Roles").Order("created_at DESC").
		Scopes(pagination.ApplyOffsetLimit(query.Paging)).Find(&users).Error; err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	return users, total, nil
}

func (repository *Repository) FindByID(ctx context.Context, userID uint) (model.User, error) {
	var storedUser model.User
	if err := repository.db.WithContext(ctx).Preload("Roles").First(&storedUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.User{}, user.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("find target user: %w", err)
	}
	return storedUser, nil
}

func (repository *Repository) HasRole(ctx context.Context, userID uint, roleName string) (bool, error) {
	var count int64
	if err := repository.db.WithContext(ctx).Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ? AND roles.name = ?", userID, roleName).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check caller role: %w", err)
	}
	return count > 0, nil
}

func (repository *Repository) VerifyEmail(ctx context.Context, userID uint, verifiedAt time.Time) error {
	if err := repository.db.WithContext(ctx).Model(&model.User{ID: userID}).
		Update("email_verified_at", verifiedAt).Error; err != nil {
		return fmt.Errorf("update email verification: %w", err)
	}
	return nil
}

func (repository *Repository) UpdatePassword(ctx context.Context, userID uint, password string) error {
	if err := repository.db.WithContext(ctx).Model(&model.User{ID: userID}).
		Update("password", password).Error; err != nil {
		return fmt.Errorf("persist password: %w", err)
	}
	return nil
}

func (repository *Repository) Delete(ctx context.Context, userID uint) error {
	if err := repository.db.WithContext(ctx).Delete(&model.User{ID: userID}).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
