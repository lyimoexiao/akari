package textureadapter

import (
	"context"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/skinlib"
	"gorm.io/gorm"
)

// Repository implements skinlib.Repository with GORM.
type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, record *skinlib.TextureRecord) error {
	m := toModel(record)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return fmt.Errorf("create texture: %w", err)
	}
	record.TID = m.TID
	record.UploadAt = m.UploadAt.Format("2006-01-02T15:04:05+08:00")
	return nil
}

func (r *Repository) FindByID(ctx context.Context, tid uint) (*skinlib.TextureRecord, error) {
	var m model.Texture
	if err := r.db.WithContext(ctx).First(&m, tid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, skinlib.ErrTextureNotFound
		}
		return nil, fmt.Errorf("find texture by id: %w", err)
	}
	return fromModel(&m), nil
}

func (r *Repository) FindByHash(ctx context.Context, hash string) (*skinlib.TextureRecord, error) {
	var m model.Texture
	if err := r.db.WithContext(ctx).Where("hash = ?", hash).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find texture by hash: %w", err)
	}
	return fromModel(&m), nil
}

func (r *Repository) FindByHashAndUploader(ctx context.Context, hash string, uploader uint) (*skinlib.TextureRecord, error) {
	var m model.Texture
	if err := r.db.WithContext(ctx).Where("hash = ? AND uploader = ? AND deleted_at IS NULL", hash, uploader).First(&m).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("find texture by hash and uploader: %w", err)
	}
	return fromModel(&m), nil
}

func (r *Repository) List(ctx context.Context, query skinlib.ListQuery) ([]skinlib.TextureRecord, int64, error) {
	dbQuery := r.db.WithContext(ctx).Model(&model.Texture{})

	// Filter by type
	if query.Type == "skin" {
		dbQuery = dbQuery.Where("type IN ?", []string{"steve", "alex"})
	} else if query.Type == "cape" {
		dbQuery = dbQuery.Where("type = ?", "cape")
	}

	// Public only
	if query.PublicOnly {
		dbQuery = dbQuery.Where("public = ?", true)
	}

	// Filter by uploader
	if query.Uploader > 0 {
		dbQuery = dbQuery.Where("uploader = ?", query.Uploader)
	}

	// Search keyword
	if query.Search != "" {
		dbQuery = dbQuery.Where("name LIKE ?", "%"+query.Search+"%")
	}

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("count textures: %w", err)
	}

	switch query.Order {
	case "likes":
		dbQuery = dbQuery.Order("likes DESC, upload_at DESC")
	default:
		dbQuery = dbQuery.Order("upload_at DESC")
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

	var models []model.Texture
	if err := dbQuery.Offset(offset).Limit(pageSize).Find(&models).Error; err != nil {
		return nil, 0, fmt.Errorf("list textures: %w", err)
	}

	records := make([]skinlib.TextureRecord, len(models))
	for i, m := range models {
		records[i] = *fromModel(&m)
	}

	return records, total, nil
}

func (r *Repository) Update(ctx context.Context, tid uint, name string, public bool) error {
	result := r.db.WithContext(ctx).Model(&model.Texture{}).
		Where("t_id = ?", tid).
		Updates(map[string]any{
			"name":   name,
			"public": public,
		})
	if result.Error != nil {
		return fmt.Errorf("update texture: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return skinlib.ErrTextureNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, tid uint) error {
	result := r.db.WithContext(ctx).Delete(&model.Texture{}, tid)
	if result.Error != nil {
		return fmt.Errorf("delete texture: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return skinlib.ErrTextureNotFound
	}
	return nil
}

func (r *Repository) IncrementLikes(ctx context.Context, tid uint) error {
	return r.db.WithContext(ctx).Model(&model.Texture{}).
		Where("t_id = ?", tid).
		Update("likes", gorm.Expr("likes + 1")).Error
}

func (r *Repository) DecrementLikes(ctx context.Context, tid uint) error {
	return r.db.WithContext(ctx).Model(&model.Texture{}).
		Where("t_id = ? AND likes > 0", tid).
		Update("likes", gorm.Expr("likes - 1")).Error
}

func (r *Repository) CountByUploader(ctx context.Context, uploader uint) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.Texture{}).
		Where("uploader = ?", uploader).Count(&count).Error; err != nil {
		return 0, fmt.Errorf("count by uploader: %w", err)
	}
	return count, nil
}

func toModel(r *skinlib.TextureRecord) *model.Texture {
	return &model.Texture{
		Name:     r.Name,
		Type:     r.Type,
		Hash:     r.Hash,
		URL:      r.URL,
		Size:     r.Size,
		Uploader: r.Uploader,
		Public:   r.Public,
		Likes:    r.Likes,
	}
}

func fromModel(m *model.Texture) *skinlib.TextureRecord {
	return &skinlib.TextureRecord{
		TID:      m.TID,
		Name:     m.Name,
		Type:     m.Type,
		Hash:     m.Hash,
		URL:      m.URL,
		Size:     m.Size,
		Uploader: m.Uploader,
		Public:   m.Public,
		Likes:    m.Likes,
		UploadAt: m.UploadAt.Format("2006-01-02T15:04:05+08:00"),
	}
}