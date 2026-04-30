package v1

import "github.com/gogf/gf/v2/frame/g"

// InterveneReq allows human operators to override or abort an ongoing design review.
// action: "override_approve" — force pass; "abort" — stop the review; "restart" — reset and re-review.
type InterveneReq struct {
	g.Meta   `path:"/api/v3/reviews/intervene" method:"post" tags:"Reviews" summary:"Human intervention: override, abort, or restart a design review"`
	DesignID string `json:"design_id" v:"required"`
	Action   string `json:"action" v:"required|in:override_approve,abort,restart"`
	Reason   string `json:"reason"`
}

type InterveneRes struct {
	DesignID string `json:"design_id"`
	Action   string `json:"action"`
	Applied  bool   `json:"applied"`
	Message  string `json:"message"`
}
