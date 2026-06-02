COMPOSE ?= docker compose
BROWSER_OPEN ?= open
TQ_HOME ?= ./.tasq
ISSUE_TRACKER_PORT ?=
ORCHESTRATOR_PORT ?=
OPENAPI_PORT ?=
WEB_PORT ?=
WEB_ISSUE_TRACKER_URL ?=

export TQ_HOME

DEV_EXEC = $(COMPOSE) exec --user codex dev
DEV_EXEC_DETACHED = $(COMPOSE) exec -d --user codex dev
DEV_EXEC_NO_TTY = $(COMPOSE) exec -T --user codex dev
DEV_LOG_DIR = .tmp/dev-logs
AIR_VERSION ?= v1.52.3

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
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev openapi
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'set -e; \
		mkdir -p $(DEV_LOG_DIR) /workspace/.tmp; \
		pkill -f "air.*\\.air.issue-tracker.toml" 2>/dev/null || true; \
		pkill -f "air.*\\.air.orchestrator.toml" 2>/dev/null || true; \
		pkill -f "next dev.*--hostname 0.0.0.0.*--port 3000" 2>/dev/null || true; \
		trap "pkill -f \"air.*\\.air.issue-tracker.toml\" 2>/dev/null || true; pkill -f \"air.*\\.air.orchestrator.toml\" 2>/dev/null || true; pkill -f \"next dev.*--hostname 0.0.0.0.*--port 3000\" 2>/dev/null || true" INT TERM EXIT; \
		go mod download; \
		(go run github.com/air-verse/air@$(AIR_VERSION) -c .air.issue-tracker.toml) & \
		sleep 1; \
		(go run github.com/air-verse/air@$(AIR_VERSION) -c .air.orchestrator.toml) & \
		(cd web && npm install && NEXT_PUBLIC_ISSUE_TRACKER_URL="$(WEB_ISSUE_TRACKER_URL)" npm run dev -- --hostname 0.0.0.0 --port 3000) & \
		wait'

.PHONY: dev-up
dev-up: dev-check ## Start dev container, OpenAPI UI, issue-tracker, orchestrator, and web UI in the background.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev openapi
	$(MAKE) dev-ready
	$(MAKE) dev-start-processes
	$(MAKE) dev-ports
	$(MAKE) dev-open

.PHONY: dev-up-d
dev-up-d: dev-up

.PHONY: dev-start-processes
dev-start-processes: dev-check ## Start issue-tracker, orchestrator, and web inside the dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp'
	$(MAKE) dev-stop-processes
	$(DEV_EXEC_DETACHED) sh -c 'go mod download >>$(DEV_LOG_DIR)/go-mod-download.log 2>&1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.issue-tracker.toml >>$(DEV_LOG_DIR)/issue-tracker.log 2>&1'
	$(DEV_EXEC_DETACHED) sh -c 'sleep 1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.orchestrator.toml >>$(DEV_LOG_DIR)/orchestrator.log 2>&1'
	@web_issue_url="$(WEB_ISSUE_TRACKER_URL)"; \
	if [ -z "$$web_issue_url" ]; then \
		issue_port="$$($(COMPOSE) port dev 8080 | sed 's/.*://')"; \
		web_issue_url="http://127.0.0.1:$$issue_port"; \
	fi; \
	$(DEV_EXEC_DETACHED) sh -c 'cd web && npm install >>../$(DEV_LOG_DIR)/web-install.log 2>&1 && exec env NEXT_PUBLIC_ISSUE_TRACKER_URL="'"$$web_issue_url"'" npm run dev -- --hostname 0.0.0.0 --port 3000 >>../$(DEV_LOG_DIR)/web.log 2>&1'

.PHONY: dev-ready
dev-ready: dev-check ## Wait until the dev container volumes are writable by the codex user.
	@attempt=1; \
	while [ "$$attempt" -le 30 ]; do \
		if $(DEV_EXEC_NO_TTY) sh -c 'test -x /usr/local/go/bin/go && test -w /go/pkg/mod && test -w /home/codex/.cache/go-build && test -w /home/codex/.codex && test -w /workspace/web/node_modules' >/dev/null 2>&1; then \
			exit 0; \
		fi; \
		attempt=$$((attempt + 1)); \
		sleep 1; \
	done; \
	echo "dev container is not ready for the codex user"; \
	exit 1

.PHONY: dev-stop-processes
dev-stop-processes: dev-check ## Stop dev processes inside the dev container without stopping the container.
	-$(DEV_EXEC) sh -c 'pkill -f "air.*\\.air.issue-tracker.toml" 2>/dev/null || true; pkill -f "air.*\\.air.orchestrator.toml" 2>/dev/null || true; pkill -f "next dev.*--hostname 0.0.0.0.*--port 3000" 2>/dev/null || true'

.PHONY: dev-down
dev-down: dev-check ## Stop Compose services and all dev processes.
	$(COMPOSE) down

.PHONY: dev-restart
dev-restart: dev-check ## Restart all development services.
	$(COMPOSE) down
	$(MAKE) dev-up

.PHONY: dev-rebuild-schema
dev-rebuild-schema: dev-check ## Recreate local SQLite schemas. Usage: make dev-rebuild-schema CONFIRM=1
	@if [ "$(CONFIRM)" != "1" ]; then \
		echo 'This removes local SQLite data under .tasq/system/data/.'; \
		echo 'Run: make dev-rebuild-schema CONFIRM=1'; \
		exit 1; \
	fi
	$(COMPOSE) down
	rm -f .tasq/system/data/issues.sqlite .tasq/system/data/orchestrator.sqlite
	$(MAKE) dev-up

