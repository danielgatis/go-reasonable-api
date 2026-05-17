// Package repositories implements data access using sqlc-generated queries.
//
// Each repository wraps a sqlcgen.Queries instance and implements the
// corresponding interface from app/interfaces/repositories.
//
// # Transaction Support
//
// Repositories implement WithTx to participate in transactions:
//
//	func (r *UserRepository) WithTx(tx pgx.Tx) repositories.UserRepository {
//	    return &UserRepository{queries: r.queries.WithTx(tx)}
//	}
//
// The returned repository uses the transaction; the original is unchanged.
// This allows composing multiple repository operations atomically.
//
// # Error Propagation
//
// Repositories wrap pgx/sqlc errors with eris.Wrap before returning, so the
// stack trace captures the repository line. Sentinels (pgx.ErrNoRows) remain
// detectable via eris.Is through wraps, so callers can still branch on
// "not found" cases.
package repositories
