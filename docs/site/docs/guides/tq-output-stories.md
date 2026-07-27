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
