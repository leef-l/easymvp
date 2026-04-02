# Codegen 更新日志

## v1.6.1 — 2026-04-01

### BUG 修复

- **修复 Export CSV 外键字段显示 ID** — Export 缺少 `fillRefFields` 批量填充，CSV 输出外键字段改用 `RefFieldName`（关联名称）而非原始 ID
- **修复 labelToIdent 枚举常量名冲突** — "正常"和"普通"均映射为 `Normal`，同字段两个枚举值会导致编译报错；"普通"改为 `Regular`
- **修复 detail-drawer 缺少更新时间** — 详情抽屉只显示 `createdAt`，现在补充 `updatedAt` 和业务时间字段（`_at` 后缀）
- **修复 list.tpl ImageUpload slot 缺少隐藏判断** — `else if eq .Component "ImageUpload"` 分支缺少 `(not .IsHidden)` 等条件守卫
- **修复 TreeSelectMulti treeData 未声明** — `parent_ids` 字段映射 `TreeSelectMulti` 但 `treeData` 仅在 `HasParentID` 时声明，新增 `HasTreeSelect` 标志位扩展声明条件
- **修复非枚举 SelectMulti 渲染为 Input** — `_ids` 字段无枚举时 else 分支生成普通 Input，改为 `Select mode=tags` 支持多值输入

### 功能增强

- **Tree 接口支持时间范围和关键词搜索** — `TreeReq`/`TreeInput`/Tree logic 新增 `StartTime`/`EndTime` 和 searchable 字段过滤
- **types.tpl ListParams 补全** — 新增 `orderBy`/`orderDir`/`startTime`/`endTime` 和 searchable 字段，消除 `as any` 类型强转
- **types.tpl 新增 TreeParams** — 树形查询参数独立类型定义，`api.tpl` Tree 接口使用强类型参数
- **菜单生成器新增批量删除按钮** — `batch-delete` 权限按钮（第 4 项），与前端批量删除功能匹配

### 代码优化

- **List/Export 过滤逻辑提取** — `logic.tpl` 中 List 和 Export 重复的筛选代码提取为 `applyListFilter()` 和 `fillRefFields()` 私有方法

---

## v1.6.0 — 2026-04-01

### BUG 修复

- **修复密码字段 dependencies 失效** — `triggerFields: ['id']` 但表单中无 `id` 字段导致 rules 永不更新，改为函数式 `rules: () => (isEdit.value ? undefined : 'required')` 动态判断
- **修复树形表格 timeRange 参数未展开** — 树形 query 直接透传 `formValues` 导致 `timeRange` 数组传到后端，现在正确展开为 `startTime/endTime`
- **修复 JsonEditor 详情用 v-html 渲染** — JSON 数据改用 `<pre>` + `JSON.stringify` 格式化展示，消除 XSS 风险
- **修复 `_no` 后缀字段误用模糊搜索** — `order_no`/`_code`/`_sn` 等编号字段改为精确匹配（`Where` 而非 `WhereLike`）

### 功能增强

- **findDisplayField 扩展** — 关联表显示字段优先级扩展为：`title > name > username > nickname > real_name > label > phone > mobile`
- **图片字段识别扩展** — 新增 `cover/logo/banner/thumbnail/poster/pic` 等常见图片字段名和后缀
- **菜单权限按钮补全** — 新增"查看"（`:detail`）和"导出"（`:export`）权限按钮
- **Export 加上限保护** — 导出接口加 `Limit(10000)` 防止百万记录 OOM
- **consts 常量名语义化** — 枚举常量名从 `StatusV1` 改为 `StatusEnabled`，覆盖 30+ 常见中文标签映射
- **导出 formValues 改用公开 API** — `gridApi.formApi.form.values` 改为 `gridApi.formApi.getValues()`

### 代码优化

- **Parser 复用数据库连接** — `New()` 时建立连接并缓存，多表生成不再重复 Open/Close
- **提取 GenerateFiles 通用函数** — backend/frontend generator 重复代码提取到 `util.GenerateFiles()`
- **Tooltip 渲染提取 helper** — `list.tpl` 提取 `tooltipHeader()`、`form.tpl` 提取 `tooltipLabel()`，消除 20+ 处内联重复
- **树形表格去掉 checkbox 和批量删除** — 树形数据语义不适合批量删除，条件化生成

---

## v1.5.0 — 2026-04-01

### 新增功能

- **批量删除** — 前后端完整支持，API（`/batch-delete`）+ Service + Logic + 前端勾选框 + 批量删除按钮
- **CSV 导出** — 后端 Export 接口（不分页查询 + CSV 流式输出），前端导出按钮（Blob 下载），支持筛选条件透传
- **查看详情 Drawer** — 新增 `detail-drawer.tpl` 模板，使用 Ant Design `Descriptions` 组件只读展示所有字段，支持枚举 Tag、金额分→元、图片预览、富文本 HTML 渲染
- **关键词模糊搜索** — Parser 自动识别 `title`/`name`/`phone`/`email` 等字段为可搜索字段，后端 `WhereLike`，前端搜索表单自动添加 Input
- **时间范围筛选** — 列表页自动添加 `RangePicker` 时间范围筛选，后端 `WhereGTE`/`WhereLTE` 过滤 `created_at`
- **列表排序** — 前端列头点击排序（`sortConfig: remote`），后端动态 `OrderBy`/`OrderDir` 排序，`createdAt` 列默认可排序
- **金额字段自动格式化** — Parser 识别 `*_price`/`*_amount`/`*_balance`/`*_fee`/`*_cost` 等字段，列表自动 `(cellValue / 100).toFixed(2)` 分→元显示

