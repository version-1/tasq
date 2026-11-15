CREATE TABLE IF NOT EXISTS issues (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	project_id INTEGER NOT NULL,
	title TEXT NOT NULL,
	description TEXT NOT NULL DEFAULT '',
	status TEXT NOT NULL,
	priority TEXT NOT NULL,
	assignee TEXT NOT NULL DEFAULT '',
	auto_ok INTEGER NOT NULL DEFAULT 0,
	auto_ok_count INTEGER NOT NULL DEFAULT 0 CHECK(auto_ok_count BETWEEN 0 AND 5),
	auto_ok_pending INTEGER NOT NULL DEFAULT 0,
	auto_ok_claimed INTEGER NOT NULL DEFAULT 0,
	auto_ok_generation INTEGER NOT NULL DEFAULT 0,
	auto_ok_stop_reason TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);

CREATE INDEX IF NOT EXISTS issues_project_id_idx ON issues(project_id);
CREATE INDEX IF NOT EXISTS issues_status_idx ON issues(status);

CREATE TABLE IF NOT EXISTS issue_artifacts (
	issue_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	data_type TEXT NOT NULL,
	data_value TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (issue_id, type),
	FOREIGN KEY(issue_id) REFERENCES issues(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS issue_artifacts_issue_id_idx ON issue_artifacts(issue_id);

CREATE TABLE IF NOT EXISTS document_artifacts (
	id TEXT PRIMARY KEY,
	project_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	body TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS document_artifacts_project_id_idx ON document_artifacts(project_id, created_at, id);

CREATE TABLE IF NOT EXISTS document_artifact_links (
	document_artifact_id TEXT NOT NULL,
	target_type TEXT NOT NULL CHECK(target_type = 'issue'),
	target_id INTEGER NOT NULL,
	created_at TEXT NOT NULL,
	PRIMARY KEY(document_artifact_id, target_type, target_id),
	FOREIGN KEY(document_artifact_id) REFERENCES document_artifacts(id) ON DELETE RESTRICT,
	FOREIGN KEY(target_id) REFERENCES issues(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS document_artifact_links_target_idx ON document_artifact_links(target_type, target_id, document_artifact_id);

CREATE TRIGGER IF NOT EXISTS prevent_project_delete_with_external_document_links
BEFORE DELETE ON projects
WHEN EXISTS (SELECT 1 FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id = a.id JOIN issues i ON i.id = l.target_id WHERE a.project_id = OLD.id AND i.project_id <> OLD.id)
BEGIN SELECT RAISE(ABORT, 'project document artifact has external issue links'); END;

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

CREATE TABLE IF NOT EXISTS change_requests (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id INTEGER NOT NULL,
	author TEXT NOT NULL,
	body TEXT NOT NULL,
	status TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	resolved_at TEXT,
	resolved_by_run_id TEXT,
	result_comment_id INTEGER,
	FOREIGN KEY(issue_id) REFERENCES issues(id),
	FOREIGN KEY(result_comment_id) REFERENCES comments(id)
);

CREATE INDEX IF NOT EXISTS change_requests_issue_id_idx ON change_requests(issue_id);
CREATE INDEX IF NOT EXISTS change_requests_issue_status_idx ON change_requests(issue_id, status, created_at, id);

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
