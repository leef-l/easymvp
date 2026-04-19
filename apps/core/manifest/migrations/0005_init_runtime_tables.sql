CREATE TABLE IF NOT EXISTS brain_run_bindings (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT NOT NULL,
  brain_kind TEXT NOT NULL,
  brain_run_id TEXT NOT NULL,
  run_status TEXT NOT NULL,
  started_at TEXT NOT NULL,
  finished_at TEXT,
  last_sync_at TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id) ON DELETE CASCADE,
  UNIQUE(brain_run_id)
);

CREATE TABLE IF NOT EXISTS run_checkpoints (
  id TEXT PRIMARY KEY,
  run_binding_id TEXT NOT NULL,
  checkpoint_type TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(run_binding_id) REFERENCES brain_run_bindings(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS run_event_index (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_binding_id TEXT NOT NULL,
  sequence_no INTEGER NOT NULL,
  event_type TEXT NOT NULL,
  event_level TEXT,
  summary TEXT NOT NULL,
  payload_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(run_binding_id) REFERENCES brain_run_bindings(id) ON DELETE CASCADE,
  UNIQUE(run_binding_id, sequence_no)
);
