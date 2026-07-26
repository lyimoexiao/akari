package permission

import "context"

type AccessControl interface {
	EnforceUser(context.Context, Check) (bool, string, error)
	IdentifiersForUser(context.Context, uint) ([]string, error)
	Snapshot() (Snapshot, error)
}

type Service struct {
	access AccessControl
}

func NewService(access AccessControl) *Service {
	return &Service{access: access}
}

func (s *Service) EnforceUser(ctx context.Context, check Check) (bool, string, error) {
	return s.access.EnforceUser(ctx, check)
}

func (s *Service) IdentifiersForUser(ctx context.Context, userID uint) ([]string, error) {
	return s.access.IdentifiersForUser(ctx, userID)
}

func (s *Service) Snapshot() (Snapshot, error) {
	return s.access.Snapshot()
}
