package service

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"io"
	"io/fs"
	"mime"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gogf/gf/v2/errors/gerror"

	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

type runArtifactRoots struct {
	RunsRoot   string
	ReplayRoot string
}

type replayArtifactRecord struct {
	ReplayID         string
	ReplayKind       string
	SeqNo            int
	Title            string
	Summary          string
	FilePath         string
	FileExt          string
	MimeType         string
	FileSize         int64
	SHA256           string
	SourceObjectKind string
	SourceObjectID   string
	Status           string
}

type logSegmentRecord struct {
	SegmentID  string
	StreamKind string
	SeqNo      int
	FilePath   string
	FileSize   int64
	SHA256     string
	StartedAt  string
	EndedAt    string
	Status     string
}

type runArtifactFile struct {
	AbsPath  string
	RelPath  string
	FileName string
	Size     int64
	ModTime  string
	SHA256   string
	MimeType string
	FileExt  string
}

func refreshReplayArtifactsForRun(ctx context.Context, binding *entity.BrainRunBindings, runState *BrainRunState) error {
	if binding == nil {
		return nil
	}
	if strings.TrimSpace(binding.ProjectId) == "" || strings.TrimSpace(binding.BrainRunId) == "" {
		return nil
	}

	roots, err := getRunArtifactRoots(ctx, binding.ProjectId)
	if err != nil {
		return err
	}

	replayFiles, logFiles, err := collectRunArtifactFiles(roots, binding.BrainRunId, runExecutionID(runState))
	if err != nil {
		return err
	}

	replayRecords := buildReplayArtifactRecords(binding, replayFiles)
	logRecords := buildLogSegmentRecords(logFiles)
	return replaceRunReplayArtifacts(ctx, binding.ProjectId, binding.TaskId, binding.BrainRunId, replayRecords, logRecords)
}

func runExecutionID(runState *BrainRunState) string {
	if runState == nil {
		return ""
	}
	return strings.TrimSpace(runState.ExecutionID)
}

func getRunArtifactRoots(ctx context.Context, projectID string) (runArtifactRoots, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return runArtifactRoots{}, err
	}
	defer closeFn()

	row := db.QueryRowContext(
		ctx,
		`SELECT COALESCE(runs_root, ''), COALESCE(replay_root, '') FROM `+dao.ProjectWorkspaces.Table()+` WHERE project_id = ? LIMIT 1`,
		projectID,
	)

	var roots runArtifactRoots
	if err = row.Scan(&roots.RunsRoot, &roots.ReplayRoot); err != nil {
		if err == sql.ErrNoRows {
			return runArtifactRoots{}, gerror.Newf("project workspace not found: %s", projectID)
		}
		return runArtifactRoots{}, gerror.Wrap(err, "query project workspace roots failed")
	}
	return roots, nil
}

func collectRunArtifactFiles(roots runArtifactRoots, runID string, executionID string) ([]runArtifactFile, []runArtifactFile, error) {
	candidates := collectRunArtifactCandidateRoots(roots, runID, executionID)
	seenFiles := make(map[string]struct{})
	replayFiles := make([]runArtifactFile, 0, 8)
	logFiles := make([]runArtifactFile, 0, 8)

	for _, root := range candidates {
		if root == "" {
			continue
		}
		info, err := os.Stat(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, nil, gerror.Wrap(err, "stat run artifact root failed")
		}
		if !info.IsDir() {
			continue
		}
		nextReplay, nextLogs, err := scanRunArtifactRoot(root, seenFiles)
		if err != nil {
			return nil, nil, err
		}
		replayFiles = append(replayFiles, nextReplay...)
		logFiles = append(logFiles, nextLogs...)
	}

	sortRunArtifactFiles(replayFiles)
	sortRunArtifactFiles(logFiles)
	return replayFiles, logFiles, nil
}

func collectRunArtifactCandidateRoots(roots runArtifactRoots, runID string, executionID string) []string {
	raw := make([]string, 0, 4)
	if strings.TrimSpace(roots.RunsRoot) != "" && strings.TrimSpace(runID) != "" {
		raw = append(raw, filepath.Join(strings.TrimSpace(roots.RunsRoot), strings.TrimSpace(runID)))
	}
	if strings.TrimSpace(roots.RunsRoot) != "" && strings.TrimSpace(executionID) != "" {
		raw = append(raw, filepath.Join(strings.TrimSpace(roots.RunsRoot), strings.TrimSpace(executionID)))
	}
	if strings.TrimSpace(roots.ReplayRoot) != "" && strings.TrimSpace(runID) != "" {
		raw = append(raw, filepath.Join(strings.TrimSpace(roots.ReplayRoot), strings.TrimSpace(runID)))
	}
	if strings.TrimSpace(roots.ReplayRoot) != "" && strings.TrimSpace(executionID) != "" {
		raw = append(raw, filepath.Join(strings.TrimSpace(roots.ReplayRoot), strings.TrimSpace(executionID)))
	}

	candidates := make([]string, 0, len(raw))
	seen := make(map[string]struct{}, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" || item == "." {
			continue
		}
		item = filepath.Clean(item)
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		candidates = append(candidates, item)
	}
	return candidates
}

