package config

import (
	"encoding/json"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type ServerConfig struct {
	Port int
}

type DatabaseConfig struct {
	Path    string
	Checker uint8
}

type NotifierConfig struct {
	UseRabbitMQ  bool
	UseRedis     bool
	UseWebSocket bool

	RabbitMQURL string
	RedisURL    string
}

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Notifier NotifierConfig
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		return nil
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnvAsInt("SERVER_PORT", 3000),
		},
		Database: DatabaseConfig{
			Path:    getEnv("DATABASE_PATH", "./data/data.db"),
			Checker: uint8(getEnvAsInt("DATABASE_CHECKER_INTERVAL", 1)),
		},
		Notifier: NotifierConfig{
			UseRabbitMQ:  getBoolEnv("NOTIFIER_USE_RABBITMQ", false),
			UseRedis:     getBoolEnv("NOTIFIER_USE_REDIS", false),
			UseWebSocket: getBoolEnv("NOTIFIER_USE_WEBSOCKET", true),

			RabbitMQURL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			RedisURL:    getEnv("REDIS_URL", "redis://username:password@localhost:6379/"),
		},
	}
}

func (c *Config) Log() {
	indentJSON, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		log.Fatalf("JSON marshal error: %v", err)
	}

	log.Println("Loaded Config:", string(indentJSON))
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
