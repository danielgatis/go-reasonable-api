// Package services implements business logic for the application.
//
// Services orchestrate repositories and external dependencies to fulfill
// use cases. They own transaction boundaries and enforce domain rules.
//
// # Dependencies
//
// Services receive dependencies via constructor injection (for Wire):
//   - Repositories for data access
//   - TxManager for transactions
//   - TaskClient for async operations
//   - Config for behavior customization
//
// # Token Security
//
// Auth tokens, password resets, and email verifications use a two-part scheme:
//  1. A random token is generated and sent to the user
//  2. Only the SHA-256 hash is stored in the database
//
// This ensures that database access doesn't compromise active tokens.
// See crypto.go for GenerateSecureToken and HashToken utilities.
package services
