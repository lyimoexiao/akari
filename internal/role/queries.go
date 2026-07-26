package role

import (
	"context"

	"github.com/lyimoexiao/akari/internal/model"
)

func (s *Service) FindByName(ctx context.Context, name string) (model.Role, error) {
	return s.repository.RoleByName(ctx, name)
}

func (s *Service) FindRegistrationRole(ctx context.Context) (*model.Role, error) {
	return s.repository.RegistrationRole(ctx)
}

func (s *Service) List(ctx context.Context) ([]Item, error) {
	roles, err := s.repository.ListRoles(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]Item, len(roles))
	for index := range roles {
		items[index] = toItem(roles[index])
	}
	return items, nil
}
