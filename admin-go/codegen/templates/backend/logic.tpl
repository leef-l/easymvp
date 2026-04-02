package {{.PackageName}}

import (
	"context"
{{- if .HasImport}}
	"encoding/csv"
{{- end}}
{{- if or .EnableOpLog .HasImport}}
	"fmt"
{{- end}}

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
{{- if .HasImport}}
	"github.com/gogf/gf/v2/net/ghttp"
{{- end}}
	"github.com/gogf/gf/v2/os/gtime"
{{- if .HasPassword}}
	"golang.org/x/crypto/bcrypt"
{{- end}}

	"easymvp/app/{{.AppName}}/internal/dao"
{{- if or .HasCreatedBy .HasDeptID}}
	"easymvp/app/{{.AppName}}/internal/middleware"
{{- end}}
	"easymvp/app/{{.AppName}}/internal/model"
	"easymvp/app/{{.AppName}}/internal/service"
	"easymvp/utility/snowflake"
{{- if .EnableOpLog}}
	"easymvp/utility/oplog"
{{- end}}
)

func init() {
	service.Register{{.ModelName}}(New())
}

func New() *s{{.ModelName}} {
	return &s{{.ModelName}}{}
}

type s{{.ModelName}} struct{}

// Create 创建{{.Comment}}
func (s *s{{.ModelName}}) Create(ctx context.Context, in *model.{{.ModelName}}CreateInput) error {
	id := snowflake.Generate()
{{- range .Fields}}
{{- if .IsPassword}}
	hashed{{.NameCamel}}, err := bcrypt.GenerateFromPassword([]byte(in.{{.NameCamel}}), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
{{- end}}
{{- end}}
	_, err{{if not .HasPassword}} :={{else}} ={{end}} dao.{{.DaoName}}.Ctx(ctx).Data(g.Map{
		dao.{{.DaoName}}.Columns().Id:        id,
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden)}}
{{- if .IsPassword}}
		dao.{{$.DaoName}}.Columns().{{.NameDao}}: string(hashed{{.NameCamel}}),
{{- else}}
		dao.{{$.DaoName}}.Columns().{{.NameDao}}: in.{{.NameCamel}},
{{- end}}
{{- end}}
{{- end}}
{{- if .HasCreatedBy}}
		dao.{{.DaoName}}.Columns().CreatedBy: middleware.GetUserID(ctx),
{{- end}}
{{- if .HasDeptID}}
		dao.{{.DaoName}}.Columns().DeptId: middleware.GetDeptID(ctx),
{{- end}}
		dao.{{.DaoName}}.Columns().CreatedAt: gtime.Now(),
		dao.{{.DaoName}}.Columns().UpdatedAt: gtime.Now(),
	}).Insert()
{{- if .EnableOpLog}}
	if err == nil {
		go oplog.Record(ctx, "{{.ModuleName}}", "create", fmt.Sprintf("%v", id), "")
	}
{{- end}}
	return err
}

// Update 更新{{.Comment}}
func (s *s{{.ModelName}}) Update(ctx context.Context, in *model.{{.ModelName}}UpdateInput) error {
	data := g.Map{
{{- range .Fields}}
{{- if and (not .IsID) (not .IsHidden) (not .IsPassword)}}
		dao.{{$.DaoName}}.Columns().{{.NameDao}}: in.{{.NameCamel}},
{{- end}}
{{- end}}
		dao.{{.DaoName}}.Columns().UpdatedAt: gtime.Now(),
	}
{{- range .Fields}}
{{- if .IsPassword}}
	if in.{{.NameCamel}} != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(in.{{.NameCamel}}), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		data[dao.{{$.DaoName}}.Columns().{{.NameDao}}] = string(hashed)
	}
{{- end}}
{{- end}}
{{- if .HasMoney}}
	// 含金额字段，使用事务 + 行锁保证并发安全
	err := dao.{{.DaoName}}.Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		// FOR UPDATE 行锁
		_, err := tx.Model(dao.{{.DaoName}}.Table()).Ctx(ctx).
			Where(dao.{{.DaoName}}.Columns().Id, in.ID).
			LockUpdate().
			Value(dao.{{.DaoName}}.Columns().Id)
		if err != nil {
			return err
		}
		_, err = tx.Model(dao.{{.DaoName}}.Table()).Ctx(ctx).
			Where(dao.{{.DaoName}}.Columns().Id, in.ID).
			Data(data).Update()
		return err
	})
{{- else}}
	_, err := dao.{{.DaoName}}.Ctx(ctx).Where(dao.{{.DaoName}}.Columns().Id, in.ID).Data(data).Update()
{{- end}}
{{- if .EnableOpLog}}
	if err == nil {
		go oplog.Record(ctx, "{{.ModuleName}}", "update", fmt.Sprintf("%v", in.ID), "")
	}
{{- end}}
	return err
}

