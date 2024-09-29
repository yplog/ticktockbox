package config

import (
	"os"
	"strconv"
)

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Path string
}

type NotifierConfig struct {
	UseRabbitMQ  bool
	RabbitMQURL  string
	UseWebSocket bool
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Notifier NotifierConfig
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 3000),
		},
		Database: DatabaseConfig{
			Path: getEnv("DATABASE_PATH", "./data/db"),
		},
		Notifier: NotifierConfig{
			UseRabbitMQ:  getBoolEnv("NOTIFIER_USE_RABBITMQ", false),
			RabbitMQURL:  getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
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
