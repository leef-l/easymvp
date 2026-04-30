package v1

import "github.com/gogf/gf/v2/frame/g"

type RunReviewLoopReq struct {
	g.Meta    `path:"/api/v3/reviews/loop" method:"post" tags:"Reviews" summary:"Run automated review-fix loop until convergence or max rounds"`
	DesignID  string `json:"design_id" v:"required"`
	ProjectID string `json:"project_id" v:"required"`
	MaxRounds int    `json:"max_rounds"`
}

type RunReviewLoopRes struct {
	Passed        bool   `json:"passed"`
	Rounds        int    `json:"rounds"`
	Reason        string `json:"reason"`
	FinalReviewID string `json:"final_review_id"`
}
