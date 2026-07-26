package router

import "context"

type HealthChecker interface {
	Check(context.Context) error
}
