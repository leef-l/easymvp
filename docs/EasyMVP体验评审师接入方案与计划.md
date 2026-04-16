# EasyMVP 体验评审师接入方案与计划

更新日期：2026-04-12

## 0. 当前状态（2026-04-12）

按“最小闭环”口径，本专题已经完成代码层接入：

- 已为 `software_dev / game_dev` 补齐 `experience_reviewer/max` 默认预设
- 默认预设统一绑定启用模型 `315100000000000007 (hunyuan-turbos)`
- 系统自检已新增体验评审师 readiness 检查，校验预设存在、`model_id` 非 0、模型记录存在且启用
- 质量标准、acceptance 规则、默认提示词和前端角色定义原有支持继续保留，无需新增 schema / stage / 前端硬编码

## 1. 背景与目标

“体验师”在当前系统中的规范名称统一为：

- `roleType`: `experience_reviewer`
- 展示名：`体验评审师`

本方案的目标不是从零设计一个新角色，而是在现有 Workflow V2 主链上，把已经存在的体验评审能力补齐为“可默认生效、可自检、可验收、可交接”的生产级配置。

目标达成后，系统应满足以下结果：

- 编码类交互项目命中体验标准时，可以稳定解析到 `experience_reviewer`
- review 阶段只在真正缺角色时报警，不再因为默认预设缺失产生伪缺口
- accept 阶段可以用体验评审师对应模型做最终体验评审
- 前端项目创建和角色预设管理页面可以正常展示该角色，无需额外前端写死
- 系统自检能明确提示“体验评审师是否已就绪”

## 2. 现状判断

当前代码并不是“完全没做”，而是处于“主干已在，落地配置未闭环”的状态。

### 2.1 已经具备的能力

1. 角色定义已存在。
   - `admin-go/app/mvp/internal/workflow/rolecatalog/catalog.go`
   - 内置定义已经包含 `experience_reviewer / 体验评审师 / AcceptanceJudge=true`

2. 默认提示词和推荐等级已存在。
   - `admin-go/app/mvp/internal/workflow/presetutil/prompts.go`
   - 已支持 `PreferredRoleLevels("experience_reviewer")`
   - 已支持 `buildExperienceReviewerRolePrompt(...)`

3. 标准层已经会要求该角色。
   - `admin-go/app/mvp/internal/workflow/qualitygate/standard.go`
   - Web 交互、游戏客户端、Android 原生、iOS 原生几类标准已挂上 `RequiredProjectRoles`

4. review / accept 链路已经消费该角色。
   - `admin-go/app/mvp/internal/workflow/stage/review/service.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/rule_engine.go`
   - `admin-go/app/mvp/internal/workflow/acceptance/judge.go`

5. 前端角色定义兜底也已经带上该角色。
   - `vue-vben-admin/apps/web-antd/src/views/mvp/role-definitions.ts`

### 2.2 此前真正缺的不是定义，而是可运行闭环

截至 2026-04-12，这条专题的关键缺口已经补齐，当前状态如下：

1. 默认角色预设 seed 已补齐。
   - `admin-go/manifest/sql/seed/mysql_seed.sql`
   - 已新增 `software_dev / game_dev -> experience_reviewer/max`

2. 系统自检已覆盖体验评审师就绪性。
   - `admin-go/app/mvp/internal/controller/chat/workflow_system_check.go`
   - 当前会单独检查体验评审师默认预设与模型可用性

3. V2 默认预设解析仍然依赖 `mvp_role_preset`，不是只靠角色注册表。
   - `admin-go/app/mvp/internal/workflow/repo/project_role_repo.go`
   - `GetProjectRoleByLevel(...)` 的默认回退来源仍是分类默认预设，这也是本次必须补 seed 的根因

4. 模型绑定仍然不能留空。
   - `admin-go/app/mvp/internal/engine/role_resolver.go`
   - `ResolveProjectModelInfo(...)` 在 `model_id=0` 时会直接失败，所以本次新增预设全部绑定了非 0 模型

