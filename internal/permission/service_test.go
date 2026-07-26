package permission_test

import (
	"context"
	"net/http"
	"slices"
	"testing"

	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/permission"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Service_returns_effective_permission_identifiers_for_each_role(t *testing.T) {
	// Given
	db := newTestDB(t)
	users := []model.User{
		{Username: "member", Email: "member@example.com", Password: "hash"},
		{Username: "operator", Email: "operator@example.com", Password: "hash"},
		{Username: "root", Email: "root@example.com", Password: "hash"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	assignTestRole(t, db, users[0].ID, model.RoleUser)
	assignTestRole(t, db, users[1].ID, model.RoleStaff)
	assignTestRole(t, db, users[2].ID, model.RoleSuperAdmin)
	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	svc := permission.NewService(backend)
	cases := []struct {
		name   string
		userID uint
		want   []string
	}{
		{
			name:   "user receives own permission",
			userID: users[0].ID,
			want:   []string{"closet.add", "closet.list", "closet.remove", "closet.rename", "skinlib.delete", "skinlib.update", "skinlib.upload", "yggdrasil.status.read"},
		},
		{
			name:   "staff receives inherited user permission",
			userID: users[1].ID,
			want:   []string{"closet.add", "closet.list", "closet.remove", "closet.rename", "skinlib.delete", "skinlib.manage", "skinlib.update", "skinlib.upload", "users.delete", "users.read", "users.reset-password", "users.verify-email", "yggdrasil.status.read"},
		},
		{
			name:   "super admin receives every inherited permission",
			userID: users[2].ID,
			want:   []string{"closet.add", "closet.list", "closet.remove", "closet.rename", "permissions.read", "request-logs.read", "roles.assign", "roles.create", "roles.delete", "roles.read", "roles.set-default", "roles.update", "skinlib.delete", "skinlib.manage", "skinlib.update", "skinlib.upload", "users.delete", "users.read", "users.reset-password", "users.verify-email", "yggdrasil.status.read"},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			// When
			got, queryErr := svc.IdentifiersForUser(context.Background(), testCase.userID)

			// Then
			if queryErr != nil {
				t.Fatalf("get identifiers: %v", queryErr)
			}
			if !slices.Equal(got, testCase.want) {
				t.Fatalf("identifiers = %v, want %v", got, testCase.want)
			}
		})
	}
}

func Test_Service_enforces_default_role_permissions(t *testing.T) {
	// Given
	db := newTestDB(t)
	users := []model.User{
		{Username: "member", Email: "member@example.com", Password: "hash"},
		{Username: "operator", Email: "operator@example.com", Password: "hash"},
		{Username: "root", Email: "root@example.com", Password: "hash"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	assignTestRole(t, db, users[0].ID, model.RoleUser)
	assignTestRole(t, db, users[1].ID, model.RoleStaff)
	assignTestRole(t, db, users[2].ID, model.RoleSuperAdmin)

	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	svc := permission.NewService(backend)

	tests := []struct {
		name   string
		userID uint
		path   string
		method string
		want   bool
	}{
		{name: "user reads own status", userID: users[0].ID, path: "/api/v1/yggdrasil/user/status", method: http.MethodGet, want: true},
		{name: "user cannot list users", userID: users[0].ID, path: "/api/v1/users", method: http.MethodGet, want: false},
		{name: "staff inherits user permission", userID: users[1].ID, path: "/api/v1/yggdrasil/user/status", method: http.MethodGet, want: true},
		{name: "staff lists users", userID: users[1].ID, path: "/api/v1/users", method: http.MethodGet, want: true},
		{name: "staff cannot assign roles", userID: users[1].ID, path: "/api/v1/roles/assign", method: http.MethodPost, want: false},
		{name: "super admin inherits staff permission", userID: users[2].ID, path: "/api/v1/users", method: http.MethodGet, want: true},
		{name: "super admin assigns roles", userID: users[2].ID, path: "/api/v1/roles/assign", method: http.MethodPost, want: true},
		{name: "super admin creates roles", userID: users[2].ID, path: "/api/v1/roles", method: http.MethodPost, want: true},
		{name: "super admin updates roles", userID: users[2].ID, path: "/api/v1/roles/42", method: http.MethodPut, want: true},
		{name: "super admin sets default role", userID: users[2].ID, path: "/api/v1/roles/42/default", method: http.MethodPut, want: true},
		{name: "super admin reads request log", userID: users[2].ID, path: "/api/v1/request-logs/request-123", method: http.MethodGet, want: true},
		{name: "staff cannot read request log", userID: users[1].ID, path: "/api/v1/request-logs/request-123", method: http.MethodGet, want: false},
		{name: "legacy admin route is denied", userID: users[2].ID, path: "/api/v1/admin/users", method: http.MethodGet, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// When
			allowed, role, enforceErr := svc.EnforceUser(context.Background(), permission.Check{
				UserID: tt.userID,
				Object: tt.path,
				Action: tt.method,
			})

			// Then
			if enforceErr != nil {
				t.Fatalf("enforce permission: %v", enforceErr)
			}
			if allowed != tt.want {
				t.Fatalf("allowed = %v, want %v for role %q", allowed, tt.want, role)
			}
		})
	}
}

func Test_Service_reads_current_role_on_every_enforcement(t *testing.T) {
	// Given
	db := newTestDB(t)
	user := model.User{Username: "operator", Email: "operator@example.com", Password: "hash"}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	assignTestRole(t, db, user.ID, model.RoleStaff)
	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	svc := permission.NewService(backend)

	// When
	var userRole model.Role
	if err := db.Where("name = ?", model.RoleUser).First(&userRole).Error; err != nil {
		t.Fatalf("load user role: %v", err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", user.ID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Create(&model.UserRole{UserID: user.ID, RoleID: userRole.ID}).Error
	}); err != nil {
		t.Fatalf("downgrade user: %v", err)
	}
	allowed, role, err := svc.EnforceUser(context.Background(), permission.Check{
		UserID: user.ID,
		Object: "/api/v1/users",
		Action: http.MethodGet,
	})
	// Then
	if err != nil {
		t.Fatalf("enforce permission: %v", err)
	}
	if allowed {
		t.Fatal("downgraded user retained staff permission")
	}
	if role != model.RoleUser {
		t.Fatalf("role = %q, want %q", role, model.RoleUser)
	}
}

func newTestDB(t *testing.T) *gorm.DB {
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

	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return db
}

func assignTestRole(t *testing.T, db *gorm.DB, userID uint, roleName string) {
	t.Helper()
	var stored model.Role
	if err := db.Where("name = ?", roleName).First(&stored).Error; err != nil {
		t.Fatalf("load role %s: %v", roleName, err)
	}
	if err := db.Create(&model.UserRole{UserID: userID, RoleID: stored.ID}).Error; err != nil {
		t.Fatalf("assign role %s: %v", roleName, err)
	}
}
