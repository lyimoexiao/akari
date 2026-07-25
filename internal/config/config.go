package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Cache     CacheConfig     `mapstructure:"cache"`
	Logger    LoggerConfig    `mapstructure:"logger"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Captcha   CaptchaConfig   `mapstructure:"captcha"`
	SMTP      SMTPConfig      `mapstructure:"smtp"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Yggdrasil YggdrasilConfig `mapstructure:"yggdrasil"`
}

type YggdrasilConfig struct {
	ServerName         string `mapstructure:"server_name"`
	ImplementationName string `mapstructure:"implementation_name"`
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

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Issuer     string `mapstructure:"issuer"`
	Expiration string `mapstructure:"expiration"`
}

type CaptchaConfig struct {
	Enabled     bool   `mapstructure:"enabled"`
	Type        string `mapstructure:"type"`
	CachePrefix string `mapstructure:"cache_prefix"`
}

// AuthConfig controls user registration and email verification behavior.
type AuthConfig struct {
	RegistrationEnabled      bool   `mapstructure:"registration_enabled"`
	EmailVerificationEnabled bool   `mapstructure:"email_verification_enabled"`
	VerifyEmailTokenTTL      string `mapstructure:"verify_email_token_ttl"`
	PasswordResetEnabled     bool   `mapstructure:"password_reset_enabled"`
	PasswordResetTokenTTL    string `mapstructure:"password_reset_token_ttl"`
}

type SMTPConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
	SSL      bool   `mapstructure:"ssl"`
}

type CacheConfig struct {
	Type   string            `mapstructure:"type"` // memory | redis | file
	Memory MemoryCacheConfig `mapstructure:"memory"`
	Redis  RedisCacheConfig  `mapstructure:"redis"`
	File   FileCacheConfig   `mapstructure:"file"`
}

type MemoryCacheConfig struct {
	DefaultTTL string `mapstructure:"default_ttl"`
	Size       int    `mapstructure:"size"`
}

type RedisCacheConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type FileCacheConfig struct {
	Dir string `mapstructure:"dir"`
}

// Load reads configuration from multiple sources with the following priority
// (highest to lowest):
//  1. OS environment variables
//  2. .env file (loaded into environment before viper reads)
//  3. config.yaml (in ./, configs/, or custom CONFIG_PATH)
//  4. Built-in defaults
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
	v.SetDefault("auth.registration_enabled", true)
	v.SetDefault("auth.email_verification_enabled", false)
	v.SetDefault("auth.verify_email_token_ttl", "2h")
	v.SetDefault("auth.password_reset_enabled", true)
	v.SetDefault("auth.password_reset_token_ttl", "30m")

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

	v.SetDefault("cache.type", "memory")
	v.SetDefault("cache.memory.default_ttl", "5m")
	v.SetDefault("cache.memory.size", 1024)
	v.SetDefault("cache.redis.addr", "localhost:6379")
	v.SetDefault("cache.redis.password", "")
	v.SetDefault("cache.redis.db", 0)
	v.SetDefault("cache.file.dir", "data/cache")

	v.SetDefault("logger.level", "info")
	v.SetDefault("logger.format", "console")
	v.SetDefault("logger.output_path", "")
	v.SetDefault("logger.max_size", 100)
	v.SetDefault("logger.max_age", 30)
	v.SetDefault("logger.max_backups", 3)
	v.SetDefault("logger.compress", true)
	v.SetDefault("logger.local_time", true)

	v.SetDefault("jwt.secret", "change-me-in-production")
	v.SetDefault("jwt.issuer", "akari")
	v.SetDefault("jwt.expiration", "24h")

	v.SetDefault("captcha.enabled", false)
	v.SetDefault("captcha.type", "click")
	v.SetDefault("captcha.cache_prefix", "captcha:")

	v.SetDefault("smtp.host", "localhost")
	v.SetDefault("smtp.port", "25")
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "noreply@example.com")
	v.SetDefault("smtp.ssl", false)

	v.SetDefault("yggdrasil.server_name", "Akari Yggdrasil")
	v.SetDefault("yggdrasil.implementation_name", "akari-yggdrasil")
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
