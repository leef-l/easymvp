# EasyMVP项目收尾计划与进度

> 更新时间：2026-04-10
>
> 目的：把当前工作树中已经启动但尚未完全收口的主线能力持续推进到可编译、可验证、可交付状态。该文档在本轮收尾过程中持续更新，避免中断后失去上下文。
>
> 约束：本轮按要求不编译、不运行 `vue-vben-admin/apps/web-antd`，前端收尾仅做静态联动核查与代码修正。

## 1. 收尾目标

本轮必须完成以下 5 项，未全部完成前不视为收尾完成：

1. 盘点当前未完成改动，识别编译、测试和联动缺口
2. 补齐并维护项目计划与进度文档
3. 收口后端 `workflow / acceptance / workspace / provider routing` 主链实现
4. 收口前端工作流控制台对应接口与页面
5. 完成至少一轮后端、前端验证，并回写最终结果

## 2. 当前状态快照

### 2.1 工作流

- 状态：本轮收尾完成
- 范围：`workflow` 控制器拆分、验收闭环、workspace 交付元数据、轨迹/回放/评测视图
- 当前已知问题：
  - 主链后端已通过测试，前端按约束完成静态联动收口
  - 保留的验证约束是 `web-antd` 未执行构建级验证，且 `validate.sh` 未在当前服务器直接执行

### 2.2 进度总览

| 序号 | 任务 | 状态 | 备注 |
|------|------|------|------|
| 1 | 盘点当前未完成改动、识别编译/测试缺口 | 已完成 | 已完成仓库状态、diff 范围、文档与新增控制器/页面梳理 |
| 2 | 新增并维护项目计划进度文档 | 已完成 | 本文档已持续更新，文档索引已纳入 |
| 3 | 补完后端工作流/验收/workspace 主链实现 | 已完成 | `go test ./app/mvp/... ./utility/provider/... ./app/ai/...` 通过 |
| 4 | 补完前端工作流控制台对应页面与接口 | 已完成 | 已完成页面/接口静态对照，并修复项目切换刷新与阶段状态展示问题 |
| 5 | 执行验证并回写最终进度 | 已完成 | 后端测试通过；前端按约束完成静态核查，未编译 `web-antd` |

## 3. 执行策略

1. 以编译和测试结果为主线，不靠人工猜测“可能完成”
2. 优先修复阻塞全量测试的低风险问题，再处理逻辑与联动问题
3. 每完成一个阶段，必须回写本文档中的状态和日志
4. 若发现新缺口，直接并入本文档，不另起临时 TODO

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
  - migration 已执行完成，当前系统检查已达到 `gate 9/9, profile 9/9`
  - 已通过 EasyMVP 系统再次触发贪吃蛇项目验证：`verificationRunID=317752601615536128`，结果 `passed`
  - 最新验证证据已明确记录 `profileSource=category:software_dev` 与 `gateSource=category:software_dev`，说明分类级 profile 与 gate 都已进入真实验收链
- 当前剩余风险：
  - 前端未做构建级验证，若后续解除约束，建议补一轮 `web-antd` 类型检查/构建验证
  - `validate.sh` 未在当前服务器直接执行；如需脚本级确认，请在隔离验证环境运行

## 5. 完成定义

满足以下条件后，本文档状态才能改为“本轮收尾完成”：

1. 后端相关包通过编译，关键测试可执行
2. 前端工作流页面与接口完成联动核对；如允许，再补类型检查或构建验证
3. 本文档和文档索引已反映最终结果
4. 未完成项被压缩到明确、可追踪、非主链阻塞范围
