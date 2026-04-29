-- Migration 0019: Add browser_run_id and verifier_run_id to acceptance_runs
-- so that the adjudication process can wait for and consume their results.

ALTER TABLE acceptance_runs ADD COLUMN browser_run_id TEXT;
ALTER TABLE acceptance_runs ADD COLUMN verifier_run_id TEXT;
ALTER TABLE acceptance_runs ADD COLUMN validation_results_json TEXT DEFAULT '{}';

CREATE INDEX IF NOT EXISTS idx_acceptance_runs_browser_run ON acceptance_runs(browser_run_id);
CREATE INDEX IF NOT EXISTS idx_acceptance_runs_verifier_run ON acceptance_runs(verifier_run_id);
