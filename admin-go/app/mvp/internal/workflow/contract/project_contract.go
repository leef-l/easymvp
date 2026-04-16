package contract

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"

	"easymvp/app/mvp/internal/workflow/repo"
)

const technicalContractKey = "technicalContract"

type ProjectContract struct {
	RequiredTechnologies  []string `json:"requiredTechnologies,omitempty"`
	ForbiddenTechnologies []string `json:"forbiddenTechnologies,omitempty"`
}

type technologyCatalogEntry struct {
	Canonical string
	Aliases   []string
}

var technologyCatalog = []technologyCatalogEntry{
	{Canonical: "sqlite", Aliases: []string{"sqlite", "sqlite3"}},
	{Canonical: "redis", Aliases: []string{"redis"}},
	{Canonical: "mysql", Aliases: []string{"mysql"}},
	{Canonical: "postgresql", Aliases: []string{"postgresql", "postgres"}},
	{Canonical: "mongodb", Aliases: []string{"mongodb", "mongo"}},
}

func Load(ctx context.Context, projectID int64) (*ProjectContract, error) {
	project, err := repo.NewProjectRepo().GetByID(ctx, projectID, "objective_json")
	if err != nil || len(project) == 0 {
		return &ProjectContract{}, err
	}
	return contractFromObjectiveJSON(gconv.String(project["objective_json"])), nil
}

func SyncFromConversation(ctx context.Context, projectID, conversationID int64) (*ProjectContract, error) {
	var (
		projectRepo = repo.NewProjectRepo()
		messageRepo = repo.NewMessageRepo()
	)

	project, err := projectRepo.GetByID(ctx, projectID, "description", "objective_json")
	if err != nil || len(project) == 0 {
		return &ProjectContract{}, err
	}

	latestUserPrompt := ""
	if conversationID > 0 {
		message, msgErr := messageRepo.GetLatestByConversationRoleStatus(ctx, conversationID, "user", "completed", "content")
		if msgErr == nil && len(message) > 0 {
			latestUserPrompt = gconv.String(message["content"])
		}
	}

	inferred := InferFromTexts(gconv.String(project["description"]), latestUserPrompt)
	rawJSON, changed, mergeErr := MergeContractIntoObjectiveJSON(gconv.String(project["objective_json"]), inferred)
	if mergeErr != nil {
		return &ProjectContract{}, mergeErr
	}
	if changed {
		if updateErr := projectRepo.UpdateFields(ctx, projectID, map[string]interface{}{"objective_json": rawJSON}); updateErr != nil {
			return &ProjectContract{}, updateErr
		}
	}
	return contractFromObjectiveJSON(rawJSON), nil
}

func InferFromTexts(texts ...string) *ProjectContract {
	raw := strings.ToLower(strings.Join(texts, "\n"))
	if strings.TrimSpace(raw) == "" {
		return &ProjectContract{}
	}

	required := make(map[string]struct{})
	forbidden := make(map[string]struct{})

	for _, tech := range technologyCatalog {
		if matchesRequiredDirective(raw, tech.Aliases) {
			required[tech.Canonical] = struct{}{}
		}
		if matchesForbiddenDirective(raw, tech.Aliases) {
			forbidden[tech.Canonical] = struct{}{}
		}
	}

	return &ProjectContract{
		RequiredTechnologies:  sortedKeys(required),
		ForbiddenTechnologies: sortedKeys(forbidden),
	}
}

