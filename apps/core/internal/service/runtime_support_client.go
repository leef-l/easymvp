package service

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"

	runtimev1 "github.com/leef-l/easymvp/apps/core/api/runtime/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/do"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const (
	runtimeRunEventDefaultLimit = 50
	runtimeRunEventMaxLimit     = 200
)

type normalizedStartBrainRunCommand struct {
	ProjectID string
	TaskID    string
	BrainKind string
	Prompt    string
	Workdir   string
	MaxTurns  int
	Provider  string
}

type runtimeCreateRunRequest struct {
	Prompt      string                  `json:"prompt"`
	Brain       string                  `json:"brain"`
	MaxTurns    int                     `json:"max_turns"`
	Workdir     string                  `json:"workdir,omitempty"`
	ModelConfig *runtimeModelConfigItem `json:"model_config,omitempty"`
}

type runtimeModelConfigItem struct {
	Provider string `json:"provider,omitempty"`
}

type runtimeCreateRunResponse struct {
	RunID       string `json:"run_id"`
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
}

type runtimeExecutionStopResponse struct {
	ExecutionID string `json:"execution_id"`
	Status      string `json:"status"`
}

type runtimeResumeRunResponse struct {
	RunID      string `json:"run_id"`
	StoreRunID string `json:"store_run_id"`
	ResumedAs  string `json:"resumed_as"`
	State      string `json:"state"`
	Turns      int    `json:"turns"`
	Reply      string `json:"reply"`
}

func normalizeStartBrainRunCommand(req StartBrainRunCommand) (*normalizedStartBrainRunCommand, error) {
	normalized := &normalizedStartBrainRunCommand{
		ProjectID: strings.TrimSpace(req.ProjectID),
		TaskID:    strings.TrimSpace(req.TaskID),
		BrainKind: strings.TrimSpace(req.BrainKind),
		Prompt:    strings.TrimSpace(req.Prompt),
		Workdir:   cleanProjectPath(req.Workdir),
		MaxTurns:  req.MaxTurns,
		Provider:  strings.TrimSpace(req.Provider),
	}
	if normalized.ProjectID == "" {
		return nil, gerror.New("project id is required")
	}
	if normalized.TaskID == "" {
		return nil, gerror.New("task id is required")
	}
	if normalized.Prompt == "" {
		return nil, gerror.New("prompt is required")
	}
	if normalized.BrainKind == "" {
		normalized.BrainKind = "central"
	}
	if normalized.Workdir != "" && !filepath.IsAbs(normalized.Workdir) {
		return nil, gerror.New("workdir must be an absolute path")
	}
	if normalized.MaxTurns <= 0 {
		normalized.MaxTurns = 20
	}
	if normalized.MaxTurns > 200 {
		return nil, gerror.New("max turns must be less than or equal to 200")
	}
	if len(normalized.BrainKind) > 64 {
		return nil, gerror.New("brain kind is too long")
	}
	if len(normalized.Provider) > 64 {
		return nil, gerror.New("provider is too long")
	}
	return normalized, nil
}

func runtimeBaseURL(ctx context.Context) (string, error) {
	baseURL := strings.TrimRight(strings.TrimSpace(CurrentStartupConfig(ctx).BrainServeBaseURL), "/")
	if baseURL == "" {
		return "", gerror.New("easymvp.brainServeBaseURL is empty")
	}
	return baseURL, nil
}

func runtimeHealthCheck(ctx context.Context, client *http.Client, baseURL string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+"/health", nil)
	if err != nil {
		return gerror.Wrap(err, "build runtime health request failed")
	}
	resp, err := client.Do(req)
	if err != nil {
		return gerror.Wrap(err, "runtime health request failed")
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return gerror.Newf("runtime health check failed: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func runtimeCreateRun(ctx context.Context, client *http.Client, baseURL string, reqData *normalizedStartBrainRunCommand) (*runtimeCreateRunResponse, error) {
	body := &runtimeCreateRunRequest{
		Prompt:   reqData.Prompt,
		Brain:    reqData.BrainKind,
		MaxTurns: reqData.MaxTurns,
		Workdir:  reqData.Workdir,
	}
	if reqData.Provider != "" {
		body.ModelConfig = &runtimeModelConfigItem{Provider: reqData.Provider}
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return nil, gerror.Wrap(err, "marshal runtime create request failed")
	}
	primaryPath := "/v1/executions"
	primaryRes, primaryErr := runtimeCreateRunAtPath(ctx, client, baseURL, primaryPath, payload)
	if primaryErr == nil {
		return primaryRes, nil
	}
	if !isRuntimeEndpointFallbackError(primaryErr) {
		return nil, primaryErr
	}
	return runtimeCreateRunAtPath(ctx, client, baseURL, "/v1/runs", payload)
}

func runtimeGetRun(ctx context.Context, client *http.Client, baseURL, runID string) (*BrainRunState, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, gerror.New("run id is required")
	}
	primaryRes, primaryErr := runtimeGetRunAtPath(ctx, client, baseURL, "/v1/executions/"+runID, runID)
	if primaryErr == nil {
		return primaryRes, nil
	}
	if !isRuntimeEndpointFallbackError(primaryErr) {
		return nil, primaryErr
	}
	return runtimeGetRunAtPath(ctx, client, baseURL, "/v1/runs/"+runID, runID)
}

