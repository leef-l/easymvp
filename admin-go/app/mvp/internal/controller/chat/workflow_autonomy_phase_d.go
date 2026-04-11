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
	"easymvp/app/mvp/internal/workflow/repo"
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

	records, err := repo.NewObservationRecordRepo().ListByProject(ctx, projectID, limit)
	if err != nil {
		return nil, err
	}
	return &v1.WorkflowMetaObservationsRes{Observations: records}, nil
}

// MetaObservationStats 查询观测统计。
func (c *cWorkflow) MetaObservationStats(ctx context.Context, req *v1.WorkflowMetaObservationStatsReq) (res *v1.WorkflowMetaObservationStatsRes, err error) {
	projectID := int64(req.ProjectID)
	if err = checkProjectOwnership(ctx, projectID); err != nil {
		return nil, err
	}

	// 总数
	observationRepo := repo.NewObservationRecordRepo()

	total, err := observationRepo.CountByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	// 按 outcome 分组
	outcomeRecords, err := observationRepo.CountGroupByField(ctx, projectID, "outcome")
	if err != nil {
		return nil, err
	}

	outcomeDist := g.Map{}
	for _, r := range outcomeRecords {
		outcomeDist[mapString(r, "outcome")] = mapInt(r, "cnt")
	}

	// 按 decision_level 分��
	levelRecords, err := observationRepo.CountGroupByField(ctx, projectID, "decision_level")
	if err != nil {
		return nil, err
	}

	levelDist := g.Map{}
	for _, r := range levelRecords {
		levelDist[mapString(r, "decision_level")] = mapInt(r, "cnt")
	}

	// 人工干预率
	overrideCount, err := observationRepo.CountHumanOverrideByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	overrideRate := 0.0
	if total > 0 {
		overrideRate = float64(overrideCount) / float64(total)
	}

	return &v1.WorkflowMetaObservationStatsRes{
		Stats: g.Map{
			"total":               total,
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

	records, err := repo.NewLearningRecordRepo().ListByProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	var list []g.Map
	for _, r := range records {
		item := g.Map{
			"id":          mapString(r, "id"),
			"metricKey":   mapString(r, "metric_key"),
			"projectId":   mapString(r, "project_id"),
			"emaValue":    g.NewVar(r["ema_value"]).Float64(),
			"rawValue":    g.NewVar(r["raw_value"]).Float64(),
			"sampleCount": mapInt(r, "sample_count"),
			"lastUpdated": mapString(r, "last_updated"),
			"decayFactor": g.NewVar(r["decay_factor"]).Float64(),
		}
		list = append(list, item)
	}
	return &v1.WorkflowMetaLearningRes{Records: list}, nil
}

// unused import guard
var _ = json.Marshal
