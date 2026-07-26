package cache

// Config holds cache backend selection and per-backend settings.
// The mapstructure tags allow viper (in internal/config) to unmarshal
// directly into this type.
type Config struct {
	Type   string            `mapstructure:"type"` // memory | redis | file
	Memory MemoryCacheConfig `mapstructure:"memory"`
	Redis  RedisCacheConfig  `mapstructure:"redis"`
	File   FileCacheConfig   `mapstructure:"file"`
}

// MemoryCacheConfig configures the in-process LRU cache.
type MemoryCacheConfig struct {
	DefaultTTL string `mapstructure:"default_ttl"`
	Size       int    `mapstructure:"size"`
}

// RedisCacheConfig configures the Redis cache backend.
type RedisCacheConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// FileCacheConfig configures the file-based cache backend.
type FileCacheConfig struct {
	Dir string `mapstructure:"dir"`
}
