package v1

import (
	"github.com/gogf/gf/v2/frame/g"
)

type HealthReq struct {
	g.Meta `path:"/api/v3/system/healthz" tags:"System" method:"get" summary:"System health check"`
}

type HealthRes struct {
	Status    string `json:"status"`
	Service   string `json:"service"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}
