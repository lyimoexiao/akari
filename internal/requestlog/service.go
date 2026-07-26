package requestlog

import (
	"context"
	"errors"

	"github.com/lyimoexiao/akari/internal/model"
)

var ErrNotFound = errors.New("请求日志不存在")

type Service struct {
	reader Reader
}

func NewService(reader Reader) *Service {
	return &Service{reader: reader}
}

func (s *Service) GetByRequestID(ctx context.Context, requestID string) (model.RequestLog, error) {
	return s.reader.GetByRequestID(ctx, requestID)
}
