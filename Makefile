DEVCONTAINER ?= devcontainer
DEVCONTAINER_WORKSPACE ?= $(CURDIR)
DEVCONTAINER_EXEC = $(DEVCONTAINER) exec --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

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
		--format "table {{.ID}}\t{{.Names}}\t{{.Status}}\t{{.Image}}"
	@printf "\nRuntime\n"
	@if $(DEVCONTAINER_EXEC) /bin/bash -lc ' \
		printf "  %-12s %s\n" "Status:" "running"; \
		printf "  %-12s %s\n" "User:" "$$(whoami)"; \
		printf "  %-12s %s\n" "Directory:" "$$(pwd)"; \
		printf "  %-12s %s\n" "Go:" "$$(go version)"; \
		printf "  %-12s %s\n" "Node:" "$$(node --version)"; \
		printf "  %-12s %s\n" "npm:" "$$(npm --version)"; \
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

.PHONY: dev-orchestrator
dev-orchestrator: dev-check ## Run the orchestrator inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go run ./cmd/orchestrator -addr :8080 -db tasq.sqlite'

.PHONY: web-up
web-up: dev-check ## Run the Next.js web UI inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'cd web && npm run dev'

.PHONY: tui-up
tui-up: dev-check ## Run the TUI inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go run ./cmd/tasq-tui -api http://localhost:8080 -watch'

.PHONY: dev-gui
dev-gui: dev-check ## Run the Next.js GUI inside the Dev Container. Prefer web-up.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'cd web && npm run dev'
