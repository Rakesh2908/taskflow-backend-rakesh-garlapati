package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

const (
	DefaultPort       = "8080"
	MinBcryptCost     = 12
	DefaultBcryptCost = 12
)

type Config struct {
	DBUrl      string
	JWTSecret  string
	Port       string
	BcryptCost int
}

func Load() Config {
	_ = godotenv.Load()

	cfg := Config{
		DBUrl:     os.Getenv("DB_URL"),
		JWTSecret: os.Getenv("JWT_SECRET"),
		Port:      os.Getenv("PORT"),
	}

	if cfg.JWTSecret == "" {
		panic("config: JWT_SECRET must not be empty")
	}

	if cfg.Port == "" {
		cfg.Port = DefaultPort
	}

	costStr := os.Getenv("BCRYPT_COST")
	if costStr == "" {
		cfg.BcryptCost = DefaultBcryptCost
	} else {
		cost, err := strconv.Atoi(costStr)
		if err != nil {
			panic("config: BCRYPT_COST must be an integer")
		}
		cfg.BcryptCost = cost
	}

	if cfg.BcryptCost < MinBcryptCost {
		cfg.BcryptCost = MinBcryptCost
	}

	return cfg
}

