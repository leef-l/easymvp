package autonomy

import "easymvp/app/mvp/internal/consts"

// DecisionMeta 决策元数据：四个一等公民。
type DecisionMeta struct {
	Confidence          float64 `json:"confidence"`
	EvidenceSufficiency float64 `json:"evidenceSufficiency"`
	Reversibility       string  `json:"reversibility"` // full / partial / none
	BlastRadius         string  `json:"blastRadius"`   // task / batch / stage / workflow / project
}

// ValidationResult 决策元数据校验结果。
type ValidationResult struct {
	Allowed   bool   `json:"allowed"`
	UpgradeTo string `json:"upgradeTo,omitempty"`
	Reason    string `json:"reason,omitempty"`
}

var blastRadiusMinLevel = map[string]string{
	"task":     consts.DecisionLevelA,
	"batch":    consts.DecisionLevelA,
	"stage":    consts.DecisionLevelB,
	"workflow": consts.DecisionLevelB,
	"project":  consts.DecisionLevelC,
}

const (
	EvidenceThresholdAuto   = 0.7
	EvidenceThresholdAssist = 0.4
)

// Validate 校验当前决策元数据是否允许按请求等级自动执行。
func (m *DecisionMeta) Validate(requestedLevel string) *ValidationResult {
	result := &ValidationResult{Allowed: true}
	if m == nil {
		return result
	}
	if m.EvidenceSufficiency < EvidenceThresholdAssist {
		result.Allowed = false
		result.Reason = "evidence_insufficient"
		return result
	}

	if minLevel, ok := blastRadiusMinLevel[m.BlastRadius]; ok && levelToInt(requestedLevel) < levelToInt(minLevel) {
		result.UpgradeTo = minLevel
		result.Reason = "blast_radius_requires_upgrade"
	}

	if m.Confidence < 0.5 && m.Reversibility == "none" {
		result.UpgradeTo = consts.DecisionLevelC
		result.Reason = "low_confidence_irreversible"
	}

	return result
}

func levelToInt(level string) int {
	switch level {
	case consts.DecisionLevelA:
		return 1
	case consts.DecisionLevelB:
		return 2
	case consts.DecisionLevelC:
		return 3
	default:
		return 99
	}
}
