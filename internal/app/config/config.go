package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/mcuadros/go-defaults"
)

type Config struct {
	HTTP     HTTPConfig
	DB       DBConfig
	Kafka    KafkaConfig
	JWT      JWTConfig
	Email    EmailConfig
}

type HTTPConfig struct {
	Addr string `env:"HTTP_ADDR" default:":8090"`
}

type DBConfig struct {
	URL string `env:"DB_URL" default:"postgres://postgres:postgres@localhost:5432/auth?sslmode=disable"`
}

type KafkaConfig struct {
	Brokers   []string `env:"KAFKA_BROKERS" default:"localhost:9092"`
	Topic     string   `env:"KAFKA_TOPIC" default:"emails"`
	UserTopic string   `env:"KAFKA_USER_TOPIC" default:"user-registered"`
}

type JWTConfig struct {
	Secret string `env:"JWT_SECRET" default:"your-secret-key"`
}

type EmailConfig struct {
	From string `env:"EMAIL_FROM" default:"auth-service@example.com"`
}

func New(filenames ...string) (*Config, error) {
	cfg := new(Config)

	if len(filenames) > 0 {
		if err := godotenv.Load(filenames...); err != nil {
			return nil, err
		}
	}

	if err := env.Parse(cfg); err != nil {
		return nil, err
	}

	defaults.SetDefaults(cfg)

	return cfg, nil
}
