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
	"easymvp/app/mvp/internal/workflow/repo"
	"easymvp/utility/snowflake"
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

	runnerLocal      = "local"
	runnerDockerExec = "docker_exec"
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
	runRepo      *repo.VerificationRunRepo
	issueRepo    *repo.VerificationIssueRepo
	evidenceRepo *repo.VerificationEvidenceRepo
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
		runRepo:      runRepo,
		issueRepo:    issueRepo,
		evidenceRepo: evidenceRepo,
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
				Detail:          "Docker 环境启动失败，容器内验证步骤被跳过。",
				ExpectedValue:   "环境启动成功并继续执行容器内验证步骤",
				ActualValue:     "环境启动失败，步骤被跳过",
				SuggestedAction: "修复 compose/Dockerfile 或在 .easymvp/verification.json 中指定可执行的本机回退步骤",
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

	issues = append(issues, evaluateVerificationGate(plan, executedSteps, stepExecutions, issues)...)

	if executedSteps == 0 {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "未执行任何验证步骤",
			Detail:          "系统没有检测到可执行的验证命令，当前结果不足以判断项目是否可发布。",
			ExpectedValue:   "至少执行 1 个有效验证步骤",
			ActualValue:     "执行步骤数为 0",
			SuggestedAction: "在项目根目录补充 .easymvp/verification.json，显式声明 Docker 启动和验证命令",
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

	project, err := g.DB().Model("mvp_project").Ctx(ctx).
		Where("id", g.NewVar(run["project_id"]).Int64()).
		WhereNull("deleted_at").
		Fields("id, name, work_dir, category_code, project_category, created_by, dept_id").
		One()
	if err != nil || project.IsEmpty() {
		return nil, fmt.Errorf("project(%d) 不存在", g.NewVar(run["project_id"]).Int64())
	}

	return &runMeta{
		RunID:           runID,
		WorkflowRunID:   g.NewVar(run["workflow_run_id"]).Int64(),
		ProjectID:       project["id"].Int64(),
		ProjectName:     project["name"].String(),
		WorkDir:         strings.TrimSpace(project["work_dir"].String()),
		CategoryCode:    strings.TrimSpace(project["category_code"].String()),
		ProjectCategory: strings.TrimSpace(project["project_category"].String()),
		CreatedBy:       g.NewVar(run["created_by"]).Int64(),
		DeptID:          g.NewVar(run["dept_id"]).Int64(),
	}, nil
}

