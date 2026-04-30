CREATE TABLE IF NOT EXISTS design_reviews (
  id TEXT PRIMARY KEY,
  design_id TEXT NOT NULL DEFAULT '',
  project_id TEXT NOT NULL REFERENCES projects(id),
  round INTEGER NOT NULL DEFAULT 1,
  passed INTEGER NOT NULL DEFAULT 0,
  score INTEGER NOT NULL DEFAULT 0,
  dimensions_json TEXT NOT NULL DEFAULT '[]',
  issues_json TEXT NOT NULL DEFAULT '[]',
  suggestions_json TEXT NOT NULL DEFAULT '[]',
  fix_tasks_json TEXT NOT NULL DEFAULT '[]',
  brain_run_id TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_design_reviews_design_id ON design_reviews(design_id);
CREATE INDEX idx_design_reviews_project_id ON design_reviews(project_id);
