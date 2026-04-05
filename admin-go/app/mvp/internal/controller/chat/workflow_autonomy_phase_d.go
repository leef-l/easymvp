package chat

import (
	"context"
	"encoding/json"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"

	v1 "easymvp/app/mvp/api/mvp/v1"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/workflow/orchestrator"
)

// MetaObservations 查询决策观测记录。
func (c *cWorkflow) MetaObservations(ctx context.Context, req *v1.WorkflowMetaObservationsReq) (res *v1.WorkflowMetaObservationsRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 100
	}

	records, err := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderDesc("created_at").
		Limit(limit).
		All()
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowMetaObservationsRes{Observations: records.List()}, nil
}

// MetaObservationStats 查询观测统计。
func (c *cWorkflow) MetaObservationStats(ctx context.Context, req *v1.WorkflowMetaObservationStatsReq) (res *v1.WorkflowMetaObservationStatsRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 总数
	total, _ := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).WhereNull("deleted_at").Count()

	// 按 outcome 分组
	outcomeRecords, _ := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).WhereNull("deleted_at").
		Fields("outcome, COUNT(*) as cnt").Group("outcome").All()

	outcomeDist := g.Map{}
	for _, r := range outcomeRecords {
		outcomeDist[r["outcome"].String()] = r["cnt"].Int()
	}

	// 按 decision_level 分��
	levelRecords, _ := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).WhereNull("deleted_at").
		Fields("decision_level, COUNT(*) as cnt").Group("decision_level").All()

	levelDist := g.Map{}
	for _, r := range levelRecords {
		levelDist[r["decision_level"].String()] = r["cnt"].Int()
	}

	// 人工干预率
	overrideCount, _ := g.DB().Model("mvp_observation_record").Ctx(ctx).
		Where("project_id", projectID).WhereNull("deleted_at").
		Where("human_override", 1).Count()

	overrideRate := 0.0
	if total > 0 {
		overrideRate = float64(overrideCount) / float64(total)
	}

	return &v1.WorkflowMetaObservationStatsRes{
		Stats: g.Map{
			"total":             total,
			"outcomeDistribution": outcomeDist,
			"levelDistribution":   levelDist,
			"humanOverrideCount":  overrideCount,
			"humanOverrideRate":   overrideRate,
		},
	}, nil
}

// MetaAssessment 查询最新评估结果。
func (c *cWorkflow) MetaAssessment(ctx context.Context, req *v1.WorkflowMetaAssessmentReq) (res *v1.WorkflowMetaAssessmentRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	assessor := orchestrator.GetMetaAssessor()
	result, err := assessor.GetLatest(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &v1.WorkflowMetaAssessmentRes{Assessment: g.Map{}}, nil
	}
	return &v1.WorkflowMetaAssessmentRes{Assessment: gconv.Map(result)}, nil
}

// MetaAssessmentHistory 查询评估历史。
func (c *cWorkflow) MetaAssessmentHistory(ctx context.Context, req *v1.WorkflowMetaAssessmentHistoryReq) (res *v1.WorkflowMetaAssessmentHistoryRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	limit := req.Limit
	if limit <= 0 {
		limit = 20
	}

	assessor := orchestrator.GetMetaAssessor()
	results, err := assessor.ListAssessments(ctx, projectID, limit)
	if err != nil {
		return nil, err
	}

	var list []g.Map
	for _, r := range results {
		list = append(list, gconv.Map(r))
	}
	return &v1.WorkflowMetaAssessmentHistoryRes{Assessments: list}, nil
}

// MetaRunAssessment 手动触发一次评估。
func (c *cWorkflow) MetaRunAssessment(ctx context.Context, req *v1.WorkflowMetaRunAssessmentReq) (res *v1.WorkflowMetaRunAssessmentRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	days := req.Days
	if days <= 0 {
		days = 7
	}

	now := gtime.Now()
	periodStart := gtime.New(now.AddDate(0, 0, -days))

	assessor := orchestrator.GetMetaAssessor()
	result, err := assessor.Assess(ctx, projectID, periodStart, now)
	if err != nil {
		return nil, err
	}

	// 自动生成调参建议
	tuner := orchestrator.GetMetaTuner()
	recommendations := tuner.GenerateRecommendations(ctx, result)
	if len(recommendations) > 0 {
		_ = tuner.SaveAndApply(ctx, recommendations)
	}

	return &v1.WorkflowMetaRunAssessmentRes{Assessment: gconv.Map(result)}, nil
}

// MetaRecommendations 查询调参建议。
func (c *cWorkflow) MetaRecommendations(ctx context.Context, req *v1.WorkflowMetaRecommendationsReq) (res *v1.WorkflowMetaRecommendationsRes, err error) {
	tuner := orchestrator.GetMetaTuner()
	results, err := tuner.ListAll(ctx, int64(req.ProjectID), req.Status, req.Limit)
	if err != nil {
		return nil, err
	}

	var list []g.Map
	for _, r := range results {
		list = append(list, gconv.Map(r))
	}
	return &v1.WorkflowMetaRecommendationsRes{Recommendations: list}, nil
}

// MetaApplyRecommendation 应用一条调参建议。
func (c *cWorkflow) MetaApplyRecommendation(ctx context.Context, req *v1.WorkflowMetaApplyRecommendationReq) (res *v1.WorkflowMetaApplyRecommendationRes, err error) {
	userID := middleware.GetUserID(ctx)

	tuner := orchestrator.GetMetaTuner()
	if err = tuner.ApplyRecommendation(ctx, int64(req.RecommendationID), userID); err != nil {
		return nil, err
	}
	return &v1.WorkflowMetaApplyRecommendationRes{}, nil
}

// MetaRejectRecommendation 驳回一条调参建议。
func (c *cWorkflow) MetaRejectRecommendation(ctx context.Context, req *v1.WorkflowMetaRejectRecommendationReq) (res *v1.WorkflowMetaRejectRecommendationRes, err error) {
	tuner := orchestrator.GetMetaTuner()
	if err = tuner.RejectRecommendation(ctx, int64(req.RecommendationID)); err != nil {
		return nil, err
	}
	return &v1.WorkflowMetaRejectRecommendationRes{}, nil
}

// MetaLearning 查询 EMA 学习记录。
func (c *cWorkflow) MetaLearning(ctx context.Context, req *v1.WorkflowMetaLearningReq) (res *v1.WorkflowMetaLearningRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	records, err := g.DB().Model("mvp_learning_record").Ctx(ctx).
		Where("project_id", projectID).
		WhereNull("deleted_at").
		OrderAsc("metric_key").
		All()
	if err != nil {
		return nil, err
	}

	var list []g.Map
	for _, r := range records {
		item := g.Map{
			"id":          r["id"].String(),
			"metricKey":   r["metric_key"].String(),
			"projectId":   r["project_id"].String(),
			"emaValue":    r["ema_value"].Float64(),
			"rawValue":    r["raw_value"].Float64(),
			"sampleCount": r["sample_count"].Int(),
			"lastUpdated": r["last_updated"].String(),
			"decayFactor": r["decay_factor"].Float64(),
		}
		list = append(list, item)
	}
	return &v1.WorkflowMetaLearningRes{Records: list}, nil
}

// unused import guard
var _ = json.Marshal
