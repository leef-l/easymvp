package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "easymvp/app/mvp/internal/packed"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gctx"
)

func main() {
	ctx := gctx.GetInitCtx()

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "migrate-url":
		url, err := migrateURL(ctx)
		exitIfErr(err)
		fmt.Println(url)
	case "seed":
		runSeed(ctx, os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Usage:
  go run ./app/mvp/tools/dbctl migrate-url
  go run ./app/mvp/tools/dbctl seed [-file manifest/sql/seed/mysql_seed.sql] [-force]
`)
}

func runSeed(ctx context.Context, args []string) {
	fs := flag.NewFlagSet("seed", flag.ExitOnError)
	file := fs.String("file", "manifest/sql/seed/mysql_seed.sql", "seed SQL file")
	force := fs.Bool("force", false, "allow applying seed on non-empty database")
	_ = fs.Parse(args)

	absFile, err := filepath.Abs(*file)
	exitIfErr(err)

	content, err := os.ReadFile(absFile)
	exitIfErr(err)
	sqlText := strings.TrimSpace(string(content))
	if sqlText == "" {
		exitIfErr(errors.New("seed SQL file is empty"))
	}

	dsn, err := mysqlDSN(ctx)
	exitIfErr(err)

	db, err := sql.Open("mysql", dsn)
	exitIfErr(err)
	defer db.Close()

	if err := db.PingContext(ctx); err != nil {
		exitIfErr(fmt.Errorf("database ping failed: %w", err))
	}

	if !*force {
		var count int
		err = db.QueryRowContext(ctx, "SELECT COUNT(1) FROM system_users").Scan(&count)
		if err != nil {
			exitIfErr(fmt.Errorf("seed safety check failed, run schema migrations first: %w", err))
		}
		if count > 0 {
			exitIfErr(fmt.Errorf("seed refused: system_users already has %d rows; rerun with -force if you really want this", count))
		}
	}

	if _, err := db.ExecContext(ctx, sqlText); err != nil {
		exitIfErr(fmt.Errorf("seed execution failed: %w", err))
	}

	fmt.Printf("seed applied: %s\n", absFile)
}

func migrateURL(ctx context.Context) (string, error) {
	link, err := databaseLink(ctx)
	if err != nil {
		return "", err
	}

	switch {
	case strings.HasPrefix(link, "mysql://"):
		return ensureQueryFlag(link, "multiStatements", "true"), nil
	case strings.HasPrefix(link, "mysql:"):
		return ensureQueryFlag("mysql://"+strings.TrimPrefix(link, "mysql:"), "multiStatements", "true"), nil
	default:
		return "", fmt.Errorf("unsupported database link format: %s", link)
	}
}

func mysqlDSN(ctx context.Context) (string, error) {
	link, err := databaseLink(ctx)
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(link, "mysql:") {
		return "", fmt.Errorf("unsupported database link format: %s", link)
	}
	return ensureQueryFlag(strings.TrimPrefix(link, "mysql:"), "multiStatements", "true"), nil
}

func databaseLink(ctx context.Context) (string, error) {
	value, err := g.Cfg().Get(ctx, "database.default.link")
	if err != nil {
		return "", fmt.Errorf("read database.default.link failed: %w", err)
	}
	link := strings.TrimSpace(value.String())
	if link == "" {
		return "", errors.New("database.default.link is empty")
	}
	return link, nil
}

func ensureQueryFlag(link, key, value string) string {
	flagText := key + "="
	if strings.Contains(link, flagText) {
		return link
	}
	if strings.Contains(link, "?") {
		return link + "&" + key + "=" + value
	}
	return link + "?" + key + "=" + value
}

func exitIfErr(err error) {
	if err == nil {
		return
	}
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
