package task

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

var runningTaskCancels sync.Map

type runtimeTask struct {
	ID           int64
	Title        string
	EngineCode   string
	RoleID       int64
	UserID       int64
	ProjectID    int64
	RepoPath     string
	WorktreePath string
	BranchName   string
	Instruction  string
	Status       string
}

type runtimeEngineConfig struct {
	EngineCode      string
	Name            string
	BaseURL         string
	APIKey          string
	DefaultModelID  int64
	TimeoutSeconds  int
	MaxSteps        int
	WorkspaceRoot   string
	CommandTemplate string
	CallbackURL     string
	CallbackSecret  string
	ExtraConfig     string
	Status          int
	ConfigStatus    int
}

type runtimeModelInfo struct {
	ModelID       int64
	ModelCode     string
	ProviderType  string
	BaseURL       string
	APIKey        string
	APISecret     string
	MaxTokens     int
	ContextWindow int
}

type taskExecutionResult struct {
	Summary string
	Output  string
}

type aiderExecutionConfig struct {
	Model                *runtimeModelInfo
	WorkDir              string
	Message              string
	Timeout              time.Duration
	MapTokens            int
	MaxChatHistoryTokens int
	CompactMode          bool
}

type aiderExecutionResult struct {
	Output      string
	ExitCode    int
	Error       error
	FailureHint string
}

func (s *sTask) dispatchTask(taskID int64) {
	go s.runTask(taskID)
}

func (s *sTask) runTask(taskID int64) {
	dbCtx := context.Background()

	taskInfo, err := s.loadTask(dbCtx, taskID)
	if err != nil {
		g.Log().Errorf(dbCtx, "[ai-task] load task failed: taskID=%d err=%v", taskID, err)
		return
	}

	if taskInfo.Status == "cancelled" {
		return
	}

	engineCfg, err := s.loadEngineRuntimeConfig(dbCtx, taskInfo.EngineCode)
	if err != nil {
		s.markTaskFailed(dbCtx, taskID, "", err.Error())
		return
	}

	if ok, err := s.markTaskRunning(dbCtx, taskID); err != nil {
		s.markTaskFailed(dbCtx, taskID, "", err.Error())
		return
	} else if !ok {
		return
	}

	timeout := time.Duration(engineCfg.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}

	execCtx, cancel := context.WithTimeout(context.Background(), timeout)
	runningTaskCancels.Store(taskID, cancel)
	defer func() {
		cancel()
		runningTaskCancels.Delete(taskID)
	}()

	_ = s.appendTaskLog(dbCtx, taskID, "system", fmt.Sprintf("任务开始执行。引擎=%s", taskInfo.EngineCode))

	var result *taskExecutionResult
	switch taskInfo.EngineCode {
	case "aider":
		result, err = s.executeWithAider(execCtx, taskInfo, engineCfg)
	case "openhands":
		result, err = s.executeWithOpenHands(execCtx, taskInfo, engineCfg)
	default:
		err = fmt.Errorf("暂不支持执行引擎: %s", taskInfo.EngineCode)
	}

	if errors.Is(err, context.Canceled) || errors.Is(execCtx.Err(), context.Canceled) {
		_ = s.appendTaskLog(dbCtx, taskID, "system", "任务执行已取消。")
		return
	}

	if errors.Is(err, context.DeadlineExceeded) || errors.Is(execCtx.Err(), context.DeadlineExceeded) {
		s.markTaskFailed(dbCtx, taskID, resultOutput(result), "任务执行超时")
		return
	}

	if err != nil {
		s.markTaskFailed(dbCtx, taskID, resultOutput(result), err.Error())
		return
	}

	s.markTaskSuccess(dbCtx, taskID, result)
}

func (s *sTask) authorizeEngineUsage(ctx context.Context, userID int64, engineCode string) (int64, error) {
	if userID == 1 {
		return 0, nil
	}

	record, err := g.DB().Ctx(ctx).Model("system_user_role ur").
		InnerJoin("system_role r", "r.id = ur.role_id AND r.deleted_at IS NULL AND r.status = 1").
		InnerJoin("system_role_ai_engine rae", "rae.role_id = ur.role_id").
		Fields("ur.role_id").
		Where("ur.user_id", userID).
		Where("rae.engine_code", engineCode).
		One()
	if err != nil {
		return 0, err
	}
	if record.IsEmpty() {
		return 0, fmt.Errorf("当前角色未授权使用执行引擎 %s", engineCode)
	}
	return record["role_id"].Int64(), nil
}

