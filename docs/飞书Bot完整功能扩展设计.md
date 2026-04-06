# 飞书 Bot 完整功能扩展设计

## 目标

让飞书 Bot 覆盖 EasyMVP 后台所有核心操作，用户在飞书内即可完成项目全生命周期管理。

---

## 功能分组与 action 映射

### 第一组：项目管理（已有）
| action | 说明 |
|--------|------|
| create_project | 创建项目 |
| list_projects | 列出项目 |
| project_status | 项目状态/进度 |
| pause_project | 暂停项目 |
| resume_project | 恢复项目 |
| confirm_plan | 确认方案启动执行 |

### 第二组：任务管理（新增）
| action | 说明 |
|--------|------|
| list_tasks | 查看项目任务列表（支持按状态筛选） |
| retry_task | 重试失败任务 |
| skip_task | 跳过阻塞任务 |

### 第三组：审核管理（新增）
| action | 说明 |
|--------|------|
| review_status | 查看审核状态 |
| approve_review | 人工审核通过 |
| reject_review | 人工审核驳回 |

### 第四组：验收管理（新增）
| action | 说明 |
|--------|------|
| accept_status | 查看验收状态 |
| approve_accept | 验收通过 |
| reject_accept | 验收驳回/打回返工 |

### 第五组：自治管理（新增）
| action | 说明 |
|--------|------|
| autonomy_status | 查看自治模式和待审检查点 |
| approve_checkpoint | 批准自治检查点 |
| reject_checkpoint | 拒绝自治检查点 |

### 第六组：通用（已有/扩展）
| action | 说明 |
|--------|------|
| chat | 对话/不确定意图 |
| help | 帮助 |

---

## 飞书事件订阅

### 必须订阅
| 事件 | 用途 |
|------|------|
| `im.message.receive_v1` | 接收用户发给 Bot 的消息（单聊+群聊@） |

### 推送通知（Bot 主动发）
Bot 在以下时机主动推送消息给绑定用户（通过 open_id 发消息）：
- 项目执行完成
- 任务失败超过阈值
- 自治检查点需要人工审核
- 审核/验收需要人工干预

### 不需要额外事件
- 审批事件（`approval.*`）：EasyMVP 有自己的审批流，不走飞书审批
- 机器人进/出群（`im.chat.member.*`）：当前不需要

---

## 系统提示词扩展

在 `intentSystemPrompt` 中补充所有新 action 的识别规则。

---

## 主动推送通知实现

新增 `FeishuNotifier`：
- 在 engine 的关键事件点（项目完成、任务失败、检查点触发）调用
- 通过 `mvp_user_collab_binding` 查找用户的 open_id
- 调用 `adapter.SendTextToUser(openID, text)` 发消息

---

## 实施顺序

1. 扩展 `botIntent` struct 和 `intentSystemPrompt`（新增所有 action）
2. 在 `DispatchFeishuCommand` switch 中加新路由
3. 实现各新增处理函数
4. 新增主动推送 `FeishuNotifier`（注入到 engine 事件点）
