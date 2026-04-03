-- MVP 消息类型字段，区分普通对话、任务执行和毒药消息
ALTER TABLE `mvp_message`
ADD COLUMN IF NOT EXISTS `message_type` varchar(30) NOT NULL DEFAULT 'general' COMMENT '消息类型：chat_user/chat_reply/task_prompt/task_reply/system_notice/poison/general' AFTER `role`;

ALTER TABLE `mvp_message`
ADD KEY `idx_message_type` (`message_type`);

UPDATE `mvp_message` m
LEFT JOIN `mvp_conversation` c ON c.id = m.conversation_id
SET m.message_type = CASE
  WHEN m.status = 'failed' THEN 'poison'
  WHEN m.role = 'system' THEN 'system_notice'
  WHEN (c.task_id IS NULL OR c.task_id = 0) AND m.role = 'user' THEN 'chat_user'
  WHEN (c.task_id IS NULL OR c.task_id = 0) AND m.role = 'assistant' THEN 'chat_reply'
  WHEN c.task_id IS NOT NULL AND c.task_id <> 0 AND m.role = 'user' THEN 'task_prompt'
  WHEN c.task_id IS NOT NULL AND c.task_id <> 0 AND m.role = 'assistant' THEN 'task_reply'
  ELSE 'general'
END
WHERE m.message_type IS NULL OR m.message_type = '' OR m.message_type = 'general';