func MergeContractIntoObjectiveJSON(raw string, contract *ProjectContract) (string, bool, error) {
	document, err := parseObjectiveJSON(raw)
	if err != nil {
		return "", false, err
	}

	existing := contractFromObjectiveDoc(document)
	merged := mergeContracts(existing, contract)
	if merged.IsEmpty() {
		if _, ok := document[technicalContractKey]; !ok {
			next, marshalErr := marshalObjectiveJSON(document)
			return next, false, marshalErr
		}
		delete(document, technicalContractKey)
		next, marshalErr := marshalObjectiveJSON(document)
		return next, true, marshalErr
	}

	if sameContract(existing, merged) {
		next, marshalErr := marshalObjectiveJSON(document)
		return next, false, marshalErr
	}

	document[technicalContractKey] = map[string]interface{}{
		"requiredTechnologies":  merged.RequiredTechnologies,
		"forbiddenTechnologies": merged.ForbiddenTechnologies,
	}
	next, marshalErr := marshalObjectiveJSON(document)
	return next, true, marshalErr
}

func MergeObjectiveFields(raw string, updates map[string]interface{}) (string, error) {
	document, err := parseObjectiveJSON(raw)
	if err != nil {
		return "", err
	}
	for key, value := range updates {
		if strings.TrimSpace(key) == "" {
			continue
		}
		document[key] = value
	}
	return marshalObjectiveJSON(document)
}

func BuildConstraintBlock(contract *ProjectContract) string {
	if contract == nil || contract.IsEmpty() {
		return ""
	}

	parts := []string{"## 项目级硬约束"}
	if len(contract.ForbiddenTechnologies) > 0 {
		parts = append(parts, "- 禁止使用: "+strings.Join(contract.ForbiddenTechnologies, ", "))
	}
	if len(contract.RequiredTechnologies) > 0 {
		parts = append(parts, "- 必须优先使用: "+strings.Join(contract.RequiredTechnologies, ", "))
	}
	return strings.Join(parts, "\n")
}

func AppendConstraintBlock(description string, contract *ProjectContract) string {
	block := BuildConstraintBlock(contract)
	if block == "" {
		return description
	}
	if strings.Contains(description, "## 项目级硬约束") {
		return description
	}
	if strings.TrimSpace(description) == "" {
		return block
	}
	return strings.TrimSpace(description) + "\n\n" + block
}

func DetectConflicts(contract *ProjectContract, texts ...string) []string {
	if contract == nil || contract.IsEmpty() {
		return nil
	}

	combined := strings.ToLower(strings.Join(texts, "\n"))
	var conflicts []string
	for _, tech := range contract.ForbiddenTechnologies {
		if mentionsTechnology(combined, tech) {
			conflicts = append(conflicts, "检测到禁止技术: "+tech)
		}
	}
	sort.Strings(conflicts)
	return uniqueStrings(conflicts)
}

func DetectMissingRequired(contract *ProjectContract, texts ...string) []string {
	if contract == nil || len(contract.RequiredTechnologies) == 0 {
		return nil
	}

	combined := strings.ToLower(strings.Join(texts, "\n"))
	var missing []string
	for _, tech := range contract.RequiredTechnologies {
		if !mentionsTechnology(combined, tech) {
			missing = append(missing, tech)
		}
	}
	sort.Strings(missing)
	return uniqueStrings(missing)
}

func (c *ProjectContract) IsEmpty() bool {
	if c == nil {
		return true
	}
	return len(c.RequiredTechnologies) == 0 && len(c.ForbiddenTechnologies) == 0
}

func contractFromObjectiveJSON(raw string) *ProjectContract {
	document, err := parseObjectiveJSON(raw)
	if err != nil {
		return &ProjectContract{}
	}
	return contractFromObjectiveDoc(document)
}

func contractFromObjectiveDoc(document map[string]interface{}) *ProjectContract {
	raw, ok := document[technicalContractKey]
	if !ok || raw == nil {
		return &ProjectContract{}
	}
	data, err := json.Marshal(raw)
	if err != nil {
		return &ProjectContract{}
	}

	var contract ProjectContract
	if err := json.Unmarshal(data, &contract); err != nil {
		return &ProjectContract{}
	}
	contract.RequiredTechnologies = uniqueStrings(contract.RequiredTechnologies)
	contract.ForbiddenTechnologies = uniqueStrings(contract.ForbiddenTechnologies)
	return &contract
}

