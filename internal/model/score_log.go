package model

import "time"

// ScoreLog records a single user score mutation (deduction or award).
type ScoreLog struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	Amount    int64     `gorm:"not null" json:"amount"`    // positive = award, negative = deduct
	Balance   int64     `gorm:"not null" json:"balance"`   // balance after this mutation
	Reason    string    `gorm:"size:32;not null" json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

func (ScoreLog) TableName() string {
	return "score_logs"
}
