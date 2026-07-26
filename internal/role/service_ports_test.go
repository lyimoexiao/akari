package role_test

import (
	"context"
	"testing"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/role"
)

type fakeRoleRepository struct {
	created model.Role
}

func (*fakeRoleRepository) HasUserRole(context.Context, uint, string) (bool, error) {
	return true, nil
}

func (*fakeRoleRepository) RoleNameExists(context.Context, string, uint) (bool, error) {
	return false, nil
}

func (*fakeRoleRepository) PermissionsByName(context.Context, []string) ([]model.Permission, error) {
	return []model.Permission{{Name: "users.read"}}, nil
}

func (repository *fakeRoleRepository) CreateRole(_ context.Context, storedRole *model.Role) error {
	storedRole.ID = 7
	repository.created = *storedRole
	return nil
}

func (*fakeRoleRepository) ReplaceRolePermissions(context.Context, uint, []model.Permission) error {
	return nil
}

func (repository *fakeRoleRepository) RoleByID(context.Context, uint) (model.Role, error) {
	return repository.created, nil
}

func (*fakeRoleRepository) UpdateRole(context.Context, uint, role.RoleChanges) error { return nil }

func (*fakeRoleRepository) RoleAssignmentCount(context.Context, uint) (int64, error) {
	return 0, nil
}

func (*fakeRoleRepository) ClearChildParents(context.Context, uint) error { return nil }

func (*fakeRoleRepository) DeleteRole(context.Context, uint) error { return nil }

func (*fakeRoleRepository) ReplaceDefaultRole(context.Context, uint) error { return nil }

func (*fakeRoleRepository) RoleByName(context.Context, string) (model.Role, error) {
	return model.Role{}, nil
}

func (*fakeRoleRepository) UserExists(context.Context, uint) (bool, error) { return true, nil }

func (*fakeRoleRepository) ReplaceUserRole(context.Context, uint, uint) error { return nil }

func (*fakeRoleRepository) RegistrationRole(context.Context) (*model.Role, error) {
	return nil, nil
}

func (*fakeRoleRepository) ListRoles(context.Context) ([]model.Role, error) { return nil, nil }

type fakePolicyUpdater struct {
	repository role.Repository
}

func (updater fakePolicyUpdater) UpdatePolicy(ctx context.Context, change func(role.Repository) error) error {
	return change(updater.repository)
}

func Test_Service_normalizes_role_before_repository_call(t *testing.T) {
	// Given
	repository := &fakeRoleRepository{}
	service := role.NewService(role.Dependencies{
		Repository: repository,
		Policies:   fakePolicyUpdater{repository: repository},
	})

	// When
	created, err := service.Create(t.Context(), 1, role.CreateReq{
		Name: " auditor ", Description: " audit access ", Permissions: []string{"users.read"},
	})

	// Then
	if err != nil {
		t.Fatalf("create role: %v", err)
	}
	if repository.created.Name != "auditor" || repository.created.Description != "audit access" {
		t.Fatalf("normalized role = %#v", repository.created)
	}
	if created.Name != "auditor" {
		t.Fatalf("created role name = %q, want auditor", created.Name)
	}
}
