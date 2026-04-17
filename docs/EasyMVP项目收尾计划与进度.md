# EasyMVP项目收尾计划与进度

> 更新时间：2026-04-17
>
> 目的：记录当前收尾状态、验证约束、文档治理结果与下一批次排期，避免在长链路推进中丢失上下文。
>
> 约束：所有测试与编译统一只走 GitHub Actions；不允许在宿主机、本地开发机或 AI 会话中直接执行 `go test`、`go build`、`pnpm build`、`pnpm exec vite build`、`pnpm exec vue-tsc`、`npm/pnpm test`、`docker build`。仓库内 `scripts/web-antd-*-safe.sh` 仅保留为历史受控脚本资产，不再作为人工执行入口。

> 历史说明：本文档中 2026-04-13 之前出现的 `go test`、`validate.sh`、`web-antd-*-safe.sh`、`mode:auto/local/docker*` 等记录均属于历史证据或迁移轨迹，不代表当前执行入口；现行口径只认 GitHub Actions workflow run、日志、artifact 与 `.easymvp/ci/latest.json`。

## 2026-04-11 暂停说明

本轮新增了一条更细的暂停文档，专门记录“禁止直接 DB、统一 repo 化”这条主线目前哪些已经做完、哪些还没做完：

- 详见 [EasyMVP直连DB收口暂停说明](./EasyMVP直连DB收口暂停说明.md)

恢复这条主线时，优先按该文档里的“推荐恢复顺序”和“恢复工作前的检查项”执行，不要只根据对话上下文继续。

## 1. 当前阶段目标（2026-04-12 重排）

原始 5 项收尾目标已经完成，当前阶段不再继续把已落地主线重新列为“待做项”。截至当前，阶段目标已收缩为：

1. 收口 `web-antd` 在 `1 core / 1G` 限制下的验证失败点
2. 继续维护 `docs/` 主入口、专项基线和验证记录口径一致

## 1.1 当前文档主入口

- [README](./README.md)：文档索引与状态说明
- [EasyMVP项目收尾计划与进度](./EasyMVP项目收尾计划与进度.md)：当前收尾、验证约束与剩余风险主入口
- [EasyMVP研发执行版](./EasyMVP研发执行版.md)：当前研发排期与近期任务主入口
- [系统机制类问题彻底根治计划](./系统机制类问题彻底根治计划.md)、[Watchdog双保险实施计划](./Watchdog双保险实施计划.md)、[EasyMVP直连DB收口暂停说明](./EasyMVP直连DB收口暂停说明.md)：已完成基线与防回退入口

## 2. 当前状态快照

### 2.1 工作流

- 状态：代码主链收尾完成；当前阶段收尾已完成
- 范围：`workflow` 控制器拆分、验收闭环、workspace 交付元数据、轨迹/回放/评测视图
- 当前已知问题：
  - 主链后端已通过测试，前端按约束完成静态联动收口
  - 系统机制专项文档中的 `4.9 / 4.10 / 4.15 / 直连 DB 收口` 已补齐，当前不再存在后端主链级待根治项
  - `objective / situation / dashboard / execution / review / accept / verification / autonomy` 8 个工作流页面已在专项验证链路下完成验证
  - `web-antd` 已在 `1 core / 1G` 限制下通过 full typecheck/build 与专项验证

### 2.2 进度总览

| 序号 | 任务 | 状态 | 备注 |
|------|------|------|------|
| 1 | 文档索引与主入口整理 | 已完成 | `README`、本文档、`研发执行版` 与 `路线图` 已按当前状态重排 |
| 2 | 当前阶段计划重排 | 已完成 | 已把已完成主线从“本周待启动”口径中剥离 |
| 3 | `web-antd` 构建/类型检查验证 | 已完成 | full typecheck/build 与专项验证已在 GitHub Actions `Web Antd Guard` run `24579056821` 下全部通过 |
| 4 | `validate.sh` / `regressioncheck` 守卫验证 | 已完成 | 守卫逻辑已转入 GitHub Actions `backend-guard`；历史本机通过记录仅保留为证据 |
| 5 | 剩余专题能力收口 | 已完成 | `experience_reviewer` 已补默认预设、system-check readiness 与专题文档回写 |

### 2.3 当前剩余执行顺序（2026-04-12）

当前剩余项不再按“大而泛”的主线描述，而按真正阻塞最终完成的顺序执行：

1. 先处理 `web-antd` 在 `1 core / 1G` 限制下的验证失败
   - 当前唯一权威入口是 [web-antd-guard.yml](../.github/workflows/web-antd-guard.yml)
   - `scripts/web-antd-*-safe.sh` 仅保留为历史受控脚本资产，不再作为人工执行入口
   - GitHub Actions 产物需统一回写/同步为 `.easymvp/ci/latest.json`，供验证阶段读取
   - 历史 OOM / `143` 失败已被收敛为最新 guard 方案，不再作为当前阻塞结论
   - 当前已确认 `verify-build / workflow-bundle / entry-bundles` 三条专项验证与 `full typecheck/build` 可在同一条 GitHub Actions 链路下共同通过
   - 当前 guard 链路已固定支持 `EASYMVP_WEB_ANTD_VERIFY_BUILD=1`、`EASYMVP_WEB_ANTD_WORKFLOW_BUNDLE=1`、`EASYMVP_WEB_ANTD_BUNDLE_ENTRY=<entry>` 与 full guard 专用瘦入口
   - 后续若继续维护这条硬限制，只允许继续通过 GitHub Actions 调整 guard 图谱，不恢复本机执行
2. 同步维护文档与验证记录
   - 若 `web-antd-guard` 在 GitHub Actions 内完成复跑，需把结果同步回写到本文档、`README` 与 `研发执行版`
   - 继续保持“测试与编译只认 GitHub Actions”这条工程铁律，不恢复本机执行口径

### 2.3.1 2026-04-18 最新 GitHub Actions 结果

- `Web Antd Guard` 最新权威 run：`24579056821`
- 本次 run 已真实执行当前计划口径中的：
  - `full-typecheck`
  - `full-build`
  - `verify-build`
  - `workflow-bundle`
  - `entry-bundles`
- 当前真实结论已固定：
  - `full typecheck` 通过
  - `full production build` 通过
  - `verify-build / workflow-bundle / entry-bundles` 通过
- 本次 run 已成功产出 `web-antd-guard-ci-latest` artifact，本地 `.easymvp/ci/latest.json` 也已按 artifact 回写，不再依赖人工从 run 元数据补写
- `Backend Guard` 最新权威 run：`24574470208`
- 本次 backend run 结论：`validate-regression / test-backend / test-codegen / build-services` 全部通过，并已成功产出 `backend-guard-ci-latest` artifact

### 2.4 当前状态判定

- `workflow / acceptance / workspace / provider routing` 主链：已完成
- 工作流控制台静态联动收口：已完成
- 文档主入口与索引整理：已完成
- 当前阶段全部收尾：已完成

当前阶段已经可以宣称“全部做完”，因为：

1. `web-antd` 已对“当前工作区代码 + 当前计划口径下 full typecheck/build”形成最新 GitHub Actions 权威通过结论
2. `.easymvp/ci/latest.json` 与主文档已完成同步回写

### 2.5 中断恢复执行清单（2026-04-17）

