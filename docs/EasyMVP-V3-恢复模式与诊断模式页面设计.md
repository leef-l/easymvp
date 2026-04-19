# EasyMVP V3 恢复模式与诊断模式页面设计

> 更新时间：2026-04-19
> 上游文档：[EasyMVP-V3-错误码与诊断分级设计](./EasyMVP-V3-错误码与诊断分级设计.md)
> 关联文档：[EasyMVP-V3-本地配置与启动参数设计](./EasyMVP-V3-本地配置与启动参数设计.md)
> 关联文档：[EasyMVP-V3-单机版启动时序与进程内调用链设计](./EasyMVP-V3-单机版启动时序与进程内调用链设计.md)
> 目标：定义 V3 单机版在启动失败、migration 失败、核心服务异常、worker 持续失败时的恢复态页面与诊断态页面。

## 1. 设计结论

V3 必须有正式的恢复模式，而不是启动失败后只弹原生错误框。

建议至少定义两个特殊页面：

1. `Recovery Mode`
2. `Diagnostics Mode`

## 2. 进入条件

### 2.1 Recovery Mode

当出现以下情况时进入：

1. Go 核心服务未成功启动
2. SQLite migration 失败
3. 数据目录不可写
4. 配置损坏导致核心服务不可用

### 2.2 Diagnostics Mode

当出现以下情况时进入：

1. worker 连续失败
2. `brain serve` 长时间不可用
3. evidence / replay 索引持续异常
4. 工作台进入长期 stale

## 3. Recovery Mode 页面结构

建议包含：

1. 顶部错误摘要
2. 错误码与错误域
3. 恢复建议卡片
4. 最近诊断记录
5. 快捷动作区

快捷动作建议：

1. `Retry Startup`
2. `Open Data Folder`
3. `Enter Safe Mode`
4. `Export Diagnostics`

## 4. Diagnostics Mode 页面结构

建议包含：

1. 系统健康概览
2. 核心服务状态
3. worker 状态列表
4. 最近错误时间线
5. 恢复操作入口

## 5. 展示原则

必须做到：

1. 简单明了
2. 先说能不能恢复
3. 再说为什么失败
4. 最后才展示调试细节

## 6. 与主工作台的关系

规则建议：

1. `Recovery Mode` 阻断主工作台
2. `Diagnostics Mode` 可作为独立页或抽屉
3. 用户修复后可重新进入正常工作台

## 7. 后续细分专题

1. Recovery 页面线框图
2. Diagnostics 页面组件规范
3. 错误码到恢复动作映射表

