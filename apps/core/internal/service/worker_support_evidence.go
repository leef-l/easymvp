package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
)

type evidenceScanWorker struct {
	interval time.Duration
}

func newEvidenceScanWorker() backgroundWorker {
	cfgInterval := g.Cfg().MustGet(context.Background(), "easymvp.workers.evidenceScanInterval", "30s").Duration()
	if cfgInterval <= 0 {
		cfgInterval = 30 * time.Second
	}
	return &evidenceScanWorker{
		interval: cfgInterval,
	}
}

func (w *evidenceScanWorker) Name() string {
	return "evidence_scan_worker"
}

func (w *evidenceScanWorker) Interval() time.Duration {
	return w.interval
}

func (w *evidenceScanWorker) RunOnce(ctx context.Context) error {
	projectIDs, err := listAllProjectIDs(ctx)
	if err != nil {
		return err
	}

	for _, projectID := range projectIDs {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if _, err := Evidence().CollectProjectEvidence(ctx, projectID); err != nil {
			handleWorkerFailure(
				ctx,
				w.Name(),
				projectID,
				"WORKER_EVIDENCE_SCAN",
				"evidence scan failed for project",
				map[string]any{
					"project_id": projectID,
					"error":      err.Error(),
				},
			)
			continue
		}
	}

	if len(projectIDs) > 0 {
		g.Log().Debugf(ctx, "[worker:%s] scanned %d projects", w.Name(), len(projectIDs))
	}
	return nil
}

func listAllProjectIDs(ctx context.Context) ([]string, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(ctx, `SELECT id FROM `+dao.Projects.Table()+` ORDER BY updated_at DESC`)
	if err != nil {
		return nil, gerror.Wrap(err, "query all project ids failed")
	}
	defer rows.Close()

	ids := make([]string, 0)
	for rows.Next() {
		var id string
		if err = rows.Scan(&id); err != nil {
			return nil, gerror.Wrap(err, "scan project id failed")
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate project ids failed")
	}
	return ids, nil
}

func collectProjectEvidence(ctx context.Context, projectID string) (int, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return 0, err
	}
	defer closeFn()

	evidenceRoot, err := getProjectEvidenceRoot(ctx, db, projectID)
	if err != nil {
		return 0, err
	}
	if evidenceRoot == "" {
		return 0, nil
	}

	if _, err := os.Stat(evidenceRoot); os.IsNotExist(err) {
		return 0, nil
	}

	latestRun, err := getLatestAcceptanceRunByProjectID(ctx, db, projectID)
	if err != nil {
		return 0, err
	}

	runID := ""
	if latestRun != nil {
		runID = latestRun.Id
	}

	insertedCount := 0
	surfaceCounts := make(map[string]int)
	journeyCounts := make(map[string]int)

	err = filepath.Walk(evidenceRoot, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(evidenceRoot, path)
		parts := strings.Split(filepath.ToSlash(relPath), "/")
		surface := "default"
		journey := ""
		if len(parts) > 1 {
			surface = parts[0]
		}
		if len(parts) > 2 {
			journey = parts[1]
		}

		exists, err := evidenceItemExists(ctx, db, path)
		if err != nil {
			return nil
		}
		if exists {
			return nil
		}

		fileSize := int(info.Size())
		capturedAt := info.ModTime().Format(time.RFC3339)
		contentHash := computeFileHash(path)
		evidenceType := inferEvidenceType(path)

		if err := insertEvidenceItem(ctx, db, projectID, runID, surface, journey, evidenceType, path, contentHash, fileSize, capturedAt); err != nil {
			return nil
		}

		insertedCount++
		if surface != "" {
			surfaceCounts[surface]++
		}
		if journey != "" {
			journeyCounts[journey]++
		}

		return nil
	})

	if err != nil {
		return 0, gerror.Wrap(err, "walk evidence root failed")
	}

	if runID != "" && insertedCount > 0 {
		for surface, count := range surfaceCounts {
			if err := upsertSurfaceCoverageEvidenceCount(ctx, db, projectID, runID, surface, count); err != nil {
				g.Log().Warningf(ctx, "update surface coverage failed: %v", err)
			}
		}
		for journey, count := range journeyCounts {
			if err := upsertJourneyCoverageEvidenceCount(ctx, db, projectID, runID, journey, count); err != nil {
				g.Log().Warningf(ctx, "update journey coverage failed: %v", err)
			}
		}
	}

	if insertedCount > 0 {
		if err := insertAuditLog(ctx, projectID, "evidence.collected", "worker:evidence_scan_worker",
			fmt.Sprintf("Collected %d evidence items", insertedCount),
			map[string]any{"inserted_count": insertedCount}); err != nil {
			g.Log().Warningf(ctx, "insert evidence audit log failed: %v", err)
		}
	}

	return insertedCount, nil
}