.PHONY: dev-ps
dev-ps: dev-check ## Show Compose service status.
	$(COMPOSE) ps

.PHONY: dev-ports
dev-ports: dev-check ## Show assigned host ports for Compose services.
	@issue_port="$$($(COMPOSE) port dev 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$issue_port" ]; then \
		printf "issue-tracker: http://localhost:%s\n" "$$issue_port"; \
	else \
		printf "issue-tracker: not running\n"; \
	fi
	@orchestrator_port="$$($(COMPOSE) port dev 8081 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$orchestrator_port" ]; then \
		printf "orchestrator:  http://localhost:%s\n" "$$orchestrator_port"; \
	else \
		printf "orchestrator:  not running\n"; \
	fi
	@openapi_port="$$($(COMPOSE) port openapi 8080 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$openapi_port" ]; then \
		printf "openapi:       http://localhost:%s\n" "$$openapi_port"; \
	else \
		printf "openapi:       not running\n"; \
	fi
	@web_port="$$($(COMPOSE) port dev 3000 2>/dev/null | sed 's/.*://')"; \
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
	web_port="$$($(COMPOSE) port dev 3000 2>/dev/null | sed 's/.*://')"; \
	if [ -n "$$web_port" ]; then \
		web_url="http://localhost:$$web_port"; \
		printf "opening web:     %s\n" "$$web_url"; \
		"$$opener" "$$web_url" >/dev/null 2>&1 || true; \
	fi

.PHONY: dev-logs
dev-logs: dev-check ## Follow dev process logs.
	@mkdir -p $(DEV_LOG_DIR)
	@touch $(DEV_LOG_DIR)/issue-tracker.log $(DEV_LOG_DIR)/orchestrator.log $(DEV_LOG_DIR)/web.log
	tail -f $(DEV_LOG_DIR)/*.log

.PHONY: dev-shell
dev-shell: dev-check ## Open a shell in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh

.PHONY: dev-exec
dev-exec: dev-check ## Run CMD in the dev container. Example: make dev-exec CMD="go test ./..."
	@if [ -z "$(CMD)" ]; then \
		echo 'Usage: make dev-exec CMD="go test ./..."'; \
		exit 1; \
	fi
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c '$(CMD)'

.PHONY: dev-test
dev-test: dev-check ## Run Go tests and web UI typecheck in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'go test ./... && cd web && npm install && npm run typecheck'

.PHONY: dev-build-app
dev-build-app: dev-check ## Run Go tests and the web UI production build in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'go test ./... && cd web && npm install && npm run build'

.PHONY: issue-tracker-up
issue-tracker-up: dev-check ## Start the issue-tracker process inside the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp; pkill -f "air.*\\.air.issue-tracker.toml" 2>/dev/null || true'
	$(DEV_EXEC_DETACHED) sh -c 'go mod download >>$(DEV_LOG_DIR)/go-mod-download.log 2>&1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.issue-tracker.toml >>$(DEV_LOG_DIR)/issue-tracker.log 2>&1'

.PHONY: orchestrator-up
orchestrator-up: issue-tracker-up ## Start the orchestrator process inside the dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp; pkill -f "air.*\\.air.orchestrator.toml" 2>/dev/null || true'
	$(DEV_EXEC_DETACHED) sh -c 'sleep 1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.orchestrator.toml >>$(DEV_LOG_DIR)/orchestrator.log 2>&1'

.PHONY: openapi-up
openapi-up: dev-check ## Start the OpenAPI UI with Docker Compose in the background.
	OPENAPI_PORT=$(OPENAPI_PORT) $(COMPOSE) up -d openapi
	$(MAKE) dev-ports

.PHONY: web-up
web-up: issue-tracker-up ## Start the Next.js web UI inside the dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR); pkill -f "next dev.*--hostname 0.0.0.0.*--port 3000" 2>/dev/null || true'
	@web_issue_url="$(WEB_ISSUE_TRACKER_URL)"; \
	if [ -z "$$web_issue_url" ]; then \
		issue_port="$$($(COMPOSE) port dev 8080 | sed 's/.*://')"; \
		web_issue_url="http://127.0.0.1:$$issue_port"; \
	fi; \
	$(DEV_EXEC_DETACHED) sh -c 'cd web && npm install >>../$(DEV_LOG_DIR)/web-install.log 2>&1 && exec env NEXT_PUBLIC_ISSUE_TRACKER_URL="'"$$web_issue_url"'" npm run dev -- --hostname 0.0.0.0 --port 3000 >>../$(DEV_LOG_DIR)/web.log 2>&1'

.PHONY: tui-up
tui-up: issue-tracker-up ## Run the TUI interactively inside the dev container.
	$(DEV_EXEC) sh -c 'go run ./cmd/tasq-tui -watch'

.PHONY: tq
tq: issue-tracker-up ## Run tq inside the dev container. Example: make tq ARGS="issue list"
	$(DEV_EXEC) sh -c 'go run ./cmd/tq $(ARGS)'

.PHONY: codex-login
codex-login: dev-check ## Log in to Codex inside the dev container with device auth and persist credentials in codex-home.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'codex login --device-auth'

.PHONY: codex-check
codex-check: dev-check ## Check Codex CLI availability inside the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dev-ready
	$(DEV_EXEC) sh -c 'codex --help >/dev/null && codex app-server --help >/dev/null'

.PHONY: dev-gui
dev-gui: web-up ## Alias-style command for web-up.
