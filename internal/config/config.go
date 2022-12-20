package config

import (
	"flag"

	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:":8080"`
	DatabaseURI          string `env:"DATABASE_URI"`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:"http://127.0.0.1:8888"`
	JWTExpireTiming      int64  `env:"JWT_EXPIRE_TIMING" envDefault:"10000"`
	SecretKey            string `env:"SECRET_KEY" envDefault:"secret"`
	MigrationsPath       string `env:"MIGRATIONS_PATH" envDefault:"migrations"`
	CompressLevel        int    `env:"COMPRESS_LEVEL" envDefault:"5"`
	CheckOrderDelay      uint   `env:"CHECK_ORDER_DELAY" envDefault:"10"`
	CheckOrderInterval   uint   `env:"CHECK_ORDER_INTERVAL" envDefault:"10"`
	WorkerLimit          int    `env:"WORKER_LIMIT" envDefault:"10"`
	ContextCancelTimeout int    `env:"CONTEXT_CANCEL_AMOUNT" envDefault:"10"`
}

func Init() (*Config, error) {
	cfg := &Config{}
	err := env.Parse(cfg)

	if err != nil {
		return nil, err
	}

	flag.StringVar(&cfg.RunAddress, "a", cfg.RunAddress, "server address")
	flag.StringVar(&cfg.DatabaseURI, "d", cfg.DatabaseURI, "db address")
	flag.StringVar(&cfg.AccrualSystemAddress, "r", cfg.AccrualSystemAddress, "accrual system address")
	flag.Parse()

	return cfg, nil
}
