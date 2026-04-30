-- Migration 0020: Add requirements table for MACCS Phase 1 (requirement understanding)

CREATE TABLE IF NOT EXISTS requirements (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL REFERENCES projects(id),
    raw_input TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'draft',
    requirement_doc_json TEXT NOT NULL DEFAULT '{}',
    user_confirmed INTEGER NOT NULL DEFAULT 0,
    confirmed_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_requirements_project_id ON requirements(project_id);
CREATE INDEX IF NOT EXISTS idx_requirements_status ON requirements(status);
