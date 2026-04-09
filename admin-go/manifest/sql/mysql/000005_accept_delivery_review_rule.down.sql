DELETE FROM `mvp_accept_rule`
WHERE `rule_code` = 'software.delivery_review_required'
  AND `project_type` IN ('software_dev', 'coding');
