CREATE TABLE IF NOT EXISTS project_workflows (
	project_id INTEGER NOT NULL UNIQUE,
	frontmatter_json TEXT NOT NULL,
	body TEXT NOT NULL,
	checksum TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY(project_id) REFERENCES projects(id)
);
