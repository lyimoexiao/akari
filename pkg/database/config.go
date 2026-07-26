package database

// Config holds database connection parameters.
// The mapstructure tags allow viper (in internal/config) to unmarshal
// directly into this type.
type Config struct {
	Type     string `mapstructure:"type"`
	DSN      string `mapstructure:"dsn"`
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Name     string `mapstructure:"name"`
}