func (s *sTask) resolveTaskPaths(workspaceRoot, repoPath, worktreePath string) (string, string, error) {
	repoPath = strings.TrimSpace(repoPath)
	worktreePath = strings.TrimSpace(worktreePath)
	workspaceRoot = strings.TrimSpace(workspaceRoot)

	if repoPath == "" {
		return "", "", fmt.Errorf("仓库路径不能为空")
	}

	if workspaceRoot != "" && !filepath.IsAbs(repoPath) {
		repoPath = filepath.Join(workspaceRoot, repoPath)
	}
	if worktreePath == "" {
		worktreePath = repoPath
	} else if workspaceRoot != "" && !filepath.IsAbs(worktreePath) {
		worktreePath = filepath.Join(workspaceRoot, worktreePath)
	}

	repoPath = filepath.Clean(repoPath)
	worktreePath = filepath.Clean(worktreePath)
	return repoPath, worktreePath, nil
}

func (s *sTask) loadTask(ctx context.Context, taskID int64) (*runtimeTask, error) {
	record, err := g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, fmt.Errorf("任务不存在")
	}
	return &runtimeTask{
		ID:           record["id"].Int64(),
		Title:        record["title"].String(),
		EngineCode:   record["engine_code"].String(),
		RoleID:       record["role_id"].Int64(),
		UserID:       record["user_id"].Int64(),
		ProjectID:    record["project_id"].Int64(),
		RepoPath:     record["repo_path"].String(),
		WorktreePath: record["worktree_path"].String(),
		BranchName:   record["branch_name"].String(),
		Instruction:  record["instruction"].String(),
		Status:       record["status"].String(),
	}, nil
}

func (s *sTask) loadEngineRuntimeConfig(ctx context.Context, engineCode string) (*runtimeEngineConfig, error) {
	record, err := g.DB().Ctx(ctx).Model("ai_engine e").
		LeftJoin("ai_engine_config c", "c.engine_code = e.code AND c.deleted_at IS NULL").
		Fields("e.code AS engine_code, e.name, e.status, COALESCE(c.status, 0) AS config_status, c.base_url, c.api_key, COALESCE(c.default_model_id, 0) AS default_model_id, COALESCE(c.timeout_seconds, 600) AS timeout_seconds, COALESCE(c.max_steps, 20) AS max_steps, c.workspace_root, c.command_template, c.callback_url, c.callback_secret, c.extra_config").
		Where("e.code", engineCode).
		Where("e.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, fmt.Errorf("执行引擎不存在")
	}
	cfg := &runtimeEngineConfig{
		EngineCode:      record["engine_code"].String(),
		Name:            record["name"].String(),
		BaseURL:         record["base_url"].String(),
		APIKey:          record["api_key"].String(),
		DefaultModelID:  record["default_model_id"].Int64(),
		TimeoutSeconds:  record["timeout_seconds"].Int(),
		MaxSteps:        record["max_steps"].Int(),
		WorkspaceRoot:   record["workspace_root"].String(),
		CommandTemplate: record["command_template"].String(),
		CallbackURL:     record["callback_url"].String(),
		CallbackSecret:  record["callback_secret"].String(),
		ExtraConfig:     record["extra_config"].String(),
		Status:          record["status"].Int(),
		ConfigStatus:    record["config_status"].Int(),
	}
	if cfg.DefaultModelID > 0 {
		modelInfo, modelErr := s.loadModelRuntimeInfo(ctx, cfg.DefaultModelID)
		if modelErr != nil {
			return nil, modelErr
		}
		cfg.BaseURL = modelInfo.BaseURL
		cfg.APIKey = modelInfo.APIKey
	}
	return cfg, nil
}

func (s *sTask) loadModelRuntimeInfo(ctx context.Context, modelID int64) (*runtimeModelInfo, error) {
	record, err := g.DB().Ctx(ctx).Model("ai_model m").
		LeftJoin("ai_plan p", "p.id = m.plan_id").
		LeftJoin("ai_provider pv", "pv.id = m.provider_id").
		Fields("m.id, m.model_code, COALESCE(m.max_tokens, 4096) AS max_tokens, COALESCE(m.context_window, 128000) AS context_window, pv.provider_type, pv.base_url, p.api_key, p.api_secret").
		Where("m.id", modelID).
		Where("m.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, fmt.Errorf("默认模型不存在")
	}
	return &runtimeModelInfo{
		ModelID:       record["id"].Int64(),
		ModelCode:     record["model_code"].String(),
		ProviderType:  record["provider_type"].String(),
		BaseURL:       record["base_url"].String(),
		APIKey:        record["api_key"].String(),
		APISecret:     record["api_secret"].String(),
		MaxTokens:     record["max_tokens"].Int(),
		ContextWindow: record["context_window"].Int(),
	}, nil
}

