UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"auto"}'
WHERE `family_code` = 'coding'
  AND (`verification_profile_json` IS NULL OR `verification_profile_json` = '');

UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"local"}'
WHERE `family_code` = 'analysis'
  AND (`verification_profile_json` IS NULL OR `verification_profile_json` = '');

UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"local"}'
WHERE `family_code` = 'creative'
  AND (`verification_profile_json` IS NULL OR `verification_profile_json` = '');
