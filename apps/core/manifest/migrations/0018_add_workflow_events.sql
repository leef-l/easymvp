-- Migration 0018: Add workflow_events table for event-driven architecture
-- This enables closed-loop automation: review rejected -> auto redesign ->
-- re-review -> compile -> execute -> acceptance -> repair -> auto re-execution.

CREATE TABLE IF NOT EXISTS workflow_events (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    event_type TEXT NOT NULL,
    payload_json TEXT NOT NULL DEFAULT '{}',
    status TEXT NOT NULL DEFAULT 'pending',
    retry_count INTEGER NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TEXT NOT NULL,
    processed_at TEXT,
    next_retry_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_workflow_events_status ON workflow_events(status);
CREATE INDEX IF NOT EXISTS idx_workflow_events_project_id ON workflow_events(project_id);
CREATE INDEX IF NOT EXISTS idx_workflow_events_type_status ON workflow_events(event_type, status);
CREATE INDEX IF NOT EXISTS idx_workflow_events_next_retry ON workflow_events(next_retry_at);

-- Migration 0018 part 2: Add repair_history table for tracking repair attempts
CREATE TABLE IF NOT EXISTS repair_plan_history (
    id TEXT PRIMARY KEY,
    repair_plan_draft_id TEXT NOT NULL,
    project_id TEXT NOT NULL,
    reason_class TEXT,
    repair_strategy TEXT,
    human_checkpoint_required INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'created',
    created_at TEXT NOT NULL,
    resolved_at TEXT
);

CREATE INDEX IF NOT EXISTS idx_repair_history_project ON repair_plan_history(project_id);
CREATE INDEX IF NOT EXISTS idx_repair_history_draft ON repair_plan_history(repair_plan_draft_id);
