package qualitygate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	CheckKindLint    = "lint"
	CheckKindTest    = "test"
	CheckKindBuild   = "build"
	CheckKindRuntime = "runtime"
	CheckKindBrowser = "browser"

	RoleTypeExperienceReviewer = "experience_reviewer"

	ExperienceProfileWebInteractive = "web_interactive"
	ExperienceProfileGameClient     = "game_client_runtime"
	ExperienceProfileAndroidNative  = "android_native"
	ExperienceProfileIOSNative      = "ios_native"
)

type ProjectSignals struct {
	FamilyCode           string
	ProjectTypeCode      string
	HasGoModules         bool
	HasNodePackage       bool
	HasFrontendApp       bool
	HasBrowserAutomation bool
	HasAndroidApp        bool
	HasIOSApp            bool
	Reasons              []string
}

type ProjectRoleRequirement struct {
	RoleType      string
	RoleLevel     string
	DisplayName   string
	Purpose       string
	ReviewProfile string
	Blocking      bool
	UseForJudge   bool
}

func (r ProjectRoleRequirement) Label() string {
	if strings.TrimSpace(r.DisplayName) != "" {
		return strings.TrimSpace(r.DisplayName)
	}
	return strings.TrimSpace(r.RoleType)
}

func (r ProjectRoleRequirement) ExpectedRoleRef() string {
	label := r.Label()
	if strings.TrimSpace(r.RoleLevel) == "" {
		return fmt.Sprintf("%s(%s)", label, r.RoleType)
	}
	return fmt.Sprintf("%s(%s/%s)", label, r.RoleType, r.RoleLevel)
}

type VerificationStandard struct {
	Code                      string
	DisplayName               string
	RequiredCheckKinds        []string
	RequirePassedVerification bool
	RequireBrowserPlan        bool
	RequireBrowserEvidence    bool
	RequiredProjectRoles      []ProjectRoleRequirement
}

func (s VerificationStandard) JudgeRoleRequirement() (ProjectRoleRequirement, bool) {
	for _, requirement := range s.RequiredProjectRoles {
		if requirement.UseForJudge {
			return requirement, true
		}
	}
	return ProjectRoleRequirement{}, false
}

type packageManifest struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func DetectProjectSignals(workDir string, familyCode string, projectTypeCode string) ProjectSignals {
	signals := ProjectSignals{
		FamilyCode:      strings.TrimSpace(strings.ToLower(familyCode)),
		ProjectTypeCode: strings.TrimSpace(strings.ToLower(projectTypeCode)),
	}
	if strings.TrimSpace(workDir) == "" {
		return signals
	}

	goMods := findFilesWithinDepth(workDir, 2, func(path string, info os.DirEntry) bool {
		return info.Name() == "go.mod"
	})
	if len(goMods) > 0 {
		signals.HasGoModules = true
		signals.Reasons = append(signals.Reasons, "go_module")
	}

	androidMarkers := findFilesWithinDepth(workDir, 3, func(path string, info os.DirEntry) bool {
		name := strings.ToLower(info.Name())
		switch name {
		case "androidmanifest.xml", "build.gradle", "build.gradle.kts", "settings.gradle", "settings.gradle.kts":
			return true
		default:
			return false
		}
	})
	if len(androidMarkers) > 0 {
		signals.HasAndroidApp = true
		signals.Reasons = append(signals.Reasons, "android_app")
	}

	iosMarkers := findFilesWithinDepth(workDir, 3, func(path string, info os.DirEntry) bool {
		name := strings.ToLower(info.Name())
		switch {
		case strings.HasSuffix(name, ".xcodeproj"), strings.HasSuffix(name, ".xcworkspace"):
			return true
		case name == "package.swift":
			return true
		case name == "info.plist" && containsAny(strings.ToLower(path), "ios", ".xcodeproj", ".xcworkspace"):
			return true
		default:
			return false
		}
	})
	if len(iosMarkers) > 0 {
		signals.HasIOSApp = true
		signals.Reasons = append(signals.Reasons, "ios_app")
	}

	packages := findFilesWithinDepth(workDir, 2, func(path string, info os.DirEntry) bool {
		return info.Name() == "package.json"
	})
	for _, packagePath := range packages {
		signals.HasNodePackage = true
		dir := filepath.Dir(packagePath)
		manifest, err := readPackageManifest(packagePath)
		if err != nil {
			continue
		}

		if isFrontendAppPackage(dir, manifest) {
			signals.HasFrontendApp = true
			signals.Reasons = append(signals.Reasons, "frontend_app:"+relativeDir(workDir, dir))
		}
		if hasBrowserAutomation(dir, manifest) {
			signals.HasBrowserAutomation = true
			signals.Reasons = append(signals.Reasons, "browser_automation:"+relativeDir(workDir, dir))
		}
	}

	signals.Reasons = uniqueSortedStrings(signals.Reasons)
	return signals
}

