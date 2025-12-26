package config

import (
	"fmt"
	"os"

	"go.yaml.in/yaml/v4"
)

type Config struct {
	Database DatabaseConfig `yaml:"database"`
	Kafka    KafkaConfig    `yaml:"kafka"`
	Services ServicesConfig `yaml:"services"`
	Server   ServerConfig   `yaml:"server"`
}

type DatabaseConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"ssl_mode"`
}

type RedisConfig struct {
	Address  string `yaml:"address"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

type KafkaConfig struct {
	Brokers []string    `yaml:"brokers"`
	Topics  TopicsConfig `yaml:"topics"`
}

type TopicsConfig struct {
	RentEvents   string `yaml:"rent_events"`
	StatusEvents string `yaml:"status_events"`
}

type ServicesConfig struct {
	RentService string `yaml:"rent_service"`
}

type ServerConfig struct {
	APIGatewayPort int `yaml:"api_gateway_port"`
	RentServicePort int `yaml:"rent_service_port"`
	StatsServicePort int `yaml:"stats_service_port"`
}

func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = "config.yaml"
	}
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal YAML: %w", err)
	}

	return &config, nil
}
