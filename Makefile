COMPOSE ?= docker compose
BROWSER_OPEN ?= open
TQ_HOME ?= ./.tasq
ISSUE_TRACKER_PORT ?=
ORCHESTRATOR_PORT ?=
OPENAPI_PORT ?=
WEB_PORT ?=
WEB_ISSUE_TRACKER_URL ?=
RELEASE_BRANCH ?= main
RELEASE_REMOTE ?= origin
RELEASE_REPO ?= version-1/tasq
TQ_INSTALL_DIR ?= $(HOME)/.local/bin
TQ_INSTALL_NAME ?= tq

export TQ_HOME

DEV_EXEC = $(COMPOSE) exec --user codex dev
DEV_EXEC_NO_TTY = $(COMPOSE) exec -T --user codex dev
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
	@printf "\nRelease:\n"
	@awk 'BEGIN {FS = ":.*## "}; /^(prerelease|release|install-tq|install-tq-prerelease):.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: prerelease
prerelease: ## Create and push a prerelease tag from the latest formal release.
	@RELEASE_REMOTE="$(RELEASE_REMOTE)" sh scripts/release.sh prerelease

.PHONY: release
release: ## Create and push a formal release tag. Usage: make release version=v0.1.1
	@RELEASE_BRANCH="$(RELEASE_BRANCH)" RELEASE_REMOTE="$(RELEASE_REMOTE)" sh scripts/release.sh release "$(version)"

.PHONY: install-tq
install-tq: ## Install tq from the latest formal release, or a specific tag. Usage: make install-tq version=v0.1.0
	@TQ_RELEASE_REPO="$(RELEASE_REPO)" TQ_INSTALL_DIR="$(TQ_INSTALL_DIR)" TQ_INSTALL_NAME="$(TQ_INSTALL_NAME)" sh scripts/install-tq-release.sh release "$(version)"

.PHONY: install-tq-prerelease
install-tq-prerelease: ## Install tq from the latest prerelease, or a specific tag. Usage: make install-tq-prerelease version=v0.1.0-pre.1
	@TQ_RELEASE_REPO="$(RELEASE_REPO)" TQ_INSTALL_DIR="$(TQ_INSTALL_DIR)" TQ_INSTALL_NAME="$(TQ_INSTALL_NAME)" sh scripts/install-tq-release.sh prerelease "$(version)"

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
	@COMPOSE="$(COMPOSE)" scripts/dev-ports.sh

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
	$(DEV_EXEC) sh -c 'go test ./... && cd cmd/web/frontend && npm install && npm run typecheck'

.PHONY: dev-build
dev-build: dev-check ## Run Go tests and the Web UI production build in the dev container.
	ISSUE_TRACKER_PORT=$(ISSUE_TRACKER_PORT) ORCHESTRATOR_PORT=$(ORCHESTRATOR_PORT) OPENAPI_PORT=$(OPENAPI_PORT) WEB_PORT=$(WEB_PORT) $(COMPOSE) up --build -d dev
	$(MAKE) dc-ready
	$(DEV_EXEC) sh -c 'go test ./... && cd cmd/web/frontend && npm install && npm run build'

.PHONY: dev-codex-login
dev-codex-login: dev-check ## Log in to Codex inside the dev container with device auth.
	$(DEV_EXEC) sh -c 'codex login --device-auth'

.PHONY: dev-codex-status
dev-codex-status: dev-check ## Check Codex authentication status inside the dev container.
	$(DEV_EXEC) sh -c 'codex login status'

.PHONY: dev-gh-login
dev-gh-login: dev-check ## Log in to GitHub CLI inside the dev container.
	$(DEV_EXEC) sh -c 'gh auth login && gh auth setup-git'

.PHONY: dev-gh-status
dev-gh-status: dev-check ## Check GitHub CLI authentication inside the dev container.
	$(DEV_EXEC) sh -c 'gh auth status'

.PHONY: dc-ready
dc-ready: dev-check ## Wait until the dev container tools and volumes are ready.
	@attempt=1; \
	while [ "$$attempt" -le 30 ]; do \
		if $(DEV_EXEC_NO_TTY) sh -c 'test -x /usr/local/go/bin/go && test -w /go/pkg/mod && test -w /go/pkg/sumdb && test -w /home/codex/.cache/go-build && test -w /home/codex/.codex && test -w /home/codex/.config/gh && test -w /workspace/cmd/web/frontend/node_modules' >/dev/null 2>&1; then \
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
	$(MAKE) run-stop
	$(DEV_EXEC) env AIR_VERSION="$(AIR_VERSION)" scripts/dev-run.sh issue-tracker
	$(MAKE) run-ready-issue-tracker
	$(DEV_EXEC) env AIR_VERSION="$(AIR_VERSION)" scripts/dev-run.sh orchestrator
	$(MAKE) run-web

.PHONY: run-ready-issue-tracker
run-ready-issue-tracker: dev-check
	@$(DEV_EXEC_NO_TTY) scripts/dev-ready.sh issue-tracker

.PHONY: run-ensure-issue-tracker
run-ensure-issue-tracker: dev-check
	@if $(DEV_EXEC_NO_TTY) scripts/dev-ready.sh issue-tracker --check >/dev/null 2>&1; then \
		:; \
	else \
		$(MAKE) run-issue-tracker; \
	fi

.PHONY: run-stop
run-stop: dev-check ## Stop dev processes inside the dev container without stopping the container.
	-$(DEV_EXEC) scripts/dev-stop.sh

.PHONY: run-issue-tracker run-is
run-issue-tracker: dev-check ## Start the issue-tracker process inside the running dev container.
	$(DEV_EXEC) env AIR_VERSION="$(AIR_VERSION)" scripts/dev-run.sh issue-tracker
	$(MAKE) run-ready-issue-tracker
run-is: run-issue-tracker ## Alias for run-issue-tracker.

.PHONY: run-orchestrator run-or
run-orchestrator: run-ensure-issue-tracker ## Start the orchestrator process inside the running dev container.
	$(DEV_EXEC) env AIR_VERSION="$(AIR_VERSION)" scripts/dev-run.sh orchestrator
run-or: run-orchestrator ## Alias for run-orchestrator.

.PHONY: run-web run-w
run-web: run-ensure-issue-tracker ## Start the Go Web process inside the running dev container.
	$(DEV_EXEC) env AIR_VERSION="$(AIR_VERSION)" scripts/dev-run.sh web
run-w: run-web ## Alias for run-web.

.PHONY: run-tui
run-tui: run-ensure-issue-tracker ## Run the TUI interactively inside the running dev container.
	$(DEV_EXEC) sh -c 'go run ./cmd/tasq-tui -watch'

.PHONY: run-tq
run-tq: dev-check ## Run tq inside the running dev container. Example: make run-tq ARGS="issue list"
	$(DEV_EXEC) sh -c 'tq $(ARGS)'

.PHONY: run-ps
run-ps: dev-check ## Show dev processes running inside the dev container.
	-$(DEV_EXEC) scripts/dev-ps.sh

.PHONY: run-logs
run-logs: dev-check ## Follow dev process logs.
	@TQ_HOME="$(TQ_HOME)" scripts/dev-logs.sh $(SERVICE)
