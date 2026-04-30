package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	acceptancev1 "github.com/leef-l/easymvp/apps/core/api/acceptance/v1"
	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
	"github.com/leef-l/easymvp/apps/core/internal/repo"
)

const (
	acceptanceIssuesLimit    = 20
	acceptanceEvidenceLimit  = 12
	acceptanceRunsLoadLimit  = 6
	acceptanceAuditLogsLimit = 20
)

type acceptanceViewData struct {
	Overview           acceptancev1.AcceptanceOverview
	AcceptanceRun      acceptancev1.AcceptanceRunView
	CoverageMatrix     []acceptancev1.CoverageItem
	Issues             []acceptancev1.AcceptanceIssue
	EvidenceCards      []acceptancev1.EvidenceCard
	ReleaseGate        acceptancev1.ReleaseGateView
	VerificationResult acceptancev1.VerificationResultView
	CompletionVerdict  acceptancev1.CompletionVerdictView
	RuntimeEscalation  acceptancev1.RuntimeEscalationView
	FaultSummary       acceptancev1.FaultSummaryView
	RepairPlanDraft    acceptancev1.RepairPlanDraftSummary
	ContractGap        acceptancev1.ContractGapView
}

type acceptanceAggregate struct {
	Project                entity.Projects
	AcceptanceProfile      *acceptanceProfileRecord
	ProductionProfile      *productionAcceptanceProfileRecord
	AcceptanceRuns         []entity.AcceptanceRuns
	LatestAcceptanceRun    *entity.AcceptanceRuns
	PersistedVerdict       *acceptancev1.CompletionVerdictView
	SurfaceCoverage        []entity.AcceptanceSurfaceCoverage
	JourneyCoverage        []entity.AcceptanceJourneyCoverage
	Issues                 []entity.AcceptanceIssues
	EvidenceItems          []entity.EvidenceItems
	AuditLogs              []entity.AuditLogs
	Judgements             []entity.AcceptanceJudgements
	RepairDraft            *repairPlanDraftRecord
	RunBindings            []entity.BrainRunBindings
	ChannelAvailable       bool
	EnvironmentAvailable   bool
	ChannelUnavailableReason string
	// C-02: CompiledTaskContract carries the typed verification contract from the
	// compiled task when a specific task_id is associated with the acceptance run.
	CompiledTaskContract *braincontracts.VerificationContract
}

func loadAcceptanceViewData(ctx context.Context, projectID string) (*acceptanceViewData, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	aggregate, err := loadAcceptanceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	verificationResult := buildVerificationResultView(ctx, aggregate)
	return &acceptanceViewData{
		Overview:           buildAcceptanceOverview(aggregate),
		AcceptanceRun:      buildAcceptanceRunView(aggregate),
		CoverageMatrix:     buildAcceptanceCoverageMatrix(aggregate),
		Issues:             buildAcceptanceIssues(aggregate),
		EvidenceCards:      buildAcceptanceEvidenceCards(aggregate),
		ReleaseGate:        buildAcceptanceReleaseGate(aggregate),
		VerificationResult: verificationResult,
		CompletionVerdict:  buildCompletionVerdictView(aggregate),
		RuntimeEscalation:  buildRuntimeEscalationView(aggregate),
		FaultSummary:       buildFaultSummaryView(aggregate),
		RepairPlanDraft:    buildRepairPlanDraftSummary(aggregate),
		ContractGap:        buildContractGapView(verificationResult),
	}, nil
}

func loadAcceptanceAggregate(ctx context.Context, db *sql.DB, projectID string) (*acceptanceAggregate, error) {
	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	runs, err := listProjectAcceptanceRuns(ctx, db, projectID, acceptanceRunsLoadLimit)
	if err != nil {
		return nil, err
	}
	auditLogs, err := listProjectAuditLogs(ctx, db, projectID, acceptanceAuditLogsLimit)
	if err != nil {
		return nil, err
	}
	evidenceItems, err := listProjectEvidenceItems(ctx, projectID, acceptanceEvidenceLimit)
	if err != nil {
		return nil, err
	}
	runBindings, err := listProjectBrainRunBindings(ctx, db, projectID, acceptanceRunsLoadLimit)
	if err != nil {
		return nil, err
	}

	aggregate := &acceptanceAggregate{
		Project:              *project,
		AcceptanceRuns:       runs,
		EvidenceItems:        evidenceItems,
		AuditLogs:            auditLogs,
		RunBindings:          runBindings,
		ChannelAvailable:     true,
		EnvironmentAvailable: true,
	}
	aggregate.AcceptanceProfile, err = getLatestAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.ProductionProfile, err = getLatestProductionAcceptanceProfile(ctx, db, projectID)
	if err != nil {
		return nil, err
	}
	aggregate.RepairDraft, err = getLatestRepairDraftForProject(ctx, projectID)
	if err != nil {
		return nil, err
	}
	if len(runs) > 0 {
		aggregate.LatestAcceptanceRun = &runs[0]
		aggregate.SurfaceCoverage, err = listAcceptanceSurfaceCoverageByRun(ctx, runs[0].Id)
		if err != nil {
			return nil, err
		}
		aggregate.JourneyCoverage, err = listAcceptanceJourneyCoverageByRun(ctx, runs[0].Id)
		if err != nil {
			return nil, err
		}
		aggregate.Issues, err = listAcceptanceIssuesByRun(ctx, runs[0].Id, acceptanceIssuesLimit)
		if err != nil {
			return nil, err
		}
		aggregate.Judgements, err = listAcceptanceJudgementsByRun(ctx, runs[0].Id)
		if err != nil {
			return nil, err
		}
	}
	if aggregate.LatestAcceptanceRun == nil {
		aggregate.Issues, err = listProjectAcceptanceIssues(ctx, db, projectID, acceptanceIssuesLimit)
		if err != nil {
			return nil, err
		}
	}
	aggregate.PersistedVerdict, err = loadPersistedCompletionVerdict(ctx, projectID, acceptanceRunID(aggregate))
	if err != nil {
		return nil, err
	}

	// C-02: Load verification contract from the compiled task when the
	// acceptance run references a specific task_id. This allows the
	// accepting flow to route checks according to the contract.
	aggregate.CompiledTaskContract = loadCompiledTaskVerificationContract(ctx, db, aggregate)

	return aggregate, nil
}

