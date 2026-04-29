package repo

import (
	"context"
	"database/sql"
	"encoding/json"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
)

var resourceIDCounter uint64

func openProjectDatabase(ctx context.Context) (*sql.DB, func(), error) {
	dbPath := g.Cfg().MustGet(ctx, "easymvp.dbPath").String()
	if strings.TrimSpace(dbPath) == "" {
		return nil, nil, gerror.New("easymvp.dbPath is empty")
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, nil, gerror.Wrap(err, "open sqlite failed")
	}

	closeFn := func() {
		_ = db.Close()
	}
	return db, closeFn, nil
}

// OpenProjectDatabase opens the project SQLite database directly.
// Prefer DAO methods; use this only for complex queries that can't be expressed via DAO chain.
func OpenProjectDatabase(ctx context.Context) (*sql.DB, func(), error) {
	return openProjectDatabase(ctx)
}

func nowText() string {
	return time.Now().Format(time.RFC3339)
}

func newResourceID(prefix string) string {
	sequence := atomic.AddUint64(&resourceIDCounter, 1)
	return prefix + "_" + time.Now().UTC().Format("20060102150405.000000000") + "_" + strconv.FormatUint(sequence, 36)
}

func nullIfEmpty(value any) any {
	text := asString(value)
	if strings.TrimSpace(text) == "" {
		return nil
	}
	return text
}

func asString(value any) string {
	switch v := value.(type) {
	case string:
		return v
	default:
		return ""
	}
}

func mustMarshalJSONString(value any, fallback string) string {
	data, err := json.Marshal(value)
	if err != nil {
		return fallback
	}
	return string(data)
}

func isSchemaMissingError(err error) bool {
	if err == nil {
		return false
	}
	text := strings.ToLower(err.Error())
	return strings.Contains(text, "no such table") || strings.Contains(text, "no such column")
}

func marshalDiagnosticDetail(detail map[string]any) any {
	if len(detail) == 0 {
		return nil
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return `{"marshal_error":"diagnostic detail encode failed"}`
	}
	return string(encoded)
}
