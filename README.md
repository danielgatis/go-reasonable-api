# go-reasonable-api

[![Go Report Card](https://goreportcard.com/badge/github.com/danielgatis/go-reasonable-api?style=flat-square)](https://goreportcard.com/report/github.com/danielgatis/go-reasonable-api)
[![Go Doc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/danielgatis/go-reasonable-api)

<img width="1536" height="1024" alt="logo" src="https://github.com/user-attachments/assets/1d52503e-b106-4bfc-929e-9dd045c3cac4" />

A production-ready Go API template, distributed as a [Copier](https://copier.readthedocs.io/) template. Run one command, answer a few questions, and you get a fresh Go project with PostgreSQL, Redis-backed background jobs, email, authentication, and Swagger docs already wired up.

## What you get

- **Echo** HTTP framework with structured logging, rate limiting, CORS, panic recovery and graceful shutdown
- **PostgreSQL** with type-safe queries via [sqlc](https://sqlc.dev/) and migrations via [golang-migrate](https://github.com/golang-migrate/migrate)
- **Authentication** out of the box: registration, login, logout, password reset, email verification, account deletion with grace period
- **Background jobs** with [Asynq](https://github.com/hibiken/asynq) (Redis-backed) — retries, scheduling, dead-letter queues
- **Email** with MailHog for development and SendGrid for production; templates authored in [React Email](https://react.email/)
- **Compile-time DI** via [Wire](https://github.com/google/wire) — no runtime reflection, no global state
- **Swagger UI** auto-generated from code annotations
- **Sentry** integration ready (just set `SENTRY_DSN`)
- **Hot reload** for both API and worker processes
- **CLAUDE.md** preconfigured for AI-assisted development

For the full feature tour and architecture notes, see the generated project's `README.md` (rendered from [`README.md.jinja`](README.md.jinja)).

## Requirements

- [Copier](https://copier.readthedocs.io/) `>= 9.0.0` (install with `pipx install copier` or run on demand with `uvx copier`)
- Go `>= 1.25`
- Node.js `>= 22` (for the email templates)
- Docker + Docker Compose (for local PostgreSQL, Redis, MailHog)

## Usage

Generate a new project from this template:

```bash
copier copy gh:danielgatis/go-reasonable-api my-new-api
```

Or, without installing Copier globally:

```bash
uvx copier copy gh:danielgatis/go-reasonable-api my-new-api
```

Copier will prompt you for the following values (sensible defaults are derived from `project_name`, so you can mostly press Enter):

| Variable | Example | Used in |
|----------|---------|---------|
| `project_name` | `My New API` | README title, default for other vars |
| `project_slug` | `my-new-api` | binary name, docker container names, defaults for db_*, GitHub repo |
| `project_description` | `Backend API` | Swagger description, CLI `--help`, go.mod comment |
| `module_path` | `github.com/myorg/my-new-api` | `module` line in `go.mod`, all Go imports |
| `brand_name` | `Acme` | Swagger title, email subjects, "From" name on outgoing email |
| `from_email` | `noreply@acme.com` | development default for outgoing email |
| `db_user` / `db_password` / `db_name` | `my-new-api` | docker-compose, default DSN in `config.go` |
| `github_user` / `github_repo` | `myorg` / `my-new-api` | README badges |

After Copier finishes it runs a few tasks for you:

1. Rewrites every `import "go-reasonable-api/..."` to `import "<module_path>/..."`.
2. Runs `go mod tidy` to populate `go.sum`.
3. Prints the next commands you should run.

Then, inside the generated directory:

```bash
cd my-new-api
docker compose up -d          # start Postgres + Redis + MailHog
make install                  # download Go deps and npm deps for emails
make generate                 # regenerate wire, sqlc, mockery, swag, email HTML
make migrate-up               # apply migrations
make dev                      # API + worker with hot reload
```

The API will be at `http://localhost:8080`, Swagger UI at `http://localhost:8080/swagger/index.html`, MailHog at `http://localhost:8025`.

## Updating an existing project

Copier supports re-running a template against an already-generated project to pick up upstream improvements:

```bash
copier update
```

You will be prompted for any new variables and shown a diff for files that drifted. Files you have edited are 3-way merged.

See [Copier's "Updating a project" docs](https://copier.readthedocs.io/en/stable/updating/) for the details.

## Repository layout (this template repo)

```
.
├── copier.yml              # template questions, exclusions, post-generation tasks
├── README.md               # this file (template-repo README)
├── README.md.jinja         # README that ends up in the generated project
├── go.mod.jinja            # module line is templated
├── Makefile.jinja          # ldflags + binary name templated
├── main.go.jinja           # CLI Use/Short templated
├── docker-compose.yml.jinja
├── .air.toml.jinja
├── Procfile.air.jinja
├── .mockery.yaml.jinja     # interface package paths templated
├── api/routes.go.jinja             # Swagger @title templated
├── support/config/config.go.jinja  # DSN, from/brand defaults templated
├── app/services/*.go.jinja         # email subjects templated
├── emails/package.json.jinja
├── emails/src/templates/*.tsx.jinja  # brand name in email bodies
└── ... everything else is copied verbatim
```

Files ending in `.jinja` are rendered with Jinja2; the `.jinja` suffix is stripped on output. The template uses `[[ var ]]` and `[% block %]` delimiters (configured via `_envops` in `copier.yml`) so that Go template syntax (`{{ .Field }}`) inside email templates and elsewhere passes through untouched.

Generated artifacts (`api/docs/`, `app/mocks/`, `support/wire/wire_gen.go`, `db/sqlcgen/`, `emails/templates/`, `go.sum`, `node_modules/`) are excluded from the template — the generated project rebuilds them via `make generate`.

## Contributing to the template

To develop the template locally:

```bash
# Smoke-test that a fresh generation produces a buildable project
uvx copier copy --trust --defaults --force . /tmp/template-smoke-test
cd /tmp/template-smoke-test
go build ./...
```

`--trust` is required because Copier will run the post-generation tasks (sed + go mod tidy).

PRs welcome — open one against `main`.

## License

MIT — see `LICENSE` (carried into generated projects as-is).
