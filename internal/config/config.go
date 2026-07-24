package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Cache    CacheConfig    `mapstructure:"cache"`
	Logger   LoggerConfig   `mapstructure:"logger"`
}

type ServerConfig struct {
	Port            string `mapstructure:"port"`
	Mode            string `mapstructure:"mode"`
	ShutdownTimeout int    `mapstructure:"shutdown_timeout"`
}

type LoggerConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	OutputPath string `mapstructure:"output_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
	Compress   bool   `mapstructure:"compress"`
	LocalTime  bool   `mapstructure:"local_time"`
}

type DatabaseConfig struct {
	Type     string `mapstructure:"type"`
	DSN      string `mapstructure:"dsn"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}

type CacheConfig struct {
	Type          string `mapstructure:"type"`
	RedisAddr     string `mapstructure:"redis_addr"`
	RedisPassword string `mapstructure:"redis_password"`
	RedisDB       int    `mapstructure:"redis_db"`
	FileDir       string `mapstructure:"file_dir"`
}

// Load reads configuration from multiple sources with the following priority
// (highest to lowest):
//   1. OS environment variables
//   2. .env file (loaded into environment before viper reads)
//   3. config.yaml (in ./, configs/, or custom CONFIG_PATH)
//   4. Built-in defaults
func Load(path string) (*Config, error) {
	// 1. Load .env into environment (so viper's AutomaticEnv picks it up)
	loadDotEnv(".env")

	// 2. Viper setup: defaults ← config file ← env vars (highest)
	v := viper.New()

	setDefaults(v)

	v.SetConfigName("config")
	v.SetConfigType("yaml")

	if path != "" {
		v.AddConfigPath(path)
	} else {
		v.AddConfigPath(".")
		v.AddConfigPath("configs")
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	// Environment variables: viper maps "server.port" ← "SERVER_PORT"
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// setDefaults registers all built-in default values.
func setDefaults(v *viper.Viper) {
	v.SetDefault("server.port", "8080")
	v.SetDefault("server.mode", "debug")
	v.SetDefault("server.shutdown_timeout", 10)

	v.SetDefault("database.type", "sqlite")
	v.SetDefault("database.host", "localhost")
	v.SetDefault("database.port", "3306")
	v.SetDefault("database.user", "root")
	v.SetDefault("database.password", "")
	v.SetDefault("database.name", "data/akari.db")
	v.SetDefault("database.dsn", "")

	v.SetDefault("cache.type", "file")
	v.SetDefault("cache.redis_addr", "localhost:6379")
	v.SetDefault("cache.redis_password", "")
	v.SetDefault("cache.redis_db", 0)
	v.SetDefault("cache.file_dir", "data/cache")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.output_path", "")
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_age", 30)
	v.SetDefault("logger.max_backups", 3)
	v.SetDefault("logger.compress", true)
	v.SetDefault("logger.local_time", true)
}

// loadDotEnv reads a .env file and sets each KEY=VALUE into the OS environment,
// so viper's AutomaticEnv can pick them up. Existing env vars are NOT overridden.
func loadDotEnv(path string) {
	f, err := os.Open(path)
	if err != nil {
		return // .env does not exist → fine
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		// Remove optional quotes around value
		val = strings.Trim(val, "\"'")
		// Do NOT override existing env vars (they have higher priority)
		if os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}