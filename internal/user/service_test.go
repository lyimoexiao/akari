package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/user"
	"github.com/lyimoexiao/akari/internal/useradapter"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Service_rejects_non_super_admin_mutations_of_super_admin(t *testing.T) {
	// Given
	db := newUserTestDB(t)
	users := []model.User{
		{Username: "operator", Email: "operator@example.com", Password: "hash"},
		{Username: "root", Email: "root@example.com", Password: "hash"},
	}
	if err := db.Create(&users).Error; err != nil {
		t.Fatalf("create users: %v", err)
	}
	assignUserTestRole(t, db, users[0].ID, model.RoleStaff)
	assignUserTestRole(t, db, users[1].ID, model.RoleSuperAdmin)
	svc := user.NewService(user.Dependencies{
		Repository: useradapter.NewRepository(db),
		Clock:      useradapter.Clock{},
		Hasher:     useradapter.PasswordHasher{},
	})
	cases := []struct {
		name   string
		mutate func() error
	}{
		{
			name: "verify email",
			mutate: func() error {
				return svc.VerifyEmail(context.Background(), users[0].ID, user.VerifyEmailReq{UserID: users[1].ID})
			},
		},
		{
			name: "reset password",
			mutate: func() error {
				return svc.ResetPassword(context.Background(), users[0].ID, user.ResetPasswordReq{
					UserID: users[1].ID, NewPassword: "changed-password",
				})
			},
		},
		{
			name: "delete user",
			mutate: func() error {
				return svc.DeleteUser(context.Background(), user.DeleteUserCommand{
					CallerID: users[0].ID, UserID: users[1].ID,
				})
			},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			// When
			err := testCase.mutate()

			// Then
			if !errors.Is(err, user.ErrCannotModifyAdmin) {
				t.Fatalf("error = %v, want ErrCannotModifyAdmin", err)
			}
		})
	}
}

func newUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sql db: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		if closeErr := sqlDB.Close(); closeErr != nil {
			t.Errorf("close sqlite: %v", closeErr)
		}
	})
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	return db
}

func assignUserTestRole(t *testing.T, db *gorm.DB, userID uint, roleName string) {
	t.Helper()
	var stored model.Role
	if err := db.Where("name = ?", roleName).First(&stored).Error; err != nil {
		t.Fatalf("load role %s: %v", roleName, err)
	}
	if err := db.Create(&model.UserRole{UserID: userID, RoleID: stored.ID}).Error; err != nil {
		t.Fatalf("assign role %s: %v", roleName, err)
	}
}
