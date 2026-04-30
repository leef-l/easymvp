-- Migration 0021: reserved placeholder.
-- Original 0021 slot was skipped during MACCS Phase 1/2 development;
-- this no-op fills the gap so future automation can rely on contiguous
-- version numbers without renaming subsequent migrations (which would
-- break checksum validation against existing schema_migrations rows).

SELECT 1;
