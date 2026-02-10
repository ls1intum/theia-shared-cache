package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server  ServerConfig  `mapstructure:"server"`
	Storage StorageConfig `mapstructure:"storage"`
	Cache   CacheConfig   `mapstructure:"cache"`
	Auth    AuthConfig    `mapstructure:"auth"`
	Metrics MetricsConfig `mapstructure:"metrics"`
	Logging LoggingConfig `mapstructure:"logging"`
	Sentry  SentryConfig  `mapstructure:"sentry"`
}

type ServerConfig struct {
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	TLS          TLSConfig     `mapstructure:"tls"`
}

type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

type StorageConfig struct {
	Endpoint  string `mapstructure:"endpoint"`
	AccessKey string `mapstructure:"access_key"`
	SecretKey string `mapstructure:"secret_key"`
	Bucket    string `mapstructure:"bucket"`
	UseSSL    bool   `mapstructure:"use_ssl"`
}

type CacheConfig struct {
	MaxEntrySizeMB int64 `mapstructure:"max_entry_size_mb"`
}

type AuthConfig struct {
	Enabled bool       `mapstructure:"enabled"`
	Users   []UserAuth `mapstructure:"users"`
}

type UserAuth struct {
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Role     string `mapstructure:"role"`
}

type MetricsConfig struct {
	Enabled bool `mapstructure:"enabled"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

type SentryConfig struct {
	Dsn     string `mapstructure:"dsn"`
	Enabled bool   `mapstructure:"enabled"`
}

func Load(configPath string) (*Config, error) {
	v := viper.New()

	// Set defaults
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.read_timeout", "30s")
	v.SetDefault("server.write_timeout", "120s")
	v.SetDefault("server.tls.enabled", false)
	v.SetDefault("server.tls.cert_file", "/etc/certs/tls.crt")
	v.SetDefault("server.tls.key_file", "/etc/certs/tls.key")

	v.SetDefault("storage.endpoint", "localhost:9000")
	v.SetDefault("storage.bucket", "gradle-cache")
	v.SetDefault("storage.use_ssl", false)

	v.SetDefault("cache.max_entry_size_mb", 100)

	v.SetDefault("auth.enabled", true)

	v.SetDefault("metrics.enabled", true)

	v.SetDefault("sentry.enabled", false)

	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")

	// Read from config file if provided
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
	}

	// Enable environment variable overrides
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	// Bind specific environment variables
	v.BindEnv("storage.access_key", "MINIO_ACCESS_KEY")
	v.BindEnv("storage.secret_key", "MINIO_SECRET_KEY")
	v.BindEnv("sentry.dsn", "SENTRY_DSN")

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Override user credentials from environment variables if set
	overrideUserFromEnv(&cfg, 0, "CACHE_READER_USERNAME", "CACHE_READER_PASSWORD")
	overrideUserFromEnv(&cfg, 1, "CACHE_WRITER_USERNAME", "CACHE_WRITER_PASSWORD")

	return &cfg, nil
}

// overrideUserFromEnv overrides username and password for a specific user index
// from environment variables, if they are set.
func overrideUserFromEnv(cfg *Config, index int, usernameEnv, passwordEnv string) {
	if index >= len(cfg.Auth.Users) {
		return
	}
	if val := os.Getenv(usernameEnv); val != "" {
		cfg.Auth.Users[index].Username = val
	}
	if val := os.Getenv(passwordEnv); val != "" {
		cfg.Auth.Users[index].Password = val
	}
}

func (c *Config) Validate() error {
	if c.Storage.Endpoint == "" {
		return fmt.Errorf("storage.endpoint is required")
	}
	if c.Storage.AccessKey == "" {
		return fmt.Errorf("storage.access_key is required")
	}
	if c.Storage.SecretKey == "" {
		return fmt.Errorf("storage.secret_key is required")
	}
	if c.Storage.Bucket == "" {
		return fmt.Errorf("storage.bucket is required")
	}
	if c.Auth.Enabled && len(c.Auth.Users) == 0 {
		return fmt.Errorf("auth.users is required when auth is enabled")
	}
	for i, user := range c.Auth.Users {
		if user.Username == "" {
			return fmt.Errorf("auth.users[%d].username is required", i)
		}
		if user.Password == "" {
			return fmt.Errorf("auth.users[%d].password is required", i)
		}
	}
	if c.Server.TLS.Enabled {
		if c.Server.TLS.CertFile == "" {
			return fmt.Errorf("server.tls.cert_file is required when TLS is enabled")
		}
		if c.Server.TLS.KeyFile == "" {
			return fmt.Errorf("server.tls.key_file is required when TLS is enabled")
		}
	}
	return nil
}

func (c *Config) MaxEntrySizeBytes() int64 {
	return c.Cache.MaxEntrySizeMB * 1024 * 1024
}
