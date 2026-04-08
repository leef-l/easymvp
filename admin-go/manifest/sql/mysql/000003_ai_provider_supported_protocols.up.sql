SET @has_supported_protocols := (
  SELECT COUNT(*)
  FROM information_schema.COLUMNS
  WHERE TABLE_SCHEMA = DATABASE()
    AND TABLE_NAME = 'ai_provider'
    AND COLUMN_NAME = 'supported_protocols'
);
SET @ddl := IF(
  @has_supported_protocols = 0,
  'ALTER TABLE `ai_provider` ADD COLUMN `supported_protocols` json DEFAULT NULL COMMENT ''支持的协议类型(JSON)：anthropic/openai_compatible/google 等'' AFTER `provider_type`',
  'SELECT 1'
);
PREPARE stmt FROM @ddl;
EXECUTE stmt;
DEALLOCATE PREPARE stmt;

UPDATE `ai_provider`
SET `supported_protocols` = JSON_ARRAY('anthropic', 'openai_compatible')
WHERE `code` = 'tencent_coding' OR `provider_type` = 'tencent_coding';

UPDATE `ai_provider`
SET `supported_protocols` = JSON_ARRAY(`provider_type`)
WHERE (`supported_protocols` IS NULL OR JSON_LENGTH(`supported_protocols`) = 0)
  AND `provider_type` IN ('anthropic', 'openai_compatible', 'google');
