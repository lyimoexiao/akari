package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/authadapter"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"github.com/lyimoexiao/akari/internal/role"
	"github.com/lyimoexiao/akari/pkg/cache"
	"github.com/lyimoexiao/akari/pkg/jwt"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Register_uses_configured_default_role(t *testing.T) {
	// Given
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	backend, err := rbacadapter.NewManager(db)
	if err != nil {
		t.Fatalf("create RBAC manager: %v", err)
	}
	roles := role.NewService(role.Dependencies{Repository: backend, Policies: backend})
	cacheService := cache.New(&cache.Config{Type: "memory"})
	t.Cleanup(func() { _ = cacheService.Close() })
	svc := auth.NewService(auth.Dependencies{
		Users:  authadapter.NewUserRepository(db),
		Roles:  roles,
		Tokens: authadapter.NewJWTManager(jwt.New(&jwt.Config{Secret: "test-secret", Issuer: "test", Expiration: time.Hour})),
		Store:  authadapter.NewTokenStore(cacheService, zap.NewNop().Sugar()),
		Settings: auth.Settings{
			RegistrationEnabled: true,
		},
	})
	root, err := svc.Register(context.Background(), &auth.RegisterReq{
		Username: "root", Email: "root@example.com", Password: "Secret123",
	})
	if err != nil {
		t.Fatalf("register root: %v", err)
	}
	custom, err := roles.Create(context.Background(), root.User.ID, role.CreateReq{Name: "member", Description: "注册用户"})
	if err != nil {
		t.Fatalf("create custom role: %v", err)
	}
	if err := roles.SetDefault(context.Background(), root.User.ID, custom.ID); err != nil {
		t.Fatalf("set default role: %v", err)
	}

	// When
	registered, err := svc.Register(context.Background(), &auth.RegisterReq{
		Username: "new-member", Email: "new-member@example.com", Password: "Secret123",
	})

	// Then
	if err != nil {
		t.Fatalf("register member: %v", err)
	}
	if registered.User.Role != custom.Name {
		t.Fatalf("role = %q, want %q", registered.User.Role, custom.Name)
	}
	var assignment model.UserRole
	if err := db.Where("user_id = ? AND role_id = ?", registered.User.ID, custom.ID).First(&assignment).Error; err != nil {
		t.Fatalf("load default role assignment: %v", err)
	}
}