func ResolveVerificationStandard(signals ProjectSignals) VerificationStandard {
	family := strings.TrimSpace(strings.ToLower(signals.FamilyCode))
	switch family {
	case "", "coding":
		standard := VerificationStandard{
			Code:                      "coding.backend",
			DisplayName:               "Coding Backend Standard",
			RequiredCheckKinds:        []string{CheckKindTest},
			RequirePassedVerification: true,
		}
		if signals.HasAndroidApp {
			required := nativeRequiredCheckKinds(signals)
			return VerificationStandard{
				Code:                      "coding.android_native_app",
				DisplayName:               "Coding Android Native App Standard",
				RequiredCheckKinds:        required,
				RequirePassedVerification: len(required) > 0,
				RequiredProjectRoles:      buildExperienceReviewerRequirements(ExperienceProfileAndroidNative),
			}
		}
		if signals.HasIOSApp {
			required := nativeRequiredCheckKinds(signals)
			return VerificationStandard{
				Code:                      "coding.ios_native_app",
				DisplayName:               "Coding iOS Native App Standard",
				RequiredCheckKinds:        required,
				RequirePassedVerification: len(required) > 0,
				RequiredProjectRoles:      buildExperienceReviewerRequirements(ExperienceProfileIOSNative),
			}
		}
		if signals.HasFrontendApp {
			required := []string{CheckKindBuild, CheckKindBrowser}
			if signals.HasGoModules {
				required = append(required, CheckKindTest)
			}
			profile := resolveExperienceReviewProfile(signals)
			code := "coding.interactive_delivery"
			displayName := "Coding Interactive Delivery Standard"
			if profile == ExperienceProfileGameClient {
				code = "coding.game_client_runtime"
				displayName = "Coding Game Client Runtime Standard"
			}
			standard = VerificationStandard{
				Code:                      code,
				DisplayName:               displayName,
				RequiredCheckKinds:        uniqueSortedStrings(required),
				RequirePassedVerification: true,
				RequireBrowserPlan:        true,
				RequireBrowserEvidence:    true,
				RequiredProjectRoles:      buildExperienceReviewerRequirements(profile),
			}
		}
		return standard
	case "analysis":
		return VerificationStandard{
			Code:        "analysis.default",
			DisplayName: "Analysis Default Standard",
		}
	case "creative":
		return VerificationStandard{
			Code:        "creative.default",
			DisplayName: "Creative Default Standard",
		}
	default:
		return VerificationStandard{
			Code:        family + ".default",
			DisplayName: "Default Standard",
		}
	}
}

func ExperienceReviewProfileLabel(profile string) string {
	switch strings.TrimSpace(profile) {
	case ExperienceProfileGameClient:
		return "游戏玩法与交互体验"
	case ExperienceProfileAndroidNative:
		return "Android 原生产品体验"
	case ExperienceProfileIOSNative:
		return "iOS 原生产品体验"
	case ExperienceProfileWebInteractive:
		return "Web 关键交互体验"
	default:
		return ""
	}
}