func getProjectEvidenceRoot(ctx context.Context, db *sql.DB, projectID string) (string, error) {
	row := db.QueryRowContext(ctx, `SELECT evidence_root FROM `+dao.ProjectWorkspaces.Table()+` WHERE project_id = ? LIMIT 1`, projectID)
	var root string
	if err := row.Scan(&root); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", gerror.Wrap(err, "query evidence root failed")
	}
	return root, nil
}

func evidenceItemExists(ctx context.Context, db *sql.DB, filePath string) (bool, error) {
	row := db.QueryRowContext(ctx, `SELECT 1 FROM `+dao.EvidenceItems.Table()+` WHERE file_path = ? LIMIT 1`, filePath)
	var one int
	if err := row.Scan(&one); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func insertEvidenceItem(ctx context.Context, db *sql.DB, projectID, runID, surface, journey, evidenceType, filePath, contentHash string, fileSize int, capturedAt string) error {
	_, err := db.ExecContext(ctx,
		`INSERT INTO `+dao.EvidenceItems.Table()+` (id, project_id, run_id, surface, journey, evidence_type, file_path, content_hash, file_size, captured_at, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("evidence"),
		projectID,
		nullIfEmpty(runID),
		nullIfEmpty(surface),
		nullIfEmpty(journey),
		evidenceType,
		filePath,
		nullIfEmpty(contentHash),
		fileSize,
		capturedAt,
		nowText(),
	)
	return err
}

func upsertSurfaceCoverageEvidenceCount(ctx context.Context, db *sql.DB, projectID, runID, surface string, delta int) error {
	now := nowText()
	_, err := db.ExecContext(ctx,
		`INSERT INTO `+dao.AcceptanceSurfaceCoverage.Table()+` (id, project_id, acceptance_run_id, surface, coverage_status, evidence_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(acceptance_run_id, surface) DO UPDATE SET
			evidence_count = evidence_count + excluded.evidence_count,
			updated_at = excluded.updated_at`,
		newResourceID("asc"),
		projectID,
		runID,
		surface,
		"pending",
		delta,
		now,
		now,
	)
	return err
}

func upsertJourneyCoverageEvidenceCount(ctx context.Context, db *sql.DB, projectID, runID, journey string, delta int) error {
	now := nowText()
	_, err := db.ExecContext(ctx,
		`INSERT INTO `+dao.AcceptanceJourneyCoverage.Table()+` (id, project_id, acceptance_run_id, journey, coverage_status, evidence_count, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(acceptance_run_id, journey) DO UPDATE SET
			evidence_count = evidence_count + excluded.evidence_count,
			updated_at = excluded.updated_at`,
		newResourceID("ajc"),
		projectID,
		runID,
		journey,
		"pending",
		delta,
		now,
		now,
	)
	return err
}

func inferEvidenceType(path string) string {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(path), "."))
	switch ext {
	case "png", "jpg", "jpeg", "gif", "webp", "bmp", "svg":
		return "screenshot"
	case "json":
		return "artifact"
	case "txt", "log", "md":
		return "log"
	case "mp4", "webm", "mov", "avi":
		return "video"
	case "pdf":
		return "document"
	default:
		return "file"
	}
}

func computeFileHash(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))[:16]
}
