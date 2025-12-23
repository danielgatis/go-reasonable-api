# Architecture

This document explains the architectural decisions, patterns, and trade-offs in this template.

## Guiding Principles

### 1. Dependencies Flow Inward

The codebase follows a layered architecture where dependencies point inward:

```
HTTP Request
    ↓
┌─────────────────────────────────────┐
│  api/handlers (HTTP adapters)       │
└─────────────────────────────────────┘
    ↓ calls
┌─────────────────────────────────────┐
│  app/services (business logic)      │
└─────────────────────────────────────┘
    ↓ calls
┌─────────────────────────────────────┐
│  app/repositories (data access)     │
└─────────────────────────────────────┘
    ↓ uses
┌─────────────────────────────────────┐
│  db/sqlcgen (generated queries)     │
└─────────────────────────────────────┘
```

Inner layers don't know about outer layers. Services don't know about HTTP. Repositories don't know about services. This makes each layer independently testable and replaceable.

### 2. Interfaces at Boundaries

Interfaces are defined where they're needed, not where they're implemented:

```go
// app/interfaces/repositories/user.go
// The SERVICE defines what it needs from a repository
type UserRepository interface {
    Create(ctx context.Context, ...) (*sqlcgen.User, error)
    GetByEmail(ctx context.Context, email string) (*sqlcgen.User, error)
}

// app/repositories/user_repository.go
// The implementation satisfies the interface
type UserRepository struct { ... }
```

This inverts the dependency: services depend on abstractions they control, not on concrete implementations.

### 3. Explicit Over Magic

Every dependency is explicitly passed through constructors:

```go
func NewUserService(
    cfg *config.Config,
    txManager *db.TxManager,
    userRepo repositories.UserRepository,
    taskClient support.TaskClient,
) *UserService
```

No service locators. No context-based injection. No init() magic. When you read a constructor, you know exactly what the component needs.

## Request Lifecycle

A typical request flows through these stages:

```
1. HTTP Request arrives
       ↓
2. Middleware chain executes:
   - RequestID (generates unique ID)
   - Logger (structured logging)
   - RateLimiter (protects from abuse)
   - Recover (catches panics)
   - CORS (handles preflight)
       ↓
3. Route matches handler
       ↓
4. Handler:
   - Binds & validates request body
   - Extracts context values (user ID, etc.)
   - Calls service method
   - Returns response DTO
       ↓
5. Service:
   - Enforces business rules
   - Orchestrates repositories
   - Manages transactions
   - Enqueues background tasks
       ↓
6. Repository:
   - Executes sqlc-generated queries
   - Returns domain models
       ↓
7. Response serialized as JSON
```

## Dependency Injection with Wire

Wire generates initialization code at compile time. No reflection, no runtime overhead.

### How It Works

1. You define **providers** (functions that create things):

```go
// support/wire/providers/db.go
func ProvideDB(cfg *config.Config) (*sql.DB, func(), error) {
    db, err := sql.Open("postgres", cfg.Database.URL)
    cleanup := func() { db.Close() }
    return db, cleanup, err
}
```

2. You group providers into **sets**:

```go
// support/wire/wire.go
var RepositoryProviderSet = wire.NewSet(
    repoImpl.NewUserRepository,
    wire.Bind(new(repositories.UserRepository), new(*repoImpl.UserRepository)),
)
```

3. You define **injectors** that Wire implements:

```go
func InitializeRouter() (*http.Router, func(), error) {
    wire.Build(APIProviderSet)
    return nil, nil, nil  // Wire replaces this
}
```

4. Run `make generate` and Wire creates `wire_gen.go` with real initialization.

### Adding New Dependencies

1. Create the provider function in `support/wire/providers/`
2. Add to the appropriate provider set in `support/wire/wire.go`
3. Run `make generate`

Wire will fail at compile time if dependencies can't be satisfied. No runtime surprises.

## Error Handling Strategy

### Two Types of Errors

**Domain errors** are expected conditions in business logic:

```go
// app/errors/errors.go - sentinels
var ErrUserNotFound = errors.NotFoundf("user")
var ErrEmailAlreadyExists = errors.New("EMAIL_ALREADY_EXISTS", "email already exists")

// Usage in services - return directly
if user == nil {
    return apperrors.ErrUserNotFound
}
```

**Infrastructure errors** come from external systems:

```go
// Wrap to add context and preserve stack trace
user, err := s.userRepo.GetByEmail(ctx, email)
if err != nil {
    return eris.Wrap(err, "failed to get user by email")
}
```

### Why eris Instead of Standard errors?

The standard library's `errors.Wrap` loses the stack trace. When debugging production issues, you need to know *where* the error originated:

```go
// eris gives you:
// "failed to create user: failed to insert: pq: duplicate key value"
//   at services.(*UserService).Create (user_service.go:45)
//   at handlers.(*UserHandler).Create (user_handler.go:32)
```

### Error Response Transformation

The global error handler (`support/http/error_handler.go`) converts errors to JSON:

```go
// AppError → uses Code, Message, StatusCode directly
// echo.HTTPError → extracts status and message
// Other errors → 500 with generic message (logged, not exposed)
```

Server errors (5xx) are automatically reported to Sentry with request context.

## Transaction Management

### The Problem

Many operations need to update multiple tables atomically. If user creation succeeds but welcome email token creation fails, you have an inconsistent state.

### The Solution

Services own transaction boundaries using `TxManager`:

