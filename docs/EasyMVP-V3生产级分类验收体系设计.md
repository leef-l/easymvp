# EasyMVP V3 生产级分类验收体系设计

> 更新时间：2026-04-19  
> 上游文档：[EasyMVP-V3总体架构设计](./EasyMVP-V3总体架构设计.md)  
> 目标：把 EasyMVP 的 accept 升级成按项目分类驱动的生产级验收编排系统。

## 1. 设计结论

V3 的最终目标不是“功能通过”，而是：

> 按项目分类证明项目达到生产可交付级别。

因此验收目标必须升级为：

- `functional_passed`
- `production_passed`
- 必要时 `manual_release_required`

只有 `production_passed` 才能算真正完成。

## 2. 为什么 V2 不够

V2 的 accept 更接近：

- 通用规则检查
- CI / 文件 / issue 聚合
- 最后裁决

但对以下项目不够：

1. web 类：需要用户端 + 后台端真实操作
2. game 类：需要真实引擎运行、玩法循环验证、后台联动验证
3. video editing 类：需要真实编辑 runtime、导出结果、后台联动验证

所以 V3 必须按项目分类做验收编排。

## 3. 核心对象

### 3.1 AcceptanceProfile

按项目分类定义验收框架。

### 3.2 ProductionAcceptanceProfile

按项目分类定义生产级验收目标。

### 3.3 Surface

统一抽象需要验收的面，例如：

- `user_frontend`
- `admin_backend`
- `api_backend`
- `game_runtime`
- `editor_runtime`

### 3.4 Journey

定义必须覆盖的真实功能路径。

### 3.5 Evidence

定义验收必须收集的结构化证据。

## 4. 验收的三层结构

### 4.1 通用验收框架

所有项目共用：

- 验收 Run 生命周期
- 证据采集
- issue 分级
- 最终裁决

### 4.2 分类验收 Profile

每个分类决定：

- 需要哪些 surface
- 需要哪些 journey
- 需要哪些 evidence
- 需要哪些 brain

### 4.3 验收执行脑

由不同 brain 执行：

- `browser-brain`
- `verifier-brain`
- `code/engine brain`
- `easymvp-brain`

## 5. 生产级的额外要求

生产级验收除了功能通过，还必须覆盖：

1. `operability`
2. `release_readiness`
3. `recovery_readiness`

也就是：

- 可运营
- 可上线
- 可恢复

## 6. 不同分类的核心差异

### 6.1 Web

至少要覆盖：

- 用户端真实操作
- 后台端真实操作
- 关键状态读写一致
- 构建与 CI 证据完整

### 6.2 Game

至少要覆盖：

- 真实运行态启动
- 核心玩法循环
- 输入 / 暂停 / 结算 / 重开
- 后台端功能
- 引擎运行证据和构建证据

### 6.3 Video Editing

至少要覆盖：

- 编辑器 runtime
- 导入 / 编辑 / 导出主路径
- 后台功能
- 导出产物和运行证据

## 7. 后续细分专题

本专题后续继续拆：

1. [EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则](./EasyMVP-V3-ProductionAcceptanceProfile证据结构与裁决规则.md)
2. `AcceptanceProfile` schema
3. `ProductionAcceptanceProfile` schema
4. `surface/journey/evidence` 结构定义
5. 不同分类 profile 样例
6. 生产级裁决规则
