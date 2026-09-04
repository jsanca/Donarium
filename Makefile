.PHONY: help postgres-up postgres-down postgres-logs server-build server-run server-up server-down health-live health-ready lint test build

GOLANGCI_LINT_VERSION := v2.7.2

help: ## Show available commands
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

postgres-up: ## Start PostgreSQL
	docker compose up -d postgres

postgres-down: ## Stop PostgreSQL and server
	docker compose down

postgres-logs: ## Tail PostgreSQL logs
	docker compose logs -f postgres

server-build: ## Build server binary locally
	cd server && CGO_ENABLED=0 go build -ldflags="-s -w" -o donarium ./cmd/donarium/

server-run: ## Run server locally (requires PostgreSQL)
	cd server && go run ./cmd/donarium/

server-up: ## Start all services via Docker Compose
	docker compose up -d --build

server-down: ## Stop all services
	docker compose down

server-logs: ## Tail server logs
	docker compose logs -f server

health-live: ## Check liveness endpoint
	@HTTP_PORT=$${HTTP_PORT:-8080}; \
	curl -sS -o /dev/null -w "HTTP %{http_code}\n" http://localhost:$$HTTP_PORT/health/live || \
		{ echo "FAIL"; exit 1; }

health-ready: ## Check readiness endpoint
	@HTTP_PORT=$${HTTP_PORT:-8080}; \
	curl -sS -o /dev/null -w "HTTP %{http_code}\n" http://localhost:$$HTTP_PORT/health/ready || \
		{ echo "FAIL"; exit 1; }

lint: ## Run golangci-lint via pinned Docker image
	@if [ -z "$$(find server -name '*.go' 2>/dev/null)" ]; then \
		echo "lint: No Go source files yet. Skipping."; \
	else \
		docker run --rm \
			-v "$$(PWD)/server:/app" \
			-w /app \
			golangci/golangci-lint:$(GOLANGCI_LINT_VERSION) \
			golangci-lint run; \
	fi

test: ## Run tests
	@if [ -z "$$(find server -name '*_test.go' 2>/dev/null)" ]; then \
		echo "test: No test files yet. Skipping."; \
	else \
		cd server && go test ./...; \
	fi

build: ## Build the server binary (alias for server-build)
	@$(MAKE) server-build

qa-up: ## Start QA environment (PostgreSQL + backend)
	docker compose -f compose.qa.yml up -d --build

qa-down: ## Stop QA environment (preserves data)
	docker compose -f compose.qa.yml down

qa-reset: ## Destroy QA data and recreate clean environment
	docker compose -f compose.qa.yml down -v
	docker compose -f compose.qa.yml up -d --build

qa-status: ## Show QA service status
	@echo "=== QA Services ==="
	@docker compose -f compose.qa.yml ps
	@echo ""
	@echo "=== Health Checks ==="
	@DONARIUM_QA_URL=$${DONARIUM_QA_URL:-http://127.0.0.1:18080}; \
	curl -sS "$$DONARIUM_QA_URL/health/live" || echo "Liveness: FAIL"
	@DONARIUM_QA_URL=$${DONARIUM_QA_URL:-http://127.0.0.1:18080}; \
	curl -sS "$$DONARIUM_QA_URL/health/ready" || echo "Readiness: FAIL"

qa-logs: ## Tail QA environment logs
	docker compose -f compose.qa.yml logs -f
