CREATE TABLE IF NOT EXISTS projects (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL,
  project_category TEXT NOT NULL,
  goal_summary TEXT NOT NULL,
  status TEXT NOT NULL,
  production_status TEXT NOT NULL,
  workspace_root TEXT NOT NULL,
  repo_root TEXT,
  current_plan_draft_id TEXT,
  current_compiled_plan_id TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS project_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  category_profile_version TEXT NOT NULL,
  acceptance_profile_version TEXT NOT NULL,
  role_profile_version TEXT,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS project_workspaces (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  workspace_root TEXT NOT NULL,
  evidence_root TEXT NOT NULL,
  runs_root TEXT NOT NULL,
  replay_root TEXT NOT NULL,
  diagnostics_root TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);
