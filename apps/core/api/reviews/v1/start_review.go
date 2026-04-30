package v1

import "github.com/gogf/gf/v2/frame/g"

type StartReviewReq struct {
	g.Meta    `path:"/api/v3/reviews/start" method:"post" tags:"Reviews" summary:"Trigger a single design review round"`
	DesignID  string `json:"design_id" v:"required"`
	ProjectID string `json:"project_id" v:"required"`
}

type StartReviewRes struct {
	ReviewID string   `json:"review_id"`
	Passed   bool     `json:"passed"`
	Score    int      `json:"score"`
	Issues   []string `json:"issues"`
	Round    int      `json:"round"`
}
