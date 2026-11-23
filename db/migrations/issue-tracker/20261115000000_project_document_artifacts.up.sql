CREATE TABLE document_artifacts (
 id TEXT PRIMARY KEY, project_id INTEGER NOT NULL, type TEXT NOT NULL, title TEXT NOT NULL, body TEXT NOT NULL, created_at TEXT NOT NULL, updated_at TEXT NOT NULL,
 FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
INSERT INTO document_artifacts SELECT a.id, i.project_id, a.type, a.title, a.body, a.created_at, a.updated_at FROM issue_document_artifacts a JOIN issues i ON i.id = a.issue_id;
CREATE TABLE document_artifact_links (
 document_artifact_id TEXT NOT NULL, target_type TEXT NOT NULL CHECK(target_type = 'issue'), target_id INTEGER NOT NULL, created_at TEXT NOT NULL,
 PRIMARY KEY(document_artifact_id, target_type, target_id),
 FOREIGN KEY(document_artifact_id) REFERENCES document_artifacts(id) ON DELETE RESTRICT,
 FOREIGN KEY(target_id) REFERENCES issues(id) ON DELETE CASCADE
);
INSERT INTO document_artifact_links SELECT id, 'issue', issue_id, created_at FROM issue_document_artifacts;
DROP TABLE issue_document_artifacts;
CREATE INDEX document_artifacts_project_id_idx ON document_artifacts(project_id, created_at, id);
CREATE INDEX document_artifact_links_target_idx ON document_artifact_links(target_type, target_id, document_artifact_id);
CREATE TRIGGER prevent_project_delete_with_external_document_links
BEFORE DELETE ON projects
WHEN EXISTS (SELECT 1 FROM document_artifacts a JOIN document_artifact_links l ON l.document_artifact_id = a.id JOIN issues i ON i.id = l.target_id WHERE a.project_id = OLD.id AND i.project_id <> OLD.id)
BEGIN SELECT RAISE(ABORT, 'project document artifact has external issue links'); END;
