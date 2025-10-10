package configs

import (
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

type config struct {
	Server ServerConfig   `koanf:"server"`
	DB     DatabaseConfig `koanf:"postgres"`
	Redis  RedisConfig    `koanf:"redis"`
	JWT    JWTConfig      `koanf:"jwt"`
}

type ServerConfig struct {
	Port string `koanf:"port"`
}

type DatabaseConfig struct {
	Host     string `koanf:"host"`
	Port     int    `koanf:"port"`
	User     string `koanf:"user"`
	Password string `koanf:"password"`
	DB       string `koanf:"db"`
	SSLMode  string `koanf:"sslmode"`
}

type RedisConfig struct {
	Addr     string `koanf:"addr"`
	Password string `koanf:"password"`
	DB       int    `koanf:"db"`
}

type JWTConfig struct {
	SecretKey              string `koanf:"secret_key"`
	ExpirationHours        int    `koanf:"expiration_hours"`
	RefreshExpirationHours int    `koanf:"refresh_expiration_hours"`
}

func LoadConfig(path string) *config {
	var k = koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		zap.L().Fatal("failed to load config", zap.Error(err))
	}

	var cfg config
	if err := k.Unmarshal("", &cfg); err != nil {
		zap.L().Fatal("failed to unmarshal config", zap.Error(err))
	}

	return &cfg
}