结论已经从“缺闭环”变成“闭环已补齐，后续只需按正常验证流程继续跟进环境实跑”。

## 3. 关键结论

### 3.1 本次接入不需要新增数据库结构

当前表结构已经够用，接入应优先复用：

- `workflow.role_definitions`
- `mvp_role_preset`
- `mvp_project_role`
- 现有 review / accept / verification 标准链路

本次不建议新增表、不建议新增 stage、不建议新增单独的体验验收流程分叉。

### 3.2 规范名必须统一为 `experience_reviewer`

文档、后端、前端、测试、验收口径全部以 `experience_reviewer` 为唯一代码标识。

- “体验师”只作为口头简称
- “体验评审师”作为产品展示名
- 不再引入 `tester`、`qa_guardian`、`ux_reviewer` 等平行编码

### 3.3 默认预设必须走 `category_code`，不能再以中文展示名为主

V2 默认回退优先使用 `categoryCode` 查询：

- `admin-go/app/mvp/internal/workflow/repo/role_preset_repo.go`
- `admin-go/app/mvp/internal/workflow/repo/project_role_repo.go`

因此本次新增预设应以这些稳定编码为主：

- `software_dev`
- `game_dev`

旧的中文分类值如“软件开发”“游戏开发”只保留兼容意义，不应再作为主方案继续扩张。

### 3.4 体验评审师新增预设不能复制当前 `model_id=0` 的旧模式

当前部分 V2 分类默认预设行使用了 `model_id=0`。但 `ResolveProjectModelInfo(...)` 不接受零模型。

所以体验评审师接入必须明确一条硬约束：

- 新增的 `experience_reviewer` 默认预设必须绑定可用 `model_id`
- 不能依赖“后续运行时再猜一个模型”
- `execution_mode` 固定为 `chat`

这也是本方案最重要的实施约束。

## 4. 方案范围

### 4.1 本次范围

1. 补齐 `experience_reviewer` 默认角色预设
2. 补齐系统自检项
3. 验证前端项目创建页/角色预设页能展示该角色
4. 补测试与交付文档

### 4.2 本次不做

1. 不新增数据库表或字段
2. 不新增独立“体验验收阶段”
3. 不强制给所有项目分类接入体验评审师
4. 不强制引入新的 `ai_model.capability=experience_reviewer`
5. 不顺手重构全部历史 `model_id=0` 默认预设

第 5 点很重要。本次只要求“体验评审师新增预设不能继续为 0”，不把全量历史预设治理绑进同一需求。

## 5. 推荐设计

### 5.1 接入分类

首批只接入真正会被当前标准层要求的编码类分类：

- `software_dev`
- `game_dev`

原因：

- `qualitygate` 当前只会在交互型编码项目上要求体验评审师
- `product_design` 虽然业务上也可能需要体验视角，但现有标准层并未强依赖该角色
- 首批先把 Workflow V2 的硬约束跑通，避免过度扩张

### 5.2 预设形态

每个分类至少补 1 条默认预设：

- `role_type=experience_reviewer`
- `role_level=max`
- `execution_mode=chat`
- `is_default=1`
- `status=1`
- `project_category` 使用 `software_dev` / `game_dev`
- `system_prompt` 允许为空，直接复用内置 `buildExperienceReviewerRolePrompt(...)`
- `model_id` 必须为非 0 的可用模型

建议不要在 MVP 首版同时铺 `max/pro/lite` 三档。首版只需要一条可稳定解析的默认预设，把主链打通即可。

### 5.3 模型绑定策略

体验评审师本质更接近“验收评审”而不是“代码实现”，因此模型策略建议如下：

1. 优先选择现有启用的高上下文 `chat` / `auditor` 型模型
2. 不要求新增 `capability=experience_reviewer`
3. 对内置 demo seed，可直接复用现有审核模型

