package service

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestEnsureDataRootLayoutCreatesExpectedDirectories(t *testing.T) {
	t.Parallel()

	root := filepath.Join(t.TempDir(), "runtime")
	if err := ensureDataRootLayout(root); err != nil {
		t.Fatalf("ensureDataRootLayout failed: %v", err)
	}

	expected := []string{
		root,
		filepath.Join(root, "data"),
		filepath.Join(root, "projects"),
		filepath.Join(root, "settings"),
		filepath.Join(root, "backups"),
		filepath.Join(root, "temp"),
		filepath.Join(root, "diagnostics"),
	}

	for _, dir := range expected {
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("expected directory %s: %v", dir, err)
		}
		if !info.IsDir() {
			t.Fatalf("expected %s to be a directory", dir)
		}
	}
}

func TestEnsureDataRootLayoutRejectsDot(t *testing.T) {
	t.Parallel()

	if err := ensureDataRootLayout("."); err == nil {
		t.Fatal("expected dot path to be rejected")
	}
}

func TestLoadMigrationsSortsByVersion(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	files := map[string]string{
		"0010_zeta.sql": "CREATE TABLE zeta(id INTEGER);",
		"0002_beta.sql": "CREATE TABLE beta(id INTEGER);",
		"0001_alpha.sql": "CREATE TABLE alpha(id INTEGER);",
		"README.md":      "ignored",
	}

	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatalf("write migration fixture %s failed: %v", name, err)
		}
	}

	migrations, err := loadMigrations(dir)
	if err != nil {
		t.Fatalf("loadMigrations failed: %v", err)
	}

	got := []int{
		migrations[0].version,
		migrations[1].version,
		migrations[2].version,
	}
	want := []int{1, 2, 10}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("migration versions = %#v, want %#v", got, want)
	}
}

func TestParseMigrationVersion(t *testing.T) {
	t.Parallel()

	if got, err := parseMigrationVersion("0012_add_index.sql"); err != nil || got != 12 {
		t.Fatalf("parseMigrationVersion returned (%d, %v), want (12, nil)", got, err)
	}
	if _, err := parseMigrationVersion("broken-name.sql"); err == nil {
		t.Fatal("expected invalid migration file name to fail")
	}
}
