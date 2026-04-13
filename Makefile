.PHONY: deps build run-webserver run-loadtest profile fmt vet lint test coverage ci sabotage reset-sabotage swag help

BASE_URL   ?= http://localhost:7777
CODE       ?= 500
URL_TARGET ?= http://localhost:7777/sabotage
N          ?= 1000
C          ?= 10
SECONDS    ?= 30

deps: ## Download and tidy Go dependencies
	go mod tidy

build: ## Compile binary to bin/server
	go build -o bin/server ./cmd/webserver/

run-webserver: ## Run server locally
	go run ./cmd/webserver/

run-loadtest: ## Run load test (flags: URL, N, C)
	go run ./cmd/loadtest/ -url $(URL_TARGET) -n $(N) -c $(C)

profile: ## Capture CPU pprof profile for SECONDS seconds and open web UI on :8080
	go tool pprof -http=:8080 http://localhost:6060/debug/pprof/profile?seconds=$(SECONDS)

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
