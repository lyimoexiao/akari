package database

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/lyimoexiao/akari/internal/config"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func ensureDir(dbType, dsn string) {
	if dbType == "sqlite" {
		os.MkdirAll(filepath.Dir(dsn), 0o755)
	}
}

func New(cfg *config.DatabaseConfig) (*gorm.DB, error) {
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

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("get underlying sql.DB: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	if err := Migrate(db); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

func Close(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("get sql.DB for close: %w", err)
	}
	return sqlDB.Close()
}

func openDriver(cfg *config.DatabaseConfig) (gorm.Dialector, error) {
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
