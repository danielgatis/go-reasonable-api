# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Development Commands

```bash
make dev              # Hot reload development (API + Worker)
make build            # Build binary (runs tidy, lint, fmt, generate)
make lint             # Run golangci-lint
make generate         # Generate all code (wire, swagger, mocks, sqlc, emails)

# Migrations
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make migrate-status   # Check migration status
make migrate-create name=add_foo  # Create new migration

# Run tests
go test ./...                           # All tests
go test ./app/services/...              # Single package
go test -run TestUserService ./...      # Single test
```

## Architecture

This is a Go API using clean architecture with dependency injection via Wire.

**Layer structure:**
- `api/handlers` - HTTP handlers (receive requests, call services, return responses)
- `app/services` - Business logic (orchestrates repositories, contains domain rules)
- `app/repositories` - Data access (wraps sqlc-generated queries)
- `app/interfaces` - Contracts between layers (repository and service interfaces)
- `support/` - Infrastructure (config, logging, email, http middleware, wire providers)

**Key patterns:**
- Wire injects dependencies at startup (`support/wire/wire.go`)
- Two entry points: `InitializeRouter()` for API, `InitializeWorker()` for background jobs
- sqlc generates type-safe queries from `db/queries/*.sql` into `db/sqlcgen/`
- Background jobs use Asynq (Redis-backed) with tasks defined in `app/tasks/`

## Error Handling

Uses `eris` for error wrapping. Standard `errors` package is blocked by linter.

```go
// Domain errors: return directly
return apperrors.ErrUserNotFound

// Infrastructure errors: wrap with context
return eris.Wrap(err, "failed to get user")

// New internal errors
return eris.New("something went wrong")
```

Domain errors are defined in `app/errors/errors.go` using helpers from `support/errors/`.

## Code Generation

Run `make generate` after modifying:
- `db/queries/*.sql` or `db/migrations/*.sql` → regenerates sqlc
- `app/interfaces/` → regenerates mocks
- `support/wire/wire.go` → regenerates wire_gen.go
- Handler annotations → regenerates swagger docs
- `emails/src/templates/*.tsx` → regenerates HTML templates
