package yggdrasil

import (
	"context"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
)

type Repository interface {
	FindUserByID(context.Context, uint) (model.User, error)
	FindUserByLogin(context.Context, string) (model.User, error)
	CountValidTokens(context.Context, string) (int64, error)
	RevokeOldestValidToken(context.Context, string) error
	CreateToken(context.Context, *YggdrasilToken) error
	FindToken(context.Context, string) (YggdrasilToken, error)
	RevokeToken(context.Context, string) error
	RevokeValidTokens(context.Context, string) error
	ProfilesForUser(context.Context, string) ([]YggdrasilProfile, error)
	CreateProfile(context.Context, *YggdrasilProfile) error
	ProfileByUUID(context.Context, string) (YggdrasilProfile, error)
	ProfilesByNames(context.Context, []string) ([]YggdrasilProfile, error)
	LastLoginToken(context.Context, string) (YggdrasilToken, error)

	// Profile texture management
	UpdateProfileSkin(ctx context.Context, profileID uint, textureID *uint) error
	UpdateProfileCape(ctx context.Context, profileID uint, textureID *uint) error
	UpdateProfileModel(ctx context.Context, profileID uint, model string) error
	TextureByID(ctx context.Context, tid uint) (*model.Texture, error)
}

type ServerSession struct {
	AccessToken string `json:"access_token"`
	ProfileUUID string `json:"profile_uuid"`
	IP          string `json:"ip"`
}

type SessionStore interface {
	Save(context.Context, string, ServerSession, time.Duration) error
	Load(context.Context, string) (ServerSession, error)
}

type Signer interface {
	Sign([]byte) (string, error)
	PublicKey() string
}

type SigningFailureReporter interface {
	ReportSigningFailure(error)
}

type Settings struct {
	EmailVerificationEnabled bool
	ServerName               string
	ImplementationName       string
	ImplementationVersion    string
	TextureBaseURL           string // base URL for texture file serving, e.g. "https://example.com"
	SkinDomains              []string // extra skin domains beyond defaults
}

type Dependencies struct {
	Repository Repository
	Sessions   SessionStore
	Signer     Signer
	Reporter   SigningFailureReporter
	Settings   Settings
}