为避免对话、会话或执行链路中断后再次回到“还剩什么没做完”的模糊状态，当前恢复顺序固定如下；恢复时优先看本节，不再重新从历史批次倒推。

#### 2.5.1 项目层剩余主阻塞

当前项目层主阻塞已清零：

1. `web-antd` 已在 run `24579056821` 下形成“当前工作区代码 + 当前计划口径下 full typecheck/build”的 GitHub Actions 通过结果

恢复执行时按以下顺序推进：

1. 继续只使用 `.github/workflows/web-antd-guard.yml` 作为权威验证入口
2. 后续若 guard 口径有变化，先修改 workflow/脚本，再通过 GitHub Actions 复跑验证
3. 每次复跑后都保留 `web-antd-guard-ci-latest` artifact
4. 将最新 GitHub Actions 结果同步回写到 `.easymvp/ci/latest.json`
5. 按最新结果同步更新 `README`、本文档和 `EasyMVP研发执行版`
6. 若未来再次出现失败，再从该通过基线继续收敛，而不是回退到本机执行

判定规则固定为：

1. full `typecheck/build` 在受限条件下通过，则当前阶段可判定完成
2. 若未来再次失败，则继续保持“当前阶段未完成”，并把失败原因、约束条件、替代验证路径回写到文档，禁止口头宣称“差不多完成”

#### 2.5.2 当前工作区未提交改动的收口计划

截至 2026-04-17，上一轮后端改动已经收口并进入 GitHub Actions 验证完成状态。已完成项包括：

1. `acceptance/evidence_collector*`
   - 已支持从当前目录、repo root、主仓 worktree 根回退查找 `.easymvp/ci/latest.json`
2. `verification/service*`
   - 已复用同类回退读取逻辑，并补齐 GitHub Actions fallback step / runner type 归一化
3. `qualitygate/standard*`
   - 已在缺少 browser automation 时放宽 browser verification 强制要求
4. `backend-guard`
   - 已修复 `publish-ci-result`、日志目录前置准备、集成测试 DB 条件化与最新 `latest.json` 生成链路
   - 最新通过 run：`24574470208`

#### 2.5.3 推荐执行顺序

恢复后不要并行发散，按以下顺序推进：

1. 先确认是否存在新的代码或 workflow 变更需要验证
2. 再触发 GitHub Actions 获取最新受限验证结果
3. 再回写 `.easymvp/ci/latest.json`
4. 再同步文档主入口
5. 最后重新判定“当前阶段全部完成”是否成立

#### 2.5.4 完成条件

当前这一轮已经满足以下两条，因此对外口径可改成“这一轮都做完了”：

1. 当前这批后端改动已完成收口，并具备对应 CI 验证证据
2. `web-antd` 在 GitHub Actions 下已经对“当前工作区代码 + 当前计划口径”形成最新通过证据，并同步回写到 `.easymvp/ci/latest.json` 与主文档

## 3. 执行策略与历史批次

1. 以编译和测试结果为主线，不靠人工猜测“可能完成”
2. 优先修复阻塞全量测试的低风险问题，再处理逻辑与联动问题
3. 每完成一个阶段，必须回写本文档中的状态和日志
4. 若发现新缺口，直接并入本文档，不另起临时 TODO

说明：

- `2.3` 和 `2.4` 是当前仍未完成的执行顺序与状态判定
- `3.1` 之后的条目主要用于追溯已完成的历史批次，不再代表当前待执行项

## 3.1 本轮追加批次（2026-04-09）

在上一批收尾完成后，继续按 `docs/EasyMVP研发执行版.md` 的剩余重点推进两项：

1. 补强 `C3`：把高风险交付、PR 草稿、人工回写等待等闸门信息补成更完整的审计与操作视图
2. 扩充回归样例：把 `test-workspaces/regression-manifest.json` 中仍为 `planned` 的样例继续落地

当前状态：

- 已完成

## 3.2 下一批次（2026-04-09）

继续按研发执行版推进仍未完全兑现的两项：

1. `A4`：补 execute / accept / rework / delivery review 关键路径自动回归
2. `A5`：补“样例一键回归”能力，至少先实现样例结构与口径的一键校验

当前状态：

- 已完成

## 3.3 下一批次（2026-04-09）

继续补回归与评测面板的可用性收口：

1. 把 `regression manifest` 校验结果正式接入后端接口返回
2. 把系统检查从“清单存在”升级到“清单存在且结构有效”
3. 把控制台“评测样例”面板补成可直接查看校验状态与 ready/planned 统计

当前状态：

- 已完成

## 3.4 下一批次（2026-04-09）

继续补阶段二 `B4` 的风险策略收口：

1. 给 `workspace delivery policy` 增加可注入测试入口
2. 补低/中/高风险交付矩阵自动回归
3. 把系统检查从“默认交付策略”升级到“风险交付矩阵”检查

当前状态：

- 已完成

## 3.5 下一批次（2026-04-09）

继续补“可接 CI / 运维守卫”的收口能力：

1. 把 `regressioncheck` 升级成同时校验样例清单和风险交付矩阵
2. 复用 `workspace` 风险矩阵校验逻辑，避免页面/脚本口径分叉
3. 更新 `validate.sh` 与说明文档，形成真正可执行的 guard 入口

当前状态：

- 已完成（代码、测试与说明文档已收口；`validate.sh` 按当前环境约束仅在隔离验证环境执行）

## 3.6 下一批次（2026-04-09）

继续补“计划文档与自动回归一致性”的收口能力：

1. 把 `EasyMVP研发执行版` 中阶段二、三已落地项正式回写为当前状态
2. 给 `B2` 的 CI 证据收集补纯逻辑自动回归
3. 继续保持“不运行 `web-antd` / 不在当前服务器执行 guard 脚本”的约束说明

当前状态：

- 已完成

## 3.7 下一批次（2026-04-09）

继续补“执行版逐项任务状态闭环”的最后一段：

1. 把 `A1 / A2 / A3` 的详细当前状态回写到执行版
2. 确认执行版 `A1-A5 / B1-B4 / C1-C4` 已全部具备明确状态
3. 继续保持当前环境约束说明，避免误执行 `web-antd` 与 guard 脚本

当前状态：

- 已完成

## 3.8 下一批次（2026-04-09）

继续补“控制器纯逻辑回归与时间线文案一致性”：

1. 给 `workflow_timeline / workflow_execution` 的纯逻辑辅助函数补更多测试
2. 修正 `workflow.force_stage` 时间线标签，避免丢失“人工切换”语义
3. 给审核问题回流文案补对称测试，和验收问题回流保持一致

当前状态：

- 已完成

## 3.9 下一批次（2026-04-09）

继续补“运行态聚合辅助函数”的纯逻辑自动回归：

1. 给 `workflow_runtime` 的快照选择逻辑补测试
2. 给任务统计与时间选择辅助函数补测试
3. 继续保持只做静态验证与 Go 测试，不触发受限运行项

当前状态：

- 已完成

## 3.10 下一批次（2026-04-09）

继续补 `workspace` 交付闸门纯逻辑回归：

1. 给交付事件载荷构造补测试
2. 给交付审核闸门开关与原因文案补测试
3. 给交付枚举归一化补测试

当前状态：

- 已完成

## 3.11 下一批次（2026-04-09）

