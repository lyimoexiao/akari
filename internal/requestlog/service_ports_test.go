package requestlog_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/internal/requestlog"
	requestmiddleware "github.com/lyimoexiao/akari/pkg/middleware"
	"go.uber.org/zap"
)

type fakeRequestLogRepository struct {
	record    model.RequestLog
	saveErr   error
	saved     model.RequestLog
	saveCalls int
}

func (repository *fakeRequestLogRepository) GetByRequestID(context.Context, string) (model.RequestLog, error) {
	return repository.record, nil
}

func (repository *fakeRequestLogRepository) Save(_ context.Context, record model.RequestLog) error {
	repository.saveCalls++
	repository.saved = record
	return repository.saveErr
}

func Test_Service_reads_through_repository_port(t *testing.T) {
	// Given
	repository := &fakeRequestLogRepository{record: model.RequestLog{RequestID: "request-123"}}
	service := requestlog.NewService(repository)

	// When
	record, err := service.GetByRequestID(t.Context(), "request-123")

	// Then
	if err != nil {
		t.Fatalf("get request log: %v", err)
	}
	if record.RequestID != "request-123" {
		t.Fatalf("request_id = %q, want request-123", record.RequestID)
	}
}

func Test_Middleware_recorder_failure_does_not_change_response(t *testing.T) {
	// Given
	repository := &fakeRequestLogRepository{saveErr: errors.New("persistence failed")}
	engine := gin.New()
	engine.Use(requestmiddleware.RequestID())
	engine.Use(requestlog.Middleware(repository, zap.NewNop().Sugar()))
	engine.GET("/api/v1/example", func(ctx *gin.Context) {
		ctx.JSON(http.StatusCreated, gin.H{"ok": true})
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/example", nil)

	// When
	engine.ServeHTTP(recorder, request)

	// Then
	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if recorder.Body.String() != `{"ok":true}` {
		t.Fatalf("body = %q, want successful response", recorder.Body.String())
	}
	if repository.saveCalls != 1 {
		t.Fatalf("save calls = %d, want 1", repository.saveCalls)
	}
	if repository.saved.Status != http.StatusCreated {
		t.Fatalf("saved status = %d, want %d", repository.saved.Status, http.StatusCreated)
	}
}
