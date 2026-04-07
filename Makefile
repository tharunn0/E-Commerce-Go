APP_NAME := ecommerce-backend
CMD_PATH := cmd/main.go

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
