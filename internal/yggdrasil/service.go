package yggdrasil

import (
	"context"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrTokenLimitReached  = errors.New("token limit reached")
)

const (
	tokenExpiration     = 15 * 24 * time.Hour
	maxTokensPerUser    = 10
	serverSessionTTL    = 30 * time.Second
	userUUIDPrefix      = "YggdrasilUser:"
	offlinePlayerPrefix = "OfflinePlayer:"
)

type Service struct {
	repository Repository
	sessions   SessionStore
	signer     Signer
	reporter   SigningFailureReporter
	settings   Settings
}

func NewService(deps Dependencies) *Service {
	return &Service{
		repository: deps.Repository,
		sessions:   deps.Sessions,
		signer:     deps.Signer,
		reporter:   deps.Reporter,
		settings:   deps.Settings,
	}
}

func OfflineUUID(name string) string {
	hash := md5.Sum([]byte(offlinePlayerPrefix + name))
	hash[6] = (hash[6] & 0x0f) | 0x30
	hash[8] = (hash[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:])
}

func userUUID(email string) string {
	hash := md5.Sum([]byte(userUUIDPrefix + email))
	hash[6] = (hash[6] & 0x0f) | 0x30
	hash[8] = (hash[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", hash[0:4], hash[4:6], hash[6:8], hash[8:10], hash[10:])
}

func generateAccessToken() string {
	token := make([]byte, 16)
	_, _ = rand.Read(token)
	token[6] = (token[6] & 0x0f) | 0x40
	token[8] = (token[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", token[0:4], token[4:6], token[6:8], token[8:10], token[10:])
}

func formatUUIDWithoutDashes(uuid string) string {
	if len(uuid) == 36 {
		return uuid[0:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:36]
	}
	return uuid
}

func (s *Service) findUserByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := s.repository.FindUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) findUserByLogin(ctx context.Context, login string) (*model.User, error) {
	user, err := s.repository.FindUserByLogin(ctx, login)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) createToken(ctx context.Context, userEmail, clientToken, profileUUID, loginIP string) (*YggdrasilToken, error) {
	count, err := s.repository.CountValidTokens(ctx, userEmail)
	if err != nil {
		return nil, fmt.Errorf("count tokens: %w", err)
	}
	if int(count) >= maxTokensPerUser {
		if err := s.repository.RevokeOldestValidToken(ctx, userEmail); err != nil {
			return nil, fmt.Errorf("revoke oldest token: %w", err)
		}
	}
	now := time.Now()
	token := &YggdrasilToken{
		AccessToken: generateAccessToken(),
		ClientToken: clientToken,
		UserEmail:   userEmail,
		ProfileUUID: profileUUID,
		Status:      TokenStatusValid,
		LoginIP:     loginIP,
		IssuedAt:    now,
		ExpiresAt:   now.Add(tokenExpiration),
	}
	if err := s.repository.CreateToken(ctx, token); err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	return token, nil
}

func (s *Service) findValidToken(ctx context.Context, accessToken, clientToken string) (*YggdrasilToken, error) {
	token, err := s.repository.FindToken(ctx, accessToken)
	if err != nil {
		return nil, err
	}
	if clientToken != "" && token.ClientToken != clientToken {
		return nil, ErrInvalidToken
	}
	if token.Status == TokenStatusInvalid {
		return nil, ErrInvalidToken
	}
	if time.Now().After(token.ExpiresAt) {
		_ = s.repository.RevokeToken(ctx, token.AccessToken)
		return nil, ErrTokenExpired
	}
	return &token, nil
}

func (s *Service) revokeToken(ctx context.Context, token *YggdrasilToken) error {
	return s.repository.RevokeToken(ctx, token.AccessToken)
}

func (s *Service) getProfilesForUser(ctx context.Context, email string) ([]YggdrasilProfile, error) {
	return s.repository.ProfilesForUser(ctx, email)
}

func (s *Service) ensureProfile(ctx context.Context, userEmail, username string) (*YggdrasilProfile, error) {
	profiles, err := s.getProfilesForUser(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	if len(profiles) > 0 {
		return &profiles[0], nil
	}
	profile := &YggdrasilProfile{
		UUID:      OfflineUUID(username),
		Name:      username,
		UserEmail: userEmail,
		Model:     "default",
	}
	if err := s.repository.CreateProfile(ctx, profile); err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}

func (s *Service) getProfileByUUID(ctx context.Context, uuid string) (*YggdrasilProfile, error) {
	profile, err := s.repository.ProfileByUUID(ctx, uuid)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (s *Service) getProfilesByNames(ctx context.Context, names []string) ([]YggdrasilProfile, error) {
	return s.repository.ProfilesByNames(ctx, names)
}

func toProfileResp(profile *YggdrasilProfile) ProfileResp {
	return ProfileResp{ID: formatUUIDWithoutDashes(profile.UUID), Name: profile.Name}
}

type texturesPayload struct {
	Timestamp   int64          `json:"timestamp"`
	ProfileID   string         `json:"profileId"`
	ProfileName string         `json:"profileName"`
	Textures    map[string]any `json:"textures"`
}

func (s *Service) buildTexturesProperty(profileID, profileName string, sign bool) Property {
	payload := texturesPayload{
		Timestamp:   time.Now().UnixMilli(),
		ProfileID:   formatUUIDWithoutDashes(profileID),
		ProfileName: profileName,
		Textures:    map[string]any{},
	}
	raw, _ := json.Marshal(payload)
	property := Property{
		Name:  "textures",
		Value: base64.StdEncoding.EncodeToString(raw),
	}
	if sign && s.signer != nil {
		signature, err := s.signer.Sign([]byte(property.Value))
		if err != nil {
			if s.reporter != nil {
				s.reporter.ReportSigningFailure(err)
			}
		} else {
			property.Signature = signature
		}
	}
	return property
}

func (s *Service) profileWithTextures(profile *YggdrasilProfile, sign bool) ProfileResp {
	response := toProfileResp(profile)
	response.Properties = []Property{s.buildTexturesProperty(profile.UUID, profile.Name, sign)}
	return response
}
