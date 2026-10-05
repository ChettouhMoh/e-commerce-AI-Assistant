package app

import (
	"context"
	"fmt"

	"ecommerce-ai-assistant/internal/platform/config"
	"ecommerce-ai-assistant/internal/platform/logging"
)

type Dependencies struct {
	Config *config.Config
	Logger *logging.Logger
}

func Wire(ctx context.Context, cfg *config.Config) (*Dependencies, error) {
	logger := logging.New(cfg.AppEnv)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	return &Dependencies{Config: cfg, Logger: logger}, nil
}
