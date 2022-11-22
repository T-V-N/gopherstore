package config

import (
	"flag"
	"fmt"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccuralSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" `
	JWTExpireTiming      int    `env:"JWT_EXPIRE_TIMING" envDefault:"100000"`
	SecretKey            string `env:"SECRET_KEY" envDefault:"secret"`
}

func Init() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)

	if err != nil {
		return nil, fmt.Errorf("error: %w", err)
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "db address")
	flag.StringVar(&cfg.AccuralSystemAddress, "r", cfg.AccuralSystemAddress, "accural system address")
	flag.Parse()

	return cfg, nil
}
