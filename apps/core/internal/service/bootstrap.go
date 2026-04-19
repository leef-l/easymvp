package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"

	_ "modernc.org/sqlite"
)

type migrationFile struct {
	version  int
	name     string
	checksum string
	sqlBody  string
}

func Bootstrap(ctx context.Context) error {
	var (
		startup       = CurrentStartupConfig(ctx)
		dataRoot      = startup.DataRoot
		dbPath        = startup.DBPath
		migrationPath = startup.MigrationPath
	)

	if dataRoot == "" {
		return gerror.New("easymvp.dataRoot is empty")
	}
	if dbPath == "" {
		return gerror.New("easymvp.dbPath is empty")
	}
	if migrationPath == "" {
		return gerror.New("easymvp.migrationPath is empty")
	}

	if err := ensureDataRootLayout(dataRoot); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return gerror.Wrap(err, "create db dir failed")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return gerror.Wrap(err, "open sqlite failed")
	}
	defer func() {
		_ = db.Close()
	}()

	if err = applyPragmas(db); err != nil {
		return gerror.Wrap(err, "apply sqlite pragmas failed")
	}
	if err = ensureSchemaMigrationsTable(db); err != nil {
		return gerror.Wrap(err, "ensure schema_migrations table failed")
	}
	if err = runMigrations(db, migrationPath); err != nil {
		return err
	}
	return nil
}

func ensureDataRootLayout(dataRoot string) error {
	root := filepath.Clean(strings.TrimSpace(dataRoot))
	if root == "" || root == "." {
		return gerror.New("easymvp.dataRoot is invalid")
	}

	dirs := []string{
		root,
		filepath.Join(root, "data"),
		filepath.Join(root, "projects"),
		filepath.Join(root, "settings"),
		filepath.Join(root, "backups"),
		filepath.Join(root, "temp"),
		filepath.Join(root, "diagnostics"),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return gerror.Wrapf(err, "create data root dir failed: %s", dir)
		}
	}
	return nil
}

func applyPragmas(db *sql.DB) error {
	var pragmas = []string{
		`PRAGMA journal_mode = WAL;`,
		`PRAGMA foreign_keys = ON;`,
		`PRAGMA busy_timeout = 5000;`,
		`PRAGMA synchronous = NORMAL;`,
	}

	for _, stmt := range pragmas {
		if _, err := db.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func ensureSchemaMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  error_message TEXT
);
`)
	return err
}

func runMigrations(db *sql.DB, migrationPath string) error {
	migrations, err := loadMigrations(migrationPath)
	if err != nil {
		return err
	}

	for _, migration := range migrations {
		applied, err := isMigrationApplied(db, migration.version, migration.checksum)
		if err != nil {
			return gerror.Wrapf(err, "check migration %d failed", migration.version)
		}
		if applied {
			continue
		}

		if err = applyMigration(db, migration); err != nil {
			return err
		}
	}
	return nil
}

func loadMigrations(dir string) ([]migrationFile, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, gerror.Wrap(err, "read migration dir failed")
	}

	var migrations []migrationFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		fullPath := filepath.Join(dir, entry.Name())
		content, err := os.ReadFile(fullPath)
		if err != nil {
			return nil, gerror.Wrapf(err, "read migration file failed: %s", entry.Name())
		}

		version, err := parseMigrationVersion(entry.Name())
		if err != nil {
			return nil, err
		}

		sum := sha256.Sum256(content)
		migrations = append(migrations, migrationFile{
			version:  version,
			name:     entry.Name(),
			checksum: hex.EncodeToString(sum[:]),
			sqlBody:  string(content),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}

func parseMigrationVersion(fileName string) (int, error) {
	parts := strings.SplitN(fileName, "_", 2)
	if len(parts) < 2 {
		return 0, gerror.Newf("invalid migration file name: %s", fileName)
	}

	version, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, gerror.Wrapf(err, "parse migration version failed: %s", fileName)
	}
	return version, nil
}

func isMigrationApplied(db *sql.DB, version int, checksum string) (bool, error) {
	var (
		storedChecksum string
		status         string
	)

	err := db.QueryRow(
		`SELECT checksum, status FROM schema_migrations WHERE version = ?`,
		version,
	).Scan(&storedChecksum, &status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if storedChecksum != checksum {
		return false, gerror.Newf("migration checksum mismatch for version %d", version)
	}
	return status == "applied", nil
}

func applyMigration(db *sql.DB, migration migrationFile) error {
	tx, err := db.Begin()
	if err != nil {
		return gerror.Wrapf(err, "begin migration %d failed", migration.version)
	}

	if _, err = tx.Exec(migration.sqlBody); err != nil {
		_ = tx.Rollback()
		recordMigrationFailure(db, migration, err)
		return gerror.Wrapf(err, "execute migration %d failed", migration.version)
	}

	if _, err = tx.Exec(
		`INSERT OR REPLACE INTO schema_migrations(version,name,checksum,applied_at,duration_ms,status,error_message) VALUES(?,?,?,?,?,?,?)`,
		migration.version,
		migration.name,
		migration.checksum,
		gtime.Now().String(),
		0,
		"applied",
		"",
	); err != nil {
		_ = tx.Rollback()
		recordMigrationFailure(db, migration, err)
		return gerror.Wrapf(err, "record migration %d failed", migration.version)
	}

	if err = tx.Commit(); err != nil {
		return gerror.Wrapf(err, "commit migration %d failed", migration.version)
	}
	return nil
}

func recordMigrationFailure(db *sql.DB, migration migrationFile, migrationErr error) {
	if db == nil || migrationErr == nil {
		return
	}
	_, _ = db.Exec(
		`INSERT OR REPLACE INTO schema_migrations(version,name,checksum,applied_at,duration_ms,status,error_message) VALUES(?,?,?,?,?,?,?)`,
		migration.version,
		migration.name,
		migration.checksum,
		gtime.Now().String(),
		0,
		"failed",
		migrationErr.Error(),
	)
}

func _unusedContext(_ context.Context) {
	fmt.Print("")
}
