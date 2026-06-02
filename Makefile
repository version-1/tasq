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
help: ## Show target prefixes and available targets.
	@printf "Prefixes:\n"
	@printf "  dev-*  Host-facing local development entry points.\n"
	@printf "  dc-*   Docker Compose and dev-container operations.\n"
	@printf "  run-*  Processes and commands that run inside the dev container.\n"
	@printf "\nGeneral:\n"
	@awk 'BEGIN {FS = ":.*## "}; /^help:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\ndev-* targets:\n"
	@awk 'BEGIN {FS = ":.*## "}; /^dev-[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\ndc-* targets:\n"
	@awk 'BEGIN {FS = ":.*## "}; /^dc-[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)
	@printf "\nrun-* targets:\n"
	@awk 'BEGIN {FS = ":.*## "}; /^run-[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev-check
dev-check:
	@command -v docker >/dev/null 2>&1 || { \
		echo "Docker CLI is required. Install and start Docker Desktop."; \
		exit 1; \
	}
	@docker compose version >/dev/null 2>&1 || { \
		echo "Docker Compose is required. Install a Docker version with the compose plugin."; \
		exit 1; \
	}

.PHONY: dev-up
dev-up: dev-check ## Start the dev environment and print assigned URLs.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev openapi
	$(MAKE) dc-ready
	$(MAKE) run-all
	$(MAKE) dev-ports

.PHONY: dev-down
dev-down: dev-check ## Stop Compose services.
	$(COMPOSE) down

.PHONY: dev-restart
dev-restart: dev-check ## Restart the dev environment.
	$(COMPOSE) down
	$(MAKE) dev-up

.PHONY: dev-reset-db
dev-reset-db: dev-check ## Reset local SQLite data. Usage: make dev-reset-db CONFIRM=1
	@if [ "$(CONFIRM)" != "1" ]; then \
		echo 'This removes local SQLite data under .tasq/system/data/.'; \
		echo 'Run: make dev-reset-db CONFIRM=1'; \
		exit 1; \
	fi
	$(COMPOSE) down
	rm -f .tasq/system/data/issues.sqlite .tasq/system/data/orchestrator.sqlite
	$(MAKE) dev-up

.PHONY: dev-openapi
dev-openapi: dev-check ## Start only the OpenAPI UI Compose service and print assigned URLs.
	OPENAPI_PORT=$(OPENAPI_PORT) $(COMPOSE) up -d openapi
	$(MAKE) dev-ports

.PHONY: dev-ports
dev-ports: dev-check ## Show assigned local URLs for dev services.
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
dev-open: dev-check ## Open the Web UI and OpenAPI UI in a browser.
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

.PHONY: dev-test
dev-test: dev-check ## Run Go tests and Web UI typecheck in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dc-ready
	$(DEV_EXEC) sh -c 'go test ./... && cd web && npm install && npm run typecheck'

.PHONY: dev-build
dev-build: dev-check ## Run Go tests and the Web UI production build in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dc-ready
	$(DEV_EXEC) sh -c 'go test ./... && cd web && npm install && npm run build'

.PHONY: dev-codex-login
dev-codex-login: dev-check ## Log in to Codex inside the dev container with device auth.
	$(DEV_EXEC) sh -c 'codex login --device-auth'

.PHONY: dev-codex-status
dev-codex-status: dev-check ## Check Codex authentication status inside the dev container.
	$(DEV_EXEC) sh -c 'codex login status'

.PHONY: dev-gh-login
dev-gh-login: dev-check ## Log in to GitHub CLI inside the dev container.
	$(DEV_EXEC) sh -c 'gh auth login'

.PHONY: dev-gh-status
dev-gh-status: dev-check ## Check GitHub CLI authentication inside the dev container.
	$(DEV_EXEC) sh -c 'gh auth status'

.PHONY: dc-ready
dc-ready: dev-check ## Wait until the dev container tools and volumes are ready.
	@attempt=1; \
	while [ "$$attempt" -le 30 ]; do \
		if $(DEV_EXEC_NO_TTY) sh -c 'test -x /usr/local/go/bin/go && test -w /go/pkg/mod && test -w /go/pkg/sumdb && test -w /home/codex/.cache/go-build && test -w /home/codex/.codex && test -w /home/codex/.config/gh && test -w /workspace/web/node_modules' >/dev/null 2>&1; then \
			exit 0; \
		fi; \
		attempt=$$((attempt + 1)); \
		sleep 1; \
	done; \
	echo "dev container is not ready for the codex user"; \
	exit 1

.PHONY: dc-ps
dc-ps: dev-check ## Show Docker Compose service status.
	$(COMPOSE) ps

.PHONY: dc-shell
dc-shell: dev-check ## Open a shell in the running dev container.
	$(DEV_EXEC) sh

.PHONY: dc-exec
dc-exec: dev-check ## Run CMD in the running dev container. Example: make dc-exec CMD="go test ./..."
	@if [ -z "$(CMD)" ]; then \
		echo 'Usage: make dc-exec CMD="go test ./..."'; \
		exit 1; \
	fi
	$(DEV_EXEC) sh -c '$(CMD)'

