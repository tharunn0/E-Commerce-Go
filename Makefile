APP_NAME := ecommerce-backend

DB_URL ?= postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable
MIGRATIONS_PATH := migrations

.PHONY: help
help:
	@echo "Available commands:"
	@echo " make run             - Run the application"
	@echo " make test            - Run tests"
	@echo " make fmt             - Format code"
	@echo " make migrate-up      - Apply database migrations"
	@echo " make migrate-down    - Rollback last migration"
	@echo " make migrate-create  - Create new migration (name=<migration_name>)"

.PHONY: run
run:
	go run cmd/main.go

.PHONY: test
test:
	go test ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: migrate-up
	migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

.PHONY: migrate-down
	migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

.PHONY: migrate-create
	migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)
