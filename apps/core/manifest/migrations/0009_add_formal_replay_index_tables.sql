CREATE TABLE IF NOT EXISTS workflow_replay_index (
  id TEXT PRIMARY KEY,
  replay_id TEXT NOT NULL,
  project_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  domain_task_id TEXT,
  compiled_task_id TEXT,
  event_id TEXT,
  trace_id TEXT,
  span_id TEXT,
  replay_kind TEXT NOT NULL,
  seq_no INTEGER NOT NULL,
  title TEXT NOT NULL,
  summary TEXT,
  file_path TEXT NOT NULL,
  file_ext TEXT,
  mime_type TEXT,
  file_size INTEGER,
  sha256 TEXT,
  source_object_kind TEXT,
  source_object_id TEXT,
  status TEXT NOT NULL DEFAULT 'available',
  created_at TEXT NOT NULL,
  updated_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(replay_id)
);

CREATE TABLE IF NOT EXISTS workflow_run_log_segments (
  id TEXT PRIMARY KEY,
  project_id TEXT NOT NULL,
  run_id TEXT NOT NULL,
  segment_id TEXT NOT NULL,
  stream_kind TEXT NOT NULL,
  seq_no INTEGER NOT NULL,
  file_path TEXT NOT NULL,
  file_size INTEGER,
  sha256 TEXT,
  started_at TEXT,
  ended_at TEXT,
  status TEXT NOT NULL DEFAULT 'available',
  created_at TEXT NOT NULL,
  FOREIGN KEY(project_id) REFERENCES projects(id) ON DELETE CASCADE,
  UNIQUE(run_id, stream_kind, seq_no)
);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_run_seq
  ON workflow_replay_index(run_id, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_project_created
  ON workflow_replay_index(project_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_event_id
  ON workflow_replay_index(event_id);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_trace_span
  ON workflow_replay_index(trace_id, span_id);

CREATE INDEX IF NOT EXISTS idx_workflow_replay_index_run_kind_seq
  ON workflow_replay_index(run_id, replay_kind, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_run_log_segments_run_stream_seq
  ON workflow_run_log_segments(run_id, stream_kind, seq_no);

CREATE INDEX IF NOT EXISTS idx_workflow_run_log_segments_project_run_created
  ON workflow_run_log_segments(project_id, run_id, created_at DESC);
