# EasyMVP V3 Acceptance 线框图设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Acceptance详细页面设计](./EasyMVP-V3-Acceptance详细页面设计.md)
> 关联文档：[EasyMVP-V3-Evidence卡片组件规范](./EasyMVP-V3-Evidence卡片组件规范.md)
> 关联文档：[EasyMVP-V3-Evidence Preview交互设计](./EasyMVP-V3-Evidence%20Preview%E4%BA%A4%E4%BA%92%E8%AE%BE%E8%AE%A1.md)
> 关联文档：[EasyMVP-V3-Replay Drawer页面设计](./EasyMVP-V3-Replay%20Drawer页面设计.md)
> 关联文档：[EasyMVP-V3-Release Gate抽屉设计](./EasyMVP-V3-Release%20Gate%E6%8A%BD%E5%B1%89%E8%AE%BE%E8%AE%A1.md)
> 目标：把 Acceptance 页的模块、区域关系、信息层级和关键交互收口为统一线框结构，支撑页面实施与后续高保真设计。

## 1. 设计结论

Acceptance 页不是普通详情页，也不是大表格页。

它的正确形态应该是：

> 一个以“验证合同 + 结果差距 + 证据实体 + 完成裁决 + 必要放行”为主线的交付页面。

因此线框图必须把四件事同时放在一个页面里：

1. 合同要求
2. Coverage / Contract Gap
3. Evidence 实体
4. 完成裁决与放行

## 2. 页面总结构

建议采用四段式主布局：

1. 顶部验证与裁决条
2. 中部左侧 Coverage 区
3. 中部右侧 Evidence 区
4. 底部 Release Gate 区

推荐线框如下：

```text
┌──────────────────────────────────────────────────────────────────────┐
│ Verification + Completion Bar                                       │
├───────────────────────────────┬──────────────────────────────────────┤
│ Coverage Matrix / Gap List    │ Evidence Cards                      │
│                               │                                      │
│                               │                                      │
├───────────────────────────────┴──────────────────────────────────────┤
│ Release Gate / Blocking Reason / Manual Release                     │
└──────────────────────────────────────────────────────────────────────┘
```

## 3. 顶部最终裁决条线框

### 3.1 目标

让用户在一眼内知道：

1. 合同要求了什么
2. 当前验证结论是什么
3. 为什么还不能 `completed`
4. 是否还需人工放行

### 3.2 线框内容

```text
┌──────────────────────────────────────────────────────────────────────┐
│ verdict | decision | completed | preferred_channel                  │
│ missing_evidence | failed_checks | manual_release_required          │
└──────────────────────────────────────────────────────────────────────┘
```

### 3.3 视觉要求

建议：

1. `decision / completed` 是最强视觉锚点
2. `manual_release_required` 单独高亮
3. `missing_evidence / failed_checks` 不能埋进小字
4. 若通道是 `github_actions`，必须显示“替代通道”提示

## 4. Coverage 区线框

### 4.1 目标

让用户迅速看懂：

1. 哪些 surface 已覆盖
2. 哪些 journey 还缺
3. 缺口是证据缺口还是功能缺口

### 4.2 推荐结构

```text
┌───────────────────────────────┐
│ Coverage Header               │
├───────────────────────────────┤
│ Surface / Journey Matrix      │
├───────────────────────────────┤
│ Blocking Gap List             │
└───────────────────────────────┘
```

### 4.3 覆盖矩阵

每个格子至少显示：

1. `surface`
2. `journey`
3. `status`
4. `evidence_count`
5. `blocking_gap`

### 4.4 阻塞缺口列表

矩阵下方建议单独列出：

1. 当前 blocking gap
2. 缺少哪些 evidence type
3. 哪些 gap 会阻塞 `completed`

## 5. Evidence 区线框

### 5.1 目标

把抽象裁决翻译成用户能实际看到、点开、验证的证据实体。

### 5.2 推荐结构

```text
┌──────────────────────────────────────┐
│ Evidence Header / Group Switch       │
├──────────────────────────────────────┤
│ Group: browser                       │
│   [Evidence Card] [Evidence Card]    │
├──────────────────────────────────────┤
│ Group: runtime                       │
│   [Evidence Card] [Evidence Card]    │
├──────────────────────────────────────┤
│ Group: verification                  │
│   [Evidence Card] [Evidence Card]    │
└──────────────────────────────────────┘
```

### 5.3 分组

建议默认分组：

1. `browser`
2. `ci/build`
3. `runtime`
4. `verification`
5. `export`

### 5.4 卡片网格

桌面端建议：

1. 双列或三列

移动端建议：

1. 单列

## 6. Release Gate 区线框

### 6.1 目标

把“最后一步能不能交付、为什么还不能完成”单独拉出来，而不是散在上面各区。

### 6.2 推荐结构

```text
┌──────────────────────────────────────────────────────────────────────┐
│ can_release | requires_manual_release | current_blocking_reason     │
│ channel_note | manual_checkpoint | escalation_hint                 │
│ [View Reason] [Manual Release] [View History]                       │
└──────────────────────────────────────────────────────────────────────┘
```

### 6.3 状态要求

#### 功能通过但未生产通过

底部主文案应强调：

1. 为什么不能 release
2. 还差哪些 gap

#### 验证通过但仍未 completed

底部主文案应强调：

1. 当前是验证通过但未完成，还是人工检查点未完成
2. 决策卡住在 `manual_checkpoint / blocked / rework` 哪一类

#### 生产通过但需人工放行

底部主按钮应为：

1. `人工放行`

#### 真正完成

底部主文案应为：

1. `Ready for delivery`

但前提必须是：

1. `CompletionVerdict.completed = true`

## 7. 抽屉与弹层关系

Acceptance 页本身不应塞太多展开细节。

建议通过两个侧向能力承接深层内容：

1. `Evidence Preview Drawer`
2. `Replay Drawer`

### 7.1 Evidence Preview Drawer

入口：

1. 点击 Evidence 卡的 `查看预览`

### 7.2 Replay Drawer

入口：

1. 点击 Evidence 卡的 `查看回放`
2. 从 Preview Drawer 再进入

## 8. 响应式结构

### 8.1 桌面端

建议：

1. 左右双栏
2. Coverage 在左
3. Evidence 在右

### 8.2 平板端

建议：

1. 上下两段
2. Coverage 在上
3. Evidence 在下

### 8.3 移动端

建议：

1. 单列
2. Final Judgement 在顶部
3. Coverage、Evidence、Release Gate 纵向堆叠

## 9. 不该怎么做

不建议：

1. 把 Acceptance 页做成 issue 列表页
2. 把 Evidence 卡埋在二级 tab 里
3. 把 Release Gate 藏到页面最深处
4. 让 Preview/Replay 弹层和主页面互相抢焦点
5. 只显示“验收通过/失败”，不展示 contract gap

## 10. 对后续文档的约束

本专题确定后，后续文档应遵守：

1. 高保真设计必须保留四段式信息层级
2. Evidence 区必须复用卡片组件规范
3. Preview 与 Replay 的入口必须留在卡片级

## 11. 后续细分专题

本专题后续继续拆：

1. Acceptance 高保真视觉设计
2. [EasyMVP-V3-Release Gate抽屉设计](./EasyMVP-V3-Release%20Gate%E6%8A%BD%E5%B1%89%E8%AE%BE%E8%AE%A1.md)
3. [EasyMVP-V3-Coverage-Matrix组件规范](./EasyMVP-V3-Coverage-Matrix组件规范.md)
