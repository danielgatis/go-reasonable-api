// Package db provides database utilities including transaction management.
package db

import (
	"context"
	"database/sql"

	"github.com/rotisserie/eris"
)

// TxManager provides transaction lifecycle management with automatic rollback.
type TxManager struct {
	db *sql.DB
}

func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTx executes fn within a transaction. If fn returns an error or panics,
// the transaction is rolled back. On success, the transaction is committed.
//
// Repositories should use WithTx to participate:
//
//	err := tm.RunInTx(ctx, func(tx *sql.Tx) error {
//	    return userRepo.WithTx(tx).Create(ctx, ...)
//	})
func (tm *TxManager) RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return eris.Wrap(err, "failed to begin transaction")
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			// Wrap the original error with rollback failure info
			return eris.Wrapf(err, "rollback also failed: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return eris.Wrap(err, "failed to commit transaction")
	}
	return nil
}

func (tm *TxManager) DB() *sql.DB {
	return tm.db
}
