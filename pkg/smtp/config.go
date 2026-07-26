package smtp

// Config holds SMTP server connection parameters.
// The mapstructure tags allow viper (in internal/config) to unmarshal directly
// into this type.
type Config struct {
	Host      string `mapstructure:"host"`
	Port      string `mapstructure:"port"`
	Username  string `mapstructure:"username"`
	Password  string `mapstructure:"password"`
	From      string `mapstructure:"from"`
	SSL       bool   `mapstructure:"ssl"`
	Timeout   int    `mapstructure:"timeout"`
	QueueSize int    `mapstructure:"queue_size"`
}