# tasq

Symphony-compatible task queue orchestrator.

tasq provides a Go REST API backed by SQLite, a Next.js web UI, and a Go TUI that reads the same API.

## Components

- Orchestrator: Go REST API backed by SQLite.
- Web UI: React, TypeScript, and Next.js.
- TUI: Go terminal client for the orchestrator API.

## Recommended Development Flow

Use the Dev Container through `make` when developing this repository.

The Dev Container provides Go, Node.js, npm, and GitHub CLI. It also installs Go modules and web dependencies with the `postCreateCommand` in `.devcontainer/devcontainer.json`.

```sh
make dev-up
```

If this is the first run, or if the Dev Container image changed, the command can take a while because it downloads the base image, installs features, and runs dependency setup.

Run the services in separate terminals:

```sh
make dev-orchestrator
```

```sh
make web-up
```

```sh
make tui-up
```

Open `http://localhost:3000` for the web UI. The orchestrator API runs on `http://localhost:8080`.

Check the Dev Container status:

```sh
make dev-status
```

## Make Commands

Run `make help` to list available commands.

| Command | Purpose |
| --- | --- |
| `make help` | Show available targets. |
| `make dev-check` | Check that Docker CLI and Dev Container CLI are installed. |
| `make dev-build` | Build the Dev Container image. |
| `make dev-up` | Create or start the Dev Container. |
| `make dev-rebuild` | Recreate the Dev Container from scratch. |
| `make dev-status` | Show the Dev Container Docker status and print tool versions when it is running. |
| `make dev-shell` | Open a shell inside the running Dev Container. |
| `make dev-exec CMD="go test ./..."` | Run an arbitrary command inside the Dev Container. |
| `make dev-test` | Run Go tests and web UI typecheck inside the Dev Container. |
| `make dev-build-app` | Run Go tests and the web UI production build inside the Dev Container. |
| `make dev-orchestrator` | Run the orchestrator API inside the Dev Container. |
| `make web-up` | Run the Next.js web UI inside the Dev Container. |
| `make tui-up` | Run the TUI inside the Dev Container. |
| `make dev-gui` | Alias-style command for running the Next.js GUI inside the Dev Container. Prefer `make web-up`. |

## If `make dev-up` Fails

Check these in order:

1. Make sure Docker Desktop is running.
2. Make sure Docker CLI and Dev Container CLI are installed:

   ```sh
   make dev-check
   ```

3. Build the image explicitly:

   ```sh
   make dev-build
   ```

4. If an old container is stuck or was created from an older config, recreate it:

   ```sh
   make dev-rebuild
   ```

5. Check whether the container is running:

   ```sh
   make dev-status
   ```

6. If the error mentions downloading images or features, check network access to `mcr.microsoft.com` and `ghcr.io`.

7. If the error happens during dependency setup, open a shell and rerun the setup commands:

   ```sh
   make dev-shell
   go mod download
   cd web
   npm install
   ```

The Dev Container forwards `8080` for the orchestrator API and `3000` for the Next.js web UI.

## Host-Only Development

You can also run the project directly on the host if Go, Node.js, and npm are installed.

Run the orchestrator:

```sh
go run ./cmd/orchestrator -addr :8080 -db tasq.sqlite
```

Run the web UI:

```sh
cd web
npm install
npm run dev
```

Run the TUI:

```sh
go run ./cmd/tasq-tui -api http://localhost:8080 -watch
```

Set `NEXT_PUBLIC_ORCHESTRATOR_URL` when the API is not available at `http://localhost:8080`.

## Create a Task

```sh
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Wire Symphony workflow","description":"Define the first workflow contract","status":"backlog","priority":"high"}'
```

## REST API

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PATCH /api/v1/tasks/{id}`
- `GET /api/v1/settings`
- `PUT /api/v1/settings`