func runtimeCancelRun(ctx context.Context, client *http.Client, baseURL, runID string) error {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return gerror.New("run id is required")
	}
	primaryErr := runtimeStopExecution(ctx, client, baseURL, runID)
	if primaryErr == nil {
		return nil
	}
	if !isRuntimeEndpointFallbackError(primaryErr) {
		return primaryErr
	}
	return runtimeCancelRunLegacy(ctx, client, baseURL, runID)
}

func runtimeCreateRunAtPath(ctx context.Context, client *http.Client, baseURL, path string, payload []byte) (*runtimeCreateRunResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+path, bytes.NewReader(payload))
	if err != nil {
		return nil, gerror.Wrap(err, "build runtime create request failed")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return nil, gerror.Wrap(err, "runtime create request failed")
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "read runtime create response failed")
	}
	if resp.StatusCode >= 300 {
		return nil, runtimeEndpointError("runtime create run failed", path, resp.StatusCode, raw)
	}

	var result runtimeCreateRunResponse
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, gerror.Wrap(err, "decode runtime create response failed")
	}
	if result.ExecutionID == "" && result.RunID != "" {
		result.ExecutionID = result.RunID
	}
	if result.RunID == "" && result.ExecutionID != "" {
		result.RunID = result.ExecutionID
	}
	result.Status = normalizeRuntimeStatus(result.Status)
	if result.RunID == "" {
		return nil, gerror.New("runtime create run returned empty run_id")
	}
	return &result, nil
}

func runtimeGetRunAtPath(ctx context.Context, client *http.Client, baseURL, path, runID string) (*BrainRunState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, baseURL+path, nil)
	if err != nil {
		return nil, gerror.Wrap(err, "build runtime get request failed")
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, gerror.Wrap(err, "runtime get request failed")
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, gerror.Wrap(err, "read runtime get response failed")
	}
	if resp.StatusCode >= 300 {
		return nil, runtimeEndpointError("runtime get run failed", path, resp.StatusCode, raw)
	}

	var result BrainRunState
	if err = json.Unmarshal(raw, &result); err != nil {
		return nil, gerror.Wrap(err, "decode runtime get response failed")
	}
	if result.ExecutionID == "" && result.RunID != "" {
		result.ExecutionID = result.RunID
	}
	if result.RunID == "" && result.ExecutionID != "" {
		result.RunID = result.ExecutionID
	}
	if result.RunID == "" {
		result.RunID = runID
	}
	if result.ExecutionID == "" {
		result.ExecutionID = result.RunID
	}
	result.Status = normalizeRuntimeStatus(result.Status)
	return &result, nil
}

func runtimeStopExecution(ctx context.Context, client *http.Client, baseURL, runID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, baseURL+"/v1/executions/"+runID+"/stop", nil)
	if err != nil {
		return gerror.Wrap(err, "build runtime stop request failed")
	}
	resp, err := client.Do(req)
	if err != nil {
		return gerror.Wrap(err, "runtime stop request failed")
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return gerror.Wrap(err, "read runtime stop response failed")
	}
	if resp.StatusCode >= 300 {
		return runtimeEndpointError("runtime stop execution failed", "/v1/executions/{id}/stop", resp.StatusCode, raw)
	}

	var result runtimeExecutionStopResponse
	if len(raw) > 0 {
		if err = json.Unmarshal(raw, &result); err != nil {
			return gerror.Wrap(err, "decode runtime stop response failed")
		}
	}
	return nil
}

func runtimeCancelRunLegacy(ctx context.Context, client *http.Client, baseURL, runID string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, baseURL+"/v1/runs/"+runID, nil)
	if err != nil {
		return gerror.Wrap(err, "build runtime cancel request failed")
	}
	resp, err := client.Do(req)
	if err != nil {
		return gerror.Wrap(err, "runtime cancel request failed")
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return runtimeEndpointError("runtime cancel run failed", "/v1/runs/{id}", resp.StatusCode, body)
	}
	return nil
}

func runtimeEndpointError(prefix, path string, statusCode int, body []byte) error {
	return gerror.Newf("%s: path=%s status=%d body=%s", prefix, path, statusCode, strings.TrimSpace(string(body)))
}

