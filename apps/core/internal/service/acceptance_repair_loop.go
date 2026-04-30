package service

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// defaultMaxRepairRounds is the maximum number of repair+re-acceptance loops
// allowed before the loop gives up and marks the project as needing redesign.
const defaultMaxRepairRounds = 3

// AcceptanceRepairResult is the final outcome of the repair loop.
type AcceptanceRepairResult struct {
	// Passed indicates whether the project eventually cleared all acceptance criteria.
	Passed bool

	// Rounds is the total number of repair+re-acceptance iterations performed.
	Rounds int

	// FinalResult is the MultiLayerAcceptanceResult from the last acceptance run.
	FinalResult *MultiLayerAcceptanceResult

	// RepairHistory records every repair round that was executed.
	RepairHistory []RepairRound

	// NeedsRedesign is true when the remaining failures involve architectural
	// problems that cannot be resolved through targeted repair tasks and the
	// project must return to Phase 2 (design).
	NeedsRedesign bool
}

// RepairRound captures what was attempted and what happened in a single
// repair iteration.
type RepairRound struct {
	// Round is a 1-based counter.
	Round int

	// FailedItems are the FailedCriterion entries that triggered this round.
	FailedItems []FailedCriterion

	// RepairTasks are the tasks that were generated to address the failures.
	RepairTasks []RepairTask

	// RepairResult is a short status string: "success", "partial", or "failed".
	RepairResult string
}

// RepairTask describes a single unit of work required to fix a failed criterion.
type RepairTask struct {
	// TaskName is a short human-readable identifier for the task.
	TaskName string

	// Description explains what needs to be fixed.
	Description string

	// Severity classifies how serious the failure is.
	// Allowed values: "minor", "major", "critical".
	Severity string

	// BrainKind names the specialist brain best suited to execute this task.
	// Allowed values: "code", "verifier", "browser".
	BrainKind string

	// Instruction is the full prompt that will be sent to the Brain.
	Instruction string
}

// RunAcceptanceRepairLoop is the Phase 5 closed-loop repair orchestrator.
//
// Algorithm:
//  1. Run multi-layer acceptance.
//  2. If all criteria pass → return success immediately.
//  3. Analyse failing criteria and classify each into a RepairTask.
//  4. Execute repair tasks via the appropriate Brain.
//  5. Re-run acceptance.
//  6. Repeat at most maxRepairRounds times (default 3 when ≤ 0 is supplied).
//  7. If architectural failures persist after all rounds, set NeedsRedesign=true.
func RunAcceptanceRepairLoop(
	ctx context.Context,
	projectID string,
	criteria AcceptanceCriteria,
	maxRepairRounds int,
) (*AcceptanceRepairResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if len(criteria.Layers) == 0 {
		return nil, gerror.New("acceptance criteria must contain at least one layer")
	}
	if maxRepairRounds <= 0 {
		maxRepairRounds = defaultMaxRepairRounds
	}

	out := &AcceptanceRepairResult{
		RepairHistory: make([]RepairRound, 0, maxRepairRounds),
	}

	for round := 0; round <= maxRepairRounds; round++ {
		// ── Step 1: run acceptance ─────────────────────────────────────────
		acceptResult, err := RunMultiLayerAcceptance(ctx, projectID, criteria)
		if err != nil {
			return nil, gerror.Wrapf(err, "acceptance run failed on round %d", round)
		}
		out.FinalResult = acceptResult

		// ── Step 2: all passed → done ──────────────────────────────────────
		if acceptResult.Passed {
			out.Passed = true
			out.Rounds = round
			g.Log().Infof(ctx, "[RepairLoop] project=%s passed acceptance after %d repair round(s)", projectID, round)
			return out, nil
		}

		// ── Step 3: max rounds exhausted ──────────────────────────────────
		if round == maxRepairRounds {
			out.Rounds = round
			out.NeedsRedesign = hasArchitecturalFailures(acceptResult.FailedItems)
			g.Log().Warningf(ctx, "[RepairLoop] project=%s exhausted %d repair round(s); needs_redesign=%v", projectID, round, out.NeedsRedesign)
			return out, nil
		}

		// ── Step 4: analyse failures and generate repair tasks ─────────────
		repairTasks := AnalyzeFailures(acceptResult.FailedItems)
		if len(repairTasks) == 0 {
			// Nothing actionable — bail out.
			out.Rounds = round
			out.NeedsRedesign = true
			g.Log().Warningf(ctx, "[RepairLoop] project=%s no repair tasks generated for %d failed items", projectID, len(acceptResult.FailedItems))
			return out, nil
		}

		// ── Step 5: execute repair tasks ──────────────────────────────────
		repairErr := ExecuteRepairTasks(ctx, projectID, repairTasks)
		repairStatus := "success"
		if repairErr != nil {
			g.Log().Warningf(ctx, "[RepairLoop] project=%s round=%d repair execution partial error: %v", projectID, round+1, repairErr)
			repairStatus = "partial"
		}

		out.RepairHistory = append(out.RepairHistory, RepairRound{
			Round:        round + 1,
			FailedItems:  acceptResult.FailedItems,
			RepairTasks:  repairTasks,
			RepairResult: repairStatus,
		})
	}

	// Should be unreachable, but be defensive.
	out.Rounds = maxRepairRounds
	return out, nil
}

