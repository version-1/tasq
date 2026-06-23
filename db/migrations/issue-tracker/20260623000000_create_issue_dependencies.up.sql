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
