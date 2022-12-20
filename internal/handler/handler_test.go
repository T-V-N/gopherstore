package handler_test

import (
	"github.com/T-V-N/gopherstore/internal/config"

	"github.com/caarlos0/env/v6"
)

func InitTestConfig() (*config.Config, error) {
	cfg := &config.Config{}
	err := env.Parse(cfg)

	if err != nil {
		return nil, err
	}

	return cfg, nil
}
