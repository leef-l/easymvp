package service

import (
	"testing"

	workspacev1 "github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

func makeBenchProjects(n int) []projectHomeAggregate {
	projects := make([]projectHomeAggregate, 0, n)
	statuses := []string{"created", "planning", "compiled", "executing", "acceptance", "completed"}
	productionStatuses := []string{"pending", "functional_passed", "production_passed"}
	for i := 0; i < n; i++ {
		projects = append(projects, projectHomeAggregate{
			Project: entity.Projects{
				Id:                "proj-" + string(rune('a'+i%26)),
				Name:              "Project " + string(rune('0'+i%10)),
				Status:            statuses[i%len(statuses)],
				ProjectCategory:   "web_app",
				ProductionStatus:  productionStatuses[i%len(productionStatuses)],
			},
			BlockingCount: i % 3,
		})
	}
	return projects
}

func BenchmarkBuildWorkspaceProjectCards_100(b *testing.B) {
	projects := makeBenchProjects(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildWorkspaceProjectCards(projects)
	}
}

func BenchmarkBuildWorkspaceProjectCards_1000(b *testing.B) {
	projects := makeBenchProjects(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildWorkspaceProjectCards(projects)
	}
}

func BenchmarkBuildWorkspaceReleaseReadiness_100(b *testing.B) {
	projects := makeBenchProjects(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildWorkspaceReleaseReadiness(projects, 12)
	}
}

func BenchmarkBuildWorkspaceReleaseReadiness_1000(b *testing.B) {
	projects := makeBenchProjects(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildWorkspaceReleaseReadiness(projects, 12)
	}
}

func BenchmarkDeriveProjectProgress(b *testing.B) {
	statuses := []string{"created", "planning", "plan_draft", "plan_review", "compiled", "execution_ready", "executing", "acceptance", "completed", "unknown"}
	prodStatuses := []string{"", "pending", "functional_passed", "production_passed"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = deriveProjectProgress(statuses[i%len(statuses)], prodStatuses[i%len(prodStatuses)])
	}
}

func BenchmarkNormalizeProjectStage(b *testing.B) {
	inputs := []string{"", "  created  ", "planning", "compiled", "executing", "acceptance", "completed"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = normalizeProjectStage(inputs[i%len(inputs)])
	}
}

func makeBenchCoverageItems(n int) []entity.AcceptanceSurfaceCoverage {
	items := make([]entity.AcceptanceSurfaceCoverage, 0, n)
	statuses := []string{"covered", "partial", "missing"}
	for i := 0; i < n; i++ {
		items = append(items, entity.AcceptanceSurfaceCoverage{
			Surface:        "surface-" + string(rune('a'+i%26)),
			CoverageStatus: statuses[i%len(statuses)],
			EvidenceCount:  i % 10,
		})
	}
	return items
}

func BenchmarkBuildCoverageSummaryRows_100(b *testing.B) {
	items := makeBenchCoverageItems(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildCoverageSummaryRows(items)
	}
}

func BenchmarkBuildCoverageSummaryRows_1000(b *testing.B) {
	items := makeBenchCoverageItems(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildCoverageSummaryRows(items)
	}
}

func makeBenchJourneyItems(n int) []entity.AcceptanceJourneyCoverage {
	items := make([]entity.AcceptanceJourneyCoverage, 0, n)
	statuses := []string{"covered", "partial", "missing"}
	for i := 0; i < n; i++ {
		items = append(items, entity.AcceptanceJourneyCoverage{
			Journey:        "journey-" + string(rune('a'+i%26)),
			CoverageStatus: statuses[i%len(statuses)],
			EvidenceCount:  i % 10,
		})
	}
	return items
}

func BenchmarkBuildJourneySummaryRows_100(b *testing.B) {
	items := makeBenchJourneyItems(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildJourneySummaryRows(items)
	}
}

func BenchmarkBuildJourneySummaryRows_1000(b *testing.B) {
	items := makeBenchJourneyItems(1000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = buildJourneySummaryRows(items)
	}
}

func BenchmarkBuildWorkspaceHomeDataOverview(b *testing.B) {
	projects := makeBenchProjects(100)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		active := 0
		blocked := 0
		for _, item := range projects {
			if !isFinishedProjectStatus(item.Project.Status) {
				active++
			}
			if item.BlockingCount > 0 {
				blocked++
			}
		}
		_ = workspacev1.HomeOverview{
			TotalProjects:   len(projects),
			ActiveProjects:  active,
			BlockedProjects: blocked,
		}
	}
}
