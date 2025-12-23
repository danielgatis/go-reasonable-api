package main

//go:generate go tool wire ./support/wire
//go:generate go tool swag init -g main.go -o api/docs --parseDependency --parseInternal
//go:generate go tool mockery
//go:generate npm run --prefix emails build
//go:generate go tool sqlc generate -f db/sqlc.yaml