继续补自治映射与事件 DTO 的纯逻辑回归：

1. 给 `workflow_autonomy` 的 DTO 映射函数补测试
2. 给 `buildTimelineEvent` 和 `jsonInt64SliceToInt64` 补测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.12 下一批次（2026-04-09）

继续补 `acceptance` 的 LLM Judge 纯逻辑回归：

1. 给 `buildJudgeUserPrompt` 补空输入与长证据摘要截断测试
2. 给 `parseJudgeResponse / validateJudgeResult` 补 JSON 解析与归一化测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.13 下一批次（2026-04-09）

继续补 `acceptance / workspace / runtime` 的兼容与边界纯逻辑回归：

1. 给证据摘要截断补测试
2. 给 `workspace` 的 legacy 列兼容过滤与隔离模式判断补测试
3. 给时间归一化的 nil/zero/future 边界补测试

当前状态：

- 已完成

## 3.14 下一批次（2026-04-09）

继续补 `acceptance rule_engine` 的文件规则纯逻辑回归：

1. 给必需文件检查补测试
2. 给必需扩展名检查补“命中/未命中/目录读取失败”测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.15 下一批次（2026-04-09）

继续补 `regression manifest / telegram` 的轻量纯逻辑回归：

1. 给 manifest 路径解析和目录/文件类型检查补边界测试
2. 给 Telegram 默认命令菜单补测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.16 下一批次（2026-04-09）

继续补简单包装函数与空分支的纯逻辑回归：

1. 给 `ResolveManifestPath` 包装入口补测试
2. 给 `NewJudge / NewRuleEngine / isUnknownColumnErr / 风险矩阵空摘要` 补测试
3. 再次验证当前可执行范围内是否只剩受限验证项

当前状态：

- 已完成

## 3.17 下一批次（2026-04-09）

继续补 `manifest` 校验器的错误口径边界回归：

1. 给 `LoadManifest` 的非法 JSON 与空 scenarios 归一化补测试
2. 给 `ValidateManifest` 的非法版本、非法日期、空场景、非法状态、ready 缺检查点等分支补测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.18 下一批次（2026-04-09）

继续补 `regressioncheck / workspace` 的路径包装逻辑回归：

1. 给 `resolveManifestPath` 的显式参数清理与自动探测补测试
2. 给 `resolveMainWorkDir` 的 worktree 反推补测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.19 下一批次（2026-04-09）

继续补 `regressioncheck` 的包装层与退出路径回归：

1. 给 `fail` 的 stderr 输出与退出码补子进程测试
2. 给 `main` 的成功输出路径补 helper 子进程测试
3. 继续只做静态验证与 Go 测试

当前状态：

- 已完成

## 3.20 下一批次（2026-04-09）

继续做文档整理与过期文档清理：

1. 统一 `README.md` 与 `docs/README.md` 的文档入口
2. 删除已被执行版/收尾进度吸收的一次性联调记录文档
3. 修正文档间的引用，避免继续指向已删除文档

当前状态：

- 已完成

## 3.21 下一批次（2026-04-09）

继续做 Workflow V2 真实创建链路验证与专项记录：

1. 直接调用 `POST /api/mvp/workflow/create-project` 验证真实创建
2. 补查 `project-status / timeline / stage-history / project-trace`
3. 把请求、响应、问题、修复和复测落到 `docs/workflow-v2-create-verify/`
4. 修复验证中发现的真实问题并补回归测试

当前状态：

- 已完成

## 3.22 下一批次（2026-04-09）

继续做 Workflow V2 从创建到完成的真实主链验证与专项归档：

1. 直接驱动 `create-project -> parse-tasks -> confirm-plan -> manual-approve -> execute -> accept -> complete`
2. 记录首次全链路运行暴露出的真实问题
3. 修复问题并补针对性回归测试
4. 把请求、响应、问题、修复与复测结果追加到 `docs/workflow-v2-create-verify/`

当前状态：

- 已完成

## 3.23 下一批次（2026-04-10）

继续补“完成后可用性验证”的低配服务器保护与旁路收口：

1. 给 Docker-first 验证步骤统一注入命令资源限制
2. 给 Codex / Claude / Gemini / Aider / OpenHands 执行器统一注入命令资源限制
3. 给 `app/ai` 运行时旁路补同口径资源限制，避免绕过 `workflow` 入口
4. 把资源限制配置、系统检查和专项说明文档同步回写

当前状态：

- 已完成

## 3.24 下一批次（2026-04-10）

继续把“项目分类验证配置”落到 EasyMVP 主系统：

1. 为 `mvp_project_category` 增加 `verification_profile_json / verification_gate_json`
2. 让验证链支持“项目级配置 -> 分类级配置 -> 自动探测”的优先级
3. 把分类级 gate 固化到验证结果判定和证据快照
4. 回写分类管理前端、数据库 migration 与系统内复验结果

当前状态：

- 已完成

## 4. 进度日志

### 2026-04-09 进度记录

- 已确认当前未提交改动集中在：
  - `admin-go/app/mvp/internal/controller/chat`
  - `admin-go/app/mvp/internal/workflow/acceptance`
  - `admin-go/app/mvp/internal/workspace`
  - `admin-go/utility/provider`
  - `vue-vben-admin/apps/web-antd/src/views/mvp/workflow`
- 已读取并对照：
  - `docs/EasyMVP研发执行版.md`
  - `docs/README.md`
  - `README.md`
- 已启动后端验证：
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/...`
  - `cd admin-go && go test ./app/ai/...`
- 后端验证结果：
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/...` 通过
  - `cd admin-go && go test ./app/ai/...` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 前端静态收尾结果：
  - 已核对 `workflow` 控制台新增接口与页面联动
  - 已修复工作流页面在 `projectId / workflowRunId` 切换时未重新加载数据的问题
  - 已修复执行控制台把 `stageStatus` 误按 `stageType` 着色的问题
- 文档回写结果：
  - `docs/README.md` 已补入 [EasyMVP项目收尾计划与进度](EasyMVP项目收尾计划与进度.md)，并修正文档计数
- 本轮第二批追加结果：
  - 已补 execute / accept / rework / delivery review 关键路径自动回归
  - 已新增 `admin-go/app/mvp/internal/regression/manifest.go` 样例清单校验器
  - 已新增 `admin-go/app/mvp/regressioncheck` 与 `test-workspaces/validate.sh` 一键校验入口
  - `workflow_regression.go` 与 `workflow_system_check.go` 已统一复用回归样例清单解析入口
- 本轮第二批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/regression ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workflow/stage/rework` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
  - `bash ./test-workspaces/validate.sh` 通过，输出 `scenarios=4 ready=4 planned=0`
- 本轮第三批追加结果：
- 本轮第十二批追加结果：
  - 已完成 Workflow V2 `create-project` 真实接口验证，并创建专项记录目录 `docs/workflow-v2-create-verify/`
  - 已发现并修复 `stage-history` 未做 UTC 本地化导致的 8 小时时间漂移
  - 已同步修正 `ReworkStatus` 的相关时间字段，避免返工链路重复出现同类问题
  - 已新增 `buildStageHistoryItem` 时间归一化回归测试
- 本轮第十二批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat` 通过
  - `POST /api/mvp/workflow/create-project` 成功返回 `projectID / conversationID / workflowRunID`
  - `GET /api/mvp/workflow/project-status` 返回 `workflowStatus=designing`、`currentStage=design`
  - `GET /api/mvp/workflow/stage-history` 与 `GET /api/mvp/workflow/project-trace` 时间口径已一致，复测样例均为 `2026-04-09 15:15:43`
  - `workflow/regression-scenarios` 已返回 `valid / readyCount / plannedCount / message`
  - 系统检查已从“样例清单存在”升级为“样例清单存在且校验通过”
  - 控制台“评测样例”面板已补校验状态与 ready/planned 统计的静态展示
