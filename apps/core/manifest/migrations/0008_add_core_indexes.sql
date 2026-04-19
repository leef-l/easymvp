CREATE INDEX IF NOT EXISTS idx_projects_status_updated_at
  ON projects(status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_projects_category_status
  ON projects(project_category, status);
CREATE INDEX IF NOT EXISTS idx_projects_production_status_updated_at
  ON projects(production_status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_plan_drafts_project_version
  ON workflow_plan_drafts(project_id, version DESC);
CREATE INDEX IF NOT EXISTS idx_plan_reviews_plan_draft
  ON workflow_plan_review_results(plan_draft_id, review_version DESC);
CREATE INDEX IF NOT EXISTS idx_compiled_plans_project_version
  ON workflow_compiled_plans(project_id, compiled_version DESC);
CREATE INDEX IF NOT EXISTS idx_compiled_tasks_plan_phase
  ON workflow_compiled_tasks(compiled_plan_id, phase);
CREATE INDEX IF NOT EXISTS idx_domain_tasks_project_phase_status
  ON domain_tasks(project_id, phase, status);
CREATE INDEX IF NOT EXISTS idx_domain_tasks_project_updated
  ON domain_tasks(project_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_run_bindings_project_status
  ON brain_run_bindings(project_id, run_status, started_at DESC);
CREATE INDEX IF NOT EXISTS idx_run_events_run_seq
  ON run_event_index(run_binding_id, sequence_no);
CREATE INDEX IF NOT EXISTS idx_acceptance_runs_project_status
  ON acceptance_runs(project_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_acceptance_issues_run_severity
  ON acceptance_issues(acceptance_run_id, severity);
CREATE INDEX IF NOT EXISTS idx_evidence_project_captured
  ON evidence_items(project_id, captured_at DESC);
CREATE INDEX IF NOT EXISTS idx_evidence_surface_journey
  ON evidence_items(project_id, surface, journey);
CREATE INDEX IF NOT EXISTS idx_replay_project_time
  ON replay_items(project_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_audit_project_time
  ON audit_logs(project_id, created_at DESC);
