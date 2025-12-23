package providers

import (
	"database/sql"

	"go-reasonable-api/support/config"

	_ "github.com/lib/pq"
	"github.com/rotisserie/eris"
)

func ProvideDB(cfg *config.Config) (*sql.DB, func(), error) {
	db, err := sql.Open("postgres", cfg.Database.URL)
	if err != nil {
		return nil, nil, eris.Wrap(err, "failed to open database connection")
	}

	// Configure connection pool
	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, nil, eris.Wrap(err, "failed to ping database")
	}

	cleanup := func() {
		_ = db.Close()
	}

	return db, cleanup, nil
}
