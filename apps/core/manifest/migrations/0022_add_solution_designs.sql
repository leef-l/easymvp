-- Migration 0022: Add solution_designs table for MACCS Phase 2 (solution design)

CREATE TABLE IF NOT EXISTS solution_designs (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    requirement_id TEXT NOT NULL DEFAULT '',
    version INTEGER NOT NULL DEFAULT 1,
    status TEXT NOT NULL DEFAULT 'draft', -- draft/designing/reviewing/approved/rejected
    architecture TEXT NOT NULL DEFAULT '',
    modules_json TEXT NOT NULL DEFAULT '[]',
    data_models_json TEXT NOT NULL DEFAULT '[]',
    pages_json TEXT NOT NULL DEFAULT '[]',
    task_drafts_json TEXT NOT NULL DEFAULT '[]',
    user_confirmed INTEGER NOT NULL DEFAULT 0,
    confirmed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_solution_designs_project_id ON solution_designs(project_id);
CREATE INDEX IF NOT EXISTS idx_solution_designs_status ON solution_designs(status);
