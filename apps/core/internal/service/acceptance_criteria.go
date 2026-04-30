package service

// AcceptanceCriteria defines the multi-layer acceptance criteria structure
// for MACCS Phase 5 systematic verification.
type AcceptanceCriteria struct {
	Layers []AcceptanceLayer
}

// AcceptanceLayer represents a single acceptance verification layer.
type AcceptanceLayer struct {
	Name     string
	Required bool
	Criteria []Criterion
}

// Criterion represents a single acceptance check within a layer.
type Criterion struct {
	Name   string
	Weight int
}

// AcceptanceVerdict is the final evaluation result from acceptance criteria.
type AcceptanceVerdict struct {
	Passed     bool
	TotalScore int
	MaxScore   int
	Summary    string
}

// DefaultAcceptanceCriteria returns the default multi-layer acceptance criteria
// for the given project type. Supported types: web_app, game, api, data_pipeline, general.
func DefaultAcceptanceCriteria(projectType string) AcceptanceCriteria {
	switch projectType {
	case "web_app":
		return webAppAcceptanceCriteria()
	case "game":
		return gameAcceptanceCriteria()
	case "api":
		return apiAcceptanceCriteria()
	case "data_pipeline":
		return dataPipelineAcceptanceCriteria()
	default:
		return generalAcceptanceCriteria()
	}
}

// EvaluateAcceptanceResults evaluates the collected layer results against the
// defined acceptance criteria and returns the overall verdict.
func EvaluateAcceptanceResults(criteria AcceptanceCriteria, results []LayerResult) AcceptanceVerdict {
	totalScore := 0
	maxScore := 0

	for _, layer := range criteria.Layers {
		layerWeight := 0
		for _, c := range layer.Criteria {
			layerWeight += c.Weight
		}
		maxScore += layerWeight

		lr := findLayerResult(layer.Name, results)
		if lr == nil {
			if layer.Required {
				return AcceptanceVerdict{
					Passed:     false,
					TotalScore: totalScore,
					MaxScore:   maxScore,
					Summary:    "Required layer " + layer.Name + " has no results",
				}
			}
			continue
		}

		if layer.Required && !lr.Passed {
			return AcceptanceVerdict{
				Passed:     false,
				TotalScore: totalScore + lr.Score,
				MaxScore:   maxScore,
				Summary:    "Required layer " + layer.Name + " failed",
			}
		}
		totalScore += lr.Score
	}

	passed := true
	summary := "All acceptance criteria passed"
	if maxScore > 0 {
		ratio := float64(totalScore) / float64(maxScore)
		if ratio < 0.7 {
			passed = false
			summary = "Acceptance score below threshold (70%)"
		}
	}

	return AcceptanceVerdict{
		Passed:     passed,
		TotalScore: totalScore,
		MaxScore:   maxScore,
		Summary:    summary,
	}
}

func findLayerResult(name string, results []LayerResult) *LayerResult {
	for i := range results {
		if results[i].LayerName == name {
			return &results[i]
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// Per-project-type criteria definitions
// ---------------------------------------------------------------------------

func webAppAcceptanceCriteria() AcceptanceCriteria {
	return AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic_coverage", Weight: 20},
					{Name: "model_validation", Weight: 10},
					{Name: "utility_functions", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "api_endpoint_integration", Weight: 20},
					{Name: "database_operations", Weight: 15},
					{Name: "auth_flow", Weight: 15},
				},
			},
			{
				Name:     "e2e_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "user_sign_in_flow", Weight: 15},
					{Name: "primary_user_journey", Weight: 20},
					{Name: "admin_operations", Weight: 10},
				},
			},
			{
				Name:     "security_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "xss_prevention", Weight: 15},
					{Name: "csrf_protection", Weight: 10},
					{Name: "input_sanitization", Weight: 10},
					{Name: "auth_bypass_check", Weight: 15},
				},
			},
			{
				Name:     "performance_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "page_load_time", Weight: 10},
					{Name: "api_response_time", Weight: 10},
					{Name: "concurrent_users", Weight: 10},
				},
			},
		},
	}
}

