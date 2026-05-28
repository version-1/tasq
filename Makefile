DEVCONTAINER ?= devcontainer
DEVCONTAINER_WORKSPACE ?= $(CURDIR)
DEVCONTAINER_EXEC = $(DEVCONTAINER) exec --workspace-folder "$(DEVCONTAINER_WORKSPACE)"
ISSUE_TRACKER_CMD = go run ./cmd/issue-tracker -addr :8080 -db tasq-issues.sqlite
ISSUE_TRACKER_LOG ?= /tmp/tasq-issue-tracker.log
ISSUE_TRACKER_PID ?= /tmp/tasq-issue-tracker.pid
ORCHESTRATOR_CMD = go run ./cmd/orchestrator -db tasq-orchestrator.sqlite -issue-tracker http://localhost:8080
ORCHESTRATOR_LOG ?= /tmp/tasq-orchestrator.log
ORCHESTRATOR_PID ?= /tmp/tasq-orchestrator.pid
WEB_CMD = cd web && ./node_modules/.bin/next dev --hostname 0.0.0.0
WEB_LOG ?= /tmp/tasq-web.log
WEB_PID ?= /tmp/tasq-web.pid

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev-check
dev-check: ## Check that Docker CLI and Dev Container CLI are installed.
	@command -v docker >/dev/null 2>&1 || { \
		echo "Docker CLI is required. Install and start Docker Desktop."; \
		exit 1; \
	}
	@command -v $(DEVCONTAINER) >/dev/null 2>&1 || { \
		echo "devcontainer CLI is required. Install @devcontainers/cli or use VS Code Dev Containers."; \
		exit 1; \
	}

.PHONY: dev-build
dev-build: dev-check ## Build the Dev Container image.
	$(DEVCONTAINER) build --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: dev-up
dev-up: dev-check ## Create or start the Dev Container.
	$(DEVCONTAINER) up --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: dev-rebuild
dev-rebuild: dev-check ## Recreate the Dev Container from scratch.
	$(DEVCONTAINER) up --workspace-folder "$(DEVCONTAINER_WORKSPACE)" --remove-existing-container

.PHONY: dev-status
dev-status: dev-check ## Show Dev Container status and tool versions.
	@printf "Dev Container\n"
	@printf "  %-12s %s\n" "Workspace:" "$(DEVCONTAINER_WORKSPACE)"
	@printf "\nDocker container\n"
	@docker ps -a \
		--filter "label=devcontainer.local_folder=$(DEVCONTAINER_WORKSPACE)" \
		--format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Ports}}\t{{.Image}}"
	@printf "\nRuntime\n"
	@if $(DEVCONTAINER_EXEC) /bin/bash -lc ' \
		printf "  %-12s %s\n" "Status:" "running"; \
		printf "  %-12s %s\n" "User:" "$$(whoami)"; \
		printf "  %-12s %s\n" "Directory:" "$$(pwd)"; \
		printf "  %-12s %s\n" "Go:" "$$(go version)"; \
		printf "  %-12s %s\n" "Node:" "$$(node --version)"; \
		printf "  %-12s %s\n" "npm:" "$$(npm --version)"; \
		printf "\nProcesses\n"; \
		ps -eo pid,stat,command | grep -E "go run ./cmd/issue-tracker|go run ./cmd/orchestrator|next dev|cmd/tasq-tui" | grep -v grep || true; \
	' 2>/dev/null; then \
		:; \
	else \
		printf "  %-12s %s\n" "Status:" "not running or not created"; \
		printf "  %-12s %s\n" "Next step:" "make dev-up"; \
		exit 1; \
	fi

.PHONY: dev-shell
dev-shell: dev-check ## Open a shell inside the running Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash

