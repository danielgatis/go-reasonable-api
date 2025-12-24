# go-reasonable-api

[![Go Report Card](https://goreportcard.com/badge/github.com/danielgatis/go-reasonable-api?style=flat-square)](https://goreportcard.com/report/github.com/danielgatis/go-reasonable-api)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/danielgatis/go-reasonable-api)

<img width="1536" height="1024" alt="logo" src="https://github.com/user-attachments/assets/1d52503e-b106-4bfc-929e-9dd045c3cac4" />

This template provides a production-ready foundation for building Go APIs. It makes opinionated choices about structure, tooling, and patterns—choices that have proven effective in real projects—while remaining flexible enough to adapt to your needs.

## Features

### Observability Out of the Box

Every log line includes `request_id` and `user_id` (when authenticated). Trace any request from ingress to database and back. **Sentry integration** is pre-configured—just set `SENTRY_DSN` and get full stack traces, request context, and error grouping in production. No setup, no boilerplate, it just works.

### Email That Just Works

Development uses **MailHog**—emails are captured locally at `http://localhost:8025`. Production uses **SendGrid**—flip one environment variable and you're sending real emails. Templates are built with **React Email**, so you write JSX components and get beautiful, tested HTML output.

### Authentication Ready to Ship

User registration, login, logout, password reset, email verification—all implemented, tested, and secure. **Account deletion** includes a 30-day grace period (configurable), protecting users from accidental or malicious deletions. Tokens are SHA-256 hashed before storage. Passwords use bcrypt with configurable cost.

### Radical Simplicity

Three layers. That's it. **Handlers** receive HTTP requests. **Services** contain business logic. **Repositories** talk to the database. No managers, no adapters, no abstract factory factories. Your domain code lives in `app/`, infrastructure lives in `support/`. When you need to find something, you know where to look.

### Developer Experience

**Hot reload** out of the box—save a file, see the change. **Swagger UI** auto-generated from code annotations. **Type-safe SQL** via sqlc—write SQL, get Go structs. **Compile-time DI** via Wire—if it compiles, dependencies are satisfied. Configuration via **Viper** with environment variables, files, and sensible defaults.

### Background Jobs Built In

Redis-backed job queue with **Asynq**. Automatic retries, scheduled tasks, dead letter queues. Email sending is already async—users don't wait for SMTP. Add your own jobs in minutes.

### Production Hardened

Rate limiting, graceful shutdown, health checks (database + Redis), CORS, panic recovery, request logging—all configured. Structured JSON logs ready for your log aggregator. Connection pooling tuned for real workloads.

### AI-Agent Ready

Built for the age of AI-assisted development. Ships with a **CLAUDE.md** that teaches AI agents how your codebase works—build commands, architecture patterns, error handling conventions. The `docs/` folder contains **architecture.md** and **project-structure.md** with the context AI agents need to make informed decisions. Claude Code, Cursor, Copilot—they all understand this codebase from day one.

## Philosophy

This template values:

- **Clarity over cleverness**: Code should be obvious to the next developer
- **Explicit dependencies**: No global state, no magic injection
- **Practical testing**: Interfaces where they enable testing, not everywhere
- **Minimal abstractions**: Only add layers that earn their complexity

## Quick Start

```bash
# Start infrastructure
docker compose up -d

# Install dependencies
make install

# Run database migrations
make migrate-up

# Start development server (hot reload)
make dev
```

The API runs at `http://localhost:8080`. Swagger docs at `http://localhost:8080/swagger/index.html`.

## Tech Stack

| Component | Choice | Why |
|-----------|--------|-----|
| Framework | Echo | Fast, minimal, good middleware ecosystem |
| Database | PostgreSQL | Reliable, feature-rich, excellent tooling |
| SQL | sqlc | Type-safe queries, no ORM complexity |
| DI | Wire | Compile-time injection, clear dependency graph |
| Background Jobs | Asynq | Redis-backed, simple API, good reliability |
| Validation | go-playground/validator | Standard, declarative, extensible |
| Errors | eris | Stack traces, wrapping, better debugging |

## Project Layout

```
.
├── api/            # HTTP layer (handlers, routes, request/response DTOs)
├── app/            # Application layer (services, repositories, domain errors)
├── cmd/            # CLI commands (api, worker, migrate)
├── db/             # Database (migrations, queries, generated code)
├── docs/           # Documentation
├── emails/         # Email templates (React Email)
└── support/        # Infrastructure (config, logging, middleware, DI)
```

See [docs/project-structure.md](docs/project-structure.md) for detailed package responsibilities.

## Development Commands

```bash
# Development
make dev              # Run API + Worker with hot reload
make run-api          # Run API only
make run-worker       # Run worker only

# Build & Quality
make build            # Build binary (includes lint, fmt, generate)
make lint             # Run linter
make generate         # Regenerate code (sqlc, wire, mocks, swagger)

# Database
make migrate-up       # Apply pending migrations
make migrate-down     # Rollback last migration
make migrate-status   # Show current migration version
make migrate-create name=add_users_table  # Create new migration

# Testing
go test ./...                          # All tests
go test ./app/services/...             # Single package
go test -run TestUserService ./...     # Single test
go test -v -race ./...                 # Verbose with race detection

# Email Templates
make emails-dev       # Preview email templates in browser
```

## Configuration

Configuration loads from environment variables (recommended) or `config.yaml`:

```bash
# Required for production
DATABASE_URL=postgres://user:pass@host:5432/dbname?sslmode=require
REDIS_ADDR=localhost:6379
AUTH_SECRET=your-secret-key-min-32-chars

# Email (choose one provider)
EMAIL_PROVIDER=sendgrid
EMAIL_SENDGRID_API_KEY=SG.xxx

# Or for development
EMAIL_PROVIDER=smtp
EMAIL_SMTP_HOST=localhost
EMAIL_SMTP_PORT=1025
```

See `support/config/config.go` for all options with defaults.

## API Endpoints

| Method | Path | Description | Auth |
|--------|------|-------------|------|
| POST | /users | Register new user | - |
| GET | /users/me | Get current user | Required |
| DELETE | /users/me | Schedule account deletion | Required |
| POST | /sessions | Login | - |
| DELETE | /sessions/current | Logout | Required |
| POST | /password-resets | Request password reset | - |
| PUT | /password-resets/:token | Complete password reset | - |
| POST | /email-verifications | Request verification email | Optional |
| PUT | /email-verifications/:token | Verify email | - |
| GET | /health | Health check | - |

## Architecture

See [docs/architecture.md](docs/architecture.md) for:
- Layered architecture and data flow
- Dependency injection patterns
- Error handling strategy
- Background job processing
- Security considerations

## Extending the Template

### Adding a New Feature

1. **Define the interface** in `app/interfaces/services/`
2. **Implement the service** in `app/services/`
3. **Add repository** if needed in `app/interfaces/repositories/` and `app/repositories/`
4. **Create handler** in `api/handlers/`
5. **Wire it up** in `support/wire/wire.go`
6. **Add routes** in `api/routes.go`
7. **Run `make generate`** to regenerate Wire code

### Adding a Background Job

1. Define task type and payload in `app/tasks/`
2. Implement handler function
3. Register in `app/tasks/registry.go`
4. Enqueue from services via `TaskClient.EnqueueCtx()`

### Adding a New Migration

```bash
make migrate-create name=add_notifications_table
# Edit db/migrations/XXXX_add_notifications_table.up.sql
# Edit db/migrations/XXXX_add_notifications_table.down.sql
make migrate-up
```

## Testing Strategy

- **Unit tests**: Services with mocked repositories (`app/services/*_test.go`)
- **Repository tests**: Against real database using testcontainers (`app/repositories/*_test.go`)
- **Handler tests**: HTTP tests with mocked services (`api/handlers/*_test.go`)

Mocks are generated with mockery. Run `make generate` after changing interfaces.

## Deployment

Build the binary:
```bash
make build
./bin/go-reasonable-api api      # Start API server
./bin/go-reasonable-api worker   # Start background worker
```

Both processes should run concurrently in production. Use your orchestrator's process management or a process supervisor.

Required infrastructure:
- PostgreSQL 14+
- Redis 6+