func loadAcceptanceAggregateByRunID(ctx context.Context, db *sql.DB, projectID string, acceptanceRunID string) (*acceptanceAggregate, error) {
	aggregate, err := loadAcceptanceAggregate(ctx, db, projectID)
	if err != nil {
		return nil, err
	}

	acceptanceRunID = strings.TrimSpace(acceptanceRunID)
	if acceptanceRunID == "" {
		return nil, gerror.New("acceptance run id is required")
	}
	if aggregate.LatestAcceptanceRun != nil && aggregate.LatestAcceptanceRun.Id == acceptanceRunID {
		return aggregate, nil
	}

	run, err := getAcceptanceRunByID(ctx, acceptanceRunID)
	if err != nil {
		return nil, err
	}
	if run == nil || strings.TrimSpace(run.ProjectId) != strings.TrimSpace(projectID) {
		return nil, gerror.New("acceptance run does not belong to project")
	}

	aggregate.LatestAcceptanceRun = run
	aggregate.SurfaceCoverage, err = listAcceptanceSurfaceCoverageByRun(ctx, run.Id)
	if err != nil {
		return nil, err
	}
	aggregate.JourneyCoverage, err = listAcceptanceJourneyCoverageByRun(ctx, run.Id)
	if err != nil {
		return nil, err
	}
	aggregate.Issues, err = listAcceptanceIssuesByRun(ctx, run.Id, acceptanceIssuesLimit)
	if err != nil {
		return nil, err
	}
	aggregate.Judgements, err = listAcceptanceJudgementsByRun(ctx, run.Id)
	if err != nil {
		return nil, err
	}
	aggregate.PersistedVerdict, err = loadPersistedCompletionVerdict(ctx, projectID, run.Id)
	if err != nil {
		return nil, err
	}
	return aggregate, nil
}

// loadCompiledTaskVerificationContract attempts to load and parse the
// verification_contract_json from the CompiledTask associated with the
// latest acceptance run's task_id. Returns nil when no contract is found
// or parsing fails (non-fatal). C-02: enables accepting flow to route
// verification based on the task-level contract.
func loadCompiledTaskVerificationContract(ctx context.Context, db *sql.DB, aggregate *acceptanceAggregate) *braincontracts.VerificationContract {
	if aggregate == nil || aggregate.LatestAcceptanceRun == nil {
		return nil
	}
	taskID := strings.TrimSpace(aggregate.LatestAcceptanceRun.TaskId)
	if taskID == "" {
		return nil
	}

	// Try to resolve the compiled task by looking up domain_task → source_compiled_task_id.
	domainTask, err := getDomainTaskByID(ctx, db, taskID)
	if err != nil {
		g.Log().Debugf(ctx, "[loadCompiledTaskVerificationContract] getDomainTaskByID(%s) failed: %v", taskID, err)
	}
	compiledTaskID := taskID
	if domainTask != nil && strings.TrimSpace(domainTask.SourceCompiledTaskId) != "" {
		compiledTaskID = strings.TrimSpace(domainTask.SourceCompiledTaskId)
	}

	compiledTask, err := getCompiledTaskByID(ctx, db, compiledTaskID)
	if err != nil || compiledTask == nil {
		return nil
	}

	raw := strings.TrimSpace(compiledTask.VerificationContractJson)
	if raw == "" || raw == "{}" || raw == "null" {
		return nil
	}

	vc, err := braincontracts.ParseVerificationContract(json.RawMessage(raw))
	if err != nil {
		g.Log().Debugf(ctx, "[loadCompiledTaskVerificationContract] parse verification contract failed for task %s: %v", compiledTaskID, err)
		return nil
	}
	return vc
}

func loadPersistedCompletionVerdict(
	ctx context.Context,
	projectID string,
	acceptanceRunID string,
) (*acceptancev1.CompletionVerdictView, error) {
	return repo.LoadPersistedCompletionVerdict(ctx, projectID, acceptanceRunID)
}

func getAcceptanceRunByID(ctx context.Context, acceptanceRunID string) (*entity.AcceptanceRuns, error) {
	return repo.GetAcceptanceRunByID(ctx, acceptanceRunID)
}

func listAcceptanceIssuesByRun(ctx context.Context, acceptanceRunID string, limit int) ([]entity.AcceptanceIssues, error) {
	return repo.ListAcceptanceIssuesByRun(ctx, acceptanceRunID, limit)
}

func listAcceptanceSurfaceCoverageByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceSurfaceCoverage, error) {
	return repo.ListAcceptanceSurfaceCoverageByRun(ctx, acceptanceRunID)
}

func listAcceptanceJourneyCoverageByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceJourneyCoverage, error) {
	return repo.ListAcceptanceJourneyCoverageByRun(ctx, acceptanceRunID)
}

func listAcceptanceJudgementsByRun(ctx context.Context, acceptanceRunID string) ([]entity.AcceptanceJudgements, error) {
	return repo.ListAcceptanceJudgementsByRun(ctx, acceptanceRunID)
}

func listProjectEvidenceItems(ctx context.Context, projectID string, limit int) ([]entity.EvidenceItems, error) {
	return repo.ListProjectEvidenceItems(ctx, projectID, limit)
}

