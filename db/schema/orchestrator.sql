CREATE TABLE IF NOT EXISTS runs (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	run_id TEXT NOT NULL UNIQUE,
	issue_id INTEGER NOT NULL,
	work_item_id INTEGER NOT NULL,
	claim_token TEXT NOT NULL,
	status TEXT NOT NULL,
	workspace TEXT NOT NULL DEFAULT '',
	attempt INTEGER NOT NULL DEFAULT 0,
	error TEXT NOT NULL DEFAULT '',
	orchestrator_id TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS runs_work_item_idx ON runs(work_item_id);
CREATE INDEX IF NOT EXISTS runs_issue_idx ON runs(issue_id, id);
CREATE INDEX IF NOT EXISTS runs_status_updated_idx ON runs(status, updated_at, id);

CREATE TABLE IF NOT EXISTS runner_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	run_id TEXT NOT NULL,
	event_type TEXT NOT NULL,
	message TEXT NOT NULL DEFAULT '',
	payload_json TEXT NOT NULL DEFAULT '',
	occurred_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS runner_events_run_idx ON runner_events(run_id, id);

CREATE TABLE IF NOT EXISTS workspace_metadata (
	workspace_key TEXT PRIMARY KEY,
	issue_id INTEGER NOT NULL,
	path TEXT NOT NULL,
	created_now INTEGER NOT NULL DEFAULT 0,
	source_path TEXT NOT NULL DEFAULT '',
	populated_at TEXT NOT NULL DEFAULT '',
	cleanup_status TEXT NOT NULL DEFAULT '',
	cleanup_at TEXT NOT NULL DEFAULT '',
	last_error TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS workspace_setup_failures (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id INTEGER NOT NULL,
	workspace_key TEXT NOT NULL DEFAULT '',
	path TEXT NOT NULL DEFAULT '',
	error TEXT NOT NULL,
	occurred_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS workspace_setup_failures_issue_idx ON workspace_setup_failures(issue_id, id);

CREATE TABLE IF NOT EXISTS outbox_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	event_id TEXT NOT NULL UNIQUE,
	run_id TEXT NOT NULL,
	issue_id INTEGER NOT NULL,
	work_item_id INTEGER NOT NULL,
	claim_token TEXT NOT NULL,
	status TEXT NOT NULL,
	workspace TEXT NOT NULL DEFAULT '',
	attempt INTEGER NOT NULL DEFAULT 0,
	error TEXT NOT NULL DEFAULT '',
	orchestrator_id TEXT NOT NULL,
	occurred_at TEXT NOT NULL,
	sent_at TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS outbox_events_unsent_idx ON outbox_events(sent_at, id);
