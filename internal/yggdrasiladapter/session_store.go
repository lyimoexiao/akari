package yggdrasiladapter

import (
	"context"
	"time"

	"github.com/lyimoexiao/akari/internal/yggdrasil"
	"github.com/lyimoexiao/akari/pkg/cache"
)

const joinSessionPrefix = "yggdrasil:join:"

type SessionStore struct {
	cache cache.Cache
}

func NewSessionStore(cacheService cache.Cache) *SessionStore {
	return &SessionStore{cache: cacheService}
}

func (s *SessionStore) Save(ctx context.Context, serverID string, session yggdrasil.ServerSession, ttl time.Duration) error {
	return s.cache.Set(ctx, joinSessionPrefix+serverID, session, ttl)
}

func (s *SessionStore) Load(ctx context.Context, serverID string) (yggdrasil.ServerSession, error) {
	var session yggdrasil.ServerSession
	err := s.cache.Get(ctx, joinSessionPrefix+serverID, &session)
	return session, err
}
