package engine

import (
	"fmt"
	"sort"
	"strings"
)

type ArchitectPlanMeta struct {
	PlanID        string `json:"plan_id,omitempty"`
	DeclaredTotal *int   `json:"declared_total,omitempty"`
	ChunkIndex    *int   `json:"chunk_index,omitempty"`
	ChunkTotal    *int   `json:"chunk_total,omitempty"`
	IsFinal       *bool  `json:"is_final,omitempty"`
}

type TaskNormalizationReport struct {
	DuplicateDropped   []string
	PlaceholderDropped []string
	EmptyNameDropped   int
}

type ArchitectPlanParseReport struct {
	PlanID                string
	DeclaredTotal         *int
	ChunkTotal            *int
	ChunkIndexes          []int
	MissingChunkIndexes   []int
	ReplacedChunkIndexes  []int
	MetaConflicts         []string
	RawTaskCount          int
	NormalizedTaskCount   int
	CodeBlockCount        int
	ParsedBlockCount      int
	InvalidCodeBlockCount int
	MixedChunkEncoding    bool
	MissingPlanMeta       bool
	DuplicateDropped      []string
	PlaceholderDropped    []string
	EmptyNameDropped      int
	UsedAIRecovery        bool
}

func (r *ArchitectPlanParseReport) mergePlanMeta(meta *ArchitectPlanMeta) {
	if r == nil || meta == nil {
		return
	}

	if planID := strings.TrimSpace(meta.PlanID); planID != "" {
		if r.PlanID != "" && r.PlanID != planID {
			r.MetaConflicts = append(r.MetaConflicts, fmt.Sprintf("plan_id 不一致: %s / %s", r.PlanID, planID))
		} else if r.PlanID == "" {
			r.PlanID = planID
		}
	}

	if meta.DeclaredTotal != nil && *meta.DeclaredTotal > 0 {
		if r.DeclaredTotal != nil && *r.DeclaredTotal != *meta.DeclaredTotal {
			r.MetaConflicts = append(r.MetaConflicts, fmt.Sprintf("declared_total 不一致: %d / %d", *r.DeclaredTotal, *meta.DeclaredTotal))
		} else if r.DeclaredTotal == nil {
			value := *meta.DeclaredTotal
			r.DeclaredTotal = &value
		}
	}

	if meta.ChunkTotal != nil && *meta.ChunkTotal > 0 {
		if r.ChunkTotal != nil && *r.ChunkTotal != *meta.ChunkTotal {
			r.MetaConflicts = append(r.MetaConflicts, fmt.Sprintf("chunk_total 不一致: %d / %d", *r.ChunkTotal, *meta.ChunkTotal))
		} else if r.ChunkTotal == nil {
			value := *meta.ChunkTotal
			r.ChunkTotal = &value
		}
	}

	if meta.ChunkIndex != nil && *meta.ChunkIndex > 0 {
		r.ChunkIndexes = append(r.ChunkIndexes, *meta.ChunkIndex)
	}
}

func (r *ArchitectPlanParseReport) applyNormalization(report *TaskNormalizationReport, normalizedCount int) {
	if r == nil {
		return
	}
	r.NormalizedTaskCount = normalizedCount
	if report == nil {
		return
	}
	r.DuplicateDropped = append(r.DuplicateDropped, report.DuplicateDropped...)
	r.PlaceholderDropped = append(r.PlaceholderDropped, report.PlaceholderDropped...)
	r.EmptyNameDropped += report.EmptyNameDropped
}

func (r *ArchitectPlanParseReport) finalize() {
	if r == nil {
		return
	}

	r.ChunkIndexes = dedupeSortedInts(r.ChunkIndexes)
	r.ReplacedChunkIndexes = dedupeSortedInts(r.ReplacedChunkIndexes)
	r.MissingChunkIndexes = nil
	r.DuplicateDropped = dedupeSortedStrings(r.DuplicateDropped)
	r.PlaceholderDropped = dedupeSortedStrings(r.PlaceholderDropped)
	r.MetaConflicts = dedupeSortedStrings(r.MetaConflicts)

	if len(r.ChunkIndexes) > 0 && r.ChunkTotal != nil && *r.ChunkTotal > 0 {
		seen := make(map[int]struct{}, len(r.ChunkIndexes))
		for _, index := range r.ChunkIndexes {
			seen[index] = struct{}{}
		}
		for index := 1; index <= *r.ChunkTotal; index++ {
			if _, ok := seen[index]; !ok {
				r.MissingChunkIndexes = append(r.MissingChunkIndexes, index)
			}
		}
	}

	if r.ParsedBlockCount > 1 && (r.DeclaredTotal == nil || r.ChunkTotal == nil || len(r.ChunkIndexes) == 0) {
		r.MissingPlanMeta = true
	}
	if len(r.ChunkIndexes) > 0 && (r.DeclaredTotal == nil || r.ChunkTotal == nil) {
		r.MissingPlanMeta = true
	}
}

func (r *ArchitectPlanParseReport) declaredTotalMismatch() bool {
	return r != nil && r.DeclaredTotal != nil && *r.DeclaredTotal > 0 && r.NormalizedTaskCount != *r.DeclaredTotal
}

func (r *ArchitectPlanParseReport) HasBlockingIssue() bool {
	if r == nil {
		return false
	}
	if r.MissingPlanMeta || r.MixedChunkEncoding || len(r.MetaConflicts) > 0 || len(r.MissingChunkIndexes) > 0 {
		return true
	}
	if r.declaredTotalMismatch() {
		return true
	}
	if r.InvalidCodeBlockCount > 0 && (r.ParsedBlockCount > 1 || len(r.ChunkIndexes) > 0 || r.DeclaredTotal != nil) {
		return true
	}
	return false
}

