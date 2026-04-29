package entity

// VerificationResults is the golang structure for table verification_results.
type VerificationResults struct {
	Id                     string `json:"id" orm:"id"`
	ProjectId              string `json:"project_id" orm:"project_id"`
	AcceptanceRunId        string `json:"acceptance_run_id" orm:"acceptance_run_id"`
	Status                 string `json:"status" orm:"status"`
	Decision               string `json:"decision" orm:"decision"`
	Completed              int    `json:"completed" orm:"completed"`
	Summary                string `json:"summary" orm:"summary"`
	PreferredChannel       string `json:"preferred_channel" orm:"preferred_channel"`
	RequiredChecksJson     string `json:"required_checks_json" orm:"required_checks_json"`
	RequiredEvidenceJson   string `json:"required_evidence_json" orm:"required_evidence_json"`
	MissingEvidenceJson    string `json:"missing_evidence_json" orm:"missing_evidence_json"`
	FailedChecksJson       string `json:"failed_checks_json" orm:"failed_checks_json"`
	VerificationContractJson string `json:"verification_contract_json" orm:"verification_contract_json"`
	SourceRunId            string `json:"source_run_id" orm:"source_run_id"`
	ChannelAvailable       int    `json:"channel_available" orm:"channel_available"`
	EnvironmentAvailable   int    `json:"environment_available" orm:"environment_available"`
	CreatedAt              string `json:"created_at" orm:"created_at"`
	UpdatedAt              string `json:"updated_at" orm:"updated_at"`
}
