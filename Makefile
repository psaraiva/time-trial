.PHONY: deps build run vet lint test coverage ci sabotage reset-sabotage help

BASE_URL ?= http://localhost:7777
CODE     ?= 500

deps: ## Download and tidy Go dependencies
	go mod tidy

build: ## Compile binary to bin/server
	go build -o bin/server ./cmd/

run: ## Run server locally
	go run ./cmd/

vet: ## Run go vet
	go vet ./...

lint: ## Run golangci-lint
	golangci-lint run ./...

ci: vet lint test ## Run vet, lint and tests (mirrors CI pipeline)

test: ## Run unit tests with race detector
	go test -race ./...

coverage: ## Generate test coverage file (coverage.out)
	go test ./... -coverprofile=coverage.out -covermode=atomic
	go tool cover -func=coverage.out

sabotage: ## POST /sabotage with JSON body (default code=500)
	curl -s -X POST "$(BASE_URL)/sabotage" -H "Content-Type: application/json" -d '{"code":$(CODE)}' | jq .

reset-sabotage: ## POST /sabotage with empty body — reset to random
	curl -s -X POST "$(BASE_URL)/sabotage" | jq .

help: ## List all available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-18s\033[0m %s\n", $$1, $$2}'
