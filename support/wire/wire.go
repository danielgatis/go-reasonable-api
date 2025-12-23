//go:build wireinject
// +build wireinject

// Package wire configures dependency injection using Google Wire.
//
// This file defines provider sets and injector functions. Wire generates
// wire_gen.go with concrete initialization code. Run `make generate` after
// modifying this file.
//
// # Provider Sets
//
// Providers are grouped by layer:
//   - BaseProviderSet: shared infrastructure (config, logger, email)
//   - RepositoryProviderSet: all repository implementations
//   - ServiceProviderSet: all service implementations
//   - HandlerProviderSet: all HTTP handlers
//   - APIProviderSet: combines above for the API server
//   - WorkerProviderSet: combines for the background worker
//
// # Adding New Dependencies
//
// To add a new service:
//  1. Define the interface in app/interfaces/services
//  2. Implement in app/services
//  3. Add to ServiceProviderSet with wire.Bind
//  4. Run `make generate`
package wire

import (
	"go-reasonable-api/api/handlers"
	"go-reasonable-api/app/interfaces/repositories"
	"go-reasonable-api/app/interfaces/services"
	repoImpl "go-reasonable-api/app/repositories"
	svcImpl "go-reasonable-api/app/services"
	"go-reasonable-api/support/config"
	"go-reasonable-api/support/http"
	"go-reasonable-api/support/wire/providers"
	"go-reasonable-api/support/worker"

	"github.com/google/wire"
	"github.com/hibiken/asynq"
)

// BaseProviderSet contains providers shared between API and Worker
var BaseProviderSet = wire.NewSet(
	config.Load,
	providers.ProvideLogger,
	providers.ProvideEmailSender,
)

// RepositoryProviderSet contains all repository providers
var RepositoryProviderSet = wire.NewSet(
	repoImpl.NewUserRepository,
	wire.Bind(new(repositories.UserRepository), new(*repoImpl.UserRepository)),
	repoImpl.NewAuthTokenRepository,
	wire.Bind(new(repositories.AuthTokenRepository), new(*repoImpl.AuthTokenRepository)),
	repoImpl.NewPasswordResetRepository,
	wire.Bind(new(repositories.PasswordResetRepository), new(*repoImpl.PasswordResetRepository)),
	repoImpl.NewEmailVerificationRepository,
	wire.Bind(new(repositories.EmailVerificationRepository), new(*repoImpl.EmailVerificationRepository)),
)

// ServiceProviderSet contains all service providers
var ServiceProviderSet = wire.NewSet(
	svcImpl.NewUserService,
	wire.Bind(new(services.UserService), new(*svcImpl.UserService)),
	svcImpl.NewSessionService,
	wire.Bind(new(services.SessionService), new(*svcImpl.SessionService)),
	svcImpl.NewPasswordResetService,
	wire.Bind(new(services.PasswordResetService), new(*svcImpl.PasswordResetService)),
	svcImpl.NewEmailVerificationService,
	wire.Bind(new(services.EmailVerificationService), new(*svcImpl.EmailVerificationService)),
)

// HandlerProviderSet contains all handler providers
var HandlerProviderSet = wire.NewSet(
	handlers.NewUserHandler,
	handlers.NewSessionHandler,
	handlers.NewPasswordResetHandler,
	handlers.NewEmailVerificationHandler,
	handlers.NewHealthHandler,
)

// APIProviderSet contains providers specific to the API
var APIProviderSet = wire.NewSet(
	BaseProviderSet,
	providers.ProvideDB,
	providers.ProvideTxManager,
	providers.ProvideAsynqClient,
	wire.Bind(new(handlers.RedisPinger), new(*asynq.Client)),
	providers.ProvideTaskClient,
	RepositoryProviderSet,
	ServiceProviderSet,
	HandlerProviderSet,
	http.NewRouter,
)

// WorkerProviderSet contains providers specific to the Worker
var WorkerProviderSet = wire.NewSet(
	BaseProviderSet,
	providers.ProvideDB,
	RepositoryProviderSet,
	providers.ProvideAsynqServer,
	providers.ProvideScheduler,
	providers.ProvideEmailTask,
	providers.ProvideCleanupTask,
	providers.ProvideTaskRegistry,
	providers.ProvideServeMux,
	providers.ProvideWorker,
)

// InitializeRouter creates the API router with all dependencies.
// The cleanup function closes database connections and should be
// deferred in main.
func InitializeRouter() (*http.Router, func(), error) {
	wire.Build(APIProviderSet)
	return nil, nil, nil
}

// InitializeWorker creates the background worker with all dependencies.
// The cleanup function closes database connections and should be
// deferred in main.
func InitializeWorker() (*worker.Worker, func(), error) {
	wire.Build(WorkerProviderSet)
	return nil, nil, nil
}
