package engine

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/ai/internal/middleware"
	"easymvp/app/ai/internal/model"
	"easymvp/app/ai/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterEngine(New())
}

func New() *sEngine {
	return &sEngine{}
}

type sEngine struct{}

func (s *sEngine) List(ctx context.Context) (list []*model.EngineListOutput, err error) {
	err = g.DB().Ctx(ctx).Model("ai_engine e").
		LeftJoin("ai_engine_config c", "c.engine_code = e.code AND c.deleted_at IS NULL").
		Fields("e.id, e.code, e.name, e.description, e.status, COALESCE(c.status, 0) AS config_status, COALESCE(c.default_model_id, 0) AS default_model_id, e.created_at, e.updated_at").
		Where("e.deleted_at IS NULL").
		OrderAsc("e.sort").
		Scan(&list)
	return
}

func (s *sEngine) Detail(ctx context.Context, engineCode string) (out *model.EngineDetailOutput, err error) {
	out = &model.EngineDetailOutput{}
	record, err := g.DB().Ctx(ctx).Model("ai_engine e").
		LeftJoin("ai_engine_config c", "c.engine_code = e.code AND c.deleted_at IS NULL").
		Fields("e.id, e.code AS engine_code, e.name, e.description, c.base_url, COALESCE(c.default_model_id, 0) AS default_model_id, COALESCE(c.timeout_seconds, 600) AS timeout_seconds, COALESCE(c.max_steps, 20) AS max_steps, c.workspace_root, c.command_template, c.callback_url, c.callback_secret, c.extra_config, e.status, COALESCE(c.status, 0) AS config_status, c.api_key").
		Where("e.code", engineCode).
		Where("e.deleted_at IS NULL").
		One()
	if err != nil {
		return nil, err
	}
	if record.IsEmpty() {
		return nil, fmt.Errorf("执行引擎不存在")
	}
	out.ID = snowflake.JsonInt64(record["id"].Int64())
	out.EngineCode = record["engine_code"].String()
	out.Name = record["name"].String()
	out.Description = record["description"].String()
	out.BaseURL = record["base_url"].String()
	out.DefaultModelID = snowflake.JsonInt64(record["default_model_id"].Int64())
	out.TimeoutSeconds = record["timeout_seconds"].Int()
	out.MaxSteps = record["max_steps"].Int()
	out.WorkspaceRoot = record["workspace_root"].String()
	out.CommandTemplate = record["command_template"].String()
	out.CallbackURL = record["callback_url"].String()
	out.CallbackSecret = record["callback_secret"].String()
	out.ExtraConfig = record["extra_config"].String()
	out.Status = record["status"].Int()
	out.ConfigStatus = record["config_status"].Int()
	out.APIKeyMasked = maskSecret(record["api_key"].String())
	return
}

func (s *sEngine) Update(ctx context.Context, in *model.EngineUpdateInput) error {
	now := gtime.Now()
	count, err := g.DB().Ctx(ctx).Model("ai_engine_config").
		Where("engine_code", in.EngineCode).
		Where("deleted_at IS NULL").
		Count()
	if err != nil {
		return err
	}

	data := g.Map{
		"engine_code":      in.EngineCode,
		"base_url":         in.BaseURL,
		"api_key":          in.APIKey,
		"default_model_id": in.DefaultModelID,
		"timeout_seconds":  in.TimeoutSeconds,
		"max_steps":        in.MaxSteps,
		"workspace_root":   in.WorkspaceRoot,
		"command_template": in.CommandTemplate,
		"callback_url":     in.CallbackURL,
		"callback_secret":  in.CallbackSecret,
		"extra_config":     in.ExtraConfig,
		"status":           in.Status,
		"dept_id":          middleware.GetDeptID(ctx),
		"updated_at":       now,
	}

	if count > 0 {
		_, err = g.DB().Ctx(ctx).Model("ai_engine_config").
			Where("engine_code", in.EngineCode).
			Where("deleted_at IS NULL").
			Data(data).
			Update()
		return err
	}

	data["id"] = snowflake.Generate()
	data["created_by"] = middleware.GetUserID(ctx)
	data["created_at"] = now
	_, err = g.DB().Ctx(ctx).Model("ai_engine_config").Data(data).Insert()
	return err
}

func (s *sEngine) TestConnection(ctx context.Context, engineCode string) (*model.EngineTestOutput, error) {
	detail, err := s.Detail(ctx, engineCode)
	if err != nil {
		return nil, err
	}

	switch engineCode {
	case "aider":
		if _, err = exec.LookPath("aider"); err == nil {
			return &model.EngineTestOutput{Success: true, Message: "aider 可执行文件可用"}, nil
		}
		if _, err = exec.LookPath("uv"); err == nil {
			return &model.EngineTestOutput{Success: true, Message: "本机未安装 aider，将通过 uv 自动安装/执行"}, nil
		}
		if _, err = exec.LookPath("docker"); err == nil {
			return &model.EngineTestOutput{Success: true, Message: "本机未安装 aider，将使用 Docker 镜像执行"}, nil
		}
		return &model.EngineTestOutput{Success: false, Message: "未找到 aider 可执行文件，且 uv/docker 都不可用"}, nil
	case "openhands":
		if strings.TrimSpace(detail.CommandTemplate) != "" {
			if _, err = exec.LookPath("docker"); err == nil {
				return &model.EngineTestOutput{Success: true, Message: "OpenHands 已配置命令模板，将通过 Docker 执行"}, nil
			}
			return &model.EngineTestOutput{Success: true, Message: "OpenHands 已配置命令模板"}, nil
		}
		if _, err = exec.LookPath("openhands"); err == nil {
			return &model.EngineTestOutput{Success: true, Message: "OpenHands CLI 可用"}, nil
		}
		if _, err = exec.LookPath("uv"); err == nil {
			return &model.EngineTestOutput{Success: true, Message: "OpenHands 将通过 uv 自动安装/执行"}, nil
		}
		if detail.BaseURL == "" {
			return &model.EngineTestOutput{Success: false, Message: "OpenHands Base URL 未配置"}, nil
		}
		client := &http.Client{Timeout: 5 * time.Second}
		target := strings.TrimRight(detail.BaseURL, "/")
		resp, err := client.Get(target)
		if err != nil {
			return &model.EngineTestOutput{Success: false, Message: "OpenHands 连接失败: " + err.Error()}, nil
		}
		defer resp.Body.Close()
		return &model.EngineTestOutput{Success: resp.StatusCode < 500, Message: fmt.Sprintf("OpenHands 响应状态: %d", resp.StatusCode)}, nil
	default:
		return &model.EngineTestOutput{Success: false, Message: "暂不支持该引擎测试"}, nil
	}
}

func maskSecret(secret string) string {
	if secret == "" {
		return ""
	}
	if len(secret) <= 8 {
		return "****"
	}
	return secret[:4] + "****" + secret[len(secret)-4:]
}