func BlueprintNeedsBrowserVerification(name string, description string, resources []string) bool {
	text := strings.ToLower(strings.TrimSpace(name + " " + description + " " + strings.Join(resources, " ")))
	return containsAny(text,
		"frontend", "react", "vue", "vite", "next", "nuxt", "svelte", "android", "ios", "iphone",
		"页面", "界面", "前端", "交互", "移动端", "客户端", "手游", "游戏界面",
		"src/app", "src/pages", "src/components", "app/src", ".tsx", ".jsx", ".vue", ".kt", ".swift", ".xcworkspace",
	)
}

func BlueprintProvidesBrowserVerification(name string, description string, resources []string) bool {
	text := strings.ToLower(strings.TrimSpace(name + " " + description + " " + strings.Join(resources, " ")))
	return containsAny(text,
		"e2e", "playwright", "cypress", "puppeteer", "webdriver", "web-test-runner",
		"espresso", "uiautomator", "xctest", "xcuitest", "detox", "maestro",
		"浏览器验证", "浏览器测试", "端到端", "真机验证", "关键交互验证",
		"/e2e/", "playwright.config", "cypress.config",
	)
}

func IsBrowserScript(name string, script string) bool {
	text := strings.ToLower(strings.TrimSpace(name + " " + script))
	return containsAny(text,
		"e2e", "playwright", "cypress", "puppeteer", "webdriver", "web-test-runner",
		"espresso", "uiautomator", "xctest", "xcuitest", "detox", "maestro",
		"browser", "端到端", "浏览器",
	)
}

func InferCheckKind(name string, command []string, runner string) string {
	text := strings.ToLower(strings.TrimSpace(name + " " + strings.Join(command, " ")))
	switch {
	case IsBrowserScript(name, strings.Join(command, " ")):
		return CheckKindBrowser
	case strings.Contains(text, "lint"):
		return CheckKindLint
	case strings.Contains(text, "go test"),
		strings.Contains(text, " vitest"),
		strings.Contains(text, " jest"),
		strings.Contains(text, " pytest"),
		strings.Contains(text, " test"):
		return CheckKindTest
	case strings.Contains(text, "build"):
		return CheckKindBuild
	case strings.Contains(text, "compose up"),
		strings.Contains(text, "compose ps"),
		strings.Contains(text, " start"),
		strings.Contains(text, " serve"),
		runner == "docker_exec":
		return CheckKindRuntime
	default:
		return ""
	}
}

func nativeRequiredCheckKinds(signals ProjectSignals) []string {
	required := make([]string, 0, 2)
	if signals.HasGoModules {
		required = append(required, CheckKindTest)
	}
	if signals.HasNodePackage {
		required = append(required, CheckKindBuild)
	}
	return uniqueSortedStrings(required)
}

func resolveExperienceReviewProfile(signals ProjectSignals) string {
	switch {
	case signals.HasAndroidApp:
		return ExperienceProfileAndroidNative
	case signals.HasIOSApp:
		return ExperienceProfileIOSNative
	case signals.ProjectTypeCode == "game_dev":
		return ExperienceProfileGameClient
	case signals.HasFrontendApp:
		return ExperienceProfileWebInteractive
	default:
		return ""
	}
}

func buildExperienceReviewerRequirements(profile string) []ProjectRoleRequirement {
	switch profile {
	case ExperienceProfileGameClient:
		return []ProjectRoleRequirement{{
			RoleType:      RoleTypeExperienceReviewer,
			DisplayName:   "游戏体验评审师",
			Purpose:       "验收关键玩法循环、交互反馈和可玩性闭环",
			ReviewProfile: profile,
			Blocking:      true,
			UseForJudge:   true,
		}}
	case ExperienceProfileAndroidNative:
		return []ProjectRoleRequirement{{
			RoleType:      RoleTypeExperienceReviewer,
			DisplayName:   "Android 体验评审师",
			Purpose:       "验收 Android 原生端关键流程、状态反馈和交互一致性",
			ReviewProfile: profile,
			Blocking:      true,
			UseForJudge:   true,
		}}
	case ExperienceProfileIOSNative:
		return []ProjectRoleRequirement{{
			RoleType:      RoleTypeExperienceReviewer,
			DisplayName:   "iOS 体验评审师",
			Purpose:       "验收 iOS 原生端关键流程、状态反馈和交互一致性",
			ReviewProfile: profile,
			Blocking:      true,
			UseForJudge:   true,
		}}
	case ExperienceProfileWebInteractive:
		return []ProjectRoleRequirement{{
			RoleType:      RoleTypeExperienceReviewer,
			DisplayName:   "产品体验评审师",
			Purpose:       "验收关键用户路径、交互反馈、视觉一致性和可发布性",
			ReviewProfile: profile,
			Blocking:      true,
			UseForJudge:   true,
		}}
	default:
		return nil
	}
}