建议默认值：

- 若沿用仓库自带 seed，可优先复用已存在的审核模型，例如 `ai_model` 中的 auditor 类模型
- 若目标环境不是仓库内置 seed 环境，则由实施人先在模型管理中确认一条稳定模型，再写入 `mvp_role_preset`

### 5.4 系统自检策略

系统自检不建议新增一条过重的“新模型能力校验”，而应做两层检查：

1. `experience_reviewer` 默认预设是否存在
2. 该预设引用的 `model_id` 是否非 0，且模型记录存在、启用

建议口径：

- 缺少预设：`error`
- 预设存在但模型缺失或模型禁用：`error`
- 预设和模型都可用：`ok`

这样可以直接反映“体验评审师是否可真正跑起来”，而不是只检查元数据。

### 5.5 前端策略

前端不需要为这个角色单独开发一套新 UI。

现有页面已经具备基础能力：

- 项目创建页会按分类读取角色预设
  - `vue-vben-admin/apps/web-antd/src/views/mvp/project/modules/form.vue`
- 角色预设页会按角色定义接口加载角色列表
  - `vue-vben-admin/apps/web-antd/src/views/mvp/role_preset/modules/form.vue`

因此前端本次以“验证与轻提示”为主，不以新增功能为主：

- 验证 `getRolePresets(categoryCode)` 返回后页面能展示体验评审师
- 如需补体验提示，只做文案级增强，不做结构性改造

## 6. 分阶段实施计划

### Phase 0：基线确认

目标：确认体验评审师接入不再沿用错误假设。

工作项：

1. 确认统一命名为 `experience_reviewer / 体验评审师`
2. 确认首批接入分类为 `software_dev`、`game_dev`
3. 确认新增默认预设必须绑定非 0 `model_id`
4. 确认不新增 schema、不新增 stage

产出：

- 本方案文档定稿

### Phase 1：后端默认预设补齐

目标：让 V2 在没有显式项目角色时，也能解析到体验评审师。

建议改动文件：

- `admin-go/manifest/sql/seed/mysql_seed.sql`

工作项：

1. 为 `software_dev` 新增 1 条 `experience_reviewer/max` 默认预设
2. 为 `game_dev` 新增 1 条 `experience_reviewer/max` 默认预设
3. `execution_mode` 使用 `chat`
4. `model_id` 绑定一条已启用、可用于体验评审的模型
5. `system_prompt` 可置空，交给 `presetutil` 默认提示词生成

完成标准：

- 通过 `repo.ListRolePresets(...CategoryCode...)` 可以查到 `experience_reviewer`
- `engine.ResolveProjectRoleByLevel(projectID, "experience_reviewer", "")` 可稳定回退到默认预设

### Phase 2：系统自检补齐

目标：让运维/实施人员能一眼看出体验评审师是否真的可用。

建议改动文件：

- `admin-go/app/mvp/internal/controller/chat/workflow_system_check.go`
- `admin-go/app/mvp/internal/controller/chat/workflow_system_check_test.go`

工作项：

1. 增加 `experience_reviewer` 默认预设数量检查
2. 增加 `experience_reviewer` 默认预设模型有效性检查
3. 返回信息中明确提示已找到多少条预设、对应模型是否可用

完成标准：

- 系统自检页面能单独显示体验评审师就绪状态
- 当预设存在但模型无效时，不得误报为正常

### Phase 3：前端展示与创建链路验证

目标：确认新增预设可以自然进入现有创建/管理流程。

建议关注文件：

- `vue-vben-admin/apps/web-antd/src/views/mvp/project/modules/form.vue`
- `vue-vben-admin/apps/web-antd/src/views/mvp/role_preset/modules/form.vue`
- `vue-vben-admin/apps/web-antd/src/views/mvp/role-definitions.ts`

工作项：

