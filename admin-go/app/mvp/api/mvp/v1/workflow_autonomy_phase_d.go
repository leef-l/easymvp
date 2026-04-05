package v1

import (
	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// ==================== 元认知观测 ====================

// WorkflowMetaObservationsReq 查询观测记录。
type WorkflowMetaObservationsReq struct {
	g.Meta    `path:"/workflow/meta/observations" method:"get" tags:"自治L7" summary:"查询决策观测记录"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Limit     int                 `json:"limit" dc:"数量限制"`
}

type WorkflowMetaObservationsRes struct {
	g.Meta       `mime:"application/json"`
	Observations []g.Map `json:"observations"`
}

// WorkflowMetaObservationStatsReq 查询观测统计。
type WorkflowMetaObservationStatsReq struct {
	g.Meta    `path:"/workflow/meta/observation-stats" method:"get" tags:"自治L7" summary:"查询观测统计"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

type WorkflowMetaObservationStatsRes struct {
	g.Meta `mime:"application/json"`
	Stats  g.Map `json:"stats"`
}

// ==================== 评估结果 ====================

// WorkflowMetaAssessmentReq 查询最新评估结果。
type WorkflowMetaAssessmentReq struct {
	g.Meta    `path:"/workflow/meta/assessment" method:"get" tags:"自治L7" summary:"查询最新评估结果"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

type WorkflowMetaAssessmentRes struct {
	g.Meta     `mime:"application/json"`
	Assessment g.Map `json:"assessment"`
}

// WorkflowMetaAssessmentHistoryReq 查询评估历史。
type WorkflowMetaAssessmentHistoryReq struct {
	g.Meta    `path:"/workflow/meta/assessment-history" method:"get" tags:"自治L7" summary:"查询评估历史"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Limit     int                 `json:"limit" dc:"数量限制"`
}

type WorkflowMetaAssessmentHistoryRes struct {
	g.Meta      `mime:"application/json"`
	Assessments []g.Map `json:"assessments"`
}

// WorkflowMetaRunAssessmentReq 手动触发一次评估。
type WorkflowMetaRunAssessmentReq struct {
	g.Meta    `path:"/workflow/meta/run-assessment" method:"post" tags:"自治L7" summary:"手动触发评估"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
	Days      int                 `json:"days" dc:"评估周期(天,默认7)"`
}

type WorkflowMetaRunAssessmentRes struct {
	g.Meta     `mime:"application/json"`
	Assessment g.Map `json:"assessment"`
}

// ==================== 调参建议 ====================

// WorkflowMetaRecommendationsReq 查询调参建议。
type WorkflowMetaRecommendationsReq struct {
	g.Meta    `path:"/workflow/meta/recommendations" method:"get" tags:"自治L7" summary:"查询调参建议"`
	ProjectID snowflake.JsonInt64 `json:"projectID" dc:"项目ID(0=全局)"`
	Status    string              `json:"status" dc:"状态过滤(pending/applied/rejected)"`
	Limit     int                 `json:"limit" dc:"数量限制"`
}

type WorkflowMetaRecommendationsRes struct {
	g.Meta          `mime:"application/json"`
	Recommendations []g.Map `json:"recommendations"`
}

// WorkflowMetaApplyRecommendationReq 应用一条建议。
type WorkflowMetaApplyRecommendationReq struct {
	g.Meta           `path:"/workflow/meta/apply-recommendation" method:"post" tags:"自治L7" summary:"应用调参建议"`
	RecommendationID snowflake.JsonInt64 `json:"recommendationID" v:"required" dc:"建议ID"`
}

type WorkflowMetaApplyRecommendationRes struct {
	g.Meta `mime:"application/json"`
}

// WorkflowMetaRejectRecommendationReq 驳回一条建议。
type WorkflowMetaRejectRecommendationReq struct {
	g.Meta           `path:"/workflow/meta/reject-recommendation" method:"post" tags:"自治L7" summary:"驳回调参建议"`
	RecommendationID snowflake.JsonInt64 `json:"recommendationID" v:"required" dc:"建议ID"`
}

type WorkflowMetaRejectRecommendationRes struct {
	g.Meta `mime:"application/json"`
}

// ==================== 学习记录 ====================

// WorkflowMetaLearningReq 查询学习记录。
type WorkflowMetaLearningReq struct {
	g.Meta    `path:"/workflow/meta/learning" method:"get" tags:"自治L7" summary:"查询EMA学习记录"`
	ProjectID snowflake.JsonInt64 `json:"projectID" v:"required" dc:"项目ID"`
}

type WorkflowMetaLearningRes struct {
	g.Meta  `mime:"application/json"`
	Records []g.Map `json:"records"`
}
