package auth

import (
	"context"
	"testing"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/bcrypt"
)

type fakeUserRepository struct {
	user model.User
}

func (f *fakeUserRepository) Count(context.Context) (int64, error) {
	return 1, nil
}

func (f *fakeUserRepository) PrepareRegistration(context.Context, string, string) error {
	return nil
}

func (f *fakeUserRepository) Create(context.Context, *model.User, *model.Role) error {
	return nil
}

func (f *fakeUserRepository) FindByLogin(context.Context, string) (model.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) FindByID(context.Context, uint) (model.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) FindByEmail(context.Context, string) (model.User, error) {
	return f.user, nil
}

func (f *fakeUserRepository) MarkEmailVerified(context.Context, uint, time.Time) error {
	return nil
}

func (f *fakeUserRepository) UpdatePassword(context.Context, uint, string) error {
	return nil
}

type fakeTokenManager struct{}

func (fakeTokenManager) GenerateToken(uint, string, string) (string, error) {
	return "test-token", nil
}

func (fakeTokenManager) ValidateToken(string) (TokenClaims, error) {
	return TokenClaims{}, nil
}

func Test_Service_Login_uses_ports_without_infrastructure(t *testing.T) {
	// Given
	password, err := bcrypt.HashPassword("Secret123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users := &fakeUserRepository{user: model.User{
		ID:       7,
		Username: "alice",
		Password: password,
		Roles:    []model.Role{{Name: model.RoleUser}},
	}}
	service := NewService(Dependencies{
		Users:    users,
		Tokens:   fakeTokenManager{},
		Settings: Settings{RegistrationEnabled: true},
	})

	// When
	response, err := service.Login(context.Background(), &LoginReq{
		Username: "alice",
		Password: "Secret123",
	})

	// Then
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if response.Token != "test-token" {
		t.Fatalf("token = %q, want test-token", response.Token)
	}
}