func buildAcceptanceRunView(data *acceptanceAggregate) acceptancev1.AcceptanceRunView {
	releaseGate := buildAcceptanceReleaseGate(data)
	completionVerdict := buildCompletionVerdictView(data)
	blockingCount := countBlockingIssues(data.Issues)
	if data.LatestAcceptanceRun == nil {
		return acceptancev1.AcceptanceRunView{
			ID:                     "",
			Status:                 deriveAcceptanceRunStatusFromProject(data.Project.Status),
			FunctionalStatus:       "pending",
			ProductionStatus:       normalizeProductionStatus(data.Project.ProductionStatus),
			ManualReleaseRequired:  false,
			LatestJudgementSummary: "",
			ReleaseGateStatus:      normalizeAcceptanceOverviewStatus(completionVerdict),
			NextAction:             firstNonEmpty(completionVerdict.NextAction, releaseGate.NextAction),
			BlockingIssueCount:     blockingCount,
		}
	}

	view := acceptancev1.AcceptanceRunView{
		ID:                    data.LatestAcceptanceRun.Id,
		TaskID:                strings.TrimSpace(data.LatestAcceptanceRun.TaskId),
		ProfileVersion:        strings.TrimSpace(data.LatestAcceptanceRun.ProfileVersion),
		Status:                normalizeAcceptanceRunStatus(data.LatestAcceptanceRun.Status),
		FunctionalStatus:      normalizeFunctionalStatus(data.LatestAcceptanceRun.FunctionalStatus),
		ProductionStatus:      normalizeProductionStatus(data.LatestAcceptanceRun.ProductionStatus),
		ManualReleaseRequired: data.LatestAcceptanceRun.ManualReleaseRequired == 1,
		FinishedAt:            strings.TrimSpace(data.LatestAcceptanceRun.FinishedAt),
		ReleaseGateStatus:     normalizeAcceptanceOverviewStatus(completionVerdict),
		NextAction:            firstNonEmpty(completionVerdict.NextAction, releaseGate.NextAction),
		BlockingIssueCount:    blockingCount,
	}
	if len(data.Judgements) > 0 {
		view.LatestJudgementKind = strings.TrimSpace(data.Judgements[0].JudgementKind)
		view.LatestJudgementResult = strings.TrimSpace(data.Judgements[0].JudgementResult)
		view.LatestJudgementSummary = strings.TrimSpace(data.Judgements[0].Summary)
		view.LatestJudgementAt = strings.TrimSpace(data.Judgements[0].CreatedAt)
	}
	return view
}

func buildAcceptanceOverview(data *acceptanceAggregate) acceptancev1.AcceptanceOverview {
	var (
		completionVerdict = buildCompletionVerdictView(data)
		coverageMatrix    = buildAcceptanceCoverageMatrix(data)
		acceptanceRun     = buildAcceptanceRunView(data)
		coveredItemCount  int
	)

	for _, item := range coverageMatrix {
		switch normalizeCoverageStatus(item.CoverageStatus) {
		case "covered", "ready", "passed":
			coveredItemCount++
		}
	}

	return acceptancev1.AcceptanceOverview{
		ProjectID:           data.Project.Id,
		CurrentStage:        normalizeProjectStage(data.Project.Status),
		OverallStatus:       normalizeAcceptanceOverviewStatus(completionVerdict),
		FunctionalStatus:    acceptanceRun.FunctionalStatus,
		ProductionStatus:    acceptanceRun.ProductionStatus,
		ReleaseGateStatus:   normalizeAcceptanceOverviewStatus(completionVerdict),
		NextAction:          firstNonEmpty(completionVerdict.NextAction, "inspect_acceptance"),
		BlockingIssueCount:  countBlockingIssues(data.Issues),
		CoveredItemCount:    coveredItemCount,
		RequiredItemCount:   len(coverageMatrix),
		EvidenceCardCount:   len(data.EvidenceItems),
		ManualReleaseNeeded: acceptanceRun.ManualReleaseRequired,
	}
}

func normalizeAcceptanceOverviewStatus(verdict acceptancev1.CompletionVerdictView) string {
	switch strings.TrimSpace(verdict.FinalStatus) {
	case "completed":
		return "completed"
	case "accepting", "reworking":
		return strings.TrimSpace(verdict.FinalStatus)
	}
	switch strings.TrimSpace(verdict.Decision) {
	case "complete":
		return "completed"
	case "rework":
		return "reworking"
	case "collect_evidence", "manual_review", "manual_checkpoint":
		return "accepting"
	default:
		return "pending"
	}
}

