package yggdrasil

import (
	"context"
	"testing"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

type fakeSigner struct{}

func (fakeSigner) Sign([]byte) (string, error) {
	return "signature", nil
}

func (fakeSigner) PublicKey() string {
	return "public-key"
}

func Test_Service_Metadata_uses_values_and_signer_port(t *testing.T) {
	// Given
	service := NewService(Dependencies{
		Signer: fakeSigner{},
		Settings: Settings{
			ServerName:            "Akari",
			ImplementationName:    "akari",
			ImplementationVersion: "test-version",
		},
	})

	// When
	metadata := service.Metadata()

	// Then
	if metadata.SignaturePublickey != "public-key" {
		t.Fatalf("public key = %q, want public-key", metadata.SignaturePublickey)
	}
	if metadata.Meta["implementationVersion"] != "test-version" {
		t.Fatalf("implementation version = %v, want test-version", metadata.Meta["implementationVersion"])
	}
}

type fakeRepository struct {
	user         model.User
	profiles     []YggdrasilProfile
	createdToken *YggdrasilToken
}

func (f *fakeRepository) FindUserByID(context.Context, uint) (model.User, error) {
	return f.user, nil
}

func (f *fakeRepository) FindUserByLogin(context.Context, string) (model.User, error) {
	return f.user, nil
}

func (f *fakeRepository) CountValidTokens(context.Context, string) (int64, error) {
	return 0, nil
}

func (f *fakeRepository) RevokeOldestValidToken(context.Context, string) error {
	return nil
}

func (f *fakeRepository) CreateToken(_ context.Context, token *YggdrasilToken) error {
	f.createdToken = token
	return nil
}

func (f *fakeRepository) FindToken(context.Context, string) (YggdrasilToken, error) {
	return YggdrasilToken{}, ErrInvalidToken
}

func (f *fakeRepository) RevokeToken(context.Context, string) error {
	return nil
}

func (f *fakeRepository) RevokeValidTokens(context.Context, string) error {
	return nil
}

func (f *fakeRepository) ProfilesForUser(context.Context, string) ([]YggdrasilProfile, error) {
	return f.profiles, nil
}

func (f *fakeRepository) CreateProfile(_ context.Context, profile *YggdrasilProfile) error {
	f.profiles = []YggdrasilProfile{*profile}
	return nil
}

func (f *fakeRepository) ProfileByUUID(context.Context, string) (YggdrasilProfile, error) {
	return f.profiles[0], nil
}

func (f *fakeRepository) ProfilesByNames(context.Context, []string) ([]YggdrasilProfile, error) {
	return f.profiles, nil
}

func (f *fakeRepository) LastLoginToken(context.Context, string) (YggdrasilToken, error) {
	return YggdrasilToken{}, ErrInvalidToken
}

func (f *fakeRepository) UpdateProfileSkin(_ context.Context, _ uint, _ *uint) error {
	return nil
}

func (f *fakeRepository) UpdateProfileCape(_ context.Context, _ uint, _ *uint) error {
	return nil
}

func (f *fakeRepository) UpdateProfileModel(_ context.Context, _ uint, _ string) error {
	return nil
}

func (f *fakeRepository) TextureByID(_ context.Context, _ uint) (*model.Texture, error) {
	return nil, nil
}

func Test_Service_Authenticate_uses_repository_port_without_database(t *testing.T) {
	// Given
	password, err := bcrypt.HashPassword("Secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	repository := &fakeRepository{
		user: model.User{
			Username: "alice",
			Email:    "alice@example.com",
			Password: password,
		},
		profiles: []YggdrasilProfile{{
			UUID:      OfflineUUID("alice"),
			Name:      "alice",
			UserEmail: "alice@example.com",
		}},
	}
	service := NewService(Dependencies{Repository: repository})

	// When
	response, err := service.Authenticate(context.Background(), &AuthenticateReq{
		Username: "alice",
		Password: "Secret123",
	}, "127.0.0.1")

	// Then
	if err != nil {
		t.Fatalf("authenticate: %v", err)
	}
	if response.SelectedProfile == nil || response.SelectedProfile.Name != "alice" {
		t.Fatalf("selected profile = %#v, want alice", response.SelectedProfile)
	}
	if repository.createdToken == nil || repository.createdToken.UserEmail != "alice@example.com" {
		t.Fatalf("created token = %#v, want alice@example.com", repository.createdToken)
	}
}