```go
func (s *UserService) Create(ctx context.Context, ...) (*sqlcgen.User, error) {
    var user *sqlcgen.User

    err := s.txManager.RunInTx(ctx, func(tx *sql.Tx) error {
        // All repositories use the same transaction
        userRepo := s.userRepo.WithTx(tx)
        verificationRepo := s.verificationRepo.WithTx(tx)

        user, err = userRepo.Create(ctx, ...)
        if err != nil {
            return err  // Automatic rollback
        }

        _, err = verificationRepo.Create(ctx, user.ID, ...)
        return err  // Commit if nil, rollback otherwise
    })

    return user, err
}
```

### Repository Transaction Support

Every repository implements `WithTx`:

```go
func (r *UserRepository) WithTx(tx *sql.Tx) repositories.UserRepository {
    return &UserRepository{
        queries: r.queries.WithTx(tx),
    }
}
```

This returns a new repository instance using the transaction. The original repository is unchanged (important for concurrent use).

## Background Job Processing

### Architecture

```
┌─────────────────┐     enqueue     ┌─────────────────┐
│  API Service    │ ───────────────→│  Redis Queue    │
└─────────────────┘                 └─────────────────┘
                                            │
                                            ↓ dequeue
                                    ┌─────────────────┐
                                    │  Worker Process │
                                    └─────────────────┘
```

The API and Worker are separate processes. This provides:
- Independent scaling (more workers for heavy email load)
- Isolation (worker crash doesn't affect API)
- Deployment flexibility (update worker without API downtime)

### Fire-and-Forget Enqueueing

Task enqueueing is intentionally fire-and-forget:

```go
// TaskClient.EnqueueCtx logs errors but doesn't return them
s.taskClient.EnqueueCtx(ctx, tasks.TypeEmail, payload, opts...)
```

**Why?** Email sending failures shouldn't break user registration. The task queue provides reliability through retries. If enqueueing itself fails (Redis down), it's logged for alerting, but the primary operation succeeds.

### Task Metadata

Tasks carry request context for distributed tracing:

```go
type TaskMetadata struct {
    RequestID string    `json:"request_id"`
    UserID    string    `json:"user_id,omitempty"`
    EnqueuedAt time.Time `json:"enqueued_at"`
}
```

When processing, the worker reconstructs the logger with this context:

```go
meta, err := taskqueue.UnwrapPayload(task.Payload(), &payload)
logger := t.logger.With().Str("request_id", meta.RequestID).Logger()
```

## Security Considerations

### Token Storage

Auth tokens, password reset tokens, and email verification tokens are never stored directly:

```go
// Generate random bytes, return hex-encoded string to user
token, _ := GenerateSecureToken(32)  // "a1b2c3..."

// Store only the SHA-256 hash
hash := HashToken(token)  // Store this
```

**Why?** If an attacker gains database access, they can't use the hashes to authenticate. They need the original tokens, which only exist in emails or client storage.

### Password Hashing

Passwords use bcrypt with configurable cost:

```go
cost := s.config.Auth.BcryptCost  // Default: 12
hash, _ := bcrypt.GenerateFromPassword([]byte(password), cost)
```

The cost is configurable because:
- Development: Lower cost = faster tests
- Production: Higher cost = more resistant to brute force

### Enumeration Prevention

Password reset and email verification endpoints don't reveal whether an email exists:

```go
func (s *PasswordResetService) Create(ctx context.Context, email string) error {
    user, err := s.userRepo.GetByEmail(ctx, email)
    if err != nil {
        return err
    }
    if user == nil {
        return nil  // Silently succeed - don't reveal email doesn't exist
    }
    // ... send reset email
}
```

### Account Deletion

Deletion is soft with a delay period:

```go
viper.SetDefault("auth.account_deletion_delay", "720h")  // 30 days
```

Users can cancel deletion by logging in. A scheduled task permanently deletes accounts after the delay. This protects against:
- Accidental deletion
- Account takeover followed by deletion
- Impulsive decisions

## Scaling Considerations

### Stateless API

The API server is stateless. All session state is in PostgreSQL, all job state is in Redis. You can run multiple API instances behind a load balancer.

### Database Connections

Connection pooling is configured in `config.go`:

```go
viper.SetDefault("database.max_open_conns", 25)
viper.SetDefault("database.max_idle_conns", 10)
viper.SetDefault("database.conn_max_lifetime", "5m")
```

For N API instances, total connections = N × max_open_conns. Size your PostgreSQL accordingly.

### Worker Concurrency

```go
viper.SetDefault("worker.concurrency", 10)
```

This is per worker process. Scale by running more worker processes, not just increasing concurrency (to maintain isolation and enable rolling deploys).

## Trade-offs and Alternatives

### sqlc vs ORM

**Chose sqlc because:**
- SQL is the source of truth, not Go structs
- Generated code is inspectable and type-safe
- No runtime reflection or query building
- Complex queries are just SQL, not method chains

**Trade-off:** More verbose for simple CRUD, but queries are explicit and optimized.

### Wire vs Other DI

**Chose Wire because:**
- Compile-time verification (no runtime panics)
- Generated code is readable and debuggable
- No reflection overhead

**Trade-off:** More ceremony for simple cases, but dependencies are always explicit.

### Separate API and Worker

**Chose separation because:**
- Independent scaling and deployment
- Clear responsibility boundaries
- Failure isolation

**Trade-off:** More operational complexity. For simple projects, you could run both in one process using goroutines.

### eris vs pkg/errors

**Chose eris because:**
- Better stack trace formatting
- JSON output for structured logging
- Active maintenance

**Trade-off:** Another dependency. Standard library errors work fine for simpler needs.
