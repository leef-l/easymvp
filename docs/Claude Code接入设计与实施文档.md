# Claude Code接入设计与实施文档

## 一、目标

接入 Claude Code 作为正式 CLI 执行器。

---

## 二、定位

Claude Code 适合：

1. 跨文件重构
2. 高质量代码编写
3. 仓库理解型任务

---

## 三、接入模式

推荐采用 CLI 适配模式。

统一输入：

- workdir
- files
- prompt
- timeout
- maxSteps

---

## 四、关键适配点

1. 命令行模板
2. 模型参数注入
3. 输出捕获
4. 错误分类
5. 变更文件识别

---

## 五、实施步骤

1. 增加 `claude_code` 引擎配置
2. 增加 `ClaudeCodeAdapter`
3. 接入 `ExecutorRegistry`
4. 增加系统检测项
5. 增加角色配置支持

---

## 六、风险

1. CLI 参数稳定性
2. 输出格式变化
3. 长任务取消控制

