# tasq
Symphony-compatible task queue orchestrator.

## Architecture

- Orchestrator: Go REST API backed by SQLite
- GUI: React, TypeScript, and Next.js
- TUI: Go terminal client that reads the same REST API

## Run the orchestrator

```sh
go run ./cmd/orchestrator -addr :8080 -db tasq.sqlite
```

## Create a task

```sh
curl -X POST http://localhost:8080/api/v1/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Wire Symphony workflow","description":"Define the first workflow contract","status":"backlog","priority":"high"}'
```

## Run the TUI

```sh
go run ./cmd/tasq-tui -api http://localhost:8080 -watch
```

## Run the GUI

```sh
cd web
npm install
npm run dev
```

Open `http://localhost:3000`. Set `NEXT_PUBLIC_ORCHESTRATOR_URL` when the API is not on `http://localhost:8080`.

## Dev container

Open this repository in a Dev Container to get Go, Node.js, npm, and GitHub CLI installed with the project dependencies.

You can manage the container from the host with `make`:

```sh
make dev-up
make dev-shell
make dev-test
```

After the container is created, run the services in separate terminals:

```sh
make dev-orchestrator
```

```sh
make web-up
```

```sh
make tui-up
```

The container forwards `8080` for the orchestrator API and `3000` for the Next.js GUI.

## REST API

- `GET /api/v1/health`
- `GET /api/v1/summary`
- `GET /api/v1/tasks`
- `POST /api/v1/tasks`
- `GET /api/v1/tasks/{id}`
- `PATCH /api/v1/tasks/{id}`
- `GET /api/v1/settings`
- `PUT /api/v1/settings`
