package router_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/database"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/requestlog"
	"github.com/lyimoexiao/akari/internal/requestlogadapter"
	"github.com/lyimoexiao/akari/internal/router"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Setup_records_request_id_for_API_but_not_static_assets(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	logger := zap.NewNop().Sugar()
	repository := requestlogadapter.NewRepository(db)
	engine := router.Setup(&router.Dependencies{
		RequestLogging: requestlog.Middleware(repository, logger),
		Handlers:       &router.Handlers{},
		Logger:         logger,
	})

	// When
	apiRecorder := httptest.NewRecorder()
	engine.ServeHTTP(apiRecorder, httptest.NewRequest(http.MethodGet, "/api/v1/missing", nil))

	// Then
	requestID := apiRecorder.Header().Get("X-Request-ID")
	if requestID == "" {
		t.Fatal("API response is missing X-Request-ID")
	}
	var envelope struct {
		RequestID string `json:"request_id"`
	}
	if err := json.Unmarshal(apiRecorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode API response: %v", err)
	}
	if envelope.RequestID != requestID {
		t.Fatalf("response request_id = %q, want %q", envelope.RequestID, requestID)
	}
	var persisted model.RequestLog
	if err := db.Where("request_id = ?", requestID).First(&persisted).Error; err != nil {
		t.Fatalf("load persisted request: %v", err)
	}

	// When
	staticRecorder := httptest.NewRecorder()
	engine.ServeHTTP(staticRecorder, httptest.NewRequest(http.MethodGet, "/assets/app.js", nil))

	// Then
	if staticRecorder.Header().Get("X-Request-ID") != "" {
		t.Fatal("static response unexpectedly contains X-Request-ID")
	}
	var count int64
	if err := db.Model(&model.RequestLog{}).Count(&count).Error; err != nil {
		t.Fatalf("count request logs: %v", err)
	}
	if count != 1 {
		t.Fatalf("request log count = %d, want 1", count)
	}
}
