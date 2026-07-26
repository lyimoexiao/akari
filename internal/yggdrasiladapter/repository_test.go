package yggdrasiladapter

import (
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Migrate_creates_yggdrasil_tables_from_adapter(t *testing.T) {
	// Given
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	// When
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	// Then
	for _, table := range []string{"yggdrasil_profiles", "yggdrasil_tokens"} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("missing table %s", table)
		}
	}
}
