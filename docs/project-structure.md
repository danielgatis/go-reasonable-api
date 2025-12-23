# Project Structure

This document describes the responsibility and boundaries of each package.

## Overview

```
.
├── api/                    # HTTP interface layer
│   ├── docs/               # Generated Swagger documentation
│   ├── handlers/           # Request handlers
│   ├── requests/           # Request DTOs with validation tags
│   └── responses/          # Response DTOs
│
├── app/                    # Application core
│   ├── errors/             # Domain error definitions
│   ├── interfaces/         # Contracts between layers
│   │   ├── repositories/   # Data access interfaces
│   │   ├── services/       # Business logic interfaces
│   │   └── support/        # Infrastructure interfaces
│   ├── mocks/              # Generated test mocks
│   ├── repositories/       # Data access implementations
│   ├── services/           # Business logic implementations
│   └── tasks/              # Background job handlers
│
├── cmd/                    # CLI entry points
│   ├── api/                # API server command
│   ├── migrate/            # Database migration command
│   ├── version/            # Version info command
│   └── worker/             # Background worker command
│
├── db/                     # Database layer
│   ├── migrations/         # SQL migration files
│   ├── queries/            # sqlc query definitions
│   └── sqlcgen/            # Generated Go code
│
├── emails/                 # Email templates
│   ├── src/templates/      # React Email components
│   └── templates/          # Generated HTML
│
└── support/                # Infrastructure layer
    ├── config/             # Configuration loading
    ├── db/                 # Database utilities (TxManager)
    ├── email/              # Email sending abstraction
    ├── errors/             # AppError type
    ├── http/               # HTTP server, middleware, utilities
    ├── logger/             # Structured logging
    ├── sentry/             # Error reporting
    ├── taskqueue/          # Task queue client
    ├── version/            # Build version info
    ├── wire/               # Dependency injection
    └── worker/             # Background worker setup
```

## Package Responsibilities

### api/

The HTTP interface layer. Translates HTTP to service calls.

