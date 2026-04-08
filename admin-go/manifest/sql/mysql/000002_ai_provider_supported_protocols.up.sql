ALTER TABLE `ai_provider`
  ADD COLUMN `supported_protocols` json DEFAULT NULL COMMENT '支持的协议类型(JSON)：anthropic/openai_compatible/google 等' AFTER `provider_type`;

UPDATE `ai_provider`
SET `supported_protocols` = JSON_ARRAY('anthropic', 'openai_compatible')
WHERE `code` = 'tencent_coding' OR `provider_type` = 'tencent_coding';

UPDATE `ai_provider`
SET `supported_protocols` = JSON_ARRAY(`provider_type`)
WHERE (`supported_protocols` IS NULL OR JSON_LENGTH(`supported_protocols`) = 0)
  AND `provider_type` IN ('anthropic', 'openai_compatible', 'google');
