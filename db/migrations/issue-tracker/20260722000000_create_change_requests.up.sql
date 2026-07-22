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
