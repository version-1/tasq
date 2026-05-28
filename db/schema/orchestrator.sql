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
