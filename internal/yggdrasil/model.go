package yggdrasil

import "time"

// ───────────────────────────── GORM Models ─────────────────────────────

// YggdrasilProfile maps a Minecraft profile (角色) to a user.
// UUID is offline-compatible: MD5("OfflinePlayer:" + name).
type YggdrasilProfile struct {
	ID            uint      `gorm:"primaryKey" json:"-"`
	UUID          string    `gorm:"uniqueIndex;size:36;not null" json:"uuid"`
	Name          string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	UserEmail     string    `gorm:"index;size:255;not null" json:"user_email"`
	Model         string    `gorm:"size:16;default:default" json:"model"` // default or slim
	TextureSkinID *uint     `gorm:"index" json:"texture_skin_id,omitempty"`
	TextureCapeID *uint     `gorm:"index" json:"texture_cape_id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// YggdrasilToken stores an access token for Yggdrasil auth.
type YggdrasilToken struct {
	ID          uint      `gorm:"primaryKey" json:"-"`
	AccessToken string    `gorm:"uniqueIndex;size:64;not null" json:"access_token"`
	ClientToken string    `gorm:"size:64;not null" json:"client_token"`
	UserEmail   string    `gorm:"index;size:255;not null" json:"user_email"`
	ProfileUUID string    `gorm:"size:36;index" json:"profile_uuid"`
	Status      string    `gorm:"size:16;not null;default:valid" json:"status"` // valid, temp_invalid, invalid
	LoginIP     string    `gorm:"size:45" json:"login_ip"`
	IssuedAt    time.Time `gorm:"not null" json:"issued_at"`
	ExpiresAt   time.Time `gorm:"not null" json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

const (
	TokenStatusValid       = "valid"
	TokenStatusTempInvalid = "temp_invalid"
	TokenStatusInvalid     = "invalid"
)

// ───────────────────────────── Request / Response ─────────────────────────────

// Yggdrasil error response per spec.
type ErrorResp struct {
	Error        string `json:"error"`
	ErrorMessage string `json:"errorMessage"`
	Cause        string `json:"cause,omitempty"`
}

// ── Authenticate (login) ──

type AuthenticateReq struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	ClientToken string `json:"clientToken,omitempty"`
	RequestUser bool   `json:"requestUser,omitempty"`
	Agent       *Agent `json:"agent,omitempty"`
}

type Agent struct {
	Name    string `json:"name"`
	Version int    `json:"version"`
}

type ProfileResp struct {
	ID         string     `json:"id"`
	Name       string     `json:"name"`
	Properties []Property `json:"properties,omitempty"`
}

type Property struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Signature string `json:"signature,omitempty"`
}

type YggdrasilUserResp struct {
	ID         string     `json:"id"`
	Properties []Property `json:"properties"`
}

type AuthenticateResp struct {
	AccessToken       string             `json:"accessToken"`
	ClientToken       string             `json:"clientToken"`
	AvailableProfiles []ProfileResp      `json:"availableProfiles"`
	SelectedProfile   *ProfileResp       `json:"selectedProfile,omitempty"`
	User              *YggdrasilUserResp `json:"user,omitempty"`
}

// ── Refresh ──

type RefreshReq struct {
	AccessToken     string       `json:"accessToken"`
	ClientToken     string       `json:"clientToken,omitempty"`
	RequestUser     bool         `json:"requestUser,omitempty"`
	SelectedProfile *ProfileResp `json:"selectedProfile,omitempty"`
}

type RefreshResp struct {
	AccessToken     string             `json:"accessToken"`
	ClientToken     string             `json:"clientToken"`
	SelectedProfile *ProfileResp       `json:"selectedProfile,omitempty"`
	User            *YggdrasilUserResp `json:"user,omitempty"`
}

// ── Validate ──

type ValidateReq struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken,omitempty"`
}

// ── Invalidate ──

type InvalidateReq struct {
	AccessToken string `json:"accessToken"`
	ClientToken string `json:"clientToken,omitempty"`
}

// ── Signout ──

type SignoutReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ── Join Server ──

type JoinReq struct {
	AccessToken     string `json:"accessToken"`
	SelectedProfile string `json:"selectedProfile"`
	ServerID        string `json:"serverId"`
}

// ── HasJoined ──
// (query params: username, serverId, ip)

// ── Profile ──
// (path param: uuid, query: unsigned)

// ── Profiles batch ──
// (body: []string)

// ── User Status (web UI) ──

type UserStatusResp struct {
	HasProfile    bool    `json:"has_profile"`
	ProfileUUID   string  `json:"profile_uuid,omitempty"`
	ProfileName   string  `json:"profile_name,omitempty"`
	TextureSkinID *uint   `json:"texture_skin_id,omitempty"`
	TextureCapeID *uint   `json:"texture_cape_id,omitempty"`
	LastLoginAt   *string `json:"last_login_at,omitempty"`
	LastLoginIP   string  `json:"last_login_ip,omitempty"`
}

// ── Profile texture management (app API) ──

type SetSkinReq struct {
	TextureTID uint `json:"texture_tid"`
}

type SetCapeReq struct {
	TextureTID uint `json:"texture_tid"`
}

type ClearCapeReq struct{}

// ── API Metadata ──

type MetadataResp struct {
	Meta               map[string]any `json:"meta"`
	SkinDomains        []string       `json:"skinDomains"`
	SignaturePublickey string         `json:"signaturePublickey"`
	// 站点公网地址（server.base_url），供前端展示 Yggdrasil API 地址
	BaseURL string `json:"base_url,omitempty"`
}
