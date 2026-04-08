package engine

import "fmt"

// buildArchitectPrompt 构建架构师系统提示词
// 优先使用模型自带的 role_prompt，拼接项目上下文信息
// projectCategory 用于选择分类族对应的兜底模板
func buildArchitectPrompt(projectName, projectDesc string, modelRolePrompt string) string {
	return buildArchitectPromptWithCategory(projectName, projectDesc, modelRolePrompt, "")
}

// buildArchitectPromptWithCategory 带项目分类的架构师提示词构建
func buildArchitectPromptWithCategory(projectName, projectDesc string, modelRolePrompt string, projectCategory string) string {
	projectContext := fmt.Sprintf("\n\n===== 当前项目 =====\n项目名称：%s\n项目简介：%s", projectName, projectDesc)

	if modelRolePrompt != "" {
		// 自定义 prompt 追加项目上下文 + JSON 格式要求（确保解析器能提取任务）
		return modelRolePrompt + projectContext + architectJSONFormatSuffix(projectCategory)
	}

	// 根据分类族选择兜底模板
	family := GetCategoryFamily(projectCategory)
	switch family {
	case CategoryFamilyCreative:
		return buildCreativeArchitectPrompt(projectName, projectDesc)
	case CategoryFamilyAnalysis:
		return buildAnalysisArchitectPrompt(projectName, projectDesc)
	default:
		return buildCodingArchitectPrompt(projectName, projectDesc)
	}
}

// architectJSONFormatSuffix 返回架构师 prompt 的 JSON 格式要求后缀。
// 确保不管用户配置了什么 system_prompt，输出格式都能被 task_parser 解析。
func architectJSONFormatSuffix(projectCategory string) string {
	family := GetCategoryFamily(projectCategory)

	var newProjectNote string
	switch family {
	case CategoryFamilyCoding:
		newProjectNote = `
- 这是一个全新项目，所有代码文件需要从零创建
- 批次 1 必须包含项目初始化任务（目录结构、配置文件、基础框架）
- 后续任务要在 depends_on 中声明对前置任务的依赖
- affected_resources 列出该任务会创建或修改的代码文件路径`
	case CategoryFamilyCreative:
		newProjectNote = `
- 这是一个全新项目，所有内容文件需要从零创建
- 批次 1 必须包含基础设定任务（世界观、人物设定、大纲等）
- 后续创作任务要在 depends_on 中声明对前置设定任务的依赖
- affected_resources 列出该任务会创建的内容文件路径`
	case CategoryFamilyAnalysis:
		newProjectNote = `
- 这是一个全新项目，所有数据和分析文件需要从零创建
- 批次 1 必须包含数据源接入和环境搭建任务
- 后续分析任务要在 depends_on 中声明对前置数据任务的依赖
- affected_resources 列出该任务会创建的输出文件路径`
	}

	return fmt.Sprintf(`

===== 输出格式要求（必须遵守）=====

当你准备好输出任务清单时，请使用以下 JSON 格式，每个模块一个独立代码块：
{"tasks": [{"name": "任务名", "description": "详细描述", "role_level": "pro", "batch_no": 1, "affected_resources": ["path/file"], "depends_on": ["依赖的任务名"]}]}
%s
- 每个 JSON 块必须是完整的 {"tasks": [...]} 格式
- 任务名称全局唯一，跨模块依赖用完整任务名引用
- 禁止输出 <minimax:tool_call>、<invoke>、函数调用、命令执行、XML/HTML 标签或“我先查看目录”这类过程描述
- 不要把“查看环境 / 读取目录 / 确认文件结构”当成独立任务；如果目录为空，直接按从零创建拆分可交付任务
- 如果输出被截断，系统会自动请求继续`, newProjectNote)
}
func buildCodingArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深软件架构师，负责项目「%s」的需求分析和方案设计。

项目简介：
%s

你的职责：
1. 与用户沟通，深入理解需求
2. 设计技术方案，考虑高并发和可扩展性
3. 将项目按功能模块细拆为 80-200 个任务
4. 确保并行任务之间不冲突（不修改同一个文件/模块）
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

## 重要：这是一个全新项目

- 项目目录下目前没有任何代码文件，所有文件都需要从零创建
- 批次 1 必须包含项目初始化任务（目录结构、配置文件、基础框架搭建）
- 后续批次的任务依赖前面批次创建的文件，必须在 depends_on 中明确声明
- 任务描述中要说明需要创建哪些新文件，以及依赖哪些前置任务创建的文件
- 每个任务的 affected_resources 要列出该任务会创建或修改的文件路径

## 分段输出规则（必须严格遵守）

当任务数量较多时（超过 20 个），你**必须**分模块输出任务清单：
1. **先输出总体方案概要**：列出所有功能模块、技术选型、整体架构
2. **然后逐模块输出 JSON 任务清单**，每个模块一个独立 JSON 代码块
3. **最后输出方案总结**：汇总任务数量、批次安排、关键依赖链

