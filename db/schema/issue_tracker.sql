CREATE TABLE IF NOT EXISTS issues (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	assignee TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE INDEX IF NOT EXISTS issues_project_id_idx ON issues(project_id);
CREATE INDEX IF NOT EXISTS issues_status_idx ON issues(status);

CREATE TABLE IF NOT EXISTS issue_dependencies (
	parent_issue_id INTEGER NOT NULL,
	dependency_issue_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY(parent_issue_id, dependency_issue_id),
	CHECK(parent_issue_id <> dependency_issue_id),
	FOREIGN KEY(parent_issue_id) REFERENCES issues(id) ON DELETE CASCADE,
	FOREIGN KEY(dependency_issue_id) REFERENCES issues(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS issue_dependencies_dependency_issue_id_idx ON issue_dependencies(dependency_issue_id);

CREATE TABLE IF NOT EXISTS comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id INTEGER NOT NULL,
	author TEXT NOT NULL,
	type TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY(issue_id) REFERENCES issues(id)
);

CREATE INDEX IF NOT EXISTS comments_issue_id_idx ON comments(issue_id);

CREATE TABLE IF NOT EXISTS projects (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	key TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	location TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS projects_key_idx ON projects(key);

CREATE TABLE IF NOT EXISTS project_workflows (
	project_id INTEGER NOT NULL UNIQUE,
	frontmatter_json TEXT NOT NULL,
	body TEXT NOT NULL,
	checksum TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE TABLE IF NOT EXISTS attachments (
	id TEXT PRIMARY KEY,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	filename TEXT NOT NULL,
	path TEXT NOT NULL,
	content_type TEXT NOT NULL,
	size INTEGER NOT NULL,
	created_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS attachments_entity_idx ON attachments(entity_type, entity_id);

CREATE TABLE IF NOT EXISTS schema_migrations (
	version TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);
