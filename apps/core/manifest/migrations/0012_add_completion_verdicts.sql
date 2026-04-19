CREATE TABLE IF NOT EXISTS completion_verdicts (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  decision TEXT NOT NULL,
  final_status TEXT NOT NULL,
  reason TEXT NOT NULL,
  summary TEXT NOT NULL,
  next_action TEXT NOT NULL,
  blocker_count INTEGER NOT NULL DEFAULT 0,
  release_ready INTEGER NOT NULL DEFAULT 0,
  completed INTEGER NOT NULL DEFAULT 0,
  manual_review_required INTEGER NOT NULL DEFAULT 0,
  manual_release_required INTEGER NOT NULL DEFAULT 0,
  manual_release_completed INTEGER NOT NULL DEFAULT 0,
  source_run_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE,
  UNIQUE(acceptance_run_id)
);

CREATE INDEX IF NOT EXISTS idx_completion_verdicts_project_updated
  ON completion_verdicts(project_id, updated_at DESC);