func (s *sTask) markTaskRunning(ctx context.Context, taskID int64) (bool, error) {
	result, err := g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		Where("status", "pending").
		Data(g.Map{
			"status":     "running",
			"started_at": gtime.Now(),
			"updated_at": gtime.Now(),
		}).
		Update()
	if err != nil {
		return false, err
	}
	if result == nil {
		return false, nil
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return rows > 0, nil
}

func (s *sTask) markTaskSuccess(ctx context.Context, taskID int64, result *taskExecutionResult) {
	if result == nil {
		result = &taskExecutionResult{}
	}
	_, _ = g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		Where("status", "running").
		Data(g.Map{
			"status":           "success",
			"response_summary": result.Summary,
			"error_message":    "",
			"finished_at":      gtime.Now(),
			"updated_at":       gtime.Now(),
		}).
		Update()
	if strings.TrimSpace(result.Output) != "" {
		_ = s.appendTaskLog(ctx, taskID, "stdout", result.Output)
	}
	_ = s.appendTaskLog(ctx, taskID, "system", "任务执行完成。")
}

func (s *sTask) markTaskFailed(ctx context.Context, taskID int64, output string, errMessage string) {
	_, _ = g.DB().Ctx(ctx).Model("ai_task").
		Where("id", taskID).
		Where("deleted_at IS NULL").
		WhereIn("status", []string{"pending", "running"}).
		Data(g.Map{
			"status":           "failed",
			"response_summary": shortenText(output, 2000),
			"error_message":    errMessage,
			"finished_at":      gtime.Now(),
			"updated_at":       gtime.Now(),
		}).
		Update()
	if strings.TrimSpace(output) != "" {
		_ = s.appendTaskLog(ctx, taskID, "stderr", output)
	}
	_ = s.appendTaskLog(ctx, taskID, "system", "任务执行失败："+errMessage)
}

func (s *sTask) appendTaskLog(ctx context.Context, taskID int64, logType string, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}

	count, err := g.DB().Ctx(ctx).Model("ai_task_log").Where("task_id", taskID).Count()
	if err != nil {
		return err
	}

	_, err = g.DB().Ctx(ctx).Model("ai_task_log").Data(g.Map{
		"id":         snowflake.Generate(),
		"task_id":    taskID,
		"seq":        count + 1,
		"log_type":   logType,
		"content":    content,
		"created_at": gtime.Now(),
	}).Insert()
	return err
}

func (s *sTask) executeWithAider(ctx context.Context, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig) (*taskExecutionResult, error) {
	if engineCfg.DefaultModelID == 0 {
		return nil, fmt.Errorf("Aider 默认模型未配置")
	}

	modelInfo, err := s.loadModelRuntimeInfo(context.Background(), engineCfg.DefaultModelID)
	if err != nil {
		return nil, err
	}
	if modelInfo.APIKey == "" {
		return nil, fmt.Errorf("Aider 所使用模型的 API Key 未配置")
	}

	if err := ensureDirExists(taskInfo.WorktreePath); err != nil {
		return nil, err
	}

	cfg := &aiderExecutionConfig{
		Model:                modelInfo,
		WorkDir:              taskInfo.WorktreePath,
		Message:              taskInfo.Instruction,
		Timeout:              time.Duration(engineCfg.TimeoutSeconds) * time.Second,
		MapTokens:            512,
		MaxChatHistoryTokens: 2048,
	}

	_ = s.appendTaskLog(context.Background(), taskInfo.ID, "system", "Aider 将优先使用本机安装或 uv 执行，缺失时再回退到 Docker。")
	result := s.runAider(ctx, cfg)
	if result.Error != nil {
		errMsg := result.FailureHint
		if errMsg == "" {
			errMsg = result.Error.Error()
		}
		return &taskExecutionResult{
			Summary: shortenText(result.Output, 1200),
			Output:  result.Output,
		}, fmt.Errorf("Aider执行失败(code=%d): %s", result.ExitCode, errMsg)
	}

	return &taskExecutionResult{
		Summary: shortenText(result.Output, 1200),
		Output:  result.Output,
	}, nil
}

func (s *sTask) executeWithOpenHands(ctx context.Context, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig) (*taskExecutionResult, error) {
	if strings.TrimSpace(engineCfg.CommandTemplate) != "" {
		return s.executeWithCommandTemplate(ctx, taskInfo, engineCfg)
	}
	if result, err, ok := s.executeWithOpenHandsCLI(ctx, taskInfo, engineCfg); ok {
		return result, err
	}
	return s.executeWithOpenHandsHTTP(ctx, taskInfo, engineCfg)
}