func buildVerificationResultView(ctx context.Context, data *acceptanceAggregate) acceptancev1.VerificationResultView {
	requiredChecks := []string{"acceptance_profile", "coverage_matrix", "evidence_artifacts"}
	requiredEvidence := []string(nil)
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.ReleaseRequirementsJSON); len(items) > 0 {
			requiredChecks = items
		}
		requiredEvidence = parseStringArrayJSON(data.ProductionProfile.RequiredEvidenceJSON)
	}

	// C-02: Merge requirements from the compiled task's verification_contract_json
	// so that the accepting flow honours the task-level contract.
	requiredChecks, requiredEvidence = mergeCompiledTaskContractRequirements(
		requiredChecks, requiredEvidence, data.CompiledTaskContract,
	)

	failedChecks := make([]string, 0, len(data.Issues))
	for _, issue := range data.Issues {
		label := firstNonEmpty(strings.TrimSpace(issue.IssueKind), strings.TrimSpace(issue.Summary), issue.Id)
		if label != "" {
			failedChecks = append(failedChecks, label)
		}
	}

	missingEvidence := make([]string, 0, len(requiredEvidence))
	evidenceReady := len(data.EvidenceItems)
	if len(requiredEvidence) > evidenceReady {
		missingEvidence = append(missingEvidence, requiredEvidence[evidenceReady:]...)
	}

	status := "pending"
	decision := "continue_verification"
	completed := false
	summary := "Verification is waiting for acceptance evidence."
	manualReviewRequired := false
	if len(failedChecks) > 0 {
		status = "failed"
		decision = "repair_required"
		summary = fmt.Sprintf("%d verification check(s) failed", len(failedChecks))
	} else if len(missingEvidence) > 0 {
		status = "incomplete"
		decision = "collect_evidence"
		summary = fmt.Sprintf("%d required evidence item(s) are still missing", len(missingEvidence))
	} else if data.LatestAcceptanceRun != nil {
		status = "passed"
		decision = "ready_for_completion"
		completed = true
		summary = "Verification requirements are currently satisfied."
	}
	if data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1 {
		manualReviewRequired = true
	}
	for _, issue := range data.Issues {
		if issue.Blocking == 1 {
			manualReviewRequired = true
			break
		}
	}

	// C-02: Use the compiled task contract's preferred_channel when available.
	preferredChannel := deriveVerificationCurrentChannel(manualReviewRequired)
	if data.CompiledTaskContract != nil && strings.TrimSpace(data.CompiledTaskContract.PreferredChannel) != "" {
		preferredChannel = strings.TrimSpace(data.CompiledTaskContract.PreferredChannel)
	}

	return acceptancev1.VerificationResultView{
		Status:           status,
		PreferredChannel: preferredChannel,
		RequiredChecks:   requiredChecks,
		RequiredEvidence: requiredEvidence,
		MissingEvidence:  missingEvidence,
		FailedChecks:     failedChecks,
		VerificationContractJSON: buildVerificationContractJSON(ctx, verificationContractParams{
			ProjectCategory:          strings.TrimSpace(data.Project.ProjectCategory),
			ProfileVersion:           firstNonEmpty(profileVersionForVerification(data.AcceptanceProfile, data.ProductionProfile), strings.TrimSpace(data.Project.Id)),
			RequiredChecks:           requiredChecks,
			RequiredEvidence:         requiredEvidence,
			ManualReviewRequired:     manualReviewRequired,
			ManualReviewSummary:      summary,
			ChannelUnavailable:       !data.ChannelAvailable,
			EnvironmentUnavailable:   !data.EnvironmentAvailable,
			ChannelUnavailableReason: data.ChannelUnavailableReason,
		}),
		SourceRunID:          firstNonEmpty(strings.TrimSpace(acceptanceRunID(data)), strings.TrimSpace(data.Project.Id)),
		UpdatedAt:            firstNonEmpty(latestAcceptanceUpdatedAt(data), strings.TrimSpace(data.Project.UpdatedAt)),
		Decision:             decision,
		Completed:            completed,
		Summary:              summary,
		ChannelAvailable:     data.ChannelAvailable,
		EnvironmentAvailable: data.EnvironmentAvailable,
	}
}

// mergeCompiledTaskContractRequirements merges check/evidence items from the
// compiled task's verification contract into the profile-derived lists.
// Duplicates are suppressed. C-02: ensures the accepting flow honours
// task-level required_checks.
func mergeCompiledTaskContractRequirements(
	checks []string,
	evidence []string,
	contract *braincontracts.VerificationContract,
) ([]string, []string) {
	if contract == nil {
		return checks, evidence
	}

	existing := make(map[string]bool, len(checks))
	for _, c := range checks {
		existing[strings.ToLower(strings.TrimSpace(c))] = true
	}
	for _, item := range contract.RequiredChecks {
		key := strings.TrimSpace(item.CheckID)
		if key == "" {
			key = strings.TrimSpace(item.Kind)
		}
		if key != "" && !existing[strings.ToLower(key)] {
			checks = append(checks, key)
			existing[strings.ToLower(key)] = true
		}
	}

	existingEvidence := make(map[string]bool, len(evidence))
	for _, e := range evidence {
		existingEvidence[strings.ToLower(strings.TrimSpace(e))] = true
	}
	for _, item := range contract.RequiredEvidence {
		key := strings.TrimSpace(item.EvidenceID)
		if key == "" {
			key = strings.TrimSpace(item.Kind)
		}
		if key != "" && !existingEvidence[strings.ToLower(key)] {
			evidence = append(evidence, key)
			existingEvidence[strings.ToLower(key)] = true
		}
	}

	return checks, evidence
}

func profileVersionForVerification(
	acceptanceProfile *acceptanceProfileRecord,
	productionProfile *productionAcceptanceProfileRecord,
) string {
	if productionProfile != nil && strings.TrimSpace(productionProfile.ProfileVersion) != "" {
		return strings.TrimSpace(productionProfile.ProfileVersion)
	}
	if acceptanceProfile != nil {
		return strings.TrimSpace(acceptanceProfile.ProfileVersion)
	}
	return ""
}

func buildCompletionVerdictView(data *acceptanceAggregate) acceptancev1.CompletionVerdictView {
	if data != nil && data.PersistedVerdict != nil {
		return *data.PersistedVerdict
	}
	return deriveCompletionVerdictView(
		deriveVerificationResultForCompletion(context.Background(), data),
		buildRuntimeEscalationView(data),
		buildFaultSummaryView(data),
		acceptanceHasVerificationConflict(data),
		data.LatestAcceptanceRun != nil && data.LatestAcceptanceRun.ManualReleaseRequired == 1,
		projectHasBlockingManualReview(data.Issues),
		countBlockingIssues(data.Issues),
		firstNonEmpty(latestAcceptanceUpdatedAt(data), strings.TrimSpace(data.Project.UpdatedAt)),
		firstNonEmpty(strings.TrimSpace(acceptanceRunID(data)), strings.TrimSpace(data.Project.Id)),
	)
}

