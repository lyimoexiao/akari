package requestlog_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/auth"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/requestlog"
	"github.com/lyimoexiao/akari/internal/requestlogadapter"
	requestmiddleware "github.com/lyimoexiao/akari/pkg/middleware"
	"go.uber.org/zap"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func Test_Middleware_persists_redacted_audit_data_for_API_request(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	db := openRequestLogTestDB(t)
	engine := gin.New()
	repository := requestlogadapter.NewRepository(db)
	engine.Use(requestmiddleware.RequestID(), requestlog.Middleware(repository, zap.NewNop().Sugar()))
	engine.POST("/api/v1/auth/login", func(ctx *gin.Context) {
		ctx.Set(auth.CtxKeyUserID, uint(42))
		ctx.Header("Set-Cookie", "session=response-secret")
		ctx.JSON(http.StatusOK, gin.H{
			"access_token": "response-token",
			"email":        "alice@example.com",
			"result":       "accepted",
		})
	})
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"username":"alice","password":"request-secret"}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Authorization", "Bearer header-secret")
	recorder := httptest.NewRecorder()

	// When
	engine.ServeHTTP(recorder, request)

	// Then
	var record model.RequestLog
	if err := db.First(&record).Error; err != nil {
		t.Fatalf("load request log: %v", err)
	}
	if record.UserID == nil || *record.UserID != 42 {
		t.Fatalf("user_id = %v, want 42", record.UserID)
	}
	if record.Module != "auth" {
		t.Fatalf("module = %q, want auth", record.Module)
	}
	for field, value := range map[string]string{
		"request_headers":  record.RequestHeaders,
		"request_body":     record.RequestBody,
		"response_headers": record.ResponseHeaders,
		"response_body":    record.ResponseBody,
	} {
		if strings.Contains(value, "request-secret") || strings.Contains(value, "header-secret") || strings.Contains(value, "response-secret") || strings.Contains(value, "response-token") || strings.Contains(value, "alice@example.com") {
			t.Fatalf("%s contains unredacted sensitive data: %s", field, value)
		}
	}
	if !strings.Contains(record.RequestHeaders, "[REDACTED]") || !strings.Contains(record.RequestBody, "[REDACTED]") {
		t.Fatalf("redaction marker missing: headers=%s body=%s", record.RequestHeaders, record.RequestBody)
	}
	if !strings.Contains(record.RequestBody, "alice") || !strings.Contains(record.ResponseBody, "accepted") {
		t.Fatalf("non-sensitive values were not preserved: request=%s response=%s", record.RequestBody, record.ResponseBody)
	}
}

func Test_Service_GetByRequestID_returns_persisted_request(t *testing.T) {
	// Given
	db := openRequestLogTestDB(t)
	want := model.RequestLog{
		RequestID: "request-123",
		Method:    http.MethodPost,
		Path:      "/api/v1/auth/login",
		Status:    http.StatusUnauthorized,
		LatencyMS: 12,
		IP:        "127.0.0.1",
		UserAgent: "request-log-test",
	}
	if err := db.Create(&want).Error; err != nil {
		t.Fatalf("seed request log: %v", err)
	}
	service := requestlog.NewService(requestlogadapter.NewRepository(db))

	// When
	got, err := service.GetByRequestID(t.Context(), want.RequestID)
	// Then
	if err != nil {
		t.Fatalf("get request log: %v", err)
	}
	if got.RequestID != want.RequestID || got.Path != want.Path || got.Status != want.Status {
		t.Fatalf("request log = %#v, want request_id=%q path=%q status=%d", got, want.RequestID, want.Path, want.Status)
	}
}

func Test_Handler_Get_returns_request_log_for_request_id(t *testing.T) {
	// Given
	gin.SetMode(gin.TestMode)
	db := openRequestLogTestDB(t)
	want := model.RequestLog{RequestID: "request-456", Method: http.MethodGet, Path: "/api/v1/users", Status: http.StatusOK}
	if err := db.Create(&want).Error; err != nil {
		t.Fatalf("seed request log: %v", err)
	}
	engine := gin.New()
	repository := requestlogadapter.NewRepository(db)
	requestlog.NewHandler(requestlog.NewService(repository), zap.NewNop().Sugar()).RegisterRoutes(engine.Group("/api/v1"))
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/request-logs/"+want.RequestID, nil)

	// When
	engine.ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body=%s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	var envelope struct {
		Data model.RequestLog `json:"data"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if envelope.Data.RequestID != want.RequestID {
		t.Fatalf("request_id = %q, want %q", envelope.Data.RequestID, want.RequestID)
	}
}

func openRequestLogTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.RequestLog{}); err != nil {
		t.Fatalf("migrate request logs: %v", err)
	}
	return db
}
