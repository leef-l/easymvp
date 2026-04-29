-- Migration 0017: Add conversation and message tables for AI architect chat

CREATE TABLE IF NOT EXISTS conversations (
    id TEXT PRIMARY KEY,
    project_id TEXT NOT NULL,
    conversation_kind TEXT NOT NULL DEFAULT 'architect',
    status TEXT NOT NULL DEFAULT 'active',
    plan_draft_id TEXT,
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_conversations_project_id ON conversations(project_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);

CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    conversation_id TEXT NOT NULL,
    sender_role TEXT NOT NULL,
    sender_name TEXT NOT NULL,
    content TEXT NOT NULL,
    message_kind TEXT NOT NULL DEFAULT 'chat',
    created_at TEXT NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at);
