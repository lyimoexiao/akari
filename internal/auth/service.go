package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

var (
	ErrEmailAlreadyExists        = errors.New("邮箱已存在")
	ErrUsernameAlreadyUsed       = errors.New("用户名已被使用")
	ErrInvalidCredentials        = errors.New("用户名/邮箱或密码错误")
	ErrUserNotFound              = errors.New("用户不存在")
	ErrEmailAlreadyVerified      = errors.New("邮箱已验证")
	ErrInvalidToken              = errors.New("令牌无效或已过期")
	ErrRegistrationClosed        = errors.New("注册已关闭")
	ErrEmailVerificationDisabled = errors.New("邮箱验证已禁用")
	ErrTokenBlacklisted          = errors.New("令牌已被吊销")
	ErrPasswordResetDisabled     = errors.New("密码重置已禁用")
	ErrWeakPassword              = errors.New("密码强度不足")
)

type Service struct {
	users    UserRepository
	roles    RoleFinder
	tokens   TokenManager
	store    TokenStore
	mailer   EmailSender
	settings Settings
}

func NewService(deps Dependencies) *Service {
	return &Service{
		users:    deps.Users,
		roles:    deps.Roles,
		tokens:   deps.Tokens,
		store:    deps.Store,
		mailer:   deps.Mailer,
		settings: deps.Settings,
	}
}

func (s *Service) Register(ctx context.Context, req *RegisterReq) (*AuthResp, error) {
	if !s.settings.RegistrationEnabled {
		return nil, ErrRegistrationClosed
	}
	count, err := s.users.Count(ctx)
	if err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}
	if err := s.users.PrepareRegistration(ctx, req.Email, req.Username); err != nil {
		return nil, err
	}
	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	var assignedRole *model.Role
	if count == 0 {
		role, findErr := s.roles.FindByName(ctx, model.RoleSuperAdmin)
		if findErr != nil {
			return nil, fmt.Errorf("resolve registration role: %w", findErr)
		}
		assignedRole = &role
	} else {
		role, findErr := s.roles.FindRegistrationRole(ctx)
		if findErr != nil {
			return nil, fmt.Errorf("resolve registration role: %w", findErr)
		}
		assignedRole = role
	}
	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
	}
	if err := s.users.Create(ctx, user, assignedRole); err != nil {
		if errors.Is(err, ErrUsernameAlreadyUsed) || errors.Is(err, ErrEmailAlreadyExists) {
			return nil, err
		}
		return nil, fmt.Errorf("create user: %w", err)
	}
	roleName := ""
	if assignedRole != nil {
		user.Roles = []model.Role{*assignedRole}
		roleName = assignedRole.Name
	}
	token, err := s.tokens.GenerateToken(user.ID, user.Username, roleName)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	return &AuthResp{Token: token, User: toUserResp(user)}, nil
}

func (s *Service) Login(ctx context.Context, req *LoginReq) (*AuthResp, error) {
	user, err := s.users.FindByLogin(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	if !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}
	token, err := s.tokens.GenerateToken(user.ID, user.Username, model.PrimaryRole(user.Roles))
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	return &AuthResp{Token: token, User: toUserResp(&user)}, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID uint) (*UserResp, error) {
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	response := toUserResp(&user)
	return &response, nil
}
