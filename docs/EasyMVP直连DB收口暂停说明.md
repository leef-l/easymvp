# EasyMVP直连DB收口说明

更新日期：2026-04-12

这份文档记录“禁止直接 DB、统一走 repo / service”主线在当前代码基线上的完成状态，供后续回归时快速核对真实边界。

## 1. 当前判定

当前这条治理主线已经完成到可持续守护状态：

- `controller/chat` 禁区层运行时代码中的 `g.DB()` 为 `0`
- `workflow` 中非 `repo`、非测试代码的运行时代码 `g.DB()` 为 `0`
- 业务层直接打开事务入口的 `g.DB().Transaction(...)` 为 `0`
- 配置仓储实现已收敛为单一叶子包，`repo.ConfigRepo` 作为对外 facade 保持业务侧命名稳定
- 提示词文案已同步到“禁止业务编排层依赖 repo 之外 DB 入口”的最新口径
- 新增 AST guard 测试，后续回流会直接在测试阶段失败

`repo` 目录内允许使用 `g.DB()`，因为它就是数据访问层；治理范围从一开始就不是“全仓禁止 DB”，而是“禁止业务编排层直接穿透到 DB”。

## 2. 已完成收口

主链 repo 化已经覆盖：

- `controller/chat`
- `workflow/acceptance`
- `workflow/verification`
- `workflow/stage/accept`
- `workflow/stage/complete`
- `workflow/stage/execute`
- `workflow/stage/rework`
- `workflow/scheduler`
- `workflow/watchdog`
- `workflow/orchestrator`
- `workflow/domain/plan`
- `workflow/configstore`
- `workflow/executor/auto`

本轮治理一致性收尾另外完成了三件事：

- `workflow/configrepo` 成为唯一配置仓储实现，`repo.ConfigRepo` 只保留 facade 能力
- `configstore.Store` 继续依赖叶子仓储，避免 `repo -> presetutil -> rolecatalog -> configstore -> repo` 环路
- `workflow/db_access_guard_test.go` 落地，静态扫描禁区目录里的真实 `g.DB()` 调用和 `dao` 直接依赖

## 3. 运行期与文本口径

当前禁区层的判定标准是：

- 业务编排层只能依赖 `repo / service` 暴露的入口
- 不允许直接依赖 DB 原生句柄、`dao.*`、直接拼表读写
- 事务统一通过 `repo.WithTx(...)` 这类 repo 侧入口打开

提示词也已同步到这套口径，不再保留旧的、与当前代码现实不一致的描述。

## 4. 验证基线

当前权威验证入口：

- `.github/workflows/backend-guard.yml`
- GitHub Actions workflow run、日志、artifact，以及同步回项目工作目录的 `.easymvp/ci/latest.json`

以下命令块是 2026-04-12 的历史最小回归记录，只作为当时的完成证据保留；现行铁律下不再允许在本机、宿主机或 AI 会话中直接执行这些 `go test` 命令。

历史最小回归记录：

```bash
go test ./app/mvp/internal/workflow
go test ./app/mvp/internal/workflow/presetutil ./app/mvp/internal/workflow/configstore ./app/mvp/internal/workflow/repo
go test ./app/mvp/internal/controller/chat ./app/mvp/internal/workflow/acceptance ./app/mvp/internal/workflow/verification ./app/mvp/internal/workflow/stage/accept ./app/mvp/internal/workflow/stage/complete ./app/mvp/internal/workflow/stage/execute ./app/mvp/internal/workflow/stage/rework ./app/mvp/internal/workflow/scheduler ./app/mvp/internal/workflow/watchdog ./app/mvp/internal/workflow/domain/plan ./app/mvp/internal/workflow/orchestrator ./app/mvp/internal/workflow/repo ./app/mvp/internal/workflow/configstore ./app/mvp/internal/workflow/executor ./app/mvp/internal/workflow/event
```

推荐辅助检查：

```bash
rg -n "g\\.DB\\(" admin-go/app/mvp/internal/controller/chat admin-go/app/mvp/internal/workflow --glob '!**/repo/**' --glob '!**/configrepo/**' --glob '!**/*_test.go'
```

如果这条 `rg` 有新命中，先判断是否为真实运行时代码；如果是，默认按回归缺陷处理，而不是再把问题留给后续批次。

## 5. 结论

这条主线现在不再处于“暂停待继续”状态，而是已经进入“守护已完成基线、防止回流”的阶段。

后续若再出现禁区层直连 DB，不要重开一轮大范围清理，直接按回归修复并补到对应 repo / service 抽象即可。
