package role

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/lyimoexiao/akari/internal/model"
)

var (
	ErrUserNotFound           = errors.New("用户不存在")
	ErrInvalidRole            = errors.New("角色无效")
	ErrRoleNotFound           = errors.New("角色不存在")
	ErrInvalidRoleName        = errors.New("角色标识符格式无效")
	ErrInvalidPermission      = errors.New("权限标识符无效")
	ErrRoleExists             = errors.New("角色标识符已存在")
	ErrRoleInUse              = errors.New("角色正在被用户使用")
	ErrDefaultRole            = errors.New("请先设置其他默认注册角色")
	ErrProtectedRole          = errors.New("超级管理员角色不可删除或改名")
	ErrCannotModifySuperAdmin = errors.New("仅超级管理员可管理角色")
)

var roleNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,63}$`)

type Service struct {
	repository Repository
	policies   PolicyUpdater
}

func NewService(dependencies Dependencies) *Service {
	return &Service{repository: dependencies.Repository, policies: dependencies.Policies}
}

func (s *Service) Create(ctx context.Context, callerID uint, req CreateReq) (Item, error) {
	req.Name = strings.TrimSpace(req.Name)
	if !roleNamePattern.MatchString(req.Name) {
		return Item{}, ErrInvalidRoleName
	}
	req.Description = strings.TrimSpace(req.Description)
	var created model.Role
	err := s.policies.UpdatePolicy(ctx, func(repository Repository) error {
		if err := requireSuperAdmin(ctx, repository, callerID); err != nil {
			return err
		}
		exists, err := repository.RoleNameExists(ctx, req.Name, 0)
		if err != nil {
			return err
		}
		if exists {
			return ErrRoleExists
		}
		permissions, err := resolvePermissions(ctx, repository, req.Permissions)
		if err != nil {
			return err
		}
		created = model.Role{Name: req.Name, Description: req.Description}
		if err := repository.CreateRole(ctx, &created); err != nil {
			return err
		}
		return repository.ReplaceRolePermissions(ctx, created.ID, permissions)
	})
	if err != nil {
		return Item{}, err
	}
	return itemByID(ctx, s.repository, created.ID)
}

func (s *Service) Update(ctx context.Context, callerID, roleID uint, req UpdateReq) (Item, error) {
	req.Name = strings.TrimSpace(req.Name)
	if !roleNamePattern.MatchString(req.Name) {
		return Item{}, ErrInvalidRoleName
	}
	req.Description = strings.TrimSpace(req.Description)
	err := s.policies.UpdatePolicy(ctx, func(repository Repository) error {
		if err := requireSuperAdmin(ctx, repository, callerID); err != nil {
			return err
		}
		storedRole, err := repository.RoleByID(ctx, roleID)
		if err != nil {
			return err
		}
		if storedRole.Name == model.RoleSuperAdmin && req.Name != model.RoleSuperAdmin {
			return ErrProtectedRole
		}
		exists, err := repository.RoleNameExists(ctx, req.Name, roleID)
		if err != nil {
			return err
		}
		if exists {
			return ErrRoleExists
		}
		permissions, err := resolvePermissions(ctx, repository, req.Permissions)
		if err != nil {
			return err
		}
		if err := repository.UpdateRole(ctx, roleID, RoleChanges{
			Name: req.Name, Description: req.Description,
		}); err != nil {
			return err
		}
		return repository.ReplaceRolePermissions(ctx, roleID, permissions)
	})
	if err != nil {
		return Item{}, err
	}
	return itemByID(ctx, s.repository, roleID)
}
