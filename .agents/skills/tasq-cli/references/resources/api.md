# `tq api`

Send a raw request to an explicitly allowlisted issue-tracker endpoint. Prefer typed `tq issue`, `comment`, `project`, or `workflow` commands when they expose the operation. `tq api` is not a general-purpose HTTP client.

Global flags and API URL resolution are documented in [globals.md](globals.md). Place global flags before `api`.

## Syntax

```text
tq [--api-url URL] api <method> <path> \
  [--query key=value] \
  [--header 'Name: value'] \
  [--data value|@file|-]
```

- Methods are case-insensitive and normalized to uppercase.
- Repeat `--query` to append multiple query values in order.
- Repeat `--header` to set multiple headers. For duplicate names, the last value wins.
- Use `--data` only with `POST`, `PUT`, or `PATCH`.

## Examples

```sh
# Read queued issues.
tq api GET /api/v1/queue

# Preserve a query already in the path and append another value.
tq api GET '/api/v1/issues?states=ready' --query states=review

# Send a literal request body.
tq api PATCH /api/v1/issues/42 --data '{"status":"done"}'

# Read a request body from a file or standard input.
tq api POST /api/v1/issues --data @request.json
printf '%s' '{"status":"ready"}' | tq api PATCH /api/v1/issues/42 --data -

# Add a request header. Quote the whole Name: value argument.
tq api POST /api/v1/issues \
  --header 'X-Request-ID: local-123' \
  --data @request.json
```

## Allowed routes

The method and path must match this fail-closed allowlist. New server routes remain unavailable until the CLI allowlist is updated.

| Area | Allowed method and path |
| --- | --- |
| Health | `GET /api/v1/health` |
| Summary | `GET /api/v1/summary` |
| Projects | `GET, POST /api/v1/projects` |
| Project | `GET, PATCH, DELETE /api/v1/projects/{id}` |
| Project check | `POST /api/v1/projects/{id}/check` |
| Project workflow | `GET, PUT, DELETE /api/v1/projects/{id}/workflow` |
| Issues | `GET, POST /api/v1/issues` |
| Issue states | `POST /api/v1/issues/states` |
| Queue | `GET /api/v1/queue` |
| Issue | `GET, PATCH /api/v1/issues/{id}` |
| Issue comments | `GET, POST /api/v1/issues/{issueId}/comments` |
| Comment | `PATCH /api/v1/comments/{id}` |
| Issue change requests | `GET, POST /api/v1/issues/{issueId}/change-requests` |
| Change request | `GET, PATCH /api/v1/change-requests/{id}` |
| Cancel change request | `POST /api/v1/change-requests/{id}/cancel` |
| Attachments | `GET /api/v1/attachments` |
| Attachment content | `GET /api/v1/attachments/{attachmentId}/content` |
| Delete attachment | `DELETE /api/v1/attachments/{attachmentId}` |

`{id}` and `{issueId}` must be positive signed 64-bit integers. `{attachmentId}` must be non-empty.

`POST /api/v1/attachments` is not allowed because `tq api` does not construct multipart requests. Attachment `PATCH` is also not allowed.

## Path and query rules

- Pass an absolute, unencoded `/api/v1/...` path rather than a complete URL.
- Do not use fragments, encoded path segments, dot segments, empty segments, or a trailing slash.
- A query already present in the path is preserved.
- Each `--query` must use `key=value`; names and values are URL-encoded and appended without API-specific validation.

## Headers and request bodies

Transport-managed headers are rejected, including `Host`, `Content-Length`, `Transfer-Encoding`, `Connection`, `Trailer`, `Upgrade`, `Proxy-Connection`, `Keep-Alive`, and `TE`.

`--data` accepts:

| Value | Body source |
| --- | --- |
| `value` | literal argument |
| `@file` | file contents |
| `-` | standard input |

The body is not validated as JSON. When a body is present and `Content-Type` is not specified, the command sets `Content-Type: application/json`.

## Output, redirects, and exit codes

- Response bytes are written unchanged to stdout, including binary data and HTTP error bodies.
- Global `--output text|json` does not transform the response.
- Redirects are not followed.
- The HTTP client timeout is 10 seconds.
- The command does not prompt before write or delete operations.

| Exit code | Meaning |
| --- | --- |
| `0` | HTTP `2xx` |
| `1` | HTTP `3xx`-`5xx` or transport failure |
| `2` | Usage, input, or allowlist error |
