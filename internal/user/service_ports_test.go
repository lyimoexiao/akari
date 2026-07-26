package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/user"
)

type fakeUserRepository struct {
	target         model.User
	findErr        error
	findCalls      int
	verifiedAt     time.Time
	password       string
	passwordWrites int
}

func (*fakeUserRepository) List(context.Context, user.ListQuery) ([]model.User, int64, error) {
	return nil, 0, nil
}

func (repository *fakeUserRepository) FindByID(context.Context, uint) (model.User, error) {
	repository.findCalls++
	return repository.target, repository.findErr
}

func (*fakeUserRepository) HasRole(context.Context, uint, string) (bool, error) {
	return false, nil
}

func (repository *fakeUserRepository) VerifyEmail(_ context.Context, _ uint, verifiedAt time.Time) error {
	repository.verifiedAt = verifiedAt
	return nil
}

func (repository *fakeUserRepository) UpdatePassword(_ context.Context, _ uint, password string) error {
	repository.passwordWrites++
	repository.password = password
	return nil
}

func (*fakeUserRepository) Delete(context.Context, uint) error { return nil }

type fixedClock struct {
	now time.Time
}

func (clock fixedClock) Now() time.Time { return clock.now }

type fakePasswordHasher struct{}

func (fakePasswordHasher) Hash(string) (string, error) { return "hashed-password", nil }

type failingPasswordHasher struct {
	err error
}

func (hasher failingPasswordHasher) Hash(string) (string, error) { return "", hasher.err }

func Test_Service_uses_clock_and_password_hasher_ports(t *testing.T) {
	// Given
	now := time.Date(2026, time.July, 26, 10, 30, 0, 0, time.UTC)
	repository := &fakeUserRepository{target: model.User{ID: 8, Username: "member"}}
	service := user.NewService(user.Dependencies{
		Repository: repository,
		Clock:      fixedClock{now: now},
		Hasher:     fakePasswordHasher{},
	})

	// When
	if err := service.VerifyEmail(t.Context(), 1, user.VerifyEmailReq{UserID: 8}); err != nil {
		t.Fatalf("verify email: %v", err)
	}
	if err := service.ResetPassword(t.Context(), 1, user.ResetPasswordReq{UserID: 8, NewPassword: "secret"}); err != nil {
		t.Fatalf("reset password: %v", err)
	}

	// Then
	if !repository.verifiedAt.Equal(now) {
		t.Fatalf("verified_at = %v, want %v", repository.verifiedAt, now)
	}
	if repository.password != "hashed-password" {
		t.Fatalf("password = %q, want hashed-password", repository.password)
	}
}

func Test_Service_stops_when_password_hasher_fails(t *testing.T) {
	// Given
	hashErr := errors.New("hasher unavailable")
	repository := &fakeUserRepository{target: model.User{ID: 8, Username: "member"}}
	service := user.NewService(user.Dependencies{
		Repository: repository,
		Clock:      fixedClock{},
		Hasher:     failingPasswordHasher{err: hashErr},
	})

	// When
	err := service.ResetPassword(t.Context(), 1, user.ResetPasswordReq{UserID: 8, NewPassword: "secret"})

	// Then
	if !errors.Is(err, hashErr) {
		t.Fatalf("reset password error = %v, want wrapped hasher error", err)
	}
	if repository.passwordWrites != 0 {
		t.Fatalf("password writes = %d, want 0", repository.passwordWrites)
	}
}

func Test_Service_rejects_self_delete_before_repository_lookup(t *testing.T) {
	// Given
	repository := &fakeUserRepository{}
	service := user.NewService(user.Dependencies{
		Repository: repository,
		Clock:      fixedClock{},
		Hasher:     fakePasswordHasher{},
	})

	// When
	err := service.DeleteUser(t.Context(), user.DeleteUserCommand{CallerID: 8, UserID: 8})

	// Then
	if !errors.Is(err, user.ErrCannotDeleteSelf) {
		t.Fatalf("delete user error = %v, want ErrCannotDeleteSelf", err)
	}
	if repository.findCalls != 0 {
		t.Fatalf("repository lookups = %d, want 0", repository.findCalls)
	}
}

func Test_Service_preserves_repository_not_found_error(t *testing.T) {
	// Given
	repository := &fakeUserRepository{findErr: user.ErrUserNotFound}
	service := user.NewService(user.Dependencies{
		Repository: repository,
		Clock:      fixedClock{},
		Hasher:     fakePasswordHasher{},
	})

	// When
	err := service.VerifyEmail(t.Context(), 1, user.VerifyEmailReq{UserID: 99})

	// Then
	if !errors.Is(err, user.ErrUserNotFound) {
		t.Fatalf("verify email error = %v, want ErrUserNotFound", err)
	}
}
