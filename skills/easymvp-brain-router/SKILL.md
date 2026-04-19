---
name: easymvp-brain-router
description: EasyMVP 的脑协作路由技能。用于判断中央大脑、代码大脑、浏览器大脑、审核大脑、故障大脑、easymvp-brain 在不同阶段和不同任务中的职责边界、调用顺序与升级条件。涉及脑选择、阶段调用、验证路径、返工恢复时必须使用。
---

# EasyMVP Brain Router

定义 EasyMVP 中谁该干什么，什么时候该调用哪个脑。

本技能只解决“路由与边界”，不替代 `easymvp-control-doctrine`。使用时默认已经遵守总纲。

## 何时使用

出现下列情况时使用：

- 设计 `brain-v3` 与 EasyMVP 的接入关系
- 判断某任务应该由哪个脑承担
- 设计阶段调用矩阵
- 设计验证、返工、故障恢复的脑协作路径
- 遇到“审核大脑 / 故障大脑 / easymvp-brain 谁来判”这类边界问题

## 基本分工

### Central Brain

职责：

- 任务理解
- 路由与委派
- 脑间协调
- 结果汇总与仲裁

不负责：

- 直接大规模代码实现
- 直接浏览器采证
- 直接承担 EasyMVP 业务裁决

### 代码大脑 `code`

职责：

- 文件读写
- 代码修改
- 局部命令执行
- 局部实现验证

不负责：

- 最终验收裁决
- 浏览器侧证据
- 业务规则判定

### 浏览器大脑 `browser`

职责：

- 页面访问
- 行为路径验证
- 页面内容提取
- 截图、录屏、页面级证据

不负责：

- 代码实现
- 业务合同编译
- 通用故障裁决

### 审核大脑 `verifier`

职责：

- 只读核验
- 断言
- 检查执行结果是否符合期望
- 审核证据是否足够支撑当前结论

不负责：

- 大规模写代码
- 生成业务计划
- 浏览器交互操作本身

### 故障大脑 `fault`

职责：

- 失败分类
- 故障原因分析
- 恢复路径建议
- 判断该 retry、rework、pause 还是 handoff

不负责：

- 直接替代 `easymvp-brain` 生成完整业务返工方案

### `easymvp-brain`

职责：

- 方案审核
- 方案编译
- 返工重构
- 验收规则映射
- 完成语义裁决

不负责：

- 通用代码编辑
- 浏览器自动化
- 通用故障执行

## 路由规则

### 规则 1：涉及业务语义，先过 `easymvp-brain`

下面这些事不应直接交给通用脑：

- 计划是否合理
- 任务是否需要拆分
- 返工是否需要重编译
- 什么证据才算 acceptance
- 执行成功是否等于业务完成

这些先交给 `easymvp-brain`。

### 规则 2：涉及代码落地，走代码大脑

一旦任务已经被编译为明确合同，代码实现默认交给：

- `central -> code`

### 规则 3：涉及页面事实与交互证据，走浏览器大脑

以下情况优先调 `browser`：

- 页面流程验证
- UI 状态确认
- 截图、页面提取、行为证据
- 前端 acceptance 证据补充

### 规则 4：涉及是否满足条件，走审核大脑

以下情况优先调 `verifier`：

- 合同检查
- 输出断言
- 结果核验
- 证据充分性检查

### 规则 5：涉及失败原因与恢复路径，走故障大脑

以下情况优先调 `fault`：

- run failed
- verify failed
- acceptance failed
- 多次重试仍失败
- 状态异常、路径不一致、日志冲突

### 规则 6：故障分析后，返工方案仍回到 `easymvp-brain`

`fault` 只负责：

- 解释失败
- 给出恢复方向

真正生成结构化返工方案时，应走：

```text
fault -> easymvp-brain -> RepairPlanDraft
```

## 阶段调用建议

### designing

默认顺序：

```text
easymvp-brain
```

必要时：

```text
easymvp-brain -> central
```

### reviewing

默认顺序：

```text
central -> 审核大脑
```

如果是业务方案审核：

```text
easymvp-brain -> 审核大脑
```

### executing

默认顺序：

```text
central -> 代码大脑
```

若任务涉及页面取证或交互：

```text
central -> 代码大脑 + 浏览器大脑
```

### accepting

默认顺序：

```text
审核大脑 -> 浏览器大脑（如需页面证据） -> easymvp-brain（完成裁决）
```

### reworking

默认顺序：

```text
故障大脑 -> easymvp-brain -> 代码大脑
```

## 升级与降级条件

### 必须升级到人工介入

出现下列任一情况：

- 同类失败重复出现 2 次以上
- 验收与返工来回震荡
- 证据冲突，无法形成单一裁决
- 高风险任务需要突破默认策略

### 可降级处理

出现下列情况可降级：

- 浏览器证据缺失，但后端验证充分，可先保留 warning
- 页面非关键缺陷，不阻断核心交付
- 临时远端验证通道不可用，可挂起等待，不强行误判通过

## 路由输出格式

当用本技能做决策时，输出尽量固定为：

```markdown
## Brain Routing Decision

- Stage: [designing/reviewing/executing/reworking/accepting]
- Primary brain: [...]
- Supporting brains: [...]
- Why this route:
  - ...
- Not delegated to:
  - ...
- Escalation condition:
  - ...
- Evidence required:
  - ...
```

## 禁止事项

- 不要让 `code` 直接决定 acceptance
- 不要让 `fault` 直接替代返工方案编译
- 不要让 `browser` 直接承担业务规则裁决
- 不要让 `central` 变成所有事都亲自执行的单体脑
- 不要跳过 `easymvp-brain` 直接把复杂业务任务扔给通用脑

