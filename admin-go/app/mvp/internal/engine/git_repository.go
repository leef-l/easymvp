package engine

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/utility/snowflake"
)

type GitAccount struct {
	ID            int64
	AccountName   string
	ProviderType  string
	GitUserName   string
	GitUserEmail  string
	RepoRoot      string
	DefaultBranch string
	InitReadme    bool
}

type ProjectRepositoryBinding struct {
	ID            int64
	ProjectID     int64
	GitAccountID  int64
	RepoName      string
	RepoPath      string
	DefaultBranch string
	InitCommitSHA string
	Status        string
	IsManaged     bool
	CreatedBy     int64
	DeptID        int64
}

var repoNameSanitizer = regexp.MustCompile(`[^a-z0-9]+`)

func ResolveDefaultGitAccount(ctx context.Context, userID int64) (*GitAccount, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("用户未登录，无法解析 Git 配置")
	}

	record, err := g.DB().Ctx(ctx).Model("mvp_git_account").
		Where("created_by", userID).
		Where("status", 1).
		WhereNull("deleted_at").
		OrderDesc("is_default").
		OrderDesc("updated_at").
		One()
	if err != nil {
		return nil, fmt.Errorf("查询默认 Git 配置失败: %w", err)
	}
	if record.IsEmpty() {
		return nil, fmt.Errorf("请先在 Git 账号菜单配置并启用默认 Git 账号")
	}

	account := &GitAccount{
		ID:            record["id"].Int64(),
		AccountName:   record["account_name"].String(),
		ProviderType:  strings.TrimSpace(record["provider_type"].String()),
		GitUserName:   strings.TrimSpace(record["git_user_name"].String()),
		GitUserEmail:  strings.TrimSpace(record["git_user_email"].String()),
		RepoRoot:      strings.TrimSpace(record["repo_root"].String()),
		DefaultBranch: strings.TrimSpace(record["default_branch"].String()),
		InitReadme:    record["init_readme"].Int() == 1,
	}
	if account.ProviderType == "" {
		account.ProviderType = "local"
	}
	if account.DefaultBranch == "" {
		account.DefaultBranch = "main"
	}
	if account.GitUserName == "" || account.GitUserEmail == "" {
		return nil, fmt.Errorf("默认 Git 账号缺少 user.name 或 user.email")
	}
	if account.RepoRoot == "" {
		return nil, fmt.Errorf("默认 Git 账号缺少仓库根目录 repoRoot")
	}
	return account, nil
}

func PrepareProjectRepository(ctx context.Context, projectID, userID, deptID int64, projectName, requestedWorkDir string) (*ProjectRepositoryBinding, error) {
	account, err := ResolveDefaultGitAccount(ctx, userID)
	if err != nil {
		return nil, err
	}
	if account.ProviderType != "local" {
		return nil, fmt.Errorf("当前仅支持 local 类型 Git 账号自动建仓，请先将默认账号切换为 local")
	}
	if _, lookErr := exec.LookPath("git"); lookErr != nil {
		return nil, fmt.Errorf("系统未安装 git，无法自动初始化仓库: %w", lookErr)
	}

	repoPath, managed, err := resolveProjectRepoPath(account, projectID, projectName, requestedWorkDir)
	if err != nil {
		return nil, err
	}

	if err := ensureProjectRepositoryInitialized(ctx, repoPath, account); err != nil {
		if managed {
			_ = cleanupRepositoryPath(repoPath)
		}
		return nil, err
	}

	commitSHA, err := resolveHeadCommit(ctx, repoPath)
	if err != nil {
		if managed {
			_ = cleanupRepositoryPath(repoPath)
		}
		return nil, err
	}

	return &ProjectRepositoryBinding{
		ID:            int64(snowflake.Generate()),
		ProjectID:     projectID,
		GitAccountID:  account.ID,
		RepoName:      filepath.Base(repoPath),
		RepoPath:      repoPath,
		DefaultBranch: account.DefaultBranch,
		InitCommitSHA: commitSHA,
		Status:        "ready",
		IsManaged:     managed,
		CreatedBy:     userID,
		DeptID:        deptID,
	}, nil
}

