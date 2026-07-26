package role

import (
	"context"

	"github.com/lyimoexiao/akari/internal/model"
)

func (s *Service) Delete(ctx context.Context, callerID, roleID uint) error {
	return s.policies.UpdatePolicy(ctx, func(repository Repository) error {
		if err := requireSuperAdmin(ctx, repository, callerID); err != nil {
			return err
		}
		storedRole, err := repository.RoleByID(ctx, roleID)
		if err != nil {
			return err
		}
		if storedRole.Name == model.RoleSuperAdmin {
			return ErrProtectedRole
		}
		if storedRole.IsDefault {
			return ErrDefaultRole
		}
		assignments, err := repository.RoleAssignmentCount(ctx, roleID)
		if err != nil {
			return err
		}
		if assignments > 0 {
			return ErrRoleInUse
		}
		if err := repository.ClearChildParents(ctx, roleID); err != nil {
			return err
		}
		return repository.DeleteRole(ctx, roleID)
	})
}

func (s *Service) SetDefault(ctx context.Context, callerID, roleID uint) error {
	return s.policies.UpdatePolicy(ctx, func(repository Repository) error {
		if err := requireSuperAdmin(ctx, repository, callerID); err != nil {
			return err
		}
		storedRole, err := repository.RoleByID(ctx, roleID)
		if err != nil {
			return err
		}
		if storedRole.Name == model.RoleSuperAdmin {
			return ErrProtectedRole
		}
		return repository.ReplaceDefaultRole(ctx, roleID)
	})
}

func (s *Service) SetRole(ctx context.Context, callerID uint, req SetRoleReq) error {
	return s.policies.UpdatePolicy(ctx, func(repository Repository) error {
		if err := requireSuperAdmin(ctx, repository, callerID); err != nil {
			return err
		}
		storedRole, err := repository.RoleByName(ctx, req.Role)
		if err != nil {
			return err
		}
		exists, err := repository.UserExists(ctx, req.UserID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrUserNotFound
		}
		return repository.ReplaceUserRole(ctx, req.UserID, storedRole.ID)
	})
}
