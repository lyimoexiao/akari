package requestlogadapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/requestlog"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (repository *Repository) GetByRequestID(ctx context.Context, requestID string) (model.RequestLog, error) {
	var record model.RequestLog
	if err := repository.db.WithContext(ctx).Where("request_id = ?", requestID).
		Order("id DESC").First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return model.RequestLog{}, requestlog.ErrNotFound
		}
		return model.RequestLog{}, fmt.Errorf("query request log: %w", err)
	}
	return record, nil
}

func (repository *Repository) Save(ctx context.Context, record model.RequestLog) error {
	if err := repository.db.WithContext(ctx).Create(&record).Error; err != nil {
		return fmt.Errorf("persist request log: %w", err)
	}
	return nil
}