func SaveProjectRepositoryBinding(ctx context.Context, binding *ProjectRepositoryBinding) error {
	if binding == nil || binding.ProjectID <= 0 {
		return fmt.Errorf("项目仓库绑定信息为空")
	}
	_, err := g.DB().Ctx(ctx).Model("mvp_project_repository").Insert(g.Map{
		"id":              binding.ID,
		"project_id":      binding.ProjectID,
		"git_account_id":  binding.GitAccountID,
		"repo_name":       binding.RepoName,
		"repo_path":       binding.RepoPath,
		"default_branch":  binding.DefaultBranch,
		"init_commit_sha": binding.InitCommitSHA,
		"status":          binding.Status,
		"is_managed":      boolToInt(binding.IsManaged),
		"created_by":      binding.CreatedBy,
		"dept_id":         binding.DeptID,
		"created_at":      gtime.Now(),
		"updated_at":      gtime.Now(),
	})
	if err != nil {
		return fmt.Errorf("保存项目仓库绑定失败: %w", err)
	}
	return nil
}

func ListProjectRepositoryBindings(ctx context.Context, projectIDs []int64) ([]ProjectRepositoryBinding, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	records, err := g.DB().Ctx(ctx).Model("mvp_project_repository").
		WhereIn("project_id", projectIDs).
		WhereNull("deleted_at").
		OrderAsc("id").
		All()
	if err != nil {
		return nil, fmt.Errorf("查询项目仓库绑定失败: %w", err)
	}

	result := make([]ProjectRepositoryBinding, 0, len(records))
	for _, record := range records {
		result = append(result, ProjectRepositoryBinding{
			ID:            record["id"].Int64(),
			ProjectID:     record["project_id"].Int64(),
			GitAccountID:  record["git_account_id"].Int64(),
			RepoName:      record["repo_name"].String(),
			RepoPath:      record["repo_path"].String(),
			DefaultBranch: record["default_branch"].String(),
			InitCommitSHA: record["init_commit_sha"].String(),
			Status:        record["status"].String(),
			IsManaged:     record["is_managed"].Int() == 1,
			CreatedBy:     record["created_by"].Int64(),
			DeptID:        record["dept_id"].Int64(),
		})
	}
	return result, nil
}

func CleanupManagedRepositories(ctx context.Context, bindings []ProjectRepositoryBinding) {
	for _, binding := range bindings {
		if !binding.IsManaged || strings.TrimSpace(binding.RepoPath) == "" {
			continue
		}
		if err := cleanupRepositoryPath(binding.RepoPath); err != nil {
			g.Log().Warningf(ctx, "[GitRepository] 清理托管仓库失败: projectID=%d repo=%s err=%v",
				binding.ProjectID, binding.RepoPath, err)
		}
	}
}

func CleanupProvisionedRepository(ctx context.Context, binding *ProjectRepositoryBinding) {
	if binding == nil || !binding.IsManaged {
		return
	}
	if err := cleanupRepositoryPath(binding.RepoPath); err != nil {
		g.Log().Warningf(ctx, "[GitRepository] 回滚仓库目录失败: repo=%s err=%v", binding.RepoPath, err)
	}
}

func resolveProjectRepoPath(account *GitAccount, projectID int64, projectName, requestedWorkDir string) (string, bool, error) {
	if account == nil {
		return "", false, fmt.Errorf("Git 账号不存在")
	}

	if requested := strings.TrimSpace(requestedWorkDir); requested != "" {
		path, created, err := EnsureWorkDir(requested)
		if err != nil {
			return "", false, err
		}
		return path, created, nil
	}

	root, _, err := EnsureWorkDir(account.RepoRoot)
	if err != nil {
		return "", false, fmt.Errorf("Git 仓库根目录不可用: %w", err)
	}
	repoPath := filepath.Join(root, buildManagedRepoDirName(projectName, projectID))
	if _, statErr := os.Stat(repoPath); statErr == nil {
		return "", false, fmt.Errorf("自动仓库目录已存在: %s", repoPath)
	} else if !os.IsNotExist(statErr) {
		return "", false, fmt.Errorf("检查自动仓库目录失败: %w", statErr)
	}

	path, _, err := EnsureWorkDir(repoPath)
	if err != nil {
		return "", false, err
	}
	return path, true, nil
}

