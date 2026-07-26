package role_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/permission"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"github.com/lyimoexiao/akari/internal/role"
)

func Test_Service_updates_custom_role_permissions_without_restart(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	permissions := permission.NewService(backend)
	svc := role.NewService(role.Dependencies{Repository: backend, Policies: backend})
	created, err := svc.Create(context.Background(), root.ID, role.CreateReq{
		Name: "auditor", Description: "审计员",
	})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	member := model.User{Username: "auditor", Email: "auditor@example.com", Password: "hash"}
	if err := db.Create(&member).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	if err := svc.SetRole(context.Background(), root.ID, role.SetRoleReq{UserID: member.ID, Role: created.Name}); err != nil {
		t.Fatalf("assign custom role: %v", err)
	}

	// When
	_, err = svc.Update(context.Background(), root.ID, created.ID, role.UpdateReq{
		Name: "auditor", Description: "审计员", Permissions: []string{"users.read"},
	})

	// Then
	if err != nil {
		t.Fatalf("update role: %v", err)
	}
	allowed, primaryRole, err := permissions.EnforceUser(context.Background(), permission.Check{
		UserID: member.ID, Object: "/api/v1/users", Action: http.MethodGet,
	})
	if err != nil {
		t.Fatalf("enforce updated permissions: %v", err)
	}
	if !allowed || primaryRole != "auditor" {
		t.Fatalf("allowed = %v, role = %q; want true, auditor", allowed, primaryRole)
	}
}

func Test_Service_protects_super_admin_role_and_users(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	svc := newRoleTestService(t, db)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	member := createRoleTestUser(t, db, "member", model.RoleUser)
	var superAdmin model.Role
	if err := db.Where("name = ?", model.RoleSuperAdmin).First(&superAdmin).Error; err != nil {
		t.Fatalf("load super admin role: %v", err)
	}

	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{
			name: "non-super-admin cannot edit super-admin permissions",
			run: func() error {
				_, updateErr := svc.Update(context.Background(), member.ID, superAdmin.ID, role.UpdateReq{
					Name: model.RoleSuperAdmin, Description: "changed", Permissions: []string{"roles.read"},
				})
				return updateErr
			},
			want: role.ErrCannotModifySuperAdmin,
		},
		{
			name: "non-super-admin cannot modify a super-admin user",
			run: func() error {
				return svc.SetRole(context.Background(), member.ID, role.SetRoleReq{UserID: root.ID, Role: model.RoleUser})
			},
			want: role.ErrCannotModifySuperAdmin,
		},
		{
			name: "non-super-admin cannot grant super-admin",
			run: func() error {
				return svc.SetRole(context.Background(), member.ID, role.SetRoleReq{UserID: member.ID, Role: model.RoleSuperAdmin})
			},
			want: role.ErrCannotModifySuperAdmin,
		},
		{
			name: "super-admin role cannot be deleted",
			run: func() error {
				return svc.Delete(context.Background(), root.ID, superAdmin.ID)
			},
			want: role.ErrProtectedRole,
		},
		{
			name: "super-admin role cannot be the registration default",
			run: func() error {
				return svc.SetDefault(context.Background(), root.ID, superAdmin.ID)
			},
			want: role.ErrProtectedRole,
		},
		{
			name: "non-super-admin cannot create ordinary roles",
			run: func() error {
				_, createErr := svc.Create(context.Background(), member.ID, role.CreateReq{Name: "operator"})
				return createErr
			},
			want: role.ErrCannotModifySuperAdmin,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			// When
			err := testCase.run()

			// Then
			if !errors.Is(err, testCase.want) {
				t.Fatalf("error = %v, want %v", err, testCase.want)
			}
		})
	}
}

func Test_Service_keeps_deleted_builtin_role_deleted_after_migration(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	svc := newRoleTestService(t, db)
	replacement, err := svc.Create(context.Background(), root.ID, role.CreateReq{
		Name: "member", Description: "注册用户",
	})
	if err != nil {
		t.Fatalf("create replacement role: %v", err)
	}
	if err := svc.SetDefault(context.Background(), root.ID, replacement.ID); err != nil {
		t.Fatalf("set replacement default role: %v", err)
	}
	var builtin model.Role
	if err := db.Where("name = ?", model.RoleUser).First(&builtin).Error; err != nil {
		t.Fatalf("load builtin user role: %v", err)
	}

	// When
	if err := svc.Delete(context.Background(), root.ID, builtin.ID); err != nil {
		t.Fatalf("delete builtin user role: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("rerun migration: %v", err)
	}

	// Then
	var count int64
	if err := db.Model(&model.Role{}).Where("name = ?", model.RoleUser).Count(&count).Error; err != nil {
		t.Fatalf("count active user roles: %v", err)
	}
	if count != 0 {
		t.Fatalf("active user roles = %d, want 0", count)
	}
}

func Test_Service_rejects_deleting_current_default_role(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	svc := newRoleTestService(t, db)
	var defaultRole model.Role
	if err := db.Where("is_default = ?", true).First(&defaultRole).Error; err != nil {
		t.Fatalf("load default role: %v", err)
	}

	// When
	err := svc.Delete(context.Background(), root.ID, defaultRole.ID)

	// Then
	if !errors.Is(err, role.ErrDefaultRole) {
		t.Fatalf("error = %v, want ErrDefaultRole", err)
	}
}

func Test_Service_rejects_unknown_permission_without_creating_role(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	svc := newRoleTestService(t, db)

	// When
	_, err := svc.Create(context.Background(), root.ID, role.CreateReq{
		Name: "auditor", Permissions: []string{"missing.permission"},
	})

	// Then
	if !errors.Is(err, role.ErrInvalidPermission) {
		t.Fatalf("error = %v, want ErrInvalidPermission", err)
	}
	var count int64
	if err := db.Model(&model.Role{}).Where("name = ?", "auditor").Count(&count).Error; err != nil {
		t.Fatalf("count roles: %v", err)
	}
	if count != 0 {
		t.Fatalf("roles = %d, want 0", count)
	}
}

func Test_Service_rejects_deleting_assigned_role(t *testing.T) {
	// Given
	db := newRoleTestDB(t)
	root := createRoleTestUser(t, db, "root", model.RoleSuperAdmin)
	member := createRoleTestUser(t, db, "member", model.RoleUser)
	svc := newRoleTestService(t, db)
	custom, err := svc.Create(context.Background(), root.ID, role.CreateReq{Name: "auditor"})
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if err := svc.SetRole(context.Background(), root.ID, role.SetRoleReq{UserID: member.ID, Role: custom.Name}); err != nil {
		t.Fatalf("assign role: %v", err)
	}

	// When
	err = svc.Delete(context.Background(), root.ID, custom.ID)

	// Then
	if !errors.Is(err, role.ErrRoleInUse) {
		t.Fatalf("error = %v, want ErrRoleInUse", err)
	}
}
