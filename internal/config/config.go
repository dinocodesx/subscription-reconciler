package config

import "time"

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
