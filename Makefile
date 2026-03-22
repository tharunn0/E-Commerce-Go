APP_NAME := ecommerce-backend
CMD_PATH := cmd/main.go

DB_URL ?= postgres://postgres:postgres@localhost:5432/ecommerce?sslmode=disable
MIGRATIONS_PATH := migrations

BIN_DIR := bin

.PHONY: help
help:
	@echo "Available commands:"
	@echo " make run               - Run the application"
	@echo " make build             - Build binary"
	@echo " make build-release     - Build optimized production binary"
	@echo " make test              - Run tests with race detection"
	@echo " make fmt               - Format code"
	@echo " make lint              - Run linter (golangci-lint)"
	@echo " make clean             - Remove build artifacts"
	@echo " make migrate-up        - Apply all migrations"
	@echo " make migrate-down      - Rollback last migration"
	@echo " make migrate-create    - Create new migration (name=<migration_name>)"

.PHONY: run
run:
	go run $(CMD_PATH)

.PHONY: build
build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

.PHONY: build-release
build-release:
	mkdir -p $(BIN_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
	go build -ldflags="-s -w" -o $(BIN_DIR)/$(APP_NAME) $(CMD_PATH)

.PHONY: test
test:
	go test -race -cover ./...

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golangci-lint run

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

.PHONY: migrate-up
migrate-up:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" up

.PHONY: migrate-down
migrate-down:
	migrate -path $(MIGRATIONS_PATH) -database "$(DB_URL)" down 1

.PHONY: migrate-create
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_PATH) -seq $(name)