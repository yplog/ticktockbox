package config

import (
	"os"
	"strconv"
)

type Config struct {
	ServerPort int

	Database struct {
		Path string
	}

	Notifier struct {
		UseRabbitMQ  bool
		RabbitMQURL  string
		UseWebSocket bool
	}
}

func Load() *Config {
	return &Config{
		ServerPort: getEnvAsInt("SERVER_PORT", 3000),
		Database: struct {
			Path string
		}{
			Path: getEnv("DATABASE_PATH", "./data/db"),
		},
		Notifier: struct {
			UseRabbitMQ  bool
			RabbitMQURL  string
			UseWebSocket bool
		}{
			UseWebSocket: getBoolEnv("NOTIFIER_USE_WEBSOCKET", true),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	strValue := getEnv(key, "")
	if value, err := strconv.Atoi(strValue); err == nil {
		return value
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return defaultValue
	}
	return boolValue
}
