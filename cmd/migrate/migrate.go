package migrate

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go-reasonable-api/db/migrations"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/logger"

	"github.com/golang-migrate/migrate/v4"
	migratepgx "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rotisserie/eris"
	"github.com/spf13/cobra"
)

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration commands",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return eris.Wrap(err, "failed to load config")
			}
			logger.Init(cfg)
			return nil
		},
	}

	cmd.AddCommand(newUpCommand())
	cmd.AddCommand(newDownCommand())
	cmd.AddCommand(newStatusCommand())
	cmd.AddCommand(newCreateCommand())

	return cmd
}

func newUpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "up",
		Short: "Run all pending migrations",
		RunE:  runUp,
	}
}

func newDownCommand() *cobra.Command {
	var steps int

	cmd := &cobra.Command{
		Use:   "down",
		Short: "Rollback migrations",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDown(steps)
		},
	}

	cmd.Flags().IntVarP(&steps, "steps", "n", 1, "Number of migrations to rollback")

	return cmd
}

func newStatusCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show migration status",
		RunE:  runStatus,
	}
}

func newCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create [name]",
		Short: "Create a new migration",
		Args:  cobra.ExactArgs(1),
		RunE:  runCreate,
	}
}

func getMigrator() (*migrate.Migrate, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, eris.Wrap(err, "failed to load config")
	}

	connConfig, err := pgx.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, eris.Wrap(err, "failed to parse database url")
	}

	db := stdlib.OpenDB(*connConfig)

	driver, err := migratepgx.WithInstance(db, &migratepgx.Config{})
	if err != nil {
		return nil, eris.Wrap(err, "failed to create driver")
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return nil, eris.Wrap(err, "failed to create source")
	}

	m, err := migrate.NewWithInstance("iofs", source, "pgx5", driver)
	if err != nil {
		return nil, eris.Wrap(err, "failed to create migrator")
	}

	return m, nil
}

func runUp(cmd *cobra.Command, args []string) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && !eris.Is(err, migrate.ErrNoChange) {
		return eris.Wrap(err, "migration failed")
	}

	logger.Info().Msg("migrations applied successfully")
	return nil
}

func runDown(steps int) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	if err := m.Steps(-steps); err != nil && !eris.Is(err, migrate.ErrNoChange) {
		return eris.Wrap(err, "rollback failed")
	}

	logger.Info().Int("steps", steps).Msg("rolled back migrations successfully")
	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	m, err := getMigrator()
	if err != nil {
		return err
	}

	version, dirty, err := m.Version()
	if err != nil && !eris.Is(err, migrate.ErrNilVersion) {
		return eris.Wrap(err, "failed to get version")
	}

	if eris.Is(err, migrate.ErrNilVersion) {
		logger.Info().Msg("no migrations applied yet")
		return nil
	}

	logger.Info().Uint("version", version).Bool("dirty", dirty).Msg("current migration version")
	return nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	name := args[0]
	timestamp := time.Now().Format("20060102150405")

	migrationsDir := "db/migrations"

	upFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.up.sql", timestamp, name))
	downFile := filepath.Join(migrationsDir, fmt.Sprintf("%s_%s.down.sql", timestamp, name))

	if err := os.WriteFile(upFile, []byte(""), 0644); err != nil {
		return eris.Wrap(err, "failed to create up migration")
	}

	if err := os.WriteFile(downFile, []byte(""), 0644); err != nil {
		return eris.Wrap(err, "failed to create down migration")
	}

	logger.Info().
		Str("up", upFile).
		Str("down", downFile).
		Msg("created migration files")

	return nil
}
