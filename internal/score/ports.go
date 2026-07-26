// Package score defines the interface for managing user score/points.
// All score mutations (deduct, award) are handled atomically at the DB level.
package score

import (
	"context"
	"time"

	"github.com/lyimoexiao/akari/pkg/pagination"
)

// LogEntry represents a single score mutation record.
type LogEntry struct {
	ID        uint   `json:"id"`
	Amount    int64  `json:"amount"`  // positive = award, negative = deduct
	Balance   int64  `json:"balance"` // balance after this mutation
	Reason    string `json:"reason"`
	CreatedAt string `json:"created_at"`
}

// LogQuery defines pagination for score log queries.
type LogQuery struct {
	pagination.Paging
}

// LogList is the paginated result of score log queries.
type LogList = pagination.Paged[LogEntry]

// Operator handles atomic score mutations for a user.
type Operator interface {
	// Balance returns the current score of a user.
	Balance(ctx context.Context, userID uint) (int64, error)

	// Deduct subtracts `amount` from the user's score.
	// Returns the actual amount deducted (0 if insufficient).
	Deduct(ctx context.Context, userID uint, amount int64, reason string) (int64, error)

	// Award adds `amount` to the user's score.
	Award(ctx context.Context, userID uint, amount int64, reason string) error

	// HasEnough checks whether the user has at least `amount` score.
	HasEnough(ctx context.Context, userID uint, amount int64) (bool, error)

	// ListLogs returns paginated score history for a user, newest first.
	ListLogs(ctx context.Context, userID uint, query LogQuery) (*LogList, error)

	// ScoreInfo returns the current score and last sign-in time for a user.
	ScoreInfo(ctx context.Context, userID uint) (score int64, lastSignAt *time.Time, err error)
}