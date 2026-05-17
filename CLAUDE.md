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

### Wrapping rule

**Always wrap with `eris.Wrap` at every layer that propagates an error from
another package**, so the deepest possible stack frame is captured. `eris`
records the call stack at the moment `Wrap` is invoked, using `runtime.Callers`;
by the time the error reaches the next layer, the inner frame has already
returned and is gone. The frame itself is the value — without the repo wrap you
lose the exact `user_repository.go:NN` line from the trace and have to grep to
find which SQL call failed.

Wrap at every layer boundary:

- **Repo → sqlc/pgx**: every repo method that propagates a sqlc/pgx error wraps
  with a short description of what the repo was doing
  (`eris.Wrap(err, "failed to get user by id")`). Sentinel detection
  (`eris.Is(err, pgx.ErrNoRows)`) still works through wraps, so the service can
  match on the original error.
- **Service → repo**: same pattern. When a service method calls `s.repo.Foo()`
  and propagates the error, wrap with what the service was doing.
- **HTTP handler → service**: every handler that returns a service error wraps
  with `eris.Wrap(err, "failed to X")`. The middleware error handler walks the
  chain via `errors.As` to recover the original `AppError` for the status code,
  so wrapping doesn't break the response.
- **Worker handler → service**: same as HTTP handlers — wrap so the asynq log
  line shows where the failure originated.

Don't wrap when:

- Returning a domain sentinel directly (`return apperrors.ErrUserNotFound`) —
  wrapping a sentinel obscures it from `errors.As` checks downstream.
- Re-returning an error you already wrapped one level up in the same function.
  Each wrap adds a frame; double-wrapping the same call site is noise.

This rule is enforced by the `wrapcheck` linter in `.golangci.yml`. If
`make lint` complains about an unwrapped error from an external package, add an
`eris.Wrap` with a description of what the current function was doing.

If you find yourself reading a stack trace and asking "wait, where did this come
from?", that's a missing wrap.

## Code Generation

Run `make generate` after modifying:
- `db/queries/*.sql` or `db/migrations/*.sql` → regenerates sqlc
- `app/interfaces/` → regenerates mocks
- `support/wire/wire.go` → regenerates wire_gen.go
- Handler annotations → regenerates swagger docs
- `emails/src/templates/*.tsx` → regenerates HTML templates
