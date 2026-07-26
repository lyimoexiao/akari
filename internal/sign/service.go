package sign

import (
	"context"
	"fmt"
	"math/rand"
)

type Service struct {
	repo  Repository
	ops   ScoreOperator
	cfg   SignConfig
	clock Clock
}

func NewService(deps Dependencies) *Service {
	return &Service{
		repo:  deps.Repository,
		ops:   deps.ScoreOps,
		cfg:   deps.Config,
		clock: deps.Clock,
	}
}

// Sign performs the daily sign-in for the given user.
// Returns the acquired score on success.
func (s *Service) Sign(ctx context.Context, userID uint) (int64, error) {
	last, err := s.repo.LastSignAt(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("check last sign: %w", err)
	}

	now := s.clock.Now()
	if last != nil {
		if s.cfg.AfterZero {
			// Same calendar day → already signed
			if last.Year() == now.Year() && last.YearDay() == now.YearDay() {
				return 0, ErrAlreadySigned
			}
		} else {
			// Gap hours check
			hours := now.Sub(*last).Hours()
			if hours < float64(s.cfg.GapHours) {
				return 0, ErrSignTooEarly
			}
		}
	}

	low := s.cfg.ScoreMin
	high := s.cfg.ScoreMax
	if high < low {
		high = low
	}
	acquired := low + rand.Int63n(high-low+1)

	if acquired > 0 {
		if err := s.ops.Award(ctx, userID, acquired, "daily_sign"); err != nil {
			return 0, fmt.Errorf("award sign score: %w", err)
		}
	}

	if err := s.repo.RecordSign(ctx, userID, now); err != nil {
		return 0, fmt.Errorf("record sign: %w", err)
	}

	return acquired, nil
}

// Status returns the sign-in status for the user.
func (s *Service) Status(ctx context.Context, userID uint) (signedToday bool, score int64, err error) {
	score, err = s.ops.Balance(ctx, userID)
	if err != nil {
		return false, 0, err
	}

	last, err := s.repo.LastSignAt(ctx, userID)
	if err != nil {
		return false, 0, err
	}

	if last == nil {
		return false, score, nil
	}

	now := s.clock.Now()
	if s.cfg.AfterZero {
		signedToday = last.Year() == now.Year() && last.YearDay() == now.YearDay()
	} else {
		hours := now.Sub(*last).Hours()
		signedToday = hours < float64(s.cfg.GapHours)
	}

	return signedToday, score, nil
}
