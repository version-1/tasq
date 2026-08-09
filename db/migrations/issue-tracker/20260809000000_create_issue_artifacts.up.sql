CREATE TABLE issue_artifacts (
	issue_id INTEGER NOT NULL,
	type TEXT NOT NULL,
	data_type TEXT NOT NULL,
	data_value TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (issue_id, type),
	FOREIGN KEY(issue_id) REFERENCES issues(id) ON DELETE CASCADE
);

CREATE INDEX issue_artifacts_issue_id_idx ON issue_artifacts(issue_id);
