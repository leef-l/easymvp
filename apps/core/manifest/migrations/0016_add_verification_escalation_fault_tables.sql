-- Migration 0016: Add independent persistence tables for VerificationResult, RuntimeEscalation, and FaultSummary
-- These tables store snapshot records per acceptance run, enabling historical audit and offline analysis
-- of the three sub-views that were previously only computed dynamically in Go.

CREATE TABLE IF NOT EXISTS verification_results (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    acceptance_run_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'pending',
    decision TEXT,
    completed INTEGER NOT NULL DEFAULT 0,
    summary TEXT,
    preferred_channel TEXT,
    required_checks_json TEXT,
    required_evidence_json TEXT,
    missing_evidence_json TEXT,
    failed_checks_json TEXT,
    verification_contract_json TEXT,
    source_run_id TEXT,
    channel_available INTEGER NOT NULL DEFAULT 1,
    environment_available INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS runtime_escalations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    acceptance_run_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'none',
    reason_class TEXT,
    source_brain TEXT,
    source_task_id TEXT,
    run_binding_id TEXT,
    run_status TEXT,
    severity TEXT,
    action TEXT,
    task_id TEXT,
    run_id TEXT,
    summary TEXT,
    policy_denied INTEGER NOT NULL DEFAULT 0,
    evidence_refs_json TEXT,
    resolved_at TEXT,
    resolution_status TEXT,
    resolver_kind TEXT,
    linked_fault_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS fault_summaries (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    acceptance_run_id TEXT NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'clean',
    blocking_issue_count INTEGER NOT NULL DEFAULT 0,
    advisory_issue_count INTEGER NOT NULL DEFAULT 0,
    top_issue TEXT,
    fault_loop_detected INTEGER NOT NULL DEFAULT 0,
    fault_kind TEXT,
    severity TEXT,
    summary TEXT,
    failed_checks_json TEXT,
    affected_tasks_json TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_verification_results_project ON verification_results(project_id);
CREATE INDEX IF NOT EXISTS idx_verification_results_run ON verification_results(acceptance_run_id);
CREATE INDEX IF NOT EXISTS idx_runtime_escalations_project ON runtime_escalations(project_id);
CREATE INDEX IF NOT EXISTS idx_runtime_escalations_run ON runtime_escalations(acceptance_run_id);
CREATE INDEX IF NOT EXISTS idx_fault_summaries_project ON fault_summaries(project_id);
CREATE INDEX IF NOT EXISTS idx_fault_summaries_run ON fault_summaries(acceptance_run_id);