**handlers/** — Request handlers

```go
type UserHandler struct {
    userService    services.UserService
    sessionService services.SessionService
}

func (h *UserHandler) Create(c echo.Context) error {
    // 1. Bind and validate request
    // 2. Call service
    // 3. Transform to response DTO
}
```

Handlers should be thin. No business logic. No direct database access.

**requests/** — Input validation

```go
type CreateUserRequest struct {
    Name     string `json:"name" validate:"required,min=2,max=100"`
    Email    string `json:"email" validate:"required,email"`
    Password string `json:"password" validate:"required,min=8"`
}
```

Validation tags are enforced by the validator middleware. Invalid requests never reach handlers.

**responses/** — Output shaping

```go
type UserResponse struct {
    ID              uuid.UUID  `json:"id"`
    Name            string     `json:"name"`
    Email           string     `json:"email"`
    EmailVerifiedAt *time.Time `json:"email_verified_at"`
}
```

Response DTOs control what's exposed to clients. Never return database models directly.

**routes.go** — Route registration

Central place for all route definitions. Makes it easy to see the full API surface.

---

### app/

The application core. Business logic lives here.

**interfaces/** — Contracts

Interfaces are defined by consumers, not implementers. Services define what they need from repositories:

```go
// app/interfaces/repositories/user.go
type UserRepository interface {
    WithTx(tx *sql.Tx) UserRepository
    Create(ctx context.Context, ...) (*sqlcgen.User, error)
    GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error)
}
```

This inverts dependencies: services depend on abstractions, not concrete implementations.

**services/** — Business logic

Services orchestrate operations and enforce rules:

```go
func (s *UserService) Create(ctx context.Context, name, email, password string) (*sqlcgen.User, error) {
    // Check if email exists (business rule)
    // Hash password (security concern)
    // Create user in transaction
    // Queue welcome email
}
```

Services:
- Own transaction boundaries
- Validate business rules
- Coordinate multiple repositories
- Enqueue background tasks

**repositories/** — Data access

Thin wrappers around sqlc-generated queries:

```go
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error) {
    user, err := r.queries.GetUserByEmail(ctx, email)
    if err == sql.ErrNoRows {
        return nil, nil  // Not found is not an error
    }
    return &user, err
}
```

Repositories:
- Execute queries
- Handle sql.ErrNoRows (return nil, not error)
- Support transactions via WithTx

**errors/** — Domain errors

Sentinel errors for known business conditions:

```go
var ErrUserNotFound = errors.NotFoundf("user")
var ErrEmailAlreadyExists = errors.New("EMAIL_ALREADY_EXISTS", "email already exists")
```

Services return these directly. The HTTP layer maps them to status codes.

**tasks/** — Background jobs

Task handlers for async operations:

```go
func (t *EmailTask) Handle(ctx context.Context, task *asynq.Task) error {
    // Unmarshal payload
    // Send email
    // Return error to retry, nil to complete
}
```

Tasks are registered in `registry.go` and processed by the worker.

---

### cmd/

CLI commands using Cobra.

**api/** — Starts the HTTP server

```go
func NewCommand() *cobra.Command {
    return &cobra.Command{
        Use:   "api",
        Short: "Start the API server",
        RunE:  run,
    }
}
```

Initializes dependencies via Wire, sets up graceful shutdown.

**worker/** — Starts the background worker

Same pattern as api. Runs Asynq server with registered task handlers.

**migrate/** — Database migrations

Subcommands: `up`, `down`, `status`, `create`. Uses golang-migrate.

---

### db/

Database layer.

**migrations/** — SQL files

```
0001_initial_schema.up.sql
0001_initial_schema.down.sql
```

Migrations are embedded in the binary. No external files needed at runtime.

**queries/** — sqlc definitions

```sql
-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING *;
```

sqlc generates type-safe Go code from these.

**sqlcgen/** — Generated code

Never edit directly. Regenerate with `make generate`.

---

### emails/

Email templates using React Email.

**src/templates/** — React components

```tsx
export const Welcome = ({ name, verificationUrl }) => (
    <Email>
        <Text>Welcome, {name}!</Text>
        <Button href={verificationUrl}>Verify Email</Button>
    </Email>
);
```

**templates/** — Generated HTML

React components are compiled to HTML. The Go code uses these HTML files.

---

### support/

Infrastructure that doesn't contain business logic.

**config/** — Configuration

```go
type Config struct {
    Database DatabaseConfig
    Auth     AuthConfig
    // ...
}

func Load() (*Config, error) {
    // Load from env, files, defaults
}
```

Viper handles loading from multiple sources with precedence.

**db/** — Database utilities

```go
type TxManager struct {
    db *sql.DB
}

func (tm *TxManager) RunInTx(ctx context.Context, fn func(tx *sql.Tx) error) error {
    // Begin, execute, commit or rollback
}
```

**errors/** — Error types

```go
type AppError struct {
    Code       string
    Message    string
    StatusCode int
    Details    map[string]any
}
```

The foundation for domain errors. Includes HTTP status mapping.

**http/** — Server setup

- `router.go` — Echo configuration, middleware setup
- `error_handler.go` — Global error transformation
- `validator.go` — Request validation
- `middlewares/` — Auth, logging, rate limiting, etc.
- `reqctx/` — Request-scoped context values

**logger/** — Structured logging

```go
logger.Info().Str("user_id", id).Msg("user created")
```

Uses zerolog. Context-aware (request ID, user ID).

**taskqueue/** — Task enqueueing

```go
func (c *Client) EnqueueCtx(ctx context.Context, taskType string, payload any, opts ...asynq.Option) {
    // Wrap payload with metadata
    // Enqueue to Redis
    // Log errors but don't return them (fire-and-forget)
}
```

**wire/** — Dependency injection

- `wire.go` — Provider sets and injector definitions
- `wire_gen.go` — Generated initialization code
- `providers/` — Individual provider functions

**worker/** — Background worker

```go
type Worker struct {
    server    *asynq.Server
    mux       *asynq.ServeMux
    scheduler *asynq.Scheduler
}

func (w *Worker) Run() error {
    // Start scheduler for periodic tasks
    // Start server for queue processing
}
```

---

## Boundaries and Rules

### What Goes Where

| I need to... | Put it in... |
|--------------|--------------|
| Handle an HTTP request | api/handlers |
| Validate request input | api/requests (validation tags) |
| Shape API response | api/responses |
| Implement business logic | app/services |
| Access the database | app/repositories |
| Define a service contract | app/interfaces/services |
| Define a repository contract | app/interfaces/repositories |
| Create a background job | app/tasks |
| Define domain error | app/errors |
| Add infrastructure | support/ |

### Import Rules

These are conventions, not enforced by tooling:

```
api/         → can import app/, support/
app/services → can import app/interfaces, app/errors, support/, db/sqlcgen
app/repos    → can import app/interfaces, db/sqlcgen, support/
app/tasks    → can import app/interfaces, support/
support/     → should not import api/ or app/
db/          → should not import anything except stdlib
```

### Extending Safely

**Adding a new entity:**

1. Add migration in `db/migrations/`
2. Add queries in `db/queries/`
3. Run `make generate` (creates sqlcgen code)
4. Add repository interface in `app/interfaces/repositories/`
5. Implement repository in `app/repositories/`
6. Add service interface in `app/interfaces/services/`
7. Implement service in `app/services/`
8. Add handler in `api/handlers/`
9. Add request/response DTOs in `api/requests/`, `api/responses/`
10. Wire up in `support/wire/wire.go`
11. Add routes in `api/routes.go`
12. Run `make generate` (creates wire code)

**Adding a new background job:**

1. Define task type constant and payload struct in `app/tasks/`
2. Implement handler
3. Register in `app/tasks/registry.go`
4. Enqueue from services via `TaskClient.EnqueueCtx()`

**Adding new configuration:**

1. Add field to appropriate config struct in `support/config/`
2. Add `viper.SetDefault()` in `Load()`
3. Access via injected `*config.Config`
