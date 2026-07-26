package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

func (s *Service) SendVerificationEmail(ctx context.Context, userID uint) error {
	if !s.settings.EmailVerificationEnabled {
		return ErrEmailVerificationDisabled
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}
	if user.IsEmailVerified() {
		return ErrEmailAlreadyVerified
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	ttl := s.settings.VerifyEmailTokenTTL
	if err := s.store.SaveVerificationToken(ctx, user.ID, token, ttl); err != nil {
		return fmt.Errorf("cache token: %w", err)
	}
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

func (s *Service) VerifyEmail(ctx context.Context, userID uint, token string) error {
	if !s.settings.EmailVerificationEnabled {
		return ErrEmailVerificationDisabled
	}
	cachedUserID, err := s.store.VerificationUser(ctx, userID, token)
	if err != nil || cachedUserID != userID {
		return ErrInvalidToken
	}
	_ = s.store.DeleteVerificationToken(ctx, userID, token)
	if err := s.users.MarkEmailVerified(ctx, userID, time.Now()); err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *Service) HasVerifiedEmail(ctx context.Context, userID uint) (bool, error) {
	if !s.settings.EmailVerificationEnabled {
		return true, nil
	}
	user, err := s.users.FindByID(ctx, userID)
	if err != nil {
		return false, ErrUserNotFound
	}
	return user.IsEmailVerified(), nil
}

func (s *Service) ValidateToken(token string) (TokenClaims, error) {
	return s.tokens.ValidateToken(token)
}

func (s *Service) Logout(ctx context.Context, token string) error {
	claims, err := s.tokens.ValidateToken(token)
	if err != nil {
		return fmt.Errorf("validate token: %w", err)
	}
	remaining := time.Until(claims.ExpiresAt)
	if remaining <= 0 {
		return nil
	}
	if err := s.store.BlacklistToken(ctx, token, remaining); err != nil {
		return fmt.Errorf("cache blacklist: %w", err)
	}
	return nil
}

func (s *Service) IsTokenBlacklisted(ctx context.Context, token string) bool {
	return s.store.IsTokenBlacklisted(ctx, token)
}

func (s *Service) ForgotPassword(ctx context.Context, req *ForgotPasswordReq) error {
	if !s.settings.PasswordResetEnabled {
		return ErrPasswordResetDisabled
	}
	user, err := s.users.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil
	}
	token, err := randomToken()
	if err != nil {
		return err
	}
	ttl := s.settings.PasswordResetTokenTTL
	if err := s.store.SavePasswordResetToken(ctx, token, user.ID, ttl); err != nil {
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
	if !s.settings.PasswordResetEnabled {
		return ErrPasswordResetDisabled
	}
	userID, err := s.store.PasswordResetUser(ctx, req.Token)
	if err != nil {
		return ErrInvalidToken
	}
	_ = s.store.DeletePasswordResetToken(ctx, req.Token)
	hashedPassword, err := bcrypt.HashPassword(req.Password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, userID, hashedPassword); err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	return nil
}

func randomToken() (string, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return hex.EncodeToString(tokenBytes), nil
}
