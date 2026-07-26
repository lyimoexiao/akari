package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/lyimoexiao/akari/pkg/cache"
	"github.com/lyimoexiao/akari/pkg/captcha"
	"github.com/lyimoexiao/akari/pkg/database"
	"github.com/lyimoexiao/akari/pkg/logger"
	"github.com/lyimoexiao/akari/pkg/smtp"
	"github.com/spf13/viper"
)

type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  database.Config `mapstructure:"database"`
	Cache     cache.Config    `mapstructure:"cache"`
	Logger    logger.Config   `mapstructure:"logger"`
	JWT       JWTConfig       `mapstructure:"jwt"`
	Captcha   captcha.Config  `mapstructure:"captcha"`
	SMTP      smtp.Config     `mapstructure:"smtp"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Yggdrasil YggdrasilConfig `mapstructure:"yggdrasil"`
	Score     ScoreConfig     `mapstructure:"score"`
	Storage   StorageConfig   `mapstructure:"storage"`
}

type YggdrasilConfig struct {
	ServerName         string `mapstructure:"server_name"`
	ImplementationName string `mapstructure:"implementation_name"`
}

// StorageConfig controls texture file storage.
type StorageConfig struct {
	Dir string `mapstructure:"dir"`
}

// ScoreConfig controls the user score/points economy.
type ScoreConfig struct {
	InitialScore        int64 `mapstructure:"initial_score"`
	SignGapHours        int   `mapstructure:"sign_gap_hours"`
	SignScoreMin        int64 `mapstructure:"sign_score_min"`
	SignScoreMax        int64 `mapstructure:"sign_score_max"`
	SignAfterZero       bool  `mapstructure:"sign_after_zero"`
	CostPerStorageKB    int64 `mapstructure:"cost_per_storage_kb"`
	CostPrivateStorage  int64 `mapstructure:"cost_private_storage"`
	CostPerPlayer       int64 `mapstructure:"cost_per_player"`
	CostPerClosetItem   int64 `mapstructure:"cost_per_closet_item"`
	ReturnScoreOnRemove bool  `mapstructure:"return_score_on_remove"`
	AwardPerUpload      int64 `mapstructure:"award_per_upload"`
	AwardPerLike        int64 `mapstructure:"award_per_like"`
	TakeBackOnDelete    bool  `mapstructure:"take_back_on_delete"`
}

type ServerConfig struct {
	Port              string `mapstructure:"port"`
	Mode              string `mapstructure:"mode"`
	BaseURL           string `mapstructure:"base_url"` // public-facing URL, e.g. "https://example.com"
	ReadHeaderTimeout int    `mapstructure:"read_header_timeout"`
	ReadTimeout       int    `mapstructure:"read_timeout"`
	WriteTimeout      int    `mapstructure:"write_timeout"`
	IdleTimeout       int    `mapstructure:"idle_timeout"`
	ShutdownTimeout   int    `mapstructure:"shutdown_timeout"`
}

type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	Issuer     string `mapstructure:"issuer"`
	Expiration string `mapstructure:"expiration"`
}

// AuthConfig controls user registration and email verification behavior.
type AuthConfig struct {
	RegistrationEnabled      bool   `mapstructure:"registration_enabled"`
	EmailVerificationEnabled bool   `mapstructure:"email_verification_enabled"`
	VerifyEmailTokenTTL      string `mapstructure:"verify_email_token_ttl"`
	PasswordResetEnabled     bool   `mapstructure:"password_reset_enabled"`
	PasswordResetTokenTTL    string `mapstructure:"password_reset_token_ttl"`
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

	// If base_url is not configured, derive from server.port
	if cfg.Server.BaseURL == "" {
		port := cfg.Server.Port
		if port == "" {
			port = "8080"
		}
		cfg.Server.BaseURL = fmt.Sprintf("http://localhost:%s", port)
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
	v.SetDefault("server.base_url", "")
	v.SetDefault("server.read_header_timeout", 5)
	v.SetDefault("server.read_timeout", 15)
	v.SetDefault("server.write_timeout", 30)
	v.SetDefault("server.idle_timeout", 60)
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
	v.SetDefault("captcha.provider", "gocaptcha")
	v.SetDefault("captcha.type", "click")
	v.SetDefault("captcha.cache_prefix", "captcha:")
	v.SetDefault("captcha.turnstile.site_key", "")
	v.SetDefault("captcha.turnstile.secret_key", "")

	v.SetDefault("smtp.host", "localhost")
	v.SetDefault("smtp.port", "25")
	v.SetDefault("smtp.username", "")
	v.SetDefault("smtp.password", "")
	v.SetDefault("smtp.from", "noreply@example.com")
	v.SetDefault("smtp.ssl", false)
	v.SetDefault("smtp.timeout", 10)
	v.SetDefault("smtp.queue_size", 100)

	v.SetDefault("yggdrasil.server_name", "Akari Yggdrasil")
	v.SetDefault("yggdrasil.implementation_name", "akari-yggdrasil")

	v.SetDefault("score.initial_score", 1000)
	v.SetDefault("score.sign_gap_hours", 24)
	v.SetDefault("score.sign_score_min", 10)
	v.SetDefault("score.sign_score_max", 100)
	v.SetDefault("score.sign_after_zero", false)
	v.SetDefault("score.cost_per_storage_kb", 0)
	v.SetDefault("score.cost_private_storage", 0)
	v.SetDefault("score.cost_per_player", 100)
	v.SetDefault("score.cost_per_closet_item", 0)
	v.SetDefault("score.return_score_on_remove", true)
	v.SetDefault("score.award_per_upload", 0)
	v.SetDefault("score.award_per_like", 0)
	v.SetDefault("score.take_back_on_delete", true)

	v.SetDefault("storage.dir", "data/textures")
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
