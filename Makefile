APP_NAME=rasanta
BUILD_DIR=bin
MAIN_PATH=cmd/server/main.go

.PHONY: help deps tidy build run run-alt dev air-install test fmt clean parser-sim

help:
	@echo "Available targets:"
	@echo "  deps      - Download dependencies"
	@echo "  tidy      - Tidy go.mod/go.sum"
	@echo "  build     - Build binary to bin/"
	@echo "  run       - Run server on default port"
	@echo "  dev       - Run hot reload using Air (.air.toml)"
	@echo "  air-install - Install Air CLI"
	@echo "  test      - Run tests"
	@echo "  fmt       - Format Go files"
	@echo "  clean     - Clean build artifacts"

deps:
	@go mod download

tidy:
	@go mod tidy

build:
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_PATH)

run:
	@go run $(MAIN_PATH)

dev:
	@air -c .air.toml

air-install:
	@go install github.com/air-verse/air@latest

test:
	@go test ./...

fmt:
	@go fmt ./...

clean:
	@go clean
	@rm -rf $(BUILD_DIR)
