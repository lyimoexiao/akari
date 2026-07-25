package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/smtp"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
	"github.com/lyimoexiao/akari/pkg/jwt"
	"github.com/lyimoexiao/akari/pkg/util"
	"gorm.io/gorm"
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

// Service provides authentication business logic.
type Service struct {
	db     *gorm.DB
	cache  cache.Cache
	jwt    *jwt.Manager
	mailer *smtp.Mailer
	cfg    *config.AuthConfig
	jwtCfg *config.JWTConfig
}

// NewService creates a new auth service.
func NewService(db *gorm.DB, c cache.Cache, m *jwt.Manager, mailer *smtp.Mailer, cfg *config.AuthConfig, jwtCfg *config.JWTConfig) *Service {
	return &Service{
		db:     db,
		cache:  c,
		jwt:    m,
		mailer: mailer,
		cfg:    cfg,
		jwtCfg: jwtCfg,
	}
}

// Register creates a new user account.
// The first registered user is automatically assigned the super_admin role.
func (s *Service) Register(ctx context.Context, req *RegisterReq) (*AuthResp, error) {
	if !s.cfg.RegistrationEnabled {
		return nil, ErrRegistrationClosed
	}

	// Check if this is the first user (super admin)
	var count int64
	if err := s.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error; err != nil {
		return nil, fmt.Errorf("count users: %w", err)
	}

	// Hard-delete any soft-deleted records with the same email or username,
	// so the new registration can reuse them.
	for _, field := range []string{"email", "username"} {
		val := req.Email
		if field == "username" {
			val = req.Username
		}
		var existing model.User
		if err := s.db.WithContext(ctx).Unscoped().Where(field+" = ?", val).First(&existing).Error; err == nil {
			if existing.DeletedAt.Valid {
				// Hard-delete the soft-deleted record, then continue to create a new one.
				if err := s.db.WithContext(ctx).Unscoped().Delete(&existing).Error; err != nil {
					return nil, fmt.Errorf("purge stale %s: %w", field, err)
				}
			} else {
				// Active record — this is a genuine duplicate.
				if field == "email" {
					return nil, ErrEmailAlreadyExists
				}
				return nil, ErrUsernameAlreadyUsed
			}
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("check %s: %w", field, err)
		}
	}

	// Hash password
	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	// Determine role: first user is super admin
	role := model.RoleUser
	if count == 0 {
		role = model.RoleSuperAdmin
	}

	user := &model.User{
		Username: req.Username,
		Email:    req.Email,
		Password: hashedPassword,
		Role:     role,
	}

	if err := s.db.WithContext(ctx).Create(user).Error; err != nil {
		// Race condition: another request created the same username/email
		// between our SELECT check and this INSERT.
		if isUniqueConstraint(err, "username") {
			return nil, ErrUsernameAlreadyUsed
		}
		if isUniqueConstraint(err, "email") {
			return nil, ErrEmailAlreadyExists
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	// Generate JWT token
	token, err := s.jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResp{
		Token: token,
		User:  toUserResp(user),
	}, nil
}

// Login authenticates a user and returns a JWT token.
func (s *Service) Login(ctx context.Context, req *LoginReq) (*AuthResp, error) {
	var user model.User
	if err := s.db.WithContext(ctx).Where("username = ? OR email = ?", req.Username, req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("find user: %w", err)
	}

	if !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := s.jwt.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &AuthResp{
		Token: token,
		User:  toUserResp(&user),
	}, nil
}

// GetCurrentUser returns the authenticated user's profile.
func (s *Service) GetCurrentUser(ctx context.Context, userID uint) (*UserResp, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	resp := toUserResp(&user)
	return &resp, nil
}

// SendVerificationEmail sends a verification email to the authenticated user.
// It generates a token, caches it, and sends the verification link via SMTP.
func (s *Service) SendVerificationEmail(ctx context.Context, userID uint) error {
	if !s.cfg.EmailVerificationEnabled {
		return ErrEmailVerificationDisabled
	}

	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return ErrUserNotFound
	}

	if user.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}

	// Generate a random verification token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	// Cache token with a TTL
	ttl := util.ParseDuration(s.cfg.VerifyEmailTokenTTL, 2*time.Hour)
	cacheKey := fmt.Sprintf("verify_email:%d:%s", user.ID, token)
	if err := s.cache.Set(ctx, cacheKey, user.ID, ttl); err != nil {
		return fmt.Errorf("cache token: %w", err)
	}

	// Send verification email
	subject := "验证你的邮箱地址"
	body := fmt.Sprintf(
		"你好 %s，\n\n请使用以下令牌验证你的邮箱地址：\n\n"+
			"验证令牌：%s\n\n用户 ID：%d\n\n此令牌将在 %s 后过期。\n\n如果你没有发起此请求，请忽略此邮件。",
		user.Username, token, user.ID, ttl.String(),
	)

	if err := s.mailer.Send([]string{user.Email}, subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

// VerifyEmail verifies a user's email address using a token.
func (s *Service) VerifyEmail(ctx context.Context, userID uint, token string) error {
	if !s.cfg.EmailVerificationEnabled {
		return ErrEmailVerificationDisabled
	}

	cacheKey := fmt.Sprintf("verify_email:%d:%s", userID, token)
	var cachedUserID uint
	if err := s.cache.Get(ctx, cacheKey, &cachedUserID); err != nil {
		return ErrInvalidToken
	}

	if cachedUserID != userID {
		return ErrInvalidToken
	}

	// Delete the token from cache (one-time use)
	_ = s.cache.Del(ctx, cacheKey)

	// Mark email as verified
	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("email_verified_at", time.Now()).Error; err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	return nil
}

// HasVerifiedEmail returns whether the given user has verified their email.
// If email verification is disabled, returns true (skipped).
func (s *Service) HasVerifiedEmail(ctx context.Context, userID uint) (bool, error) {
	if !s.cfg.EmailVerificationEnabled {
		return true, nil
	}
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return false, ErrUserNotFound
	}
	return user.IsEmailVerified(), nil
}

// GetJWTManager returns the JWT manager (used by middleware).
func (s *Service) GetJWTManager() *jwt.Manager {
	return s.jwt
}

// Logout blacklists the given token so it can no longer be used.
func (s *Service) Logout(ctx context.Context, tokenString string) error {
	// Parse token to get expiration time
	claims, err := s.jwt.ValidateToken(tokenString)
	if err != nil {
		return fmt.Errorf("validate token: %w", err)
	}

	// Calculate remaining TTL
	remaining := time.Until(claims.ExpiresAt.Time)
	if remaining <= 0 {
		return nil // token already expired
	}

	// Hash the token for use as a cache key (avoid storing the full token)
	hash := sha256String(tokenString)
	cacheKey := "token_blacklist:" + hash

	// Store in cache with same TTL as the token
	if err := s.cache.Set(ctx, cacheKey, true, remaining); err != nil {
		return fmt.Errorf("cache blacklist: %w", err)
	}

	return nil
}

// IsTokenBlacklisted checks whether a token has been blacklisted.
func (s *Service) IsTokenBlacklisted(ctx context.Context, tokenString string) bool {
	hash := sha256String(tokenString)
	cacheKey := "token_blacklist:" + hash
	var val bool
	if err := s.cache.Get(ctx, cacheKey, &val); err != nil {
		return false
	}
	return val
}

// isUniqueConstraint reports whether err is a UNIQUE constraint violation
// for the given column name. Handles both SQLite and MySQL dialects.
func isUniqueConstraint(err error, column string) bool {
	msg := err.Error()
	// SQLite: "UNIQUE constraint failed: users.username"
	// MySQL:  "Duplicate entry 'X' for key 'users.username'"
	return strings.Contains(msg, "UNIQUE constraint failed") && strings.Contains(msg, column) ||
		strings.Contains(msg, "Duplicate entry") && strings.Contains(msg, column)
}
func sha256String(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordReq) error {
	if !s.cfg.PasswordResetEnabled {
		return ErrPasswordResetDisabled
	}

	var user model.User
	if err := s.db.WithContext(ctx).Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil
	}

	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	ttl := util.ParseDuration(s.cfg.PasswordResetTokenTTL, 30*time.Minute)
	cacheKey := "password_reset:" + sha256String(token)
	if err := s.cache.Set(ctx, cacheKey, user.ID, ttl); err != nil {
		return fmt.Errorf("cache token: %w", err)
	}

	subject := "重置你的密码"
	body := fmt.Sprintf(
		"你好 %s，\n\n请使用以下令牌重置你的密码：\n\n"+
			"重置令牌：%s\n\n用户 ID：%d\n\n此令牌将在 %s 后过期。\n\n如果你没有发起此请求，请忽略此邮件。",
		user.Username, token, user.ID, ttl.String(),
	)

	if err := s.mailer.Send([]string{user.Email}, subject, body); err != nil {
		return fmt.Errorf("send email: %w", err)
	}

	return nil
}

func (s *Service) ResetPassword(ctx context.Context, req *ResetPasswordReq) error {
	if !s.cfg.PasswordResetEnabled {
		return ErrPasswordResetDisabled
	}

	var userID uint
	cacheKey := "password_reset:" + sha256String(req.Token)
	if err := s.cache.Get(ctx, cacheKey, &userID); err != nil {
		return ErrInvalidToken
	}

	_ = s.cache.Del(ctx, cacheKey)

	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).
		Update("password", hashedPassword).Error; err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	return nil
}
