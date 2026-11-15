-- The old model can represent exactly one same-project issue link per artifact.
CREATE TABLE _document_artifact_rollback_guard (ok INTEGER NOT NULL CHECK(ok = 1));
INSERT INTO _document_artifact_rollback_guard (ok)
SELECT 0 WHERE EXISTS (
 SELECT 1 FROM document_artifacts a
 LEFT JOIN document_artifact_links l ON l.document_artifact_id = a.id
 LEFT JOIN issues i ON i.id = l.target_id
 GROUP BY a.id
 HAVING COUNT(l.target_id) <> 1 OR MIN(i.project_id) <> MIN(a.project_id)
);
DROP TABLE _document_artifact_rollback_guard;
CREATE TABLE issue_document_artifacts (
 id TEXT PRIMARY KEY, issue_id INTEGER NOT NULL, type TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
 FOREIGN KEY(issue_id) REFERENCES issues(id) ON DELETE CASCADE
);
INSERT INTO issue_document_artifacts (id, issue_id, type, title, body, created_at, updated_at)
SELECT a.id, l.target_id, a.type, a.title, a.body, a.created_at, a.updated_at FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id = a.id;
DROP TRIGGER prevent_project_delete_with_external_document_links;
DROP TABLE document_artifact_links;
DROP TABLE document_artifacts;
CREATE INDEX issue_document_artifacts_issue_id_idx ON issue_document_artifacts(issue_id, created_at, id);
