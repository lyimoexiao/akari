package textureadapter

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/gorm"
)

// TextureChecker implements closet.TextureRepository and closet.TextureDeleter.
type TextureChecker struct {
	db *gorm.DB
}

func NewTextureChecker(db *gorm.DB) *TextureChecker {
	return &TextureChecker{db: db}
}

func (c *TextureChecker) Exists(ctx context.Context, tid uint) (bool, error) {
	var count int64
	if err := c.db.WithContext(ctx).Model(&model.Texture{}).Where("tid = ?", tid).Count(&count).Error; err != nil {
		return false, fmt.Errorf("check texture exists: %w", err)
	}
	return count > 0, nil
}

func (c *TextureChecker) IsPublicOrUploader(ctx context.Context, tid, userID uint) (bool, error) {
	var texture model.Texture
	if err := c.db.WithContext(ctx).Select("public", "uploader").First(&texture, tid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, fmt.Errorf("check texture access: %w", err)
	}
	return texture.Public || texture.Uploader == userID, nil
}

func (c *TextureChecker) GetUploader(ctx context.Context, tid uint) (uint, error) {
	var texture model.Texture
	if err := c.db.WithContext(ctx).Select("uploader").First(&texture, tid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("纹理不存在")
		}
		return 0, fmt.Errorf("get uploader: %w", err)
	}
	return texture.Uploader, nil
}

// DeleteByID soft-deletes a texture by ID. callerID must match uploader unless force=true.
func (c *TextureChecker) DeleteByID(ctx context.Context, tid, callerID uint, force bool) error {
	var texture model.Texture
	if err := c.db.WithContext(ctx).Select("uploader").First(&texture, tid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return fmt.Errorf("find texture: %w", err)
	}
	if texture.Uploader != callerID && !force {
		return fmt.Errorf("无权删除此纹理")
	}
	return c.db.WithContext(ctx).Delete(&model.Texture{}, tid).Error
}