### 模板优化

- **Tree 接口筛选参数贯通** — Tree 请求支持枚举筛选参数透传（`TreeReq` → `TreeInput` → Logic 条件过滤）
- **编辑时密码字段条件隐藏** — 编辑模式下密码字段使用 `dependencies` 联动隐藏，placeholder 显示"不填则不修改"
- **含 RichText/JsonEditor 时弹窗自动加宽** — 表单弹窗宽度根据 `HasRichText` 动态切换 800px/600px
- **Export 接口筛选条件支持** — 导出接口复用 ListInput，支持枚举筛选 + 关键词搜索 + 时间范围

---

## v1.4.0 — 2026-04-01

### BUG 修复

- **修复树形表格 `treeNode` 不生效** — `list.tpl` 中 `$firstDataCol` 在 RichText/JsonEditor 字段（不渲染列）时被错误消耗，导致树形表格首列永远不会标记 `treeNode: true`，展开功能失效
- **修复 `form.tpl` TreeSelect `fieldNames.label` 不一致** — `TreeSelectSingle`/`TreeSelectMulti`（parent_id）统一使用 `RefDisplayLower`（camelCase），与外键 TreeSelect 行为一致

### 性能优化

- **List 外键关联改为批量查询** — 原来每条记录每个外键字段逐条发 SQL（N+1 问题），改为先收集所有外键 ID，批量 `WHERE id IN (...)` 查询后 map 回填，性能从 O(n×k) 降至 O(k)

### 模板完善

- **`form.tpl` 补充 `IconPicker` 和 `InputUrl` 组件分支** — 之前 `field_mapper.go` 映射了这两个组件类型，但 `form.tpl` 缺少对应渲染分支，静默回退为普通 Input。现在 `IconPicker` 渲染图标选择器，`InputUrl` 渲染带 `https://` 前缀的输入框
- **`skip_fields` 配置生效** — `codegen.yaml` 中的 `skip_fields` 之前加载了但从未使用，现在会将配置中列出的字段标记为隐藏（不生成前端组件）
- **Go 包名去下划线** — 多段模块名（如 `user_role`）生成的 Go 包名自动去除下划线（`userrole`），避免 `go vet` 警告

### 代码整洁

- **提取 `replacePlaceholders` 到 `generator/util` 公共包** — 消除 `backend/generator.go` 和 `frontend/generator.go` 中的重复函数
- **合并 `renderTemplate` 和 `renderTemplateWithFuncs`** — 统一为一个内置 `ModuleCamel` 模板函数的渲染函数
- **删除 `snakeToCamelLocal`** — 改用 `parser.SnakeToCamelSimple` 导出函数
- **删除 `router.tpl` 死文件** — 该模板从未被任何 generator 使用（路由注册通过 `cmd.tpl` 完成）

### 可配置性增强

- **菜单应用目录配置化** — 应用名到标题/图标的映射从 `menu/generator.go` 硬编码移到 `codegen.yaml` 的 `menu_apps` 配置项，新增应用无需修改源码
- **数据库密码支持环境变量** — `codegen.yaml` 中 `password` 字段支持 `${ENV_VAR}` 语法，从环境变量读取，避免明文存储

---

## v1.3.0 — 2026-03-31

### Tooltip 括号语法

字段注释中的中文括号 `（）` 或英文括号 `()` 自动提取为 Tooltip 提示，括号前文字作为精简标签。

**示例：** `排序（升序）` → 标签显示"排序"，鼠标悬停显示"升序"

**变更文件：**

- `parser/meta.go` — `FieldMeta` 新增 `ShortLabel`、`TooltipText`；`TableMeta` 新增 `HasTooltip`
- `parser/comment_parser.go` — 新增 `extractParentheses()` 函数，`ParseComment()` 返回值扩展为 4 个
- `parser/parser.go` — `buildFieldMeta()` 适配新返回值，自动检测 `HasTooltip`
- `templates/frontend/form.tpl` — 表单 label 条件渲染 Tooltip + QuestionCircleOutlined 图标
- `templates/frontend/list.tpl` — 列头使用 `ShortLabel`，有提示时渲染 `slots.header` Tooltip

**生成效果：**

```vue
<!-- 无括号：普通文字 -->
label: '部门名称'

<!-- 有括号：Tooltip 渲染（v1.6.0+ 使用 helper 函数） -->
label: tooltipLabel('排序', '升序')
```

---

## v1.2.0 — 2026-03-28

### 菜单生成器

新增 `--menu` 和 `--only menu` 参数，支持将菜单数据直接写入 `system_menu` 表。

每个模块自动生成目录 + 页面 + 按钮权限（新增/编辑/删除）三级菜单。

---

## v1.1.0 — 初始版本

### 核心功能

- 数据库表结构自动解析，支持 `{应用名}_{模块名}` 表名规范
- 后端生成：API / Controller / Logic / Service / Model / Consts
- 前端生成：TypeScript 类型 / API 函数 / 列表页 / 表单弹窗
- 智能组件映射：按字段名、枚举数量、数据库类型自动选择前端组件
- 树形结构检测（`parent_id`）、密码加密、外键关联、Snowflake ID 处理
- 枚举常量自动生成（Go 常量 + 前端 options）
- `--dry-run` 预览模式、`--force` 强制覆盖、`--only backend/frontend` 部分生成
