package consts

// 项目状态
const (
	ProjectStatusDesigning = "designing" // 设计中（架构师对话）
	ProjectStatusRunning   = "running"   // 执行中（调度器调度）
	ProjectStatusPaused    = "paused"    // 已暂停（任务失败或手动暂停）
	ProjectStatusCompleted = "completed" // 已完成
)

// 项目分类
const (
	ProjectCategorySoftware = "软件开发"
	ProjectCategoryData     = "数据分析"
	ProjectCategoryProduct  = "产品设计"
)

// AllProjectStatuses 所有有效的项目状态
var AllProjectStatuses = []string{
	ProjectStatusDesigning,
	ProjectStatusRunning,
	ProjectStatusPaused,
	ProjectStatusCompleted,
}
