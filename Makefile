COMPOSE ?= docker compose
BROWSER_OPEN ?= open
ISSUE_TRACKER_PORT ?=
OPENAPI_PORT ?=
WEB_PORT ?=
ISSUE_TRACKER_URL ?=

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev-check
dev-check: ## Check that Docker CLI and Compose are installed.
	@command -v docker >/dev/null 2>&1 || { \
		echo "Docker CLI is required. Install and start Docker Desktop."; \
		exit 1; \
	}
	@docker compose version >/dev/null 2>&1 || { \
		echo "Docker Compose is required. Install a Docker version with the compose plugin."; \
		exit 1; \
	}

.PHONY: dev-up-forward
dev-up-forward: dev-check ## Start issue-tracker, orchestrator, OpenAPI UI, and web UI in the foreground.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build issue-tracker orchestrator openapi web

.PHONY: dev-up
dev-up: dev-check ## Start issue-tracker, orchestrator, OpenAPI UI, and web UI in the background.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d issue-tracker orchestrator openapi
	@issue_port="$$($(COMPOSE) port issue-tracker 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -z "$$issue_port" ]; then \
		echo "issue-tracker port is not assigned"; \
		exit 1; \
	fi; \
	issue_url="http://localhost:$$issue_port"; \
	printf "web NEXT_PUBLIC_ISSUE_TRACKER_URL=%s\n" "$$issue_url"; \
	NEXT_PUBLIC_ISSUE_TRACKER_URL="$$issue_url" ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d web
	$(MAKE) dev-ports
	$(MAKE) dev-open

.PHONY: dev-up-d
dev-up-d: dev-up

.PHONY: dev-down
dev-down: dev-check ## Stop Compose services.
	$(COMPOSE) down

.PHONY: dev-restart
dev-restart: dev-check ## Restart all Compose development services.
	$(COMPOSE) down
	$(MAKE) dev-up

.PHONY: dev-ps
dev-ps: dev-check ## Show Compose service status.
	$(COMPOSE) ps

.PHONY: dev-ports
dev-ports: dev-check ## Show assigned host ports for Compose services.
	@issue_port="$$($(COMPOSE) port issue-tracker 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$issue_port" ]; then \
		printf "issue-tracker: http://localhost:%s\n" "$$issue_port"; \
	else \
		printf "issue-tracker: not running\n"; \
	fi
	@openapi_port="$$($(COMPOSE) port openapi 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$openapi_port" ]; then \
		printf "openapi:       http://localhost:%s\n" "$$openapi_port"; \
	else \
		printf "openapi:       not running\n"; \
	fi
	@web_port="$$($(COMPOSE) port web 3000 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$web_port" ]; then \
		printf "web:           http://localhost:%s\n" "$$web_port"; \
	else \
		printf "web:           not running\n"; \
	fi

.PHONY: dev-open
dev-open: dev-check ## Open the web UI and OpenAPI UI in a browser.
	@opener="$(BROWSER_OPEN)"; \
	if ! command -v "$$opener" >/dev/null 2>&1; then \
		echo "browser opener '$$opener' is not available"; \
		exit 0; \
	fi; \
	openapi_port="$$($(COMPOSE) port openapi 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$openapi_port" ]; then \
		openapi_url="http://localhost:$$openapi_port"; \
		printf "opening openapi: %s\n" "$$openapi_url"; \
		"$$opener" "$$openapi_url" >/dev/null 2>&1 || true; \
	fi; \
	web_port="$$($(COMPOSE) port web 3000 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$web_port" ]; then \
		web_url="http://localhost:$$web_port"; \
		printf "opening web:     %s\n" "$$web_url"; \
		"$$opener" "$$web_url" >/dev/null 2>&1 || true; \
	fi

.PHONY: dev-logs
dev-logs: dev-check ## Follow Compose service logs.
	$(COMPOSE) logs -f issue-tracker orchestrator openapi web

.PHONY: dev-shell
dev-shell: dev-check ## Open a shell in a Compose Go tool container.
	$(COMPOSE) run --rm go-tools sh

.PHONY: dev-exec
dev-exec: dev-check ## Run CMD in a Compose Go tool container. Example: make dev-exec CMD="go test ./..."
	@if [ -z "$(CMD)" ]; then \
		echo 'Usage: make dev-exec CMD="go test ./..."'; \
		exit 1; \
	fi
	$(COMPOSE) run --rm go-tools sh -c '$(CMD)'

.PHONY: dev-test
dev-test: dev-check ## Run Go tests and web UI typecheck in Compose containers.
	$(COMPOSE) run --rm go-tools go test ./...
	$(COMPOSE) run --rm web npm run typecheck

.PHONY: dev-build-app
dev-build-app: dev-check ## Run Go tests and the web UI production build in Compose containers.
	$(COMPOSE) run --rm go-tools go test ./...
	$(COMPOSE) run --rm web npm run build

.PHONY: issue-tracker-up
issue-tracker-up: dev-check ## Start the issue-tracker API with Docker Compose in the background.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) $(COMPOSE) up --build -d issue-tracker

.PHONY: orchestrator-up
orchestrator-up: dev-check ## Start the issue-tracker and orchestrator with Docker Compose in the background.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) $(COMPOSE) up --build -d issue-tracker orchestrator

.PHONY: openapi-up
openapi-up: dev-check ## Start the OpenAPI UI with Docker Compose in the background.
	OPENAPI_PORT=$(OPENAPI_PORT) $(COMPOSE) up -d openapi
	$(MAKE) dev-ports

.PHONY: web-up
web-up: dev-up ## Start issue-tracker, orchestrator, OpenAPI UI, and Next.js web UI in the background.

.PHONY: tui-up
tui-up: orchestrator-up ## Run the TUI on the host against the Compose issue-tracker API.
	@api="$(ISSUE_TRACKER_URL)"; \
	if [ -z "$$api" ]; then \
		api="http://localhost:$$($(COMPOSE) port issue-tracker 8080 | sed 's/.*://')"; \
	fi; \
	go run ./cmd/tasq-tui -api "$$api" -watch

.PHONY: tq
tq: issue-tracker-up ## Run tq against the Compose issue-tracker API. Example: make tq ARGS="issue list"
	@api="$(ISSUE_TRACKER_URL)"; \
	if [ -z "$$api" ]; then \
		api="http://localhost:$$($(COMPOSE) port issue-tracker 8080 | sed 's/.*://')"; \
	fi; \
	go run ./cmd/tq --api-url "$$api" $(ARGS)

.PHONY: dev-gui
dev-gui: web-up ## Alias-style command for web-up.
