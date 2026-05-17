// Package db provides database utilities including transaction management.
package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/rotisserie/eris"
)

// Pool is the subset of *pgxpool.Pool methods TxManager needs.
// Defining it as an interface lets tests substitute pgxmock.
type Pool interface {
	Begin(ctx context.Context) (pgx.Tx, error)
}

// TxManager provides transaction lifecycle management with automatic rollback.
type TxManager struct {
	pool Pool
}

func NewTxManager(pool Pool) *TxManager {
	return &TxManager{pool: pool}
}

// RunInTx executes fn within a transaction. If fn returns an error or panics,
// the transaction is rolled back. On success, the transaction is committed.
//
// Repositories should use WithTx to participate:
//
//	err := tm.RunInTx(ctx, func(tx pgx.Tx) error {
//	    return userRepo.WithTx(tx).Create(ctx, ...)
//	})
func (tm *TxManager) RunInTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return eris.Wrap(err, "failed to begin transaction")
	}

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return eris.Wrapf(err, "rollback also failed: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return eris.Wrap(err, "failed to commit transaction")
	}
	return nil
}
