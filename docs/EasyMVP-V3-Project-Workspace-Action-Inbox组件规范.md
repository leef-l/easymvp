# EasyMVP V3 Project Workspace Action Inbox 组件规范

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-Workspace详细页面设计](./EasyMVP-V3-Workspace详细页面设计.md)
> 关联文档：[EasyMVP-V3-创建初始化事件流接口设计](./EasyMVP-V3-创建初始化事件流接口设计.md)
> 目标：定义单项目 `Action Inbox` 区域的卡片结构、排序策略、动作模型和首次进入特殊表现。

## 1. 设计结论

`Action Inbox` 是单项目工作台最重要的右侧区域。

它必须把“现在该你做的事”从所有状态和事件中剥离出来。

## 2. 组件目标

每张卡必须回答：

1. 发生了什么
2. 这是不是 blocker
3. 推荐怎么处理
4. 点下去去哪

## 3. 卡片结构

建议固定为：

1. 标题
2. 一句话摘要
3. 严重级别
4. 是否阻塞
5. 推荐动作按钮

## 4. 建议类型

1. `review_blocker`
2. `risk_confirmation`
3. `acceptance_blocker`
4. `manual_release_required`
5. `run_sync_failed`
6. `creation_recovery_required`
7. `verification_conflict`
8. `fault_loop_detected`
9. `policy_denied`

## 5. 排序

建议：

1. `blocking=true` 最优先
2. `manual_release_required` 第二
3. `verification_conflict / fault_loop_detected` 第三
4. `risk_confirmation` 第四
4. 其他按更新时间

## 6. 首次进入特殊规则

项目首次进入时，`Action Inbox` 不应堆 blocker 卡。

此时优先展示：

1. `Open Plan`
2. `Confirm project goal`
3. `Check workspace path`

不要在首次进入时伪造“已进入执行/验收”类动作卡。

## 7. 线框

```text
┌──────────────────────────────────┐
│ Review Blocker                   │
│ 2 blocking issues found in plan  │
│ high · blocking                  │
│ [Open Plan]                      │
└──────────────────────────────────┘
```

## 8. 不该怎么做

不应该：

1. 右侧只是普通消息列表
2. 没有主动作
3. blocker 和建议动作混在一起没有层级
4. 把 `verification_conflict` 或 `fault_loop_detected` 隐藏成普通失败

## 9. 后续细分专题

1. Action Inbox 卡片视觉稿
2. Action Inbox 数据字段约束
