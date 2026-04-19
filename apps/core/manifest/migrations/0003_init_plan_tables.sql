CREATE TABLE IF NOT EXISTS workflow_plan_drafts (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  version INTEGER NOT NULL,
  source_kind TEXT NOT NULL,
  source_run_id TEXT,
  project_category TEXT NOT NULL,
  goal_summary TEXT NOT NULL,
  input_requirements_json TEXT NOT NULL,
  draft_tasks_json TEXT NOT NULL,
  status TEXT NOT NULL,
  created_by TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(project_id, version)
);

CREATE TABLE IF NOT EXISTS workflow_plan_review_results (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_draft_id TEXT NOT NULL,
  review_version INTEGER NOT NULL,
  review_run_id TEXT,
  decision TEXT NOT NULL,
  blocking_issue_count INTEGER NOT NULL DEFAULT 0,
  advisory_issue_count INTEGER NOT NULL DEFAULT 0,
  issues_json TEXT NOT NULL,
  split_suggestions_json TEXT,
  override_suggestions_json TEXT,
  status TEXT NOT NULL,
  reviewed_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id) ON DELETE CASCADE,
  UNIQUE(plan_draft_id, review_version)
);

CREATE TABLE IF NOT EXISTS workflow_compiled_plans (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  plan_draft_id TEXT NOT NULL,
  plan_review_result_id TEXT NOT NULL,
  compiled_version INTEGER NOT NULL,
  compile_run_id TEXT,
  project_category TEXT NOT NULL,
  status TEXT NOT NULL,
  risk_summary_json TEXT,
  compile_diff_json TEXT,
  generated_at TEXT NOT NULL,
  activated_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_draft_id) REFERENCES workflow_plan_drafts(id) ON DELETE CASCADE,
  FOREIGN KEY(plan_review_result_id) REFERENCES workflow_plan_review_results(id) ON DELETE CASCADE,
  UNIQUE(project_id, compiled_version)
);

CREATE TABLE IF NOT EXISTS workflow_compiled_tasks (
  id TEXT PRIMARY KEY,
  compiled_plan_id TEXT NOT NULL,
  task_key TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT,
  phase TEXT NOT NULL,
  task_kind TEXT NOT NULL,
  role_type TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  affected_resources_json TEXT NOT NULL,
  delivery_contract_json TEXT NOT NULL,
  verification_contract_json TEXT NOT NULL,
  manual_review_required INTEGER NOT NULL DEFAULT 0,
  depends_on_task_keys_json TEXT,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(compiled_plan_id) REFERENCES workflow_compiled_plans(id) ON DELETE CASCADE,
  UNIQUE(compiled_plan_id, task_key)
);
