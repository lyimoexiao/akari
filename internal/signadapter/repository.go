package signadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) LastSignAt(ctx context.Context, userID uint) (*time.Time, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("last_sign_at").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("user %d: not found", userID)
		}
		return nil, fmt.Errorf("last sign at: %w", err)
	}
	return user.LastSignAt, nil
}

func (r *Repository) RecordSign(ctx context.Context, userID uint, at time.Time) error {
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("last_sign_at", at).Error; err != nil {
		return fmt.Errorf("record sign: %w", err)
	}
	return nil
}

// Clock implements sign.Clock using time.Now.
type Clock struct{}

func (Clock) Now() time.Time { return time.Now() }