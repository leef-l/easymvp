package engine

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

func TestNormalizeTasksDropsExplorationPlaceholder(t *testing.T) {
	parser := &TaskParser{}

	tasks := parser.normalizeTasks(context.Background(), []ArchitectTask{
		{
			Name:        "查看项目目录结构",
			Description: "查看 /www/wwwroot/project/easymvp/admin-go 目录结构，了解现有代码和模块组织",
			RoleLevel:   "lite",
			BatchNo:     1,
		},
	}, "软件开发")

	if len(tasks) != 0 {
		t.Fatalf("expected placeholder task to be dropped, got %d tasks", len(tasks))
	}
}

func TestNormalizeTasksKeepsDeliverableTask(t *testing.T) {
	parser := &TaskParser{}

	tasks := parser.normalizeTasks(context.Background(), []ArchitectTask{
		{
			Name:        "输出后端主链修复方案",
			Description: "梳理 Workflow V2 后端主链阻断点，并输出修复方案与接口回归清单",
			RoleLevel:   "lite",
			BatchNo:     1,
		},
	}, "软件开发")

	if len(tasks) != 1 {
		t.Fatalf("expected deliverable task to be kept, got %d tasks", len(tasks))
	}
	if tasks[0].RoleType != "implementer" {
		t.Fatalf("expected default role_type implementer, got %q", tasks[0].RoleType)
	}
}

func TestNormalizeTasksWithReportPrefersLatestDuplicate(t *testing.T) {
	parser := &TaskParser{}

	tasks, report := parser.normalizeTasksWithReport(context.Background(), []ArchitectTask{
		{
			Name:        "初始化前端",
			Description: "旧描述",
			RoleLevel:   "pro",
			BatchNo:     1,
		},
		{
			Name:        "初始化前端",
			Description: "新描述",
			RoleLevel:   "pro",
			BatchNo:     1,
		},
	}, "软件开发")

	if len(tasks) != 1 {
		t.Fatalf("expected one normalized task, got %d", len(tasks))
	}
	if tasks[0].Description != "新描述" {
		t.Fatalf("expected latest duplicate to win, got %+v", tasks[0])
	}
	if len(report.DuplicateDropped) != 1 || report.DuplicateDropped[0] != "初始化前端" {
		t.Fatalf("unexpected duplicate report: %+v", report)
	}
}