// Delete 软删除{{.Comment}}
func (s *s{{.ModelName}}) Delete(ctx context.Context, id snowflake.JsonInt64) error {
	_, err := dao.{{.DaoName}}.Ctx(ctx).Where(dao.{{.DaoName}}.Columns().Id, id).Data(g.Map{
		dao.{{.DaoName}}.Columns().DeletedAt: gtime.Now(),
	}).Update()
{{- if .EnableOpLog}}
	if err == nil {
		go oplog.Record(ctx, "{{.ModuleName}}", "delete", fmt.Sprintf("%v", id), "")
	}
{{- end}}
	return err
}

// BatchDelete 批量软删除{{.Comment}}
func (s *s{{.ModelName}}) BatchDelete(ctx context.Context, ids []snowflake.JsonInt64) error {
	_, err := dao.{{.DaoName}}.Ctx(ctx).WhereIn(dao.{{.DaoName}}.Columns().Id, ids).Data(g.Map{
		dao.{{.DaoName}}.Columns().DeletedAt: gtime.Now(),
	}).Update()
{{- if .EnableOpLog}}
	if err == nil {
		go oplog.Record(ctx, "{{.ModuleName}}", "batch-delete", fmt.Sprintf("%v", ids), "")
	}
{{- end}}
	return err
}

// Detail 获取{{.Comment}}详情
func (s *s{{.ModelName}}) Detail(ctx context.Context, id snowflake.JsonInt64) (out *model.{{.ModelName}}DetailOutput, err error) {
	out = &model.{{.ModelName}}DetailOutput{}
	err = dao.{{.DaoName}}.Ctx(ctx).Where(dao.{{.DaoName}}.Columns().Id, id).Where(dao.{{.DaoName}}.Columns().DeletedAt, nil).Scan(out)
	if err != nil {
		return nil, err
	}
{{- range .Fields}}
{{- if and .RefFieldName (not .IsHidden)}}
	// 查询{{.Label}}关联显示
	if out.{{.NameCamel}} != 0 {
		val, err := g.DB().Ctx(ctx).Model("{{.RefTableDB}}").Where("id", out.{{.NameCamel}}).Where("deleted_at", nil).Value("{{.RefDisplayField}}")
		if err == nil {
			out.{{.RefFieldName}} = val.String()
		}
	}
{{- end}}
{{- end}}
	return
}

// applyListFilter 应用列表通用过滤条件
func (s *s{{.ModelName}}) applyListFilter(ctx context.Context, in *model.{{.ModelName}}ListInput) *gdb.Model {
	m := dao.{{.DaoName}}.Ctx(ctx).Where(dao.{{.DaoName}}.Columns().DeletedAt, nil)
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (.IsEnum)}}
	if in.{{.NameCamel}} != nil {
		m = m.Where(dao.{{$.DaoName}}.Columns().{{.NameDao}}, *in.{{.NameCamel}})
	}
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	if in.{{.NameCamel}} != "" {
{{- if .IsExactSearch}}
		m = m.Where(dao.{{$.DaoName}}.Columns().{{.NameDao}}, in.{{.NameCamel}})
{{- else}}
		m = m.WhereLike(dao.{{$.DaoName}}.Columns().{{.NameDao}}, "%"+in.{{.NameCamel}}+"%")
{{- end}}
	}
{{- end}}
{{- end}}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.{{.DaoName}}.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.{{.DaoName}}.Columns().CreatedAt, in.EndTime)
	}
{{- if or .HasCreatedBy .HasDeptID}}
	// 数据权限过滤
	m = middleware.ApplyDataScope(ctx, m{{if .HasCreatedBy}}, dao.{{.DaoName}}.Columns().CreatedBy{{end}}{{if .HasDeptID}}, dao.{{.DaoName}}.Columns().DeptId{{end}})
{{- end}}
	return m
}

{{- $hasRef := false}}
{{- range .Fields}}
{{- if and .RefFieldName (not .IsHidden)}}
{{- $hasRef = true}}
{{- end}}
{{- end}}
{{- if $hasRef}}

// fillRefFields 批量填充关联显示字段（避免 N+1 查询）
func (s *s{{.ModelName}}) fillRefFields(ctx context.Context, list []*model.{{.ModelName}}ListOutput) {
{{- range .Fields}}
{{- if and .RefFieldName (not .IsHidden)}}
	{
		idSet := make(map[int64]struct{})
		for _, item := range list {
			if item.{{.NameCamel}} != 0 {
				idSet[int64(item.{{.NameCamel}})] = struct{}{}
			}
		}
		if len(idSet) > 0 {
			ids := make([]int64, 0, len(idSet))
			for id := range idSet {
				ids = append(ids, id)
			}
			rows, err := g.DB().Ctx(ctx).Model("{{.RefTableDB}}").
				Fields("id", "{{.RefDisplayField}}").
				Where("deleted_at", nil).
				WhereIn("id", ids).
				All()
			if err == nil {
				refMap := make(map[int64]string, len(rows))
				for _, row := range rows {
					refMap[row["id"].Int64()] = row["{{.RefDisplayField}}"].String()
				}
				for _, item := range list {
					if val, ok := refMap[int64(item.{{.NameCamel}})]; ok {
						item.{{.RefFieldName}} = val
					}
				}
			}
		}
	}
{{- end}}
{{- end}}
}
{{- end}}

