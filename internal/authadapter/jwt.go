package authadapter

import (
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/pkg/jwt"
)

type JWTManager struct {
	manager *jwt.Manager
}

func NewJWTManager(manager *jwt.Manager) *JWTManager {
	return &JWTManager{manager: manager}
}

func (m *JWTManager) GenerateToken(userID uint, username, role string) (string, error) {
	return m.manager.GenerateToken(userID, username, role)
}

func (m *JWTManager) ValidateToken(token string) (auth.TokenClaims, error) {
	claims, err := m.manager.ValidateToken(token)
	if err != nil {
		return auth.TokenClaims{}, err
	}
	return auth.TokenClaims{
		UserID:    claims.UserID,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: claims.ExpiresAt.Time,
	}, nil
}