func gameAcceptanceCriteria() AcceptanceCriteria {
	return AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "game_logic_coverage", Weight: 20},
					{Name: "physics_calculations", Weight: 15},
					{Name: "state_management", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "game_loop_integration", Weight: 20},
					{Name: "asset_loading", Weight: 10},
					{Name: "save_load_cycle", Weight: 15},
				},
			},
			{
				Name:     "e2e_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "launch_to_gameplay", Weight: 15},
					{Name: "core_game_loop", Weight: 20},
					{Name: "pause_resume_flow", Weight: 10},
				},
			},
			{
				Name:     "security_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "input_validation", Weight: 10},
					{Name: "save_file_integrity", Weight: 10},
				},
			},
			{
				Name:     "performance_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "frame_rate_stability", Weight: 15},
					{Name: "memory_usage", Weight: 10},
					{Name: "load_time", Weight: 10},
				},
			},
		},
	}
}

func apiAcceptanceCriteria() AcceptanceCriteria {
	return AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "handler_logic", Weight: 20},
					{Name: "request_validation", Weight: 15},
					{Name: "response_serialization", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "endpoint_contracts", Weight: 20},
					{Name: "database_operations", Weight: 15},
					{Name: "external_service_calls", Weight: 10},
				},
			},
			{
				Name:     "e2e_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "full_api_workflow", Weight: 15},
					{Name: "error_handling_paths", Weight: 10},
				},
			},
			{
				Name:     "security_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "auth_token_validation", Weight: 15},
					{Name: "rate_limiting", Weight: 10},
					{Name: "injection_prevention", Weight: 15},
					{Name: "data_exposure_check", Weight: 10},
				},
			},
			{
				Name:     "performance_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "response_latency", Weight: 15},
					{Name: "throughput_under_load", Weight: 10},
					{Name: "connection_pool_stability", Weight: 10},
				},
			},
		},
	}
}

func dataPipelineAcceptanceCriteria() AcceptanceCriteria {
	return AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "transform_logic", Weight: 20},
					{Name: "data_validation_rules", Weight: 15},
					{Name: "edge_case_handling", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "pipeline_stage_handoff", Weight: 20},
					{Name: "source_sink_connectivity", Weight: 15},
					{Name: "schema_compatibility", Weight: 10},
				},
			},
			{
				Name:     "e2e_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "full_pipeline_run", Weight: 20},
					{Name: "data_integrity_check", Weight: 15},
				},
			},
			{
				Name:     "security_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "credential_handling", Weight: 10},
					{Name: "data_encryption_at_rest", Weight: 10},
				},
			},
			{
				Name:     "performance_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "throughput_volume", Weight: 15},
					{Name: "memory_efficiency", Weight: 10},
					{Name: "backpressure_handling", Weight: 10},
				},
			},
		},
	}
}

func generalAcceptanceCriteria() AcceptanceCriteria {
	return AcceptanceCriteria{
		Layers: []AcceptanceLayer{
			{
				Name:     "unit_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "core_logic_coverage", Weight: 20},
					{Name: "edge_case_handling", Weight: 10},
				},
			},
			{
				Name:     "integration_test",
				Required: true,
				Criteria: []Criterion{
					{Name: "component_integration", Weight: 20},
					{Name: "data_flow_validation", Weight: 10},
				},
			},
			{
				Name:     "e2e_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "primary_workflow", Weight: 15},
					{Name: "error_recovery", Weight: 10},
				},
			},
			{
				Name:     "security_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "input_validation", Weight: 10},
					{Name: "access_control", Weight: 10},
				},
			},
			{
				Name:     "performance_test",
				Required: false,
				Criteria: []Criterion{
					{Name: "baseline_performance", Weight: 10},
					{Name: "resource_usage", Weight: 10},
				},
			},
		},
	}
}
