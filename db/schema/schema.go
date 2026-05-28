package schema

import _ "embed"

//go:embed issue_tracker.sql
var IssueTracker string

//go:embed orchestrator.sql
var Orchestrator string
