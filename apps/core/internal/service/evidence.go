package service

import (
	"context"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	evidencev1 "github.com/leef-l/easymvp/apps/core/api/evidence/v1"
)

type IEvidence interface {
	CollectProjectEvidence(ctx context.Context, projectID string) (*evidencev1.CollectEvidenceRes, error)
}

var localEvidence IEvidence = (*sEvidence)(nil)

type sEvidence struct{}

func Evidence() IEvidence {
	if localEvidence == nil {
		localEvidence = &sEvidence{}
	}
	return localEvidence
}

func (s *sEvidence) CollectProjectEvidence(ctx context.Context, projectID string) (*evidencev1.CollectEvidenceRes, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, gerror.New("project id is required")
	}

	count, err := collectProjectEvidence(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return &evidencev1.CollectEvidenceRes{
		CommandID:      newResourceID("cmd"),
		Accepted:       true,
		ResourceID:     projectID,
		NextAction:     "refresh_acceptance_view",
		InsertedCount:  count,
	}, nil
}
