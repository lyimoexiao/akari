package yggdrasiladapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func Migrate(db *gorm.DB) error {
	if err := db.AutoMigrate(&yggdrasil.YggdrasilProfile{}, &yggdrasil.YggdrasilToken{}); err != nil {
		return fmt.Errorf("migrate yggdrasil schema: %w", err)
	}
	return nil
}

func (r *Repository) FindUserByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, mapNotFound(err, yggdrasil.ErrUserNotFound)
}

func (r *Repository) FindUserByLogin(ctx context.Context, login string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).
		Where("email = ? OR username = ?", login, login).
		First(&user).Error
	return user, mapNotFound(err, yggdrasil.ErrUserNotFound)
}

func (r *Repository) CountValidTokens(ctx context.Context, email string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilToken{}).
		Where("user_email = ? AND status = ?", email, yggdrasil.TokenStatusValid).
		Count(&count).Error
	return count, err
}

func (r *Repository) RevokeOldestValidToken(ctx context.Context, email string) error {
	var token yggdrasil.YggdrasilToken
	err := r.db.WithContext(ctx).
		Where("user_email = ? AND status = ?", email, yggdrasil.TokenStatusValid).
		Order("issued_at ASC").
		First(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	return r.RevokeToken(ctx, token.AccessToken)
}

func (r *Repository) CreateToken(ctx context.Context, token *yggdrasil.YggdrasilToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *Repository) FindToken(ctx context.Context, accessToken string) (yggdrasil.YggdrasilToken, error) {
	var token yggdrasil.YggdrasilToken
	err := r.db.WithContext(ctx).
		Where("access_token = ?", accessToken).
		First(&token).Error
	return token, mapNotFound(err, yggdrasil.ErrInvalidToken)
}

func (r *Repository) RevokeToken(ctx context.Context, accessToken string) error {
	return r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilToken{}).
		Where("access_token = ?", accessToken).
		Update("status", yggdrasil.TokenStatusInvalid).Error
}

func (r *Repository) RevokeValidTokens(ctx context.Context, email string) error {
	return r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilToken{}).
		Where("user_email = ? AND status = ?", email, yggdrasil.TokenStatusValid).
		Update("status", yggdrasil.TokenStatusInvalid).Error
}

func (r *Repository) ProfilesForUser(ctx context.Context, email string) ([]yggdrasil.YggdrasilProfile, error) {
	var profiles []yggdrasil.YggdrasilProfile
	err := r.db.WithContext(ctx).Where("user_email = ?", email).Find(&profiles).Error
	return profiles, err
}

func (r *Repository) CreateProfile(ctx context.Context, profile *yggdrasil.YggdrasilProfile) error {
	return r.db.WithContext(ctx).Create(profile).Error
}

func (r *Repository) ProfileByUUID(ctx context.Context, uuid string) (yggdrasil.YggdrasilProfile, error) {
	var profile yggdrasil.YggdrasilProfile
	err := r.db.WithContext(ctx).Where("uuid = ?", uuid).First(&profile).Error
	return profile, mapNotFound(err, yggdrasil.ErrProfileNotFound)
}

func (r *Repository) ProfilesByNames(ctx context.Context, names []string) ([]yggdrasil.YggdrasilProfile, error) {
	var profiles []yggdrasil.YggdrasilProfile
	err := r.db.WithContext(ctx).Where("name IN ?", names).Find(&profiles).Error
	return profiles, err
}

func (r *Repository) LastLoginToken(ctx context.Context, email string) (yggdrasil.YggdrasilToken, error) {
	var token yggdrasil.YggdrasilToken
	err := r.db.WithContext(ctx).
		Where("user_email = ? AND login_ip != ''", email).
		Order("issued_at DESC").
		First(&token).Error
	return token, err
}

func mapNotFound(err, notFound error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return err
}
