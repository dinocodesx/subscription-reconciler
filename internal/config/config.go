package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr              string
	DatabaseURL           string
	SelfBaseURL           string
	ShutdownTimeout       time.Duration
	DBConnectTimeout      time.Duration
	DBConnectMaxAttempts  int
	CarrierPollInterval   time.Duration
	CarrierPollBatchSize  int
	NotificationInterval  time.Duration
	NotificationBatchSize int
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:              getEnv("HTTP_ADDR", ":8080"),
		DatabaseURL:           getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/adora?sslmode=disable"),
		SelfBaseURL:           getEnv("SELF_BASE_URL", "http://127.0.0.1:8080"),
		ShutdownTimeout:       10 * time.Second,
		DBConnectTimeout:      3 * time.Second,
		DBConnectMaxAttempts:  30,
		CarrierPollInterval:   5 * time.Minute,
		CarrierPollBatchSize:  25,
		NotificationInterval:  15 * time.Second,
		NotificationBatchSize: 50,
	}

	var err error

	if cfg.ShutdownTimeout, err = getDuration("SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DBConnectTimeout, err = getDuration("DB_CONNECT_TIMEOUT", cfg.DBConnectTimeout); err != nil {
		return Config{}, err
	}
	if cfg.DBConnectMaxAttempts, err = getInt("DB_CONNECT_MAX_ATTEMPTS", cfg.DBConnectMaxAttempts); err != nil {
		return Config{}, err
	}
	if cfg.CarrierPollInterval, err = getDuration("CARRIER_POLL_INTERVAL", cfg.CarrierPollInterval); err != nil {
		return Config{}, err
	}
	if cfg.CarrierPollBatchSize, err = getInt("CARRIER_POLL_BATCH_SIZE", cfg.CarrierPollBatchSize); err != nil {
		return Config{}, err
	}
	if cfg.NotificationInterval, err = getDuration("NOTIFICATION_INTERVAL", cfg.NotificationInterval); err != nil {
		return Config{}, err
	}
	if cfg.NotificationBatchSize, err = getInt("NOTIFICATION_BATCH_SIZE", cfg.NotificationBatchSize); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return fallback
}

func getDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return duration, nil
}

func getInt(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", key, err)
	}

	return parsed, nil
}
