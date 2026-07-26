package authadapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *UserRepository) PrepareRegistration(ctx context.Context, email, username string) error {
	for _, identity := range []struct {
		column    string
		value     string
		duplicate error
	}{
		{"email", email, auth.ErrEmailAlreadyExists},
		{"username", username, auth.ErrUsernameAlreadyUsed},
	} {
		var existing model.User
		err := r.db.WithContext(ctx).Unscoped().
			Where(identity.column+" = ?", identity.value).
			First(&existing).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("check %s: %w", identity.column, err)
		}
		if !existing.DeletedAt.Valid {
			return identity.duplicate
		}
		if err := r.purge(ctx, existing); err != nil {
			return fmt.Errorf("purge stale %s: %w", identity.column, err)
		}
	}
	return nil
}

func (r *UserRepository) purge(ctx context.Context, user model.User) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Unscoped().Delete(&user).Error
	})
}

func (r *UserRepository) Create(ctx context.Context, user *model.User, role *model.Role) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if role == nil {
			return nil
		}
		return tx.Create(&model.UserRole{UserID: user.ID, RoleID: role.ID}).Error
	})
	if err == nil {
		return nil
	}
	if isUniqueConstraint(err, "username") {
		return auth.ErrUsernameAlreadyUsed
	}
	if isUniqueConstraint(err, "email") {
		return auth.ErrEmailAlreadyExists
	}
	return err
}

func (r *UserRepository) FindByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").
		Where("username = ? OR email = ?", login, login).
		First(&user).Error
	return user, mapUserNotFound(err)
}

func (r *UserRepository) FindByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Preload("Roles").First(&user, id).Error
	return user, mapUserNotFound(err)
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, mapUserNotFound(err)
}

func (r *UserRepository) MarkEmailVerified(ctx context.Context, id uint, verifiedAt time.Time) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Update("email_verified_at", verifiedAt).Error
}

func (r *UserRepository) UpdatePassword(ctx context.Context, id uint, password string) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", id).
		Update("password", password).Error
}

func mapUserNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return auth.ErrUserNotFound
	}
	return err
}

func isUniqueConstraint(err error, column string) bool {
	message := err.Error()
	return strings.Contains(message, "UNIQUE constraint failed") && strings.Contains(message, column) ||
		strings.Contains(message, "Duplicate entry") && strings.Contains(message, column)
}
