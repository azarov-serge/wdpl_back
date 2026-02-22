package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config описывает конфигурацию приложения.
// Значения читаются из переменных окружения с помощью cleanenv.
type Config struct {
	ServerHost         string `env:"SERVER_HOST" env-default:"0.0.0.0"`
	ServerPort         int    `env:"SERVER_PORT" env-default:"3000"`
	DatabaseURL        string `env:"DATABASE_URL" env-required:"true"`
	JWTSecret          string `env:"JWT_SECRET" env-required:"true"`
	RefreshSecret      string `env:"REFRESH_SECRET" env-required:"true"`
	AccessTokenTTLMin  int    `env:"ACCESS_TOKEN_TTL" env-default:"15"`
	RefreshTokenTTLMin int    `env:"REFRESH_TOKEN_TTL" env-default:"30"`
	LogLevel           string `env:"LOG_LEVEL" env-default:"info"`
	LogFormat          string `env:"LOG_FORMAT" env-default:"text"`
	Environment        string `env:"ENVIRONMENT" env-default:"development"`
}

// Load загружает конфигурацию из переменных окружения.
// В dev окружении значения подхватываются из .env (загружается в main.go через godotenv).
func Load() (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, fmt.Errorf("load config from env: %w", err)
	}
	return &cfg, nil
}

// MustLoad как Load, но при отсутствии обязательных переменных (env-required) завершает программу через panic.
// Использовать в main при старте — без валидного конфига приложение не должно работать.
func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}
	return cfg
}

func (c *Config) ServerAddress() string {
	return fmt.Sprintf("%s:%d", c.ServerHost, c.ServerPort)
}
