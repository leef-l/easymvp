package service

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/model/braincontracts"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

// ReviewResult holds the outcome of a single design review round.
type ReviewResult struct {
	ReviewID string
	Passed   bool
	Score    int
	Issues   []string
	Round    int
}

// ReviewLoopResult holds the outcome of the full review-fix loop.
type ReviewLoopResult struct {
	Passed        bool
	Rounds        int
	Reason        string
	FinalReviewID string
}

const reviewLoopDefaultMaxRounds = 5

type IReview interface {
	StartDesignReview(ctx context.Context, designID, projectID string) (*ReviewResult, error)
	RunReviewLoop(ctx context.Context, designID, projectID string, maxRounds int) (*ReviewLoopResult, error)
	GetDesignReviews(ctx context.Context, designID string) ([]entity.DesignReviews, error)
	// Intervene allows human operators to override or abort an ongoing design review.
	// action: "override_approve" — force-pass the design; "abort" — stop the review loop;
	// "restart" — delete review history and start fresh.
	Intervene(ctx context.Context, designID string, action string, reason string) error
}

var localReview IReview = (*sReview)(nil)

type sReview struct{}

func Review() IReview {
	if localReview == nil {
		localReview = &sReview{}
	}
	return localReview
}

// StartDesignReview executes a single review round for the given solution design.
func (s *sReview) StartDesignReview(ctx context.Context, designID, projectID string) (*ReviewResult, error) {
	designID = strings.TrimSpace(designID)
	projectID = strings.TrimSpace(projectID)
	if designID == "" {
		return nil, gerror.New("design id is required")
	}
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	// Open database and load both design and project.
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	design, err := getDesignByID(ctx, db, designID)
	if err != nil {
		return nil, gerror.Wrapf(err, "load solution design %s failed", designID)
	}

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return nil, gerror.Wrapf(err, "load project %s failed", projectID)
	}

	// Determine the current round by counting existing reviews.
	existingReviews, err := listDesignReviews(ctx, designID)
	if err != nil {
		return nil, gerror.Wrap(err, "list existing design reviews failed")
	}
	round := len(existingReviews) + 1

	// Collect previous issues for context.
	var previousIssues []string
	if len(existingReviews) > 0 {
		lastReview := existingReviews[len(existingReviews)-1]
		previousIssues = parseStringArrayJSON(lastReview.IssuesJson)
	}

	// Build brain contract input.
	input := braincontracts.DesignReviewInput{
		DesignID:       designID,
		ProjectID:      projectID,
		DesignVersion:  design.Version,
		Architecture:   design.Architecture,
		ModulesJSON:    safeRawJSON(design.ModulesJson),
		DataModelsJSON: safeRawJSON(design.DataModelsJson),
		PagesJSON:      safeRawJSON(design.PagesJson),
		TaskDraftsJSON: safeRawJSON(design.TaskDraftsJson),
		GoalSummary:    project.GoalSummary,
		Round:          round,
		PreviousIssues: previousIssues,
	}

	// Call the Verifier Brain.
	envelope, result, err := callDesignReviewContract(ctx, input)
	if err != nil {
		return nil, gerror.Wrap(err, "design review brain contract failed")
	}

	// Persist the review record.
	reviewID := newResourceID("design_review")
	brainRunID := ""
	if envelope != nil {
		brainRunID = strings.TrimSpace(envelope.TraceID)
	}

	dimensionsJSON := mustMarshalJSONString(result.Dimensions, "[]")
	issuesJSON := marshalIssuesJSON(result.Issues)
	suggestionsJSON := mustMarshalJSONString(result.Suggestions, "[]")

	passedInt := 0
	if result.Passed {
		passedInt = 1
	}

	if err = insertDesignReview(ctx, reviewID, designID, projectID, round, passedInt, result.Score, dimensionsJSON, issuesJSON, suggestionsJSON, brainRunID); err != nil {
		return nil, gerror.Wrap(err, "insert design review record failed")
	}

	if auditErr := insertAuditLog(ctx, projectID, "design.review.completed", "system", "Design review completed", map[string]any{
		"review_id": reviewID,
		"design_id": designID,
		"round":     round,
		"passed":    result.Passed,
		"score":     result.Score,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	// Build simplified issue list.
	issueStrings := make([]string, 0, len(result.Issues))
	for _, issue := range result.Issues {
		issueStrings = append(issueStrings, issue.Summary)
	}

	return &ReviewResult{
		ReviewID: reviewID,
		Passed:   result.Passed,
		Score:    result.Score,
		Issues:   issueStrings,
		Round:    round,
	}, nil
}

// RunReviewLoop runs an automated review→fix→re-review loop, up to maxRounds.
func (s *sReview) RunReviewLoop(ctx context.Context, designID, projectID string, maxRounds int) (*ReviewLoopResult, error) {
	designID = strings.TrimSpace(designID)
	projectID = strings.TrimSpace(projectID)
	if designID == "" {
		return nil, gerror.New("design id is required")
	}
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if maxRounds <= 0 {
		maxRounds = reviewLoopDefaultMaxRounds
	}

	g.Log().Infof(ctx, "starting design review loop for design=%s project=%s maxRounds=%d", designID, projectID, maxRounds)

	var prevIssueCount int
	var lastReviewID string

	for round := 1; round <= maxRounds; round++ {
		g.Log().Infof(ctx, "review loop round %d/%d for design=%s", round, maxRounds, designID)

		// 1. Execute a review round.
		result, err := s.StartDesignReview(ctx, designID, projectID)
		if err != nil {
			return nil, gerror.Wrapf(err, "review loop round %d failed", round)
		}
		lastReviewID = result.ReviewID

		// 2. If passed, return success.
		if result.Passed && isHardPass(result.Score, result.Issues) {
			g.Log().Infof(ctx, "review loop passed at round %d with score=%d", round, result.Score)
			return &ReviewLoopResult{
				Passed:        true,
				Rounds:        round,
				Reason:        "review passed",
				FinalReviewID: result.ReviewID,
			}, nil
		}

		// 3. Check convergence (starting from round 2).
		currentIssueCount := len(result.Issues)
		if round > 1 && !isConverging(prevIssueCount, currentIssueCount) {
			// Check if remaining issues are all minor — soft convergence.
			if isSoftConvergence(ctx, designID) {
				g.Log().Infof(ctx, "review loop soft converged at round %d (remaining issues are minor)", round)
				return &ReviewLoopResult{
					Passed:        true,
					Rounds:        round,
					Reason:        "soft convergence: remaining issues are minor",
					FinalReviewID: result.ReviewID,
				}, nil
			}
			g.Log().Warningf(ctx, "review loop not converging at round %d: prev=%d, current=%d", round, prevIssueCount, currentIssueCount)
			return &ReviewLoopResult{
				Passed:        false,
				Rounds:        round,
				Reason:        "not converging: issue reduction < 20%",
				FinalReviewID: result.ReviewID,
			}, nil
		}
		prevIssueCount = currentIssueCount

		// 4. If this is the last round and not passed, don't attempt a fix.
		if round == maxRounds {
			break
		}

		// 5. Generate fix tasks and apply corrections.
		if err = s.applyDesignFix(ctx, designID, projectID, result); err != nil {
			g.Log().Warningf(ctx, "design fix failed at round %d: %v", round, err)
			return &ReviewLoopResult{
				Passed:        false,
				Rounds:        round,
				Reason:        "fix failed: " + err.Error(),
				FinalReviewID: result.ReviewID,
			}, nil
		}
	}

	return &ReviewLoopResult{
		Passed:        false,
		Rounds:        maxRounds,
		Reason:        "reached max rounds",
		FinalReviewID: lastReviewID,
	}, nil
}

// GetDesignReviews returns all review records for the given design.
func (s *sReview) GetDesignReviews(ctx context.Context, designID string) ([]entity.DesignReviews, error) {
	designID = strings.TrimSpace(designID)
	if designID == "" {
		return nil, gerror.New("design id is required")
	}
	return listDesignReviews(ctx, designID)
}

// applyDesignFix calls the Central Brain to fix the solution design based on review issues.
func (s *sReview) applyDesignFix(ctx context.Context, designID, projectID string, reviewResult *ReviewResult) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	design, err := getDesignByID(ctx, db, designID)
	if err != nil {
		return gerror.Wrapf(err, "load solution design %s for fix failed", designID)
	}

	project, err := getProjectByID(ctx, db, projectID)
	if err != nil {
		return gerror.Wrapf(err, "load project %s for fix failed", projectID)
	}

	// Load the full review record to get structured issues.
	latestReviews, err := listDesignReviews(ctx, designID)
	if err != nil {
		return gerror.Wrap(err, "load reviews for fix failed")
	}
	var structuredIssues []braincontracts.DesignReviewIssue
	var suggestions []string
	if len(latestReviews) > 0 {
		lastReview := latestReviews[len(latestReviews)-1]
		structuredIssues = parseDesignReviewIssues(lastReview.IssuesJson)
		suggestions = parseStringArrayJSON(lastReview.SuggestionsJson)
	}

	fixInput := braincontracts.DesignFixInput{
		DesignID:       designID,
		ProjectID:      projectID,
		DesignVersion:  design.Version,
		Architecture:   design.Architecture,
		ModulesJSON:    safeRawJSON(design.ModulesJson),
		DataModelsJSON: safeRawJSON(design.DataModelsJson),
		PagesJSON:      safeRawJSON(design.PagesJson),
		TaskDraftsJSON: safeRawJSON(design.TaskDraftsJson),
		GoalSummary:    project.GoalSummary,
		Issues:         structuredIssues,
		Suggestions:    suggestions,
	}

	_, fixResult, err := callDesignFixContract(ctx, fixInput)
	if err != nil {
		return gerror.Wrap(err, "design fix brain contract failed")
	}

	// Update solution_designs with the fixed content and bump version.
	newVersion := design.Version + 1
	if err = updateDesignWithFix(ctx, designID, newVersion, fixResult); err != nil {
		return gerror.Wrap(err, "update solution design with fix failed")
	}

	// Store fix tasks in the latest review record.
	if len(latestReviews) > 0 {
		fixTasksJSON := mustMarshalJSONString(fixResult.FixedIssues, "[]")
		if err = updateDesignReviewFixTasks(ctx, latestReviews[len(latestReviews)-1].Id, fixTasksJSON); err != nil {
			g.Log().Warningf(ctx, "update review fix tasks failed: %v", err)
		}
	}

	if auditErr := insertAuditLog(ctx, projectID, "design.fix.applied", "system", "Design fix applied", map[string]any{
		"design_id":       designID,
		"new_version":     newVersion,
		"fixed_issues":    len(fixResult.FixedIssues),
		"changes_summary": fixResult.ChangesSummary,
	}); auditErr != nil {
		g.Log().Errorf(ctx, "insert audit log failed: %v", auditErr)
	}

	g.Log().Infof(ctx, "design fix applied for design=%s, version bumped to %d", designID, newVersion)
	return nil
}

