// Package repositories implements data access using sqlc-generated queries.
//
// Each repository wraps a sqlcgen.Queries instance and implements the
// corresponding interface from app/interfaces/repositories.
//
// # Transaction Support
//
// Repositories implement WithTx to participate in transactions:
//
//	func (r *UserRepository) WithTx(tx *sql.Tx) repositories.UserRepository {
//	    return &UserRepository{queries: r.queries.WithTx(tx)}
//	}
//
// The returned repository uses the transaction; the original is unchanged.
// This allows composing multiple repository operations atomically.
//
// # Null Handling
//
// Repositories return nil (not an error) when a record is not found.
// Callers must check for nil before using the result.
package repositories
