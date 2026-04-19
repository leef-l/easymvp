package service

import (
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
)

const (
	runtimeErrorCodeUnavailable = "RUN_001"
	runtimeErrorCodeCreateRun   = "RUN_002"
	runtimeErrorCodeSyncRun     = "RUN_003"
	runtimeErrorCodeResumeRun   = "RUN_004"
)

func wrapRuntimeError(code string, summary string, err error) error {
	summary = strings.TrimSpace(summary)
	if err == nil {
		return gerror.Newf("%s: %s", code, summary)
	}
	return gerror.Wrap(err, code+": "+summary)
}

func runEventTypeForStatus(runStatus string) string {
	switch runStatus {
	case "run_pending":
		return "run.accepted"
	case "run_active":
		return "run.started"
	case "run_succeeded":
		return "run.completed"
	case "run_failed":
		return "run.failed"
	case "run_unsupported":
		return "run.unsupported"
	case "run_denied":
		return "run.denied"
	case "run_cancelled":
		return "run.cancelled"
	default:
		return "run.updated"
	}
}

func runEventSummaryForStatus(runStatus string) string {
	switch runStatus {
	case "run_pending":
		return "brain run accepted by runtime adapter"
	case "run_active":
		return "brain run is active in runtime"
	case "run_succeeded":
		return "brain run completed successfully"
	case "run_failed":
		return "brain run failed in runtime"
	case "run_unsupported":
		return "brain run hit an unsupported runtime capability"
	case "run_denied":
		return "brain run was denied by runtime policy"
	case "run_cancelled":
		return "brain run was cancelled"
	default:
		return "brain run status updated"
	}
}
