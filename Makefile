.PHONY: deps build run fmt vet lint test coverage ci sabotage reset-sabotage swag help

BASE_URL ?= http://localhost:7777
CODE     ?= 500

deps: ## Download and tidy Go dependencies
	go mod tidy

build: ## Compile binary to bin/server
	go build -o bin/server ./cmd/

run: ## Run server locally
	go run ./cmd/

fmt: ## Format code with gofmt -s
	gofmt -s -w .

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

ci: fmt vet lint test ## Run fmt, vet, lint and tests (mirrors CI pipeline)

test: ## Run unit tests with race detector
	go test -race ./...

coverage: ## Generate test coverage file (coverage.out)
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

swag: ## Generate Swagger docs (requires: go install github.com/swaggo/swag/cmd/swag@latest)
	swag init -g cmd/main.go --output docs/swagger

help: ## List all available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