.PHONY: dev-exec
dev-exec: dev-check ## Run CMD inside the Dev Container. Example: make dev-exec CMD="go test ./..."
	@if [ -z "$(CMD)" ]; then \
		echo 'Usage: make dev-exec CMD="go test ./..."'; \
		exit 1; \
	fi
	$(DEVCONTAINER_EXEC) /bin/bash -lc '$(CMD)'

.PHONY: dev-test
dev-test: dev-check ## Run Go tests and web UI typecheck inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run typecheck'

.PHONY: dev-build-app
dev-build-app: dev-check ## Run production build checks inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run build'

.PHONY: issue-tracker-up
issue-tracker-up: dev-up ## Start the issue-tracker in the Dev Container background.
	$(DEVCONTAINER_EXEC) /bin/bash -lc ' \
		if [ -s "$(ISSUE_TRACKER_PID)" ] && kill -0 "$$(cat "$(ISSUE_TRACKER_PID)")" 2>/dev/null; then \
			echo "issue-tracker already running"; \
		else \
			nohup $(ISSUE_TRACKER_CMD) > "$(ISSUE_TRACKER_LOG)" 2>&1 & echo $$! > "$(ISSUE_TRACKER_PID)"; \
			echo "issue-tracker started"; \
			echo "pid: $$(cat "$(ISSUE_TRACKER_PID)")"; \
			echo "log: $(ISSUE_TRACKER_LOG)"; \
		fi \
	'

.PHONY: orchestrator-up
orchestrator-up: issue-tracker-up ## Start the orchestrator in the Dev Container background.
	$(DEVCONTAINER_EXEC) /bin/bash -lc ' \
		if [ -s "$(ORCHESTRATOR_PID)" ] && kill -0 "$$(cat "$(ORCHESTRATOR_PID)")" 2>/dev/null; then \
			echo "orchestrator already running"; \
		else \
			nohup $(ORCHESTRATOR_CMD) > "$(ORCHESTRATOR_LOG)" 2>&1 & echo $$! > "$(ORCHESTRATOR_PID)"; \
			echo "orchestrator started"; \
			echo "pid: $$(cat "$(ORCHESTRATOR_PID)")"; \
			echo "log: $(ORCHESTRATOR_LOG)"; \
		fi \
	'

.PHONY: dev-orchestrator
dev-orchestrator: issue-tracker-up ## Run the orchestrator in the Dev Container foreground.
	$(DEVCONTAINER_EXEC) /bin/bash -lc '$(ORCHESTRATOR_CMD)'

.PHONY: web-up
web-up: orchestrator-up ## Start orchestrator and the Next.js web UI in the Dev Container background.
	$(DEVCONTAINER_EXEC) /bin/bash -lc ' \
		if [ -s "$(WEB_PID)" ] && kill -0 "$$(cat "$(WEB_PID)")" 2>/dev/null; then \
			echo "web already running"; \
		elif pgrep -f "[n]ode /workspaces/tasq/web/node_modules/.bin/next dev" >/dev/null; then \
			pgrep -f "[n]ode /workspaces/tasq/web/node_modules/.bin/next dev" | head -n 1 > "$(WEB_PID)"; \
			echo "web already running"; \
		else \
			nohup bash -lc "$(WEB_CMD)" > "$(WEB_LOG)" 2>&1 & echo $$! > "$(WEB_PID)"; \
			sleep 1; \
			pgrep -f "[n]ode /workspaces/tasq/web/node_modules/.bin/next dev" | head -n 1 > "$(WEB_PID)" || true; \
			echo "web started"; \
			echo "pid: $$(cat "$(WEB_PID)")"; \
			echo "log: $(WEB_LOG)"; \
		fi \
	'

.PHONY: tui-up
tui-up: orchestrator-up ## Run issue-tracker and orchestrator in the background, then the TUI in the foreground.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go run ./cmd/tasq-tui -api http://localhost:8080 -watch'

.PHONY: dev-gui
dev-gui: web-up ## Run orchestrator and the Next.js GUI inside the Dev Container. Prefer web-up.
