package engine

import (
	"context"
	"testing"
)

func TestFastExtractWithReportDetectsMissingChunkAndCountMismatch(t *testing.T) {
	t.Parallel()

	parser := &TaskParser{}
	content := "```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":3,\"chunk_index\":1,\"chunk_total\":3,\"is_final\":false},\"tasks\":[{\"name\":\"初始化前端\",\"description\":\"创建前端骨架\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"frontend/package.json\"],\"depends_on\":[]}]}\n" +
		"```\n\n```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":3,\"chunk_index\":3,\"chunk_total\":3,\"is_final\":true},\"tasks\":[{\"name\":\"初始化后端\",\"description\":\"创建后端骨架\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"backend/go.mod\"],\"depends_on\":[]}]}\n" +
		"```"

	tasks, report, err := parser.FastExtractWithReport(context.Background(), content, "software_dev")
	if err != nil {
		t.Fatalf("FastExtractWithReport() error = %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 parsed tasks, got %d", len(tasks))
	}
	if report == nil || !report.HasBlockingIssue() {
		t.Fatalf("expected blocking report, got %+v", report)
	}
	if len(report.MissingChunkIndexes) != 1 || report.MissingChunkIndexes[0] != 2 {
		t.Fatalf("unexpected missing chunks: %+v", report)
	}
	if !report.declaredTotalMismatch() {
		t.Fatalf("expected declared_total mismatch, got %+v", report)
	}
}

func TestFastExtractWithReportReplacesWholeChunkByLatestVersion(t *testing.T) {
	t.Parallel()

	parser := &TaskParser{}
	content := "```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":2,\"chunk_index\":1,\"chunk_total\":2,\"is_final\":false},\"tasks\":[{\"name\":\"旧前端初始化\",\"description\":\"旧块里的任务\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"frontend/package.json\"],\"depends_on\":[]},{\"name\":\"会被移除的任务\",\"description\":\"旧块多余任务\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"frontend/README.md\"],\"depends_on\":[]}]}\n" +
		"```\n\n```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":2,\"chunk_index\":2,\"chunk_total\":2,\"is_final\":true},\"tasks\":[{\"name\":\"后端初始化\",\"description\":\"后端骨架\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"backend/go.mod\"],\"depends_on\":[]}]}\n" +
		"```\n\n```json\n" +
		"{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":2,\"chunk_index\":1,\"chunk_total\":2,\"is_final\":false},\"tasks\":[{\"name\":\"前端初始化\",\"description\":\"修正后的完整块\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"frontend/package.json\"],\"depends_on\":[]}]}\n" +
		"```"

	tasks, report, err := parser.FastExtractWithReport(context.Background(), content, "software_dev")
	if err != nil {
		t.Fatalf("FastExtractWithReport() error = %v", err)
	}
	if report == nil || report.HasBlockingIssue() {
		t.Fatalf("expected non-blocking report, got %+v", report)
	}
	if len(report.ReplacedChunkIndexes) != 1 || report.ReplacedChunkIndexes[0] != 1 {
		t.Fatalf("unexpected replaced chunks: %+v", report)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks after replacement, got %d", len(tasks))
	}
	if tasks[0].Name != "前端初始化" || tasks[1].Name != "后端初始化" {
		t.Fatalf("unexpected task order/content: %+v", tasks)
	}
}

func TestFastExtractWithReportPreservesPlanMetaFromRawJSONObject(t *testing.T) {
	t.Parallel()

	parser := &TaskParser{}
	content := "{\"plan_meta\":{\"plan_id\":\"demo\",\"declared_total\":3,\"chunk_index\":1,\"chunk_total\":3,\"is_final\":false},\"tasks\":[{\"name\":\"初始化前端\",\"description\":\"创建前端骨架\",\"role_level\":\"pro\",\"batch_no\":1,\"affected_resources\":[\"frontend/package.json\"],\"depends_on\":[]}]}"

	tasks, report, err := parser.FastExtractWithReport(context.Background(), content, "software_dev")
	if err != nil {
		t.Fatalf("FastExtractWithReport() error = %v", err)
	}
	if len(tasks) != 1 {
		t.Fatalf("expected 1 parsed task, got %d", len(tasks))
	}
	if report == nil || !report.HasBlockingIssue() {
		t.Fatalf("expected blocking report, got %+v", report)
	}
	if report.ChunkTotal == nil || *report.ChunkTotal != 3 {
		t.Fatalf("unexpected chunk_total in report: %+v", report)
	}
	if len(report.MissingChunkIndexes) != 2 || report.MissingChunkIndexes[0] != 2 || report.MissingChunkIndexes[1] != 3 {
		t.Fatalf("unexpected missing chunks: %+v", report)
	}
}
