package v1

import "github.com/gogf/gf/v2/frame/g"

// ProjectProgressStreamReq 订阅指定项目的 MACCS 闭环工作流实时进度事件（SSE）。
type ProjectProgressStreamReq struct {
	g.Meta `path:"/api/v3/projects/{id}/progress-stream" tags:"Projects" method:"get" summary:"MACCS workflow progress event stream (SSE)"`
	Id     string `json:"id" in:"path" v:"required"`
}

// ProjectProgressStreamRes 是 SSE 端点的占位响应结构（实际响应通过 raw writer 写入）。
type ProjectProgressStreamRes struct{}
