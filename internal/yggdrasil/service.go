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
	ErrTextureNotFound    = errors.New("texture not found")
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

// ── Profile Texture Management ──

// SetSkin sets the active skin texture for the user's first profile.
func (s *Service) SetSkin(ctx context.Context, userID uint, textureTID uint) error {
	profile, err := s.getProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the texture exists and is of a valid skin type
	texture, err := s.repository.TextureByID(ctx, textureTID)
	if err != nil {
		return fmt.Errorf("check texture: %w", err)
	}
	if texture == nil {
		return ErrTextureNotFound
	}
	if texture.Type != "steve" && texture.Type != "alex" {
		return fmt.Errorf("纹理类型不是皮肤")
	}

	// Update model based on texture type
	modelType := "default"
	if texture.Type == "alex" {
		modelType = "slim"
	}

	if err := s.repository.UpdateProfileSkin(ctx, profile.ID, &textureTID); err != nil {
		return fmt.Errorf("update profile skin: %w", err)
	}
	// Also update the profile model
	_ = s.repository.UpdateProfileModel(ctx, profile.ID, modelType)

	return nil
}

// SetCape sets the active cape texture for the user's first profile.
func (s *Service) SetCape(ctx context.Context, userID uint, textureTID uint) error {
	profile, err := s.getProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	// Verify the texture exists and is a cape
	texture, err := s.repository.TextureByID(ctx, textureTID)
	if err != nil {
		return fmt.Errorf("check texture: %w", err)
	}
	if texture == nil {
		return ErrTextureNotFound
	}
	if texture.Type != "cape" {
		return fmt.Errorf("纹理类型不是披风")
	}

	if err := s.repository.UpdateProfileCape(ctx, profile.ID, &textureTID); err != nil {
		return fmt.Errorf("update profile cape: %w", err)
	}
	return nil
}

// ClearSkin removes the active skin from the user's first profile.
func (s *Service) ClearSkin(ctx context.Context, userID uint) error {
	profile, err := s.getProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.repository.UpdateProfileSkin(ctx, profile.ID, nil); err != nil {
		return fmt.Errorf("clear profile skin: %w", err)
	}
	return nil
}

// ClearCape removes the active cape from the user's first profile.
func (s *Service) ClearCape(ctx context.Context, userID uint) error {
	profile, err := s.getProfileByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if err := s.repository.UpdateProfileCape(ctx, profile.ID, nil); err != nil {
		return fmt.Errorf("clear profile cape: %w", err)
	}
	return nil
}

func (s *Service) getProfileByUserID(ctx context.Context, userID uint) (*YggdrasilProfile, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	profiles, err := s.getProfilesForUser(ctx, user.Email)
	if err != nil {
		return nil, err
	}
	if len(profiles) == 0 {
		return nil, ErrProfileNotFound
	}
	return &profiles[0], nil
}

type texturesPayload struct {
	Timestamp   int64          `json:"timestamp"`
	ProfileID   string         `json:"profileId"`
	ProfileName string         `json:"profileName"`
	Textures    map[string]any `json:"textures"`
}

func (s *Service) buildTexturesProperty(ctx context.Context, profile *YggdrasilProfile, sign bool) Property {
	textures := map[string]any{}

	// Resolve skin texture
	if profile.TextureSkinID != nil {
		skin, err := s.repository.TextureByID(ctx, *profile.TextureSkinID)
		if err == nil && skin != nil {
			modelType := "default"
			if skin.Type == "alex" {
				modelType = "slim"
			}
			texURL := skin.URL
			if texURL == "" {
				texURL = s.textureURL(skin.Hash)
			}
			textures["SKIN"] = map[string]any{
				"url":      texURL,
				"metadata": map[string]string{"model": modelType},
			}
		}
	}

	// Resolve cape texture
	if profile.TextureCapeID != nil {
		cape, err := s.repository.TextureByID(ctx, *profile.TextureCapeID)
		if err == nil && cape != nil {
			texURL := cape.URL
			if texURL == "" {
				texURL = s.textureURL(cape.Hash)
			}
			textures["CAPE"] = map[string]any{
				"url": texURL,
			}
		}
	}

	payload := texturesPayload{
		Timestamp:   time.Now().UnixMilli(),
		ProfileID:   formatUUIDWithoutDashes(profile.UUID),
		ProfileName: profile.Name,
		Textures:    textures,
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

func (s *Service) profileWithTextures(ctx context.Context, profile *YggdrasilProfile, sign bool) ProfileResp {
	response := toProfileResp(profile)
	response.Properties = []Property{s.buildTexturesProperty(ctx, profile, sign)}
	return response
}

func (s *Service) textureURL(hash string) string {
	base := s.settings.TextureBaseURL
	if base == "" {
		base = "http://" + s.settings.ServerName
	}
	return fmt.Sprintf("%s/api/v1/raw/%s", base, hash)
}
