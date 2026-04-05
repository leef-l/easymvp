package autonomy

import (
	"context"
	"strings"

	"easymvp/app/mvp/internal/consts"
)

// EngineSelectionStrategy 执行器选择策略（B3）。
//
// 在任务执行前，根据任务特征（role_type、affected_resources 数量、历史成功率、任务描述）
// 推荐最优执行器（aider / claude_code / openhands / chat）。
//
// 触发时机：task.pending（任务即将开始），也可由 task.failed 触发换引擎重试。
type EngineSelectionStrategy struct{}

func NewEngineSelectionStrategy() *EngineSelectionStrategy {
	return &EngineSelectionStrategy{}
}

func (s *EngineSelectionStrategy) Name() string { return "engine_selection" }

func (s *EngineSelectionStrategy) Priority() int { return 70 }

func (s *EngineSelectionStrategy) Applicable(sit *Situation, trigger string) bool {
	switch trigger {
	case consts.TriggerTaskFailed, consts.TriggerTaskRetryExhausted:
		// 任务失败时，判断是否值得换引擎重试
		return sit != nil && sit.Health != nil && sit.Health.RetryCount < 2
	}
	return false
}

func (s *EngineSelectionStrategy) Evaluate(ctx context.Context, sit *Situation, req *DecisionRequest) *ActionPlan {
	if sit == nil || sit.Health == nil {
		return nil
	}

	// 从 TriggerContext 中获取任务信息
	taskInfo := s.extractTaskInfo(req)
	if taskInfo == nil {
		return nil
	}

	// 已用引擎（从 context 获取，避免重复推荐相同引擎）
	currentEngine := s.extractCurrentEngine(req)

	rec := s.recommend(taskInfo, currentEngine)
	if rec == nil {
		return nil
	}

	// 若推荐引擎与当前引擎相同，无需切换
	if rec.EngineType == currentEngine {
		return nil
	}

	level := consts.DecisionLevelA
	confidence := rec.Confidence
	if confidence < 0.65 {
		level = consts.DecisionLevelB
	}

	return &ActionPlan{
		StrategyName:    s.Name(),
		Trigger:         req.TriggerSource,
		DecisionLevel:   level,
		ActionType:      consts.ActionTypeSwitchExecutor,
		TargetID:        req.DomainTaskID,
		Reasoning:       rec.Reason,
		RollbackAction:  consts.ActionTypePauseWorkflow,
		ExpectedOutcome: "切换至 " + rec.EngineType + " 引擎后任务成功完成",
		Parameters: map[string]interface{}{
			"engine_type": rec.EngineType,
			"task_id":     req.DomainTaskID,
		},
		Meta: &DecisionMeta{
			Confidence:          confidence,
			EvidenceSufficiency: 0.7,
			Reversibility:       "full",
			BlastRadius:         "task",
		},
	}
}

// recommend 根据任务特征推荐执行器。
func (s *EngineSelectionStrategy) recommend(task *taskFeature, currentEngine string) *EngineRecommendation {
	// 非实现者角色 → chat
	if task.roleType != consts.RoleTypeImplementer {
		if currentEngine == "chat" {
			return nil
		}
		return &EngineRecommendation{EngineType: "chat", Confidence: 0.95, Reason: "非实现者角色（" + task.roleType + "）使用 chat 引擎"}
	}

	desc := strings.ToLower(task.description + " " + task.taskName)
	resourceCount := task.resourceCount

	// 规则 1：大范围多文件重构 → claude_code
	if resourceCount >= 5 || containsKeyword(desc, "重构", "refactor", "迁移", "migrate", "架构调整") {
		return &EngineRecommendation{EngineType: "claude_code", Confidence: 0.80,
			Reason: "多文件/重构类任务，推荐全局理解能力强的 claude_code"}
	}

	// 规则 2：需要运行测试/部署 → openhands
	if containsKeyword(desc, "测试", "test", "部署", "deploy", "运行", "启动", "安装依赖") {
		return &EngineRecommendation{EngineType: "openhands", Confidence: 0.75,
			Reason: "需要运行环境的任务，推荐有沙箱的 openhands"}
	}

	// 规则 3：前端 UI → gemini_cli
	if containsKeyword(desc, "ui", "界面", "前端", "组件", "样式", "css", "html", "vue", "react") {
		return &EngineRecommendation{EngineType: "gemini_cli", Confidence: 0.65,
			Reason: "前端 UI 类任务，推荐多模态理解的 gemini_cli"}
	}

	// 规则 4：文档/分析类 → chat
	if containsKeyword(desc, "文档", "分析", "设计", "评审", "doc", "review", "analysis") {
		return &EngineRecommendation{EngineType: "chat", Confidence: 0.80,
			Reason: "文档/分析类任务，使用 chat 引擎"}
	}

	// 规则 5：历史失败率高且资源少 → claude_code（更强的理解能力）
	if task.recentFailureRate >= 0.5 && resourceCount <= 3 {
		return &EngineRecommendation{EngineType: "claude_code", Confidence: 0.70,
			Reason: "历史失败率较高，升级至 claude_code 提升成功率"}
	}

	// 默认：单文件精确编辑 → aider
	return &EngineRecommendation{EngineType: "aider", Confidence: 0.60,
		Reason: "默认推荐 aider（专注、快速、成本低）"}
}

// taskFeature 从 TriggerContext 提取的任务特征。
type taskFeature struct {
	taskName          string
	description       string
	roleType          string
	resourceCount     int
	recentFailureRate float64
}

func (s *EngineSelectionStrategy) extractTaskInfo(req *DecisionRequest) *taskFeature {
	if req.TriggerContext == nil {
		return nil
	}
	tc := req.TriggerContext
	f := &taskFeature{}
	if v, ok := tc["task_name"].(string); ok {
		f.taskName = v
	}
	if v, ok := tc["description"].(string); ok {
		f.description = v
	}
	if v, ok := tc["role_type"].(string); ok {
		f.roleType = v
	}
	if v, ok := tc["resource_count"].(int); ok {
		f.resourceCount = v
	} else if v, ok := tc["resource_count"].(float64); ok {
		f.resourceCount = int(v)
	}
	if v, ok := tc["recent_failure_rate"].(float64); ok {
		f.recentFailureRate = v
	}
	// 若无 role_type，默认为 implementer（触发场景一般是实现任务）
	if f.roleType == "" {
		f.roleType = consts.RoleTypeImplementer
	}
	return f
}

func (s *EngineSelectionStrategy) extractCurrentEngine(req *DecisionRequest) string {
	if req.TriggerContext == nil {
		return ""
	}
	if v, ok := req.TriggerContext["engine_type"].(string); ok {
		return v
	}
	return ""
}

func containsKeyword(s string, patterns ...string) bool {
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}
