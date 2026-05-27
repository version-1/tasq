DEVCONTAINER ?= devcontainer
DEVCONTAINER_WORKSPACE ?= $(CURDIR)
DEVCONTAINER_EXEC = $(DEVCONTAINER) exec --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: dev-check
dev-check: ## Check that the Dev Container CLI is installed.
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
dev-test: dev-check ## Run Go tests and GUI typecheck inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run typecheck'

.PHONY: dev-build-app
dev-build-app: dev-check ## Run production build checks inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run build'

.PHONY: dev-orchestrator
dev-orchestrator: dev-check ## Run the orchestrator inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go run ./cmd/orchestrator -addr :8080 -db tasq.sqlite'

.PHONY: dev-gui
dev-gui: dev-check ## Run the Next.js GUI inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'cd web && npm run dev'
