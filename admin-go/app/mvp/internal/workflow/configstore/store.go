package configstore

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"

	"easymvp/app/mvp/internal/workflow/configrepo"
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
	configRepo := configrepo.NewConfigRepo()
	defer func() {
		if recover() != nil {
			value = ""
			err = nil
		}
	}()
	return configRepo.GetValueByKey(ctx, key)
}

func (s *Store) UpsertByKey(ctx context.Context, in UpsertInput) error {
	configRepo := configrepo.NewConfigRepo()
	return configRepo.UpsertByKey(ctx, in.Key, g.Map{
		"config_value": in.Value,
		"config_type":  in.ConfigType,
		"category":     in.Category,
		"description":  in.Description,
		"created_by":   in.CreatedBy,
		"dept_id":      in.DeptID,
	})
}

func (s *Store) DeleteByKey(ctx context.Context, key string) error {
	return configrepo.NewConfigRepo().SoftDeleteByKey(ctx, key)
}