- 本轮第二十一批追加结果：
  - 已完成 Workflow V2 从 `create-project` 到 `completed` 的真实全链路验证
  - 首次真实运行项目 `317591057413967872` 暴露了 `review-status` 时间未归一化、`skip-task` 时间二次偏移、`completion-summary` 统计异常、`accept` 时间线标签缺失 4 个问题
  - 已修复 `workflow_review.go`、`workflow.go`、`workflow_timeline.go`、`stage/complete/service.go` 的相关实现
  - 已补 `workflow_time_test.go` 与 `stage/complete/service_test.go` 针对性回归测试
  - 已把全链路请求、问题样例、修复后复测样例落到 `docs/workflow-v2-create-verify/`
- 本轮第二十一批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/stage/complete` 通过
  - 复测项目 `317593202297147392` 已完成全链路流转，`workflowStatus=completed`、`currentStage=complete`
  - `review-status` 的 `stageTasks.startedAt` 已恢复为本地时间 `2026-04-09 15:35:20`
  - `skip-task` 后任务 `completedAt` 已恢复为 `2026-04-09 15:35:22`
  - `completion-summary` 已返回 `skippedTasks=2`、`avgTaskDuration=1s`
  - `timeline` 已返回 `验收阶段已启动`
- 本轮第三批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/regression ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workspace` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
  - `bash ./test-workspaces/validate.sh` 通过，输出 `scenarios=4 ready=4 planned=0`
- 本轮第四批追加结果：
  - 已为 `workspace delivery policy` 增加可注入配置入口，补齐风险矩阵自动回归
  - 系统检查已新增 low/medium/high 风险交付矩阵摘要与异常配置告警
  - `B4` 所需的“中高风险不直写主结果”约束已具备代码、测试和检查三重收口
- 本轮第四批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workspace ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
  - `bash ./test-workspaces/validate.sh` 通过，输出 `scenarios=4 ready=4 planned=0`
- 本轮第五批追加结果：
  - 已将 `admin-go/app/mvp/regressioncheck/main.go` 重构为可单测入口，guard 失败信息可区分 manifest 与风险交付矩阵
  - 已新增 `admin-go/app/mvp/regressioncheck/main_test.go`，补齐 guard 主入口成功/manifest 失败/风险矩阵失败三类用例
  - `test-workspaces/README.md` 与 `docs/EasyMVP研发执行版.md` 已补 guard 脚本执行限制，避免在业务服务器直接运行
  - `3.5` 状态已回写为完成，并明确当前服务器不直接执行 `validate.sh`
- 本轮第五批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/regression ./app/mvp/internal/workspace ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
  - 未执行 `bash ./test-workspaces/validate.sh`，遵循当前服务器约束
- 本轮第六批追加结果：
  - `docs/EasyMVP研发执行版.md` 已补齐 `B1 / B2 / B3 / C1 / C2 / C3 / C4` 的当前状态，避免执行版与实际代码脱节
  - 已新增 `admin-go/app/mvp/internal/workflow/acceptance/evidence_collector_test.go`，补齐 CI 证据文件探测、JSON 摘要和日志识别回归
  - 本轮文档继续明确未运行 `web-antd`、未在当前服务器执行 `validate.sh`
- 本轮第六批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第七批追加结果：
  - `docs/EasyMVP研发执行版.md` 已补齐 `A1 / A2 / A3` 的详细当前状态
  - 执行版现已覆盖 `A1-A5 / B1-B4 / C1-C4` 全部任务的完成状态，不再只靠章节摘要判断
  - 当前约束已继续保留：未运行 `web-antd`，未在当前服务器执行 `validate.sh`
- 本轮第七批验证结果：
  - 本批仅涉及文档回写，未新增代码改动，沿用第六批后端验证结果
- 本轮第八批追加结果：
  - 已补 `workflow_time_test.go`，覆盖 `domainTaskErrorMessage`、异常资源解析和 `formatTimelineLabel`
  - 已修正 `workflow.force_stage` 时间线标签为“工作流已人工切换到X阶段”，避免事件语义丢失
  - `docs/EasyMVP研发执行版.md` 已把 `C1 / C2 / C3` 拆成逐项状态，不再挂在 `C4` 小节下
  - 已新增 `admin-go/app/mvp/internal/controller/chat/workflow_review_test.go`，补齐审核问题回流文案测试
- 本轮第八批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/acceptance` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第九批追加结果：
  - 已新增 `admin-go/app/mvp/internal/controller/chat/workflow_runtime_test.go`
  - 已补 `shouldUseRuntimeSnapshot / taskStatFromProgress / latestNonNilTime` 三个运行态辅助函数测试
- 本轮第九批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十批追加结果：
  - 已新增 `admin-go/app/mvp/internal/workspace/delivery_event_test.go`
  - 已补交付事件载荷、交付审核闸门、交付审核原因和枚举归一化测试
- 本轮第十批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workspace` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十一批追加结果：
  - 已新增 `admin-go/app/mvp/internal/controller/chat/workflow_autonomy_test.go`
  - 已新增 `admin-go/app/mvp/internal/controller/chat/workflow_event_test.go`
  - 已补自治决策/检查点/规则 DTO 映射，以及时间线事件与 ID 切片转换测试
- 本轮第十一批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十二批追加结果：
  - 已新增 `admin-go/app/mvp/internal/workflow/acceptance/judge_test.go`
  - 已补 `buildJudgeUserPrompt` 的空输入兜底、规则命中格式化和长证据摘要截断测试
  - 已补 `parseJudgeResponse / validateJudgeResult` 的原始 JSON、代码块 JSON、非法内容和结论归一化测试
- 本轮第十二批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workflow/acceptance` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十三批追加结果：
  - 已补 `trimSummary` 的截断与空白裁剪测试
  - 已新增 `admin-go/app/mvp/internal/workspace/manager_test.go` 与 `workspace_repo_test.go`
  - 已补 `NeedsIsolation`、`isDeliveryReferenceColumnErr`、`filterWorkspaceDataByError` 以及 `normalizeDBUTCGTime / isRecentGTime` 的边界测试
- 本轮第十三批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workspace ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十四批追加结果：
  - 已补 `checkRequiredFiles` 的存在/缺失/越界路径忽略测试
  - 已补 `checkRequiredExtensions` 的扩展名命中、缺失命中和目录读取失败告警测试
  - `acceptance` 的文件型规则在当前允许范围内已具备直接自动回归
- 本轮第十四批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/workflow/acceptance` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十五批追加结果：
  - 已补 `ResolveManifestPathFromCWD` 未命中、`resolveWorkspacePath` 越界/绝对路径保护，以及 `requireDir / requireRegularFile` 的类型检查测试
  - 已新增 `admin-go/app/mvp/internal/controller/chat/workflow_telegram_test.go`
  - 已补 `defaultTelegramCommands` 默认菜单顺序与文案测试