func ensureProjectRepositoryInitialized(ctx context.Context, repoPath string, account *GitAccount) error {
	if repoPath == "" {
		return fmt.Errorf("仓库路径不能为空")
	}

	isRepo, err := isGitRepository(repoPath)
	if err != nil {
		return err
	}
	if !isRepo {
		if err := runGitCommand(ctx, repoPath, "init", "-b", account.DefaultBranch); err != nil {
			return fmt.Errorf("初始化 Git 仓库失败: %w", err)
		}
	}

	if err := runGitCommand(ctx, repoPath, "config", "user.name", account.GitUserName); err != nil {
		return fmt.Errorf("设置 git user.name 失败: %w", err)
	}
	if err := runGitCommand(ctx, repoPath, "config", "user.email", account.GitUserEmail); err != nil {
		return fmt.Errorf("设置 git user.email 失败: %w", err)
	}

	if account.InitReadme {
		readmePath := filepath.Join(repoPath, "README.md")
		if _, statErr := os.Stat(readmePath); os.IsNotExist(statErr) {
			content := []byte("# " + filepath.Base(repoPath) + "\n")
			if writeErr := os.WriteFile(readmePath, content, 0o644); writeErr != nil {
				return fmt.Errorf("创建 README 失败: %w", writeErr)
			}
		}
	}

	if err := runGitCommand(ctx, repoPath, "add", "-A"); err != nil {
		return fmt.Errorf("Git add 失败: %w", err)
	}
	if err := runGitCommand(ctx, repoPath, "commit", "--allow-empty", "-m", "chore: initialize project repository"); err != nil {
		if !strings.Contains(err.Error(), "nothing to commit") &&
			!strings.Contains(err.Error(), "working tree clean") {
			return fmt.Errorf("初始化提交失败: %w", err)
		}
	}
	return nil
}

func resolveHeadCommit(ctx context.Context, repoPath string) (string, error) {
	output, err := runGitCommandOutput(ctx, repoPath, "rev-parse", "HEAD")
	if err != nil {
		return "", fmt.Errorf("读取仓库提交 SHA 失败: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func isGitRepository(repoPath string) (bool, error) {
	info, err := os.Stat(filepath.Join(repoPath, ".git"))
	if err == nil {
		return info != nil, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("检查 Git 仓库失败: %w", err)
}

func runGitCommand(ctx context.Context, repoPath string, args ...string) error {
	_, err := runGitCommandOutput(ctx, repoPath, args...)
	return err
}

func runGitCommandOutput(ctx context.Context, repoPath string, args ...string) (string, error) {
	cmdCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(cmdCtx, "git", args...)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		text := strings.TrimSpace(string(output))
		if text == "" {
			text = err.Error()
		}
		return "", fmt.Errorf("%s", text)
	}
	return strings.TrimSpace(string(output)), nil
}

func cleanupRepositoryPath(repoPath string) error {
	repoPath = filepath.Clean(strings.TrimSpace(repoPath))
	if repoPath == "" || !filepath.IsAbs(repoPath) {
		return fmt.Errorf("仓库路径非法: %s", repoPath)
	}
	if err := checkPathBlacklist(repoPath); err == nil {
		return fmt.Errorf("拒绝删除系统目录: %s", repoPath)
	}
	if _, statErr := os.Stat(filepath.Join(repoPath, ".git")); statErr != nil {
		return fmt.Errorf("目录不是受管 Git 仓库: %s", repoPath)
	}
	return os.RemoveAll(repoPath)
}

func buildManagedRepoDirName(projectName string, projectID int64) string {
	name := strings.ToLower(strings.TrimSpace(projectName))
	name = repoNameSanitizer.ReplaceAllString(name, "-")
	name = strings.Trim(name, "-")
	if name == "" {
		name = "project"
	}
	if len(name) > 48 {
		name = strings.Trim(name[:48], "-")
	}
	if name == "" {
		name = "project"
	}
	return fmt.Sprintf("%s-%d", name, projectID)
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
