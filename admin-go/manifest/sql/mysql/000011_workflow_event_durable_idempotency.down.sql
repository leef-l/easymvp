DROP TABLE IF EXISTS `mvp_workflow_event_ledger`;

ALTER TABLE `mvp_workflow_event`
  DROP INDEX `uk_workflow_event_event_id`,
  DROP INDEX `uk_workflow_event_idempotency_key`,
  DROP COLUMN `event_id`,
  DROP COLUMN `attempt`,
  DROP COLUMN `idempotency_key`;
