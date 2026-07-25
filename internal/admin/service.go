package admin

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("用户不存在")
	ErrCannotDeleteSelf  = errors.New("不能删除自己")
	ErrCannotModifyAdmin = errors.New("staff 不能操作超级管理员")
	ErrRoleNotAllowed    = errors.New("只有超级管理员才能设置角色")
)

// Service provides admin business logic.
type Service struct {
	db *gorm.DB
}

// NewService creates an admin service.
func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ListUsers returns a paginated user list.
func (s *Service) ListUsers(ctx context.Context, req *ListUsersReq) (*ListUsersResp, error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 || req.PageSize > 100 {
		req.PageSize = 20
	}

	q := s.db.WithContext(ctx).Model(&model.User{})
	if req.Query != "" {
		like := "%" + req.Query + "%"
		q = q.Where("username LIKE ? OR email LIKE ?", like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	var users []model.User
	offset := (req.Page - 1) * req.PageSize
	if err := q.Order("created_at DESC").Offset(offset).Limit(req.PageSize).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	items := make([]UserItem, len(users))
	for i, u := range users {
		items[i] = toUserItem(&u)
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	return &ListUsersResp{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

// VerifyEmail manually marks a user's email as verified.
func (s *Service) VerifyEmail(ctx context.Context, req *AdminVerifyEmailReq) error {
	result := s.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", req.UserID).
		Update("email_verified_at", time.Now())
	if result.Error != nil {
		return fmt.Errorf("verify email: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return ErrUserNotFound
	}
	return nil
}

// SetRole changes a user's role (super_admin only — enforced by caller).
func (s *Service) SetRole(ctx context.Context, callerIsSuperAdmin bool, req *SetRoleReq) error {
	if !callerIsSuperAdmin {
		return ErrRoleNotAllowed
	}

	// Fetch target user
	var target model.User
	if err := s.db.WithContext(ctx).First(&target, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find target user: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&target).Update("role", req.Role).Error; err != nil {
		return fmt.Errorf("update role: %w", err)
	}
	return nil
}

// ResetPassword forcibly resets a user's password.
func (s *Service) ResetPassword(ctx context.Context, callerIsSuperAdmin bool, req *AdminResetPasswordReq) error {
	var target model.User
	if err := s.db.WithContext(ctx).First(&target, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find target user: %w", err)
	}

	// Staff cannot reset super_admin passwords
	if !callerIsSuperAdmin && target.IsSuperAdmin() {
		return ErrCannotModifyAdmin
	}

	hashed, err := bcrypt.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&target).Update("password", hashed).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

// DeleteUser soft-deletes a user.
func (s *Service) DeleteUser(ctx context.Context, callerID uint, callerIsSuperAdmin bool, req *DeleteUserReq) error {
	if req.UserID == callerID {
		return ErrCannotDeleteSelf
	}

	var target model.User
	if err := s.db.WithContext(ctx).First(&target, req.UserID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("find target user: %w", err)
	}

	// Staff cannot delete super_admin
	if !callerIsSuperAdmin && target.IsSuperAdmin() {
		return ErrCannotModifyAdmin
	}

	if err := s.db.WithContext(ctx).Delete(&target).Error; err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}
