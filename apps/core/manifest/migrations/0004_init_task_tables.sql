CREATE TABLE IF NOT EXISTS domain_tasks (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  source_compiled_plan_id TEXT NOT NULL,
  source_compiled_task_id TEXT NOT NULL,
  source_task_key TEXT NOT NULL,
  compiled_version INTEGER NOT NULL,
  name TEXT NOT NULL,
  phase TEXT NOT NULL,
  task_kind TEXT NOT NULL,
  role_type TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  risk_level TEXT NOT NULL,
  status TEXT NOT NULL,
  manual_review_required INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_dependencies (
  task_id TEXT NOT NULL,
  depends_on_task_id TEXT NOT NULL,
  created_at TEXT NOT NULL,
  PRIMARY KEY(task_id, depends_on_task_id),
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE,
  FOREIGN KEY(depends_on_task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS task_manual_gates (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  gate_kind TEXT NOT NULL,
  gate_status TEXT NOT NULL,
  comment TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE
);
