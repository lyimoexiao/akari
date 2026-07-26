package sign

import (
	"context"
	"time"
)

type Repository interface {
	// LastSignAt returns the last sign-in time for the user, or nil if never signed.
	LastSignAt(ctx context.Context, userID uint) (*time.Time, error)

	// RecordSign persists the sign-in timestamp.
	RecordSign(ctx context.Context, userID uint, at time.Time) error
}

type Clock interface {
	Now() time.Time
}

type Dependencies struct {
	Repository Repository
	ScoreOps   ScoreOperator
	Config     SignConfig
	Clock      Clock
}

type SignConfig struct {
	GapHours  int
	ScoreMin  int64
	ScoreMax  int64
	AfterZero bool
}

type ScoreOperator interface {
	Award(ctx context.Context, userID uint, amount int64, reason string) error
	Balance(ctx context.Context, userID uint) (int64, error)
}