func projectHasBlockingManualReview(items []entity.AcceptanceIssues) bool {
	for _, issue := range items {
		if issue.Blocking == 1 {
			return true
		}
	}
	return false
}

func deriveVerificationResultForCompletion(ctx context.Context, data *acceptanceAggregate) acceptancev1.VerificationResultView {
	verification := buildVerificationResultView(ctx, data)
	return verification
}

func deriveCompletionVerdictView(
	verification acceptancev1.VerificationResultView,
	runtimeEscalation acceptancev1.RuntimeEscalationView,
	faultSummary acceptancev1.FaultSummaryView,
	verificationConflict bool,
	manualReleaseRequired bool,
	manualReviewRequired bool,
	blockerCount int,
	updatedAt string,
	sourceRunID string,
) acceptancev1.CompletionVerdictView {
	manualReviewRequired = manualReviewRequired || manualReleaseRequired
	decision := "blocked"
	finalStatus := ""
	reason := "Awaiting acceptance adjudication."
	completed := false
	manualReleaseCompleted := false
	nextAction := verification.Decision
	executorSucceeded := false
	deliveryVerified := false
	acceptancePassed := false

	switch verification.Status {
	case "failed":
		decision = "rework"
		finalStatus = "reworking"
		reason = verification.Summary
		nextAction = "prepare_rework"
		executorSucceeded = false
		deliveryVerified = false
		acceptancePassed = false
	case "incomplete":
		decision = "blocked"
		finalStatus = "accepting"
		reason = verification.Summary
		nextAction = "collect_evidence"
		executorSucceeded = true
		deliveryVerified = false
		acceptancePassed = false
	case "passed":
		acceptancePassed = true
		executorSucceeded = true
		deliveryVerified = true
		switch {
		case runtimeEscalation.Status != "none":
			decision = "manual_checkpoint"
			finalStatus = "accepting"
			switch runtimeEscalation.ReasonClass {
			case "channel_unavailable":
				reason = "Verification passed, but verification channel is unavailable. Switch to fallback channel."
				nextAction = "switch_verification_channel"
			case "environment_unavailable":
				reason = "Verification passed, but execution environment is unavailable. Restore environment before completion."
				nextAction = "restore_environment"
			default:
				reason = "Verification passed, but runtime escalation must be resolved before completion."
				nextAction = "resolve_runtime_escalation"
			}
		case faultSummary.FaultLoopDetected:
			decision = "manual_checkpoint"
			finalStatus = "reworking"
			reason = "Verification passed, but repeated acceptance faults require manual review before completion."
			nextAction = "review_fault_loop"
		case verificationConflict:
			decision = "manual_checkpoint"
			finalStatus = "accepting"
			reason = "Verification passed, but conflicting blocking signals require manual review before completion."
			nextAction = "resolve_verification_conflict"
		case manualReleaseRequired:
			decision = "manual_checkpoint"
			finalStatus = "accepting"
			reason = "Verification passed, but manual release confirmation is still required."
			nextAction = "apply_manual_release"
		case manualReviewRequired:
			decision = "manual_checkpoint"
			finalStatus = "accepting"
			reason = "Verification passed, but manual review is still required before completion."
			nextAction = "manual_checkpoint"
		default:
			decision = "complete"
			finalStatus = "completed"
			reason = "Verification passed and no manual release hold remains."
			completed = true
			manualReleaseCompleted = true
			nextAction = "complete_project"
		}
	}

	return acceptancev1.CompletionVerdictView{
		Decision:               decision,
		FinalStatus:            finalStatus,
		Reason:                 reason,
		ManualReviewRequired:   manualReviewRequired,
		ManualReleaseRequired:  manualReleaseRequired,
		ManualReleaseCompleted: manualReleaseCompleted,
		ReleaseReady:           completed,
		BlockerCount:           blockerCount,
		NextAction:             nextAction,
		SourceRunID:            sourceRunID,
		UpdatedAt:              updatedAt,
		Completed:              completed,
		Summary:                reason,
		ExecutorSucceeded:      executorSucceeded,
		DeliveryVerified:       deliveryVerified,
		AcceptancePassed:       acceptancePassed,
	}
}

func buildRuntimeEscalationView(data *acceptanceAggregate) acceptancev1.RuntimeEscalationView {
	// Check aggregate-level channel/environment unavailability before run bindings.
	if !data.ChannelAvailable {
		reason := data.ChannelUnavailableReason
		if reason == "" {
			reason = "Preferred verification channel (high_spec_remote) is unavailable."
		}
		return acceptancev1.RuntimeEscalationView{
			Status:      "escalated",
			ReasonClass: "channel_unavailable",
			Severity:    "blocking",
			Action:      "switch_verification_channel",
			Summary:     reason,
		}
	}
	if !data.EnvironmentAvailable {
		return acceptancev1.RuntimeEscalationView{
			Status:      "escalated",
			ReasonClass: "environment_unavailable",
			Severity:    "blocking",
			Action:      "restore_environment",
			Summary:     "Execution environment is unavailable. Cannot proceed with verification.",
		}
	}

	for _, binding := range data.RunBindings {
		status := normalizeBindingStatus(binding.RunStatus)
		if !bindingNeedsAttention(status) {
			continue
		}
		reasonClass := "runtime_attention"
		policyDenied := false
		switch status {
		case "run_denied":
			reasonClass = "policy_denied"
			policyDenied = true
		case "run_unsupported":
			reasonClass = "capability_gap"
		case "run_failed":
			reasonClass = "execution_failed"
		}
		return acceptancev1.RuntimeEscalationView{
			Status:       "escalated",
			ReasonClass:  reasonClass,
			SourceBrain:  strings.TrimSpace(binding.BrainKind),
			SourceTaskID: strings.TrimSpace(binding.TaskId),
			RunBindingID: strings.TrimSpace(binding.Id),
			RunStatus:    status,
			Severity:     deriveBindingSeverity(status),
			Action:       "open_task_review",
			TaskID:       strings.TrimSpace(binding.TaskId),
			RunID:        firstNonEmpty(strings.TrimSpace(binding.BrainRunId), strings.TrimSpace(binding.Id)),
			UpdatedAt:    firstNonEmpty(strings.TrimSpace(binding.LastSyncAt), strings.TrimSpace(binding.UpdatedAt)),
			Summary:      deriveBindingInboxTitle(binding, nil),
			PolicyDenied: policyDenied,
		}
	}
	return acceptancev1.RuntimeEscalationView{
		Status: "none",
	}
}

