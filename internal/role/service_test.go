package role_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"github.com/lyimoexiao/akari/internal/role"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Service_set_role_updates_user(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	user := model.User{Username: "member", Email: "member@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	svc := newRoleTestService(t, db)

	// When
	err := svc.SetRole(context.Background(), root.ID, role.SetRoleReq{UserID: user.ID, Role: model.RoleStaff})

	// Then
	if err != nil {
		t.Fatalf("set role: %v", err)
	}
	var updated model.User
	if err := db.Preload("Roles").First(&updated, user.ID).Error; err != nil {
		t.Fatalf("reload user: %v", err)
	}
	if got := model.PrimaryRole(updated.Roles); got != model.RoleStaff {
		t.Fatalf("role = %q, want %q", got, model.RoleStaff)
	}
}

func Test_Service_set_role_rejects_unknown_role(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	svc := newRoleTestService(t, db)

	// When
	err := svc.SetRole(context.Background(), root.ID, role.SetRoleReq{UserID: 999, Role: "owner"})

	// Then
	if !errors.Is(err, role.ErrInvalidRole) {
		t.Fatalf("error = %v, want ErrInvalidRole", err)
	}
}

func createRoleTestUser(t *testing.T, db *gorm.DB, username, roleName string) model.User {
	t.Helper()

	user := model.User{Username: username, Email: username + "@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create %s: %v", username, err)
	}
	var storedRole model.Role
	if err := db.Where("name = ?", roleName).First(&storedRole).Error; err != nil {
		t.Fatalf("load role %s: %v", roleName, err)
	}
	if err := db.Create(&model.UserRole{UserID: user.ID, RoleID: storedRole.ID}).Error; err != nil {
		t.Fatalf("assign role %s: %v", roleName, err)
	}
	return user
}

func newRoleTestService(t *testing.T, db *gorm.DB) *role.Service {
	t.Helper()

	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	return role.NewService(role.Dependencies{Repository: backend, Policies: backend})
}

func newRoleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return db
}
