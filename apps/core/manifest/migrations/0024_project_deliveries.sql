CREATE TABLE IF NOT EXISTS project_deliveries (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL REFERENCES projects(id),
  status TEXT NOT NULL DEFAULT 'pending',
  workspace_path TEXT NOT NULL DEFAULT '',
  readme TEXT NOT NULL DEFAULT '',
  architecture_doc TEXT NOT NULL DEFAULT '',
  api_docs TEXT NOT NULL DEFAULT '',
  deployment_doc TEXT NOT NULL DEFAULT '',
  test_report_json TEXT NOT NULL DEFAULT '{}',
  statistics_json TEXT NOT NULL DEFAULT '{}',
  user_accepted INTEGER NOT NULL DEFAULT 0,
  accepted_at TEXT,
  delivered_at TEXT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX IF NOT EXISTS idx_project_deliveries_project_id ON project_deliveries(project_id);
