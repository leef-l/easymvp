# EasyMVP 文档分层与使用说明

> 更新时间：2026-04-20  
> 上位文档：[README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md)  
> 关联文档：[EasyMVP-术语与枚举统一表.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-术语与枚举统一表.md)

---

## 1. 文档目的

这份文档用于给 `钱学森总纲设计` 目录做分层。

它只解决两个问题：

1. 每份文档在整个体系里是什么层级
2. 实际推进时应该先读哪类，再读哪类

它不新增设计内容，只负责导航和去重。

---

## 2. 分层原则

这个目录里的文档统一分成 4 层：

1. 总纲层
2. 结构层
3. 落地层
4. 对齐层

含义如下：

| 层级 | 作用 | 什么时候看 |
|---|---|---|
| 总纲层 | 定义总方向、总边界、总原则 | 刚进入项目或重新校准方向时 |
| 结构层 | 定义职责边界、阶段协作、验证与 I/O 结构 | 设计对象和流程时 |
| 落地层 | 定义缺口、施工项、字段、页面、状态机 | 真正准备实现时 |
| 对齐层 | 统一命名、枚举、术语、阅读入口 | 收口和防漂移时 |

---

## 3. 文档分层结果

## 3.1 总纲层

1. [钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/钱学森工程控制论总方案-brain-v3-easymvp-easymvp-brain.md)
2. [EasyMVP工程铁律.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP工程铁律.md)
3. [EasyMVP-三层验证架构说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-三层验证架构说明.md)

这 3 份回答的是：

- 为什么这样设计
- 哪些硬约束不能碰
- 验证环境的总口径是什么

## 3.2 结构层

1. [easymvp-brain-职责与边界定义.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-职责与边界定义.md)
2. [easymvp-brain-输入输出契约.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/easymvp-brain-输入输出契约.md)
3. [EasyMVP-四基础专精大脑阶段调用矩阵.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-四基础专精大脑阶段调用矩阵.md)
4. [EasyMVP-Verification-Contract统一设计.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-Verification-Contract统一设计.md)
5. [EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-中央大脑与四专精大脑IO合同及升级规则.md)

这 5 份回答的是：

- 谁负责什么
- 对象如何流动
- 验证合同如何建
- 多脑输入输出和升级规则是什么

## 3.3 落地层

1. [EasyMVP-钱学森总纲落地缺口与实施顺序.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-钱学森总纲落地缺口与实施顺序.md)
2. [EasyMVP-专项实施清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-专项实施清单.md)
3. [EasyMVP-对象级字段清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-对象级字段清单.md)
4. [EasyMVP-页面读取与展示清单.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-页面读取与展示清单.md)
5. [EasyMVP-闭环状态机补充说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-闭环状态机补充说明.md)

这 5 份回答的是：

- 还有什么没做
- 先做什么
- 对象字段怎么定
- 页面怎么展示
- 自动推进和阻断怎么定

## 3.4 对齐层

1. [EasyMVP-术语与枚举统一表.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-术语与枚举统一表.md)
2. [README.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/README.md)
3. [EasyMVP-文档分层与使用说明.md](/www/wwwroot/project/easymvp/docs/钱学森总纲设计/EasyMVP-文档分层与使用说明.md)

这 3 份回答的是：

- 用什么词
- 从哪篇进入
- 哪篇是主文、哪篇是辅文

---

## 4. 主文与辅文

为了避免重复阅读，建议固定 5 份主文，其余默认按需查阅：

### 主文

1. 总方案
2. `easymvp-brain` 输入输出契约
3. `Verification Contract` 统一设计
4. 专项实施清单
5. 对象级字段清单

### 辅文

1. 职责边界定义
2. 四基础专精大脑阶段调用矩阵
3. 中央大脑与四专精大脑 I/O 合同及升级规则
4. 页面读取与展示清单
5. 闭环状态机补充说明
6. 术语与枚举统一表
7. 三层验证架构说明
8. 工程铁律

原则：

- 主文负责建立完整工作认知
- 辅文负责局部决策时查细节

---

## 5. 推荐使用方式

## 5.1 设计阶段

先看：

1. 总方案
2. 职责边界定义
3. 输入输出契约
4. 阶段调用矩阵

## 5.2 实现阶段

先看：

1. 专项实施清单
2. 对象级字段清单
3. Verification Contract 统一设计
4. 闭环状态机补充说明

## 5.3 页面联调阶段

先看：

1. 页面读取与展示清单
2. 对象级字段清单
3. 术语与枚举统一表

## 5.4 收口阶段

先看：

1. 术语与枚举统一表
2. 工程铁律
3. 文档分层与使用说明

---

## 6. 后续新增文档规则

后续如果还要往这个目录加文档，只允许落入以下 4 层之一。

必须满足：

1. 能明确说出属于哪一层
2. 能明确说出替代或补充哪篇现有文档
3. 不能只是换个标题重复同一层内容

不再接受：

1. 重复性的总纲稿
2. 只讲理念不讲对象的补充稿
3. 不说明层级和用途的零散文档

---

## 7. 一句话结论

这个目录后续不应再被当作文档堆栈来读，而应当按层读：

> 先总纲，再结构，再落地，最后用对齐层防漂移。
