---
id: tq-output-stories
title: Check tq Command Output
sidebar_position: 2
---

# Check tq Command Output

Use `tqstory` to review `tq` command output without starting services, creating databases, or calling the issue-tracker API. Each scenario uses fixed data and calls the same output renderers as the CLI.

Run a scenario from the repository root:

```sh
go run ./cmd/tqstory <scenario>
```

`go run` requires the `./` package prefix in a Go module.

## Scenarios

| Scenario | Output to review |
| --- | --- |
| `tq_issue_list` | Issue list with project and status colors. |
| `tq_issue_detail` | Full issue detail fields. |
| `tq_issue_action` | Successful issue action followed by issue detail. |
| `tq_project_list` | Project table and identifier colors. |
| `tq_empty` | Empty project-list message. |
| `tq_project_check` | PASS and FAIL project checks. |
| `tq_service_status` | Running and stopped service states. |
| `tq_migration_status` | Applied and pending migration states. |
| `tq_service_start_success` | Successful service-start result. |
| `tq_service_stop_success` | Successful service-stop result. |
| `tq_project_remove_success` | Successful project-removal result. |
| `tq_workflow_remove_success` | Successful workflow-override removal result. |
| `tq_project_remove_cancelled` | Cancelled project-removal result. |
| `tq_project_remove_confirmation` | Project-removal warning and confirmation prompt. |
| `tq_warning` | Destructive-operation warning. |
| `tq_service_start_fail` | Text-mode service-start failure. |
| `tq_json_success` | Successful JSON output without ANSI escape sequences. |
| `tq_json_error` | JSON error output without ANSI escape sequences. |
| `all` | Every scenario, with a heading before each result. |

## Check Command Results

Use these scenarios when reviewing success, cancellation, and confirmation output from commands:

```sh
go run ./cmd/tqstory tq_service_start_success
go run ./cmd/tqstory tq_service_stop_success
go run ./cmd/tqstory tq_project_remove_success
go run ./cmd/tqstory tq_workflow_remove_success
go run ./cmd/tqstory tq_project_remove_cancelled
go run ./cmd/tqstory tq_project_remove_confirmation
```

The confirmation scenario only renders the prompt; it does not read input or remove a project.

## Check Lists, Statuses, and Errors

```sh
go run ./cmd/tqstory tq_issue_list
go run ./cmd/tqstory tq_service_status
go run ./cmd/tqstory tq_migration_status
go run ./cmd/tqstory tq_service_start_fail
go run ./cmd/tqstory tq_json_error
```

Failure scenarios are visual previews and exit with status `0`. This makes them suitable for manual checks and screenshots. JSON scenarios do not include ANSI escape sequences.

## Review Every Scenario

Run all scenarios, with a heading before each result:

```sh
go run ./cmd/tqstory all
```

Pass an unknown scenario, or omit the scenario, to print the available scenario names and exit with status `2`.
