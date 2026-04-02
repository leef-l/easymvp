package response

import (
	"net/http"

	"github.com/gogf/gf/v2/net/ghttp"
)

// R 统一响应结构
type R struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Ok 成功响应
func Ok(r *ghttp.Request, data ...interface{}) {
	resp := R{Code: 0, Message: "ok"}
	if len(data) > 0 {
		resp.Data = data[0]
	}
	r.Response.WriteJsonExit(resp)
}

// OkMsg 成功响应（自定义消息）
func OkMsg(r *ghttp.Request, msg string) {
	r.Response.WriteJsonExit(R{Code: 0, Message: msg})
}

// Fail 失败响应
func Fail(r *ghttp.Request, msg string) {
	r.Response.WriteJsonExit(R{Code: -1, Message: msg})
}

// FailCode 失败响应（自定义 code）
func FailCode(r *ghttp.Request, code int, msg string) {
	r.Response.WriteJsonExit(R{Code: code, Message: msg})
}

// Unauthorized 401 未授权
func Unauthorized(r *ghttp.Request, msg ...string) {
	m := "未登录或登录已过期"
	if len(msg) > 0 {
		m = msg[0]
	}
	r.Response.Status = http.StatusUnauthorized
	r.Response.WriteJsonExit(R{Code: 401, Message: m})
}

// Forbidden 403 无权限
func Forbidden(r *ghttp.Request, msg ...string) {
	m := "无访问权限"
	if len(msg) > 0 {
		m = msg[0]
	}
	r.Response.Status = http.StatusForbidden
	r.Response.WriteJsonExit(R{Code: 403, Message: m})
}
