# OpenHands接入设计与实施文档

## 一、目标

将 OpenHands 从当前的“预留执行模式”升级为正式执行器。

---

## 二、当前现状

当前系统已有：

1. `execution_mode=openhands`
2. `ai_engine/ai_engine_config`
3. 环境检查与配置入口

当前缺失：

1. 正式 `OpenHandsAdapter`
2. 任务执行链接通
3. session 生命周期管理
4. 流输出回写

---

## 三、接入方式

优先采用两种模式之一：

### 模式 A：CLI/command_template 模式

通过 `ai_engine_config.command_template` 启动 OpenHands。

优点：

- 快速落地
- 与 Aider 风格接近

### 模式 B：HTTP/API 模式

通过远端 OpenHands 服务管理 session。

优点：

- 更适合长任务
- 更容易做状态查询和取消

建议顺序：先 A，后 B。

---

## 四、适配器设计

新增：

- `internal/executor/openhands_adapter.go`

职责：

1. 读取 OpenHands 配置
2. 构建执行请求
3. 维护 session
4. 回写流输出和结果

---

## 五、输入输出

输入：

- `taskPrompt`
- `workDir`
- `writeFiles`
- `readFiles`
- `modelInfo`
- `timeout`

输出：

- 标准输出
- 错误输出
- 结果摘要
- 失败原因
- 变更文件

---

## 六、取消与超时

必须支持：

1. workflow 暂停
2. workflow 取消
3. 单任务超时

---

## 七、实施步骤

1. 新建 `OpenHandsAdapter`
2. 接 `ExecutorRegistry`
3. 在 `execute stage` 中支持 `openhands`
4. 增加回写与日志
5. 增加 cancel/timeout

---

## 八、验收标准

1. `execution_mode=openhands` 的任务可真正执行
2. 可查看运行日志
3. 可取消
4. 可回写失败原因

