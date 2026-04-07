-- 数据权限索引 idx_datascope (dept_id, created_by)
-- 支撑 ApplyDataScope 五级数据权限过滤查询
CREATE INDEX idx_datascope ON mvp_project (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_workflow_run (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_conversation (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_domain_task (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_plan_version (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_stage_run (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_task_blueprint (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_project_role (dept_id, created_by);
CREATE INDEX idx_datascope ON mvp_project_report (dept_id, created_by);

-- 软删除索引 idx_deleted_at
-- 几乎所有列表查询都带 WHERE deleted_at IS NULL
CREATE INDEX idx_deleted_at ON mvp_domain_task (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_stage_run (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_plan_version (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_task_blueprint (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_review_issue (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_autonomy_decision (deleted_at);
CREATE INDEX idx_deleted_at ON mvp_workflow_run (deleted_at);

-- 业务查询索引
CREATE INDEX idx_workflow_run ON mvp_review_issue (workflow_run_id);
CREATE INDEX idx_project_role ON mvp_conversation (project_id, role_type(20));