// AnalyzeFailures maps a list of FailedCriterion entries to concrete RepairTask
// instructions. Each failure is classified by severity and routed to the most
// appropriate Brain.
//
// Classification rules:
//   - style/ui/layout failures → minor, code brain
//   - functional/logic failures → major, code brain + verifier confirmation
//   - security/architecture failures → critical, needs redesign flag
//   - performance failures → major, verifier brain
//   - e2e/browser failures → major, browser brain
func AnalyzeFailures(failures []FailedCriterion) []RepairTask {
	if len(failures) == 0 {
		return nil
	}

	tasks := make([]RepairTask, 0, len(failures))
	for _, f := range failures {
		task := classifyFailure(f)
		tasks = append(tasks, task)
	}
	return tasks
}

// classifyFailure maps a single FailedCriterion to a RepairTask.
func classifyFailure(f FailedCriterion) RepairTask {
	layer := strings.ToLower(strings.TrimSpace(f.Layer))
	criterion := strings.ToLower(strings.TrimSpace(f.Criterion))

	switch {
	case layer == "e2e_test":
		return RepairTask{
			TaskName:    fmt.Sprintf("repair_e2e_%s", sanitizeTaskName(f.Criterion)),
			Description: fmt.Sprintf("Fix E2E test failure in criterion '%s': %s", f.Criterion, f.Reason),
			Severity:    "major",
			BrainKind:   "browser",
			Instruction: buildBrowserRepairInstruction(f),
		}

	case layer == "security_test":
		return RepairTask{
			TaskName:    fmt.Sprintf("repair_security_%s", sanitizeTaskName(f.Criterion)),
			Description: fmt.Sprintf("Fix security failure in criterion '%s': %s", f.Criterion, f.Reason),
			Severity:    "critical",
			BrainKind:   "code",
			Instruction: buildSecurityRepairInstruction(f),
		}

	case layer == "performance_test":
		return RepairTask{
			TaskName:    fmt.Sprintf("repair_perf_%s", sanitizeTaskName(f.Criterion)),
			Description: fmt.Sprintf("Fix performance failure in criterion '%s': %s", f.Criterion, f.Reason),
			Severity:    "major",
			BrainKind:   "verifier",
			Instruction: buildVerifierRepairInstruction(f),
		}

	case isStyleCriterion(criterion):
		return RepairTask{
			TaskName:    fmt.Sprintf("repair_style_%s", sanitizeTaskName(f.Criterion)),
			Description: fmt.Sprintf("Fix style/UI issue in criterion '%s': %s", f.Criterion, f.Reason),
			Severity:    "minor",
			BrainKind:   "code",
			Instruction: buildCodeRepairInstruction(f),
		}

	default:
		// functional / integration / unit failures
		return RepairTask{
			TaskName:    fmt.Sprintf("repair_func_%s", sanitizeTaskName(f.Criterion)),
			Description: fmt.Sprintf("Fix functional failure in criterion '%s' (layer: %s): %s", f.Criterion, f.Layer, f.Reason),
			Severity:    "major",
			BrainKind:   "code",
			Instruction: buildCodeRepairInstruction(f),
		}
	}
}

