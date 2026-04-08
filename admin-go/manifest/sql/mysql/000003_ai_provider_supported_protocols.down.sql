SET @has_supported_protocols := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ai_provider'
    AND COLUMN_NAME = 'supported_protocols'
);
SET @ddl := IF(
  @has_supported_protocols = 1,
  'ALTER TABLE `ai_provider` DROP COLUMN `supported_protocols`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;
