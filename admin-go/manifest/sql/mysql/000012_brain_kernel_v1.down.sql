-- 回滚 BrainKernel v1 持久化层
-- 按依赖逆序 DROP（delta 依赖 plan，所以先 drop delta）

DROP TABLE IF EXISTS `mvp_artifact_meta`;
DROP TABLE IF EXISTS `mvp_brain_usage`;
DROP TABLE IF EXISTS `mvp_run_checkpoint`;
DROP TABLE IF EXISTS `mvp_brain_plan_delta`;
DROP TABLE IF EXISTS `mvp_brain_plan`;
