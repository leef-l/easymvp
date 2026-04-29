-- Add missing foreign keys to domain_tasks for data integrity
-- SQLite requires table recreation to add foreign keys

PRAGMA foreign_keys = OFF;

BEGIN TRANSACTION;

ALTER TABLE domain_tasks RENAME TO _domain_tasks_old;

CREATE TABLE domain_tasks (
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
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(source_compiled_plan_id) REFERENCES workflow_compiled_plans(id) ON DELETE CASCADE,
  FOREIGN KEY(source_compiled_task_id) REFERENCES workflow_compiled_tasks(id) ON DELETE CASCADE
);

INSERT INTO domain_tasks SELECT * FROM _domain_tasks_old;

DROP TABLE _domain_tasks_old;

COMMIT;

PRAGMA foreign_keys = ON;
