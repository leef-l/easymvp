# EasyMVP V3 Acceptance 详细页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3实时工作台页面设计](./EasyMVP-V3实时工作台页面设计.md)
> 关联文档：[EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
> 目标：把 `Acceptance` 页细化为 V3 的最终交付裁决页面。

## 1. 设计结论

`Acceptance` 页不是 issue 列表。

它必须一眼回答：

1. 该项目分类要求什么
2. 当前覆盖到了什么
3. 还缺什么证据
4. 是否已经 `production_passed`

## 2. 页面布局

建议采用：

1. 顶部最终裁决条
2. 中部左侧 coverage 区
3. 中部右侧 evidence 区
4. 底部 release gate 区

## 3. 顶部最终裁决条

建议固定展示：

1. `functional_passed`
2. `production_passed`
3. `manual_release_required`
4. `released_by_human`
5. `blocking_issue_count`

## 4. Coverage 区

### 4.1 目标

展示分类要求与当前覆盖差距。

### 4.2 展示方式

推荐矩阵或分组卡片。

每项显示：

1. `surface`
2. `journey`
3. `status`
4. `evidence_count`
5. `blocking_gap`

### 4.3 排序

建议：

1. blocking gap 优先
2. 核心 journey 优先
3. 最近变更优先

## 5. Evidence 区

### 5.1 目标

让用户看到“证据”而不是抽象结论。

### 5.2 每张证据卡

建议展示：

1. `evidence_type`
2. `title`
3. `summary`
4. `source_brain`
5. `generated_at`
6. `preview_target`

### 5.3 证据分类

建议分组：

1. `browser`
2. `ci/build`
3. `runtime`
4. `verification`
5. `export`

## 6. Release Gate 区

### 6.1 目标

把最后放行动作单独拉出来。

### 6.2 展示内容

建议展示：

1. `can_release`
2. `requires_manual_release`
3. `release_button_text`
4. `release_action_target`
5. `current_blocking_reason`

### 6.3 交互

允许：

1. 查看人工放行原因
2. 执行放行动作
3. 查看放行记录

## 7. 状态设计

### 7.1 功能通过但未生产通过

特征：

1. `functional_passed = true`
2. `production_passed = false`

页面主提示应为：

1. 还缺哪些 surface / evidence
2. 为什么不能 release

### 7.2 生产通过但需人工放行

特征：

1. `production_passed = true`
2. `manual_release_required = true`
3. `released_by_human = false`

页面主按钮应为：

1. `人工放行`

### 7.3 真正完成

特征：

1. `production_passed = true`
2. 必要时 `released_by_human = true`

页面主提示应为：

1. `Ready for delivery`

## 8. 不该怎么做

不应该：

1. 只显示 issue 数量
2. 不显示 evidence 实体
3. 把 `manual_release_required` 混在普通 warning 里

## 9. 后续细分专题

本专题后续继续拆：

1. Acceptance 线框图
2. Evidence 卡片组件规范
3. Release gate 抽屉设计
