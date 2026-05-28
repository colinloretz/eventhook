package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	DatabaseURL   string
	RedisURL      string
	APIKey        string
	Port          int
	WorkerCount   int
	MaxPayloadKB  int
	Env           string
}

func Load() *Config {
	viper.SetEnvPrefix("EVENTHOOK")
	viper.AutomaticEnv()

	viper.SetDefault("PORT", 7676)
	viper.SetDefault("WORKER_COUNT", 10)
	viper.SetDefault("MAX_PAYLOAD_KB", 1024)
	viper.SetDefault("ENV", "development")
	viper.SetDefault("DATABASE_URL", "postgres://eventhook:eventhook@localhost:5432/eventhook")
	viper.SetDefault("REDIS_URL", "redis://localhost:6379")
	viper.SetDefault("API_KEY", "dev-api-key")

	return &Config{
		DatabaseURL:  viper.GetString("DATABASE_URL"),
		RedisURL:     viper.GetString("REDIS_URL"),
		APIKey:       viper.GetString("API_KEY"),
		Port:         viper.GetInt("PORT"),
		WorkerCount:  viper.GetInt("WORKER_COUNT"),
		MaxPayloadKB: viper.GetInt("MAX_PAYLOAD_KB"),
		Env:          viper.GetString("ENV"),
	}
}
