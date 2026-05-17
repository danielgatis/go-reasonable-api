package providers

import (
	"go-reasonable-api/support/db"

	"github.com/jackc/pgx/v5/pgxpool"
)

func ProvideTxManager(pool *pgxpool.Pool) *db.TxManager {
	return db.NewTxManager(pool)
}
