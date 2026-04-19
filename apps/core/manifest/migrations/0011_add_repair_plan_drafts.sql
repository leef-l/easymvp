CREATE TABLE IF NOT EXISTS repair_plan_drafts (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  failed_task_context_json TEXT NOT NULL,
  failure_reason_json TEXT NOT NULL,
  original_contracts_json TEXT NOT NULL,
  runtime_summary_json TEXT NOT NULL,
  artifact_refs_json TEXT,
  repair_plan_json TEXT NOT NULL,
  repair_reasoning_summary TEXT NOT NULL,
  replaced_constraints_json TEXT,
  status TEXT NOT NULL,
  created_by TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_repair_plan_drafts_project_updated
  ON repair_plan_drafts(project_id, updated_at DESC);
