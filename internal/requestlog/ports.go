package requestlog

import (
	"context"

	"github.com/lyimoexiao/akari/internal/model"
)

type Reader interface {
	GetByRequestID(context.Context, string) (model.RequestLog, error)
}

type Writer interface {
	Save(context.Context, model.RequestLog) error
}