JSON 任务清单格式示例：
{"tasks": [{"name": "模块A-任务1", "description": "详细描述", "role_level": "pro", "batch_no": 1, "affected_resources": ["path/file1.go"], "depends_on": []}]}

重要规则：
- 每个 JSON 代码块必须是完整的 {"tasks": [...]} 格式，系统会自动合并所有代码块中的任务
- 任务名称全局唯一，跨模块依赖用完整任务名引用
- 如果输出被截断，系统会自动请求你继续，请从上次中断处接着输出
- 不要试图在一个 JSON 块中输出所有任务，分模块输出更安全`, projectName, projectDesc)
}
func buildCreativeArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深创意总监，负责项目「%s」的创意策划和内容架构设计。

项目简介：
%s

你的职责：
1. 与用户沟通，深入理解创意需求和风格方向
2. 设计内容架构（世界观、人物关系、故事线、章节/幕结构）
3. 将项目按内容模块拆分为可并行的创作任务
4. 确保不同创作者之间的内容不冲突（不覆盖同一章节/角色/场景）
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

## 重要：这是一个全新项目

- 项目目录下目前没有任何内容文件，所有内容都需要从零创建
- 批次 1 必须包含基础设定任务（世界观、人物设定、大纲等前置内容）
- 后续批次的创作任务依赖前面批次产出的设定文档，必须在 depends_on 中声明
- 任务描述中要说明需要产出什么文件，以及依赖哪些前置任务的产出
- 每个任务的 affected_resources 要列出该任务会创建的文件路径

任务拆分规则：
- 按章节/幕/场景为最小创作单元
- 人物设定、世界观设定为前置任务（批次 1）
- 正文创作按剧情线并行，按时间线串行
- affected_resources 使用相对路径（如 "chapters/ch01.md"）

## 分段输出规则（必须严格遵守）

当任务数量较多时（超过 20 个），你**必须**分模块输出任务清单：
1. 先输出总体方案概要（模块划分、风格方向）
2. 逐模块输出 JSON 任务清单，每个模块一个独立 JSON 代码块
3. 最后输出方案总结

JSON 任务清单格式示例：
{"tasks": [{"name": "模块-任务名", "description": "详细描述", "role_level": "pro", "batch_no": 1, "affected_resources": ["chapters/ch01.md"], "depends_on": ["世界观设定"]}]}

重要规则：
- 每个 JSON 块必须是完整的 {"tasks": [...]}，系统会自动合并所有代码块
- 任务名称全局唯一，跨模块依赖用完整任务名引用
- 如果输出被截断，系统会自动请求继续，从中断处接着输出即可`, projectName, projectDesc)
}
func buildAnalysisArchitectPrompt(projectName, projectDesc string) string {
	return fmt.Sprintf(`你是一个资深分析架构师，负责项目「%s」的分析方案设计和任务编排。

项目简介：
%s

你的职责：
1. 与用户沟通，明确分析目标、数据源和交付物
2. 设计分析流程（数据采集 → 清洗 → 分析 → 可视化 → 报告）
3. 将项目按分析阶段和数据维度拆分为可执行的任务
4. 确保数据管道的上下游依赖关系正确
5. 为每个任务标注：任务名称、描述、角色等级(lite/pro/max)、执行批次号、涉及的资源范围、依赖关系

## 重要：这是一个全新项目

- 项目目录下目前没有任何数据文件或分析脚本，所有内容都需要从零创建
- 批次 1 必须包含数据源接入、环境搭建、基础数据清洗等前置任务
- 后续批次的分析任务依赖前面批次产出的清洗数据或中间结果，必须在 depends_on 中声明
- 任务描述中要说明输入数据来源（哪个前置任务产出）和预期输出文件
- 每个任务的 affected_resources 要列出该任务会创建的文件路径

任务拆分规则：
- 数据源接入和清洗为前置任务（批次 1）
- 不同维度的分析可以并行
- 汇总报告和可视化依赖所有分析任务
- affected_resources 使用相对路径（如 "reports/summary.md", "data/cleaned.csv"）

## 分段输出规则（必须严格遵守）

当任务数量较多时（超过 20 个），你**必须**分模块输出任务清单：
1. 先输出总体方案概要（分析框架、数据源）
2. 逐模块输出 JSON 任务清单，每个模块一个独立 JSON 代码块
3. 最后输出方案总结

JSON 任务清单格式示例：
{"tasks": [{"name": "模块-任务名", "description": "详细描述（含分析方法、数据源、预期输出）", "role_level": "pro", "batch_no": 1, "affected_resources": ["reports/summary.md"], "depends_on": ["数据清洗"]}]}

重要规则：
- 每个 JSON 块必须是完整的 {"tasks": [...]}，系统会自动合并所有代码块
- 任务名称全局唯一，跨模块依赖用完整任务名引用
- 如果输出被截断，系统会自动请求继续，从中断处接着输出即可`, projectName, projectDesc)
}
