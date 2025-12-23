package repositories

import (
	"context"
	"database/sql"
	"os"
	"testing"

	"go-reasonable-api/db/migrations"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/lib/pq"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
)

var testDB *sql.DB
var testContainer *tcpostgres.PostgresContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.Run(ctx,
		"postgres:16-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		tcpostgres.BasicWaitStrategies(),
	)
	if err != nil {
		panic(err)
	}
	testContainer = container

	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}
	testDB = db

	if err := runMigrations(db); err != nil {
		panic(err)
	}

	code := m.Run()

	_ = db.Close()
	_ = container.Terminate(ctx)

	os.Exit(code)
}

func runMigrations(db *sql.DB) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return err
	}

	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		return err
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	return nil
}

func setupTest(t *testing.T) *sql.Tx {
	tx, err := testDB.Begin()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback()
	})

	return tx
}
