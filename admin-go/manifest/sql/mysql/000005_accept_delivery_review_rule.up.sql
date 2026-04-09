INSERT INTO `mvp_accept_rule`
(`id`,`project_type`,`rule_code`,`rule_name`,`rule_type`,`scope_type`,`config_json`,`enabled`,`priority`,`created_at`,`updated_at`,`deleted_at`)
VALUES
(318090000000000001,'software_dev','software.delivery_review_required','高风险或待人工交付结果需要人工审核','process','task','{\"require_manual_review_delivery_modes\":[\"manual\",\"pr\"],\"require_manual_review_sync_statuses\":[\"pending\",\"failed\"],\"min_risk_level\":\"high\"}',1,25,'2026-04-09 00:00:00','2026-04-09 00:00:00',NULL),
(318090000000000002,'coding','software.delivery_review_required','高风险或待人工交付结果需要人工审核','process','task','{\"require_manual_review_delivery_modes\":[\"manual\",\"pr\"],\"require_manual_review_sync_statuses\":[\"pending\",\"failed\"],\"min_risk_level\":\"high\"}',1,25,'2026-04-09 00:00:00','2026-04-09 00:00:00',NULL)
ON DUPLICATE KEY UPDATE
`rule_name`=VALUES(`rule_name`),
`rule_type`=VALUES(`rule_type`),
`scope_type`=VALUES(`scope_type`),
`config_json`=VALUES(`config_json`),
`enabled`=VALUES(`enabled`),
`priority`=VALUES(`priority`),
`updated_at`=VALUES(`updated_at`),
`deleted_at`=NULL;
