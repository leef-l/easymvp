SET @stmt = IF(
  (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'mvp_project_category'
      AND COLUMN_NAME = 'verification_profile_json'
  ) = 0,
  'ALTER TABLE `mvp_project_category` ADD COLUMN `verification_profile_json` longtext COLLATE utf8mb4_unicode_ci COMMENT ''分类默认验证配置(JSON)'' AFTER `description`',
  'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

SET @stmt = IF(
  (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'mvp_project_category'
      AND COLUMN_NAME = 'verification_gate_json'
  ) = 0,
  'ALTER TABLE `mvp_project_category` ADD COLUMN `verification_gate_json` longtext COLLATE utf8mb4_unicode_ci COMMENT ''分类验证放行规则(JSON)'' AFTER `verification_profile_json`',
  'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed"],"minExecutedSteps":1,"requiredCheckKinds":["test"]}'
WHERE `category_code` = 'software_dev'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed"],"minExecutedSteps":2,"requiredCheckKinds":["test","build"]}'
WHERE `category_code` = 'game_dev'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed","manual_review"],"minExecutedSteps":0}'
WHERE `category_code` = 'data_analysis'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed","manual_review"],"minExecutedSteps":0}'
WHERE `family_code` = 'creative'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');
