-- Add executor_succeeded, delivery_verified, acceptance_passed to completion_verdicts
-- These three boolean flags distinguish the four layers of completion:
--   executor_succeeded = run completed
--   delivery_verified  = delivery contract satisfied
--   acceptance_passed  = acceptance criteria met
--   completed          = business closure (requires all above + manual release if required)

ALTER TABLE completion_verdicts
  ADD COLUMN executor_succeeded INTEGER NOT NULL DEFAULT 0;

ALTER TABLE completion_verdicts
  ADD COLUMN delivery_verified INTEGER NOT NULL DEFAULT 0;

ALTER TABLE completion_verdicts
  ADD COLUMN acceptance_passed INTEGER NOT NULL DEFAULT 0;
