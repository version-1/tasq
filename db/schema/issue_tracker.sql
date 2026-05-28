CREATE TABLE IF NOT EXISTS issues (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	assignee TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS issues_status_idx ON issues(status);

CREATE TABLE IF NOT EXISTS projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS projects_key_idx ON projects(key);

CREATE TABLE IF NOT EXISTS workspaces (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	name TEXT NOT NULL,
	path TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE INDEX IF NOT EXISTS workspaces_project_id_idx ON workspaces(project_id);
CREATE INDEX IF NOT EXISTS workspaces_status_idx ON workspaces(status);

CREATE TABLE IF NOT EXISTS work_items (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id INTEGER NOT NULL,
	status TEXT NOT NULL,
	claimed_by TEXT NOT NULL DEFAULT '',
	claim_token TEXT NOT NULL DEFAULT '',
	lease_until TEXT NOT NULL DEFAULT '',
	attempt INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(issue_id) REFERENCES issues(id)
);

CREATE INDEX IF NOT EXISTS work_items_status_lease_idx ON work_items(status, lease_until);

CREATE TABLE IF NOT EXISTS orchestrator_events (
	event_id TEXT PRIMARY KEY,
	work_item_id INTEGER NOT NULL,
	issue_id INTEGER NOT NULL,
	run_id TEXT NOT NULL,
	claim_token TEXT NOT NULL,
	status TEXT NOT NULL,
	workspace TEXT NOT NULL DEFAULT '',
	attempt INTEGER NOT NULL DEFAULT 0,
	error TEXT NOT NULL DEFAULT '',
	occurred_at TEXT NOT NULL,
	orchestrator_id TEXT NOT NULL DEFAULT '',
	received_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS run_snapshots (
	issue_id INTEGER PRIMARY KEY,
	work_item_id INTEGER NOT NULL,
	run_id TEXT NOT NULL,
	status TEXT NOT NULL,
	workspace TEXT NOT NULL DEFAULT '',
	attempt INTEGER NOT NULL DEFAULT 0,
	error TEXT NOT NULL DEFAULT '',
	orchestrator_id TEXT NOT NULL DEFAULT '',
	updated_at TEXT NOT NULL
);
