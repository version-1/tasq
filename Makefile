DEVCONTAINER ?= devcontainer
DEVCONTAINER_WORKSPACE ?= $(CURDIR)
DEVCONTAINER_EXEC = $(DEVCONTAINER) exec --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: help
help: ## Show available targets.
	@awk 'BEGIN {FS = ":.*## "}; /^[a-zA-Z0-9_-]+:.*## / {printf "%-28s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: devcontainer-check
devcontainer-check: ## Check that the Dev Container CLI is installed.
	@command -v $(DEVCONTAINER) >/dev/null 2>&1 || { \
		echo "devcontainer CLI is required. Install @devcontainers/cli or use VS Code Dev Containers."; \
		exit 1; \
	}

.PHONY: devcontainer-build
devcontainer-build: devcontainer-check ## Build the Dev Container image.
	$(DEVCONTAINER) build --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: devcontainer-up
devcontainer-up: devcontainer-check ## Create or start the Dev Container.
	$(DEVCONTAINER) up --workspace-folder "$(DEVCONTAINER_WORKSPACE)"

.PHONY: devcontainer-rebuild
devcontainer-rebuild: devcontainer-check ## Recreate the Dev Container from scratch.
	$(DEVCONTAINER) up --workspace-folder "$(DEVCONTAINER_WORKSPACE)" --remove-existing-container

.PHONY: devcontainer-shell
devcontainer-shell: devcontainer-check ## Open a shell inside the running Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash

.PHONY: devcontainer-exec
devcontainer-exec: devcontainer-check ## Run CMD inside the Dev Container. Example: make devcontainer-exec CMD="go test ./..."
	@if [ -z "$(CMD)" ]; then \
		echo 'Usage: make devcontainer-exec CMD="go test ./..."'; \
		exit 1; \
	fi
	$(DEVCONTAINER_EXEC) /bin/bash -lc '$(CMD)'

.PHONY: devcontainer-test
devcontainer-test: devcontainer-check ## Run Go tests and GUI typecheck inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run typecheck'

.PHONY: devcontainer-build-app
devcontainer-build-app: devcontainer-check ## Run production build checks inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go test ./... && cd web && npm run build'

.PHONY: devcontainer-orchestrator
devcontainer-orchestrator: devcontainer-check ## Run the orchestrator inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'go run ./cmd/orchestrator -addr :8080 -db tasq.sqlite'

.PHONY: devcontainer-gui
devcontainer-gui: devcontainer-check ## Run the Next.js GUI inside the Dev Container.
	$(DEVCONTAINER_EXEC) /bin/bash -lc 'cd web && npm run dev'