func buildFaultSummaryView(data *acceptanceAggregate) acceptancev1.FaultSummaryView {
	blocking := 0
	advisory := 0
	topIssue := ""
	for _, issue := range data.Issues {
		if strings.TrimSpace(topIssue) == "" {
			topIssue = firstNonEmpty(strings.TrimSpace(issue.Summary), strings.TrimSpace(issue.IssueKind), issue.Id)
		}
		if issue.Blocking == 1 {
			blocking++
		} else {
			advisory++
		}
	}

	faultLoopDetected := false
	failedRuns := 0
	for _, run := range data.AcceptanceRuns {
		if normalizeAcceptanceRunStatus(run.Status) == "failed" {
			failedRuns++
		}
	}
	if failedRuns >= 2 {
		faultLoopDetected = true
	}

	status := "healthy"
	if blocking > 0 {
		status = "blocking"
	} else if advisory > 0 || faultLoopDetected {
		status = "attention"
	}

	return acceptancev1.FaultSummaryView{
		Status:             status,
		BlockingIssueCount: blocking,
		AdvisoryIssueCount: advisory,
		TopIssue:           topIssue,
		FaultLoopDetected:  faultLoopDetected,
		FaultKind:          firstNonEmpty(topIssue, status),
		Severity:           deriveFaultSeverity(blocking, advisory),
		Summary:            deriveFaultSummaryText(status, topIssue, blocking, advisory, faultLoopDetected),
		FailedChecks:       summarizeIssueLabels(data.Issues),
		AffectedTasks:      summarizeEscalatedTasks(data.RunBindings),
		UpdatedAt:          firstNonEmpty(latestAcceptanceUpdatedAt(data), strings.TrimSpace(data.Project.UpdatedAt)),
	}
}

func buildRepairPlanDraftSummary(data *acceptanceAggregate) acceptancev1.RepairPlanDraftSummary {
	if data.RepairDraft == nil {
		return acceptancev1.RepairPlanDraftSummary{
			Status: "idle",
		}
	}
	return acceptancev1.RepairPlanDraftSummary{
		ID:                   strings.TrimSpace(data.RepairDraft.ID),
		Status:               normalizePlanState(data.RepairDraft.Status, "ready"),
		ReasonClass:          "acceptance_failure",
		RepairStrategy:       "repair_plan_draft",
		ReasoningSummary:     strings.TrimSpace(data.RepairDraft.RepairReasoningSummary),
		Summary:              strings.TrimSpace(data.RepairDraft.RepairReasoningSummary),
		UpdatedTasks:         summarizeEscalatedTasks(data.RunBindings),
		ManualReviewRequired: buildCompletionVerdictView(data).ManualReviewRequired,
		UpdatedAt:            strings.TrimSpace(data.RepairDraft.UpdatedAt),
	}
}

func acceptanceRunID(data *acceptanceAggregate) string {
	if data.LatestAcceptanceRun == nil {
		return ""
	}
	return data.LatestAcceptanceRun.Id
}

func acceptanceHasVerificationConflict(data *acceptanceAggregate) bool {
	if countBlockingIssues(data.Issues) == 0 || data.LatestAcceptanceRun == nil {
		return false
	}
	return isFunctionalPassed(data.LatestAcceptanceRun.FunctionalStatus) ||
		isProductionReady(data.LatestAcceptanceRun.ProductionStatus) ||
		data.LatestAcceptanceRun.ManualReleaseRequired == 1
}

// acceptanceHasFaultLoop detects if the project has entered a rework oscillation
// (failed acceptance runs >= 2). This maps Engineering Cybernetics ch.11
// "nonlinear system / limit cycle" into the EasyMVP fault loop concept.
func acceptanceHasFaultLoop(data *acceptanceAggregate) bool {
	if data == nil {
		return false
	}
	failedRuns := 0
	for _, run := range data.AcceptanceRuns {
		if normalizeAcceptanceRunStatus(run.Status) == "failed" {
			failedRuns++
		}
	}
	return failedRuns >= 2
}

func latestAcceptanceUpdatedAt(data *acceptanceAggregate) string {
	if data.LatestAcceptanceRun != nil {
		return firstNonEmpty(strings.TrimSpace(data.LatestAcceptanceRun.FinishedAt), strings.TrimSpace(data.LatestAcceptanceRun.CreatedAt))
	}
	if len(data.Judgements) > 0 {
		return strings.TrimSpace(data.Judgements[0].CreatedAt)
	}
	return ""
}

func deriveFaultSeverity(blocking, advisory int) string {
	switch {
	case blocking > 0:
		return "error"
	case advisory > 0:
		return "warning"
	default:
		return "info"
	}
}

func deriveFaultSummaryText(status, topIssue string, blocking, advisory int, faultLoopDetected bool) string {
	switch {
	case faultLoopDetected && topIssue != "":
		return "Repeated acceptance faults detected: " + topIssue
	case faultLoopDetected:
		return "Repeated acceptance faults detected."
	case status == "blocking" && topIssue != "":
		return topIssue
	case status == "attention" && advisory > 0:
		return fmt.Sprintf("%d advisory acceptance issue(s) remain open", advisory)
	default:
		return "No blocking acceptance fault is currently aggregated."
	}
}