1. 验证项目创建页在 `software_dev`、`game_dev` 下能看到体验评审师预设
2. 验证角色预设管理页能正常创建/编辑/展示 `experience_reviewer`
3. 若页面文案需要说明“体验师即体验评审师”，只补轻量说明

完成标准：

- 无需前端硬编码新角色，也能正确展示
- 不新增新的静态常量分叉

### Phase 4：测试与验收

目标：把“代码里有定义”升级为“场景上真的能跑通”。

建议补的测试面：

1. `workflow_system_check_test.go`
   - 新增体验评审师预设存在/缺失/模型失效三种断言

2. `qualitygate/standard_test.go`
   - 确认 Web / Game / Android / iOS 标准仍然正确要求 `experience_reviewer`

3. `acceptance/rule_engine_test.go`
   - 覆盖缺少 `experience_reviewer` 时命中 `software.required_project_role_missing`

4. `presetutil/prompts_test.go`
   - 确认体验评审师默认提示词未回退丢失

建议补的联调验证：

1. 创建 `software_dev` 项目，不显式选择项目角色，只依赖默认预设
2. 命中交互型标准后，review 阶段不再因为缺体验评审师报伪缺口
3. accept 阶段能成功解析体验评审师模型并完成评审

## 7. 验收标准

以下 6 条全部满足，才视为“体验评审师已接入完成”：

1. `workflow.role_definitions` 返回中包含 `experience_reviewer`
2. `getRolePresets("software_dev")` 和 `getRolePresets("game_dev")` 能返回默认体验评审师预设
3. 默认预设中的 `model_id` 非 0，且模型真实存在并启用
4. `engine.ResolveProjectRoleByLevel(..., "experience_reviewer", "")` 可成功解析
5. 交互型项目在 accept 阶段不再因缺体验评审师命中阻塞错误
6. 系统自检能明确展示体验评审师 readiness

## 8. 风险与注意事项

### 8.1 `acceptanceJudge` 不是当前运行时选择开关

角色定义中的 `acceptanceJudge` 目前主要是元数据，并不是运行时最终选 judge 的唯一开关。

当前真正决定验收角色的是：

- `qualitygate.VerificationStandard.RequiredProjectRoles`
- `VerificationStandard.JudgeRoleRequirement()`
- `acceptance/judge.go`

因此不要误以为只把 `acceptanceJudge=true` 打开就算完成接入。

### 8.2 不要继续扩张旧中文分类 seed

新增 `experience_reviewer` 时，如果只往“软件开发”“游戏开发”这类中文分类值上补数据，V2 默认回退仍然可能绕不过去。

本次以 `software_dev / game_dev` 为准。

### 8.3 不要复制 `model_id=0` 旧模式

体验评审师是验收关键角色，必须直接可解析模型。

如果新增预设仍为 `model_id=0`，该角色大概率会在 accept 阶段直接失败，而不是优雅降级。

### 8.4 不借这次需求顺手扩大改动面

这次需求只做“体验评审师闭环接入”，不要顺手把以下内容绑进主交付：

- 全量历史默认预设回填模型
- 新增专用 capability 枚举
- 新增体验验收 UI 页面
- 重写项目创建流程

这些都可以单列后续债务。

## 9. 交接清单

交给实现同学时，至少要同步这 5 件事：

1. 规范名统一用 `experience_reviewer`
2. 默认预设只补 `software_dev`、`game_dev`
3. 新增预设必须绑定非 0 `model_id`
4. 系统自检要检查“预设存在 + 模型有效”
5. 验收不以 `acceptanceJudge` 字段为完成标志，而以真实运行链路为准

## 10. 建议提交拆分

建议拆成 3 个提交，避免混在一起难回滚：

1. `feat: add experience reviewer default presets`
2. `feat: add experience reviewer readiness checks`
3. `docs: document experience reviewer rollout plan`

这样即使前两步只落一部分，也能独立定位问题和回滚。
