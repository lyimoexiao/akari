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

	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
	"github.com/lyimoexiao/akari/pkg/version"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ───────────────────────────── Errors ─────────────────────────────

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token expired")
	ErrProfileNotFound    = errors.New("profile not found")
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailNotVerified   = errors.New("email not verified")
	ErrTokenLimitReached  = errors.New("token limit reached")
)

// ───────────────────────────── Types ─────────────────────────────

// serverSession is stored temporarily in cache for join/hasJoined flow.
type serverSession struct {
	AccessToken string `json:"access_token"`
	ProfileUUID string `json:"profile_uuid"`
	IP          string `json:"ip"`
}

const (
	tokenExpiration     = 15 * 24 * time.Hour // 15 days
	maxTokensPerUser    = 10
	serverSessionTTL    = 30 * time.Second
	cacheKeyJoin        = "yggdrasil:join:"
	userUUIDPrefix      = "YggdrasilUser:"
	offlinePlayerPrefix = "OfflinePlayer:"
)

// ───────────────────────────── Service ─────────────────────────────

type Service struct {
	db           *gorm.DB
	cache        cache.Cache
	authCfg      *config.AuthConfig
	logger       *zap.SugaredLogger
	km           *KeyManager // nil = signing disabled
	yggdrasilCfg *config.YggdrasilConfig
}

func NewService(db *gorm.DB, c cache.Cache, authCfg *config.AuthConfig, logger *zap.SugaredLogger, km *KeyManager, yggdrasilCfg *config.YggdrasilConfig) *Service {
	return &Service{db: db, cache: c, authCfg: authCfg, logger: logger, km: km, yggdrasilCfg: yggdrasilCfg}
}

// ── UUID helpers ──

