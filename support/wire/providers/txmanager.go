package providers

import (
	"database/sql"

	"go-reasonable-api/support/db"
)

func ProvideTxManager(database *sql.DB) *db.TxManager {
	return db.NewTxManager(database)
}
