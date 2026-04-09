UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed"],"minExecutedSteps":1,"requiredCheckKinds":["test"]}'
WHERE `family_code` = 'coding'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed","manual_review"],"minExecutedSteps":0}'
WHERE `family_code` = 'analysis'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');

UPDATE `mvp_project_category`
SET `verification_gate_json` = '{"allowedDecisions":["passed","manual_review"],"minExecutedSteps":0}'
WHERE `family_code` = 'creative'
  AND (`verification_gate_json` IS NULL OR `verification_gate_json` = '');
