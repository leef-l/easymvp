package autonomy

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

// DomainTaskInfo 任务信息（用于执行器选择）。
type DomainTaskInfo struct {
	TaskID            int64
	TaskName          string
	Description       string
	RoleType          string
	AffectedResources []string
	ProjectType       string
}

// EngineSelector 执行器智能选择器（纯规则版）。
type EngineSelector struct{}

// NewEngineSelector 创建选择器。
func NewEngineSelector() *EngineSelector { return &EngineSelector{} }

// Recommend 为任务推荐最佳执行器。
func (s *EngineSelector) Recommend(ctx context.Context, task *DomainTaskInfo) *EngineRecommendation {
	// 非实现者角色 → chat
	if task.RoleType != "implementer" {
		return &EngineRecommendation{
			EngineType: "chat",
			Confidence: 0.95,
			Reason:     "非实现者角色（" + task.RoleType + "）使用 chat 引擎",
		}
	}

	desc := strings.ToLower(task.Description + " " + task.TaskName)
	resourceCount := len(task.AffectedResources)

	// 规则 1：大范围多文件重构 → claude_code
	if resourceCount >= 5 || containsAny(desc, []string{"重构", "refactor", "迁移", "migrate", "架构调整"}) {
		return &EngineRecommendation{
			EngineType: "claude_code",
			Confidence: 0.75,
			Reason:     "多文件/重构类任务，推荐全局理解能力强的 claude_code",
		}
	}

	// 规则 2：需要运行测试/部署 → openhands
	if containsAny(desc, []string{"测试", "test", "部署", "deploy", "运行", "启动", "安装依赖"}) {
		return &EngineRecommendation{
			EngineType: "openhands",
			Confidence: 0.7,
			Reason:     "需要运行环境的任务，推荐有沙箱的 openhands",
		}
	}

	// 规则 3：前端 UI → gemini_cli（多模态）
	if containsAny(desc, []string{"ui", "界面", "前端", "组件", "样式", "css", "html", "vue", "react"}) {
		return &EngineRecommendation{
			EngineType: "gemini_cli",
			Confidence: 0.6,
			Reason:     "前端 UI 类任务，推荐多模态理解的 gemini_cli",
		}
	}

	// 规则 4：文档/分析类 → chat
	if containsAny(desc, []string{"文档", "分析", "设计", "评审", "doc", "review", "analysis"}) {
		return &EngineRecommendation{
			EngineType: "chat",
			Confidence: 0.8,
			Reason:     "文档/分析类任务，使用 chat 引擎",
		}
	}

	// 默认：单文件精确编辑 → aider（成本低、速度快）
	g.Log().Debugf(ctx, "[EngineSelector] 默认推荐 aider: task=%d", task.TaskID)
	return &EngineRecommendation{
		EngineType: "aider",
		Confidence: 0.6,
		Reason:     "默认推荐 aider（专注、快速、成本低）",
	}
}

func containsAny(s string, patterns []string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
