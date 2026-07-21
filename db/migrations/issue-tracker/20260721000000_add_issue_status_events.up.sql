ALTER TABLE issues ADD COLUMN changed_reason TEXT;

CREATE TABLE IF NOT EXISTS issue_status_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	issue_id INTEGER NOT NULL,
	from_status TEXT NOT NULL,
	to_status TEXT NOT NULL,
	changed_reason TEXT,
	actor TEXT NOT NULL DEFAULT '',
	comment_id INTEGER,
	created_at TEXT NOT NULL,
	FOREIGN KEY(issue_id) REFERENCES issues(id),
	FOREIGN KEY(comment_id) REFERENCES comments(id)
);

CREATE INDEX IF NOT EXISTS issue_status_events_issue_id_idx ON issue_status_events(issue_id);
