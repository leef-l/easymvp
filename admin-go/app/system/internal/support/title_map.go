package support

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

// UniqueNonZeroIDs 去重并过滤零值
func UniqueNonZeroIDs(ids []int64) []int64 {
	if len(ids) == 0 {
		return make([]int64, 0)
	}

	result := make([]int64, 0, len(ids))
	seen := make(map[int64]struct{}, len(ids))
	for _, id := range ids {
		if id == 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// LoadTitleMap 批量加载 title 字段映射
func LoadTitleMap(ctx context.Context, table string, ids []int64) map[int64]string {
	ids = UniqueNonZeroIDs(ids)
	if len(ids) == 0 {
		return map[int64]string{}
	}

	var rows []struct {
		Id    int64  `json:"id"`
		Title string `json:"title"`
	}
	_ = g.DB().Ctx(ctx).Model(table).
		Fields("id,title").
		Where("id", ids).
		Where("deleted_at", nil).
		Scan(&rows)

	titleMap := make(map[int64]string, len(rows))
	for _, row := range rows {
		titleMap[row.Id] = row.Title
	}
	return titleMap
}

// LoadTitle 加载单条 title
func LoadTitle(ctx context.Context, table string, id int64) string {
	if id == 0 {
		return ""
	}
	return LoadTitleMap(ctx, table, []int64{id})[id]
}
