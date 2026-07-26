// Package database provides multi-driver database connection management.
//
// It supports MySQL, PostgreSQL, and SQLite via GORM. The connection is
// configured through a self-contained Config struct, and an optional
// migration function can be supplied for schema initialization.
package database

import (
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// MigrationFunc is an optional callback invoked after the database connection
// is established, before connection pooling is configured.
// Return an error to abort initialization.
type MigrationFunc func(*gorm.DB) error

// New opens a database connection, runs the optional migration function, and
// configures the connection pool.
// Pass nil for migrator if no schema initialization is needed.
func New(cfg *Config, migrator MigrationFunc) (*gorm.DB, error) {
	ensureDir(cfg.Type, cfg.Name)

	dialector, err := openDriver(cfg)
	if err != nil {
		return nil, err
	}

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	// Run optional migration callback before pool config
	if migrator != nil {
		if err := migrator(db); err != nil {
			return nil, err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	return db, nil
}

// Close gracefuly shuts down the underlying sql.DB.
func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB for close: %w", err)
	}
	return sqlDB.Close()
}

func ensureDir(dbType, dsn string) {
	if dbType == "sqlite" {
		_ = os.MkdirAll(filepath.Dir(dsn), 0o755)
	}
}

func openDriver(cfg *Config) (gorm.Dialector, error) {
	// If DSN is set explicitly, use it directly
	if cfg.DSN != "" {
		return openDSN(cfg.Type, cfg.DSN)
	}

	switch cfg.Type {
	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
		return mysql.Open(dsn), nil

	case "postgres":
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name)
		return postgres.Open(dsn), nil

	case "sqlite":
		dsn := cfg.Name
		if dsn == "" {
			dsn = "akari.db"
		}
		return sqlite.Open(dsn), nil

	default:
		return nil, fmt.Errorf("unsupported database type: %s", cfg.Type)
	}
}

func openDSN(dbType, dsn string) (gorm.Dialector, error) {
	switch dbType {
	case "mysql":
		return mysql.Open(dsn), nil
	case "postgres":
		return postgres.Open(dsn), nil
	case "sqlite":
		return sqlite.Open(dsn), nil
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbType)
	}
}