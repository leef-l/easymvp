package verification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/engine"
	"easymvp/app/mvp/internal/workflow/event"
	"easymvp/app/mvp/internal/workflow/qualitygate"
	"easymvp/app/mvp/internal/workflow/repo"
)

const (
	statusRunning   = "running"
	statusCompleted = "completed"
	statusFailed    = "failed"

	decisionPassed       = "passed"
	decisionFailed       = "failed"
	decisionManualReview = "manual_review"

	modeAuto          = "auto"
	modeDockerCompose = "docker_compose"
	modeDockerfile    = "dockerfile"
	modeLocal         = "local"
	modeGitHubActions = "github_actions"

	runnerLocal      = "local"
	runnerDockerExec = "docker_exec"
	runnerGitHub     = "github_actions"
)

// StartRequest 启动验证请求。
type StartRequest struct {
	ProjectID     int64
	WorkflowRunID int64
	CreatedBy     int64
	DeptID        int64
	TriggerSource string
	Reason        string
}

// Service 项目验证服务。
type Service struct {
	runRepo        *repo.VerificationRunRepo
	issueRepo      *repo.VerificationIssueRepo
	evidenceRepo   *repo.VerificationEvidenceRepo
	projectRepo    *repo.ProjectRepo
	projectCatRepo *repo.ProjectCategoryRepo
	domainTaskRepo *repo.DomainTaskRepo
}

// NewService 创建验证服务。
func NewService(
	runRepo *repo.VerificationRunRepo,
	issueRepo *repo.VerificationIssueRepo,
	evidenceRepo *repo.VerificationEvidenceRepo,
) *Service {
	if runRepo == nil {
		runRepo = repo.NewVerificationRunRepo()
	}
	if issueRepo == nil {
		issueRepo = repo.NewVerificationIssueRepo()
	}
	if evidenceRepo == nil {
		evidenceRepo = repo.NewVerificationEvidenceRepo()
	}
	return &Service{
		runRepo:        runRepo,
		issueRepo:      issueRepo,
		evidenceRepo:   evidenceRepo,
		projectRepo:    repo.NewProjectRepo(),
		projectCatRepo: repo.NewProjectCategoryRepo(),
		domainTaskRepo: repo.NewDomainTaskRepo(),
	}
}

type runMeta struct {
	RunID           int64
	WorkflowRunID   int64
	ProjectID       int64
	ProjectName     string
	WorkDir         string
	CategoryCode    string
	ProjectCategory string
	CreatedBy       int64
	DeptID          int64
}

type verificationProfile struct {
	Mode          string             `json:"mode"`
	Docker        *dockerProfile     `json:"docker,omitempty"`
	DownAfter     *bool              `json:"downAfter,omitempty"`
	SetupSteps    []verificationStep `json:"setupSteps,omitempty"`
	Steps         []verificationStep `json:"steps,omitempty"`
	TeardownSteps []verificationStep `json:"teardownSteps,omitempty"`
}

type verificationGate struct {
	AllowedDecisions   []string `json:"allowedDecisions,omitempty"`
	MinExecutedSteps   int      `json:"minExecutedSteps,omitempty"`
	RequiredCheckKinds []string `json:"requiredCheckKinds,omitempty"`
	AllowedRunnerTypes []string `json:"allowedRunnerTypes,omitempty"`
}

type categoryVerificationConfig struct {
	CategoryCode string
	DisplayName  string
	Source       string
	Profile      *verificationProfile
	ProfileRaw   string
	Gate         *verificationGate
	GateRaw      string
}

type dockerProfile struct {
	ComposeFile string `json:"composeFile,omitempty"`
	EnvFile     string `json:"envFile,omitempty"`
	ProjectName string `json:"projectName,omitempty"`
	Build       *bool  `json:"build,omitempty"`
}

type verificationStep struct {
	Name           string            `json:"name"`
	Runner         string            `json:"runner,omitempty"`
	Service        string            `json:"service,omitempty"`
	WorkDir        string            `json:"workDir,omitempty"`
	Command        []string          `json:"command"`
	TimeoutSeconds int               `json:"timeoutSeconds,omitempty"`
	Optional       bool              `json:"optional,omitempty"`
	DomainTaskID   int64             `json:"domainTaskID,omitempty"`
	ResourceRef    string            `json:"resourceRef,omitempty"`
	Expected       string            `json:"expected,omitempty"`
	Env            map[string]string `json:"env,omitempty"`
}

type executionPlan struct {
	Mode             string
	RunnerType       string
	ConfigSnapshot   string
	DetectionSummary string
	Profile          *verificationProfile
	Gate             *verificationGate
	GateSource       string
	SetupSteps       []verificationStep
	VerifySteps      []verificationStep
	TeardownSteps    []verificationStep
	CIResult         *ciLatestResult
	CIResultRaw      string
}

type ciLatestResult struct {
	Status     string          `json:"status"`
	Tool       string          `json:"tool,omitempty"`
	Pipeline   string          `json:"pipeline,omitempty"`
	Summary    string          `json:"summary,omitempty"`
	Workflow   string          `json:"workflow,omitempty"`
	RunID      string          `json:"runId,omitempty"`
	RunURL     string          `json:"runUrl,omitempty"`
	CheckKinds []string        `json:"checkKinds,omitempty"`
	Checks     []ciLatestCheck `json:"checks,omitempty"`
}

