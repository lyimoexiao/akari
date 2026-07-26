package routeradapter

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/pkg/cache"
	"gorm.io/gorm"
)

var (
	errDatabaseUnavailable = errors.New("database unavailable")
	errCacheUnavailable    = errors.New("cache unavailable")
)

type HealthChecker struct {
	db    *gorm.DB
	cache cache.Cache
}

func NewHealthChecker(db *gorm.DB, cacheBackend cache.Cache) *HealthChecker {
	return &HealthChecker{db: db, cache: cacheBackend}
}

func (checker *HealthChecker) Check(ctx context.Context) error {
	if checker.db == nil {
		return errDatabaseUnavailable
	}
	sqlDB, err := checker.db.DB()
	if err != nil {
		return fmt.Errorf("open database handle: %w", err)
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	if checker.cache == nil {
		return errCacheUnavailable
	}
	if err := checker.cache.Ping(ctx); err != nil {
		return fmt.Errorf("ping cache: %w", err)
	}
	return nil
}
