// Package repositories defines data access contracts.
//
// Repositories abstract database operations behind interfaces, enabling
// transaction support and testability via mocks.
//
// # Transaction Support
//
// All repositories implement WithTx(tx *sql.Tx) to participate in
// transactions managed by support/db.TxManager. The returned repository
// uses the transaction connection; the original instance is unchanged.
//
//	err := txManager.RunInTx(ctx, func(tx *sql.Tx) error {
//	    userRepo := r.userRepo.WithTx(tx)
//	    tokenRepo := r.tokenRepo.WithTx(tx)
//	    // operations on both repos share the transaction
//	})
//
// # Cleanup Methods
//
// Repositories with expirable records (tokens, resets) expose Delete*
// methods for batch cleanup. These are called by scheduled tasks,
// not by application code.
package repositories
