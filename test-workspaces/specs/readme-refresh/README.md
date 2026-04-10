# README 修订回归样例

场景编码：`readme_refresh`

目标：

- 覆盖低风险 `patch + auto_apply` 路径
- 观察交付物、回写状态和验收证据是否闭环

建议输入：

1. 创建一个最小仓库，只包含 `README.md`
2. 让架构师拆出单任务：
   - 修改 README 标题
   - 增加一段使用说明
   - 不改动代码和依赖
3. 确认方案并推进到执行、验收

关键观察点：

1. `mvp_task_workspace.delivery_mode = patch`
2. `mvp_task_workspace.sync_strategy = auto_apply`
3. 低风险任务完成后 `sync_status = applied`
4. 执行控制台可见 `patch_ref / diff_summary`
5. 验收证据中可看到 `diff / delivery / task_output`

备注：

- 当前样例固定口径，不要求在本轮自动执行
- 如果后续补脚本，应保持 `scenarioCode` 不变