func (s *Service) buildExecutionPlan(ctx context.Context, meta *runMeta) (*executionPlan, []issueDraft, []evidenceDraft, error) {
	plan := &executionPlan{Mode: modeLocal, RunnerType: modeLocal}
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
		plan.ConfigSnapshot = `{"mode":"local","reason":"missing_workdir"}`
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
		plan.ConfigSnapshot = fmt.Sprintf(`{"mode":"local","workDir":%q}`, meta.WorkDir)
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

	composeFile, _ := findComposeFile(meta.WorkDir)
	dockerfile, _ := findDockerfile(meta.WorkDir)
	autoSteps := s.discoverLocalSteps(ctx, meta, meta.WorkDir)

	mode := modeAuto
	if profile != nil && strings.TrimSpace(profile.Mode) != "" {
		mode = normalizeMode(profile.Mode)
	}
	if mode == modeAuto {
		switch {
		case composeFile != "":
			mode = modeDockerCompose
		case dockerfile != "":
			mode = modeDockerfile
		default:
			mode = modeLocal
		}
	}
	plan.Mode = mode
	plan.RunnerType = mode

	var customSetup []verificationStep
	var customSteps []verificationStep
	var customTeardown []verificationStep
	if profile != nil {
		customSetup = s.normalizeSteps(ctx, meta, profile.SetupSteps)
		customSteps = s.normalizeSteps(ctx, meta, profile.Steps)
		customTeardown = s.normalizeSteps(ctx, meta, profile.TeardownSteps)
	}

	detection := map[string]interface{}{
		"mode":          mode,
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
	if composeFile != "" {
		detection["composeFile"] = composeFile
	}
	if dockerfile != "" {
		detection["dockerfile"] = dockerfile
	}
	if profileRaw != "" {
		detection["profile"] = json.RawMessage(profileRaw)
	}
	if categoryConfig != nil && categoryConfig.GateRaw != "" {
		detection["gate"] = json.RawMessage(categoryConfig.GateRaw)
	}

	switch mode {
	case modeDockerCompose:
		composePath := composeFile
		projectName := ""
		envFile := ""
		buildEnabled := true
		downAfter := true
		if profile != nil && profile.Docker != nil {
			if profile.Docker.ComposeFile != "" {
				if resolved, err := resolveWithinRoot(meta.WorkDir, profile.Docker.ComposeFile); err == nil {
					composePath = resolved
				} else {
					issues = append(issues, issueDraft{
						IssueType:       "config",
						Severity:        "error",
						Title:           "验证 composeFile 无效",
						Detail:          err.Error(),
						ExpectedValue:   "composeFile 位于项目目录内",
						ActualValue:     profile.Docker.ComposeFile,
						SuggestedAction: "修正 .easymvp/verification.json 中的 docker.composeFile",
					})
				}
			}
			if profile.Docker.EnvFile != "" {
				if resolved, err := resolveWithinRoot(meta.WorkDir, profile.Docker.EnvFile); err == nil {
					envFile = resolved
				} else {
					issues = append(issues, issueDraft{
						IssueType:       "config",
						Severity:        "error",
						Title:           "验证 envFile 无效",
						Detail:          err.Error(),
						ExpectedValue:   "envFile 位于项目目录内",
						ActualValue:     profile.Docker.EnvFile,
						SuggestedAction: "修正 .easymvp/verification.json 中的 docker.envFile",
					})
				}
			}
			projectName = strings.TrimSpace(profile.Docker.ProjectName)
			if profile.Docker.Build != nil {
				buildEnabled = *profile.Docker.Build
			}
		}
		if profile != nil && profile.DownAfter != nil {
			downAfter = *profile.DownAfter
		}

		prefix, dockerErr := detectComposePrefix()
		if composePath == "" {
			issues = append(issues, issueDraft{
				IssueType:       "environment",
				Severity:        "blocker",
				Title:           "未找到 docker compose 文件",
				Detail:          "自动检测和配置都没有找到 compose.yaml / docker-compose.yml。",
				ExpectedValue:   "项目包含可启动的 compose 文件",
				ActualValue:     "compose 文件缺失",
				SuggestedAction: "补充 compose 文件或改用 .easymvp/verification.json 指定本机验证命令",
			})
			plan.RunnerType = modeLocal
		} else if dockerErr != nil {
			issues = append(issues, issueDraft{
				IssueType:       "environment",
				Severity:        "blocker",
				Title:           "Docker Compose 不可用",
				Detail:          dockerErr.Error(),
				ExpectedValue:   "宿主机可执行 docker compose",
				ActualValue:     dockerErr.Error(),
				SuggestedAction: "安装 docker compose，或在验证配置中提供不依赖 Docker 的回退步骤",
			})
			plan.RunnerType = modeLocal
		} else {
			composeArgs := append([]string{}, prefix...)
			composeArgs = append(composeArgs, "-f", composePath, "--project-name", projectName)
			if envFile != "" {
				composeArgs = append(composeArgs, "--env-file", envFile)
			}
			up := append([]string{}, composeArgs...)
			up = append(up, "up", "-d")
			if buildEnabled {
				up = append(up, "--build")
			}
			plan.SetupSteps = append(plan.SetupSteps, verificationStep{
				Name:           "docker compose up",
				Runner:         runnerLocal,
				Command:        up,
				WorkDir:        ".",
				TimeoutSeconds: 900,
				ResourceRef:    ".",
				Expected:       "Docker compose 环境成功启动",
			})
			ps := append([]string{}, composeArgs...)
			ps = append(ps, "ps")
			plan.VerifySteps = append(plan.VerifySteps, verificationStep{
				Name:           "docker compose ps",
				Runner:         runnerLocal,
				Command:        ps,
				WorkDir:        ".",
				TimeoutSeconds: 120,
				ResourceRef:    ".",
				Expected:       "服务容器均处于 running/healthy 状态",
			})
			if downAfter {
				down := append([]string{}, composeArgs...)
				down = append(down, "down", "--remove-orphans")
				plan.TeardownSteps = append(plan.TeardownSteps, verificationStep{
					Name:           "docker compose down",
					Runner:         runnerLocal,
					Command:        down,
					WorkDir:        ".",
					TimeoutSeconds: 300,
					Optional:       true,
					ResourceRef:    ".",
				})
			}
		}
	case modeDockerfile:
		dockerPath, dockerErr := exec.LookPath("docker")
		buildPath := dockerfile
		buildEnabled := true
		if profile != nil && profile.Docker != nil {
			if profile.Docker.Build != nil {
				buildEnabled = *profile.Docker.Build
			}
		}
		if buildPath == "" {
			issues = append(issues, issueDraft{
				IssueType:       "environment",
				Severity:        "blocker",
				Title:           "未找到 Dockerfile",
				Detail:          "项目未检测到可用于构建验证镜像的 Dockerfile。",
				ExpectedValue:   "项目包含 Dockerfile",
				ActualValue:     "Dockerfile 缺失",
				SuggestedAction: "补充 Dockerfile 或改用 .easymvp/verification.json 声明验证方案",
			})
			plan.RunnerType = modeLocal
		} else if dockerErr != nil {
			issues = append(issues, issueDraft{
				IssueType:       "environment",
				Severity:        "blocker",
				Title:           "Docker 不可用",
				Detail:          dockerErr.Error(),
				ExpectedValue:   "宿主机可执行 docker build",
				ActualValue:     dockerErr.Error(),
				SuggestedAction: "安装 Docker，或在验证配置中声明可执行的本机验证步骤",
			})
			plan.RunnerType = modeLocal
		} else if buildEnabled {
			buildCmd := []string{
				dockerPath,
				"build",
				"-f", buildPath,
				"-t", fmt.Sprintf("easymvp-verify-%d-%d", meta.ProjectID, meta.RunID),
				meta.WorkDir,
			}
			plan.SetupSteps = append(plan.SetupSteps, verificationStep{
				Name:           "docker build",
				Runner:         runnerLocal,
				Command:        buildCmd,
				WorkDir:        ".",
				TimeoutSeconds: 900,
				ResourceRef:    ".",
				Expected:       "Docker 镜像成功构建",
			})
		}
	default:
		plan.RunnerType = modeLocal
	}

	plan.SetupSteps = append(plan.SetupSteps, customSetup...)
	if len(customSteps) > 0 {
		plan.VerifySteps = append(plan.VerifySteps, customSteps...)
	} else {
		plan.VerifySteps = append(plan.VerifySteps, autoSteps...)
	}
	plan.TeardownSteps = append(plan.TeardownSteps, customTeardown...)

	if len(plan.VerifySteps) == 0 {
		issues = append(issues, issueDraft{
			IssueType:       "config",
			Severity:        "warn",
			Title:           "未检测到验证命令",
			Detail:          "自动检测未找到 go test / npm lint / npm test / npm build 等可执行步骤。",
			ExpectedValue:   "至少存在 1 个验证命令",
			ActualValue:     "检测结果为空",
			SuggestedAction: "在 .easymvp/verification.json 中声明 steps，或补充测试/构建脚本",
		})
	}

	if plan.ConfigSnapshot == "" {
		encoded, _ := json.Marshal(detection)
		plan.ConfigSnapshot = string(encoded)
	}
	plan.DetectionSummary = fmt.Sprintf(
		"mode=%s profileSource=%s gateSource=%s compose=%t dockerfile=%t steps=%d",
		plan.RunnerType,
		profileSource,
		valueOrDefault(plan.GateSource, "none"),
		composeFile != "",
		dockerfile != "",
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
		if script, ok := scripts["test"]; ok {
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
	return issueDraft{
		IssueType:       issueType,
		Severity:        severity,
		Title:           fmt.Sprintf("验证步骤失败: %s", step.Name),
		Detail:          detail,
		ExpectedValue:   valueOrDefault(step.Expected, "命令执行成功"),
		ActualValue:     actual,
		SuggestedAction: "修复步骤对应的代码、配置或 Docker 环境后重新发起验证",
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
	var payloadJSON string
	if len(payload) > 0 {
		if data, err := json.Marshal(payload); err == nil {
			payloadJSON = string(data)
		}
	}

	data := g.Map{
		"id":              int64(snowflake.Generate()),
		"workflow_run_id": workflowRunID,
		"entity_type":     entityType,
		"event_type":      eventType,
		"payload":         payloadJSON,
		"created_at":      gtime.Now(),
	}
	if entityID != nil {
		data["entity_id"] = *entityID
	}
	if _, err := g.DB().Model("mvp_workflow_event").Ctx(ctx).Insert(data); err != nil {
		g.Log().Warningf(ctx, "[Verification] 写入事件失败: event=%s workflowRunID=%d err=%v", eventType, workflowRunID, err)
	}
}

func (s *Service) findRelatedDomainTaskID(ctx context.Context, workflowRunID int64, resourceRef string) int64 {
	resourceRef = strings.TrimSpace(resourceRef)
	if workflowRunID == 0 || resourceRef == "" || resourceRef == "." {
		return 0
	}
	patternA := fmt.Sprintf("%%%s%%", resourceRef)
	record, err := g.DB().Model("mvp_domain_task").Ctx(ctx).
		Where("workflow_run_id", workflowRunID).
		Where("affected_resources LIKE ?", patternA).
		WhereNull("deleted_at").
		OrderDesc("updated_at").
		Fields("id").
		One()
	if err != nil || record.IsEmpty() {
		return 0
	}
	return record["id"].Int64()
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
	model := g.DB().Model("mvp_project_category").Ctx(ctx).
		Where("status", 1).
		WhereNull("deleted_at")

	if meta.CategoryCode != "" {
		record, err := model.Clone().Where("category_code", meta.CategoryCode).
			Fields("category_code, display_name, verification_profile_json, verification_gate_json").
			One()
		if err != nil {
			return nil, err
		}
		if !record.IsEmpty() {
			return record.Map(), nil
		}
	}
	if meta.ProjectCategory != "" {
		record, err := model.Clone().Where("display_name", meta.ProjectCategory).
			Fields("category_code, display_name, verification_profile_json, verification_gate_json").
			One()
		if err != nil {
			return nil, err
		}
		if !record.IsEmpty() {
			return record.Map(), nil
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
	case modeLocal:
		return modeLocal
	case modeDockerCompose, "compose":
		return modeDockerCompose
	case modeDockerfile:
		return modeDockerfile
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
			SuggestedAction: "调整项目分类 verification_gate_json，或补充适配该分类的验证运行方式",
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
			SuggestedAction: "补充该分类要求的验证步骤，或调整 verification_gate_json 的最小执行步数",
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
					SuggestedAction: "在项目级 .easymvp/verification.json 或分类默认验证配置中补充对应验证步骤",
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
					SuggestedAction: "修复环境启动/命令配置问题，确保该分类要求的验证步骤能够实际执行",
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
			SuggestedAction: "修复验证问题，或调整分类 verification_gate_json 的 allowedDecisions",
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
	text := strings.ToLower(step.Name + " " + strings.Join(step.Command, " "))
	switch {
	case strings.Contains(text, "lint"):
		return "lint"
	case strings.Contains(text, "go test"), strings.Contains(text, " test"), strings.Contains(text, "vitest"), strings.Contains(text, "jest"), strings.Contains(text, "pytest"):
		return "test"
	case strings.Contains(text, "build"):
		return "build"
	case strings.Contains(text, "compose up"), strings.Contains(text, "compose ps"), strings.Contains(text, " start"), strings.Contains(text, " serve"), step.Runner == runnerDockerExec:
		return "runtime"
	default:
		return ""
	}
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
	case modeDockerCompose:
		return modeDockerCompose
	case modeDockerfile:
		return modeDockerfile
	case modeLocal:
		return modeLocal
	default:
		return modeAuto
	}
}

func normalizeRunner(runner string) string {
	switch strings.TrimSpace(strings.ToLower(runner)) {
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
