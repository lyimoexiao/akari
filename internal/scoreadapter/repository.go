// Package scoreadapter provides a GORM-based implementation of score.Operator.
//
// All mutations use atomic SQL: UPDATE ... SET score = score +/- ? WHERE id = ?
// This avoids read-modify-write races without explicit locking.
// Every mutation is also recorded in the score_logs table for audit/history.
package scoreadapter

import (
	"context"
	"fmt"
	"time"

	"github.com/lyimoexiao/akari/internal/model"
	"github.com/lyimoexiao/akari/pkg/pagination"
	"github.com/lyimoexiao/akari/internal/score"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Balance(ctx context.Context, userID uint) (int64, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("score").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, fmt.Errorf("user %d: %w", userID, score.ErrInsufficientScore)
		}
		return 0, fmt.Errorf("get balance: %w", err)
	}
	return user.Score, nil
}

// Deduct atomically subtracts amount from the user's score.
// If deduction succeeds, a score_log entry is written.
func (r *Repository) Deduct(ctx context.Context, userID uint, amount int64, reason string) (int64, error) {
	if amount <= 0 {
		return 0, nil
	}

	var balance int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&model.User{}).
			Where("id = ? AND score >= ?", userID, amount).
			Update("score", gorm.Expr("score - ?", amount))
		if err := result.Error; err != nil {
			return err
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound // signal: no rows updated
		}

		// Read new balance
		var user model.User
		if err := tx.Select("score").First(&user, userID).Error; err != nil {
			return err
		}
		balance = user.Score

		// Write log entry
		log := model.ScoreLog{
			UserID:  userID,
			Amount:  -amount,
			Balance: balance,
			Reason:  reason,
		}
		return tx.Create(&log).Error
	})

	if err == gorm.ErrRecordNotFound {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("deduct score: %w", err)
	}
	return amount, nil
}

func (r *Repository) Award(ctx context.Context, userID uint, amount int64, reason string) error {
	if amount <= 0 {
		return nil
	}

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&model.User{}).
			Where("id = ?", userID).
			Update("score", gorm.Expr("score + ?", amount)).Error; err != nil {
			return err
		}

		// Read new balance
		var user model.User
		if err := tx.Select("score").First(&user, userID).Error; err != nil {
			return err
		}

		// Write log entry
		log := model.ScoreLog{
			UserID:  userID,
			Amount:  amount,
			Balance: user.Score,
			Reason:  reason,
		}
		return tx.Create(&log).Error
	})

	if err != nil {
		return fmt.Errorf("award score: %w", err)
	}
	return nil
}

func (r *Repository) HasEnough(ctx context.Context, userID uint, amount int64) (bool, error) {
	if amount <= 0 {
		return true, nil
	}

	var count int64
	if err := r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ? AND score >= ?", userID, amount).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("check score: %w", err)
	}
	return count > 0, nil
}

func (r *Repository) ScoreInfo(ctx context.Context, userID uint) (int64, *time.Time, error) {
	var user model.User
	if err := r.db.WithContext(ctx).Select("score", "last_sign_at").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return 0, nil, fmt.Errorf("user %d: %w", userID, score.ErrInsufficientScore)
		}
		return 0, nil, fmt.Errorf("get score info: %w", err)
	}
	return user.Score, user.LastSignAt, nil
}

func (r *Repository) ListLogs(ctx context.Context, userID uint, query score.LogQuery) (*score.LogList, error) {
	query.Normalise()

	dbQuery := r.db.WithContext(ctx).Model(&model.ScoreLog{}).
		Where("user_id = ?", userID)

	var total int64
	if err := dbQuery.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("count score logs: %w", err)
	}

	var logs []model.ScoreLog
	if err := dbQuery.Order("id DESC").
		Scopes(pagination.ApplyOffsetLimit(query.Paging)).Find(&logs).Error; err != nil {
		return nil, fmt.Errorf("list score logs: %w", err)
	}

	items := make([]score.LogEntry, len(logs))
	for i, l := range logs {
		items[i] = score.LogEntry{
			ID:        l.ID,
			Amount:    l.Amount,
			Balance:   l.Balance,
			Reason:    l.Reason,
			CreatedAt: l.CreatedAt.Format("2006-01-02T15:04:05+08:00"),
		}
	}

	return pagination.NewPaged(items, total, query.Paging), nil
}