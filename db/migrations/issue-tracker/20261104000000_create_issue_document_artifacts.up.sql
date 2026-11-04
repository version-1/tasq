CREATE TABLE issue_document_artifacts (
 id TEXT PRIMARY KEY,
 issue_id INTEGER NOT NULL,
 type TEXT NOT NULL,
 title TEXT NOT NULL,
 body TEXT NOT NULL,
 created_at TEXT NOT NULL,
 updated_at TEXT NOT NULL,
 FOREIGN KEY(issue_id) REFERENCES issues(id) ON DELETE CASCADE
);
CREATE INDEX issue_document_artifacts_issue_id_idx ON issue_document_artifacts(issue_id, created_at, id);