// ExecuteRepairTasks sends each repair task to its designated Brain and collects
// errors. A non-nil return value means at least one task failed, but execution
// continues for all remaining tasks (partial repair is better than no repair).
func ExecuteRepairTasks(ctx context.Context, projectID string, tasks []RepairTask) error {
	if len(tasks) == 0 {
		return nil
	}

	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return gerror.Wrap(err, "resolve brain serve base url for repair tasks failed")
	}
	client := &http.Client{Timeout: 60 * time.Second}

	var firstErr error
	for _, task := range tasks {
		if err := executeSingleRepairTask(ctx, client, baseURL, projectID, task); err != nil {
			g.Log().Warningf(ctx, "[ExecuteRepairTasks] project=%s task=%s brain=%s error: %v",
				projectID, task.TaskName, task.BrainKind, err)
			if firstErr == nil {
				firstErr = err
			}
		} else {
			g.Log().Infof(ctx, "[ExecuteRepairTasks] project=%s task=%s brain=%s dispatched",
				projectID, task.TaskName, task.BrainKind)
		}
	}
	return firstErr
}

// executeSingleRepairTask dispatches a RepairTask to its Brain via the existing
// multi-brain infrastructure.
func executeSingleRepairTask(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	projectID string,
	task RepairTask,
) error {
	multiReq := &MultiBrainRunRequest{
		ProjectID: projectID,
		Prompt:    task.Instruction,
		MaxTurns:  8,
		Brains:    []string{task.BrainKind},
	}
	multiRes := ExecuteMultiBrain(ctx, client, baseURL, multiReq)

	if brainErr, ok := multiRes.Errors[task.BrainKind]; ok && brainErr != nil {
		return gerror.Wrapf(brainErr, "brain %s failed for repair task %s", task.BrainKind, task.TaskName)
	}
	brainResult, ok := multiRes.Results[task.BrainKind]
	if !ok || brainResult == nil {
		return gerror.Newf("no result from %s brain for repair task %s", task.BrainKind, task.TaskName)
	}
	if brainResult.Status != "completed" && brainResult.Status != "success" {
		return gerror.Newf("brain %s returned status %q for repair task %s", task.BrainKind, brainResult.Status, task.TaskName)
	}
	return nil
}

// hasArchitecturalFailures returns true when any of the failed criteria indicate
// a structural problem (security or integration layer) that repair tasks alone
// cannot resolve and requires returning to Phase 2 (design).
func hasArchitecturalFailures(failures []FailedCriterion) bool {
	for _, f := range failures {
		layer := strings.ToLower(strings.TrimSpace(f.Layer))
		if layer == "security_test" || layer == "integration_test" {
			return true
		}
	}
	return false
}

// isStyleCriterion returns true when the criterion name hints at a
// presentation-layer (CSS/UI/layout) concern.
func isStyleCriterion(criterion string) bool {
	styleKeywords := []string{"style", "css", "layout", "ui", "visual", "theme", "color", "font", "spacing", "responsive"}
	for _, kw := range styleKeywords {
		if strings.Contains(criterion, kw) {
			return true
		}
	}
	return false
}

// sanitizeTaskName converts a criterion name to a safe snake_case identifier
// suitable for use in task names.
func sanitizeTaskName(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	replacer := strings.NewReplacer(" ", "_", "-", "_", "/", "_", ".", "_")
	return replacer.Replace(name)
}

// ── Instruction builders ────────────────────────────────────────────────────

func buildCodeRepairInstruction(f FailedCriterion) string {
	return fmt.Sprintf(
		"Acceptance layer '%s' failed on criterion '%s'. Failure reason: %s\n"+
			"Please analyse the codebase, identify the root cause, and apply a targeted fix. "+
			"Verify the fix by running the relevant tests before finishing.",
		f.Layer, f.Criterion, f.Reason,
	)
}

func buildBrowserRepairInstruction(f FailedCriterion) string {
	return fmt.Sprintf(
		"E2E acceptance criterion '%s' failed. Failure reason: %s\n"+
			"Please open the application in the browser, reproduce the failure, and fix the "+
			"underlying code or UI interaction causing the test to fail.",
		f.Criterion, f.Reason,
	)
}

func buildVerifierRepairInstruction(f FailedCriterion) string {
	return fmt.Sprintf(
		"Performance acceptance criterion '%s' (layer: %s) failed. Failure reason: %s\n"+
			"Please profile the relevant code path, identify the bottleneck, and apply an "+
			"optimisation. Re-run the performance check to confirm the criterion now passes.",
		f.Criterion, f.Layer, f.Reason,
	)
}

func buildSecurityRepairInstruction(f FailedCriterion) string {
	return fmt.Sprintf(
		"Security acceptance criterion '%s' failed. Failure reason: %s\n"+
			"This is a critical security issue. Please identify the vulnerability, apply a "+
			"hardening fix, and verify that the security check now passes. Do NOT introduce "+
			"any temporary workarounds that simply suppress the test output.",
		f.Criterion, f.Reason,
	)
}
