package authadapter

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/pkg/cache"
	"go.uber.org/zap"
)

type TokenStore struct {
	cache  cache.Cache
	logger *zap.SugaredLogger
}

func NewTokenStore(cacheService cache.Cache, logger *zap.SugaredLogger) *TokenStore {
	if logger == nil {
		logger = zap.NewNop().Sugar()
	}
	return &TokenStore{cache: cacheService, logger: logger}
}

func (s *TokenStore) SaveVerificationToken(ctx context.Context, userID uint, token string, ttl time.Duration) error {
	return s.cache.Set(ctx, verificationKey(userID, token), userID, ttl)
}

func (s *TokenStore) VerificationUser(ctx context.Context, userID uint, token string) (uint, error) {
	var storedUserID uint
	err := s.cache.Get(ctx, verificationKey(userID, token), &storedUserID)
	return storedUserID, err
}

func (s *TokenStore) DeleteVerificationToken(ctx context.Context, userID uint, token string) error {
	return s.cache.Del(ctx, verificationKey(userID, token))
}

func (s *TokenStore) SavePasswordResetToken(ctx context.Context, token string, userID uint, ttl time.Duration) error {
	return s.cache.Set(ctx, passwordResetKey(token), userID, ttl)
}

func (s *TokenStore) PasswordResetUser(ctx context.Context, token string) (uint, error) {
	var userID uint
	err := s.cache.Get(ctx, passwordResetKey(token), &userID)
	return userID, err
}

func (s *TokenStore) DeletePasswordResetToken(ctx context.Context, token string) error {
	return s.cache.Del(ctx, passwordResetKey(token))
}

func (s *TokenStore) BlacklistToken(ctx context.Context, token string, ttl time.Duration) error {
	return s.cache.Set(ctx, blacklistKey(token), true, ttl)
}

func (s *TokenStore) IsTokenBlacklisted(ctx context.Context, token string) bool {
	var blacklisted bool
	if err := s.cache.Get(ctx, blacklistKey(token), &blacklisted); err != nil {
		if errors.Is(err, cache.ErrCacheMiss) {
			// Cache miss (key not found) → token is definitively not blacklisted.
			return false
		}
		// Any other error (Redis down, deserialization failure) means the
		// revocation state is indeterminate. Fail closed: treat as blacklisted
		// to preserve security over availability.
		s.logger.Errorf("TokenStore.IsTokenBlacklisted: cache.Get failed, failing closed: %v", err)
		return true
	}
	return blacklisted
}

func verificationKey(userID uint, token string) string {
	return fmt.Sprintf("verify_email:%d:%s", userID, token)
}

func passwordResetKey(token string) string {
	return "password_reset:" + tokenHash(token)
}

func blacklistKey(token string) string {
	return "token_blacklist:" + tokenHash(token)
}

func tokenHash(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
