# 回归样例集

本目录用于沉淀 EasyMVP 工作流回归样例。

当前环境不做验证执行，但样例清单和预期口径需要固定，避免后续每轮回归临时拼接输入。

配套清单见：

- [regression-manifest.json](./regression-manifest.json)
- [validate.sh](./validate.sh)（仅供 GitHub Actions 调用）

使用原则：

1. 每个样例都要有稳定的 `scenarioCode`
2. 每个样例都要明确目标阶段和关键观察点
3. 结构变更通过 migration 交付，样例只负责输入与期望，不承载 schema 初始化逻辑
4. 同一个样例目录既可以放真实仓库，也可以先放占位说明；是否可执行由 `status` 字段标记

一键校验：

- 统一通过 `.github/workflows/backend-guard.yml` 触发
- `validate.sh` 仅允许在 GitHub Actions 环境内执行，本机直接运行会被拒绝

执行约束：

1. `validate.sh` 会在 GitHub Actions 内执行 `go run ./app/mvp/regressioncheck`
2. 本机、业务服务器和 AI 会话不允许直接执行这条 guard 链
3. 成功输出会同时包含 `regression manifest OK` 与 `workspace delivery policies OK`

当前校验覆盖：

1. `regression-manifest.json` 版本、日期、状态、场景编码去重
2. `ready` 场景目录是否存在
3. `specs/*` 场景是否包含 `README.md / input.md / expected.md`
4. 非 `specs/*` 的 `ready` 场景是否至少具备 `README.md` 或 `repo/` 目录
5. `workspace delivery` 的 low / medium / high 风险交付矩阵是否保持安全配置
