package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lyimoexiao/akari/internal/config"
	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	Close() error
}

func New(cfg *config.CacheConfig) Cache {
	switch cfg.Type {
	case "redis":
		return newRedis(cfg)
	default:
		return newFileCache(cfg.FileDir)
	}
}

// --- Redis ---

type redisCache struct {
	client *redis.Client
}

func newRedis(cfg *config.CacheConfig) *redisCache {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})
	return &redisCache{client: rdb}
}

func (c *redisCache) Get(_ context.Context, key string, dest interface{}) error {
	val, err := c.client.Get(context.Background(), key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(val, dest)
}

func (c *redisCache) Set(_ context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(context.Background(), key, data, ttl).Err()
}

func (c *redisCache) Del(_ context.Context, keys ...string) error {
	return c.client.Del(context.Background(), keys...).Err()
}

func (c *redisCache) Close() error {
	return c.client.Close()
}

// --- File Cache ---

type fileCache struct {
	dir string
	mu  sync.RWMutex
}

type fileEntry struct {
	Value     interface{} `json:"value"`
	ExpiresAt time.Time   `json:"expires_at"`
}

func newFileCache(dir string) *fileCache {
	os.MkdirAll(dir, 0o755)
	return &fileCache{dir: dir}
}

func (c *fileCache) Get(_ context.Context, key string, dest interface{}) error {
	c.mu.RLock()
	defer c.mu.RUnlock()

	data, err := os.ReadFile(filepath.Join(c.dir, key+".json"))
	if err != nil {
		return err
	}

	var entry fileEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return err
	}
	if !entry.ExpiresAt.IsZero() && time.Now().After(entry.ExpiresAt) {
		os.Remove(filepath.Join(c.dir, key+".json"))
		return fmt.Errorf("cache key expired")
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
		os.Remove(filepath.Join(c.dir, key+".json"))
	}
	return nil
}

func (c *fileCache) Close() error {
	return nil
}