// List 获取{{.Comment}}列表
func (s *s{{.ModelName}}) List(ctx context.Context, in *model.{{.ModelName}}ListInput) (list []*model.{{.ModelName}}ListOutput, total int, err error) {
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
		m = m.OrderAsc(dao.{{.DaoName}}.Columns().Id)
	}
	err = m.Page(in.PageNum, in.PageSize).Scan(&list)
	if err != nil {
		return
	}
{{- if $hasRef}}
	s.fillRefFields(ctx, list)
{{- end}}
	return
}
// Export 导出{{.Comment}}（不分页）
func (s *s{{.ModelName}}) Export(ctx context.Context, in *model.{{.ModelName}}ListInput) (list []*model.{{.ModelName}}ListOutput, err error) {
	m := s.applyListFilter(ctx, in)
	err = m.OrderAsc(dao.{{.DaoName}}.Columns().Id).Limit(10000).Scan(&list)
	if err != nil {
		return
	}
{{- if $hasRef}}
	s.fillRefFields(ctx, list)
{{- end}}
	return
}

{{if .HasParentID}}
// Tree 获取{{.Comment}}树形结构
func (s *s{{.ModelName}}) Tree(ctx context.Context, in *model.{{.ModelName}}TreeInput) (tree []*model.{{.ModelName}}TreeOutput, err error) {
	var list []*model.{{.ModelName}}TreeOutput
	m := dao.{{.DaoName}}.Ctx(ctx).Where(dao.{{.DaoName}}.Columns().DeletedAt, nil)
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsParentID) (.IsEnum)}}
	if in.{{.NameCamel}} != nil {
		m = m.Where(dao.{{$.DaoName}}.Columns().{{.NameDao}}, *in.{{.NameCamel}})
	}
{{- end}}
{{- end}}
{{- range .Fields}}
{{- if .IsSearchable}}
	if in.{{.NameCamel}} != "" {
{{- if .IsExactSearch}}
		m = m.Where(dao.{{$.DaoName}}.Columns().{{.NameDao}}, in.{{.NameCamel}})
{{- else}}
		m = m.WhereLike(dao.{{$.DaoName}}.Columns().{{.NameDao}}, "%"+in.{{.NameCamel}}+"%")
{{- end}}
	}
{{- end}}
{{- end}}
	if in.StartTime != "" {
		m = m.WhereGTE(dao.{{.DaoName}}.Columns().CreatedAt, in.StartTime)
	}
	if in.EndTime != "" {
		m = m.WhereLTE(dao.{{.DaoName}}.Columns().CreatedAt, in.EndTime)
	}
	err = m.{{if .HasSort}}OrderAsc(dao.{{.DaoName}}.Columns().Sort).{{end}}Scan(&list)
	if err != nil {
		return
	}

	// 使用 map 迭代方式组装树
	nodeMap := make(map[int64]*model.{{.ModelName}}TreeOutput, len(list))
	for _, item := range list {
		item.Children = make([]*model.{{.ModelName}}TreeOutput, 0)
		nodeMap[int64(item.ID)] = item
	}

	tree = make([]*model.{{.ModelName}}TreeOutput, 0)
	for _, item := range list {
		if int64(item.ParentID) == 0 {
			tree = append(tree, item)
		} else if parent, ok := nodeMap[int64(item.ParentID)]; ok {
			parent.Children = append(parent.Children, item)
		}
	}
	return
}
{{end}}
{{if .HasBatchEdit}}
// BatchUpdate 批量编辑{{.Comment}}
func (s *s{{.ModelName}}) BatchUpdate(ctx context.Context, in *model.{{.ModelName}}BatchUpdateInput) error {
	data := g.Map{
		dao.{{.DaoName}}.Columns().UpdatedAt: gtime.Now(),
	}
	if in.Status != nil {
		data[dao.{{.DaoName}}.Columns().Status] = *in.Status
	}
	_, err := dao.{{.DaoName}}.Ctx(ctx).WhereIn(dao.{{.DaoName}}.Columns().Id, in.IDs).Data(data).Update()
	return err
}
{{end}}
{{if .HasImport}}
// Import 导入{{.Comment}}
func (s *s{{.ModelName}}) Import(ctx context.Context, file *ghttp.UploadFile) (success int, fail int, err error) {
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
			dao.{{.DaoName}}.Columns().Id:        id,
			dao.{{.DaoName}}.Columns().CreatedAt: gtime.Now(),
			dao.{{.DaoName}}.Columns().UpdatedAt: gtime.Now(),
		}
		idx := 0
{{- range .Fields}}
{{- if and (not .IsHidden) (not .IsID) (not .IsPassword) (not .IsTimeField)}}
		if idx < len(record) {
			data[dao.{{$.DaoName}}.Columns().{{.NameDao}}] = record[idx]
		}
		idx++
{{- end}}
{{- end}}
		if _, insertErr := dao.{{.DaoName}}.Ctx(ctx).Data(data).Insert(); insertErr != nil {
			fail++
		} else {
			success++
		}
	}
	return
}
{{end}}