func scanRunArtifactRoot(root string, seenFiles map[string]struct{}) ([]runArtifactFile, []runArtifactFile, error) {
	replayFiles := make([]runArtifactFile, 0, 8)
	logFiles := make([]runArtifactFile, 0, 8)

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		normalized := filepath.Clean(path)
		if _, ok := seenFiles[normalized]; ok {
			return nil
		}
		kind := classifyRunArtifactPath(root, normalized)
		if kind == "" {
			return nil
		}
		item, err := buildRunArtifactFile(root, normalized)
		if err != nil {
			return err
		}
		seenFiles[normalized] = struct{}{}
		switch kind {
		case "log":
			logFiles = append(logFiles, item)
		case "replay":
			replayFiles = append(replayFiles, item)
		}
		return nil
	})
	if err != nil {
		return nil, nil, gerror.Wrap(err, "scan run artifact root failed")
	}
	return replayFiles, logFiles, nil
}

func classifyRunArtifactPath(root string, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return ""
	}
	rel = filepath.ToSlash(rel)
	lowerRel := strings.ToLower(rel)
	name := strings.ToLower(filepath.Base(path))

	if strings.HasPrefix(lowerRel, "checkpoints/") || lowerRel == "meta.json" {
		return ""
	}
	if strings.HasPrefix(lowerRel, "logs/") || strings.Contains(name, "stdout") || strings.Contains(name, "stderr") || strings.HasSuffix(name, ".log") {
		return "log"
	}
	if strings.HasPrefix(lowerRel, "replay/") || strings.HasPrefix(lowerRel, "artifacts/") || strings.Contains(name, "_replay_") || strings.Contains(name, "tool-call") || strings.Contains(name, "tool_result") {
		return "replay"
	}
	return ""
}

func buildRunArtifactFile(root string, path string) (runArtifactFile, error) {
	info, err := os.Stat(path)
	if err != nil {
		return runArtifactFile{}, gerror.Wrap(err, "stat run artifact file failed")
	}
	if info.IsDir() {
		return runArtifactFile{}, gerror.New("artifact path is a directory")
	}
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return runArtifactFile{}, gerror.Wrap(err, "resolve artifact relative path failed")
	}
	sum, err := computeFileSHA256(path)
	if err != nil {
		return runArtifactFile{}, err
	}
	fileExt := strings.ToLower(filepath.Ext(path))
	mimeType := mime.TypeByExtension(fileExt)
	if mimeType == "" && (fileExt == ".log" || fileExt == ".txt" || fileExt == ".jsonl") {
		mimeType = "text/plain; charset=utf-8"
	}
	return runArtifactFile{
		AbsPath:  path,
		RelPath:  filepath.ToSlash(rel),
		FileName: filepath.Base(path),
		Size:     info.Size(),
		ModTime:  info.ModTime().UTC().Format("2006-01-02T15:04:05Z07:00"),
		SHA256:   sum,
		MimeType: mimeType,
		FileExt:  fileExt,
	}, nil
}