func isFrontendAppPackage(dir string, manifest packageManifest) bool {
	if !hasFrontendFramework(manifest) {
		return false
	}
	if hasFrontendEntryFiles(dir) {
		return true
	}
	return hasAnyScript(manifest.Scripts, "dev", "start", "build")
}

func hasBrowserAutomation(dir string, manifest packageManifest) bool {
	for name, script := range manifest.Scripts {
		if IsBrowserScript(name, script) {
			return true
		}
	}
	for _, candidate := range []string{
		"playwright.config.ts",
		"playwright.config.js",
		"cypress.config.ts",
		"cypress.config.js",
		"detox.config.js",
		".maestro",
		"e2e",
	} {
		if info, err := os.Stat(filepath.Join(dir, candidate)); err == nil {
			if info.IsDir() || info.Mode().IsRegular() {
				return true
			}
		}
	}
	return false
}

func hasFrontendFramework(manifest packageManifest) bool {
	for _, deps := range []map[string]string{manifest.Dependencies, manifest.DevDependencies} {
		for name := range deps {
			switch strings.ToLower(strings.TrimSpace(name)) {
			case "react", "react-dom", "next", "vue", "nuxt", "svelte", "@sveltejs/kit",
				"solid-js", "preact", "@angular/core", "expo", "react-native", "@ionic/react":
				return true
			}
		}
	}
	return false
}

func hasFrontendEntryFiles(dir string) bool {
	candidates := []string{
		"index.html",
		"src/App.tsx",
		"src/App.jsx",
		"src/App.vue",
		"src/main.tsx",
		"src/main.jsx",
		"src/main.ts",
		"src/main.js",
		"app/page.tsx",
		"pages/index.tsx",
		"App.tsx",
		"App.vue",
	}
	for _, candidate := range candidates {
		if info, err := os.Stat(filepath.Join(dir, candidate)); err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}

func hasAnyScript(scripts map[string]string, names ...string) bool {
	for _, name := range names {
		if _, ok := scripts[name]; ok {
			return true
		}
	}
	return false
}

func readPackageManifest(path string) (packageManifest, error) {
	var manifest packageManifest
	content, err := os.ReadFile(path)
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(content, &manifest); err != nil {
		return manifest, err
	}
	if manifest.Scripts == nil {
		manifest.Scripts = map[string]string{}
	}
	if manifest.Dependencies == nil {
		manifest.Dependencies = map[string]string{}
	}
	if manifest.DevDependencies == nil {
		manifest.DevDependencies = map[string]string{}
	}
	return manifest, nil
}

func findFilesWithinDepth(root string, maxDepth int, match func(path string, info os.DirEntry) bool) []string {
	root = filepath.Clean(root)
	results := make([]string, 0, 8)
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if path != root {
			rel, relErr := filepath.Rel(root, path)
			if relErr == nil && depth(rel) > maxDepth {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if match(path, d) {
			results = append(results, path)
		}
		return nil
	})
	sort.Strings(results)
	return results
}

func depth(rel string) int {
	if rel == "." || rel == "" {
		return 0
	}
	return len(strings.Split(rel, string(filepath.Separator)))
}

func relativeDir(root string, dir string) string {
	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return dir
	}
	if rel == "." {
		return "."
	}
	return rel
}

func containsAny(text string, keywords ...string) bool {
	for _, keyword := range keywords {
		if strings.Contains(text, keyword) {
			return true
		}
	}
	return false
}

func uniqueSortedStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	items := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	sort.Strings(items)
	return items
}
