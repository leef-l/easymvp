SET @stmt = IF(
  (
    SELECT COUNT(1)
    FROM INFORMATION_SCHEMA.COLUMNS
    WHERE TABLE_SCHEMA = DATABASE()
      AND TABLE_NAME = 'mvp_project_category'
      AND COLUMN_NAME = 'verification_gate_json'
  ) > 0,
  'ALTER TABLE `mvp_project_category` DROP COLUMN `verification_gate_json`',
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
      AND COLUMN_NAME = 'verification_profile_json'
  ) > 0,
  'ALTER TABLE `mvp_project_category` DROP COLUMN `verification_profile_json`',
  'SELECT 1'
);
PREPARE stmt FROM @stmt;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