- 本轮第十五批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/regression ./app/mvp/internal/controller/chat` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十六批追加结果：
  - 已补 `ResolveManifestPath` 当前目录包装入口测试
  - 已补 `NewJudge / NewRuleEngine / isUnknownColumnErr / RiskDeliveryPolicyReport.Summary(empty)` 测试
  - 当前允许范围内已未再发现新的高价值纯逻辑测试缺口，剩余项收敛为受限验证
- 本轮第十六批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/regression ./app/mvp/internal/workspace ./app/mvp/internal/workflow/acceptance` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十七批追加结果：
  - 已补 `LoadManifest` 的非法 JSON 失败口径与 `scenarios == nil` 归一化测试
  - 已补 `ValidateManifest` 的非法 version / updatedAt / scenarios、非法字段、ready 缺 checkpoints 和空工作区分支测试
  - `regression manifest` 的结构与路径校验在当前允许范围内已进一步收口
- 本轮第十七批验证结果：
  - `cd admin-go && go test ./app/mvp/internal/regression` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十八批追加结果：
  - 已补 `regressioncheck.resolveManifestPath` 的显式路径清理与当前目录自动探测测试
  - 已补 `workspace.resolveMainWorkDir` 的 worktree 路径反推测试
  - `regressioncheck` 与 `workspace` 剩余低成本路径包装分支已进一步收口
- 本轮第十八批验证结果：
  - `cd admin-go && go test ./app/mvp/regressioncheck ./app/mvp/internal/workspace` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第十九批追加结果：
  - 已补 `regressioncheck.fail` 的 stderr 输出与退出码子进程测试
  - 已补 `regressioncheck.main` 的成功输出 helper 子进程测试，并在 helper 内覆写校验依赖以避开配置读取
  - `regressioncheck` 包装层在当前允许范围内已进一步收口
