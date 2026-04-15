package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const (
	AppEnvLocal = "local"
	AppEnvTest  = "test"
	AppEnvProd  = "prod"
)

var allowedAppEnvs = []string{AppEnvLocal, AppEnvTest, AppEnvProd}

// Config contains the full application configuration loaded from env or .env.
type Config struct {
	AppEnv   string `env:"APP_ENV" env-default:"local"`
	LogLevel string `env:"LOG_LEVEL" env-default:"info"`
	HTTP     HTTPConfig
	DB       DBConfig
	Auth     AuthConfig
	I18N     I18NConfig
}

// HTTPConfig defines HTTP server binding and timeout settings.
type HTTPConfig struct {
	Host            string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port            string        `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout     time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout    time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
	IdleTimeout     time.Duration `env:"HTTP_IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `env:"HTTP_SHUTDOWN_TIMEOUT" env-default:"10s"`
}

// DBConfig defines PostgreSQL connection settings.
type DBConfig struct {
	Host           string        `env:"DB_HOST" env-default:"localhost"`
	Port           string        `env:"DB_PORT" env-default:"5432"`
	Name           string        `env:"DB_NAME" env-required:"true"`
	User           string        `env:"DB_USER" env-required:"true"`
	Password       string        `env:"DB_PASSWORD" env-required:"true"`
	SSLMode        string        `env:"DB_SSLMODE" env-default:"disable"`
	ConnectTimeout time.Duration `env:"DB_CONNECT_TIMEOUT" env-default:"5s"`
}

// AuthConfig defines JWT and password validation settings.
type AuthConfig struct {
	AccessSecret   string        `env:"JWT_ACCESS_SECRET" env-required:"true"`
	AccessTTL      time.Duration `env:"JWT_ACCESS_TTL" env-default:"15m"`
	PasswordMinLen int           `env:"AUTH_PASSWORD_MIN_LEN" env-default:"5"`
	PasswordMaxLen int           `env:"AUTH_PASSWORD_MAX_LEN" env-default:"72"`
}

// I18NConfig defines localization defaults.
type I18NConfig struct {
	DefaultLocale string `env:"DEFAULT_LOCALE" env-default:"en"`
}

// MustLoad loads configuration from .env or environment variables and panics on failure.
func MustLoad() Config {
	cfg, err := Load(".env")
	if err != nil {
		panic(err)
	}

	return cfg
}

// Load reads configuration from the given file when it exists, otherwise from environment variables.
func Load(path string) (Config, error) {
	var cfg Config

	if path != "" {
		if _, err := os.Stat(path); err == nil {
			if err := cleanenv.ReadConfig(path, &cfg); err != nil {
				return Config{}, fmt.Errorf("read config %s: %w", path, err)
			}

			if err := cfg.Validate(); err != nil {
				return Config{}, err
			}

			return cfg, nil
		} else if !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("stat config %s: %w", path, err)
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return Config{}, fmt.Errorf("read env: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

// Validate checks cross-field application configuration invariants.
func (c Config) Validate() error {
	if !isAllowedAppEnv(c.AppEnv) {
		return fmt.Errorf("validate app env: app env must be one of %s", joinQuoted(allowedAppEnvs))
	}

	if err := c.Auth.Validate(); err != nil {
		return fmt.Errorf("validate auth config: %w", err)
	}

	return nil
}

// Validate checks auth-specific configuration constraints.
func (c AuthConfig) Validate() error {
	const bcryptMaxPasswordBytes = 72

	if c.AccessSecret == "" {
		return errors.New("auth access secret is required")
	}

	if c.AccessTTL <= 0 {
		return errors.New("auth access ttl must be greater than 0")
	}

	if c.PasswordMinLen < 1 {
		return errors.New("auth password min len must be greater than 0")
	}

	if c.PasswordMaxLen < c.PasswordMinLen {
		return errors.New("auth password max len must be greater than or equal to min len")
	}

	if c.PasswordMaxLen > bcryptMaxPasswordBytes {
		return fmt.Errorf("auth password max len must be less than or equal to %d bytes", bcryptMaxPasswordBytes)
	}

	return nil
}

// Address returns the HTTP listen address in host:port form.
func (c HTTPConfig) Address() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// DSN builds a PostgreSQL connection string from the configured database settings.
func (c DBConfig) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Name,
		c.SSLMode,
	)
}

// isAllowedAppEnv reports whether the configured app environment is supported.
func isAllowedAppEnv(appEnv string) bool {
	return slices.Contains(allowedAppEnvs, appEnv)
}

// joinQuoted formats string values as a comma-separated quoted list.
func joinQuoted(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return strings.Join(quoted, ", ")
}