func summarizeIssueLabels(items []entity.AcceptanceIssues) []string {
	result := make([]string, 0, len(items))
	for _, item := range items {
		label := firstNonEmpty(strings.TrimSpace(item.Summary), strings.TrimSpace(item.IssueKind), item.Id)
		if label == "" {
			continue
		}
		result = append(result, label)
		if len(result) >= 5 {
			break
		}
	}
	return result
}

func summarizeEscalatedTasks(bindings []entity.BrainRunBindings) []string {
	result := make([]string, 0, len(bindings))
	for _, item := range bindings {
		if !bindingNeedsAttention(item.RunStatus) {
			continue
		}
		label := firstNonEmpty(strings.TrimSpace(item.TaskId), strings.TrimSpace(item.Id))
		if label == "" {
			continue
		}
		result = append(result, label)
		if len(result) >= 5 {
			break
		}
	}
	return result
}

func buildAcceptanceCoverageMatrix(data *acceptanceAggregate) []acceptancev1.CoverageItem {
	requiredSurfaces := requiredAcceptanceSurfaces(data.Project.ProjectCategory)
	requiredJourneys := requiredAcceptanceJourneys(data.Project.ProjectCategory)
	if data.ProductionProfile != nil {
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredSurfacesJSON); len(items) > 0 {
			requiredSurfaces = items
		}
		if items := parseStringArrayJSON(data.ProductionProfile.RequiredJourneysJSON); len(items) > 0 {
			requiredJourneys = items
		}
	}
	evidenceBySurface, evidenceByJourney := groupEvidenceCounts(data.EvidenceItems)

	items := make([]acceptancev1.CoverageItem, 0, len(requiredSurfaces)+len(requiredJourneys))
	surfaceMap := make(map[string]entity.AcceptanceSurfaceCoverage, len(data.SurfaceCoverage))
	for _, item := range data.SurfaceCoverage {
		surfaceMap[strings.ToLower(strings.TrimSpace(item.Surface))] = item
	}
	journeyMap := make(map[string]entity.AcceptanceJourneyCoverage, len(data.JourneyCoverage))
	for _, item := range data.JourneyCoverage {
		journeyMap[strings.ToLower(strings.TrimSpace(item.Journey))] = item
	}

	for _, key := range requiredSurfaces {
		lowerKey := strings.ToLower(key)
		if item, ok := surfaceMap[lowerKey]; ok {
			items = append(items, acceptancev1.CoverageItem{
				Key:            key,
				Kind:           "surface",
				Name:           humanizeAcceptanceKey(key),
				CoverageStatus: normalizeCoverageStatus(item.CoverageStatus),
				EvidenceCount:  maxInt(item.EvidenceCount, evidenceBySurface[lowerKey]),
			})
			continue
		}
		items = append(items, acceptancev1.CoverageItem{
			Key:            key,
			Kind:           "surface",
			Name:           humanizeAcceptanceKey(key),
			CoverageStatus: deriveCoverageStatusFromEvidence(evidenceBySurface[lowerKey], data),
			EvidenceCount:  evidenceBySurface[lowerKey],
		})
	}

	for _, key := range requiredJourneys {
		lowerKey := strings.ToLower(key)
		if item, ok := journeyMap[lowerKey]; ok {
			items = append(items, acceptancev1.CoverageItem{
				Key:            key,
				Kind:           "journey",
				Name:           humanizeAcceptanceKey(key),
				CoverageStatus: normalizeCoverageStatus(item.CoverageStatus),
				EvidenceCount:  maxInt(item.EvidenceCount, evidenceByJourney[lowerKey]),
			})
			continue
		}
		items = append(items, acceptancev1.CoverageItem{
			Key:            key,
			Kind:           "journey",
			Name:           humanizeAcceptanceKey(key),
			CoverageStatus: deriveCoverageStatusFromEvidence(evidenceByJourney[lowerKey], data),
			EvidenceCount:  evidenceByJourney[lowerKey],
		})
	}

	return items
}

func buildAcceptanceIssues(data *acceptanceAggregate) []acceptancev1.AcceptanceIssue {
	items := make([]acceptancev1.AcceptanceIssue, 0, len(data.Issues))
	for _, issue := range data.Issues {
		items = append(items, acceptancev1.AcceptanceIssue{
			ID:              issue.Id,
			AcceptanceRunID: strings.TrimSpace(issue.AcceptanceRunId),
			Severity:        normalizeSeverity(issue.Severity),
			IssueKind:       strings.TrimSpace(issue.IssueKind),
			Blocking:        issue.Blocking == 1,
			Summary:         firstNonEmpty(issue.Summary, fmt.Sprintf("Acceptance issue: %s", issue.IssueKind)),
			DetailJSON:      strings.TrimSpace(issue.DetailJson),
			CreatedAt:       strings.TrimSpace(issue.CreatedAt),
		})
	}
	return items
}

func buildAcceptanceEvidenceCards(data *acceptanceAggregate) []acceptancev1.EvidenceCard {
	items := make([]acceptancev1.EvidenceCard, 0, len(data.EvidenceItems))
	for _, item := range data.EvidenceItems {
		items = append(items, acceptancev1.EvidenceCard{
			ID:           item.Id,
			Surface:      item.Surface,
			Journey:      item.Journey,
			EvidenceType: item.EvidenceType,
			FilePath:     item.FilePath,
			CapturedAt:   firstNonEmpty(item.CapturedAt, item.CreatedAt),
		})
	}
	return items
}