func (r *ArchitectPlanParseReport) IssueLines() []string {
	if r == nil {
		return nil
	}

	lines := make([]string, 0, 8)
	if r.MissingPlanMeta {
		lines = append(lines, "多段任务清单缺少完整 plan_meta（至少需要 declared_total、chunk_index、chunk_total）")
	}
	if r.MixedChunkEncoding {
		lines = append(lines, "同一轮回复混用了带 chunk_index 和不带 chunk_index 的 JSON 块")
	}
	if len(r.MetaConflicts) > 0 {
		lines = append(lines, "plan_meta 存在冲突："+strings.Join(r.MetaConflicts, "；"))
	}
	if len(r.MissingChunkIndexes) > 0 {
		lines = append(lines, fmt.Sprintf("缺少分段: %s", formatIntList(r.MissingChunkIndexes)))
	}
	if r.InvalidCodeBlockCount > 0 && (r.ParsedBlockCount > 1 || len(r.ChunkIndexes) > 0 || r.DeclaredTotal != nil) {
		lines = append(lines, fmt.Sprintf("有 %d 个 JSON 代码块解析失败，已无法确认任务是否完整", r.InvalidCodeBlockCount))
	}
	if r.declaredTotalMismatch() {
		lines = append(lines, fmt.Sprintf("declared_total=%d，但标准化后仅保留 %d 个任务", *r.DeclaredTotal, r.NormalizedTaskCount))
	}
	if len(r.DuplicateDropped) > 0 {
		lines = append(lines, fmt.Sprintf("重复任务名已按“后发覆盖前发”收敛: %s", formatNameList(r.DuplicateDropped, 8)))
	}
	if len(r.PlaceholderDropped) > 0 {
		lines = append(lines, fmt.Sprintf("无交付占位任务已忽略: %s", formatNameList(r.PlaceholderDropped, 8)))
	}
	if r.EmptyNameDropped > 0 {
		lines = append(lines, fmt.Sprintf("空任务名条目已忽略: %d", r.EmptyNameDropped))
	}
	return lines
}

func (r *ArchitectPlanParseReport) Summary() string {
	if r == nil {
		return "未生成解析报告"
	}

	parts := []string{
		fmt.Sprintf("原始任务=%d", r.RawTaskCount),
		fmt.Sprintf("标准化后=%d", r.NormalizedTaskCount),
	}
	if r.DeclaredTotal != nil && *r.DeclaredTotal > 0 {
		parts = append(parts, fmt.Sprintf("declared_total=%d", *r.DeclaredTotal))
	}
	if len(r.ChunkIndexes) > 0 {
		parts = append(parts, fmt.Sprintf("chunks=%s", formatIntList(r.ChunkIndexes)))
	}
	if r.ChunkTotal != nil && *r.ChunkTotal > 0 {
		parts = append(parts, fmt.Sprintf("chunk_total=%d", *r.ChunkTotal))
	}
	if r.UsedAIRecovery {
		parts = append(parts, "source=ai_recovery")
	}
	return strings.Join(parts, "，")
}

func (r *ArchitectPlanParseReport) BuildContinuationPrompt() string {
	if r == nil {
		return ""
	}

	var builder strings.Builder
	builder.WriteString("继续，当前任务清单还不能落库，请只补发缺失或修正后的 JSON 代码块，不要重复自然语言说明。\n\n")
	builder.WriteString("当前解析结果：\n")
	for _, line := range r.IssueLines() {
		builder.WriteString("- ")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	builder.WriteString("\n必须使用以下格式：\n")
	builder.WriteString("{\"plan_meta\":{\"plan_id\":\"同一轮保持一致\",\"declared_total\":总任务数,\"chunk_index\":当前块序号,\"chunk_total\":总块数,\"is_final\":是否最后一块},\"tasks\":[...]}\n\n")
	builder.WriteString("补发规则：\n")
	builder.WriteString("- 如果只是补齐缺块，只发送缺失的 chunk_index 对应 JSON 块。\n")
	builder.WriteString("- 如果要修正已发送块，直接重发同一个 chunk_index 的完整 JSON 块，系统会以后发块覆盖前发块。\n")
	builder.WriteString("- declared_total 必须等于最终可执行任务总数，不要把“查看目录/环境分析”这类占位动作算进去。\n")
	builder.WriteString("- 任务名称必须全局唯一；如果重名，请重发修正后的完整块。\n")
	builder.WriteString("- 如果这是最后一次补发，请把 is_final 设为 true；否则设为 false。\n")
	return builder.String()
}

func dedupeSortedInts(values []int) []int {
	if len(values) == 0 {
		return nil
	}
	sort.Ints(values)
	result := make([]int, 0, len(values))
	for _, value := range values {
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func dedupeSortedStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	for i := range values {
		values[i] = strings.TrimSpace(values[i])
	}
	sort.Strings(values)
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if len(result) == 0 || result[len(result)-1] != value {
			result = append(result, value)
		}
	}
	return result
}

func formatIntList(values []int) string {
	if len(values) == 0 {
		return "[]"
	}
	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, fmt.Sprintf("%d", value))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func formatNameList(values []string, limit int) string {
	values = dedupeSortedStrings(values)
	if len(values) == 0 {
		return "[]"
	}
	if limit <= 0 {
		limit = len(values)
	}
	if len(values) <= limit {
		return strings.Join(values, "、")
	}
	return strings.Join(values[:limit], "、") + fmt.Sprintf(" 等 %d 项", len(values))
}
