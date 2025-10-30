package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
)

type Config struct {
	MQTTBroker   string
	MQTTClientID string
	MQTTTopic    string
	MQTTPort     int
	MQTTUseTLS   bool
	MQTTCAFile   string
	MQTTCertFile string
	MQTTKeyFile  string

	TelegramBotToken string
	TelegramChatID   int64

	HTTPPort int

	DBHost     string
	DBPort     int
	DBUser     string
	DBPassword string
	DBName     string
}

func LoadConfig() (*Config, error) {
	cfg := &Config{
		MQTTBroker:   getEnv("MQTT_BROKER", "localhost"),
		MQTTClientID: getEnv("MQTT_CLIENT_ID", "go-backend-service"),
		MQTTTopic:    getEnv("MQTT_TOPIC", "esp32/sensors"),
		MQTTPort:     getEnvInt("MQTT_PORT", 1883),
		MQTTUseTLS:   getEnvBool("MQTT_USE_TLS", false),
		MQTTCAFile:   getEnv("MQTT_CA_FILE", "./certs/ca.crt"),
		MQTTCertFile: getEnv("MQTT_CERT_FILE", "./certs/client.crt"),
		MQTTKeyFile:  getEnv("MQTT_KEY_FILE", "./certs/client.key"),

		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		TelegramChatID:   getEnvInt64("TELEGRAM_CHAT_ID", 0),
		HTTPPort:         getEnvInt("HTTP_PORT", 4000),
		DBHost:           getEnv("DB_HOST", "localhost"),
		DBPort:           getEnvInt("DB_PORT", 5432),
		DBUser:           getEnv("DB_USER", "soil_user"),
		DBPassword:       getEnv("DB_PASSWORD", ""),
		DBName:           getEnv("DB_NAME", "soil_watering"),
	}

	if cfg.TelegramBotToken == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.TelegramChatID == 0 {
		return nil, fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}
	if cfg.DBPassword == "" {
		return nil, fmt.Errorf("DB_PASSWORD is required")
	}

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid integer for %s='%s', using default %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvInt64(key string, defaultValue int64) int64 {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseInt(valueStr, 10, 64)
	if err != nil {
		log.Printf("Warning: Invalid int64 for %s='%s', using default %d", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func getEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}
	value, err := strconv.ParseBool(valueStr)
	if err != nil {
		log.Printf("Warning: Invalid boolean for %s='%s', using default %t", key, valueStr, defaultValue)
		return defaultValue
	}
	return value
}

func (c *Config) GetDSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.DBUser,
		c.DBPassword,
		c.DBHost,
		c.DBPort,
		c.DBName,
	)
}
