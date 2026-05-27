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

Start the web UI with the orchestrator. Both processes run in the Dev Container background:

```sh
make web-up
```

Or start the TUI with the orchestrator. The orchestrator runs in the Dev Container background, and the TUI runs in the foreground:

```sh
make tui-up
```

The Dev Container publishes host ports automatically so multiple worktrees can run in parallel.
Run `make dev-status` and open the host port mapped to container port `3000`.
The web UI proxies `/api/v1/...` to the orchestrator inside the same Dev Container.
If the published ports do not appear after changing `.devcontainer/devcontainer.json`, run `make dev-rebuild` so Docker recreates the container.

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
| `make dev-rebuild` | Recreate the Dev Container from scratch and apply port publishing changes. |
| `make dev-status` | Show the Dev Container Docker status and print tool versions when it is running. |
| `make dev-shell` | Open a shell inside the running Dev Container. |
| `make dev-exec CMD="go test ./..."` | Run an arbitrary command inside the Dev Container. |
| `make dev-test` | Run Go tests and web UI typecheck inside the Dev Container. |
| `make dev-build-app` | Run Go tests and the web UI production build inside the Dev Container. |
| `make orchestrator-up` | Start the orchestrator API in the Dev Container background. |
| `make dev-orchestrator` | Run the orchestrator API in the Dev Container foreground. |
| `make web-up` | Start the orchestrator and Next.js web UI in the Dev Container background. |
| `make tui-up` | Start the orchestrator in the background and run the TUI in the foreground. |
| `make dev-gui` | Alias-style command for `make web-up`. Prefer `make web-up`. |

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

4. If host ports are not listed in `make dev-status`, or if an old container was created from an older config, recreate it:

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

Set `NEXT_PUBLIC_ORCHESTRATOR_URL` only when the web UI should call an API outside the local Dev Container.

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
