package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	Session  SessionConfig
	Security SecurityConfig
	Cookie   CookieConfig
	MQ       MQConfig
	CORS     CORSConfig
}

type ServerConfig struct {
	Host string
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	Name         string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type SessionConfig struct {
	Secret           string
	Duration         time.Duration
	RefreshThreshold time.Duration
}

type SecurityConfig struct {
	BCryptCost        int
	RateLimitRequests int
	RateLimitWindow   time.Duration
}

type CookieConfig struct {
	Secure bool
	Domain string
}

type MQConfig struct {
	Type        string
	URL         string
	RabbitMQURL string
}

type CORSConfig struct {
	AllowedOrigins   []string
	AllowCredentials bool
}

func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "localhost"),
			Port: getEnv("SERVER_PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:         getEnv("DB_HOST", "localhost"),
			Port:         getEnv("DB_PORT", "5432"),
			User:         getEnv("DB_USER", "postgres"),
			Password:     getEnv("DB_PASSWORD", "postgres"),
			Name:         getEnv("DB_NAME", "kova_delivery"),
			SSLMode:      getEnv("DB_SSLMODE", "disable"),
			MaxOpenConns: getEnvAsInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns: getEnvAsInt("DB_MAX_IDLE_CONNS", 5),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		Session: SessionConfig{
			Secret:           getEnv("SESSION_SECRET", ""),
			Duration:         getEnvAsDuration("SESSION_DURATION", 24*time.Hour),
			RefreshThreshold: getEnvAsDuration("SESSION_REFRESH_THRESHOLD", 1*time.Hour),
		},
		Security: SecurityConfig{
			BCryptCost:        getEnvAsInt("BCRYPT_COST", 12),
			RateLimitRequests: getEnvAsInt("RATE_LIMIT_REQUESTS", 100),
			RateLimitWindow:   getEnvAsDuration("RATE_LIMIT_WINDOW", 1*time.Minute),
		},
		Cookie: CookieConfig{
			Secure: getEnvAsBool("COOKIE_SECURE", false),
			Domain: getEnv("COOKIE_DOMAIN", "localhost"),
		},
		MQ: MQConfig{
			Type:        getEnv("MQ_TYPE", "redis"),
			URL:         getEnv("MQ_URL", "redis://localhost:6379/1"),
			RabbitMQURL: getEnv("RABBITMQ_URL", ""),
		},
		CORS: CORSConfig{
			AllowedOrigins:   strings.Split(getEnv("CORS_ALLOWED_ORIGINS", "http://localhost:3000"), ","),
			AllowCredentials: getEnvAsBool("CORS_ALLOW_CREDENTIALS", true),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Session.Secret == "" || len(c.Session.Secret) < 32 {
		return fmt.Errorf("SESSION_SECRET must be at least 32 characters")
	}

	if c.Security.BCryptCost < 10 || c.Security.BCryptCost > 14 {
		return fmt.Errorf("BCRYPT_COST must be between 10 and 14")
	}

	return nil
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.Name,
		c.Database.SSLMode,
	)
}

func (c *Config) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Redis.Host, c.Redis.Port)
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolVal, err := strconv.ParseBool(value); err == nil {
			return boolVal
		}
	}
	return defaultValue
}

func getEnvAsDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}
