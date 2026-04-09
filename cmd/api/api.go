package api

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"
	"go-reasonable-api/support/sentry"
	"go-reasonable-api/support/wire"

	"github.com/labstack/echo/v5"
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

	e := router.Setup()

	// Create a context that is cancelled on SIGINT/SIGTERM for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sc := echo.StartConfig{
		Address:    ":" + router.Port(),
		HideBanner: true,
	}

	logger.Info().Str("port", router.Port()).Msg("starting API server")
	if err := sc.Start(ctx, e); err != nil {
		logger.Error().Err(err).Msg("server error")
	}

	logger.Info().Msg("server shutdown complete")
	return nil
}