func buildAcceptanceReleaseGate(data *acceptanceAggregate) acceptancev1.ReleaseGateView {
	blockingCount := countBlockingIssues(data.Issues)
	releasedByHuman := hasHumanReleaseApproval(data)

	switch {
	case data.LatestAcceptanceRun == nil:
		return acceptancev1.ReleaseGateView{
			Status:     "pending_acceptance",
			NextAction: "start_acceptance_run",
			Summary:    "Acceptance has not started yet.",
		}
	case blockingCount > 0:
		return acceptancev1.ReleaseGateView{
			Status:     "blocked",
			NextAction: "resolve_blocking_issues",
			Summary:    fmt.Sprintf("%d blocking acceptance issue(s) must be resolved before release.", blockingCount),
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus) && data.LatestAcceptanceRun.ManualReleaseRequired == 1 && releasedByHuman:
		return acceptancev1.ReleaseGateView{
			Status:     "released",
			NextAction: "open_release_record",
			Summary:    "Production acceptance passed and manual release has been approved.",
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus) && data.LatestAcceptanceRun.ManualReleaseRequired == 1:
		return acceptancev1.ReleaseGateView{
			Status:     "awaiting_manual_release",
			NextAction: "apply_manual_release",
			Summary:    "Production acceptance passed, waiting for manual release approval.",
		}
	case isProductionReady(data.LatestAcceptanceRun.ProductionStatus):
		return acceptancev1.ReleaseGateView{
			Status:     "ready",
			NextAction: "complete_project",
			Summary:    "Production acceptance passed and the project can be delivered.",
		}
	case isFunctionalPassed(data.LatestAcceptanceRun.FunctionalStatus):
		return acceptancev1.ReleaseGateView{
			Status:     "functional_passed",
			NextAction: "continue_production_acceptance",
			Summary:    "Functional acceptance passed, but production readiness is still incomplete.",
		}
	default:
		return acceptancev1.ReleaseGateView{
			Status:     normalizeAcceptanceRunStatus(data.LatestAcceptanceRun.Status),
			NextAction: "watch_acceptance_progress",
			Summary:    "Acceptance is still gathering coverage and evidence.",
		}
	}
}

func requiredAcceptanceSurfaces(category string) []string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "game":
		return []string{"game_runtime", "admin_backend", "api_backend", "build_output", "ops_console"}
	case "video_editing":
		return []string{"editor_runtime", "admin_backend", "api_backend", "export_output"}
	case "web_app", "web":
		return []string{"user_frontend", "admin_backend", "api_backend", "build_output"}
	default:
		return []string{"user_frontend", "admin_backend", "api_backend", "build_output"}
	}
}

func requiredAcceptanceJourneys(category string) []string {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case "game":
		return []string{"launch_game", "core_game_loop", "pause_resume", "finish_settlement", "restart_game", "admin_sync", "build_verification"}
	case "video_editing":
		return []string{"import_media", "timeline_edit", "preview_playback", "export_video", "admin_sync", "build_verification"}
	case "web_app", "web":
		return []string{"user_sign_in", "primary_user_flow", "admin_operate", "state_write_read", "api_verification", "build_verification"}
	default:
		return []string{"user_sign_in", "primary_user_flow", "admin_operate", "state_write_read", "api_verification", "build_verification"}
	}
}

func groupEvidenceCounts(items []entity.EvidenceItems) (map[string]int, map[string]int) {
	bySurface := make(map[string]int)
	byJourney := make(map[string]int)
	for _, item := range items {
		if key := strings.ToLower(strings.TrimSpace(item.Surface)); key != "" {
			bySurface[key]++
		}
		if key := strings.ToLower(strings.TrimSpace(item.Journey)); key != "" {
			byJourney[key]++
		}
	}
	return bySurface, byJourney
}

func deriveCoverageStatusFromEvidence(evidenceCount int, data *acceptanceAggregate) string {
	switch {
	case evidenceCount <= 0:
		return "missing"
	case data.LatestAcceptanceRun != nil && isProductionReady(data.LatestAcceptanceRun.ProductionStatus):
		return "pass"
	default:
		return "partial"
	}
}

func normalizeCoverageStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "pass", "passed", "ready", "covered":
		return "pass"
	case "partial", "in_progress":
		return "partial"
	default:
		return "missing"
	}
}

func normalizeAcceptanceRunStatus(status string) string {
	status = strings.ToLower(strings.TrimSpace(status))
	if status == "" {
		return "pending"
	}
	return strings.ReplaceAll(status, " ", "_")
}

func normalizeFunctionalStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "passed", "pass", "functional_passed":
		return "passed"
	case "failed", "error":
		return "failed"
	case "partial":
		return "partial"
	default:
		return "pending"
	}
}

func deriveAcceptanceRunStatusFromProject(projectStatus string) string {
	switch strings.ToLower(strings.TrimSpace(projectStatus)) {
	case "acceptance":
		return "running"
	case "completed":
		return "completed"
	default:
		return "pending"
	}
}

func humanizeAcceptanceKey(key string) string {
	parts := strings.Split(strings.TrimSpace(key), "_")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func hasHumanReleaseApproval(data *acceptanceAggregate) bool {
	for _, item := range data.Judgements {
		kind := strings.ToLower(strings.TrimSpace(item.JudgementKind))
		result := strings.ToLower(strings.TrimSpace(item.JudgementResult))
		if strings.Contains(kind, "release") && (result == "approved" || result == "released" || result == "passed") {
			return true
		}
	}
	for _, item := range data.AuditLogs {
		eventType := strings.ToLower(strings.TrimSpace(item.EventType))
		summary := strings.ToLower(strings.TrimSpace(item.Summary))
		if strings.Contains(eventType, "release") && (strings.Contains(summary, "approve") || strings.Contains(summary, "released")) {
			return true
		}
	}
	return false
}
