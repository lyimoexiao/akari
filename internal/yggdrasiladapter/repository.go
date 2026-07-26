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

func (r *Repository) UpdateProfileSkin(ctx context.Context, profileID uint, textureID *uint) error {
	return r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilProfile{}).
		Where("id = ?", profileID).
		Update("texture_skin_id", textureID).Error
}

func (r *Repository) UpdateProfileCape(ctx context.Context, profileID uint, textureID *uint) error {
	return r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilProfile{}).
		Where("id = ?", profileID).
		Update("texture_cape_id", textureID).Error
}

func (r *Repository) TextureByID(ctx context.Context, tid uint) (*model.Texture, error) {
	var texture model.Texture
	if err := r.db.WithContext(ctx).First(&texture, tid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &texture, nil
}

func (r *Repository) UpdateProfileModel(ctx context.Context, profileID uint, model string) error {
	return r.db.WithContext(ctx).Model(&yggdrasil.YggdrasilProfile{}).
		Where("id = ?", profileID).
		Update("model", model).Error
}

// ClearTextureFromProfile clears the given texture TID from the user's
// Yggdrasil profile(s) — both skin and cape — so the profile no longer
// references a texture that has been removed from the closet.
func (r *Repository) ClearTextureFromProfile(ctx context.Context, userID, textureTID uint) error {
	user, err := r.FindUserByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("find user: %w", err)
	}
	profiles, err := r.ProfilesForUser(ctx, user.Email)
	if err != nil {
		return fmt.Errorf("find profiles: %w", err)
	}
	for _, p := range profiles {
		if p.TextureSkinID != nil && *p.TextureSkinID == textureTID {
			if err := r.UpdateProfileSkin(ctx, p.ID, nil); err != nil {
				return fmt.Errorf("clear profile skin: %w", err)
			}
		}
		if p.TextureCapeID != nil && *p.TextureCapeID == textureTID {
			if err := r.UpdateProfileCape(ctx, p.ID, nil); err != nil {
				return fmt.Errorf("clear profile cape: %w", err)
			}
		}
	}
	return nil
}

func mapNotFound(err, notFound error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return notFound
	}
	return err
}
