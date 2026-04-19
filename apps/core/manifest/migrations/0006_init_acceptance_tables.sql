CREATE TABLE IF NOT EXISTS acceptance_runs (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  task_id TEXT,
  profile_version TEXT NOT NULL,
  status TEXT NOT NULL,
  functional_status TEXT NOT NULL,
  production_status TEXT NOT NULL,
  manual_release_required INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  finished_at TEXT,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(task_id) REFERENCES domain_tasks(id)
);

CREATE TABLE IF NOT EXISTS acceptance_issues (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  severity TEXT NOT NULL,
  issue_kind TEXT NOT NULL,
  blocking INTEGER NOT NULL DEFAULT 0,
  summary TEXT NOT NULL,
  detail_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS acceptance_judgements (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  judgement_kind TEXT NOT NULL,
  judgement_result TEXT NOT NULL,
  summary TEXT NOT NULL,
  detail_json TEXT,
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS acceptance_surface_coverage (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  surface TEXT NOT NULL,
  coverage_status TEXT NOT NULL,
  evidence_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE,
  UNIQUE(acceptance_run_id, surface)
);

CREATE TABLE IF NOT EXISTS acceptance_journey_coverage (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_run_id TEXT NOT NULL,
  journey TEXT NOT NULL,
  coverage_status TEXT NOT NULL,
  evidence_count INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_run_id) REFERENCES acceptance_runs(id) ON DELETE CASCADE,
  UNIQUE(acceptance_run_id, journey)
);
