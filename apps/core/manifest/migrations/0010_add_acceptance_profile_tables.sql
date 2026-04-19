CREATE TABLE IF NOT EXISTS acceptance_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  project_category TEXT NOT NULL,
  profile_version TEXT NOT NULL,
  required_surfaces_json TEXT NOT NULL,
  required_journeys_json TEXT NOT NULL,
  required_evidence_json TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_acceptance_profiles_project_version
  ON acceptance_profiles(project_id, profile_version, updated_at DESC);

CREATE TABLE IF NOT EXISTS production_acceptance_profiles (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  acceptance_profile_id TEXT NOT NULL,
  profile_version TEXT NOT NULL,
  required_surfaces_json TEXT NOT NULL,
  required_journeys_json TEXT NOT NULL,
  required_evidence_json TEXT NOT NULL,
  release_requirements_json TEXT NOT NULL,
  status TEXT NOT NULL,
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  FOREIGN KEY(acceptance_profile_id) REFERENCES acceptance_profiles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_production_acceptance_profiles_project_version
  ON production_acceptance_profiles(project_id, profile_version, updated_at DESC);