- 本轮第十九批验证结果：
  - `cd admin-go && go test ./app/mvp/regressioncheck` 通过
  - `cd admin-go && go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过
- 本轮第二十批追加结果：
  - 已整理 `README.md` 与 `docs/README.md` 的文档入口与说明文字
  - 已把路线图、执行版中的“一次性联调记录”引用替换为执行版/收尾进度文档
  - 已清理 `GitWorktree任务级环境隔离设计文档.md` 中残留的旧实施计划与方案评审段落，改为当前实现口径
  - 已删除过期文档 `docs/WorkflowV2主链后端联调问题记录.md`
- 本轮第二十批验证结果：
  - 已执行文档引用静态检查，确认仓库内不再引用已删除文档
  - 本批未触发 `web-antd` 或 `validate.sh`

### 2026-04-09 最终状态

- 后端主链：
  - 控制器拆分、验收闭环、workspace 交付元数据、provider 协议路由相关代码已通过当前包测试
- 前端主链：
  - `dashboard / execution / review / accept / autonomy / rework / timeline / regression / objective / situation / meta-cognition` 已完成当前改动范围内的静态联动收口
  - 未执行 `web-antd` 构建或运行，遵循本轮约束
- 本轮追加批次：
  - 已新增 `/workflow/delivery-reviews` 明细接口，补齐交付闸门“聚合 + 明细”视图
  - 时间线页已可直接查看待人工交付 / PR 草稿 / 待回写 / 高风险任务明细
  - `test-workspaces/regression-manifest.json` 已升级到 `version = 2`
  - `readme_refresh / multi_task_dependency / manual_takeover` 三个回归样例已补齐规格目录并从 `planned` 调整为 `ready`
- 本轮第二批追加：
  - 已补 `accept / rework / delivery review` 关键纯逻辑测试，并沿用 `executor` 既有自动回归覆盖 execute 侧
  - 已补验收问题转返工原因构造测试，收口 `accept issue -> rework` 文案链路
  - 已补 `regression manifest` 一键校验器和脚本，ready 样例现可一键做结构与口径校验
- 本轮第三批追加：
  - 回归样例接口已显式返回校验摘要，系统检查与评测面板不再只看文件是否存在
  - 评测样例面板已可直接查看 manifest 校验通过/失败和 ready/planned 数量
- 本轮第四批追加：
  - 风险交付矩阵已进入系统检查，`low / medium / high` 策略漂移会直接告警
  - `workspace delivery policy` 已有默认矩阵与配置覆写回归，防止中高风险退化为自动直写
- 本轮第五批追加：
  - `regressioncheck` 已能同时输出 manifest 与风险矩阵摘要，且主入口具备单元测试覆盖
  - guard 文档已明确要求在隔离验证环境执行，避免在业务服务器直接运行
- 本轮第六批追加：
  - 执行版文档已补齐阶段二、三的已落地状态，计划与代码现已对齐
  - `B2` 的 CI 证据采集已具备纯逻辑自动回归，降低“页面可见但缺自动验证”的风险
- 本轮第七批追加：
  - 执行版文档已补齐阶段一 `A1 / A2 / A3` 的详细状态，现有计划项已全部具备显式完成标记
- 本轮第八批追加：
  - `workflow.force_stage` 时间线标签已改为保留“人工切换”语义
  - 控制器纯逻辑测试已继续补齐到时间线、任务错误消息和审核问题回流文案
- 本轮第九批追加：
  - 运行态聚合辅助函数已补纯逻辑测试，当前控制器侧常用纯逻辑函数已基本覆盖
- 本轮第十批追加：
  - `workspace` 交付事件载荷、审核闸门、原因文案和枚举归一化已补纯逻辑测试
- 本轮第十一批追加：
  - 自治 DTO 映射、时间线事件和 ID 切片转换已补纯逻辑测试
- 本轮第十二批追加：
  - `acceptance` 的 LLM Judge prompt/response 纯逻辑已补自动回归，验收闭环在当前允许范围内进一步收口
- 本轮第十三批追加：
  - `acceptance / workspace / runtime` 的兼容与边界纯逻辑已继续补齐，legacy 列兼容和时间边界场景已有自动回归保护
- 本轮第十四批追加：
  - `acceptance rule_engine` 的文件规则已补纯逻辑测试，验收规则侧剩余未覆盖点主要收敛到数据库依赖路径
- 本轮第十五批追加：
  - `regression manifest` 的路径安全与文件类型判断已具备更完整的自动回归，控制台默认 Telegram 菜单也已纳入测试保护
- 本轮第十六批追加：
  - 简单包装函数与空摘要分支已补自动回归，代码侧可继续静态推进的低成本缺口基本收空
- 本轮第十七批追加：
  - `manifest` 校验器的错误口径边界已补齐更多自动回归，当前可静态推进的剩余空间进一步压缩
- 本轮第十八批追加：
  - `regressioncheck / workspace` 的路径包装逻辑已补自动回归，当前剩余可静态推进项继续收缩
- 本轮第十九批追加：
  - `regressioncheck` 的 helper 包装层与失败路径已补自动回归，当前可继续静态推进的包装分支更少
- 本轮第二十批追加：
  - 文档索引与规划文档已完成一次整理，过期的一次性联调文档已从工作树移除
- 本轮第二十一批追加：
  - Workflow V2 已完成一次从创建项目到 `completed` 的真实后端全链路验证，发现的时间口径与汇总统计问题已修复并复测通过
- 本轮第二十二批追加：
  - 已新增统一命令资源限制策略，默认限制 `npm/pnpm/go/turbo/compose` 类命令的内存与并发
  - Docker-first 验证步骤已统一注入资源限制，避免 `pnpm build`、`npm install` 等命令在低配服务器上失控
  - `workflow` 执行器与 `app/ai` 运行时旁路已统一复用同一套资源限制，避免从其他入口绕过保护
  - 系统检查与专项文档已补充“命令资源限制”说明，当前配置可通过 `mvp_config` 或 `config.yaml` 覆盖
- 本轮第二十三批追加：
  - 已修正 `npm` / `pnpm` 资源限制的注入方式，避免将 `pnpm` 专属参数错误下发给 `npm`
  - Linux 环境下的非 Node 安装/构建/验证命令已额外通过 `prlimit` 增加进程级地址空间上限；Node 侧保留 `NODE_OPTIONS + npm/pnpm 并发限制`，避免前端验证出现误杀
- 本轮第二十四批追加：
  - 已为 `mvp_project_category` 增加 `verification_profile_json / verification_gate_json`，并补齐分类管理后端模型、接口、导入导出与前端表单展示
  - 验证服务已支持“项目级 `.easymvp/verification.json` -> 分类级默认 profile -> 自动探测”的优先级，并在结果判定时执行分类级 gate
  - 分类级 gate 当前支持 `allowedDecisions / minExecutedSteps / requiredCheckKinds / allowedRunnerTypes`
  - MySQL migration 已升级到 `8`，默认给 `software_dev / game_dev / data_analysis / creative` 系列写入首批 gate
  - 已通过 EasyMVP 系统重新触发贪吃蛇项目验证：`verificationRunID=317748542812721152`，结果 `passed`
  - 最新验证证据已明确记录 `gateSource=category:software_dev`，说明分类级验证规则已经进入系统内真实验收链
- 本轮第二十五批追加：
  - 已新增分类验证覆盖率系统检查项，直接统计已启用分类的 `verification_profile_json / verification_gate_json` 配置覆盖情况
  - 已为分类管理列表补充“默认验证模板 / 放行规则”状态列，便于直接发现未配置项
  - 已新增 migration `9`，按 `family_code` 自动回填遗漏分类的默认 gate，避免 `product_design` 一类分类处于空白状态
- 本轮第二十六批追加：
  - 已新增 migration `10`，为 `coding / analysis / creative` 家族回填首批分类默认 `verification_profile_json`
  - `coding` 家族默认 profile 为 `{"mode":"auto"}`，保留 Docker-first 自动探测
  - `analysis / creative` 家族默认 profile 为 `{"mode":"local"}`，直接走本机验证 / 人工复核路径
  - 注：以上为历史回填记录；当前已由 migration `12` 统一归一到 `{"mode":"github_actions"}`
  - migration 已执行完成，当前系统检查已达到 `gate 9/9, profile 9/9`
  - 已通过 EasyMVP 系统再次触发贪吃蛇项目验证：`verificationRunID=317752601615536128`，结果 `passed`
  - 最新验证证据已明确记录 `profileSource=category:software_dev` 与 `gateSource=category:software_dev`，说明分类级 profile 与 gate 都已进入真实验收链
- 本轮第二十七批追加：
  - 已补齐项目级硬约束 contract 下钻链路：`architect plan -> review precheck -> domain task instantiation -> objective_json merge` 现在消费同一份结构化约束
  - 已补齐 `objective_guard` 的任务级预算判定，`retry / rework` 不再错误复用 workflow 总量压死后续新任务
  - 已重新执行高档位后端验证：`go test ./app/mvp/... ./utility/provider/... ./app/ai/...`
  - 当前后端主链、provider 路由与 `app/ai` 相关包均通过编译/测试，收尾文档里的后端验证口径已与真实代码状态重新对齐
- 本轮第二十八批追加：
  - 已继续提升验证强度，执行 `cd admin-go && go test ./...`
  - `app/mvp / app/ai / app/system / utility/*` 当前 Go 包均通过整仓测试
  - 当前后端剩余风险已进一步压缩为环境约束项，不再是 Go 代码主链或包级回归项
- 本轮第二十九批追加：
  - 已修正 `/workflow/situation-history` 的返回口径，控制器现在会把 `snapshot_data` 解包为前端实际消费的 `progress / health / resource / trend` 结构，不再直接回传原始快照表行
  - `/workflow/situation` 已支持可选 `taskID`，可返回带任务焦点的自治预算视角；`/workflow/situation-history` 也支持按 `taskID` 过滤同任务焦点快照
  - `web-antd` 的态势页已完成静态联动收口：支持 `taskId` 路由参数，并同时展示 workflow 总量与当前任务的 `retry / rework` 预算
  - 已新增控制器纯逻辑回归，覆盖快照解包、坏 JSON 兜底和任务过滤逻辑；`cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/autonomy` 与 `go test ./app/mvp/internal/workflow/... ./app/mvp/internal/controller/chat` 通过
- 本轮第三十批追加：
  - `workflow/objective` 接口已把 `technicalContract` 一并返回，控制台可直接看到当前项目识别出的 `required / forbidden technologies`
  - `web-antd` 的目标层页面已补只读“项目级硬约束”区块，并明确提示“保存目标层参数不会覆盖硬约束”
  - 已执行 `cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/contract ./app/mvp/internal/workflow/autonomy`，通过
- 本轮第三十一批追加：
  - `web-antd` 执行控制台已新增“态势”入口，能直接带 `projectId / workflowRunId / taskId` 跳到任务焦点态势页
  - 任务级预算从“后端内部可用”提升到“控制台可直接进入并查看”，前端静态联动又收掉一处手工拼参数缺口
- 本轮第三十二批追加：
  - 已在上述 `situation / objective / execution` 联动补丁后再次执行 `cd admin-go && go test ./...`
  - 最新整仓 Go 测试仍为全绿，说明本轮新增的接口扩展、快照解包与控制台静态联动没有带出后端回归
- 本轮第三十三批追加：
  - 已收紧 `/workflow/situation-history` 的过滤语义：未传 `taskID` 时只返回 workflow 总量快照，避免把任务焦点快照混入通用历史视图
  - 已补充控制器纯逻辑回归，覆盖“通用历史排除任务焦点快照”的边界
  - `web-antd` 的目标层页面已把 `technicalContract` 从表单态与保存载荷中剥离，避免只读硬约束字段被一并回传到写接口
  - 已再次执行 `cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/autonomy` 与 `cd admin-go && go test ./...`，结果继续全绿
- 本轮第三十四批追加：
  - `web-antd` 的态势页已把“刷新态势”收敛为统一 `refreshDashboard` 入口，按钮刷新与路由切换都会同时拉取当前态势和快照历史
  - 任务焦点模式下不再出现“上方态势已刷新、下方快照历史仍停留旧数据”的前端静态联动缺口
  - 本批未执行 `web-antd` 构建或类型检查，仍保持当前服务器只做静态修正的约束；后端基线已通过 `cd admin-go && go test ./app/mvp/...`
- 本轮第三十五批追加：
  - 已修正 `/workflow/situation-history` 的分页过滤缺口：现在会分页拉取快照窗口并在攒够目标结果前持续过滤，不再出现“项目最新快照很多时，任务焦点历史被无关快照挤掉”的误空结果
  - `workflow_run_id` 过滤已下推到 `SituationSnapshotRepo.ListByProjectIDWindow(...)`，控制器侧只保留 `taskID` 焦点过滤与快照解包
  - `workflow/situation` 在传入 `taskID` 时不再先做一遍多余的 workflow 全量感知，避免一次请求里重复计算两套态势
  - 已新增控制器纯逻辑回归，覆盖“跨页后才命中目标任务快照”的场景；`cd admin-go && go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/repo ./app/mvp/internal/workflow/autonomy` 与 `cd admin-go && go test ./app/mvp/...` 通过
- 本轮第三十六批追加：
  - `project-status` 接口已补充返回 `workflowRunID`，工作流仪表板不再需要靠其他页面间接推断当前运行实例
  - 仪表板“快速导航”已新增“目标约束 / 态势仪表板”入口；进入态势页时会自动带上 `projectId + workflowRunId`
  - `web-antd` 的态势页已统一用 `refreshLoading` 驱动主区块加载状态，并且缺少 `workflowRunId` 时不再误拉项目级快照历史
  - 本批继续保持“仅做前端静态联动、不跑 `web-antd` 构建”的约束；后端已再次执行 `cd admin-go && go test ./app/mvp/...`，结果全绿
- 本轮第三十七批追加：
  - 工作流仪表板中的 `重试 / 跳过 / 全部重试 / 全部跳过` 现在都会触发整页 `loadAll()` 刷新，不再只更新失败任务列表
  - 顶部总任务、失败数、阶段状态和进度条不再在人工介入后短时间停留旧值，仪表板的统计口径已和操作结果保持同步
  - 项目切换时会先重置旧的仪表板状态，再加载新项目数据，避免切页过程中短暂显示上一个项目的残留信息
  - 本批为纯前端静态联动修正，未执行 `web-antd` 构建或类型检查；静态检查 `git diff --check` 通过
- 本轮第三十八批追加：
  - 嵌入式 `execution / review / accept / verification` 子面板已统一补上 `changed` 事件，人工动作完成后会通知父级工作流仪表板刷新顶部统计和阶段状态
  - `review / verification` 在项目切换时也会先清空旧数据，再加载新项目内容，避免短时间显示上一项目的审核/验证残留
  - 这样工作流仪表板中的子面板动作和父级概览已经形成一致的刷新口径，不再出现“子页已变更、父级仍显示旧状态”的分裂
  - 本批仍为纯前端静态联动修正，未执行 `web-antd` 构建或类型检查；静态检查 `git diff --check` 通过
- 本轮第三十九批追加：
  - `autonomy` 子面板中的 `approve / reject / triggerReplan` 也已纳入统一 `changed` 事件链路，父级工作流仪表板会同步刷新
  - 到当前为止，仪表板内所有会直接改变工作流状态或阶段推进的子面板动作，都已具备“子页动作完成 -> 父级概览同步刷新”的一致行为
  - `rework / timeline / regression` 仍保持只读展示，不需要额外的父级刷新回调
  - 本批仍为纯前端静态联动修正，未执行 `web-antd` 构建或类型检查；静态检查 `git diff --check` 通过
- 本轮第四十批追加：
  - `review / accept / verification / autonomy` 在切换项目时，除了清主数据外，也会清空筛选条件、弹窗开关、驳回/返工原因、已选问题等临时 UI 状态
  - 新项目页面不再继承上一个项目残留的 severity filter、驳回理由、返工弹窗或报告详情，跨项目状态隔离更完整
  - 这批修正不改变业务接口，仅收紧控制台局部状态管理；仍未执行 `web-antd` 构建或类型检查
  - 静态检查 `git diff --check` 通过
- 本轮第四十一批追加：
  - `dashboard / execution / review / accept / verification / autonomy / situation` 已统一补上请求版本保护，切换项目或路由后，旧请求回包不再反写新页面状态
  - 这批修正重点收口“异步请求后返回覆盖当前项目数据”的前端竞态问题，避免控制台在快速切换项目时短时间闪回旧数据
  - `dashboard` 的失败任务列表也已并入同一版本保护口径，不再出现主状态已切到新项目、失败任务仍被旧请求回填的分裂
  - 本批仍为纯前端静态联动修正，未执行 `web-antd` 构建或类型检查；静态检查 `git diff --check` 通过
- 本轮第四十二批追加：
  - 已完成 `docs/README.md`、`EasyMVP项目收尾计划与进度.md`、`EasyMVP研发执行版.md`、`EasyMVP全面分析与优化路线图.md` 的统一整理，明确“主入口 / 已完成基线 / 专题规划 / 验证记录”四类口径
  - 当前计划已从“补第一轮主链实现”切换为“补隔离环境验证、继续文档治理、收缩剩余专题能力”
  - 不再把 `workflow` 控制器拆分、provider 协议路由、worktree 回写策略等已完成主线重新列为近期待启动任务
  - 本批为纯文档整理；静态检查 `git diff --check` 通过
- 本轮第四十三批追加：
  - 已修正本文档内部“已完成 / 待执行”并存的状态冲突，明确拆成“代码主链收尾完成”和“当前阶段全部完成”两个判定层级
  - 已新增“当前剩余执行顺序”，把收尾顺序固定为：`web-antd` 隔离验证 -> `validate.sh / regressioncheck` 守卫验证 -> `experience_reviewer` 专题收口
  - 已把最终未完成项压缩为三个明确阻塞，不再用泛化的“还有一些收尾工作”表述
  - 本批为纯文档修正；静态检查 `git diff --check` 通过
- 本轮第四十四批追加：
  - 已把第 3 节口径收敛为“执行策略与历史批次”，避免继续把旧批次记录误读为当前待执行事项
  - 已明确：当前真正未完成的内容只看 `2.3 当前剩余执行顺序` 和 `2.4 当前状态判定`
  - 本批为纯文档修正；静态检查 `git diff --check` 通过
- 本轮第四十五批追加：
  - 已新增 [scripts/web-antd-build-safe.sh](../scripts/web-antd-build-safe.sh)，把 `web-antd` 生产构建也纳入与类型检查一致的受控执行入口
  - 当前 `web-antd` 的构建/类型检查口径已统一为：`NODE_OPTIONS` 堆上限、`pnpm child/workspace concurrency=1`、`nice/ionice`、内存门槛检查、单实例锁
  - 已回写 `EasyMVP工程铁律`、`workflow-verification-docker` 和本文档，明确不允许直接裸跑高负载 `pnpm build`
  - 本批为脚本与文档修正；静态检查 `git diff --check` 通过
- 本轮第四十六批追加：
  - 已修正 `test-workspaces/validate.sh` 的 GoFrame 配置入口，默认注入 `GF_GCFG_PATH=app/mvp/manifest/config` 与 `GF_GCFG_FILE=config.yaml`
  - 已补 `admin-go/app/mvp/regressioncheck/main.go` 的 MySQL driver blank import，避免 `g.DB()` 初始化时缺驱动崩溃
  - 现已通过：`bash ./test-workspaces/validate.sh`、`cd admin-go && go test ./app/mvp/regressioncheck`
  - `validate.sh / regressioncheck` 这条守卫验证已从“待执行”收口为“已完成”
- 本轮第四十七批追加：
  - 已按受控脚本试跑 `web-antd` 构建/类型检查；当前实机可用内存 `1909MB`，低于脚本默认门槛 `2048MB`
  - 两条脚本均按设计自动跳过，没有裸跑高负载 `pnpm` 命令，也没有把宿主机继续压满
  - `web-antd` 这条剩余项当前不是“缺脚本”，而是“等待更空闲窗口或隔离验证环境”
- 本轮第四十八批追加：
  - 已补 `admin-go/manifest/sql/seed/mysql_seed.sql` 的 `software_dev / game_dev -> experience_reviewer/max` 默认预设，统一绑定启用模型 `315100000000000007`
  - 已补 `workflow_system_check` 的体验评审师 readiness 检查，会校验默认预设存在、`model_id` 非 0、模型记录存在且启用
  - 已补 `workflow_system_check_test` 的存在 / 缺失 / 模型失效断言，并回写 [EasyMVP体验评审师接入方案与计划](./EasyMVP体验评审师接入方案与计划.md)
  - `experience_reviewer` 专题已从“当前剩余阻塞”收口为“已完成专题”
- 本轮第四十九批追加：
  - 已把 `web-antd` 两条受控脚本升级为硬限制模式：`systemd-run --scope + AllowedCPUs=0 + CPUQuota=100% + MemoryMax=1G + MemorySwapMax=0`
  - 已完成首轮实跑：`./scripts/web-antd-typecheck-safe.sh` 在 `heap=768MB` 下触发 V8 OOM；保持 `1G cgroup`、把 heap 提到 `896MB` 后会被 scope 终止
  - 已完成首轮构建实跑：`./scripts/web-antd-build-safe.sh` 在 `transforming...` 阶段被 `1G` 限制终止，退出 `143`
  - 当前 `web-antd` 的剩余问题已从“未验证”切换为“已验证，但当前 full typecheck/build 无法在 1 core / 1G 下通过”
- 本轮第五十批追加：
  - 已新增 [scripts/web-antd-entry-typecheck-safe.sh](../scripts/web-antd-entry-typecheck-safe.sh) 的正式批量入口 [scripts/web-antd-workflow-pages-typecheck-safe.sh](../scripts/web-antd-workflow-pages-typecheck-safe.sh)
  - 已在 `1 core / 1G` 硬限制下顺序通过：`objective / situation / dashboard / execution / review / accept / verification / autonomy`
  - 当前 `web-antd` 的验证口径已从“只有 full typecheck/build”扩展为“full 验证失败 + workflow 页面可拆分验证”
- 本轮第五十一批追加：
  - 已新增 [scripts/web-antd-verify-build-safe.sh](../scripts/web-antd-verify-build-safe.sh)，用于在同样 `1 core / 1G` 限制下触发 `web-antd` 轻量验证构建
  - 已补 `vue-vben-admin/apps/web-antd/vite.config.ts` 的 `EASYMVP_WEB_ANTD_VERIFY_BUILD=1` 静态开关，会关闭 `archiver / extraAppConfig / html / i18n / injectAppLoading / injectMetadata / license`
  - 当前这批为纯静态收口，遵循“禁止编译”约束，尚未重新执行 `web-antd` 构建验证
- 本轮第五十二批追加：
  - 已新增 [scripts/web-antd-workflow-bundle-safe.sh](../scripts/web-antd-workflow-bundle-safe.sh)，用于在同样 `1 core / 1G` 限制下触发 `workflow` 最小 bundle 验证
  - 已新增 [src/verify/workflow-bundle.ts](../vue-vben-admin/apps/web-antd/src/verify/workflow-bundle.ts)，会把 `workflow` 目录下的 `vue/ts/tsx` 模块与 `api/mvp/workflow` 纳入单独构建入口
  - 已补 `vue-vben-admin/apps/web-antd/vite.config.ts` 的 `EASYMVP_WEB_ANTD_WORKFLOW_BUNDLE=1` 开关；当前这批仍为纯静态收口，未执行构建
- 本轮第五十三批追加：
  - 已新增 [scripts/web-antd-entry-bundle-safe.sh](../scripts/web-antd-entry-bundle-safe.sh)，用于在同样 `1 core / 1G` 限制下按单入口触发页面级 bundle 验证
  - 已新增 [scripts/web-antd-workflow-entry-bundles-safe.sh](../scripts/web-antd-workflow-entry-bundles-safe.sh)，用于串行执行 `objective / situation / dashboard / execution / review / accept / verification / autonomy` 的页面级 bundle 验证
  - 已补 `vue-vben-admin/apps/web-antd/vite.config.ts` 的 `EASYMVP_WEB_ANTD_BUNDLE_ENTRY` / `EASYMVP_WEB_ANTD_BUNDLE_OUT_DIR` 开关，使轻量验证构建可以进一步收窄到单页面入口
  - 同步修正了本文档第 `5.2` 节的旧完成定义残留；当前这批仍为纯静态收口，未执行构建
- 本轮第五十四批追加：
  - 已补 [README.md](./README.md) 的“当前状态”区块，把 `web-antd` 受控验证脚本矩阵、`1 core / 1G` 资源约束和禁止裸跑口径补回主索引
  - 本文档第 `5.2` 节要求的 `README` 同步口径现已收口，不再只在 `研发执行版 / 工程铁律 / 验证说明` 中单边存在
  - 当前这批仍为纯文档治理，未执行 `pnpm` / `vite build` / `vue-tsc`
- 本轮第五十五批追加：
  - 已继续收紧 `vue-vben-admin/apps/web-antd/vite.config.ts` 的轻量验证模式：除关闭附加插件外，再禁用 `minify / cssMinify / reportCompressedSize / modulePreload / treeshake`
  - `scripts/web-antd-verify-build-safe.sh`、`scripts/web-antd-workflow-bundle-safe.sh`、`scripts/web-antd-entry-bundle-safe.sh` 与 `scripts/web-antd-workflow-entry-bundles-safe.sh` 会自动复用这一层更轻的 Vite 构建口径
  - 当前这批仍为纯静态压峰值，未执行构建
- 当前剩余风险：
  - 当前批次主阻塞已清零；若后续 `web-antd` 入口图或 guard 口径继续变化，必须以 run `24579056821` 为通过基线继续复跑验证

## 5. 完成定义

### 5.1 代码主链收尾完成

满足以下条件后，可判定“代码主链收尾完成”：

1. 后端相关包通过编译，关键测试可执行
2. 前端工作流页面与接口完成静态联动核对
3. 文档主入口、专项基线与进度日志已能反映真实代码状态
4. 后端主链剩余风险已压缩到环境约束或专题能力，而非主流程缺口

当前状态：

- 已满足

### 5.2 当前阶段全部完成

满足以下条件后，本文档顶部状态才能改为“当前阶段收尾完成”：

1. `web-antd` 在 `1 core / 1G` 限制下至少完成一轮可追溯的受控验证，并把成功或失败结果回写到本文档
2. 当前受限验证入口、资源约束和禁止裸跑口径已同步更新到 `README`、`研发执行版`、`工程铁律` 与相关验证说明
3. 剩余未完成项已不再是当前阶段阻塞，或者已被正式移交到后续独立专题文档

当前状态：

- 未满足
