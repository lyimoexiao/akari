package database

import (
	"fmt"

	pkgdb "github.com/lyimoexiao/akari/pkg/database"
	"gorm.io/gorm"
)

// New opens a database connection, runs project-specific migrations, and
// configures the connection pool. Delegates connection logic to pkg/database.
func New(cfg *pkgdb.Config) (*gorm.DB, error) {
	return pkgdb.New(cfg, func(db *gorm.DB) error {
		if err := Migrate(db); err != nil {
			return fmt.Errorf("auto migrate: %w", err)
		}
		return nil
	})
}

// Close gracefuly shuts down the database connection.
func Close(db *gorm.DB) error {
	return pkgdb.Close(db)
}
