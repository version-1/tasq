# tasq

Local-first issue tracker and task orchestrator.

tasq provides a Go issue-tracker API backed by SQLite, a Go orchestrator backed by SQLite, a Next.js web UI, and a Go TUI.

## Components

- Issue Tracker: Go REST API backed by SQLite. It owns issues, work items, and UI summaries.
- Orchestrator: Go worker backed by SQLite. It claims work items and records run state.
- Web UI: React, TypeScript, and Next.js client for the issue-tracker API.
- TUI: Go terminal client for the issue-tracker API.

## Recommended Development Flow

Use Docker Compose through `make` when developing this repository.

The Compose environment runs three services:

- `issue-tracker`: Go API on container port `8080`.
- `orchestrator`: Go worker connected to `http://issue-tracker:8080`.
- `web`: Next.js dev server on container port `3000`.

Start the full local environment in the background:

```sh
make web-up
```

Open the web UI at <http://localhost:3000>.
The issue-tracker API is available at <http://localhost:8080>.
The web UI proxies `/api/v1/...` to `issue-tracker` inside the Compose network.

To run the services in the foreground:

```sh
make dev-up
```

To stop the environment:

```sh
make dev-down
```

If another worktree or local process already uses the default ports, override them:

```sh
make web-up ISSUE_TRACKER_PORT=18080 WEB_PORT=13000
```

## Make Commands

Run `make help` to list available commands.

| Command | Purpose |
| --- | --- |
| `make help` | Show available targets. |
| `make dev-check` | Check that Docker CLI and Docker Compose are installed. |
| `make dev-up` | Start issue-tracker, orchestrator, and web UI in the foreground. |
| `make dev-up-d` | Start issue-tracker, orchestrator, and web UI in the background. |
| `make dev-down` | Stop Compose services. |
| `make dev-status` | Show Compose service status. |
| `make dev-logs` | Follow Compose service logs. |
| `make dev-shell` | Open a shell in a Go tool container. |
| `make dev-exec CMD="go test ./..."` | Run an arbitrary command in a Go tool container. |
| `make dev-test` | Run Go tests and web UI typecheck in Compose containers. |
| `make dev-build-app` | Run Go tests and the web UI production build in Compose containers. |
| `make issue-tracker-up` | Start the issue-tracker API in the background. |
| `make orchestrator-up` | Start issue-tracker and orchestrator in the background. |
| `make web-up` | Start issue-tracker, orchestrator, and Next.js web UI in the background. |
| `make tui-up` | Start issue-tracker and orchestrator in Compose, then run the TUI on the host. |
| `make dev-gui` | Alias-style command for `make web-up`. |

## If `make web-up` Fails

Check these in order:

1. Make sure Docker Desktop is running.
2. Make sure Docker CLI and Docker Compose are installed:

   ```sh
   make dev-check
   ```

3. Check the service status:

   ```sh
   make dev-status
   ```

4. Check the service logs:

   ```sh
   make dev-logs
   ```

5. If the error mentions downloading images or dependencies, check network access to Docker Hub, the Go module proxy, and the npm registry.

6. If host ports are already in use, run with alternate host ports:

   ```sh
   make web-up ISSUE_TRACKER_PORT=18080 WEB_PORT=13000
   ```

Compose stores Go module/build caches and `web/node_modules` in named Docker volumes.
SQLite files are created under `.data/` in the repository and are ignored by git.

## Host-Only Development

You can also run the project directly on the host if Go, Node.js, and npm are installed.

Run the issue-tracker:

```sh
go run ./cmd/issue-tracker -addr :8080 -db tasq-issues.sqlite
```

Run the orchestrator:

```sh
go run ./cmd/orchestrator -db tasq-orchestrator.sqlite -issue-tracker http://localhost:8080
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

Set `NEXT_PUBLIC_ISSUE_TRACKER_URL` only when the browser should call an issue-tracker API outside the local web origin.
Set `ISSUE_TRACKER_INTERNAL_URL` only when the Next.js rewrite should proxy to an issue-tracker API other than `http://localhost:8080`.

## Create an Issue

```sh
curl -X POST http://localhost:8080/api/v1/issues \
  -H 'Content-Type: application/json' \
  -d '{"title":"Wire Symphony workflow","description":"Define the first workflow contract","status":"ready","priority":"high"}'
```

## REST API

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/issues`
- `POST /api/v1/issues`
- `GET /api/v1/issues/{id}`
- `PATCH /api/v1/issues/{id}`
- `POST /api/v1/work-items/claim`
- `POST /api/v1/orchestrator-events`
