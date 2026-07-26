package router_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/cache"
	"github.com/lyimoexiao/akari/internal/config"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/permission"
	"github.com/lyimoexiao/akari/internal/rbacadapter"
	"github.com/lyimoexiao/akari/internal/role"
	"github.com/lyimoexiao/akari/internal/router"
	"github.com/lyimoexiao/akari/internal/routeradapter"
	"github.com/lyimoexiao/akari/internal/user"
	"github.com/lyimoexiao/akari/internal/useradapter"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type unavailableCache struct{}

func (unavailableCache) Get(context.Context, string, interface{}) error {
	return nil
}

func (unavailableCache) Set(context.Context, string, interface{}, time.Duration) error {
	return nil
}

func (unavailableCache) Del(context.Context, ...string) error {
	return nil
}

func (unavailableCache) Ping(context.Context) error {
	return errors.New("cache unavailable")
}

func (unavailableCache) Close() error {
	return nil
}

type healthResponse struct {
	Code    int  `json:"code"`
	Success bool `json:"success"`
	Data    struct {
		Status string `json:"status"`
	} `json:"data"`
}

func Test_Health_returns_ok_when_database_and_cache_are_available(t *testing.T) {
	// Given
	db := newHealthTestDB(t)
	engine := newHealthTestEngine(t, db, cache.New(&config.CacheConfig{}))
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	responseRecorder := httptest.NewRecorder()

	// When
	engine.ServeHTTP(responseRecorder, request)

	// Then
	if responseRecorder.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, responseRecorder.Code)
	}
	var body healthResponse
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !body.Success || body.Code != 0 || body.Data.Status != "ok" {
		t.Fatalf("expected healthy response envelope, got %+v", body)
	}
}

func Test_Health_returns_service_unavailable_when_database_is_unavailable(t *testing.T) {
	// Given
	db := newHealthTestDB(t)
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get underlying database: %v", err)
	}
	if err := sqlDB.Close(); err != nil {
		t.Fatalf("close database: %v", err)
	}
	engine := newHealthTestEngine(t, db, cache.New(&config.CacheConfig{}))
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	responseRecorder := httptest.NewRecorder()

	// When
	engine.ServeHTTP(responseRecorder, request)

	// Then
	assertServiceUnavailable(t, responseRecorder)
}

func Test_Health_returns_service_unavailable_when_cache_is_unavailable(t *testing.T) {
	// Given
	db := newHealthTestDB(t)
	engine := newHealthTestEngine(t, db, unavailableCache{})
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	responseRecorder := httptest.NewRecorder()

	// When
	engine.ServeHTTP(responseRecorder, request)

	// Then
	assertServiceUnavailable(t, responseRecorder)
}

func assertServiceUnavailable(t *testing.T, responseRecorder *httptest.ResponseRecorder) {
	t.Helper()
	if responseRecorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, responseRecorder.Code)
	}
	var body healthResponse
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Success || body.Code != -1 {
		t.Fatalf("expected unavailable response envelope, got %+v", body)
	}
}

func newHealthTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	t.Cleanup(func() {
		sqlDB, dbErr := db.DB()
		if dbErr == nil {
			_ = sqlDB.Close()
		}
	})
	return db
}

func newHealthTestEngine(t *testing.T, db *gorm.DB, cacheBackend cache.Cache) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	return router.Setup(&router.Dependencies{
		Health:   routeradapter.NewHealthChecker(db, cacheBackend),
		Handlers: &router.Handlers{},
		Logger:   zap.NewNop().Sugar(),
	})
}

func Test_Setup_registers_top_level_access_management_routes(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
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
	permissionService := permission.NewService(backend)
	permissionMiddleware := permission.NewMiddleware(permissionService)
	logger := zap.NewNop().Sugar()
	handlers := &router.Handlers{
		User: user.NewHandler(user.NewService(user.Dependencies{
			Repository: useradapter.NewRepository(db),
			Clock:      useradapter.Clock{},
			Hasher:     useradapter.PasswordHasher{},
		}), permissionMiddleware, logger),
		Role: role.NewHandler(role.NewService(role.Dependencies{
			Repository: backend, Policies: backend,
		}), permissionMiddleware, logger),
		Permission: permission.NewHandler(permissionService, permissionMiddleware, logger),
	}

	// When
	engine := router.Setup(&router.Dependencies{
		Handlers: handlers,
		Auth:     auth.NewMiddleware(nil),
		Logger:   logger,
	})
	registered := make(map[string]bool)
	for _, routeInfo := range engine.Routes() {
		registered[routeInfo.Method+" "+routeInfo.Path] = true
	}

	// Then
	expected := []string{
		"GET /api/v1/users",
		"POST /api/v1/users/verify-email",
		"POST /api/v1/users/reset-password",
		"DELETE /api/v1/users/:id",
		"GET /api/v1/roles",
		"POST /api/v1/roles",
		"PUT /api/v1/roles/:id",
		"DELETE /api/v1/roles/:id",
		"PUT /api/v1/roles/:id/default",
		"POST /api/v1/roles/assign",
		"GET /api/v1/permissions",
		"GET /api/v1/auth/permission",
	}
	for _, routeKey := range expected {
		if !registered[routeKey] {
			t.Errorf("route %q is not registered", routeKey)
		}
	}
	if registered["GET /api/v1/admin/users"] {
		t.Fatal("legacy admin route is still registered")
	}
}
