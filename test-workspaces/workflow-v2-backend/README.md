# 最小后端主链回归样例

场景编码：`workflow_v2_backend`

目标：

- 覆盖 `create-project -> review -> execute -> accept -> complete` 主链
- 保留一个可追踪的 repo 型回归样例骨架，供 `validate.sh` 与 manifest guard 校验

说明：

- 当前仓库内仅保留最小目录骨架，不在本轮提交真实嵌套 Git 仓库
- 需要做真实主链回归时，可在 `repo/` 下放入独立样例项目并由工作流创建 worktree
