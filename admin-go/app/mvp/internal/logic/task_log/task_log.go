package tasklog

import (
	"context"
	"encoding/csv"
	"fmt"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"

	"easymvp/app/mvp/internal/dao"
	"easymvp/app/mvp/internal/model"
	"easymvp/app/mvp/internal/service"
	"easymvp/utility/snowflake"
)

func init() {
	service.RegisterTaskLog(New())
}

func New() *sTaskLog {
	return &sTaskLog{}
}

type sTaskLog struct{}

// Create 创建任务日志表
func (s *sTaskLog) Create(ctx context.Context, in *model.TaskLogCreateInput) error {
	id := snowflake.Generate()
	_, err := dao.MvpTaskLog.Ctx(ctx).Data(g.Map{
		dao.MvpTaskLog.Columns().Id:        id,
		dao.MvpTaskLog.Columns().TaskId: in.TaskID,
		dao.MvpTaskLog.Columns().Action: in.Action,
		dao.MvpTaskLog.Columns().FromStatus: in.FromStatus,
		dao.MvpTaskLog.Columns().ToStatus: in.ToStatus,
		dao.MvpTaskLog.Columns().Message: in.Message,
		dao.MvpTaskLog.Columns().Operator: in.Operator,
		dao.MvpTaskLog.Columns().CreatedAt: gtime.Now(),
		dao.MvpTaskLog.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
	return err
}

// Update 更新任务日志表
func (s *sTaskLog) Update(ctx context.Context, in *model.TaskLogUpdateInput) error {
	data := g.Map{
		dao.MvpTaskLog.Columns().TaskId: in.TaskID,
		dao.MvpTaskLog.Columns().Action: in.Action,
		dao.MvpTaskLog.Columns().FromStatus: in.FromStatus,
		dao.MvpTaskLog.Columns().ToStatus: in.ToStatus,
		dao.MvpTaskLog.Columns().Message: in.Message,
		dao.MvpTaskLog.Columns().Operator: in.Operator,
		dao.MvpTaskLog.Columns().UpdatedAt: gtime.Now(),
	}
	_, err := dao.MvpTaskLog.Ctx(ctx).Where(dao.MvpTaskLog.Columns().Id, in.ID).Data(data).Update()
	return err
}

// Delete 软删除任务日志表
func (s *sTaskLog) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.MvpTaskLog.Ctx(ctx).Where(dao.MvpTaskLog.Columns().Id, id).Data(g.Map{
		dao.MvpTaskLog.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// BatchDelete 批量软删除任务日志表
func (s *sTaskLog) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.MvpTaskLog.Ctx(ctx).WhereIn(dao.MvpTaskLog.Columns().Id, ids).Data(g.Map{
		dao.MvpTaskLog.Columns().DeletedAt: gtime.Now(),
	}).Update()
	return err
}

// Detail 获取任务日志表详情
func (s *sTaskLog) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.TaskLogDetailOutput, err error) {
	out = &model.TaskLogDetailOutput{}
	err = dao.MvpTaskLog.Ctx(ctx).Where(dao.MvpTaskLog.Columns().Id, id).Where(dao.MvpTaskLog.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
	// 查询任务ID关联显示
	if out.TaskID != 0 {
		val, err := g.DB().Ctx(ctx).Model("mvp_task").Where("id", out.TaskID).Where("deleted_at", nil).Value("name")
		if err == nil {
			out.TaskName = val.String()
		}
	}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *sTaskLog) applyListFilter(ctx context.Context, in *model.TaskLogListInput) *gdb.Model {
	m := dao.MvpTaskLog.Ctx(ctx).Where(dao.MvpTaskLog.Columns().DeletedAt, nil)
	if in.StartTime != "" {
		m = m.WhereGTE(dao.MvpTaskLog.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.MvpTaskLog.Columns().CreatedAt, in.EndTime)
	}
	return m
}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *sTaskLog) fillRefFields(ctx context.Context, list []*model.TaskLogListOutput) {
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.TaskID != 0 {
				idSet[int64(item.TaskID)] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("mvp_task").
				Fields("id", "name").
				Where("deleted_at", nil).
				WhereIn("id", ids).
				All()
			if err == nil {
				refMap := make(map[int64]string, len(rows))
				for _, row := range rows {
					refMap[row["id"].Int64()] = row["name"].String()
				}
				for _, item := range list {
					if val, ok := refMap[int64(item.TaskID)]; ok {
						item.TaskName = val
					}
				}
			}
		}
	}
}

// List 获取任务日志表列表
func (s *sTaskLog) List(ctx context.Context, in *model.TaskLogListInput) (list []*model.TaskLogListOutput, total int, err error) {
	m := s.applyListFilter(ctx, in)
	total, err = m.Count()
	if err != nil {
		return
	}
	// 动态排序
	if in.OrderBy != "" {
		if in.OrderDir == "desc" {
			m = m.OrderDesc(in.OrderBy)
		} else {
			m = m.OrderAsc(in.OrderBy)
		}
	} else {
		m = m.OrderAsc(dao.MvpTaskLog.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}
// Export 导出任务日志表（不分页）
func (s *sTaskLog) Export(ctx context.Context, in *model.TaskLogListInput) (list []*model.TaskLogListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.MvpTaskLog.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
	s.fillRefFields(ctx, list)
	return
}




// Import 导入任务日志表
func (s *sTaskLog) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.MvpTaskLog.Columns().Id:        id,
			dao.MvpTaskLog.Columns().CreatedAt: gtime.Now(),
			dao.MvpTaskLog.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().TaskId] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().Action] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().FromStatus] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().ToStatus] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().Message] = record[idx]
		}
		idx++
		if idx < len(record) {
			data[dao.MvpTaskLog.Columns().Operator] = record[idx]
		}
		idx++
		if _, insertErr := dao.MvpTaskLog.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}

