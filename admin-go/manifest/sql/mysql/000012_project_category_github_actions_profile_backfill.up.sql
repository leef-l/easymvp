UPDATE `mvp_project_category`
SET `verification_profile_json` = '{"mode":"github_actions"}'
WHERE `verification_profile_json` IN (
  '{"mode":"auto"}',
  '{"mode":"local"}',
  '{"mode":"docker_compose"}',
  '{"mode":"dockerfile"}'
);
