package permission_test

import (
	"context"
	"testing"

	"github.com/lyimoexiao/akari/internal/permission"
)

type fakeAccessControl struct {
	check       permission.Check
	identifiers []string
	snapshot    permission.Snapshot
}

func (backend *fakeAccessControl) EnforceUser(_ context.Context, check permission.Check) (bool, string, error) {
	backend.check = check
	return true, "staff", nil
}

func (backend *fakeAccessControl) IdentifiersForUser(context.Context, uint) ([]string, error) {
	return backend.identifiers, nil
}

func (backend *fakeAccessControl) Snapshot() (permission.Snapshot, error) {
	return backend.snapshot, nil
}

func Test_Service_uses_access_control_port_without_database(t *testing.T) {
	// Given
	backend := &fakeAccessControl{identifiers: []string{"users.read"}}
	service := permission.NewService(backend)
	check := permission.Check{UserID: 42, Object: "/api/v1/users", Action: "GET"}

	// When
	allowed, roleName, err := service.EnforceUser(t.Context(), check)

	// Then
	if err != nil {
		t.Fatalf("enforce permission: %v", err)
	}
	if !allowed || roleName != "staff" {
		t.Fatalf("allowed = %v, role = %q; want true, staff", allowed, roleName)
	}
	if backend.check != check {
		t.Fatalf("check = %#v, want %#v", backend.check, check)
	}
}