type ciLatestCheck struct {
	Name     string `json:"name,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Status   string `json:"status,omitempty"`
	Summary  string `json:"summary,omitempty"`
	Command  string `json:"command,omitempty"`
	Runner   string `json:"runner,omitempty"`
	Workflow string `json:"workflow,omitempty"`
	Job      string `json:"job,omitempty"`
}

type issueDraft struct {
	IssueType       string
	Severity        string
	Title           string
	Detail          string
	ExpectedValue   string
	ActualValue     string
	SuggestedAction string
	DomainTaskID    int64
	ResourceRef     string
}

type evidenceDraft struct {
	EvidenceType string
	SourceType   string
	SourceID     int64
	ContentRef   string
	Summary      string
}

type commandResult struct {
	ExitCode int
	Output   string
	Err      error
	Skipped  bool
}

type stepExecution struct {
	Stage  string
	Step   verificationStep
	Result commandResult
}

// Start 创建并异步启动一次验证运行。
func (s *Service) Start(ctx context.Context, req StartRequest) (int64, error) {
	if req.ProjectID == 0 || req.WorkflowRunID == 0 {
		return 0, fmt.Errorf("projectID 和 workflowRunID 不能为空")
	}

	running, err := s.runRepo.CountRunningByWorkflow(ctx, req.WorkflowRunID)
	if err != nil {
		return 0, fmt.Errorf("查询运行中验证失败: %w", err)
	}
	if running > 0 {
		return 0, fmt.Errorf("当前已有运行中的验证，请稍后再试")
	}

	round, err := s.runRepo.GetNextRound(ctx, req.WorkflowRunID)
	if err != nil {
		round = 1
	}
	if strings.TrimSpace(req.TriggerSource) == "" {
		req.TriggerSource = "manual"
	}

	now := gtime.Now()
	runID, err := s.runRepo.Create(ctx, g.Map{
		"workflow_run_id":    req.WorkflowRunID,
		"project_id":         req.ProjectID,
		"verification_round": round,
		"status":             statusRunning,
		"trigger_source":     req.TriggerSource,
		"created_by":         req.CreatedBy,
		"dept_id":            req.DeptID,
		"started_at":         now,
		"created_at":         now,
		"updated_at":         now,
	})
	if err != nil {
		return 0, fmt.Errorf("创建验证运行失败: %w", err)
	}

	s.insertWorkflowEvent(ctx, req.WorkflowRunID, "verification_run", "verification.started", &runID, map[string]interface{}{
		"project_id":      req.ProjectID,
		"verification_id": runID,
		"trigger_source":  req.TriggerSource,
		"reason":          strings.TrimSpace(req.Reason),
	})

	s.runAsync(runID)
	return runID, nil
}

func (s *Service) runAsync(runID int64) {
	go func() {
		bgCtx := context.Background()
		defer func() {
			if r := recover(); r != nil {
				g.Log().Errorf(bgCtx, "[Verification] panic: runID=%d panic=%v", runID, r)
				s.failRun(bgCtx, runID, fmt.Sprintf("验证执行 panic: %v", r))
			}
		}()

		if err := s.executeRun(bgCtx, runID); err != nil {
			g.Log().Errorf(bgCtx, "[Verification] 执行失败: runID=%d err=%v", runID, err)
			s.failRun(bgCtx, runID, err.Error())
		}
	}()
}

func (s *Service) executeRun(ctx context.Context, runID int64) error {
	meta, err := s.loadRunMeta(ctx, runID)
	if err != nil {
		return err
	}

	plan, issues, evidence, err := s.buildExecutionPlan(ctx, meta)
	if err != nil {
		return err
	}

	executedSteps := 0
	stepExecutions := make([]stepExecution, 0, len(plan.SetupSteps)+len(plan.VerifySteps))
	if plan.RunnerType == modeGitHubActions {
		executedSteps, stepExecutions, issues, evidence = s.collectGitHubActionsEvidence(meta, plan, issues, evidence)
	} else {
		setupFailed := false
		for _, step := range plan.SetupSteps {
			result, commandText := s.executeStep(ctx, meta.WorkDir, plan.Profile, step)
			evidence = append(evidence, s.buildStepEvidence(meta, "command", step, commandText, result))
			stepExecutions = append(stepExecutions, stepExecution{Stage: "setup", Step: step, Result: result})
			if !result.Skipped {
				executedSteps++
			}
			if result.Err != nil && !step.Optional {
				setupFailed = true
				issues = append(issues, s.buildStepIssue(step, "environment", "blocker", result, commandText))
			} else if result.Err != nil {
				issues = append(issues, s.buildStepIssue(step, "environment", "warn", result, commandText))
			}
		}

		for _, step := range plan.VerifySteps {
			if setupFailed && step.Runner == runnerDockerExec {
				skipped := commandResult{Skipped: true}
				commandText := describeStepCommand(step)
				evidence = append(evidence, s.buildStepEvidence(meta, "command", step, commandText, skipped))
				issues = append(issues, issueDraft{
					IssueType:       "environment",
					Severity:        "warn",
					Title:           fmt.Sprintf("跳过验证步骤: %s", step.Name),
					Detail:          "检测到 legacy Docker runner 启动失败，兼容验证步骤已被跳过。",
					ExpectedValue:   "环境启动成功并继续执行容器内验证步骤",
					ActualValue:     "环境启动失败，步骤被跳过",
					SuggestedAction: "移除 legacy Docker 验证配置，改由 GitHub Actions workflow 执行对应检查",
					DomainTaskID:    step.DomainTaskID,
					ResourceRef:     step.ResourceRef,
				})
				continue
			}

			result, commandText := s.executeStep(ctx, meta.WorkDir, plan.Profile, step)
			evidence = append(evidence, s.buildStepEvidence(meta, "command", step, commandText, result))
			stepExecutions = append(stepExecutions, stepExecution{Stage: "verify", Step: step, Result: result})
			if !result.Skipped {
				executedSteps++
			}
			if result.Err != nil && !step.Optional {
				issues = append(issues, s.buildStepIssue(step, "command", "error", result, commandText))
			} else if result.Err != nil {
				issues = append(issues, s.buildStepIssue(step, "command", "warn", result, commandText))
			}
		}

		for _, step := range plan.TeardownSteps {
			result, commandText := s.executeStep(ctx, meta.WorkDir, plan.Profile, step)
			evidence = append(evidence, s.buildStepEvidence(meta, "command", step, commandText, result))
			if result.Err != nil {
				issues = append(issues, s.buildStepIssue(step, "runtime", "warn", result, commandText))
			}
		}
	}

	issues = append(issues, evaluateVerificationGate(plan, executedSteps, stepExecutions, issues)...)

	if executedSteps == 0 {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "未执行任何验证步骤",
			Detail:          "系统没有读取到已完成的 GitHub Actions 检查结果，当前结果不足以判断项目是否可发布。",
			ExpectedValue:   "至少执行 1 个有效验证步骤",
			ActualValue:     "执行步骤数为 0",
			SuggestedAction: "等待 GitHub Actions 完成并同步最新 CI 结果后重新发起验证",
		})
	}

	if len(evidence) > 0 {
		items := make([]g.Map, 0, len(evidence))
		for _, item := range evidence {
			items = append(items, s.newEvidenceMap(meta, item))
		}
		if err := s.evidenceRepo.BatchCreate(ctx, items); err != nil {
			return fmt.Errorf("写入验证证据失败: %w", err)
		}
	}

	if len(issues) > 0 {
		items := make([]g.Map, 0, len(issues))
		for _, item := range issues {
			items = append(items, s.newIssueMap(ctx, meta, item))
		}
		if err := s.issueRepo.BatchCreate(ctx, items); err != nil {
			return fmt.Errorf("写入验证问题失败: %w", err)
		}
	}

	decision := decideRunResult(issues, executedSteps)
	summary := buildRunSummary(plan.RunnerType, executedSteps, issues)
	extra := g.Map{
		"decision":            decision,
		"runner_type":         plan.RunnerType,
		"summary":             summary,
		"config_snapshot_ref": plan.ConfigSnapshot,
		"finished_at":         gtime.Now(),
	}
	rows, err := s.runRepo.UpdateStatus(ctx, runID, statusRunning, statusCompleted, extra)
	if err != nil {
		return fmt.Errorf("更新验证状态失败: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("验证运行 %d 已不在 running 状态", runID)
	}

	eventType := "verification.completed"
	if decision == decisionFailed {
		eventType = "verification.failed"
	}
	s.insertWorkflowEvent(ctx, meta.WorkflowRunID, "verification_run", eventType, &runID, map[string]interface{}{
		"verification_id": runID,
		"decision":        decision,
		"runner_type":     plan.RunnerType,
		"summary":         summary,
	})
	return nil
}

func (s *Service) loadRunMeta(ctx context.Context, runID int64) (*runMeta, error) {
	run, err := s.runRepo.GetByID(ctx, runID)
	if err != nil || len(run) == 0 {
		return nil, fmt.Errorf("verification_run(%d) 不存在", runID)
	}

	project, err := s.projectRepo.GetByID(ctx, g.NewVar(run["project_id"]).Int64(),
		"id", "name", "work_dir", "category_code", "project_category", "created_by", "dept_id")
	if err != nil || len(project) == 0 {
		return nil, fmt.Errorf("project(%d) 不存在", g.NewVar(run["project_id"]).Int64())
	}

	return &runMeta{
		RunID:           runID,
		WorkflowRunID:   g.NewVar(run["workflow_run_id"]).Int64(),
		ProjectID:       g.NewVar(project["id"]).Int64(),
		ProjectName:     g.NewVar(project["name"]).String(),
		WorkDir:         strings.TrimSpace(g.NewVar(project["work_dir"]).String()),
		CategoryCode:    strings.TrimSpace(g.NewVar(project["category_code"]).String()),
		ProjectCategory: strings.TrimSpace(g.NewVar(project["project_category"]).String()),
		CreatedBy:       g.NewVar(project["created_by"]).Int64(),
		DeptID:          g.NewVar(project["dept_id"]).Int64(),
	}, nil
}

func (s *Service) buildExecutionPlan(ctx context.Context, meta *runMeta) (*executionPlan, []issueDraft, []evidenceDraft, error) {
	plan := &executionPlan{Mode: modeGitHubActions, RunnerType: modeGitHubActions}
	issues := make([]issueDraft, 0, 4)
	evidence := make([]evidenceDraft, 0, 4)

	if meta.WorkDir == "" {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "blocker",
			Title:           "项目未配置工作目录",
			Detail:          "当前项目 work_dir 为空，无法执行自动验证。",
			ExpectedValue:   "项目配置有效的代码工作目录",
			ActualValue:     "work_dir 为空",
			SuggestedAction: "为项目补充仓库工作目录后重新发起验证",
		})
		plan.ConfigSnapshot = `{"mode":"github_actions","reason":"missing_workdir"}`
		plan.DetectionSummary = "项目未配置工作目录"
		return plan, issues, evidence, nil
	}
	if _, err := os.Stat(meta.WorkDir); err != nil {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "blocker",
			Title:           "项目工作目录不存在",
			Detail:          fmt.Sprintf("工作目录 %s 不存在或不可访问。", meta.WorkDir),
			ExpectedValue:   "工作目录存在且可读写",
			ActualValue:     err.Error(),
			SuggestedAction: "修复工作目录映射或重新绑定仓库路径",
		})
		plan.ConfigSnapshot = fmt.Sprintf(`{"mode":"github_actions","workDir":%q}`, meta.WorkDir)
		plan.DetectionSummary = "工作目录不存在"
		return plan, issues, evidence, nil
	}

	projectProfile, configPath, rawConfig, configErr := loadVerificationProfile(meta.WorkDir)
	if configErr != nil {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "error",
			Title:           "验证配置解析失败",
			Detail:          fmt.Sprintf("%s 解析失败: %v", configPath, configErr),
			ExpectedValue:   "verification.json 为合法 JSON",
			ActualValue:     configErr.Error(),
			SuggestedAction: "修复 .easymvp/verification.json 后重新验证；当前先回退自动检测",
		})
	}

	categoryConfig, categoryIssues, err := s.loadCategoryVerificationConfig(ctx, meta)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("加载分类验证配置失败: %w", err)
	}
	issues = append(issues, categoryIssues...)

	profile := projectProfile
	profileSource := "auto_detect"
	profileRaw := ""
	if projectProfile != nil {
		profileSource = "project:.easymvp/verification.json"
		profileRaw = rawConfig
	} else if categoryConfig != nil && categoryConfig.Profile != nil {
		profile = categoryConfig.Profile
		profileSource = categoryConfig.Source
		profileRaw = categoryConfig.ProfileRaw
	}
	plan.Profile = profile
	if categoryConfig != nil && categoryConfig.Gate != nil {
		plan.Gate = categoryConfig.Gate
		plan.GateSource = categoryConfig.Source
	}

	requestedMode := modeGitHubActions
	if profile != nil && strings.TrimSpace(profile.Mode) != "" {
		requestedMode = strings.TrimSpace(strings.ToLower(profile.Mode))
	}
	plan.Mode = normalizeMode(requestedMode)
	plan.RunnerType = plan.Mode

	workflowFiles, _ := findGitHubWorkflowFiles(meta.WorkDir)
	latestCIResult, latestCIPath, latestCIRaw, latestCIErr := loadLatestCIResult(meta.WorkDir)
	plan.CIResult = latestCIResult
	plan.CIResultRaw = latestCIRaw

	detection := map[string]interface{}{
		"mode":          plan.Mode,
		"workDir":       meta.WorkDir,
		"configPath":    configPath,
		"profileSource": profileSource,
		"gateSource":    valueOrDefault(plan.GateSource, "none"),
		"projectCategory": map[string]string{
			"code":        meta.CategoryCode,
			"displayName": meta.ProjectCategory,
		},
	}
	if categoryConfig != nil {
		detection["categoryConfig"] = map[string]string{
			"code":        categoryConfig.CategoryCode,
			"displayName": categoryConfig.DisplayName,
			"source":      categoryConfig.Source,
		}
	}
	detection["githubActionsWorkflowCount"] = len(workflowFiles)
	if len(workflowFiles) > 0 {
		detection["githubActionsWorkflows"] = workflowFiles
	}
	if latestCIPath != "" {
		detection["ciLatestPath"] = latestCIPath
	}
	if profileRaw != "" {
		detection["profile"] = snapshotJSONValue(profileRaw)
	}
	if latestCIRaw != "" {
		detection["ciLatest"] = snapshotJSONValue(latestCIRaw)
	}
	if categoryConfig != nil && categoryConfig.GateRaw != "" {
		detection["gate"] = snapshotJSONValue(categoryConfig.GateRaw)
	}
	signals := qualitygate.DetectProjectSignals(
		meta.WorkDir,
		string(engine.GetCategoryFamily(valueOrDefault(meta.CategoryCode, meta.ProjectCategory))),
		valueOrDefault(meta.CategoryCode, meta.ProjectCategory),
	)
	standard := qualitygate.ResolveVerificationStandard(signals)
	detection["verificationStandard"] = map[string]interface{}{
		"code":                 standard.Code,
		"displayName":          standard.DisplayName,
		"requiredCheckKinds":   standard.RequiredCheckKinds,
		"requiredProjectRoles": standard.RequiredProjectRoles,
		"signals": map[string]interface{}{
			"familyCode":           signals.FamilyCode,
			"projectTypeCode":      signals.ProjectTypeCode,
			"hasGoModules":         signals.HasGoModules,
			"hasNodePackage":       signals.HasNodePackage,
			"hasFrontendApp":       signals.HasFrontendApp,
			"hasBrowserAutomation": signals.HasBrowserAutomation,
			"hasAndroidApp":        signals.HasAndroidApp,
			"hasIOSApp":            signals.HasIOSApp,
			"reasons":              signals.Reasons,
		},
	}
	if profileRequestsLocalExecution(profile) {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "检测到已停用的本机验证配置",
			Detail:          "当前 verification profile 仍包含本机 steps / docker / downAfter / legacy mode 配置，但工程铁律已要求测试与编译统一交给 GitHub Actions。",
			ExpectedValue:   `{"mode":"github_actions"}`,
			ActualValue:     valueOrDefault(strings.TrimSpace(profileRaw), "检测到 legacy verification profile"),
			SuggestedAction: "移除本机验证命令与 Docker 回退配置，改由 GitHub Actions 执行后回写 .easymvp/ci/latest.json",
		})
	}
	if len(workflowFiles) == 0 {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "error",
			Title:           "未检测到 GitHub Actions 工作流",
			Detail:          "项目目录下没有发现 .github/workflows/*.yml 或 *.yaml，无法按当前铁律完成测试与编译验证。",
			ExpectedValue:   "项目仓库包含 GitHub Actions workflow",
			ActualValue:     "workflow 文件缺失",
			SuggestedAction: "补齐 GitHub Actions workflow，并通过 workflow run 产出 CI 结果",
		})
	}
	if latestCIErr != nil {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "error",
			Title:           "CI 结果文件解析失败",
			Detail:          fmt.Sprintf("%s 解析失败: %v", latestCIPath, latestCIErr),
			ExpectedValue:   ".easymvp/ci/latest.json 为合法 JSON",
			ActualValue:     latestCIErr.Error(),
			SuggestedAction: "修复 latest.json 内容，或重新由 GitHub Actions 生成最新结果文件",
		})
	}
	if latestCIResult == nil {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "未找到 GitHub Actions 最新结果",
			Detail:          "当前项目尚未提供 .easymvp/ci/latest.json，验证阶段无法读取 GitHub Actions 的最新测试/编译结果。",
			ExpectedValue:   ".easymvp/ci/latest.json 包含最新 workflow run 结果",
			ActualValue:     "CI 结果文件缺失",
			SuggestedAction: "先触发对应的 GitHub Actions guard workflow，并把最新结果同步到 .easymvp/ci/latest.json",
		})
	} else {
		plan.VerifySteps = buildGitHubActionsSteps(latestCIResult)
		if len(plan.VerifySteps) == 0 {
			issues = append(issues, issueDraft{
				IssueType:       "config",
				Severity:        "warn",
				Title:           "CI 结果未声明检查项",
				Detail:          "latest.json 已存在，但没有 checks 或 checkKinds，系统无法把 GitHub Actions 结果映射到 test/build/browser 等标准检查类型。",
				ExpectedValue:   "latest.json 至少包含 checks 或 checkKinds",
				ActualValue:     "检查项为空",
				SuggestedAction: "在 CI 结果中补充 checks/checkKinds 字段，明确每个 GitHub Actions 检查的类型与状态",
			})
		}
		if tool := strings.TrimSpace(strings.ToLower(latestCIResult.Tool)); tool != "" && tool != runnerGitHub {
			issues = append(issues, issueDraft{
				IssueType:       "config",
				Severity:        "error",
				Title:           "CI 结果来源不符合铁律",
				Detail:          fmt.Sprintf("latest.json 声明的 tool=%s，不是 github_actions。", latestCIResult.Tool),
				ExpectedValue:   "tool=github_actions",
				ActualValue:     latestCIResult.Tool,
				SuggestedAction: "仅接受 GitHub Actions 作为测试与编译来源，请修正 CI 结果输出",
			})
		}
	}
	plan.Gate = mergeVerificationStandardIntoGate(plan.Gate, standard)
	plan.GateSource = mergeGateSource(plan.GateSource, standard.Code)

	if len(plan.VerifySteps) == 0 {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "未检测到可用的 GitHub Actions 验证结果",
			Detail:          "系统没有读取到可映射为标准检查项的 GitHub Actions 结果，当前结果不足以判断项目是否可发布。",
			ExpectedValue:   "至少存在 1 个 GitHub Actions 检查结果",
			ActualValue:     "检测结果为空",
			SuggestedAction: "通过 GitHub Actions 执行验证，并在 .easymvp/ci/latest.json 中声明 checks/checkKinds",
		})
	}

	if plan.ConfigSnapshot == "" {
		encoded, _ := json.Marshal(detection)
		plan.ConfigSnapshot = string(encoded)
	}
	plan.DetectionSummary = fmt.Sprintf(
		"mode=%s profileSource=%s gateSource=%s workflows=%d ci=%t steps=%d",
		plan.RunnerType,
		profileSource,
		valueOrDefault(plan.GateSource, "none"),
		len(workflowFiles),
		latestCIResult != nil,
		len(plan.VerifySteps),
	)
	evidence = append(evidence, evidenceDraft{
		EvidenceType: "profile",
		SourceType:   "verification_run",
		SourceID:     meta.RunID,
		ContentRef:   plan.ConfigSnapshot,
		Summary:      plan.DetectionSummary,
	})
	return plan, issues, evidence, nil
}

func (s *Service) collectGitHubActionsEvidence(
	meta *runMeta,
	plan *executionPlan,
	issues []issueDraft,
	evidence []evidenceDraft,
) (int, []stepExecution, []issueDraft, []evidenceDraft) {
	stepExecutions := make([]stepExecution, 0, len(plan.VerifySteps))
	if plan == nil || plan.CIResult == nil {
		return 0, stepExecutions, issues, evidence
	}

	result := plan.CIResult
	if raw := strings.TrimSpace(plan.CIResultRaw); raw != "" {
		evidence = append(evidence, evidenceDraft{
			EvidenceType: "ci",
			SourceType:   "verification_run",
			SourceID:     meta.RunID,
			ContentRef:   raw,
			Summary:      buildCISnapshotSummary(result),
		})
	}

	executedSteps := 0
	for idx, step := range plan.VerifySteps {
		checkStatus := normalizeCIStatus(result.Status)
		checkSummary := strings.TrimSpace(result.Summary)
		if idx < len(result.Checks) {
			check := result.Checks[idx]
			if normalized := normalizeCIStatus(check.Status); normalized != "" {
				checkStatus = normalized
			}
			if summary := strings.TrimSpace(check.Summary); summary != "" {
				checkSummary = summary
			}
		}

		commandText := describeStepCommand(step)
		commandResult := commandResult{
			ExitCode: 0,
			Output:   checkSummary,
		}

		switch checkStatus {
		case "passed":
			executedSteps++
		case "failed":
			executedSteps++
			commandResult.ExitCode = 1
			commandResult.Err = fmt.Errorf("GitHub Actions status=%s", checkStatus)
		default:
			commandResult.Skipped = true
			commandResult.Output = valueOrDefault(checkSummary, "GitHub Actions 检查尚未完成")
			issues = append(issues, issueDraft{
				IssueType:       "ci",
				Severity:        "warn",
				Title:           fmt.Sprintf("GitHub Actions 检查未完成: %s", step.Name),
				Detail:          fmt.Sprintf("当前检查状态为 %s，验证阶段尚未拿到可判定的最终结果。", valueOrDefault(checkStatus, "pending")),
				ExpectedValue:   "GitHub Actions check 已完成且通过",
				ActualValue:     valueOrDefault(checkStatus, "pending"),
				SuggestedAction: "等待 workflow run 完成并回写 .easymvp/ci/latest.json 后重新发起验证",
				DomainTaskID:    step.DomainTaskID,
				ResourceRef:     step.ResourceRef,
			})
		}

		evidence = append(evidence, s.buildStepEvidence(meta, "command", step, commandText, commandResult))
		stepExecutions = append(stepExecutions, stepExecution{Stage: "verify", Step: step, Result: commandResult})
		if commandResult.Err != nil {
			issues = append(issues, s.buildStepIssue(step, "ci", "error", commandResult, commandText))
		}
	}

	return executedSteps, stepExecutions, issues, evidence
}

func buildGitHubActionsSteps(result *ciLatestResult) []verificationStep {
	if result == nil {
		return nil
	}

	steps := make([]verificationStep, 0, len(result.Checks))
	for _, check := range result.Checks {
		name := strings.TrimSpace(check.Name)
		kind := normalizeCheckKind(check.Kind)
		if name == "" {
			if kind != "" {
				name = "github actions " + kind
			} else {
				name = "github actions check"
			}
		}
		steps = append(steps, verificationStep{
			Name:           name,
			Runner:         runnerGitHub,
			Command:        buildGitHubActionsCommand(check, kind),
			WorkDir:        ".",
			ResourceRef:    ".",
			TimeoutSeconds: 0,
			Expected:       "GitHub Actions 检查通过",
		})
	}

	if len(steps) > 0 {
		return steps
	}

	label := strings.TrimSpace(result.Pipeline)
	if label == "" {
		label = "github actions pipeline"
	}
	return []verificationStep{{
		Name:        label,
		Runner:      runnerGitHub,
		Command:     []string{"github_actions", "pipeline", normalizeCIStatus(result.Status)},
		WorkDir:     ".",
		ResourceRef: ".",
		Expected:    "GitHub Actions workflow 通过",
	}}
}

func buildGitHubActionsCommand(check ciLatestCheck, kind string) []string {
	command := make([]string, 0, 5)
	command = append(command, "github_actions", "check")
	switch {
	case kind != "":
		command = append(command, kind)
	case strings.TrimSpace(check.Command) != "":
		command = append(command, strings.Fields(check.Command)...)
	default:
		command = append(command, "unknown")
	}
	if status := normalizeCIStatus(check.Status); status != "" {
		command = append(command, "status="+status)
	}
	return command
}

func buildCISnapshotSummary(result *ciLatestResult) string {
	if result == nil {
		return "GitHub Actions CI 结果"
	}
	parts := make([]string, 0, 4)
	if status := normalizeCIStatus(result.Status); status != "" {
		parts = append(parts, "status="+status)
	}
	if tool := strings.TrimSpace(result.Tool); tool != "" {
		parts = append(parts, "tool="+tool)
	}
	if pipeline := strings.TrimSpace(result.Pipeline); pipeline != "" {
		parts = append(parts, "pipeline="+pipeline)
	}
	if summary := strings.TrimSpace(result.Summary); summary != "" {
		parts = append(parts, "summary="+summary)
	}
	if len(parts) == 0 {
		return "GitHub Actions CI 结果"
	}
	return "CI 结果：" + strings.Join(parts, " ")
}

func loadLatestCIResult(root string) (*ciLatestResult, string, string, error) {
	path := filepath.Join(root, ".easymvp", "ci", "latest.json")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, path, "", nil
		}
		return nil, path, "", err
	}

	var result ciLatestResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, path, string(data), err
	}
	result.Status = normalizeCIStatus(result.Status)
	result.Tool = strings.TrimSpace(strings.ToLower(result.Tool))
	result.Pipeline = strings.TrimSpace(result.Pipeline)
	result.Summary = strings.TrimSpace(result.Summary)
	result.Workflow = strings.TrimSpace(result.Workflow)
	result.RunID = strings.TrimSpace(result.RunID)
	result.RunURL = strings.TrimSpace(result.RunURL)

	result.CheckKinds = normalizeCIResultCheckKinds(result.CheckKinds)
	result.Checks = normalizeCIResultChecks(result.Checks, result.CheckKinds)
	return &result, path, string(data), nil
}

func normalizeCIResultCheckKinds(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		kind := normalizeCheckKind(value)
		if kind == "" {
			continue
		}
		if _, ok := seen[kind]; ok {
			continue
		}
		seen[kind] = struct{}{}
		result = append(result, kind)
	}
	sort.Strings(result)
	return result
}

func normalizeCIResultChecks(checks []ciLatestCheck, fallbackKinds []string) []ciLatestCheck {
	result := make([]ciLatestCheck, 0, len(checks)+len(fallbackKinds))
	for _, check := range checks {
		check.Name = strings.TrimSpace(check.Name)
		check.Kind = normalizeCheckKind(check.Kind)
		check.Status = normalizeCIStatus(check.Status)
		check.Summary = strings.TrimSpace(check.Summary)
		check.Command = strings.TrimSpace(check.Command)
		check.Runner = runnerGitHub
		check.Workflow = strings.TrimSpace(check.Workflow)
		check.Job = strings.TrimSpace(check.Job)
		result = append(result, check)
	}
	if len(result) == 0 {
		for _, kind := range fallbackKinds {
			result = append(result, ciLatestCheck{
				Name:   "github actions " + kind,
				Kind:   kind,
				Runner: runnerGitHub,
			})
		}
	}
	return result
}

func normalizeCIStatus(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "success", "succeeded", "passed", "pass", "completed":
		return "passed"
	case "failure", "failed", "error", "cancelled", "canceled", "timed_out":
		return "failed"
	case "", "queued", "queue", "running", "in_progress", "pending", "requested", "waiting":
		return "pending"
	default:
		return strings.TrimSpace(strings.ToLower(value))
	}
}

func profileRequestsLocalExecution(profile *verificationProfile) bool {
	if profile == nil {
		return false
	}
	if len(profile.SetupSteps) > 0 || len(profile.Steps) > 0 || len(profile.TeardownSteps) > 0 {
		return true
	}
	if profile.Docker != nil || profile.DownAfter != nil {
		return true
	}
	switch strings.TrimSpace(strings.ToLower(profile.Mode)) {
	case "", modeAuto, modeGitHubActions:
		return false
	default:
		return true
	}
}

func findGitHubWorkflowFiles(root string) ([]string, error) {
	workflowDir := filepath.Join(root, ".github", "workflows")
	entries, err := os.ReadDir(workflowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	files := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		if strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml") {
			files = append(files, filepath.ToSlash(filepath.Join(".github", "workflows", entry.Name())))
		}
	}
	sort.Strings(files)
	return files, nil
}

func snapshotJSONValue(raw string) interface{} {
	text := strings.TrimSpace(raw)
	if text == "" {
		return ""
	}
	if json.Valid([]byte(text)) {
		return json.RawMessage(text)
	}
	return text
}

func (s *Service) discoverLocalSetupSteps(ctx context.Context, meta *runMeta, root string) []verificationStep {
	steps := make([]verificationStep, 0, 4)
	seen := make(map[string]struct{})

	for _, dir := range findProjectSubdirs(root, "package.json") {
		scripts, err := readPackageScripts(dir)
		if err != nil || !hasAutoDiscoverablePackageScripts(scripts) {
			continue
		}
		if info, err := os.Stat(filepath.Join(dir, "node_modules")); err == nil && info.IsDir() {
			continue
		}

		rel := relativePath(root, dir)
		key := "install:" + rel
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}

		pm := detectPackageManager(dir)
		command := buildPackageInstallCommand(dir, pm)
		resourceRef := strings.TrimSpace(rel)
		steps = append(steps, verificationStep{
			Name:           displayStepName(rel, describePackageInstallCommand(command)),
			Runner:         runnerLocal,
			WorkDir:        rel,
			Command:        command,
			TimeoutSeconds: 900,
			ResourceRef:    resourceRef,
			DomainTaskID:   s.findRelatedDomainTaskID(ctx, meta.WorkflowRunID, resourceRef),
			Expected:       "项目依赖安装完成",
		})
	}

	return steps
}

func (s *Service) discoverLocalSteps(ctx context.Context, meta *runMeta, root string) []verificationStep {
	steps := make([]verificationStep, 0, 8)
	seen := make(map[string]struct{})

	for _, dir := range findProjectSubdirs(root, "go.mod") {
		rel := relativePath(root, dir)
		key := "go:" + rel
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		resourceRef := strings.TrimSpace(rel)
		steps = append(steps, verificationStep{
			Name:           displayStepName(rel, "go test ./..."),
			Runner:         runnerLocal,
			WorkDir:        rel,
			Command:        []string{"go", "test", "./..."},
			TimeoutSeconds: 600,
			ResourceRef:    resourceRef,
			DomainTaskID:   s.findRelatedDomainTaskID(ctx, meta.WorkflowRunID, resourceRef),
			Expected:       "Go 单元测试全部通过",
		})
	}

	for _, dir := range findProjectSubdirs(root, "package.json") {
		rel := relativePath(root, dir)
		scripts, err := readPackageScripts(dir)
		if err != nil || len(scripts) == 0 {
			continue
		}
		pm := detectPackageManager(dir)
		resourceRef := strings.TrimSpace(rel)
		taskID := s.findRelatedDomainTaskID(ctx, meta.WorkflowRunID, resourceRef)

		if _, ok := scripts["lint"]; ok {
			key := "lint:" + rel
			if _, seenLint := seen[key]; !seenLint {
				seen[key] = struct{}{}
				steps = append(steps, verificationStep{
					Name:           displayStepName(rel, "lint"),
					Runner:         runnerLocal,
					WorkDir:        rel,
					Command:        buildPackageScriptCommand(pm, "lint"),
					TimeoutSeconds: 300,
					ResourceRef:    resourceRef,
					DomainTaskID:   taskID,
					Expected:       "前端/Node lint 通过",
				})
			}
		}
		if script, ok := scripts["test"]; ok && !qualitygate.IsBrowserScript("test", script) {
			key := "test:" + rel
			if _, seenTest := seen[key]; !seenTest {
				seen[key] = struct{}{}
				step := verificationStep{
					Name:           displayStepName(rel, "test"),
					Runner:         runnerLocal,
					WorkDir:        rel,
					Command:        buildTestScriptCommand(pm, script),
					TimeoutSeconds: 600,
					ResourceRef:    resourceRef,
					DomainTaskID:   taskID,
					Expected:       "自动测试通过",
					Env:            map[string]string{"CI": "true"},
				}
				steps = append(steps, step)
			}
		}
		for _, scriptName := range discoverBrowserScriptNames(scripts) {
			key := "browser:" + rel + ":" + scriptName
			if _, seenBrowser := seen[key]; seenBrowser {
				continue
			}
			seen[key] = struct{}{}
			steps = append(steps, verificationStep{
				Name:           displayStepName(rel, "browser "+scriptName),
				Runner:         runnerLocal,
				WorkDir:        rel,
				Command:        buildPackageScriptCommand(pm, scriptName),
				TimeoutSeconds: 900,
				ResourceRef:    resourceRef,
				DomainTaskID:   taskID,
				Expected:       "浏览器级关键交互验证通过",
				Env:            map[string]string{"CI": "true"},
			})
		}
		if _, ok := scripts["build"]; ok {
			key := "build:" + rel
			if _, seenBuild := seen[key]; !seenBuild {
				seen[key] = struct{}{}
				steps = append(steps, verificationStep{
					Name:           displayStepName(rel, "build"),
					Runner:         runnerLocal,
					WorkDir:        rel,
					Command:        buildPackageScriptCommand(pm, "build"),
					TimeoutSeconds: 600,
					ResourceRef:    resourceRef,
					DomainTaskID:   taskID,
					Expected:       "项目构建通过",
				})
			}
		}
	}

	return steps
}

func (s *Service) normalizeSteps(ctx context.Context, meta *runMeta, steps []verificationStep) []verificationStep {
	result := make([]verificationStep, 0, len(steps))
	for _, step := range steps {
		step.Runner = normalizeRunner(step.Runner)
		if step.Runner == "" {
			step.Runner = runnerLocal
		}
		if step.Name == "" {
			step.Name = describeStepCommand(step)
		}
		if step.ResourceRef == "" {
			step.ResourceRef = strings.TrimSpace(step.WorkDir)
		}
		if step.DomainTaskID == 0 {
			step.DomainTaskID = s.findRelatedDomainTaskID(ctx, meta.WorkflowRunID, step.ResourceRef)
		}
		result = append(result, step)
	}
	return result
}

func (s *Service) executeStep(ctx context.Context, root string, profile *verificationProfile, step verificationStep) (commandResult, string) {
	timeout := time.Duration(step.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 5 * time.Minute
	}
	stepCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	step = applyCommandResourcePolicy(stepCtx, step)

	args, dir, commandText, err := buildCommand(root, profile, step)
	if err != nil {
		return commandResult{ExitCode: -1, Err: err}, commandText
	}
	if len(args) == 0 {
		return commandResult{ExitCode: -1, Err: fmt.Errorf("命令为空")}, commandText
	}

	cmd := exec.CommandContext(stepCtx, args[0], args[1:]...)
	if dir != "" {
		cmd.Dir = dir
	}
	cmd.Env = append(os.Environ(), buildEnvPairs(step.Env)...)
	engine.GetCommandResourcePolicy(stepCtx).Apply(cmd)
	output, execErr := cmd.CombinedOutput()
	result := commandResult{
		ExitCode: 0,
		Output:   truncateText(string(output), 16000),
		Err:      execErr,
	}
	if execErr != nil {
		var exitErr *exec.ExitError
		if errors.As(execErr, &exitErr) {
			result.ExitCode = exitErr.ExitCode()
		} else if errors.Is(stepCtx.Err(), context.DeadlineExceeded) {
			result.ExitCode = -1
		} else {
			result.ExitCode = -1
		}
	}
	return result, commandText
}

func applyCommandResourcePolicy(ctx context.Context, step verificationStep) verificationStep {
	policy := engine.GetCommandResourcePolicy(ctx)
	step.Env = policy.MergeEnv(step.Env)
	return step
}

func (s *Service) buildStepIssue(step verificationStep, issueType, severity string, result commandResult, commandText string) issueDraft {
	actual := fmt.Sprintf("command=%s exitCode=%d", commandText, result.ExitCode)
	if result.Err != nil {
		actual = result.Err.Error()
	}
	detail := fmt.Sprintf("命令执行失败：%s", commandText)
	if trimmed := strings.TrimSpace(result.Output); trimmed != "" {
		detail += "\n\n输出：\n" + trimmed
	}
	suggestedAction := "修复步骤对应的代码、配置或 CI 结果后重新发起验证"
	if step.Runner == runnerGitHub {
		suggestedAction = "修复 GitHub Actions 失败项并重新执行 workflow，待最新结果回写后重新发起验证"
	}
	return issueDraft{
		IssueType:       issueType,
		Severity:        severity,
		Title:           fmt.Sprintf("验证步骤失败: %s", step.Name),
		Detail:          detail,
		ExpectedValue:   valueOrDefault(step.Expected, "命令执行成功"),
		ActualValue:     actual,
		SuggestedAction: suggestedAction,
		DomainTaskID:    step.DomainTaskID,
		ResourceRef:     step.ResourceRef,
	}
}

func (s *Service) buildStepEvidence(meta *runMeta, evidenceType string, step verificationStep, commandText string, result commandResult) evidenceDraft {
	payload := map[string]interface{}{
		"name":      step.Name,
		"runner":    step.Runner,
		"workDir":   step.WorkDir,
		"command":   commandText,
		"exitCode":  result.ExitCode,
		"skipped":   result.Skipped,
		"output":    result.Output,
		"projectID": meta.ProjectID,
	}
	content, _ := json.Marshal(payload)
	summary := fmt.Sprintf("%s: exit=%d", step.Name, result.ExitCode)
	if result.Skipped {
		summary = step.Name + ": skipped"
	}
	return evidenceDraft{
		EvidenceType: evidenceType,
		SourceType:   "verification_step",
		SourceID:     step.DomainTaskID,
		ContentRef:   string(content),
		Summary:      summary,
	}
}

func (s *Service) newIssueMap(ctx context.Context, meta *runMeta, item issueDraft) g.Map {
	now := gtime.Now()
	resolvedTaskID := item.DomainTaskID
	if taskID, err := ResolveIssueTaskID(ctx, meta.WorkflowRunID, item.DomainTaskID, item.Title, item.Detail, item.ResourceRef); err == nil && taskID > 0 {
		resolvedTaskID = taskID
	}
	return g.Map{
		"verification_run_id": meta.RunID,
		"workflow_run_id":     meta.WorkflowRunID,
		"project_id":          meta.ProjectID,
		"domain_task_id":      nullableInt64(resolvedTaskID),
		"issue_type":          item.IssueType,
		"severity":            item.Severity,
		"title":               item.Title,
		"detail":              item.Detail,
		"expected_value":      item.ExpectedValue,
		"actual_value":        item.ActualValue,
		"suggested_action":    item.SuggestedAction,
		"resource_ref":        item.ResourceRef,
		"status":              "open",
		"created_by":          meta.CreatedBy,
		"dept_id":             meta.DeptID,
		"created_at":          now,
		"updated_at":          now,
	}
}

func (s *Service) newEvidenceMap(meta *runMeta, item evidenceDraft) g.Map {
	now := gtime.Now()
	return g.Map{
		"verification_run_id": meta.RunID,
		"evidence_type":       item.EvidenceType,
		"source_type":         item.SourceType,
		"source_id":           nullableInt64(item.SourceID),
		"content_ref":         item.ContentRef,
		"summary":             item.Summary,
		"created_at":          now,
		"updated_at":          now,
	}
}

func (s *Service) failRun(ctx context.Context, runID int64, reason string) {
	meta, err := s.loadRunMeta(ctx, runID)
	if err != nil {
		g.Log().Warningf(ctx, "[Verification] 载入 run meta 失败: runID=%d err=%v", runID, err)
		return
	}
	_, upErr := s.runRepo.UpdateStatus(ctx, runID, statusRunning, statusFailed, g.Map{
		"decision":    decisionFailed,
		"summary":     truncateText(reason, 2000),
		"finished_at": gtime.Now(),
	})
	if upErr != nil {
		g.Log().Warningf(ctx, "[Verification] 标记失败失败: runID=%d err=%v", runID, upErr)
	}
	s.insertWorkflowEvent(ctx, meta.WorkflowRunID, "verification_run", "verification.failed", &runID, map[string]interface{}{
		"verification_id": runID,
		"summary":         truncateText(reason, 500),
	})
}

func (s *Service) insertWorkflowEvent(ctx context.Context, workflowRunID int64, entityType, eventType string, entityID *int64, payload map[string]interface{}) {
	if err := event.PersistRecord(ctx, event.Event{
		WorkflowRunID: workflowRunID,
		EntityType:    entityType,
		EntityID:      entityID,
		EventType:     eventType,
		Payload:       payload,
	}); err != nil {
		g.Log().Warningf(ctx, "[Verification] 写入事件失败: event=%s workflowRunID=%d err=%v", eventType, workflowRunID, err)
	}
}

func (s *Service) findRelatedDomainTaskID(ctx context.Context, workflowRunID int64, resourceRef string) int64 {
	resourceRef = strings.TrimSpace(resourceRef)
	if workflowRunID == 0 || resourceRef == "" || resourceRef == "." {
		return 0
	}
	record, err := s.domainTaskRepo.FindLatestByWorkflowAndAffectedResourceLike(ctx, workflowRunID, resourceRef)
	if err != nil || len(record) == 0 {
		return 0
	}
	return g.NewVar(record["id"]).Int64()
}

func (s *Service) loadCategoryVerificationConfig(ctx context.Context, meta *runMeta) (*categoryVerificationConfig, []issueDraft, error) {
	record, err := s.findProjectCategoryRecord(ctx, meta)
	if err != nil || record == nil {
		return nil, nil, err
	}

	cfg := &categoryVerificationConfig{
		CategoryCode: strings.TrimSpace(g.NewVar(record["category_code"]).String()),
		DisplayName:  strings.TrimSpace(g.NewVar(record["display_name"]).String()),
	}
	cfg.Source = buildCategoryConfigSource(cfg.CategoryCode, cfg.DisplayName)

	issues := make([]issueDraft, 0, 2)
	if text := strings.TrimSpace(g.NewVar(record["verification_profile_json"]).String()); text != "" {
		var profile verificationProfile
		if err := json.Unmarshal([]byte(text), &profile); err != nil {
			issues = append(issues, issueDraft{
				IssueType:       "config",
				Severity:        "error",
				Title:           "分类默认验证配置解析失败",
				Detail:          fmt.Sprintf("%s 的 verification_profile_json 解析失败: %v", cfg.Source, err),
				ExpectedValue:   "分类默认验证配置为合法 verification profile JSON",
				ActualValue:     err.Error(),
				SuggestedAction: "修复项目分类中的 verification_profile_json，或删除后回退到项目级配置/自动检测",
			})
		} else {
			cfg.Profile = &profile
			cfg.ProfileRaw = text
		}
	}
	if text := strings.TrimSpace(g.NewVar(record["verification_gate_json"]).String()); text != "" {
		var gate verificationGate
		if err := json.Unmarshal([]byte(text), &gate); err != nil {
			issues = append(issues, issueDraft{
				IssueType:       "config",
				Severity:        "error",
				Title:           "分类验证放行规则解析失败",
				Detail:          fmt.Sprintf("%s 的 verification_gate_json 解析失败: %v", cfg.Source, err),
				ExpectedValue:   "分类验证放行规则为合法 verification gate JSON",
				ActualValue:     err.Error(),
				SuggestedAction: "修复项目分类中的 verification_gate_json，或删除后回退到默认放行策略",
			})
		} else if normalized, gateErr := normalizeVerificationGate(gate); gateErr != nil {
			issues = append(issues, issueDraft{
				IssueType:       "config",
				Severity:        "error",
				Title:           "分类验证放行规则无效",
				Detail:          fmt.Sprintf("%s 的 verification_gate_json 不合法: %v", cfg.Source, gateErr),
				ExpectedValue:   "allowedDecisions / requiredCheckKinds / allowedRunnerTypes 使用受支持值",
				ActualValue:     gateErr.Error(),
				SuggestedAction: "修复项目分类中的 verification_gate_json 后重新验证",
			})
		} else {
			cfg.Gate = normalized
			cfg.GateRaw = text
		}
	}

	if cfg.Profile == nil && cfg.Gate == nil && cfg.ProfileRaw == "" && cfg.GateRaw == "" {
		return nil, issues, nil
	}
	return cfg, issues, nil
}

func (s *Service) findProjectCategoryRecord(ctx context.Context, meta *runMeta) (g.Map, error) {
	if meta.CategoryCode != "" {
		record, err := s.projectCatRepo.GetByCode(ctx, meta.CategoryCode)
		if err != nil {
			return nil, err
		}
		if len(record) != 0 {
			return record, nil
		}
	}
	if meta.ProjectCategory != "" {
		record, err := s.projectCatRepo.GetByDisplayName(ctx, meta.ProjectCategory)
		if err != nil {
			return nil, err
		}
		if len(record) != 0 {
			return record, nil
		}
	}
	return nil, nil
}

func buildCategoryConfigSource(categoryCode, displayName string) string {
	if strings.TrimSpace(categoryCode) != "" {
		return "category:" + strings.TrimSpace(categoryCode)
	}
	if strings.TrimSpace(displayName) != "" {
		return "category:" + strings.TrimSpace(displayName)
	}
	return "category:unknown"
}

func normalizeVerificationGate(raw verificationGate) (*verificationGate, error) {
	allowedDecisions, err := normalizeDecisionList(raw.AllowedDecisions)
	if err != nil {
		return nil, err
	}
	requiredCheckKinds, err := normalizeCheckKindList(raw.RequiredCheckKinds)
	if err != nil {
		return nil, err
	}
	allowedRunnerTypes, err := normalizePlanRunnerTypeList(raw.AllowedRunnerTypes)
	if err != nil {
		return nil, err
	}

	gate := &verificationGate{
		AllowedDecisions:   allowedDecisions,
		MinExecutedSteps:   raw.MinExecutedSteps,
		RequiredCheckKinds: requiredCheckKinds,
		AllowedRunnerTypes: allowedRunnerTypes,
	}
	if gate.MinExecutedSteps < 0 {
		gate.MinExecutedSteps = 0
	}
	return gate, nil
}

func normalizeDecisionList(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := strings.TrimSpace(strings.ToLower(item))
		if value == "" {
			continue
		}
		switch value {
		case decisionPassed, decisionFailed, decisionManualReview:
		default:
			return nil, fmt.Errorf("allowedDecisions 包含不支持的值: %s", item)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func normalizeCheckKindList(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := normalizeCheckKind(item)
		if value == "" {
			return nil, fmt.Errorf("requiredCheckKinds 包含不支持的值: %s", item)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func normalizePlanRunnerTypeList(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := normalizePlanRunnerType(item)
		if value == "" {
			return nil, fmt.Errorf("allowedRunnerTypes 包含不支持的值: %s", item)
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		normalized = append(normalized, value)
	}
	return normalized, nil
}

func normalizePlanRunnerType(value string) string {
	switch strings.ReplaceAll(strings.TrimSpace(strings.ToLower(value)), "-", "_") {
	case modeGitHubActions, "github", "github_action", "gha":
		return modeGitHubActions
	case modeLocal:
	case modeDockerCompose, "compose":
	case modeDockerfile:
		return modeGitHubActions
	default:
		return ""
	}
}

func normalizeCheckKind(value string) string {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "lint":
		return "lint"
	case "test":
		return "test"
	case "build":
		return "build"
	case "runtime":
		return "runtime"
	case "browser", "e2e":
		return qualitygate.CheckKindBrowser
	default:
		return ""
	}
}

func evaluateVerificationGate(plan *executionPlan, executedSteps int, stepExecutions []stepExecution, existingIssues []issueDraft) []issueDraft {
	if plan == nil || plan.Gate == nil {
		return nil
	}

	gate := plan.Gate
	gateSource := valueOrDefault(plan.GateSource, "category gate")
	issues := make([]issueDraft, 0, 4)

	if len(gate.AllowedRunnerTypes) > 0 && !containsString(gate.AllowedRunnerTypes, normalizePlanRunnerType(plan.RunnerType)) {
		issues = append(issues, issueDraft{
			IssueType:       "gate",
			Severity:        "error",
			Title:           "分类验证规则未满足: runner 类型不允许",
			Detail:          fmt.Sprintf("%s 要求 runner_type 属于 %s，但本次验证使用的是 %s。", gateSource, strings.Join(gate.AllowedRunnerTypes, ", "), plan.RunnerType),
			ExpectedValue:   strings.Join(gate.AllowedRunnerTypes, ", "),
			ActualValue:     plan.RunnerType,
			SuggestedAction: "调整项目分类 verification_gate_json，确保 allowedRunnerTypes 与 github_actions 一致",
		})
	}

	if gate.MinExecutedSteps > 0 && executedSteps < gate.MinExecutedSteps {
		issues = append(issues, issueDraft{
			IssueType:       "gate",
			Severity:        "error",
			Title:           "分类验证规则未满足: 执行步骤不足",
			Detail:          fmt.Sprintf("%s 要求至少执行 %d 个验证步骤，本次仅执行 %d 个。", gateSource, gate.MinExecutedSteps, executedSteps),
			ExpectedValue:   fmt.Sprintf("执行步骤 >= %d", gate.MinExecutedSteps),
			ActualValue:     fmt.Sprintf("执行步骤 = %d", executedSteps),
			SuggestedAction: "补齐 GitHub Actions checks/checkKinds 或调整 verification_gate_json 的 minExecutedSteps",
		})
	}

	if len(gate.RequiredCheckKinds) > 0 {
		plannedKinds := collectCheckKindsFromSteps(append(append([]verificationStep{}, plan.SetupSteps...), plan.VerifySteps...))
		executedKinds := collectExecutedCheckKinds(stepExecutions)
		for _, kind := range gate.RequiredCheckKinds {
			if _, ok := plannedKinds[kind]; !ok {
				issues = append(issues, issueDraft{
					IssueType:       "gate",
					Severity:        "error",
					Title:           fmt.Sprintf("分类验证规则未满足: 缺少 %s 检查", kind),
					Detail:          fmt.Sprintf("%s 要求包含 %s 检查，但当前验证计划没有生成对应步骤。", gateSource, kind),
					ExpectedValue:   fmt.Sprintf("计划中包含 %s 检查", kind),
					ActualValue:     "未生成对应步骤",
					SuggestedAction: "在 GitHub Actions workflow 与 .easymvp/ci/latest.json 中补充对应检查类型",
				})
				continue
			}
			if _, ok := executedKinds[kind]; !ok {
				issues = append(issues, issueDraft{
					IssueType:       "gate",
					Severity:        "error",
					Title:           fmt.Sprintf("分类验证规则未满足: 未执行 %s 检查", kind),
					Detail:          fmt.Sprintf("%s 要求执行 %s 检查，但本次没有实际执行到对应步骤。", gateSource, kind),
					ExpectedValue:   fmt.Sprintf("已执行 %s 检查", kind),
					ActualValue:     "未执行",
					SuggestedAction: "确认 GitHub Actions workflow 已执行对应检查，并把最终状态回写到 .easymvp/ci/latest.json",
				})
			}
		}
	}

	currentDecision := decideRunResult(existingIssues, executedSteps)
	if len(gate.AllowedDecisions) > 0 && !containsString(gate.AllowedDecisions, currentDecision) {
		issues = append(issues, issueDraft{
			IssueType:       "gate",
			Severity:        "error",
			Title:           "分类验证规则未满足: 结果未达放行标准",
			Detail:          fmt.Sprintf("%s 允许的结果是 %s，但本次预判结果为 %s。", gateSource, strings.Join(gate.AllowedDecisions, ", "), currentDecision),
			ExpectedValue:   strings.Join(gate.AllowedDecisions, ", "),
			ActualValue:     currentDecision,
			SuggestedAction: "修复 GitHub Actions 失败项，或调整分类 verification_gate_json 的 allowedDecisions",
		})
	}

	return issues
}

func collectCheckKindsFromSteps(steps []verificationStep) map[string]struct{} {
	kinds := make(map[string]struct{})
	for _, step := range steps {
		kind := inferCheckKind(step)
		if kind == "" {
			continue
		}
		kinds[kind] = struct{}{}
	}
	return kinds
}

func collectExecutedCheckKinds(stepExecutions []stepExecution) map[string]struct{} {
	kinds := make(map[string]struct{})
	for _, item := range stepExecutions {
		if item.Result.Skipped {
			continue
		}
		kind := inferCheckKind(item.Step)
		if kind == "" {
			continue
		}
		kinds[kind] = struct{}{}
	}
	return kinds
}

func inferCheckKind(step verificationStep) string {
	return qualitygate.InferCheckKind(step.Name, step.Command, step.Runner)
}

func discoverBrowserScriptNames(scripts map[string]string) []string {
	names := make([]string, 0, len(scripts))
	for name, script := range scripts {
		if qualitygate.IsBrowserScript(name, script) {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func mergeVerificationStandardIntoGate(gate *verificationGate, standard qualitygate.VerificationStandard) *verificationGate {
	if len(standard.RequiredCheckKinds) == 0 {
		return gate
	}
	if gate == nil {
		gate = &verificationGate{}
	}
	for _, kind := range standard.RequiredCheckKinds {
		if containsString(gate.RequiredCheckKinds, kind) {
			continue
		}
		gate.RequiredCheckKinds = append(gate.RequiredCheckKinds, kind)
	}
	sort.Strings(gate.RequiredCheckKinds)
	return gate
}

func mergeGateSource(current string, standardCode string) string {
	standardCode = strings.TrimSpace(standardCode)
	if standardCode == "" {
		return current
	}
	extra := "standard:" + standardCode
	if current == "" {
		return extra
	}
	if strings.Contains(current, extra) {
		return current
	}
	return current + "+" + extra
}

func containsString(items []string, target string) bool {
	for _, item := range items {
		if item == target {
			return true
		}
	}
	return false
}

func loadVerificationProfile(root string) (*verificationProfile, string, string, error) {
	configPath := filepath.Join(root, ".easymvp", "verification.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, configPath, "", nil
		}
		return nil, configPath, "", err
	}
	var profile verificationProfile
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, configPath, string(data), err
	}
	return &profile, configPath, string(data), nil
}

func findComposeFile(root string) (string, error) {
	candidates := []string{"compose.yaml", "compose.yml", "docker-compose.yml", "docker-compose.yaml"}
	files := findFiles(root, 2, func(path string, info os.DirEntry) bool {
		name := info.Name()
		for _, candidate := range candidates {
			if name == candidate {
				return true
			}
		}
		return false
	})
	if len(files) == 0 {
		return "", os.ErrNotExist
	}
	sort.Strings(files)
	return files[0], nil
}

func findDockerfile(root string) (string, error) {
	files := findFiles(root, 2, func(path string, info os.DirEntry) bool {
		name := info.Name()
		return name == "Dockerfile" || strings.HasPrefix(name, "Dockerfile.")
	})
	if len(files) == 0 {
		return "", os.ErrNotExist
	}
	sort.Strings(files)
	return files[0], nil
}

func findFiles(root string, maxDepth int, matcher func(path string, info os.DirEntry) bool) []string {
	root = filepath.Clean(root)
	results := make([]string, 0)
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path == root {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		depth := strings.Count(filepath.ToSlash(rel), "/")
		if entry.IsDir() && depth >= maxDepth {
			return filepath.SkipDir
		}
		if entry.Type().IsRegular() && matcher(path, entry) {
			results = append(results, path)
		}
		return nil
	})
	return results
}

func findProjectSubdirs(root, filename string) []string {
	files := findFiles(root, 1, func(path string, info os.DirEntry) bool {
		return info.Name() == filename
	})
	dirs := make([]string, 0, len(files))
	for _, file := range files {
		dirs = append(dirs, filepath.Dir(file))
	}
	sort.Strings(dirs)
	return dirs
}

func readPackageScripts(dir string) (map[string]string, error) {
	data, err := os.ReadFile(filepath.Join(dir, "package.json"))
	if err != nil {
		return nil, err
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return nil, err
	}
	return pkg.Scripts, nil
}

func detectPackageManager(dir string) string {
	switch {
	case fileExists(filepath.Join(dir, "pnpm-lock.yaml")):
		return "pnpm"
	case fileExists(filepath.Join(dir, "yarn.lock")):
		return "yarn"
	default:
		return "npm"
	}
}

func hasAutoDiscoverablePackageScripts(scripts map[string]string) bool {
	if len(scripts) == 0 {
		return false
	}
	for _, name := range []string{"lint", "test", "build"} {
		if _, ok := scripts[name]; ok {
			return true
		}
	}
	return false
}

func buildPackageInstallCommand(dir, pm string) []string {
	switch pm {
	case "pnpm":
		command := []string{"pnpm", "install"}
		if fileExists(filepath.Join(dir, "pnpm-lock.yaml")) {
			command = append(command, "--frozen-lockfile")
		}
		return command
	case "yarn":
		command := []string{"yarn", "install"}
		if fileExists(filepath.Join(dir, "yarn.lock")) {
			command = append(command, "--frozen-lockfile")
		}
		return command
	default:
		if fileExists(filepath.Join(dir, "package-lock.json")) || fileExists(filepath.Join(dir, "npm-shrinkwrap.json")) {
			return []string{"npm", "ci"}
		}
		return []string{"npm", "install"}
	}
}

func buildPackageScriptCommand(pm, script string) []string {
	switch pm {
	case "yarn":
		return []string{"yarn", script}
	default:
		return []string{pm, "run", script}
	}
}

func buildTestScriptCommand(pm, scriptBody string) []string {
	lower := strings.ToLower(scriptBody)
	switch pm {
	case "yarn":
		if strings.Contains(lower, "vitest") {
			return []string{"yarn", "test", "--run"}
		}
		if strings.Contains(lower, "react-scripts test") || strings.Contains(lower, "jest") {
			return []string{"yarn", "test", "--watchAll=false"}
		}
		return []string{"yarn", "test"}
	default:
		if strings.Contains(lower, "vitest") {
			return []string{pm, "run", "test", "--", "--run"}
		}
		if strings.Contains(lower, "react-scripts test") || strings.Contains(lower, "jest") {
			return []string{pm, "run", "test", "--", "--watchAll=false"}
		}
		return []string{pm, "run", "test"}
	}
}

func describePackageInstallCommand(command []string) string {
	if len(command) == 0 {
		return "install dependencies"
	}
	return strings.Join(command, " ")
}

func displayStepName(rel, label string) string {
	if rel == "." || rel == "" {
		return label
	}
	return rel + " " + label
}

func relativePath(root, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil || rel == "." {
		return "."
	}
	return filepath.ToSlash(rel)
}

func normalizeMode(mode string) string {
	switch strings.TrimSpace(strings.ToLower(mode)) {
	case modeGitHubActions, "github", "github_action", "gha":
		return modeGitHubActions
	case modeDockerCompose:
	case modeDockerfile:
	case modeLocal:
		return modeGitHubActions
	default:
		return modeGitHubActions
	}
}

func normalizeRunner(runner string) string {
	switch strings.TrimSpace(strings.ToLower(runner)) {
	case runnerGitHub, "github", "github_action", "gha":
		return runnerGitHub
	case "docker", runnerDockerExec:
		return runnerDockerExec
	case "", runnerLocal:
		return runnerLocal
	default:
		return runnerLocal
	}
}

func detectComposePrefix() ([]string, error) {
	dockerPath, err := exec.LookPath("docker")
	if err == nil {
		if exec.Command(dockerPath, "compose", "version").Run() == nil {
			return []string{dockerPath, "compose"}, nil
		}
	}
	composePath, composeErr := exec.LookPath("docker-compose")
	if composeErr == nil {
		return []string{composePath}, nil
	}
	if err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("未找到 docker compose 或 docker-compose")
}

func buildCommand(root string, profile *verificationProfile, step verificationStep) ([]string, string, string, error) {
	switch step.Runner {
	case runnerDockerExec:
		return buildDockerExecCommand(root, profile, step)
	default:
		dir, err := resolveWithinRoot(root, valueOrDefault(step.WorkDir, "."))
		if err != nil {
			return nil, "", "", err
		}
		commandText := shellJoin(step.Command)
		return step.Command, dir, commandText, nil
	}
}

func describeStepCommand(step verificationStep) string {
	if len(step.Command) == 0 {
		return step.Name
	}
	return shellJoin(step.Command)
}

func buildDockerExecCommand(root string, profile *verificationProfile, step verificationStep) ([]string, string, string, error) {
	if step.Service == "" {
		return nil, "", "", fmt.Errorf("docker_exec 步骤缺少 service")
	}
	composeFile, _ := findComposeFile(root)
	if profile != nil && profile.Docker != nil && strings.TrimSpace(profile.Docker.ComposeFile) != "" {
		resolved, err := resolveWithinRoot(root, profile.Docker.ComposeFile)
		if err != nil {
			return nil, "", "", err
		}
		composeFile = resolved
	}
	if composeFile == "" {
		return nil, "", "", fmt.Errorf("docker_exec 步骤未找到 compose 文件")
	}

	prefix, err := detectComposePrefix()
	if err != nil {
		return nil, "", "", err
	}
	args := append([]string{}, prefix...)
	args = append(args, "-f", composeFile)
	projectName := ""
	if profile != nil && profile.Docker != nil {
		projectName = strings.TrimSpace(profile.Docker.ProjectName)
		if envFile := strings.TrimSpace(profile.Docker.EnvFile); envFile != "" {
			if resolved, envErr := resolveWithinRoot(root, envFile); envErr == nil {
				args = append(args, "--env-file", resolved)
			}
		}
	}
	if projectName != "" {
		args = append(args, "--project-name", projectName)
	}

	shellCommand := shellJoin(step.Command)
	if strings.TrimSpace(step.WorkDir) != "" && step.WorkDir != "." {
		shellCommand = "cd " + shellQuote(step.WorkDir) + " && " + shellCommand
	}
	if len(step.Env) > 0 {
		shellCommand = buildShellEnvPrefix(step.Env) + shellCommand
	}
	args = append(args, "exec", "-T", step.Service, "sh", "-lc", shellCommand)
	commandText := shellJoin(args)
	return args, root, commandText, nil
}

func resolveWithinRoot(root, rel string) (string, error) {
	root = filepath.Clean(root)
	if rel == "" || rel == "." {
		return root, nil
	}
	target := filepath.Clean(filepath.Join(root, rel))
	if target != root && !strings.HasPrefix(target, root+string(filepath.Separator)) {
		return "", fmt.Errorf("路径越界: %s", rel)
	}
	return target, nil
}

func buildEnvPairs(values map[string]string) []string {
	if len(values) == 0 {
		return nil
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}

func buildShellEnvPrefix(values map[string]string) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+shellQuote(values[key]))
	}
	return strings.Join(parts, " ") + " "
}

func decideRunResult(issues []issueDraft, executedSteps int) string {
	hasHardFailure := false
	for _, item := range issues {
		if item.Severity == "blocker" || item.Severity == "error" {
			hasHardFailure = true
			break
		}
	}
	switch {
	case hasHardFailure:
		return decisionFailed
	case executedSteps == 0:
		return decisionManualReview
	default:
		return decisionPassed
	}
}

func buildRunSummary(runnerType string, executedSteps int, issues []issueDraft) string {
	var blockers, errorsCount, warns, infos int
	for _, item := range issues {
		switch item.Severity {
		case "blocker":
			blockers++
		case "error":
			errorsCount++
		case "warn":
			warns++
		case "info":
			infos++
		}
	}
	return fmt.Sprintf(
		"验证完成：runner=%s，执行步骤=%d，blocker=%d，error=%d，warn=%d，info=%d。",
		runnerType, executedSteps, blockers, errorsCount, warns, infos,
	)
}

func truncateText(text string, limit int) string {
	if limit <= 0 || len(text) <= limit {
		return text
	}
	return text[:limit] + "\n...[truncated]"
}

func nullableInt64(value int64) interface{} {
	if value <= 0 {
		return nil
	}
	return value
}

func valueOrDefault(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func shellJoin(parts []string) string {
	if len(parts) == 0 {
		return ""
	}
	joined := make([]string, 0, len(parts))
	for _, part := range parts {
		joined = append(joined, shellQuote(part))
	}
	return strings.Join(joined, " ")
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'"'"'`) + "'"
}
