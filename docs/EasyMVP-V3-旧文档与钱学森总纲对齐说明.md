# EasyMVP V3 旧文档与钱学森总纲对齐说明

> 更新时间：2026-04-20  
> 适用范围：`/www/wwwroot/project/easymvp/docs/` 下原有 `EasyMVP-V3-*` 文档  
> 最新总纲入口：[钱学森总纲设计/README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md)

---

## 1. 文档目的

这份文档用于给原来的 V3 文档体系加一层“最新口径说明”。

它只回答三个问题：

1. 原来的 V3 文档还能不能看
2. 哪些内容仍然有效
3. 哪些内容必须以 `钱学森总纲设计` 目录为准

---

## 2. 总体结论

原来的 V3 文档**不是作废**，但它们现在的定位已经变化：

1. 原 V3 文档继续作为历史设计细稿、字段细稿、页面细稿、实现背景文档保留
2. 最新方向、边界、阶段协作、验证口径、对象口径，统一以 [钱学森总纲设计/README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md) 这一组文档为准
3. 当旧文档与钱学森总纲设计目录发生冲突时，**以钱学森总纲设计目录为准**

一句话：

> 旧 V3 文档保留“细节参考”角色，钱学森总纲设计目录承担“当前权威口径”角色。

---

## 3. 仍然有效的内容

原 V3 文档里，下面这些内容大多数仍然有效：

1. 详细 DTO / schema / 表结构
2. 页面线框与组件拆分
3. 查询接口和聚合接口的细节设计
4. migration、索引、目录规范
5. 已落地代码对应的实现背景

这些内容的问题不在“全错”，而在于：

- 顶层解释口径还是旧的
- 没有统一收束到最新钱学森总纲

---

## 4. 必须以钱学森总纲为准的内容

以下内容统一以 `钱学森总纲设计` 目录为准：

1. `brain-v3 / EasyMVP / easymvp-brain` 三层定位
2. `central / code / browser / verifier / fault` 四基础专精大脑与中央大脑的边界
3. `easymvp-brain` 的职责与输入输出契约
4. `Verification Contract` 的统一口径
5. `CompletionVerdict / FaultSummary / RepairPlanDraft / RuntimeEscalation` 这些对象的最新口径
6. `executing -> accepting -> reworking -> completed` 的闭环状态机规则
7. 验证环境口径：
   - 长期目标：`high_spec_remote`
   - 当前替代：`github_actions`

---

## 5. 重点旧文档的最新读法

## 5.1 `EasyMVP-V3总体架构设计.md`

现在应当这样读：

1. 保留其作为旧 V3 顶层草图的作用
2. 但最新三层定位和闭环原则，以总方案为准：
   - [钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md)

## 5.2 `EasyMVP-V3专精大脑接入计划.md`

现在应当这样读：

1. 保留其作为接入背景和旧设计脉络的作用
2. 最新职责边界、合同和阶段协作，以这几份为准：
   - [easymvp-brain-职责与边界定义.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md)
   - [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md)
   - [EasyMVP-四基础专精大脑阶段调用矩阵.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-四基础专精大脑阶段调用矩阵.md)

## 5.3 `EasyMVP-V3生产级分类验收体系设计.md`

现在应当这样读：

1. 保留分类验收思路
2. 最新验证合同、验证通道、完成裁决口径，以这几份为准：
   - [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md)
   - [EasyMVP-闭环状态机补充说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-闭环状态机补充说明.md)

## 5.4 `EasyMVP-V3-AgentTeam*.md`

现在应当这样读：

1. 这些文档保留为当时的实施计划和推进记录
2. 不再承担 EasyMVP 最新总纲入口职责
3. 当前 EasyMVP 专项总纲与落地口径，应回到 `钱学森总纲设计` 目录

---

## 6. 建议阅读顺序

如果现在是为了继续做 EasyMVP，而不是看历史设计，建议这样读：

1. 先看 [钱学森总纲设计/README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md)
2. 再看 [EasyMVP-专项实施清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-专项实施清单.md)
3. 再看 [EasyMVP-对象级字段清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-对象级字段清单.md)
4. 然后按需回查旧 V3 细稿

---

## 7. 一句话结论

`docs/` 下原来的 V3 文档现在应该被理解为：

> 细节资产仍然保留，顶层口径统一回收到 `钱学森总纲设计` 目录。
