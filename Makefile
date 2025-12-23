VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_DATE ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS := -ldflags "-s -w \
	-X go-reasonable-api-template/support/version.Version=$(VERSION) \
	-X go-reasonable-api-template/support/version.Commit=$(COMMIT) \
	-X go-reasonable-api-template/support/version.BuildDate=$(BUILD_DATE)"

# Development with auto-reload for both API and Worker
.PHONY: dev
dev:
	@go tool air

# Build
.PHONY: build
build: tidy lint fmt generate
	@go build $(LDFLAGS) -o bin/go-reasonable-api-template .

# Run commands
.PHONY: run-api
run-api:
	@go run . api

.PHONY: run-worker
run-worker:
	@go run . worker

# Migrations
.PHONY: migrate-up
migrate-up:
	@go run . migrate up

.PHONY: migrate-down
migrate-down:
	@go run . migrate down

.PHONY: migrate-status
migrate-status:
	@go run . migrate status

.PHONY: migrate-create
migrate-create:
	@go run . migrate create $(name)

# Go mod tidy
.PHONY: tidy
tidy:
	@go mod tidy

# Lint
.PHONY: lint
lint:
	@go tool golangci-lint run ./...

# Format
.PHONY: fmt
fmt:
	@go fmt ./...

# Generate all code
.PHONY: generate
generate:
	@go generate gen.go

# Run React Email dev server for preview
.PHONY: emails-dev
emails-dev:
	@cd emails && npm run dev

# Install dependencies
.PHONY: install
install:
	@go mod download
	@cd emails && npm install
