package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	lru "github.com/hashicorp/golang-lru/v2/expirable"
	"github.com/redis/go-redis/v9"

	"github.com/lyimoexiao/akari/pkg/logger"
	"github.com/lyimoexiao/akari/pkg/util"
)

// ErrCacheMiss indicates the requested key was not found in the cache.
// Callers can use errors.Is to distinguish "not found" from other errors
// (e.g. connection failure, deserialization error).
var ErrCacheMiss = errors.New("cache: key not found")

// Cache defines the interface for cache operations.
type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Ping(ctx context.Context) error
	Close() error
}

// New creates a single cache backend based on cache.type.
// Supported types: "memory" (default), "redis", "file".
func New(cfg *Config) Cache {
	switch cfg.Type {
	case "redis":
		return newRedisCache(&cfg.Redis)
	case "file":
		return newFileCache(&cfg.File)
	default:
		// "memory" or unknown → memory cache (safe fallback)
		logger.L.Warn("Cache type is 'memory' — token revocation state is local to this process. " +
			"Multi-instance deployments MUST use 'redis' to share revocation state across all instances.")
		return newMemoryCache(&cfg.Memory)
	}
}

// ─────────────────────────────────────────────────────────────────
// Memory Cache (hashicorp/golang-lru/v2 — expirable LRU)
// ─────────────────────────────────────────────────────────────────

type memoryCache struct {
	cache      *lru.LRU[string, any]
	defaultTTL time.Duration
}

type memoryEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}

func newMemoryCache(cfg *MemoryCacheConfig) *memoryCache {
	ttl := util.ParseDuration(cfg.DefaultTTL, 5*time.Minute)
	size := cfg.Size
	if size <= 0 {
		size = 1024
	}

	return &memoryCache{
		cache:      lru.NewLRU[string, any](size, nil, ttl),
		defaultTTL: ttl,
	}
}

func (c *memoryCache) Get(_ context.Context, key string, dest interface{}) error {
	val, ok := c.cache.Get(key)
	if !ok {
		return fmt.Errorf("%w: %q", ErrCacheMiss, key)
	}

	entry, ok := val.(memoryEntry)
	if !ok {
		return fmt.Errorf("cache key %q: invalid entry type", key)
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		c.cache.Remove(key)
		return fmt.Errorf("%w: %q (expired)", ErrCacheMiss, key)
	}

	data, err := json.Marshal(entry.Value)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

func (c *memoryCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}
	c.cache.Add(key, memoryEntry{Value: value, ExpiresAt: expiresAt})
	return nil
}

func (c *memoryCache) Del(_ context.Context, keys ...string) error {
	for _, key := range keys {
		c.cache.Remove(key)
	}
	return nil
}

func (c *memoryCache) Ping(ctx context.Context) error {
	return ctx.Err()
}

func (c *memoryCache) Close() error {
	c.cache.Purge()
	return nil
}

// ─────────────────────────────────────────────────────────────────
// Redis Cache
// ─────────────────────────────────────────────────────────────────

type redisCache struct {
	client *redis.Client
}

func newRedisCache(cfg *RedisCacheConfig) *redisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	return &redisCache{client: rdb}
}

func (c *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return fmt.Errorf("%w: %q", ErrCacheMiss, key)
		}
		return err
	}
	return json.Unmarshal(val, dest)
}

func (c *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

func (c *redisCache) Del(ctx context.Context, keys ...string) error {
	return c.client.Del(ctx, keys...).Err()
}

func (c *redisCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}

// ─────────────────────────────────────────────────────────────────
// File Cache
// ─────────────────────────────────────────────────────────────────

type fileCache struct {
	dir string
	mu  sync.RWMutex
}

type fileEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}

func newFileCache(cfg *FileCacheConfig) *fileCache {
	_ = os.MkdirAll(cfg.Dir, 0o755)
	return &fileCache{dir: cfg.Dir}
}

func (c *fileCache) Get(_ context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(c.dir, key+".json"))
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %q", ErrCacheMiss, key)
		}
		return err
	}

	var entry fileEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		return fmt.Errorf("%w: %q (expired)", ErrCacheMiss, key)
	}

	raw, _ := json.Marshal(entry.Value)
	return json.Unmarshal(raw, dest)
}

func (c *fileCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	entry := fileEntry{Value: value, ExpiresAt: expiresAt}
	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(c.dir, key+".json"), data, 0o644)
}

func (c *fileCache) Del(_ context.Context, keys ...string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, key := range keys {
		_ = os.Remove(filepath.Join(c.dir, key+".json"))
	}
	return nil
}

func (c *fileCache) Ping(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	probe, err := os.CreateTemp(c.dir, ".health-*")
	if err != nil {
		return fmt.Errorf("create cache probe: %w", err)
	}
	probePath := probe.Name()
	if err := probe.Close(); err != nil {
		_ = os.Remove(probePath)
		return fmt.Errorf("close cache probe: %w", err)
	}
	if err := os.Remove(probePath); err != nil {
		return fmt.Errorf("remove cache probe: %w", err)
	}
	return ctx.Err()
}

func (c *fileCache) Close() error {
	return nil
}

// ─────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────
