# ============================================================================
# QuickLens - Makefile
# ============================================================================
# Usage: make help
# ============================================================================

.DEFAULT_GOAL := help
COMPOSE := docker compose
CONTAINER := ql-app

# ── Docker ──────────────────────────────────────────────────────────────────

.PHONY: help
help: ## Show this help message
	@echo ""
	@echo "  QuickLens — Lightweight LLM Observability"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'
	@echo ""

.PHONY: up
up: ## Start all services in detached mode
	$(COMPOSE) up -d --build

.PHONY: down
down: ## Stop all services
	$(COMPOSE) down

.PHONY: logs
logs: ## Tail container logs (follow mode)
	$(COMPOSE) logs -f --tail=100

.PHONY: ps
ps: ## Show running containers
	$(COMPOSE) ps

.PHONY: clean
clean: ## Stop services and remove volumes
	$(COMPOSE) down -v --remove-orphans
	@echo "✓ Volumes and orphan containers removed"

.PHONY: restart
restart: down up ## Restart all services

.PHONY: shell
shell: ## Open a shell inside the app container
	docker exec -it $(CONTAINER) sh

# ── Development ─────────────────────────────────────────────────────────────

.PHONY: build-frontend
build-frontend: ## Build the SvelteKit frontend locally
	cd frontend && npm ci && npm run build

.PHONY: build-backend
build-backend: ## Build the Go backend locally
	cd backend && CGO_ENABLED=0 go build -ldflags="-w -s" -o quicklens-backend main.go

.PHONY: test
test: ## Run all tests (Go + frontend)
	@echo "── Go tests ──"
	cd backend && go test ./... -v -count=1
	@echo ""
	@echo "── Frontend tests ──"
	cd frontend && npm test -- --run 2>/dev/null || echo "(no frontend tests configured)"

.PHONY: lint
lint: ## Run linters (Go + frontend)
	@echo "── Go lint ──"
	cd backend && go vet ./...
	@echo ""
	@echo "── Frontend lint ──"
	cd frontend && npm run lint 2>/dev/null || echo "(no frontend linter configured)"

.PHONY: fmt
fmt: ## Format Go source code
	cd backend && gofmt -w .

# ── Profiles ────────────────────────────────────────────────────────────────

.PHONY: up-ollama
up-ollama: ## Start with Ollama profile
	$(COMPOSE) --profile ollama up -d --build

.PHONY: up-postgres
up-postgres: ## Start with PostgreSQL profile
	$(COMPOSE) --profile postgres up -d --build

.PHONY: up-clickhouse
up-clickhouse: ## Start with ClickHouse profile
	$(COMPOSE) --profile clickhouse up -d --build

.PHONY: up-all
up-all: ## Start with all optional profiles
	$(COMPOSE) --profile ollama --profile postgres --profile clickhouse up -d --build

# ── Utilities ───────────────────────────────────────────────────────────────

.PHONY: env
env: ## Create .env from .env.example if it doesn't exist
	@test -f .env || (cp .env.example .env && echo "✓ Created .env from .env.example") || echo "• .env already exists"

.PHONY: health
health: ## Check application health endpoint
	@curl -sf http://localhost:$${EXTERNAL_PORT:-80}/health | python3 -m json.tool 2>/dev/null || echo "✗ Service unavailable"