// OfflineUUID computes the offline-compatible UUID v3 for a profile name.
// Equivalent to Java's UUID.nameUUIDFromBytes(("OfflinePlayer:" + name).getBytes(UTF_8)).
func OfflineUUID(name string) string {
	h := md5.Sum([]byte(offlinePlayerPrefix + name))
	h[6] = (h[6] & 0x0f) | 0x30 // version 3
	h[8] = (h[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", h[0:4], h[4:6], h[6:8], h[8:10], h[10:])
}

// userUUID computes a deterministic UUID v3 from an email.
func userUUID(email string) string {
	h := md5.Sum([]byte(userUUIDPrefix + email))
	h[6] = (h[6] & 0x0f) | 0x30
	h[8] = (h[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", h[0:4], h[4:6], h[6:8], h[8:10], h[10:])
}

// generateAccessToken generates a random UUID v4 string.
func generateAccessToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant RFC 4122
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// formatUUIDWithoutDashes returns the unsigned (no-dash) UUID string.
func formatUUIDWithoutDashes(uuid string) string {
	if len(uuid) == 36 {
		return uuid[0:8] + uuid[9:13] + uuid[14:18] + uuid[19:23] + uuid[24:36]
	}
	return uuid
}

// ── authenticate user via model.User ──

func (s *Service) findUserByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return &user, nil
}

func (s *Service) findUserByLogin(ctx context.Context, login string) (*model.User, error) {
	var user model.User
	// Try email first, then username
	err := s.db.WithContext(ctx).Where("email = ? OR username = ?", login, login).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("find user: %w", err)
	}
	return &user, nil
}

// ── Token helpers ──

func (s *Service) createToken(ctx context.Context, userEmail, clientToken, profileUUID, loginIP string) (*YggdrasilToken, error) {
	// Enforce token limit
	var count int64
	s.db.WithContext(ctx).Model(&YggdrasilToken{}).Where("user_email = ? AND status = ?", userEmail, TokenStatusValid).Count(&count)
	if int(count) >= maxTokensPerUser {
		// Revoke oldest token
		var oldest YggdrasilToken
		if err := s.db.WithContext(ctx).Where("user_email = ? AND status = ?", userEmail, TokenStatusValid).
			Order("issued_at ASC").First(&oldest).Error; err == nil {
			s.db.WithContext(ctx).Model(&oldest).Update("status", TokenStatusInvalid)
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
	if err := s.db.WithContext(ctx).Create(token).Error; err != nil {
		return nil, fmt.Errorf("create token: %w", err)
	}
	return token, nil
}

func (s *Service) findValidToken(ctx context.Context, accessToken, clientToken string) (*YggdrasilToken, error) {
	var t YggdrasilToken
	if err := s.db.WithContext(ctx).Where("access_token = ?", accessToken).First(&t).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInvalidToken
		}
		return nil, fmt.Errorf("find token: %w", err)
	}
	// Check clientToken if provided
	if clientToken != "" && t.ClientToken != clientToken {
		return nil, ErrInvalidToken
	}
	// Check status
	if t.Status == TokenStatusInvalid {
		return nil, ErrInvalidToken
	}
	// Check expiration
	if time.Now().After(t.ExpiresAt) {
		// Auto-clean expired tokens
		s.db.WithContext(ctx).Model(&t).Update("status", TokenStatusInvalid)
		return nil, ErrTokenExpired
	}
	return &t, nil
}

func (s *Service) revokeToken(ctx context.Context, token *YggdrasilToken) error {
	return s.db.WithContext(ctx).Model(token).Update("status", TokenStatusInvalid).Error
}

// ── Profile helpers ──

func (s *Service) getProfilesForUser(ctx context.Context, email string) ([]YggdrasilProfile, error) {
	var profiles []YggdrasilProfile
	if err := s.db.WithContext(ctx).Where("user_email = ?", email).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// ensureProfile creates a default profile for the user if none exists.
func (s *Service) ensureProfile(ctx context.Context, userEmail, username string) (*YggdrasilProfile, error) {
	profiles, err := s.getProfilesForUser(ctx, userEmail)
	if err != nil {
		return nil, err
	}
	if len(profiles) > 0 {
		return &profiles[0], nil
	}
	// Auto-create: profile name = existing model.User.username
	uuid := OfflineUUID(username)
	profile := &YggdrasilProfile{
		UUID:      uuid,
		Name:      username,
		UserEmail: userEmail,
		Model:     "default",
	}
	if err := s.db.WithContext(ctx).Create(profile).Error; err != nil {
		return nil, fmt.Errorf("create profile: %w", err)
	}
	return profile, nil
}

func (s *Service) getProfileByUUID(ctx context.Context, uuid string) (*YggdrasilProfile, error) {
	var p YggdrasilProfile
	if err := s.db.WithContext(ctx).Where("uuid = ?", uuid).First(&p).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProfileNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (s *Service) getProfilesByNames(ctx context.Context, names []string) ([]YggdrasilProfile, error) {
	var profiles []YggdrasilProfile
	if err := s.db.WithContext(ctx).Where("name IN ?", names).Find(&profiles).Error; err != nil {
		return nil, err
	}
	return profiles, nil
}

// toProfileResp converts a YggdrasilProfile to the API response format (no properties).
func toProfileResp(p *YggdrasilProfile) ProfileResp {
	return ProfileResp{
		ID:   formatUUIDWithoutDashes(p.UUID),
		Name: p.Name,
	}
}

// texturesPayload is the JSON structure inside the Base64-encoded textures property.
type texturesPayload struct {
	Timestamp   int64          `json:"timestamp"`
	ProfileID   string         `json:"profileId"`
	ProfileName string         `json:"profileName"`
	Textures    map[string]any `json:"textures"`
}

// buildTexturesProperty builds the textures Property for a profile.
// When sign is true and KeyManager is available, the property value is signed.
func (s *Service) buildTexturesProperty(profileID, profileName string, sign bool) Property {
	id := formatUUIDWithoutDashes(profileID)
	payload := texturesPayload{
		Timestamp:   time.Now().UnixMilli(),
		ProfileID:   id,
		ProfileName: profileName,
		Textures:    map[string]any{}, // ponytail: textures not implemented yet
	}
	raw, _ := json.Marshal(payload)
	value := base64.StdEncoding.EncodeToString(raw)

	prop := Property{Name: "textures", Value: value}

	if sign && s.km != nil {
		sig, err := s.km.SignSHA1([]byte(value))
		if err != nil {
			s.logger.Warnw("failed to sign textures property", "error", err)
		} else {
			prop.Signature = sig
		}
	}

	return prop
}

// profileWithTextures returns a ProfileResp with the textures property.
// sign indicates whether to include a signature.
func (s *Service) profileWithTextures(p *YggdrasilProfile, sign bool) ProfileResp {
	resp := toProfileResp(p)
	resp.Properties = []Property{
		s.buildTexturesProperty(p.UUID, p.Name, sign),
	}
	return resp
}

// ── Business methods ──

// Authenticate handles POST /authserver/authenticate.
func (s *Service) Authenticate(ctx context.Context, req *AuthenticateReq, loginIP string) (*AuthenticateResp, error) {
	// Find user by email or username
	user, err := s.findUserByLogin(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Check password
	if !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	// Check email verification if enabled
	if s.authCfg.EmailVerificationEnabled {
		if !user.IsEmailVerified() {
			return nil, ErrEmailNotVerified
		}
	}

	// Ensure profile exists
	profile, err := s.ensureProfile(ctx, user.Email, user.Username)
	if err != nil {
		return nil, fmt.Errorf("ensure profile: %w", err)
	}

	// Determine clientToken
	clientToken := req.ClientToken
	if clientToken == "" {
		clientToken = generateAccessToken()
	}

	// Determine selected profile: if only one profile → bind; multiple → leave empty
	profiles, _ := s.getProfilesForUser(ctx, user.Email)
	profileUUID := ""
	if len(profiles) == 1 {
		profileUUID = profile.UUID
	}

	token, err := s.createToken(ctx, user.Email, clientToken, profileUUID, loginIP)
	if err != nil {
		return nil, err
	}

	// Build available profiles
	available := make([]ProfileResp, len(profiles))
	for i, p := range profiles {
		available[i] = toProfileResp(&p)
	}

	resp := &AuthenticateResp{
		AccessToken:       token.AccessToken,
		ClientToken:       token.ClientToken,
		AvailableProfiles: available,
	}

	// Include selected profile if bound
	if profileUUID != "" {
		sp := toProfileResp(profile)
		resp.SelectedProfile = &sp
	}

	// Include user info if requested
	if req.RequestUser {
		resp.User = &YggdrasilUserResp{
			ID: userUUID(user.Email),
		}
	}

	return resp, nil
}

// Refresh handles POST /authserver/refresh.
func (s *Service) Refresh(ctx context.Context, req *RefreshReq) (*RefreshResp, error) {
	// Validate the original token (temp_invalid is allowed for refresh)
	var oldToken YggdrasilToken
	if err := s.db.WithContext(ctx).Where("access_token = ?", req.AccessToken).First(&oldToken).Error; err != nil {
		return nil, ErrInvalidToken
	}
	if req.ClientToken != "" && oldToken.ClientToken != req.ClientToken {
		return nil, ErrInvalidToken
	}
	if oldToken.Status == TokenStatusInvalid {
		return nil, ErrInvalidToken
	}
	if time.Now().After(oldToken.ExpiresAt) {
		s.db.WithContext(ctx).Model(&oldToken).Update("status", TokenStatusInvalid)
		return nil, ErrTokenExpired
	}

	// Determine which profile to bind
	profileUUID := oldToken.ProfileUUID
	if req.SelectedProfile != nil {
		// Validate the selected profile exists and belongs to the same user
		rawUUID := req.SelectedProfile.ID
		if len(rawUUID) == 32 {
			// Convert unsigned UUID back to standard format
			rawUUID = rawUUID[0:8] + "-" + rawUUID[8:12] + "-" + rawUUID[12:16] + "-" + rawUUID[16:20] + "-" + rawUUID[20:32]
		}
		prof, err := s.getProfileByUUID(ctx, rawUUID)
		if err != nil {
			return nil, ErrInvalidToken
		}
		if prof.UserEmail != oldToken.UserEmail {
			return nil, ErrInvalidToken
		}
		profileUUID = rawUUID
	}

	// Revoke old token
	s.revokeToken(ctx, &oldToken)

	// Create new token (don't track refresh IP)
	newToken, err := s.createToken(ctx, oldToken.UserEmail, oldToken.ClientToken, profileUUID, "")
	if err != nil {
		return nil, err
	}

	resp := &RefreshResp{
		AccessToken: newToken.AccessToken,
		ClientToken: newToken.ClientToken,
	}

	// Include selected profile
	if profileUUID != "" {
		prof, err := s.getProfileByUUID(ctx, profileUUID)
		if err == nil {
			sp := toProfileResp(prof)
			resp.SelectedProfile = &sp
		}
	}

	// Include user info
	if req.RequestUser {
		resp.User = &YggdrasilUserResp{
			ID: userUUID(oldToken.UserEmail),
		}
	}

	return resp, nil
}

// Validate handles POST /authserver/validate.
func (s *Service) Validate(ctx context.Context, req *ValidateReq) error {
	_, err := s.findValidToken(ctx, req.AccessToken, req.ClientToken)
	return err
}

// Invalidate handles POST /authserver/invalidate.
func (s *Service) Invalidate(ctx context.Context, req *InvalidateReq) error {
	var t YggdrasilToken
	if err := s.db.WithContext(ctx).Where("access_token = ?", req.AccessToken).First(&t).Error; err != nil {
		// Per spec: always return 204 regardless
		return nil
	}
	return s.revokeToken(ctx, &t)
}

// Signout handles POST /authserver/signout.
func (s *Service) Signout(ctx context.Context, req *SignoutReq) error {
	user, err := s.findUserByLogin(ctx, req.Username)
	if err != nil {
		return nil // return 204 regardless
	}
	if !bcrypt.CheckPassword(req.Password, user.Password) {
		return nil
	}
	// Revoke all valid tokens for this user
	return s.db.WithContext(ctx).Model(&YggdrasilToken{}).
		Where("user_email = ? AND status = ?", user.Email, TokenStatusValid).
		Update("status", TokenStatusInvalid).Error
}

// JoinServer handles POST /sessionserver/session/minecraft/join.
func (s *Service) JoinServer(ctx context.Context, req *JoinReq, clientIP string) error {
	token, err := s.findValidToken(ctx, req.AccessToken, "")
	if err != nil {
		return err
	}
	// Verify the selected profile matches the token binding
	if token.ProfileUUID == "" {
		// Token must be bound to a profile
		return ErrInvalidToken
	}
	// selectedProfile is unsigned UUID — normalise
	profileUUID := req.SelectedProfile
	if len(profileUUID) == 32 {
		profileUUID = profileUUID[0:8] + "-" + profileUUID[8:12] + "-" + profileUUID[12:16] + "-" + profileUUID[16:20] + "-" + profileUUID[20:32]
	}
	if profileUUID != token.ProfileUUID {
		return ErrInvalidToken
	}

	// Store in cache with ~30s TTL
	session := serverSession{
		AccessToken: req.AccessToken,
		ProfileUUID: token.ProfileUUID,
		IP:          clientIP,
	}
	if err := s.cache.Set(ctx, cacheKeyJoin+req.ServerID, session, serverSessionTTL); err != nil {
		return fmt.Errorf("cache join session: %w", err)
	}
	return nil
}

// HasJoined handles GET /sessionserver/session/minecraft/hasJoined.
func (s *Service) HasJoined(ctx context.Context, username, serverID, ip string) (*ProfileResp, error) {
	// Look up the session from cache
	var session serverSession
	if err := s.cache.Get(ctx, cacheKeyJoin+serverID, &session); err != nil {
		return nil, ErrInvalidToken
	}

	// Verify the profile name matches
	prof, err := s.getProfileByUUID(ctx, session.ProfileUUID)
	if err != nil {
		return nil, ErrProfileNotFound
	}
	if prof.Name != username {
		return nil, ErrInvalidToken
	}

	// If IP enforcement is enabled, check it
	if ip != "" && session.IP != "" && session.IP != ip {
		return nil, ErrInvalidToken
	}

	// hasJoined always returns full profile with signed properties
	resp := s.profileWithTextures(prof, true)
	return &resp, nil
}

// GetProfile handles GET /sessionserver/session/minecraft/profile/{uuid}.
func (s *Service) GetProfile(ctx context.Context, uuid string, unsigned bool) (*ProfileResp, error) {
	// uuid may be unsigned (32 chars) or standard (36 chars)
	normUUID := uuid
	if len(normUUID) == 32 {
		normUUID = normUUID[0:8] + "-" + normUUID[8:12] + "-" + normUUID[12:16] + "-" + normUUID[16:20] + "-" + normUUID[20:32]
	}

	prof, err := s.getProfileByUUID(ctx, normUUID)
	if err != nil {
		return nil, err
	}

	// unsigned=true (default): no signature; unsigned=false: include signature
	resp := s.profileWithTextures(prof, !unsigned)
	return &resp, nil
}

// GetProfilesByName handles POST /api/profiles/minecraft.
func (s *Service) GetProfilesByName(ctx context.Context, names []string) ([]ProfileResp, error) {
	profiles, err := s.getProfilesByNames(ctx, names)
	if err != nil {
		return nil, err
	}
	result := make([]ProfileResp, len(profiles))
	for i, p := range profiles {
		result[i] = toProfileResp(&p)
	}
	return result, nil
}

// UserStatus returns Yggdrasil binding info for the given app user ID.
func (s *Service) UserStatus(ctx context.Context, userID uint) (*UserStatusResp, error) {
	user, err := s.findUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	profiles, err := s.getProfilesForUser(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	resp := &UserStatusResp{}

	if len(profiles) > 0 {
		p := profiles[0]
		resp.HasProfile = true
		resp.ProfileUUID = p.UUID
		resp.ProfileName = p.Name
	}

	var lastToken YggdrasilToken
	if err := s.db.WithContext(ctx).Where("user_email = ? AND login_ip != ''", user.Email).
		Order("issued_at DESC").First(&lastToken).Error; err == nil {
		t := lastToken.IssuedAt.Format(time.RFC3339)
		resp.LastLoginAt = &t
		resp.LastLoginIP = lastToken.LoginIP
	}

	return resp, nil
}

// Metadata returns API metadata for GET /.
func (s *Service) Metadata() *MetadataResp {
	pubKey := ""
	if s.km != nil {
		pubKey = s.km.PublicKeyPEM()
	}
	meta := map[string]any{
		"serverName":              s.yggdrasilCfg.ServerName,
		"implementationName":      s.yggdrasilCfg.ImplementationName,
		"implementationVersion":   version.Version,
		"feature.non_email_login": true,
	}
	return &MetadataResp{
		Meta:               meta,
		SkinDomains:        []string{},
		SignaturePublickey: pubKey,
	}
}
