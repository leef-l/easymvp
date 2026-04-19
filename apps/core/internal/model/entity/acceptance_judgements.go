// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

// AcceptanceJudgements is the golang structure for table acceptance_judgements.
type AcceptanceJudgements struct {
	Id              string `json:"id"              orm:"id"                ` //
	ProjectId       string `json:"projectId"       orm:"project_id"        ` //
	AcceptanceRunId string `json:"acceptanceRunId" orm:"acceptance_run_id" ` //
	JudgementKind   string `json:"judgementKind"   orm:"judgement_kind"    ` //
	JudgementResult string `json:"judgementResult" orm:"judgement_result"  ` //
	Summary         string `json:"summary"         orm:"summary"           ` //
	DetailJson      string `json:"detailJson"      orm:"detail_json"       ` //
	CreatedAt       string `json:"createdAt"       orm:"created_at"        ` //
}
