package worker

import (
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
		Use:   "worker",
		Short: "Start the background worker",
		RunE:  run,
	}
}

func run(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return eris.Wrap(err, "failed to load config")
	}

	logger.Init(cfg)

	sentryCleanup, err := sentry.Init(cfg, "worker")
	if err != nil {
		return eris.Wrap(err, "failed to initialize sentry")
	}
	defer sentryCleanup()

	w, cleanup, err := wire.InitializeWorker()
	if err != nil {
		return eris.Wrap(err, "failed to initialize worker")
	}
	defer cleanup()

	// Watch for Ctrl+C signal to trigger graceful shutdown
	ctrlc.Watch(func() {
		logger.Info().Msg("received shutdown signal, gracefully shutting down...")
		w.Shutdown()
	})

	return w.Run()
}
