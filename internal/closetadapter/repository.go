package closetadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/internal/closet"
	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/gorm"
)

// Repository implements closet.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Add(ctx context.Context, userID, textureTID uint, itemName string) error {
	entry := model.UserCloset{
		UserID:     userID,
		TextureTID: textureTID,
		ItemName:   itemName,
	}
	return r.db.WithContext(ctx).Create(&entry).Error
}

func (r *Repository) Remove(ctx context.Context, userID, textureTID uint) error {
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND texture_t_id = ?", userID, textureTID).
		Delete(&model.UserCloset{})
	if result.Error != nil {
		return fmt.Errorf("remove from closet: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return closet.ErrClosetItemNotFound
	}
	return nil
}

func (r *Repository) Rename(ctx context.Context, userID, textureTID uint, newName string) error {
	result := r.db.WithContext(ctx).
		Model(&model.UserCloset{}).
		Where("user_id = ? AND texture_t_id = ?", userID, textureTID).
		Update("item_name", newName)
	if result.Error != nil {
		return fmt.Errorf("rename closet item: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return closet.ErrClosetItemNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context, userID uint, query closet.ListQuery) ([]closet.ClosetItem, int64, error) {
	// Build count query
	countQuery := r.db.WithContext(ctx).
		Model(&model.UserCloset{}).
		Joins("LEFT JOIN textures ON textures.t_id = user_closet.texture_t_id").
		Where("user_closet.user_id = ?", userID)

	// Filter by type
	if query.Type == "skin" {
		countQuery = countQuery.Where("textures.type IN ?", []string{"steve", "alex"})
	} else if query.Type == "cape" {
		countQuery = countQuery.Where("textures.type = ?", "cape")
	}

	// Search keyword
	if query.Search != "" {
		countQuery = countQuery.Where("user_closet.item_name LIKE ?", "%"+query.Search+"%")
	}

	var total int64
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count closet: %w", err)
	}

	page := query.Pagination.Page
	pageSize := query.Pagination.PageSize
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Query with explicit join
	type closetRow struct {
		TextureTID uint      `gorm:"column:texture_t_id"`
		ItemName   string    `gorm:"column:item_name"`
		CreatedAt  time.Time `gorm:"column:created_at"`
		// Texture fields
		TexName     string `gorm:"column:tex_name"`
		TexType     string `gorm:"column:tex_type"`
		TexHash     string `gorm:"column:tex_hash"`
		TexURL      string `gorm:"column:tex_url"`
		TexSize     int64  `gorm:"column:tex_size"`
		TexPublic   bool   `gorm:"column:tex_public"`
		TexLikes    int64  `gorm:"column:tex_likes"`
		TexUploader uint   `gorm:"column:tex_uploader"`
	}

	var rows []closetRow

	// Build query dynamically
	dataQuery := r.db.WithContext(ctx).
		Table("user_closet").
		Select(`user_closet.texture_t_id, user_closet.item_name, user_closet.created_at,
			textures.name AS tex_name, textures.type AS tex_type,
			textures.hash AS tex_hash, textures.size AS tex_size,
			textures.public AS tex_public, textures.likes AS tex_likes,
			textures.uploader AS tex_uploader`).
		Joins("LEFT JOIN textures ON textures.t_id = user_closet.texture_t_id").
		Where("user_closet.user_id = ?", userID)

	if query.Type == "skin" {
		dataQuery = dataQuery.Where("textures.type IN ?", []string{"steve", "alex"})
	} else if query.Type == "cape" {
		dataQuery = dataQuery.Where("textures.type = ?", "cape")
	}

	if query.Search != "" {
		dataQuery = dataQuery.Where("user_closet.item_name LIKE ?", "%"+query.Search+"%")
	}

	if err := dataQuery.
		Order("user_closet.created_at DESC").
		Offset(offset).Limit(pageSize).
		Find(&rows).Error; err != nil {
		return nil, 0, fmt.Errorf("list closet: %w", err)
	}

	items := make([]closet.ClosetItem, len(rows))
	for i, r := range rows {
		items[i] = closet.ClosetItem{
			TextureTID: r.TextureTID,
			ItemName:   r.ItemName,
			CreatedAt:  r.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
			Texture: &struct {
				Name     string `json:"name"`
				Type     string `json:"type"`
				Hash     string `json:"hash"`
				URL      string `json:"url"`
				Size     int64  `json:"size"`
				Public   bool   `json:"public"`
				Likes    int64  `json:"likes"`
				Uploader uint   `json:"uploader"`
			}{
				Name:     r.TexName,
				Type:     r.TexType,
				Hash:     r.TexHash,
				URL:      r.TexURL,
				Size:     r.TexSize,
				Public:   r.TexPublic,
				Likes:    r.TexLikes,
				Uploader: r.TexUploader,
			},
		}
	}

	return items, total, nil
}

func (r *Repository) Exists(ctx context.Context, userID, textureTID uint) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&model.UserCloset{}).
		Where("user_id = ? AND texture_t_id = ?", userID, textureTID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *Repository) AllTextureIDs(ctx context.Context, userID uint) ([]uint, error) {
	var ids []uint
	if err := r.db.WithContext(ctx).
		Model(&model.UserCloset{}).
		Where("user_id = ?", userID).
		Pluck("texture_t_id", &ids).Error; err != nil {
		return nil, fmt.Errorf("get closet ids: %w", err)
	}
	return ids, nil
}

// RemoveByTextureExceptUploader removes all closet entries for the given texture TID
// except the uploader's own entry. Used when a texture is set to private.
func (r *Repository) RemoveByTextureExceptUploader(ctx context.Context, textureTID, uploaderID uint) error {
	return r.db.WithContext(ctx).
		Where("texture_t_id = ? AND user_id != ?", textureTID, uploaderID).
		Delete(&model.UserCloset{}).Error
}
