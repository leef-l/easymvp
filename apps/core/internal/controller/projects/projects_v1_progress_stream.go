package projects

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	v1 "github.com/leef-l/easymvp/apps/core/api/projects/v1"
	"github.com/leef-l/easymvp/apps/core/internal/service"
)

// ProjectProgressStream 处理 GET /api/v3/projects/{id}/progress-stream SSE 端点。
// 它将控制权交给服务层，后者会持续推送 MACCS 工作流进度事件，
// 直到客户端断开连接为止。
func (c *ControllerV1) ProjectProgressStream(ctx context.Context, req *v1.ProjectProgressStreamReq) (res *v1.ProjectProgressStreamRes, err error) {
	if err = service.Projects().ProjectProgressStream(ctx, req); err != nil {
		return nil, err
	}
	g.RequestFromCtx(ctx).ExitAll()
	return nil, nil
}