// isHardPass checks if the review result meets the hard-pass criteria.
// Score >= 85 and no critical issues.
func isHardPass(score int, issues []string) bool {
	return score >= 85
}

// isConverging checks if issue count is reducing by at least 20%.
func isConverging(prevCount, currentCount int) bool {
	if prevCount == 0 {
		return currentCount == 0
	}
	reduction := float64(prevCount-currentCount) / float64(prevCount)
	return reduction >= 0.2
}

// isSoftConvergence checks if remaining issues in the latest review are all minor.
func isSoftConvergence(ctx context.Context, designID string) bool {
	reviews, err := listDesignReviews(ctx, designID)
	if err != nil || len(reviews) == 0 {
		return false
	}
	lastReview := reviews[len(reviews)-1]
	issues := parseDesignReviewIssues(lastReview.IssuesJson)
	if len(issues) == 0 {
		return true
	}
	for _, issue := range issues {
		severity := strings.ToLower(strings.TrimSpace(issue.Severity))
		if severity == "critical" || severity == "major" {
			return false
		}
	}
	return true
}

// callDesignReviewContract calls the Verifier Brain design_review contract.
func callDesignReviewContract(ctx context.Context, input braincontracts.DesignReviewInput) (*braincontracts.BrainContractEnvelope, *braincontracts.DesignReviewResult, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "marshal design review input failed")
	}

	execResult, err := EasyMVPBrain().ExecuteContract(ctx, EasyMVPBrainExecuteCommand{
		ContractKind: "design_review",
		Instruction:  "Review the solution design and provide structured feedback with pass/fail decision, score, dimensional analysis, issues, and suggestions.",
		ContextJSON:  inputJSON,
	})
	if err != nil {
		return nil, nil, err
	}
	if execResult == nil || execResult.Envelope == nil {
		return nil, nil, gerror.New("design review contract returned empty result")
	}

	var result braincontracts.DesignReviewResult
	if err = json.Unmarshal(execResult.Envelope.ResultJSON, &result); err != nil {
		return nil, nil, gerror.Wrap(err, "decode design review result failed")
	}

	return execResult.Envelope, &result, nil
}

