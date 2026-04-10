package configstore

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

type UpsertInput struct {
	Key         string
	Value       string
	ConfigType  string
	Category    string
	Description string
	CreatedBy   int64
	DeptID      int64
}

type Store struct{}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) GetValueByKey(ctx context.Context, key string) (value string, err error) {
	defer func() {
		if recover() != nil {
			value = ""
			err = nil
		}
	}()
	record, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Fields("config_value").
		One()
	if err != nil || record.IsEmpty() {
		return "", err
	}
	return record["config_value"].String(), nil
}

func (s *Store) UpsertByKey(ctx context.Context, in UpsertInput) error {
	record, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", in.Key).
		WhereNull("deleted_at").
		Fields("id").
		One()
	if err != nil {
		return err
	}
	if !record.IsEmpty() {
		_, err = g.DB().Model("mvp_config").Ctx(ctx).
			Where("id", record["id"].Int64()).
			Data(g.Map{
				"config_value": in.Value,
				"config_type":  in.ConfigType,
				"category":     in.Category,
				"description":  in.Description,
				"updated_at":   gtime.Now(),
			}).
			Update()
		return err
	}

	_, err = g.DB().Model("mvp_config").Ctx(ctx).Insert(g.Map{
		"config_key":   in.Key,
		"config_value": in.Value,
		"config_type":  in.ConfigType,
		"category":     in.Category,
		"description":  in.Description,
		"created_by":   in.CreatedBy,
		"dept_id":      in.DeptID,
		"created_at":   gtime.Now(),
		"updated_at":   gtime.Now(),
	})
	return err
}

func (s *Store) DeleteByKey(ctx context.Context, key string) error {
	_, err := g.DB().Model("mvp_config").Ctx(ctx).
		Where("config_key", key).
		WhereNull("deleted_at").
		Data(g.Map{
			"deleted_at": gtime.Now(),
			"updated_at": gtime.Now(),
		}).
		Update()
	return err
}