func computeFileSHA256(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", gerror.Wrap(err, "open artifact file failed")
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err = io.Copy(hasher, file); err != nil {
		return "", gerror.Wrap(err, "hash artifact file failed")
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func sortRunArtifactFiles(items []runArtifactFile) {
	sort.SliceStable(items, func(left, right int) bool {
		if items[left].RelPath == items[right].RelPath {
			return items[left].FileName < items[right].FileName
		}
		return items[left].RelPath < items[right].RelPath
	})
}

func buildReplayArtifactRecords(binding *entity.BrainRunBindings, files []runArtifactFile) []replayArtifactRecord {
	records := make([]replayArtifactRecord, 0, len(files))
	for index, item := range files {
		seqNo := index + 1
		records = append(records, replayArtifactRecord{
			ReplayID:         buildReplayArtifactID(binding.BrainRunId, seqNo),
			ReplayKind:       classifyReplayKind(item),
			SeqNo:            seqNo,
			Title:            buildReplayArtifactTitle(item),
			Summary:          item.RelPath,
			FilePath:         item.AbsPath,
			FileExt:          item.FileExt,
			MimeType:         item.MimeType,
			FileSize:         item.Size,
			SHA256:           item.SHA256,
			SourceObjectKind: "brain_run",
			SourceObjectID:   binding.BrainRunId,
			Status:           "available",
		})
	}
	return records
}

func buildLogSegmentRecords(files []runArtifactFile) []logSegmentRecord {
	records := make([]logSegmentRecord, 0, len(files))
	for index, item := range files {
		seqNo := index + 1
		records = append(records, logSegmentRecord{
			SegmentID:  buildLogSegmentID(item, seqNo),
			StreamKind: classifyLogStreamKind(item),
			SeqNo:      seqNo,
			FilePath:   item.AbsPath,
			FileSize:   item.Size,
			SHA256:     item.SHA256,
			StartedAt:  item.ModTime,
			EndedAt:    item.ModTime,
			Status:     "available",
		})
	}
	return records
}

func buildReplayArtifactID(runID string, seqNo int) string {
	return strings.TrimSpace(runID) + "_replay_" + strconv.Itoa(seqNo)
}

func buildLogSegmentID(item runArtifactFile, seqNo int) string {
	base := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	base = strings.ReplaceAll(base, " ", "_")
	if base == "" {
		base = "segment"
	}
	return base + "_" + strconv.Itoa(seqNo)
}

func buildReplayArtifactTitle(item runArtifactFile) string {
	name := strings.TrimSuffix(item.FileName, filepath.Ext(item.FileName))
	name = strings.ReplaceAll(name, "_", " ")
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.TrimSpace(name)
	if name == "" {
		return item.FileName
	}
	return name
}

func classifyReplayKind(item runArtifactFile) string {
	value := strings.ToLower(item.RelPath + " " + item.FileName)
	switch {
	case strings.Contains(value, "tool-call"), strings.Contains(value, "tool_call"):
		return "tool_call"
	case strings.Contains(value, "tool-result"), strings.Contains(value, "tool_result"):
		return "tool_result"
	case strings.Contains(value, "browser"):
		return "browser_trace"
	case strings.Contains(value, "verification"):
		return "verification_snapshot"
	case strings.Contains(value, "thought"):
		return "thought_summary"
	case strings.Contains(value, "capture"):
		return "runtime_capture"
	default:
		return "step_snapshot"
	}
}

func classifyLogStreamKind(item runArtifactFile) string {
	value := strings.ToLower(item.RelPath + " " + item.FileName)
	switch {
	case strings.Contains(value, "stderr"):
		return "stderr"
	case strings.Contains(value, "system"):
		return "system"
	case strings.Contains(value, "tool"):
		return "tool"
	default:
		return "stdout"
	}
}

func replaceRunReplayArtifacts(ctx context.Context, projectID string, taskID string, runID string, replayRecords []replayArtifactRecord, logRecords []logSegmentRecord) error {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return err
	}
	defer closeFn()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return gerror.Wrap(err, "begin replay artifact transaction failed")
	}

	if err = deleteRunReplayArtifacts(ctx, tx, projectID, runID); err != nil {
		_ = tx.Rollback()
		return err
	}
	for _, item := range replayRecords {
		if err = insertReplayArtifactRecord(ctx, tx, projectID, taskID, runID, item); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	for _, item := range logRecords {
		if err = insertLogSegmentRecord(ctx, tx, projectID, runID, item); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err = tx.Commit(); err != nil {
		return gerror.Wrap(err, "commit replay artifact transaction failed")
	}
	return nil
}

func deleteRunReplayArtifacts(ctx context.Context, tx *sql.Tx, projectID string, runID string) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM workflow_replay_index WHERE project_id = ? AND run_id = ?`, projectID, runID); err != nil {
		return gerror.Wrap(err, "delete replay index rows failed")
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM workflow_run_log_segments WHERE project_id = ? AND run_id = ?`, projectID, runID); err != nil {
		return gerror.Wrap(err, "delete log segment rows failed")
	}
	return nil
}

func insertReplayArtifactRecord(ctx context.Context, tx *sql.Tx, projectID string, taskID string, runID string, item replayArtifactRecord) error {
	now := nowText()
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO workflow_replay_index (
id, replay_id, project_id, run_id, domain_task_id, compiled_task_id, event_id, trace_id, span_id,
replay_kind, seq_no, title, summary, file_path, file_ext, mime_type, file_size, sha256,
source_object_kind, source_object_id, status, created_at, updated_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("wfreplay"),
		item.ReplayID,
		projectID,
		runID,
		nullIfEmpty(taskID),
		nil,
		nil,
		nil,
		nil,
		item.ReplayKind,
		item.SeqNo,
		item.Title,
		nullIfEmpty(item.Summary),
		item.FilePath,
		nullIfEmpty(item.FileExt),
		nullIfEmpty(item.MimeType),
		item.FileSize,
		nullIfEmpty(item.SHA256),
		nullIfEmpty(item.SourceObjectKind),
		nullIfEmpty(item.SourceObjectID),
		item.Status,
		now,
		now,
	)
	if err != nil {
		return gerror.Wrap(err, "insert replay artifact record failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert replay artifact affected unexpected rows")
	}
	return nil
}

func insertLogSegmentRecord(ctx context.Context, tx *sql.Tx, projectID string, runID string, item logSegmentRecord) error {
	result, err := tx.ExecContext(
		ctx,
		`INSERT INTO workflow_run_log_segments (
id, project_id, run_id, segment_id, stream_kind, seq_no, file_path, file_size, sha256, started_at, ended_at, status, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		newResourceID("wflog"),
		projectID,
		runID,
		item.SegmentID,
		item.StreamKind,
		item.SeqNo,
		item.FilePath,
		item.FileSize,
		nullIfEmpty(item.SHA256),
		nullIfEmpty(item.StartedAt),
		nullIfEmpty(item.EndedAt),
		item.Status,
		nowText(),
	)
	if err != nil {
		return gerror.Wrap(err, "insert log segment record failed")
	}
	if affected, _ := result.RowsAffected(); affected != 1 {
		return gerror.New("insert log segment affected unexpected rows")
	}
	return nil
}
