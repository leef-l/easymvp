CREATE INDEX IF NOT EXISTS idx_projects_updated_created
  ON projects(updated_at DESC, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_acceptance_runs_project_created
  ON acceptance_runs(project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_acceptance_issues_project_created
  ON acceptance_issues(project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_evidence_project_effective_captured
  ON evidence_items(project_id, COALESCE(captured_at, created_at) DESC, created_at DESC);
