package database

import (
	"testing"

	"github.com/lyimoexiao/akari/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type legacyUser struct {
	ID       uint   `gorm:"primaryKey"`
	Username string `gorm:"uniqueIndex;size:64;not null"`
	Email    string `gorm:"uniqueIndex;size:255;not null"`
	Password string `gorm:"size:255;not null"`
	Role     string `gorm:"size:32;not null;default:user"`
}

func (legacyUser) TableName() string {
	return "users"
}

func Test_Migrate_preserves_legacy_user_roles_and_removes_role_column(t *testing.T) {
	// Given
	db := openMigrationTestDB(t)
	if err := db.AutoMigrate(&legacyUser{}); err != nil {
		t.Fatalf("create legacy users table: %v", err)
	}
	users := []legacyUser{
		{Username: "root", Email: "root@example.com", Password: "hash", Role: model.RoleSuperAdmin},
		{Username: "operator", Email: "operator@example.com", Password: "hash", Role: model.RoleStaff},
		{Username: "member", Email: "member@example.com", Password: "hash", Role: model.RoleUser},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("seed legacy users: %v", err)
	}

	// When
	if err := Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}

	// Then
	for _, table := range []any{&model.Role{}, &model.Permission{}, &model.UserRole{}, &model.RolePermission{}} {
		if !db.Migrator().HasTable(table) {
			t.Fatalf("missing migrated table for %T", table)
		}
	}
	if db.Migrator().HasColumn("users", "role") {
		t.Fatal("legacy users.role column still exists")
	}
	for _, indexName := range []string{"idx_users_username", "idx_users_email"} {
		if !db.Migrator().HasIndex(&model.User{}, indexName) {
			t.Fatalf("missing user uniqueness index %s", indexName)
		}
	}

	var assignments []struct {
		Username string
		Role     string
	}
	if err := db.Table("users").
		Select("users.username, roles.name AS role").
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Order("users.id").
		Scan(&assignments).Error; err != nil {
		t.Fatalf("read migrated assignments: %v", err)
	}
	want := []string{model.RoleSuperAdmin, model.RoleStaff, model.RoleUser}
	if len(assignments) != len(want) {
		t.Fatalf("assignments = %d, want %d", len(assignments), len(want))
	}
	for index, assignment := range assignments {
		if assignment.Role != want[index] {
			t.Fatalf("role for %s = %q, want %q", assignment.Username, assignment.Role, want[index])
		}
	}
}

func Test_Migrate_seeds_default_roles_permissions_and_is_idempotent(t *testing.T) {
	// Given
	db := openMigrationTestDB(t)

	// When
	if err := Migrate(db); err != nil {
		t.Fatalf("first migration: %v", err)
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("second migration: %v", err)
	}

	// Then
	var roleCount int64
	if err := db.Model(&model.Role{}).Count(&roleCount).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if roleCount != 3 {
		t.Fatalf("roles = %d, want 3", roleCount)
	}
	var defaultRoles []model.Role
	if err := db.Where("is_default = ?", true).Find(&defaultRoles).Error; err != nil {
		t.Fatalf("load default roles: %v", err)
	}
	if len(defaultRoles) != 1 || defaultRoles[0].Name != model.RoleUser {
		t.Fatalf("default roles = %v, want [%s]", defaultRoles, model.RoleUser)
	}

	var permissionCount int64
	if err := db.Model(&model.Permission{}).Count(&permissionCount).Error; err != nil {
		t.Fatalf("count permissions: %v", err)
	}
	if permissionCount != 21 {
		t.Fatalf("permissions = %d, want 21", permissionCount)
	}

	var relationCount int64
	if err := db.Model(&model.RolePermission{}).Count(&relationCount).Error; err != nil {
		t.Fatalf("count role permissions: %v", err)
	}
	if relationCount != 21 {
		t.Fatalf("role permissions = %d, want 21", relationCount)
	}
	constraints := map[string][]string{
		"user_roles":       {"fk_user_roles_user", "fk_user_roles_role"},
		"role_permissions": {"fk_role_permissions_role", "fk_role_permissions_permission"},
	}
	for tableName, names := range constraints {
		for _, constraintName := range names {
			if !db.Migrator().HasConstraint(tableName, constraintName) {
				t.Fatalf("missing constraint %s on %s", constraintName, tableName)
			}
		}
	}
}

func openMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Errorf("close sqlite: %v", closeErr)
		}
	})
	return db
}
