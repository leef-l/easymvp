package message

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/dao"
	"easymvp/app/mvp/internal/middleware"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterMessage(New())
}

func New() *sMessage {
	return &sMessage{}
}

type sMessage struct{}

// Create 创建MVP消息表
func (s *sMessage) Create(ctx context.Context, in *model.MessageCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpMessage.Ctx(ctx).Data(g.Map{
		dao.MvpMessage.Columns().Id:        id,
		dao.MvpMessage.Columns().ConversationId: in.ConversationID,
		dao.MvpMessage.Columns().Role: in.Role,
		dao.MvpMessage.Columns().MessageType: resolveMessageType(in.MessageType, in.Role, in.Status),
		dao.MvpMessage.Columns().Content: in.Content,
		dao.MvpMessage.Columns().ModelId: in.ModelID,
		dao.MvpMessage.Columns().TokenUsage: in.TokenUsage,
		dao.MvpMessage.Columns().Status: in.Status,
		dao.MvpMessage.Columns().CreatedBy: middleware.GetUserID(ctx),
		dao.MvpMessage.Columns().DeptId: middleware.GetDeptID(ctx),
		dao.MvpMessage.Columns().CreatedAt: gtime.Now(),
		dao.MvpMessage.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新MVP消息表
func (s *sMessage) Update(ctx context.Context, in *model.MessageUpdateInput) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().DeletedAt, nil), in.ID, dao.MvpMessage.Columns().Id, dao.MvpMessage.Columns().CreatedBy); err != nil {
		return err
	}
	data := g.Map{
		dao.MvpMessage.Columns().ConversationId: in.ConversationID,
		dao.MvpMessage.Columns().Role: in.Role,
		dao.MvpMessage.Columns().MessageType: resolveMessageType(in.MessageType, in.Role, in.Status),
		dao.MvpMessage.Columns().Content: in.Content,
		dao.MvpMessage.Columns().ModelId: in.ModelID,
		dao.MvpMessage.Columns().TokenUsage: in.TokenUsage,
		dao.MvpMessage.Columns().Status: in.Status,
		dao.MvpMessage.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除MVP消息表
func (s *sMessage) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	if err := middleware.CheckOwnership(ctx, dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().DeletedAt, nil), id, dao.MvpMessage.Columns().Id, dao.MvpMessage.Columns().CreatedBy); err != nil {
		return err
	}
	_, err := dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().Id, id).Data(g.Map{
		dao.MvpMessage.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除MVP消息表
func (s *sMessage) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	m := middleware.ApplyDataScope(ctx, dao.MvpMessage.Ctx(ctx), dao.MvpMessage.Columns().CreatedBy)
	_, err := m.WhereIn(dao.MvpMessage.Columns().Id, ids).Data(g.Map{
		dao.MvpMessage.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取MVP消息表详情
func (s *sMessage) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.MessageDetailOutput, err error) {
	// 权限校验
	if err = middleware.CheckOwnership(ctx,
		dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().DeletedAt, nil),
		id, dao.MvpMessage.Columns().Id, dao.MvpMessage.Columns().CreatedBy); err != nil {
		return nil, err
	}

	out = &model.MessageDetailOutput{}
	err = dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().Id, id).Where(dao.MvpMessage.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	if out.ID == 0 {
		return nil, fmt.Errorf("记录不存在")
	}
	// 查询对话ID关联显示
	if out.ConversationID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_conversation").Where("id", out.ConversationID).Where("deleted_at", nil).Value("title")
		if err == nil {
			out.ConversationTitle = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sMessage) applyListFilter(ctx context.Context, in *model.MessageListInput) *gdb.Model {
	m := dao.MvpMessage.Ctx(ctx).Where(dao.MvpMessage.Columns().DeletedAt, nil)
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpMessage.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpMessage.Columns().CreatedAt, in.EndTime)
	}
	if in.MessageType != "" {
		m = m.Where(dao.MvpMessage.Columns().MessageType, in.MessageType)
	}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m, dao.MvpMessage.Columns().CreatedBy, dao.MvpMessage.Columns().DeptId)
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sMessage) fillRefFields(ctx context.Context, list []*model.MessageListOutput) {
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.ConversationID != 0 {
				idSet[int64(item.ConversationID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("mvp_conversation").
				Fields("id", "title").
				Where("deleted_at", nil).
				WhereIn("id", ids).
				All()
			if err == nil {
				refMap := make(map[int64]string, len(rows))
				for _, row := range rows {
					refMap[row["id"].Int64()] = row["title"].String()
				}
				for _, item := range list {
					if val, ok := refMap[int64(item.ConversationID)]; ok {
						item.ConversationTitle = val
					}
				}
			}
		}
	}
}

// List 获取MVP消息表列表
func (s *sMessage) List(ctx context.Context, in *model.MessageListInput) (list []*model.MessageListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序（白名单防止 SQL 注入）
	allowedOrderBy := map[string]bool{
		"id": true, "role": true, "message_type": true, "status": true, "created_at": true, "updated_at": true,
	}
	if in.OrderBy != "" && allowedOrderBy[in.OrderBy] {
		if in.OrderDir == "desc" {
			m = m.OrderDesc(in.OrderBy)
		} else {
			m = m.OrderAsc(in.OrderBy)
		}
	} else {
		m = m.OrderAsc(dao.MvpMessage.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出MVP消息表（不分页）
func (s *sMessage) Export(ctx context.Context, in *model.MessageListInput) (list []*model.MessageListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpMessage.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}



// BatchUpdate 批量编辑MVP消息表
func (s *sMessage) BatchUpdate(ctx context.Context, in *model.MessageBatchUpdateInput) error {
	data := g.Map{
		dao.MvpMessage.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.MvpMessage.Columns().Status] = *in.Status
	}
	m := middleware.ApplyDataScope(ctx, dao.MvpMessage.Ctx(ctx), dao.MvpMessage.Columns().CreatedBy)
	_, err := m.WhereIn(dao.MvpMessage.Columns().Id, in.IDs).Data(data).Update()
	return err
}


// Import 导入MVP消息表
func (s *sMessage) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
	f, err := file.Open()
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// 跳过表头
	if _, err = reader.Read(); err != nil {
		return 0, 0, fmt.Errorf("读取CSV表头失败: %w", err)
	}

	for {
		record, readErr := reader.Read()
		if readErr != nil {
			break
		}
		if len(record) == 0 {
			continue
		}
		// 逐行插入
		id := snowflake.Generate()
		data := g.Map{
			dao.MvpMessage.Columns().Id:        id,
			dao.MvpMessage.Columns().CreatedBy: middleware.GetUserID(ctx),
			dao.MvpMessage.Columns().DeptId:    middleware.GetDeptID(ctx),
			dao.MvpMessage.Columns().CreatedAt: gtime.Now(),
			dao.MvpMessage.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpMessage.Columns().ConversationId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().Role] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().MessageType] = resolveMessageType(record[idx], fmt.Sprint(data[dao.MvpMessage.Columns().Role]), fmt.Sprint(data[dao.MvpMessage.Columns().Status]))
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().Content] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().ModelId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().TokenUsage] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpMessage.Columns().Status] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpMessage.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

func resolveMessageType(messageType string, role string, status string) string {
	if messageType != "" {
		return messageType
	}
	if status == "failed" {
		return model.MessageTypePoison
	}
	switch role {
	case "user":
		return model.MessageTypeChatUser
	case "assistant":
		return model.MessageTypeChatReply
	case "system":
		return model.MessageTypeSystemNotice
	default:
		return model.MessageTypeGeneral
	}
}

