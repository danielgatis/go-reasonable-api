// Package config loads application configuration from files and environment.
//
// Configuration is loaded via Viper with the following precedence:
//  1. Environment variables (highest priority)
//  2. config.yaml file
//  3. Built-in defaults (lowest priority)
//
// Environment variables use underscore separation: DATABASE_URL, AUTH_SECRET.
//
// Sensitive fields (secrets, passwords, API keys) implement fmt.Stringer
// with redacted output to prevent accidental logging.
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/rotisserie/eris"
	"github.com/spf13/viper"
)

// Environment controls behavior differences between deployment targets.
type Environment string

const (
	EnvDevelopment Environment = "development"
	EnvProduction  Environment = "production"
)

func (e Environment) IsDev() bool {
	return e == EnvDevelopment
}

func (e Environment) IsProd() bool {
	return e == EnvProduction
}

// Config is the root configuration structure.
// Load() populates this from environment and config files.
type Config struct {
	Environment Environment    `mapstructure:"environment"`
	Server      ServerConfig   `mapstructure:"server"`
	Database    DatabaseConfig `mapstructure:"database"`
	Auth        AuthConfig     `mapstructure:"auth"`
	Redis       RedisConfig    `mapstructure:"redis"`
	Logger      LoggerConfig   `mapstructure:"logger"`
	Worker      WorkerConfig   `mapstructure:"worker"`
	App         AppConfig      `mapstructure:"app"`
	Email       EmailConfig    `mapstructure:"email"`
	Sentry      SentryConfig   `mapstructure:"sentry"`
}

type LoggerConfig struct {
	Level  string `mapstructure:"level"`
	Pretty bool   `mapstructure:"pretty"`
}

type ServerConfig struct {
	Port               string     `mapstructure:"port"`
	RateLimitPerSecond int        `mapstructure:"rate_limit_per_second"`
	CORS               CORSConfig `mapstructure:"cors"`
}

type CORSConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowMethods     []string `mapstructure:"allow_methods"`
	AllowHeaders     []string `mapstructure:"allow_headers"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
	MaxAge           int      `mapstructure:"max_age"` // seconds
}

type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// String returns a string representation with sensitive fields masked.
func (c DatabaseConfig) String() string {
	return fmt.Sprintf("DatabaseConfig{URL: [REDACTED], MaxOpenConns: %d, MaxIdleConns: %d, ConnMaxLifetime: %s, ConnMaxIdleTime: %s}",
		c.MaxOpenConns, c.MaxIdleConns, c.ConnMaxLifetime, c.ConnMaxIdleTime)
}

type AuthConfig struct {
	Secret                    string        `mapstructure:"secret"`
	AuthTokenTTL              time.Duration `mapstructure:"auth_token_ttl"`
	PasswordResetTokenTTL     time.Duration `mapstructure:"password_reset_token_ttl"`
	EmailConfirmationTokenTTL time.Duration `mapstructure:"email_confirmation_token_ttl"`
	AccountDeletionDelay      time.Duration `mapstructure:"account_deletion_delay"`
	BcryptCost                int           `mapstructure:"bcrypt_cost"`
}

// String returns a string representation with sensitive fields masked.
func (c AuthConfig) String() string {
	return fmt.Sprintf("AuthConfig{Secret: [REDACTED], AuthTokenTTL: %s, PasswordResetTokenTTL: %s, EmailConfirmationTokenTTL: %s, AccountDeletionDelay: %s, BcryptCost: %d}",
		c.AuthTokenTTL, c.PasswordResetTokenTTL, c.EmailConfirmationTokenTTL, c.AccountDeletionDelay, c.BcryptCost)
}

type RedisConfig struct {
	Addr string `mapstructure:"addr"`
}

type WorkerConfig struct {
	Concurrency    int           `mapstructure:"concurrency"`
	EmailMaxRetry  int           `mapstructure:"email_max_retry"`
	EmailTimeout   time.Duration `mapstructure:"email_timeout"`
	EmailRetention time.Duration `mapstructure:"email_retention"`
}

type AppConfig struct {
	BaseURL string `mapstructure:"base_url"`
}

type EmailConfig struct {
	Provider       string `mapstructure:"provider"`         // "smtp" or "sendgrid"
	From           string `mapstructure:"from"`             // Sender email address
	FromName       string `mapstructure:"from_name"`        // Sender name
	SMTPHost       string `mapstructure:"smtp_host"`        // SMTP server host
	SMTPPort       int    `mapstructure:"smtp_port"`        // SMTP server port
	SMTPUser       string `mapstructure:"smtp_user"`        // SMTP username (optional for MailHog)
	SMTPPassword   string `mapstructure:"smtp_password"`    // SMTP password (optional for MailHog)
	SendGridAPIKey string `mapstructure:"sendgrid_api_key"` // SendGrid API key (for production)
}

// String returns a string representation with sensitive fields masked.
func (c EmailConfig) String() string {
	return fmt.Sprintf("EmailConfig{Provider: %s, From: %s, FromName: %s, SMTPHost: %s, SMTPPort: %d, SMTPUser: %s, SMTPPassword: [REDACTED], SendGridAPIKey: [REDACTED]}",
		c.Provider, c.From, c.FromName, c.SMTPHost, c.SMTPPort, c.SMTPUser)
}

type SentryConfig struct {
	DSN              string  `mapstructure:"dsn"`
	EnableTracing    bool    `mapstructure:"enable_tracing"`
	TracesSampleRate float64 `mapstructure:"traces_sample_rate"`
}

// String returns a string representation with sensitive fields masked.
func (c SentryConfig) String() string {
	return fmt.Sprintf("SentryConfig{DSN: [REDACTED], EnableTracing: %t, TracesSampleRate: %f}",
		c.EnableTracing, c.TracesSampleRate)
}

func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Defaults for development
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.rate_limit_per_second", 20)
	viper.SetDefault("server.cors.allow_origins", []string{"*"})
	viper.SetDefault("server.cors.allow_methods", []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})
	viper.SetDefault("server.cors.allow_headers", []string{"Origin", "Content-Type", "Accept", "Authorization"})
	viper.SetDefault("server.cors.allow_credentials", false)
	viper.SetDefault("server.cors.max_age", 86400) // 24 hours
	viper.SetDefault("database.url", "postgres://reasonable:reasonable@localhost:5433/reasonable?sslmode=disable")
	viper.SetDefault("database.max_open_conns", 25)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "5m")
	viper.SetDefault("database.conn_max_idle_time", "1m")
	viper.SetDefault("auth.secret", "dev-secret-change-in-production")
	viper.SetDefault("auth.auth_token_ttl", "0")
	viper.SetDefault("auth.password_reset_token_ttl", "1h")
	viper.SetDefault("auth.email_confirmation_token_ttl", "24h")
	viper.SetDefault("auth.account_deletion_delay", "720h") // 30 days
	viper.SetDefault("auth.bcrypt_cost", 12)
	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.pretty", true)
	viper.SetDefault("worker.concurrency", 10)
	viper.SetDefault("worker.email_max_retry", 5)
	viper.SetDefault("worker.email_timeout", "30s")
	viper.SetDefault("worker.email_retention", "24h")
	viper.SetDefault("app.base_url", "http://localhost:3000")
	viper.SetDefault("environment", "development")

	// Email defaults for development (MailHog)
	viper.SetDefault("email.provider", "smtp")
	viper.SetDefault("email.from", "noreply@go-reasonable-api.local")
	viper.SetDefault("email.from_name", "ZapAgenda")
	viper.SetDefault("email.smtp_host", "localhost")
	viper.SetDefault("email.smtp_port", 1025)
	viper.SetDefault("email.smtp_user", "")
	viper.SetDefault("email.smtp_password", "")
	viper.SetDefault("email.sendgrid_api_key", "")

	// Sentry defaults (disabled by default in development)
	viper.SetDefault("sentry.dsn", "")
	viper.SetDefault("sentry.enable_tracing", false)
	viper.SetDefault("sentry.traces_sample_rate", 0.1)

	_ = viper.ReadInConfig()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, eris.Wrap(err, "failed to unmarshal config")
	}

	if err := cfg.Validate(); err != nil {
		return nil, eris.Wrap(err, "config validation failed")
	}

	return &cfg, nil
}

// Validate checks that all required configuration values are set and valid.
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return eris.New("database.url is required")
	}

	if c.Environment.IsProd() {
		if c.Auth.Secret == "dev-secret-change-in-production" {
			return eris.New("auth.secret must be changed in production")
		}
		if c.Auth.Secret == "" {
			return eris.New("auth.secret is required in production")
		}
	}

	if c.Redis.Addr == "" {
		return eris.New("redis.addr is required")
	}

	if c.Server.Port == "" {
		return eris.New("server.port is required")
	}

	if c.Auth.BcryptCost < 4 || c.Auth.BcryptCost > 31 {
		return eris.New("auth.bcrypt_cost must be between 4 and 31")
	}

	return nil
}