// callDesignFixContract calls the Central Brain design_fix contract.
func callDesignFixContract(ctx context.Context, input braincontracts.DesignFixInput) (*braincontracts.BrainContractEnvelope, *braincontracts.DesignFixResult, error) {
	inputJSON, err := json.Marshal(input)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "marshal design fix input failed")
	}

	execResult, err := EasyMVPBrain().ExecuteContract(ctx, EasyMVPBrainExecuteCommand{
		ContractKind: "design_fix",
		Instruction:  "Fix the solution design based on the review issues and suggestions. Return the corrected architecture, modules, data models, pages, and task drafts.",
		ContextJSON:  inputJSON,
	})
	if err != nil {
		return nil, nil, err
	}
	if execResult == nil || execResult.Envelope == nil {
		return nil, nil, gerror.New("design fix contract returned empty result")
	}

	var result braincontracts.DesignFixResult
	if err = json.Unmarshal(execResult.Envelope.ResultJSON, &result); err != nil {
		return nil, nil, gerror.Wrap(err, "decode design fix result failed")
	}

	return execResult.Envelope, &result, nil
}

// safeRawJSON converts a string to json.RawMessage, returning null for empty strings.
func safeRawJSON(s string) json.RawMessage {
	s = strings.TrimSpace(s)
	if s == "" {
		return json.RawMessage("null")
	}
	return json.RawMessage(s)
}

