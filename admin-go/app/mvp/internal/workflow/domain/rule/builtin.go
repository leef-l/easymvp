package rule

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// BuiltinChecker 内置规则检查器，包含 5 条单任务可验证的基础规则。
//
// 跨任务规则（资源冲突、batch_no 一致性、depends_on 引用有效性、role_level 覆盖）
// 需要全局视图，保留在 engine/review_precheck.go 的 systemPrecheck 层处理（方案A）。
type BuiltinChecker struct {
	rules []builtinRule
}

type builtinRule struct {
	name     string
	severity string
	check    func(ctx context.Context, rctx *CheckContext) *CheckResult
}

// NewBuiltinChecker 创建内置规则检查器。
func NewBuiltinChecker() *BuiltinChecker {
	return &BuiltinChecker{
		rules: []builtinRule{
			{
				name:     "task_name_not_empty",
				severity: "error",
				check:    checkTaskNameNotEmpty,
			},
			{
				name:     "task_description_quality",
				severity: "error",
				check:    checkDescriptionQuality,
			},
			{
				name:     "affected_resources_format",
				severity: "error",
				check:    checkResourcesFormat,
			},
			{
				name:     "affected_resources_no_garbage",
				severity: "error",
				check:    checkResourcesNoGarbage,
			},
			{
				name:     "coding_no_directory_placeholder",
				severity: "error",
				check:    checkCodingNoDirectoryPlaceholder,
			},
		},
	}
}

// Check 执行所有内置规则，返回未通过的结果列表。
func (c *BuiltinChecker) Check(ctx context.Context, rctx *CheckContext) []CheckResult {
	var results []CheckResult
	for _, r := range c.rules {
		result := r.check(ctx, rctx)
		if result != nil {
			result.RuleName = r.name
			result.Severity = r.severity
			results = append(results, *result)
		}
	}
	return results
}

// ── 规则 1：任务名非空 ──────────────────────────────────────────────────────────

func checkTaskNameNotEmpty(_ context.Context, rctx *CheckContext) *CheckResult {
	if strings.TrimSpace(rctx.TaskName) == "" {
		return &CheckResult{Passed: false, Message: "任务名称为空"}
	}
	return nil // nil 表示规则通过
}

// ── 规则 2：任务描述质量 ────────────────────────────────────────────────────────

const minDescriptionRunes = 10

func checkDescriptionQuality(_ context.Context, rctx *CheckContext) *CheckResult {
	count := utf8.RuneCountInString(strings.TrimSpace(rctx.TaskDescription))
	if count < minDescriptionRunes {
		return &CheckResult{
			Passed: false,
			Message: fmt.Sprintf(
				"任务描述过短（%d字），需要至少%d字的有效描述",
				utf8.RuneCountInString(rctx.TaskDescription),
				minDescriptionRunes,
			),
		}
	}
	return nil
}

// ── 规则 3：affected_resources JSON 格式合法 ────────────────────────────────────

func checkResourcesFormat(_ context.Context, rctx *CheckContext) *CheckResult {
	// AffectedResources 由调用方解析后传入；此处做二次校验：
	// 若原始 JSON 为空/null/[] 则跳过；否则要求能解析为 string 切片。
	// 注意：CheckContext 已将 JSON 解析为 []string，格式错误会导致 AffectedResources 为 nil。
	// 规则直接检查传入字段——实际格式校验在调用方构建 CheckContext 时完成。
	// 此实现留作扩展点（可接收原始 JSON 字符串作为补充字段）。
	for _, res := range rctx.AffectedResources {
		if !json.Valid([]byte(`"` + res + `"`)) {
			return &CheckResult{
				Passed:  false,
				Message: fmt.Sprintf("affected_resources 包含非法字符串项: %s", res),
			}
		}
	}
	return nil
}

// ── 规则 4：affected_resources 无乱码路径 ──────────────────────────────────────

// garbageRe 检测路径是否包含控制字符、超长非 ASCII 序列或密集特殊符号。
// 与 engine/review_precheck.go 中的 garbageRe 保持一致，便于将来统一迁移。
var garbageRe = regexp.MustCompile(`[\x00-\x1f]|[^\x00-\x7f]{10,}|[!@#$%^&*()+=\[\]{}|\\;':",<>?]{3,}`)

func checkResourcesNoGarbage(_ context.Context, rctx *CheckContext) *CheckResult {
	for _, res := range rctx.AffectedResources {
		if garbageRe.MatchString(res) {
			return &CheckResult{
				Passed:  false,
				Message: fmt.Sprintf("affected_resources 包含疑似乱码路径: %s", res),
			}
		}
	}
	return nil
}

// ── 规则 5：编码类项目不允许目录占位 ──────────────────────────────────────────

// dirPlaceholderPatterns 目录占位符匹配模式（与 resourcepath 包保持一致）。
// 匹配以 / 结尾，或纯目录名（无扩展名且含路径分隔符）的模式。
var dirPlaceholderPatterns = []*regexp.Regexp{
	regexp.MustCompile(`/$`),            // 以 / 结尾
	regexp.MustCompile(`^\.\./`),        // 相对目录跳转
	regexp.MustCompile(`^[^.]+/[^.]+$`), // 含路径分隔符但无扩展名（如 src/utils）
}

const categoryFamilyCoding = "coding"

func checkCodingNoDirectoryPlaceholder(_ context.Context, rctx *CheckContext) *CheckResult {
	if rctx.CategoryFamily != categoryFamilyCoding {
		return nil // 仅对编码类项目启用
	}
	var placeholders []string
	for _, res := range rctx.AffectedResources {
		for _, pat := range dirPlaceholderPatterns {
			if pat.MatchString(res) {
				placeholders = append(placeholders, res)
				break
			}
		}
	}
	if len(placeholders) > 0 {
		return &CheckResult{
			Passed: false,
			Message: fmt.Sprintf(
				"affected_resources 必须写明确文件路径，不能使用目录占位: %s",
				strings.Join(placeholders, ", "),
			),
		}
	}
	return nil
}
