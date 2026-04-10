package rolecatalog

import (
	"context"
	"strings"
	"testing"

	"easymvp/app/mvp/internal/workflow/configstore"
)

type fakeConfigStore struct {
	deletedKey string
	saved      configstore.UpsertInput
	value      string
}

func (f *fakeConfigStore) DeleteByKey(ctx context.Context, key string) error {
	f.deletedKey = key
	return nil
}

func (f *fakeConfigStore) GetValueByKey(ctx context.Context, key string) (string, error) {
	return f.value, nil
}

func (f *fakeConfigStore) UpsertByKey(ctx context.Context, in configstore.UpsertInput) error {
	f.saved = in
	return nil
}

func TestServiceResetDeletesRoleDefinitionConfig(t *testing.T) {
	t.Parallel()

	store := &fakeConfigStore{}
	svc := NewService(store)
	if err := svc.Reset(context.Background()); err != nil {
		t.Fatalf("Reset returned error: %v", err)
	}
	if store.deletedKey != ConfigKeyRoleDefinitions {
		t.Fatalf("expected deleted key %q, got %q", ConfigKeyRoleDefinitions, store.deletedKey)
	}
}

func TestServiceSavePersistsNormalizedDefinitions(t *testing.T) {
	t.Parallel()

	store := &fakeConfigStore{}
	svc := NewService(store)
	err := svc.Save(context.Background(), []Definition{
		{RoleType: "experience_reviewer", DisplayName: "体验评审师", AcceptanceJudge: false},
	}, 12, 34)
	if err != nil {
		t.Fatalf("Save returned error: %v", err)
	}
	if store.saved.Key != ConfigKeyRoleDefinitions {
		t.Fatalf("expected config key %q, got %q", ConfigKeyRoleDefinitions, store.saved.Key)
	}
	if !strings.Contains(store.saved.Value, `"roleType":"experience_reviewer"`) {
		t.Fatalf("expected serialized role definition, got %s", store.saved.Value)
	}
	if !strings.Contains(store.saved.Value, `"color":"magenta"`) {
		t.Fatalf("expected builtin defaults to be filled, got %s", store.saved.Value)
	}
	store.value = store.saved.Value
	list := svc.List(context.Background())
	found := false
	for _, item := range list {
		if item.RoleType != "experience_reviewer" {
			continue
		}
		found = true
		if item.AcceptanceJudge {
			t.Fatalf("expected explicit false acceptanceJudge to survive round-trip, got %+v", item)
		}
	}
	if !found {
		t.Fatalf("expected experience_reviewer definition in round-trip list")
	}
}
