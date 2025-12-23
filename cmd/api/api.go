package api

import (
	"context"
	"time"

	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"
	"go-reasonable-api/support/sentry"
	"go-reasonable-api/support/wire"

	ctrlc "github.com/danielgatis/go-ctrlc"
	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "api",
		Short: "Start the API server",
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return eris.Wrap(err, "failed to load config")
	}

	logger.Init(cfg)

	sentryCleanup, err := sentry.Init(cfg, "api")
	if err != nil {
		return eris.Wrap(err, "failed to initialize sentry")
	}
	defer sentryCleanup()

	router, cleanup, err := wire.InitializeRouter()
	if err != nil {
		return eris.Wrap(err, "failed to initialize router")
	}
	defer cleanup()

	echo := router.Setup()

	// Start server in a goroutine
	go func() {
		logger.Info().Str("port", router.Port()).Msg("starting API server")
		if err := echo.Start(":" + router.Port()); err != nil {
			logger.Info().Msg("server stopped")
		}
	}()

	// Wait for Ctrl+C signal
	done := make(chan struct{})
	ctrlc.Watch(func() {
		logger.Info().Msg("received shutdown signal, gracefully shutting down...")

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := echo.Shutdown(ctx); err != nil {
			logger.Error().Err(err).Msg("failed to shutdown server gracefully")
		}

		close(done)
	})

	<-done
	logger.Info().Msg("server shutdown complete")
	return nil
}