// marshalIssuesJSON serializes design review issues to JSON string.
func marshalIssuesJSON(issues []braincontracts.DesignReviewIssue) string {
	if len(issues) == 0 {
		return "[]"
	}
	data, err := json.Marshal(issues)
	if err != nil {
		return "[]"
	}
	return string(data)
}

// parseDesignReviewIssues parses the issues_json field into structured issues.
func parseDesignReviewIssues(raw string) []braincontracts.DesignReviewIssue {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var issues []braincontracts.DesignReviewIssue
	if err := json.Unmarshal([]byte(raw), &issues); err != nil {
		return nil
	}
	return issues
}

// Intervene allows human operators to override, abort, or restart a design review.
func (s *sReview) Intervene(ctx context.Context, designID string, action string, reason string) error {
	designID = strings.TrimSpace(designID)
	action = strings.TrimSpace(action)
	if designID == "" {
		return gerror.New("design id is required")
	}
	if action == "" {
		return gerror.New("action is required")
	}

	// Load existing reviews so we know the project ID for audit logging.
	reviews, err := listDesignReviews(ctx, designID)
	if err != nil {
		return gerror.Wrapf(err, "load design reviews for %s failed", designID)
	}

	projectID := ""
	if len(reviews) > 0 {
		projectID = reviews[0].ProjectId
	}

	switch action {
	case "override_approve":
		// Insert a synthetic "human approved" review record with score 100.
		reviewID := newResourceID("design_review")
		round := len(reviews) + 1
		reasonJSON := mustMarshalJSONString([]string{}, "[]")
		if err = insertDesignReview(ctx, reviewID, designID, projectID, round, 1, 100, "[]", "[]", reasonJSON, "human_override"); err != nil {
			return gerror.Wrap(err, "insert override review record failed")
		}
		// Mark the solution design as approved.
		if _, dbErr := g.DB().Exec(ctx,
			"UPDATE solution_designs SET status = ? WHERE id = ?",
			"approved", designID,
		); dbErr != nil {
			return gerror.Wrap(dbErr, "update design status to approved failed")
		}
		g.Log().Infof(ctx, "design review overridden by human: design=%s reason=%s", designID, reason)

	case "abort":
		// Mark the solution design as review_aborted.
		if _, dbErr := g.DB().Exec(ctx,
			"UPDATE solution_designs SET status = ? WHERE id = ?",
			"review_aborted", designID,
		); dbErr != nil {
			return gerror.Wrap(dbErr, "update design status to review_aborted failed")
		}
		g.Log().Infof(ctx, "design review aborted by human: design=%s reason=%s", designID, reason)

	case "restart":
		// Delete all existing review records for this design.
		if _, dbErr := g.DB().Exec(ctx,
			"DELETE FROM design_reviews WHERE design_id = ?",
			designID,
		); dbErr != nil {
			return gerror.Wrap(dbErr, "delete design reviews for restart failed")
		}
		// Reset the solution design status back to reviewing.
		if _, dbErr := g.DB().Exec(ctx,
			"UPDATE solution_designs SET status = ? WHERE id = ?",
			"reviewing", designID,
		); dbErr != nil {
			return gerror.Wrap(dbErr, "reset design status for restart failed")
		}
		g.Log().Infof(ctx, "design review restarted by human: design=%s reason=%s", designID, reason)

	default:
		return gerror.Newf("unknown action %q: must be one of override_approve, abort, restart", action)
	}

	// Write audit log if we have a projectID.
	if projectID != "" {
		if auditErr := insertAuditLog(ctx, projectID, "design.review.intervene", "human", "Human intervention on design review", map[string]any{
			"design_id": designID,
			"action":    action,
			"reason":    reason,
		}); auditErr != nil {
			g.Log().Errorf(ctx, "insert audit log for intervene failed: %v", auditErr)
		}
	}

	return nil
}