.PHONY: run-all
run-all: dev-check ## Start issue-tracker, orchestrator, and Web inside the running dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp'
	$(MAKE) run-stop
	$(DEV_EXEC_DETACHED) sh -c 'go mod download >>$(DEV_LOG_DIR)/go-mod-download.log 2>&1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.issue-tracker.toml >>$(DEV_LOG_DIR)/issue-tracker.log 2>&1'
	$(MAKE) run-ready-issue-tracker
	$(DEV_EXEC_DETACHED) sh -c 'sleep 1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.orchestrator.toml >>$(DEV_LOG_DIR)/orchestrator.log 2>&1'
	@web_issue_url="$(WEB_ISSUE_TRACKER_URL)"; \
	if [ -z "$$web_issue_url" ]; then \
		issue_port="$$($(COMPOSE) port dev 8080 | sed 's/.*://')"; \
		web_issue_url="http://127.0.0.1:$$issue_port"; \
	fi; \
	$(DEV_EXEC_DETACHED) sh -c 'cd web && npm install >>../$(DEV_LOG_DIR)/web-install.log 2>&1 && exec env NEXT_PUBLIC_ISSUE_TRACKER_URL="'"$$web_issue_url"'" npm run dev -- --hostname 0.0.0.0 --port 3000 >>../$(DEV_LOG_DIR)/web.log 2>&1'

.PHONY: run-ready-issue-tracker
run-ready-issue-tracker: dev-check
	@$(DEV_EXEC_NO_TTY) sh -c 'attempt=1; \
		while [ "$$attempt" -le 30 ]; do \
			if curl -fsS http://127.0.0.1:8080/api/v1/health >/dev/null 2>&1; then \
				exit 0; \
			fi; \
			attempt=$$((attempt + 1)); \
			sleep 1; \
		done; \
		echo "issue-tracker is not ready"; \
		exit 1'

.PHONY: run-ensure-issue-tracker
run-ensure-issue-tracker: dev-check
	@if $(DEV_EXEC_NO_TTY) sh -c 'curl -fsS http://127.0.0.1:8080/api/v1/health >/dev/null 2>&1'; then \
		:; \
	else \
		$(MAKE) run-issue-tracker; \
	fi

.PHONY: run-stop
run-stop: dev-check ## Stop dev processes inside the dev container without stopping the container.
	-$(DEV_EXEC) sh -c 'pkill -f "air.*\\.air[.]issue-tracker[.]toml" 2>/dev/null || true; pkill -f "air.*\\.air[.]orchestrator[.]toml" 2>/dev/null || true; pkill -f "next dev.*--hostname 0[.]0[.]0[.]0.*--port 3000" 2>/dev/null || true'

.PHONY: run-issue-tracker run-is
run-issue-tracker: dev-check ## Start the issue-tracker process inside the running dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp; pkill -f "air.*\\.air[.]issue-tracker[.]toml" 2>/dev/null || true'
	$(DEV_EXEC_DETACHED) sh -c 'go mod download >>$(DEV_LOG_DIR)/go-mod-download.log 2>&1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.issue-tracker.toml >>$(DEV_LOG_DIR)/issue-tracker.log 2>&1'
	$(MAKE) run-ready-issue-tracker
run-is: run-issue-tracker ## Alias for run-issue-tracker.

.PHONY: run-orchestrator run-or
run-orchestrator: run-ensure-issue-tracker ## Start the orchestrator process inside the running dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR) /workspace/.tmp; pkill -f "air.*\\.air[.]orchestrator[.]toml" 2>/dev/null || true'
	$(DEV_EXEC_DETACHED) sh -c 'sleep 1; exec go run github.com/air-verse/air@$(AIR_VERSION) -c .air.orchestrator.toml >>$(DEV_LOG_DIR)/orchestrator.log 2>&1'
run-or: run-orchestrator ## Alias for run-orchestrator.

.PHONY: run-web run-w
run-web: run-ensure-issue-tracker ## Start the Next.js Web process inside the running dev container.
	$(DEV_EXEC) sh -c 'mkdir -p $(DEV_LOG_DIR); pkill -f "next dev.*--hostname 0[.]0[.]0[.]0.*--port 3000" 2>/dev/null || true'
	@web_issue_url="$(WEB_ISSUE_TRACKER_URL)"; \
	if [ -z "$$web_issue_url" ]; then \
		issue_port="$$($(COMPOSE) port dev 8080 | sed 's/.*://')"; \
		web_issue_url="http://127.0.0.1:$$issue_port"; \
	fi; \
	$(DEV_EXEC_DETACHED) sh -c 'cd web && npm install >>../$(DEV_LOG_DIR)/web-install.log 2>&1 && exec env NEXT_PUBLIC_ISSUE_TRACKER_URL="'"$$web_issue_url"'" npm run dev -- --hostname 0.0.0.0 --port 3000 >>../$(DEV_LOG_DIR)/web.log 2>&1'
run-w: run-web ## Alias for run-web.

.PHONY: run-tui
run-tui: run-ensure-issue-tracker ## Run the TUI interactively inside the running dev container.
	$(DEV_EXEC) sh -c 'go run ./cmd/tasq-tui -watch'

.PHONY: run-tq
run-tq: dev-check ## Run tq inside the running dev container. Example: make run-tq ARGS="issue list"
	$(DEV_EXEC) sh -c 'tq $(ARGS)'

.PHONY: run-ps
run-ps: dev-check ## Show dev processes running inside the dev container.
	-$(DEV_EXEC) sh -c 'ps -ef | grep -E "air|issue-tracker|orchestrator|next" | grep -v grep || true'

.PHONY: run-logs
run-logs: dev-check ## Follow dev process logs.
	@mkdir -p $(DEV_LOG_DIR)
	@touch $(DEV_LOG_DIR)/issue-tracker.log $(DEV_LOG_DIR)/orchestrator.log $(DEV_LOG_DIR)/web.log
	tail -f $(DEV_LOG_DIR)/*.log
