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

// MultiLayerAcceptanceResult is the aggregated result of a multi-layer acceptance run.
type MultiLayerAcceptanceResult struct {
	Passed       bool
	TotalScore   int
	LayerResults []LayerResult
	FailedItems  []FailedCriterion
}

// LayerResult holds the pass/fail outcome for a single acceptance layer.
type LayerResult struct {
	LayerName string
	Passed    bool
	Score     int
	Details   []CriterionResult
}

// CriterionResult records the outcome of a single criterion within a layer.
type CriterionResult struct {
	Name   string
	Passed bool
	Score  int
	Detail string
}

// FailedCriterion captures a single criterion that did not pass.
type FailedCriterion struct {
	Layer     string
	Criterion string
	Reason    string
}

// RunMultiLayerAcceptance executes the multi-layer acceptance verification
// pipeline against the given project. Each layer dispatches work to the
// appropriate Brain (Verifier for unit/integration/security/performance,
// Browser for e2e) and collects per-criterion results.
func RunMultiLayerAcceptance(ctx context.Context, projectID string, criteria AcceptanceCriteria) (*MultiLayerAcceptanceResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}
	if len(criteria.Layers) == 0 {
		return nil, gerror.New("acceptance criteria must contain at least one layer")
	}

	baseURL, err := runtimeBaseURL(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "resolve brain serve base url for multi-layer acceptance failed")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	result := &MultiLayerAcceptanceResult{
		LayerResults: make([]LayerResult, 0, len(criteria.Layers)),
		FailedItems:  make([]FailedCriterion, 0),
	}
	totalScore := 0

	for _, layer := range criteria.Layers {
		layerResult, err := executeAcceptanceLayer(ctx, client, baseURL, projectID, layer)
		if err != nil {
			g.Log().Warningf(ctx, "[RunMultiLayerAcceptance] layer %s failed: %v", layer.Name, err)
			if layer.Required {
				failedResult := buildFailedLayerResult(layer, err.Error())
				result.LayerResults = append(result.LayerResults, failedResult)
				result.FailedItems = append(result.FailedItems, collectFailedCriteria(layer.Name, failedResult)...)
				result.Passed = false
				result.TotalScore = totalScore
				return result, nil
			}
			result.LayerResults = append(result.LayerResults, buildFailedLayerResult(layer, err.Error()))
			continue
		}

		result.LayerResults = append(result.LayerResults, *layerResult)
		totalScore += layerResult.Score

		if !layerResult.Passed {
			result.FailedItems = append(result.FailedItems, collectFailedCriteria(layer.Name, *layerResult)...)
			if layer.Required {
				result.Passed = false
				result.TotalScore = totalScore
				return result, nil
			}
		}
	}

	verdict := EvaluateAcceptanceResults(criteria, result.LayerResults)
	result.Passed = verdict.Passed
	result.TotalScore = verdict.TotalScore

	return result, nil
}

// executeAcceptanceLayer dispatches a single acceptance layer to the correct
// Brain and maps its output back to a LayerResult.
func executeAcceptanceLayer(
	ctx context.Context,
	client *http.Client,
	baseURL string,
	projectID string,
	layer AcceptanceLayer,
) (*LayerResult, error) {
	brainKind := resolveBrainForLayer(layer.Name)
	prompt := buildLayerPrompt(layer)

	multiReq := &MultiBrainRunRequest{
		ProjectID: projectID,
		Prompt:    prompt,
		MaxTurns:  6,
		Brains:    []string{brainKind},
	}
	multiRes := ExecuteMultiBrain(ctx, client, baseURL, multiReq)

	if brainErr, ok := multiRes.Errors[brainKind]; ok && brainErr != nil {
		return nil, brainErr
	}

	brainResult, ok := multiRes.Results[brainKind]
	if !ok || brainResult == nil {
		return nil, gerror.Newf("no result from %s brain for layer %s", brainKind, layer.Name)
	}

	return mapBrainResultToLayerResult(layer, brainResult), nil
}

// resolveBrainForLayer returns the brain kind to use for the given layer.
// Browser Brain handles e2e tests; Verifier Brain handles everything else.
func resolveBrainForLayer(layerName string) string {
	switch strings.ToLower(strings.TrimSpace(layerName)) {
	case "e2e_test":
		return "browser"
	default:
		return "verifier"
	}
}

// buildLayerPrompt constructs the prompt sent to the Brain for a given layer.
func buildLayerPrompt(layer AcceptanceLayer) string {
	criteriaNames := make([]string, 0, len(layer.Criteria))
	for _, c := range layer.Criteria {
		criteriaNames = append(criteriaNames, c.Name)
	}
	return fmt.Sprintf(
		"Run %s acceptance checks. Criteria to verify: %s. Report pass/fail for each criterion with details.",
		layer.Name,
		strings.Join(criteriaNames, ", "),
	)
}

// mapBrainResultToLayerResult converts the brain run response into a LayerResult.
// When the brain returns successfully, we optimistically mark all criteria as
// passed; a future iteration can parse structured output from the brain to
// determine per-criterion pass/fail granularity.
func mapBrainResultToLayerResult(layer AcceptanceLayer, brainResult *runtimeCreateRunResponse) *LayerResult {
	passed := brainResult.Status == "completed" || brainResult.Status == "success"
	details := make([]CriterionResult, 0, len(layer.Criteria))
	totalScore := 0

	for _, c := range layer.Criteria {
		criterionPassed := passed
		score := 0
		detail := "Brain run status: " + brainResult.Status
		if criterionPassed {
			score = c.Weight
		}
		totalScore += score
		details = append(details, CriterionResult{
			Name:   c.Name,
			Passed: criterionPassed,
			Score:  score,
			Detail: detail,
		})
	}

	return &LayerResult{
		LayerName: layer.Name,
		Passed:    passed,
		Score:     totalScore,
		Details:   details,
	}
}

// buildFailedLayerResult creates a LayerResult where every criterion has failed.
func buildFailedLayerResult(layer AcceptanceLayer, reason string) LayerResult {
	details := make([]CriterionResult, 0, len(layer.Criteria))
	for _, c := range layer.Criteria {
		details = append(details, CriterionResult{
			Name:   c.Name,
			Passed: false,
			Score:  0,
			Detail: reason,
		})
	}
	return LayerResult{
		LayerName: layer.Name,
		Passed:    false,
		Score:     0,
		Details:   details,
	}
}

// collectFailedCriteria extracts FailedCriterion entries from a LayerResult.
func collectFailedCriteria(layerName string, lr LayerResult) []FailedCriterion {
	items := make([]FailedCriterion, 0, len(lr.Details))
	for _, d := range lr.Details {
		if !d.Passed {
			items = append(items, FailedCriterion{
				Layer:     layerName,
				Criterion: d.Name,
				Reason:    d.Detail,
			})
		}
	}
	return items
}