func isRuntimeEndpointFallbackError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "status=404") || strings.Contains(message, "status=405")
}

func runtimeResumeRun(ctx context.Context, runID string) (*runtimeResumeRunResponse, error) {
	runID = strings.TrimSpace(runID)
	if runID == "" {
		return nil, gerror.New("run id is required")
	}

	commandCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
	defer cancel()

	output, err := exec.CommandContext(commandCtx, "brain", "resume", runID, "--json").CombinedOutput()
	if err != nil {
		return nil, gerror.Wrapf(err, "runtime resume command failed: output=%s", strings.TrimSpace(string(output)))
	}

	var result runtimeResumeRunResponse
	if err = json.Unmarshal(output, &result); err != nil {
		return nil, gerror.Wrapf(err, "decode runtime resume response failed: output=%s", strings.TrimSpace(string(output)))
	}
	if strings.TrimSpace(result.RunID) == "" {
		result.RunID = runID
	}
	result.State = normalizeRuntimeStatus(result.State)
	if result.State == "" {
		return nil, gerror.New("runtime resume response returned empty state")
	}
	return &result, nil
}

func insertBrainRunBinding(ctx context.Context, row *do.BrainRunBindings) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	result, err := db.ExecContext(
		ctx,
		`INSERT INTO `+dao.BrainRunBindings.Table()+` (
id, project_id, task_id, brain_kind, brain_run_id, run_status, started_at, finished_at, last_sync_at, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		row.Id,
		row.ProjectId,
		row.TaskId,
		row.BrainKind,
		row.BrainRunId,
		row.RunStatus,
		row.StartedAt,
		nullIfEmpty(row.FinishedAt),
		nullIfEmpty(row.LastSyncAt),
		row.CreatedAt,
		row.UpdatedAt,
	)
	if err != nil {
		return gerror.Wrap(err, "insert brain run binding failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert brain run binding affected unexpected rows")
	}
	return nil
}

func getBrainRunBindingByID(ctx context.Context, bindingID string) (*entity.BrainRunBindings, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	query := `
SELECT
  id,
  project_id,
  task_id,
  brain_kind,
  brain_run_id,
  run_status,
  started_at,
  COALESCE(finished_at, ''),
  COALESCE(last_sync_at, ''),
  created_at,
  updated_at
FROM ` + dao.BrainRunBindings.Table() + `
WHERE id = ?
LIMIT 1`

	row := db.QueryRowContext(ctx, query, bindingID)
	var binding entity.BrainRunBindings
	if err = row.Scan(
		&binding.Id,
		&binding.ProjectId,
		&binding.TaskId,
		&binding.BrainKind,
		&binding.BrainRunId,
		&binding.RunStatus,
		&binding.StartedAt,
		&binding.FinishedAt,
		&binding.LastSyncAt,
		&binding.CreatedAt,
		&binding.UpdatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, gerror.Newf("brain run binding not found: %s", bindingID)
		}
		return nil, gerror.Wrap(err, "query brain run binding failed")
	}
	return &binding, nil
}

func listRunBindingEvents(ctx context.Context, bindingID string, limit int) ([]entity.RunEventIndex, error) {
	if strings.TrimSpace(bindingID) == "" {
		return nil, gerror.New("binding id is required")
	}
	if limit <= 0 {
		limit = runtimeRunEventDefaultLimit
	}
	if limit > runtimeRunEventMaxLimit {
		limit = runtimeRunEventMaxLimit
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(
		ctx,
		`SELECT
  id,
  project_id,
  run_binding_id,
  sequence_no,
  event_type,
  COALESCE(event_level, ''),
  summary,
  COALESCE(payload_json, ''),
  created_at
FROM `+dao.RunEventIndex.Table()+`
WHERE run_binding_id = ?
ORDER BY sequence_no DESC, id DESC
LIMIT ?`,
		bindingID,
		limit,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query run binding events failed")
	}
	defer rows.Close()

	items := make([]entity.RunEventIndex, 0, limit)
	for rows.Next() {
		var item entity.RunEventIndex
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.RunBindingId,
			&item.SequenceNo,
			&item.EventType,
			&item.EventLevel,
			&item.Summary,
			&item.PayloadJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan run binding event failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate run binding events failed")
	}

	for left, right := 0, len(items)-1; left < right; left, right = left+1, right-1 {
		items[left], items[right] = items[right], items[left]
	}
	return items, nil
}

func updateBrainRunBindingStatus(ctx context.Context, bindingID, runStatus string) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	now := nowText()
	finishedAt := ""
	if isTerminalBrainRunStatus(runStatus) {
		finishedAt = now
	}

	result, err := db.ExecContext(
		ctx,
		`UPDATE `+dao.BrainRunBindings.Table()+`
SET run_status = ?, finished_at = ?, last_sync_at = ?, updated_at = ?
WHERE id = ?`,
		runStatus,
		nullIfEmpty(finishedAt),
		now,
		now,
		bindingID,
	)
	if err != nil {
		return gerror.Wrap(err, "update brain run binding failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("update brain run binding affected unexpected rows")
	}
	return nil
}

func appendRunEventIndex(
	ctx context.Context,
	projectID string,
	runBindingID string,
	eventType string,
	eventLevel string,
	summary string,
	payload any,
) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return gerror.Wrap(err, "begin run event transaction failed")
	}

	var nextSequence int
	if err = tx.QueryRowContext(
		ctx,
		`SELECT COALESCE(MAX(sequence_no), 0) + 1 FROM `+dao.RunEventIndex.Table()+` WHERE run_binding_id = ?`,
		runBindingID,
	).Scan(&nextSequence); err != nil {
		_ = tx.Rollback()
		return gerror.Wrap(err, "query next run event sequence failed")
	}

	var payloadJSON any
	if payload != nil {
		encoded, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			_ = tx.Rollback()
			return gerror.Wrap(marshalErr, "marshal run event payload failed")
		}
		payloadJSON = string(encoded)
	}

	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO `+dao.RunEventIndex.Table()+` (
id, project_id, run_binding_id, sequence_no, event_type, event_level, summary, payload_json, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("runevt"),
		projectID,
		runBindingID,
		nextSequence,
		eventType,
		nullIfEmpty(eventLevel),
		summary,
		payloadJSON,
		nowText(),
	)
	if err != nil {
		_ = tx.Rollback()
		return gerror.Wrap(err, "insert run event failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		_ = tx.Rollback()
		return gerror.New("insert run event affected unexpected rows")
	}
	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "commit run event transaction failed")
	}
	return nil
}

func updateProjectTaskRuntimeStatus(ctx context.Context, taskID string, runStatus string) error {
	if strings.TrimSpace(taskID) == "" {
		return nil
	}

	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	taskStatus := domainTaskStatusForRunStatus(runStatus)
	result, err := db.ExecContext(
		ctx,
		`UPDATE `+dao.DomainTasks.Table()+` SET status = ?, updated_at = ? WHERE id = ?`,
		taskStatus,
		nowText(),
		taskID,
	)
	if err != nil {
		return gerror.Wrap(err, "update domain task runtime status failed")
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return nil
	}
	return nil
}

func mapBrainRunStatus(status string) string {
	switch normalizeRuntimeStatus(status) {
	case "queued", "waiting", "pending", "accepted", "created":
		return "run_pending"
	case "running", "in_progress", "streaming":
		return "run_active"
	case "succeeded", "success", "completed", "complete":
		return "run_succeeded"
	case "failed", "error", "crashed":
		return "run_failed"
	case "unsupported", "tool_unsupported", "not_supported":
		return "run_unsupported"
	case "denied", "permission_denied", "tool_denied", "forbidden":
		return "run_denied"
	case "cancelled", "canceled", "stopped", "aborted":
		return "run_cancelled"
	default:
		return "run_pending"
	}
}

func normalizeRuntimeStatus(status string) string {
	return strings.ToLower(strings.TrimSpace(status))
}

func isTerminalBrainRunStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case "run_succeeded", "run_failed", "run_unsupported", "run_denied", "run_cancelled":
		return true
	default:
		return false
	}
}

func domainTaskStatusForRunStatus(runStatus string) string {
	switch runStatus {
	case "run_pending":
		return "queued"
	case "run_active":
		return "running"
	case "run_succeeded":
		return "verify_pending"
	case "run_failed":
		return "failed"
	case "run_unsupported", "run_denied":
		return "blocked"
	case "run_cancelled":
		return "cancelled"
	default:
		return "queued"
	}
}

func eventLevelForRunStatus(runStatus string) string {
	switch runStatus {
	case "run_failed":
		return "error"
	case "run_denied", "run_unsupported", "run_cancelled":
		return "warning"
	default:
		return "info"
	}
}

func mapRunBindingView(binding *entity.BrainRunBindings) runtimev1.RunBindingView {
	return runtimev1.RunBindingView{
		BindingID:  binding.Id,
		ProjectID:  binding.ProjectId,
		TaskID:     binding.TaskId,
		BrainKind:  binding.BrainKind,
		RunID:      binding.BrainRunId,
		RunStatus:  binding.RunStatus,
		StartedAt:  binding.StartedAt,
		FinishedAt: binding.FinishedAt,
		LastSyncAt: binding.LastSyncAt,
	}
}