func (s *sTask) executeWithOpenHandsCLI(ctx context.Context, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig) (*taskExecutionResult, error, bool) {
	if runtime.GOOS == "windows" {
		return nil, nil, false
	}
	if err := ensureDirExists(taskInfo.WorktreePath); err != nil {
		return nil, err, true
	}

	if _, err := exec.LookPath("openhands"); err != nil {
		if _, uvErr := exec.LookPath("uv"); uvErr != nil {
			return nil, nil, false
		}
	}

	var modelInfo *runtimeModelInfo
	if engineCfg.DefaultModelID > 0 {
		loadedModelInfo, err := s.loadModelRuntimeInfo(context.Background(), engineCfg.DefaultModelID)
		if err != nil {
			return nil, err, true
		}
		modelInfo = loadedModelInfo
	}
	if modelInfo == nil || strings.TrimSpace(modelInfo.APIKey) == "" {
		return nil, fmt.Errorf("OpenHands 所使用模型的 API Key 未配置"), true
	}

	cmd, err := buildOpenHandsCLICommand(ctx, taskInfo, modelInfo)
	if err != nil {
		return nil, err, true
	}

	_ = s.appendTaskLog(context.Background(), taskInfo.ID, "system", "OpenHands 将通过官方 CLI/uv 路径执行。")

	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output := strings.TrimSpace(stdout.String() + stderr.String())
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return &taskExecutionResult{
			Summary: shortenText(output, 1200),
			Output:  output,
		}, err, true
	}

	if output == "" {
		output = "OpenHands CLI 已执行完成。"
	}

	return &taskExecutionResult{
		Summary: shortenText(output, 1200),
		Output:  output,
	}, nil, true
}