func parseObjectiveJSON(raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}

	document := make(map[string]interface{})
	if err := json.Unmarshal([]byte(raw), &document); err != nil {
		return nil, err
	}
	return document, nil
}

func marshalObjectiveJSON(document map[string]interface{}) (string, error) {
	data, err := json.Marshal(document)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func mergeContracts(base, override *ProjectContract) *ProjectContract {
	if base == nil {
		base = &ProjectContract{}
	}
	if override == nil || override.IsEmpty() {
		return &ProjectContract{
			RequiredTechnologies:  uniqueStrings(base.RequiredTechnologies),
			ForbiddenTechnologies: uniqueStrings(base.ForbiddenTechnologies),
		}
	}

	required := make(map[string]struct{}, len(base.RequiredTechnologies)+len(override.RequiredTechnologies))
	for _, item := range append(base.RequiredTechnologies, override.RequiredTechnologies...) {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" {
			continue
		}
		required[item] = struct{}{}
	}

	forbidden := make(map[string]struct{}, len(base.ForbiddenTechnologies)+len(override.ForbiddenTechnologies))
	for _, item := range append(base.ForbiddenTechnologies, override.ForbiddenTechnologies...) {
		item = strings.TrimSpace(strings.ToLower(item))
		if item == "" {
			continue
		}
		forbidden[item] = struct{}{}
	}

	return &ProjectContract{
		RequiredTechnologies:  sortedKeys(required),
		ForbiddenTechnologies: sortedKeys(forbidden),
	}
}

func sameContract(left, right *ProjectContract) bool {
	if left == nil {
		left = &ProjectContract{}
	}
	if right == nil {
		right = &ProjectContract{}
	}
	if strings.Join(uniqueStrings(left.RequiredTechnologies), ",") != strings.Join(uniqueStrings(right.RequiredTechnologies), ",") {
		return false
	}
	return strings.Join(uniqueStrings(left.ForbiddenTechnologies), ",") == strings.Join(uniqueStrings(right.ForbiddenTechnologies), ",")
}

func matchesRequiredDirective(text string, aliases []string) bool {
	for _, alias := range aliases {
		alias = strings.TrimSpace(strings.ToLower(alias))
		if alias == "" {
			continue
		}
		for _, pattern := range []string{
			"必须使用" + alias,
			"必须使用 " + alias,
			"必须用" + alias,
			"必须用 " + alias,
			"需要使用" + alias,
			"需要使用 " + alias,
			"需要用" + alias,
			"需要用 " + alias,
			"must use " + alias,
			"required to use " + alias,
			"need to use " + alias,
		} {
			if strings.Contains(text, pattern) {
				return true
			}
		}
	}
	return false
}

func matchesForbiddenDirective(text string, aliases []string) bool {
	for _, alias := range aliases {
		alias = strings.TrimSpace(strings.ToLower(alias))
		if alias == "" {
			continue
		}
		for _, pattern := range []string{
			"不要" + alias,
			"不要 " + alias,
			"不要用" + alias,
			"不要用 " + alias,
			"不能用" + alias,
			"不能用 " + alias,
			"禁止" + alias,
			"禁止 " + alias,
			"禁止使用" + alias,
			"禁止使用 " + alias,
			"不用" + alias,
			"不用 " + alias,
			"不使用" + alias,
			"不使用 " + alias,
			"no " + alias,
			"without " + alias,
			"do not use " + alias,
			"don't use " + alias,
			"must not use " + alias,
		} {
			if strings.Contains(text, pattern) {
				return true
			}
		}
	}
	return false
}

func mentionsTechnology(text, canonical string) bool {
	for _, tech := range technologyCatalog {
		if tech.Canonical != canonical {
			continue
		}
		for _, alias := range tech.Aliases {
			if strings.Contains(text, strings.ToLower(alias)) {
				return true
			}
		}
	}
	return false
}

func sortedKeys(items map[string]struct{}) []string {
	keys := make([]string, 0, len(items))
	for key := range items {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}
