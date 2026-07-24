.PHONY: help build run dev migrate test clean wire install frontend-dev frontend-build lint

GO       ?= go
BIN      ?= bin/server
BIN_DIR  ?= bin

# Build-time variables injected via -ldflags (see pkg/version)
VERSION  ?= $(shell node -p "require('./package.json').version" 2>/dev/null || echo "dev")
COMMIT   ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BRANCH   ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
DATE     ?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS  := -X "github.com/lyimoexiao/akari/pkg/version.Version=$(VERSION)" \
           -X "github.com/lyimoexiao/akari/pkg/version.Branch=$(BRANCH)" \
           -X "github.com/lyimoexiao/akari/pkg/version.Commit=$(COMMIT)" \
           -X "github.com/lyimoexiao/akari/pkg/version.Date=$(DATE)"

## —— Akari Makefile ——————————————————————————————————————————
help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

install: ## Install frontend dependencies
	pnpm install

frontend-build: install ## Build the frontend (requires install)
	pnpm build

build: frontend-build ## Build the Go server binary (with version info)
	$(GO) build -ldflags '$(LDFLAGS)' -o $(BIN) ./cmd/server/

run: ## Run the server (loads .env if present)
	@if [ -f .env ]; then set -a; . ./.env; set +a; fi; \
		$(GO) run ./cmd/server/

dev: install ## Run backend and frontend concurrently
	@echo "Starting backend and frontend..."
	@$(GO) run ./cmd/server/ & \
		pnpm dev & \
		wait

migrate: ## Run database migrations
	$(GO) run ./cmd/migrate/

test: ## Run all tests with race detection
	$(GO) test ./... -v -race -count=1

lint: frontend-build ## Run all linters
	pnpm lint
	$(GO) vet ./...
	$(GO) fmt ./...

clean: ## Remove build and runtime artifacts
	rm -rf $(BIN_DIR)/ data/ web/dist/ .cache/ *.out *.test *.exe

wire: ## Regenerate Wire dependency injection
	$(GO) run github.com/google/wire/cmd/wire@v0.7.0 ./cmd/server/

version: ## Show build version info
	@echo "Version: $(VERSION)"
	@echo "Branch:  $(BRANCH)"
	@echo "Commit:  $(COMMIT)"
	@echo "Date:    $(DATE)"

frontend-dev: ## Start Vite dev server
	pnpm dev