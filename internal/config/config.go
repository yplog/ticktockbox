package config

import "os"

type Config struct {
	Port           string
	QuestDBURL     string
	QuestDBUser    string
	QuestDBPass    string
	QuestDBName    string
	QuestDBSSLMode string
	QuestDBHTTPURL string
	RabbitMQURL    string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "3000"),
		QuestDBURL:     getEnv("QUESTDB_URL", "localhost:8812"),
		QuestDBUser:    getEnv("QUESTDB_USER", "admin"),
		QuestDBPass:    getEnv("QUESTDB_PASSWORD", "quest"),
		QuestDBName:    getEnv("QUESTDB_NAME", "qdb"),
		QuestDBSSLMode: getEnv("QUESTDB_SSL_MODE", "disable"),
		QuestDBHTTPURL: getEnv("QUESTDB_HTTP_URL", "http://localhost:9000"),
		RabbitMQURL:    getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