func (s *sTask) executeWithCommandTemplate(ctx context.Context, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig) (*taskExecutionResult, error) {
	if err := ensureDirExists(taskInfo.WorktreePath); err != nil {
		return nil, err
	}

	var modelInfo *runtimeModelInfo
	if engineCfg.DefaultModelID > 0 {
		modelInfo, _ = s.loadModelRuntimeInfo(context.Background(), engineCfg.DefaultModelID)
	}

	command := renderCommandTemplate(engineCfg.CommandTemplate, taskInfo, engineCfg, modelInfo)
	cmd := buildShellCommand(ctx, command)
	cmd.Dir = taskInfo.WorktreePath
	cmd.Env = append(os.Environ(), buildCommandTemplateEnv(taskInfo, engineCfg, modelInfo)...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	output := strings.TrimSpace(stdout.String() + stderr.String())
	if err != nil {
		if output == "" {
			output = err.Error()
		}
		return &taskExecutionResult{
			Summary: shortenText(output, 1200),
			Output:  output,
		}, err
	}

	if output == "" {
		output = "命令已执行完成。"
	}

	return &taskExecutionResult{
		Summary: shortenText(output, 1200),
		Output:  output,
	}, nil
}

func (s *sTask) executeWithOpenHandsHTTP(ctx context.Context, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig) (*taskExecutionResult, error) {
	if strings.TrimSpace(engineCfg.BaseURL) == "" {
		return nil, fmt.Errorf("OpenHands Base URL 未配置")
	}

	var modelInfo *runtimeModelInfo
	if engineCfg.DefaultModelID > 0 {
		modelInfo, _ = s.loadModelRuntimeInfo(context.Background(), engineCfg.DefaultModelID)
	}

	extraConfig := parseJSONMap(engineCfg.ExtraConfig)
	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if strings.TrimSpace(engineCfg.APIKey) != "" {
		headers["Authorization"] = "Bearer " + strings.TrimSpace(engineCfg.APIKey)
		headers["X-API-Key"] = strings.TrimSpace(engineCfg.APIKey)
	}
	if rawHeaders, ok := extraConfig["headers"].(map[string]interface{}); ok {
		for key, value := range rawHeaders {
			headers[key] = fmt.Sprint(value)
		}
	}

	payload := map[string]interface{}{
		"taskId":       taskInfo.ID,
		"title":        taskInfo.Title,
		"instruction":  taskInfo.Instruction,
		"repoPath":     taskInfo.RepoPath,
		"worktreePath": taskInfo.WorktreePath,
		"branchName":   taskInfo.BranchName,
		"engineCode":   taskInfo.EngineCode,
		"maxSteps":     engineCfg.MaxSteps,
		"callbackUrl":  engineCfg.CallbackURL,
	}
	if modelInfo != nil {
		payload["llm"] = map[string]interface{}{
			"model":        modelInfo.ModelCode,
			"providerType": modelInfo.ProviderType,
			"baseUrl":      modelInfo.BaseURL,
			"apiKey":       modelInfo.APIKey,
		}
	}
	if len(extraConfig) > 0 {
		payload["extraConfig"] = extraConfig
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	method := "POST"
	if rawMethod, ok := extraConfig["method"].(string); ok && strings.TrimSpace(rawMethod) != "" {
		method = strings.ToUpper(strings.TrimSpace(rawMethod))
	}

	req, err := http.NewRequestWithContext(ctx, method, strings.TrimSpace(engineCfg.BaseURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	respText := strings.TrimSpace(string(respBody))
	if resp.StatusCode >= 400 {
		if respText == "" {
			respText = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return &taskExecutionResult{
			Summary: shortenText(respText, 1200),
			Output:  respText,
		}, fmt.Errorf("OpenHands 请求失败: HTTP %d", resp.StatusCode)
	}

	summary := extractResponseSummary(respText)
	if summary == "" {
		summary = fmt.Sprintf("OpenHands 请求完成，HTTP %d", resp.StatusCode)
	}

	return &taskExecutionResult{
		Summary: summary,
		Output:  respText,
	}, nil
}

func (s *sTask) runAider(ctx context.Context, cfg *aiderExecutionConfig) *aiderExecutionResult {
	result := s.runAiderOnce(ctx, cfg)
	if result.Error != nil && isTokenLimitFailure(result.Output) && !cfg.CompactMode {
		compact := *cfg
		compact.CompactMode = true
		compact.MapTokens = 0
		compact.MaxChatHistoryTokens = 512
		compact.Message = compactAiderMessage(cfg.Message)
		retry := s.runAiderOnce(ctx, &compact)
		retry.Output = strings.TrimSpace(result.Output) + "\n\n[Aider] 命中 token limit，已自动切换精简上下文重试。\n\n" + strings.TrimSpace(retry.Output)
		if retry.Error == nil {
			return retry
		}
		result = retry
	}
	result.FailureHint = buildAiderFailureHint(result.Output)
	return result
}

func (s *sTask) runAiderOnce(ctx context.Context, cfg *aiderExecutionConfig) *aiderExecutionResult {
	metadataFile, _ := writeAiderMetadata(cfg.Model)
	if metadataFile != "" {
		defer os.Remove(metadataFile)
	}

	messageFile, _ := writeTempFile("aider-message-", ".txt", cfg.Message)
	if messageFile != "" {
		defer os.Remove(messageFile)
	}

	chatHistoryFile, _ := writeTempFile("aider-chat-history-", ".md", "")
	if chatHistoryFile != "" {
		defer os.Remove(chatHistoryFile)
	}

	inputHistoryFile, _ := writeTempFile("aider-input-history-", ".txt", "")
	if inputHistoryFile != "" {
		defer os.Remove(inputHistoryFile)
	}

	llmHistoryFile, _ := writeTempFile("aider-llm-history-", ".jsonl", "")
	if llmHistoryFile != "" {
		defer os.Remove(llmHistoryFile)
	}

	args := buildAiderArgs(cfg, metadataFile, messageFile, chatHistoryFile, inputHistoryFile, llmHistoryFile)
	cmd, cleanup, err := s.buildAiderCommand(ctx, cfg, args, metadataFile, messageFile, chatHistoryFile, inputHistoryFile, llmHistoryFile)
	if cleanup != nil {
		defer cleanup()
	}
	output := ""
	if err != nil {
		return &aiderExecutionResult{
			Output: err.Error(),
			Error:  err,
		}
	}

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err = cmd.Run()
	output = stdout.String() + stderr.String()
	result := &aiderExecutionResult{
		Output: strings.TrimSpace(output),
		Error:  err,
	}
	if exitErr, ok := err.(*exec.ExitError); ok {
		result.ExitCode = exitErr.ExitCode()
	}
	return result
}

func (s *sTask) buildAiderCommand(ctx context.Context, cfg *aiderExecutionConfig, args []string, metadataFile, messageFile, chatHistoryFile, inputHistoryFile, llmHistoryFile string) (*exec.Cmd, func(), error) {
	if _, err := exec.LookPath("aider"); err == nil {
		cmd := exec.CommandContext(ctx, "aider", args...)
		cmd.Dir = cfg.WorkDir
		cmd.Env = append(os.Environ(), buildAiderEnv(cfg.Model)...)
		return cmd, nil, nil
	}

	if _, err := exec.LookPath("uv"); err == nil {
		uvArgs := []string{"tool", "run", "--python", "3.12", "--from", "aider-chat", "aider"}
		uvArgs = append(uvArgs, args...)
		cmd := exec.CommandContext(ctx, "uv", uvArgs...)
		cmd.Dir = cfg.WorkDir
		cmd.Env = append(os.Environ(), buildAiderEnv(cfg.Model)...)
		return cmd, nil, nil
	}

	if _, err := exec.LookPath("docker"); err != nil {
		return nil, nil, fmt.Errorf("未找到 aider 可执行文件，且 docker 不可用")
	}

	workDir := filepath.Clean(cfg.WorkDir)
	dockerArgs := []string{
		"run", "--rm",
		"-v", workDir + ":/app",
		"-w", "/app",
	}
	for _, envItem := range buildAiderEnv(cfg.Model) {
		dockerArgs = append(dockerArgs, "-e", envItem)
	}

	containerMetadata := ""
	if metadataFile != "" {
		containerMetadata = "/tmp/" + filepath.Base(metadataFile)
		dockerArgs = append(dockerArgs, "-v", metadataFile+":"+containerMetadata)
	}
	containerMessage := ""
	if messageFile != "" {
		containerMessage = "/tmp/" + filepath.Base(messageFile)
		dockerArgs = append(dockerArgs, "-v", messageFile+":"+containerMessage)
	}
	containerChatHistory := ""
	if chatHistoryFile != "" {
		containerChatHistory = "/tmp/" + filepath.Base(chatHistoryFile)
		dockerArgs = append(dockerArgs, "-v", chatHistoryFile+":"+containerChatHistory)
	}
	containerInputHistory := ""
	if inputHistoryFile != "" {
		containerInputHistory = "/tmp/" + filepath.Base(inputHistoryFile)
		dockerArgs = append(dockerArgs, "-v", inputHistoryFile+":"+containerInputHistory)
	}
	containerLLMHistory := ""
	if llmHistoryFile != "" {
		containerLLMHistory = "/tmp/" + filepath.Base(llmHistoryFile)
		dockerArgs = append(dockerArgs, "-v", llmHistoryFile+":"+containerLLMHistory)
	}

	dockerArgs = append(dockerArgs, "paulgauthier/aider")
	dockerArgs = append(dockerArgs, buildAiderArgs(cfg, containerMetadata, containerMessage, containerChatHistory, containerInputHistory, containerLLMHistory)...)

	cmd := exec.CommandContext(ctx, "docker", dockerArgs...)
	cmd.Dir = workDir
	return cmd, nil, nil
}

func buildAiderArgs(cfg *aiderExecutionConfig, metadataFile, messageFile, chatHistoryFile, inputHistoryFile, llmHistoryFile string) []string {
	mapTokens := cfg.MapTokens
	if mapTokens == 0 && !cfg.CompactMode {
		mapTokens = 512
	}

	maxChatHistoryTokens := cfg.MaxChatHistoryTokens
	if maxChatHistoryTokens == 0 {
		if cfg.CompactMode {
			maxChatHistoryTokens = 512
		} else {
			maxChatHistoryTokens = 2048
		}
	}

	args := []string{
		"--model", formatAiderModel(cfg.Model),
		"--no-auto-commits",
		"--no-gitignore",
		"--no-check-update",
		"--no-show-release-notes",
		"--no-gui",
		"--no-show-model-warnings",
		"--no-pretty",
		"--no-stream",
		"--no-browser",
		"--no-restore-chat-history",
		"--yes-always",
		"--chat-language", "Chinese",
		"--map-tokens", strconv.Itoa(mapTokens),
		"--max-chat-history-tokens", strconv.Itoa(maxChatHistoryTokens),
	}
	if metadataFile != "" {
		args = append(args, "--model-metadata-file", metadataFile)
	}
	if cfg.CompactMode {
		args = append(args, "--subtree-only")
	}
	if chatHistoryFile != "" {
		args = append(args, "--chat-history-file", chatHistoryFile)
	}
	if inputHistoryFile != "" {
		args = append(args, "--input-history-file", inputHistoryFile)
	}
	if llmHistoryFile != "" {
		args = append(args, "--llm-history-file", llmHistoryFile)
	}
	if messageFile != "" {
		args = append(args, "--message-file", messageFile)
	} else if strings.TrimSpace(cfg.Message) != "" {
		args = append(args, "--message", cfg.Message)
	}
	return args
}

func buildAiderEnv(model *runtimeModelInfo) []string {
	env := []string{
		"PYTHONIOENCODING=utf-8",
		"PYTHONLEGACYWINDOWSSTDIO=0",
		"UV_LOCK_TIMEOUT=600",
	}
	switch model.ProviderType {
	case "openai_compatible":
		env = append(env,
			"OPENAI_API_KEY="+model.APIKey,
			"OPENAI_API_BASE="+strings.TrimRight(model.BaseURL, "/"),
		)
	default:
		baseURL := strings.TrimRight(strings.TrimSuffix(model.BaseURL, "/v1"), "/")
		env = append(env,
			"ANTHROPIC_API_KEY="+model.APIKey,
			"ANTHROPIC_BASE_URL="+baseURL,
		)
	}
	return env
}

func writeAiderMetadata(model *runtimeModelInfo) (string, error) {
	if model == nil {
		return "", nil
	}
	maxOutput := model.MaxTokens
	if maxOutput == 0 {
		maxOutput = 4096
	}
	contextWindow := model.ContextWindow
	if contextWindow == 0 {
		contextWindow = 128000
	}

	metadata := map[string]map[string]int{
		formatAiderModel(model): {
			"max_tokens":        maxOutput,
			"max_input_tokens":  contextWindow,
			"max_output_tokens": maxOutput,
		},
	}

	data, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return "", err
	}
	return writeTempFile("aider-metadata-", ".json", string(data))
}

func writeTempFile(prefix, ext, content string) (string, error) {
	file := filepath.Join(os.TempDir(), fmt.Sprintf("%s%d%s", prefix, time.Now().UnixNano(), ext))
	if err := os.WriteFile(file, []byte(content), 0644); err != nil {
		return "", err
	}
	return file, nil
}

func formatAiderModel(model *runtimeModelInfo) string {
	if model == nil {
		return ""
	}
	switch model.ProviderType {
	case "openai_compatible":
		if strings.HasPrefix(model.ModelCode, "openai/") {
			return model.ModelCode
		}
		return "openai/" + model.ModelCode
	default:
		if strings.HasPrefix(model.ModelCode, "anthropic/") {
			return model.ModelCode
		}
		return "anthropic/" + model.ModelCode
	}
}

func compactAiderMessage(message string) string {
	message = strings.TrimSpace(message)
	if len(message) > 2500 {
		message = message[:2500] + "\n...(已自动截断上下文)"
	}
	return message + "\n\n请仅完成最小必要修改，优先处理最核心问题。"
}

func isTokenLimitFailure(output string) bool {
	lower := strings.ToLower(output)
	keywords := []string{
		"token-limits.html",
		"hit a token limit",
		"exceeded output limit",
		"input tokens:",
		"output tokens:",
		"context window",
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
}

func buildAiderFailureHint(output string) string {
	if strings.TrimSpace(output) == "" {
		return "Aider 无输出即退出，请检查 aider 可执行文件、模型配置和工作目录。"
	}
	if isTokenLimitFailure(output) {
		return "Aider 命中了 token limit，已尝试精简上下文重试但仍失败。建议缩小任务范围或切换上下文更大的模型。"
	}
	return shortenText(output, 600)
}

func buildCommandTemplateEnv(taskInfo *runtimeTask, engineCfg *runtimeEngineConfig, modelInfo *runtimeModelInfo) []string {
	env := []string{
		"AI_TASK_ID=" + strconv.FormatInt(taskInfo.ID, 10),
		"AI_TASK_TITLE=" + taskInfo.Title,
		"AI_TASK_ENGINE_CODE=" + taskInfo.EngineCode,
		"AI_TASK_REPO_PATH=" + taskInfo.RepoPath,
		"AI_TASK_WORKTREE_PATH=" + taskInfo.WorktreePath,
		"AI_TASK_BRANCH_NAME=" + taskInfo.BranchName,
		"AI_TASK_INSTRUCTION=" + taskInfo.Instruction,
		"AI_ENGINE_BASE_URL=" + engineCfg.BaseURL,
		"AI_ENGINE_API_KEY=" + engineCfg.APIKey,
		"AI_ENGINE_MAX_STEPS=" + strconv.Itoa(engineCfg.MaxSteps),
		"AI_ENGINE_CALLBACK_URL=" + engineCfg.CallbackURL,
		"AI_ENGINE_CALLBACK_SECRET=" + engineCfg.CallbackSecret,
	}
	if modelInfo != nil {
		env = append(env,
			"AI_MODEL_ID="+strconv.FormatInt(modelInfo.ModelID, 10),
			"AI_MODEL_CODE="+modelInfo.ModelCode,
			"AI_MODEL_PROVIDER_TYPE="+modelInfo.ProviderType,
			"AI_MODEL_BASE_URL="+modelInfo.BaseURL,
			"AI_MODEL_API_KEY="+modelInfo.APIKey,
		)
	}
	return env
}

func buildOpenHandsCLICommand(ctx context.Context, taskInfo *runtimeTask, modelInfo *runtimeModelInfo) (*exec.Cmd, error) {
	args := []string{
		"--headless",
		"--json",
		"--override-with-envs",
		"--always-approve",
		"--exit-without-confirmation",
		"-t", taskInfo.Instruction,
	}
	if _, err := exec.LookPath("openhands"); err == nil {
		cmd := exec.CommandContext(ctx, "openhands", args...)
		cmd.Dir = taskInfo.WorktreePath
		cmd.Env = append(os.Environ(), buildOpenHandsCLIEnv(taskInfo, modelInfo)...)
		return cmd, nil
	}
	if _, err := exec.LookPath("uv"); err == nil {
		uvArgs := []string{"tool", "run", "--python", "3.12", "openhands"}
		uvArgs = append(uvArgs, args...)
		cmd := exec.CommandContext(ctx, "uv", uvArgs...)
		cmd.Dir = taskInfo.WorktreePath
		cmd.Env = append(os.Environ(), buildOpenHandsCLIEnv(taskInfo, modelInfo)...)
		return cmd, nil
	}
	return nil, fmt.Errorf("未找到 OpenHands CLI，且 uv 不可用")
}

func buildOpenHandsCLIEnv(taskInfo *runtimeTask, modelInfo *runtimeModelInfo) []string {
	var env = []string{
		"LLM_API_KEY=" + modelInfo.APIKey,
		"LLM_MODEL=" + formatOpenHandsModel(modelInfo),
		"LLM_BASE_URL=" + strings.TrimRight(modelInfo.BaseURL, "/"),
		"SANDBOX_VOLUMES=" + taskInfo.WorktreePath + ":/workspace:rw",
		"UV_LOCK_TIMEOUT=600",
	}
	return env
}

func formatOpenHandsModel(model *runtimeModelInfo) string {
	if model == nil {
		return ""
	}
	switch model.ProviderType {
	case "openai_compatible":
		return model.ModelCode
	default:
		if strings.HasPrefix(model.ModelCode, "anthropic/") {
			return model.ModelCode
		}
		return "anthropic/" + model.ModelCode
	}
}

func renderCommandTemplate(template string, taskInfo *runtimeTask, engineCfg *runtimeEngineConfig, modelInfo *runtimeModelInfo) string {
	replacements := map[string]string{
		"{{task_id}}":         strconv.FormatInt(taskInfo.ID, 10),
		"{{title}}":           shellQuote(taskInfo.Title),
		"{{instruction}}":     shellQuote(taskInfo.Instruction),
		"{{repo_path}}":       shellQuote(taskInfo.RepoPath),
		"{{worktree_path}}":   shellQuote(taskInfo.WorktreePath),
		"{{branch_name}}":     shellQuote(taskInfo.BranchName),
		"{{engine_code}}":     shellQuote(taskInfo.EngineCode),
		"{{engine_base_url}}": shellQuote(engineCfg.BaseURL),
		"{{engine_api_key}}":  shellQuote(engineCfg.APIKey),
	}
	if modelInfo != nil {
		replacements["{{model_code}}"] = shellQuote(modelInfo.ModelCode)
		replacements["{{model_code_openhands}}"] = shellQuote(formatOpenHandsModel(modelInfo))
		replacements["{{model_base_url}}"] = shellQuote(modelInfo.BaseURL)
		replacements["{{model_base_url_root}}"] = shellQuote(strings.TrimRight(strings.TrimSuffix(modelInfo.BaseURL, "/v1"), "/"))
		replacements["{{model_api_key}}"] = shellQuote(modelInfo.APIKey)
	}
	for key, value := range replacements {
		template = strings.ReplaceAll(template, key, value)
	}
	return template
}

func parseJSONMap(raw string) map[string]interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return map[string]interface{}{}
	}
	result := map[string]interface{}{}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return map[string]interface{}{}
	}
	return result
}

func extractResponseSummary(respText string) string {
	respText = strings.TrimSpace(respText)
	if respText == "" {
		return ""
	}
	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(respText), &payload); err == nil {
		for _, key := range []string{"summary", "message", "result", "output", "content"} {
			if value, ok := payload[key]; ok && strings.TrimSpace(fmt.Sprint(value)) != "" {
				return shortenText(fmt.Sprint(value), 1200)
			}
		}
	}
	return shortenText(respText, 1200)
}

func ensureDirExists(path string) error {
	return ensureDirAvailable(path, "工作目录")
}

func ensureDirAvailable(path string, label string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("%s不可用: %s", label, path)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s不是目录: %s", label, path)
	}
	return nil
}

func psQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "''") + "'"
}

func shQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func shellQuote(value string) string {
	if runtime.GOOS == "windows" {
		return psQuote(value)
	}
	return shQuote(value)
}

func shortenText(text string, max int) string {
	text = strings.ReplaceAll(text, "\r", "")
	text = strings.TrimSpace(text)
	if len(text) <= max {
		return text
	}
	return text[:max] + "...(截断)"
}

func resultOutput(result *taskExecutionResult) string {
	if result == nil {
		return ""
	}
	return result.Output
}

func buildShellCommand(ctx context.Context, command string) *exec.Cmd {
	if runtime.GOOS == "windows" {
		return exec.CommandContext(ctx, "powershell", "-NoProfile", "-Command", command)
	}
	return exec.CommandContext(ctx, "sh", "-lc", command)
}
