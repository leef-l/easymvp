# EasyMVP工程铁律

更新日期：2026-04-11

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

## 铁律 4：重验证必须资源受控

- 当前宿主机允许执行重型静态检查，但必须串行、低优先级、带堆上限运行。
- `web-antd` 全量类型检查统一走 [scripts/web-antd-typecheck-safe.sh](../scripts/web-antd-typecheck-safe.sh)。
- 不允许在生产宿主机上直接裸跑高并发 `vue-tsc`、`pnpm build`、`turbo build` 一类命令。
