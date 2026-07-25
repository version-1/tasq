package tq

import (
	"fmt"
	"io"
)

func printRootHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq [--api-url URL] [--output text|json] <resource> <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Resources:")
	fmt.Fprintln(w, "  issue    create, get, list, update, and shortcut issue actions")
	fmt.Fprintln(w, "  comment  add and list issue comments")
	fmt.Fprintln(w, "  project  add, remove, check, and list projects")
	fmt.Fprintln(w, "  workflow add, remove, and show project workflow resolution")
	fmt.Fprintln(w, "  migrate  apply, roll back, and inspect local database migrations")
	fmt.Fprintln(w, "  web      open the running Web UI in the default browser")
	fmt.Fprintln(w, "  service  start, stop, and inspect local services")
	fmt.Fprintln(w, "  logs     show and follow service logs")
	fmt.Fprintln(w, "  version  show version information")
	fmt.Fprintln(w, "  update   update tq from a GitHub Release and restart services")
}

func printIssueHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq issue <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  create   Create an issue (--project KEY required)")
	fmt.Fprintln(w, "  get      Get an issue by ID")
	fmt.Fprintln(w, "  list     List issues (--project KEY optional)")
	fmt.Fprintln(w, "  watch    Poll ready issues and emit JSON events (--interval, --seen-ttl, --verbose)")
	fmt.Fprintln(w, "  update   Update an issue")
	fmt.Fprintln(w, "  close    Mark an issue as done")
	fmt.Fprintln(w, "  cancel   Mark an issue as cancelled")
	fmt.Fprintln(w, "  ready    Mark an issue as ready")
	fmt.Fprintln(w, "  draft    Move an issue to backlog")
	fmt.Fprintln(w, "  rename   Rename an issue")
	fmt.Fprintln(w, "  edit     Update an issue description")
}

func printCommentHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq comment <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add a comment to an issue")
	fmt.Fprintln(w, "  list     List comments for an issue")
}

func printProjectHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq project <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add a project for a repository")
	fmt.Fprintln(w, "  remove   Remove a project by key (-y skips confirmation)")
	fmt.Fprintln(w, "  check    Check local project workflow files")
	fmt.Fprintln(w, "  list     List projects")
}

func printProjectRemoveHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq project remove [-y] <project-key>")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Flags:")
	fmt.Fprintln(w, "  -y  Remove without confirmation")
}

func printWorkflowHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: tq workflow <action> [flags]")
	fmt.Fprintln(w)
	fmt.Fprintln(w, "Actions:")
	fmt.Fprintln(w, "  add      Add or replace a project workflow override (--project KEY and --file PATH or --body TEXT required)")
	fmt.Fprintln(w, "  remove   Remove a project workflow override (--project KEY required)")
	fmt.Fprintln(w, "  show     Show the resolved project workflow")
}
