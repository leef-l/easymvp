-- 回滚：删除数据权限索引
DROP INDEX idx_datascope ON mvp_project;
DROP INDEX idx_datascope ON mvp_workflow_run;
DROP INDEX idx_datascope ON mvp_conversation;
DROP INDEX idx_datascope ON mvp_domain_task;
DROP INDEX idx_datascope ON mvp_plan_version;
DROP INDEX idx_datascope ON mvp_stage_run;
DROP INDEX idx_datascope ON mvp_task_blueprint;
DROP INDEX idx_datascope ON mvp_project_role;
DROP INDEX idx_datascope ON mvp_project_report;

-- 回滚：删除软删除索引
DROP INDEX idx_deleted_at ON mvp_domain_task;
DROP INDEX idx_deleted_at ON mvp_stage_run;
DROP INDEX idx_deleted_at ON mvp_plan_version;
DROP INDEX idx_deleted_at ON mvp_task_blueprint;
DROP INDEX idx_deleted_at ON mvp_review_issue;
DROP INDEX idx_deleted_at ON mvp_autonomy_decision;
DROP INDEX idx_deleted_at ON mvp_workflow_run;

-- 回滚：删除业务查询索引
DROP INDEX idx_workflow_run ON mvp_review_issue;
DROP INDEX idx_project_role ON mvp_conversation;
