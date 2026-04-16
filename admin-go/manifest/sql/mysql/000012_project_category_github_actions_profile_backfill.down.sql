UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"auto"}'
WHERE `family_code` = 'coding'
  AND `verification_profile_json` = '{"mode":"github_actions"}';

UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"local"}'
WHERE `family_code` IN ('analysis', 'creative')
  AND `verification_profile_json` = '{"mode":"github_actions"}';
