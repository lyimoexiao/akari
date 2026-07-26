package user

import (
	"context"
	"errors"
	"fmt"
	"math"

	"github.com/lyimoexiao/akari/internal/model"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrCannotDeleteSelf  = errors.New("不能删除自己")
	ErrCannotModifyAdmin = errors.New("非超级管理员不能操作超级管理员")
)

type Service struct {
	repository Repository
	clock      Clock
	hasher     PasswordHasher
}

func NewService(dependencies Dependencies) *Service {
	return &Service{
		repository: dependencies.Repository,
		clock:      dependencies.Clock,
		hasher:     dependencies.Hasher,
	}
}

func (s *Service) ListUsers(ctx context.Context, req ListUsersReq) (*ListResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}
	users, total, err := s.repository.List(ctx, ListQuery{
		Search: req.Query, Page: req.Page, PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(users))
	for index := range users {
		items[index] = toItem(&users[index])
	}
	return &ListResp{
		Items: items, Total: total, Page: req.Page, PageSize: req.PageSize,
		TotalPages: int(math.Ceil(float64(total) / float64(req.PageSize))),
	}, nil
}

func (s *Service) VerifyEmail(ctx context.Context, callerID uint, req VerifyEmailReq) error {
	target, err := s.loadMutableTarget(ctx, callerID, req.UserID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	if err := s.repository.VerifyEmail(ctx, target.ID, now); err != nil {
		return fmt.Errorf("verify email: %w", err)
	}
	return nil
}

func (s *Service) ResetPassword(ctx context.Context, callerID uint, req ResetPasswordReq) error {
	target, err := s.loadMutableTarget(ctx, callerID, req.UserID)
	if err != nil {
		return err
	}
	hashed, err := s.hasher.Hash(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.repository.UpdatePassword(ctx, target.ID, hashed); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, command DeleteUserCommand) error {
	if command.UserID == command.CallerID {
		return ErrCannotDeleteSelf
	}
	target, err := s.loadMutableTarget(ctx, command.CallerID, command.UserID)
	if err != nil {
		return err
	}
	if err := s.repository.Delete(ctx, target.ID); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}


func (s *Service) loadMutableTarget(ctx context.Context, callerID, userID uint) (model.User, error) {
	target, err := s.repository.FindByID(ctx, userID)
	if err != nil {
		return model.User{}, err
	}
	if model.PrimaryRole(target.Roles) != model.RoleSuperAdmin {
		return target, nil
	}
	allowed, err := s.repository.HasRole(ctx, callerID, model.RoleSuperAdmin)
	if err != nil {
		return model.User{}, err
	}
	if !allowed {
		return model.User{}, ErrCannotModifyAdmin
	}
	return target, nil
}