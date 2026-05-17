package providers

import (
	"context"

	"go-reasonable-api/support/config"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rotisserie/eris"
)

func ProvideDB(cfg *config.Config) (*pgxpool.Pool, func(), error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.Database.URL)
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to parse database url")
	}

	poolCfg.MaxConns = int32(cfg.Database.MaxOpenConns)
	poolCfg.MinConns = int32(cfg.Database.MaxIdleConns)
	poolCfg.MaxConnLifetime = cfg.Database.ConnMaxLifetime
	poolCfg.MaxConnIdleTime = cfg.Database.ConnMaxIdleTime

	ctx := context.Background()
	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to create database pool")
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, nil, eris.Wrap(err, "failed to ping database")
	}

	cleanup := func() {
		pool.Close()
	}

	return pool, cleanup, nil
}
