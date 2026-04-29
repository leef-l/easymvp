package main

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	
	_ "modernc.org/sqlite"
)

func main() {
	dbPath := `C:\Users\Public\project\easymvp\var\data\easymvp.db`
	migrationDir := `C:\Users\Public\project\easymvp\apps\core\manifest\migrations`
	
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		fmt.Println("open db error:", err)
		os.Exit(1)
	}
	defer db.Close()
	
	// Create schema_migrations table
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS schema_migrations (
  version INTEGER PRIMARY KEY,
  name TEXT NOT NULL,
  checksum TEXT NOT NULL,
  applied_at TEXT NOT NULL,
  duration_ms INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL,
  error_message TEXT
);`)
	if err != nil {
		fmt.Println("create schema_migrations error:", err)
		os.Exit(1)
	}
	
	// Read migration files
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		fmt.Println("read dir error:", err)
		os.Exit(1)
	}
	
	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			files = append(files, entry.Name())
		}
	}
	sort.Strings(files)
	
	for _, file := range files {
		content, err := os.ReadFile(filepath.Join(migrationDir, file))
		if err != nil {
			fmt.Println("read file error:", file, err)
			continue
		}
		
		// Check if already applied
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM schema_migrations WHERE name = ? AND status = 'applied'", file).Scan(&count)
		if err != nil {
			fmt.Println("check migration error:", file, err)
			continue
		}
		if count > 0 {
			fmt.Println("SKIP (already applied):", file)
			continue
		}
		
		// Apply migration
		_, err = db.Exec(string(content))
		if err != nil {
			fmt.Println("APPLY FAILED:", file, err)
			continue
		}
		
		// Record migration
		_, err = db.Exec(
			"INSERT INTO schema_migrations(version,name,checksum,applied_at,duration_ms,status) VALUES(?,?,?,?,?,?)",
			0, file, "manual", "2026-04-25T18:00:00", 0, "applied",
		)
		if err != nil {
			fmt.Println("record migration error:", file, err)
			continue
		}
		
		fmt.Println("APPLIED:", file)
	}
	
	fmt.Println("\nDone!")
}
