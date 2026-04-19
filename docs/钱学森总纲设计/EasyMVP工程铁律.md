# EasyMVP工程铁律

更新日期：2026-04-13

这份文档不是建议，是 EasyMVP 主项目后续开发与验收的硬约束。

## 铁律 1：禁止直接 DB

- 禁止在 `controller`、`workflow`、`stage`、`review`、`acceptance`、`verification`、`autonomy` 等业务编排层直接调用 `g.DB()`、`dao.*`，也禁止直接拼表完成数据读写。
- 所有数据库访问必须先接口化，再下沉到 `repo` 实现。
- 上层只允许依赖稳定接口，例如 `service interface`、`repo interface`、`DTO / input / output contract`。

标准链路：

`controller -> service -> repo interface -> repo implementation -> DB`

当前已经完成 repo 收口的主链：

- `workflow.category_resolver`
- `workflow.verification.service`
- `workflow.acceptance.rule_engine`
- `workflow.stage.accept.service`

这几条链路已经禁止回退到表级直查；新增字段或查询条件，必须先补到对应 repo。

## 铁律 2：新增能力先补接口，再接业务

新增配置、角色、验收证据、状态回写等数据能力时，必须先完成：

1. 明确输入输出结构
2. 定义 service / repo 接口
3. 落 repo 实现
4. 最后接 controller / workflow / stage

不允许先在上层业务里直接查表，后面再“补抽象”。

## 铁律 3：存量债务不允许继续扩散

- 当前仓内仍存在历史直连 DB 代码，这是存量债务，不是新代码可以继续复制的理由。
- 任何新功能、新入口、新配置项，必须从第一天开始遵守接口化与标准化。
- 若本次改动碰到历史直连 DB 区域，至少要在本次改动范围内完成抽象收口，不能继续把 `g.DB()` 扩散到更多文件。

## 角色定义的落地方式

`workflow.role_definitions` 已作为项目级角色注册表。

- 展示名、颜色、默认提示词、推荐等级、是否可做验收评审，统一从这个配置读取
- 后台通过专用编辑器维护，不再要求前端写死角色展示
- 标准层只依赖稳定的 `roleType` 编码，例如 `experience_reviewer`

当前新增实现示例：

- 角色定义配置读写：`rolecatalog service -> config repo`
- 控制器不直接碰 `mvp_config`
- 前端表单、列表、详情页统一通过角色定义接口解析展示

## 验收口径

从现在开始，以下情况属于阻塞项：

- 新增控制器或阶段逻辑直接写 `g.DB()`
- 新增配置功能绕过 repo / service 直接操作 `mvp_config`
- 新增角色只在前端常量里写死，后台新增后无法生效
- 在已 repo 化主链上重新引入表级直查，例如绕过 `ProjectRepo / WorkflowRunRepo / StageRunRepo / DomainTaskRepo / VerificationRunRepo`

未满足上述铁律的改动，不得视为生产级完成。

## 铁律 4：最终验证必须在高配验证环境执行

- 后端测试、前端测试、lint、typecheck、bundle、生产构建、Go 二进制编译、镜像构建，统一不得在当前低配宿主机上直接执行。
- 禁止在宿主机、本地低配开发机、线上低配机器、AI 执行会话中直接执行任何 `go test`、`go build`、`pnpm test`、`npm test`、`pnpm build`、`pnpm exec vite build`、`pnpm exec vue-tsc`、`docker build` 等重验证或重编译命令。
- 正式验收口径只认**高配验证环境**的结果；当前阶段由于服务器配置不足，GitHub Actions 暂时承担远端高配验证角色，因此现阶段可以接受 GitHub Actions 的 workflow run、日志和 artifact 作为正式验证证据。
- 需要新增验证项时，优先修改远端验证通道；当前阶段即修改 `.github/workflows/` 或其调用脚本，再通过 `push`、`pull_request` 或 `workflow_dispatch` 触发。
- 仓库中的 `scripts/web-antd-*-safe.sh` 等本机受控脚本只保留为历史资产或 CI 迁移参考，不再作为最终验收入口。

当前主入口：

- 后端守卫链：`.github/workflows/backend-guard.yml`
- 后端与部署链：`.github/workflows/deploy.yml`
- `web-antd` 守卫链：`.github/workflows/web-antd-guard.yml`

## 铁律 5：服务器负载保护

- 当前宿主机默认只允许低开销只读检查、代码编辑、日志排查与文档更新。
- 未获人工明确许可时，不得在宿主机启动任何测试、编译、批处理、长时间扫描或其他重操作。
- 若当前 GitHub Actions 不可用，默认暂停重验证与重编译，不回退到低配宿主机执行；后续如接入独立高配验证机，则切换到高配验证机通道。

### 核心原则

- 先看负载再执行：开始任何可能消耗资源的操作前，先检查当前服务器负载。
- 超过 `80` 立即停止：当负载超过 `80` 时，立即停止继续执行，不再启动扫描、批处理或其他重操作。
- 低于 `50` 才恢复：停止后进入等待，只允许保留低开销只读观察；只有负载回落到 `50` 以下，才允许继续后续操作。
- 恢复前重新检查：每次准备恢复执行前，必须重新检查一次负载；如果再次超过 `80`，立即再次停止。

### 执行口径

- 默认以 `CPU busy` 作为负载主口径，辅助参考 `load average` 与可用内存。
- 常用检查命令：`top -bn1 | head -n 5`、`uptime`、`free -h`
- 如果人工临时指定了更严格的资源限制，以人工要求为准。
