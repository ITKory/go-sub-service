package config

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

func Load() (*Config, error) {
	vip := viper.New()
	vip.SetConfigName("config")
	vip.SetConfigType("yaml")
	vip.AddConfigPath(".")

	if err := vip.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var config Config
	if err := vip.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	if port := os.Getenv("SERVER_PORT"); port != "" {
		config.Server.Port = port
	}

	return &config, nil
}

func (c *Config) DatabaseURL() string {
	// try get from env
	if url := os.Getenv("DATABASE_URL"); url != "" {
		return url
	}

	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.Database.User,
		c.Database.Password,
		c.Database.Host,
		c.Database.Port,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}
