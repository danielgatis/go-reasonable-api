package providers

import (
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"

	"github.com/rs/zerolog"
)

func ProvideLogger(cfg *config.Config) *zerolog.Logger {
	logger.Init(cfg)
	return logger.Get()
}
