package acceptance

import (
	"context"
	"time"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/utility/snowflake"
)

// QualityFinding 统一质量发现，对应 mvp_quality_finding 表。
type QualityFinding struct {
	ID            int64      `json:"id"`
	WorkflowRunID int64      `json:"workflow_run_id"`
	StageCode     string     `json:"stage_code"`  // review/accept/rework
	Producer      string     `json:"producer"`    // hard_rule/llm_judge/experience_reviewer/human
	Severity      string     `json:"severity"`    // blocker/error/warn/info
	RuleCode      string     `json:"rule_code"`   // 规则代码或 llm_auto
	TargetType    string     `json:"target_type"` // task/file/blueprint/project
	TargetID      string     `json:"target_id"`
	Message       string     `json:"message"`
	EvidenceJSON  string     `json:"evidence_json,omitempty"` // JSON 字符串，可为空
	Resolved      bool       `json:"resolved"`
	ResolvedBy    *int64     `json:"resolved_by,omitempty"`
	ResolvedAt    *time.Time `json:"resolved_at,omitempty"`
	CreatedBy     int64      `json:"created_by"`
	DeptID        int64      `json:"dept_id"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// WriteFinding 将质量发现写入 mvp_quality_finding 表。
// 若 f.ID 为 0，自动生成雪花 ID。
func WriteFinding(ctx context.Context, f *QualityFinding) error {
	if f.ID == 0 {
		f.ID = int64(snowflake.Generate())
	}
	now := time.Now()
	if f.CreatedAt.IsZero() {
		f.CreatedAt = now
	}
	f.UpdatedAt = now

	row := g.Map{
		"id":             f.ID,
		"workflow_run_id": f.WorkflowRunID,
		"stage_code":     f.StageCode,
		"producer":       f.Producer,
		"severity":       f.Severity,
		"rule_code":      f.RuleCode,
		"target_type":    f.TargetType,
		"target_id":      f.TargetID,
		"message":        f.Message,
		"evidence_json":  nil,
		"resolved":       f.Resolved,
		"resolved_by":    f.ResolvedBy,
		"resolved_at":    f.ResolvedAt,
		"created_by":     f.CreatedBy,
		"dept_id":        f.DeptID,
		"created_at":     f.CreatedAt,
		"updated_at":     f.UpdatedAt,
	}
	if f.EvidenceJSON != "" {
		row["evidence_json"] = f.EvidenceJSON
	}

	_, err := g.DB().Ctx(ctx).Model("mvp_quality_finding").Data(row).Insert()
	return err
}
