package captcha

// Config holds captcha service configuration.
// The mapstructure tags allow viper (in internal/config) to unmarshal
// directly into this type.
type Config struct {
	Enabled     bool            `mapstructure:"enabled"`
	Provider    string          `mapstructure:"provider"` // "gocaptcha" | "turnstile"
	Type        string          `mapstructure:"type"`     // used only by gocaptcha: click | rotate | slide
	CachePrefix string          `mapstructure:"cache_prefix"`
	Turnstile   TurnstileConfig `mapstructure:"turnstile"`
}

// TurnstileConfig holds Cloudflare Turnstile credentials.
type TurnstileConfig struct {
	SiteKey   string `mapstructure:"site_key"`
	SecretKey string `mapstructure:"secret_key"`
}
