package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"

	workspacev1 "github.com/leef-l/easymvp/apps/core/api/workspace/v1"
	"github.com/leef-l/easymvp/apps/core/internal/dao"
	"github.com/leef-l/easymvp/apps/core/internal/model/entity"
)

const (
	workspaceEventPollInterval      = 2 * time.Second
	workspaceEventKeepaliveInterval = 15 * time.Second
	workspaceEventDefaultLimit      = 50
	workspaceEventMaxLimit          = 200
)

type normalizedProjectEventsStreamReq struct {
	ProjectID   string
	Limit       int
	LastEventID string
}

type workspaceEventCursor struct {
	EventID   string
	CreatedAt string
}

func streamWorkspaceProjectEvents(ctx context.Context, req *workspacev1.ProjectEventsStreamReq) error {
	httpReq := g.RequestFromCtx(ctx)
	if httpReq == nil {
		return gerror.New("workspace event stream request context is unavailable")
	}
	flusher, ok := httpReq.Response.RawWriter().(http.Flusher)
	if !ok {
		return gerror.New("workspace event stream flusher is unavailable")
	}

	normalized := normalizeProjectEventsStreamReq(httpReq, req)
	httpReq.Response.RawWriter().Header().Set("Content-Type", "text/event-stream")
	httpReq.Response.RawWriter().Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	httpReq.Response.RawWriter().Header().Set("Connection", "keep-alive")
	httpReq.Response.RawWriter().Header().Set("X-Accel-Buffering", "no")
	httpReq.Response.RawWriter().WriteHeader(http.StatusOK)

	if err := writeWorkspaceSSEControlFrame(httpReq, "retry: 3000\n\n"); err != nil {
		return nil
	}
	flusher.Flush()

	cursor := workspaceEventCursor{}
	if normalized.LastEventID != "" {
		resolvedCursor, found, err := getWorkspaceEventCursor(ctx, normalized.ProjectID, normalized.LastEventID)
		if err != nil {
			g.Log().Warningf(ctx, "workspace event stream cursor lookup failed: %v", err)
			return nil
		}
		if found {
			cursor = resolvedCursor
		} else {
			if err = writeWorkspaceSSEEvent(httpReq, "", "workspace.snapshot_invalidated", map[string]any{
				"project_id":      normalized.ProjectID,
				"last_event_id":   normalized.LastEventID,
				"recommended_act": "refresh_project_workspace",
			}); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}

	emitEvents := func() error {
		events, err := listWorkspaceProjectStreamEvents(ctx, normalized.ProjectID, cursor, normalized.Limit)
		if err != nil {
			return err
		}
		for _, item := range events {
			if err = writeWorkspaceSSEEvent(httpReq, item.Id, item.EventType, mapWorkspaceProjectStreamEvent(item)); err != nil {
				return nil
			}
			cursor = workspaceEventCursor{
				EventID:   item.Id,
				CreatedAt: item.CreatedAt,
			}
		}
		if len(events) > 0 {
			flusher.Flush()
		}
		return nil
	}

	if err := emitEvents(); err != nil {
		g.Log().Warningf(ctx, "workspace event stream initial emit failed: %v", err)
		return nil
	}

	pollTicker := time.NewTicker(workspaceEventPollInterval)
	keepaliveTicker := time.NewTicker(workspaceEventKeepaliveInterval)
	defer pollTicker.Stop()
	defer keepaliveTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-httpReq.Request.Context().Done():
			return nil
		case <-pollTicker.C:
			if err := emitEvents(); err != nil {
				g.Log().Warningf(ctx, "workspace event stream polling failed: %v", err)
				return nil
			}
		case <-keepaliveTicker.C:
			if err := writeWorkspaceSSEControlFrame(httpReq, ": keepalive\n\n"); err != nil {
				return nil
			}
			flusher.Flush()
		}
	}
}

func normalizeProjectEventsStreamReq(httpReq *ghttp.Request, req *workspacev1.ProjectEventsStreamReq) normalizedProjectEventsStreamReq {
	normalized := normalizedProjectEventsStreamReq{
		ProjectID:   strings.TrimSpace(req.ProjectID),
		Limit:       req.Limit,
		LastEventID: strings.TrimSpace(req.LastEventID),
	}
	if normalized.Limit <= 0 {
		normalized.Limit = workspaceEventDefaultLimit
	}
	if normalized.Limit > workspaceEventMaxLimit {
		normalized.Limit = workspaceEventMaxLimit
	}
	if normalized.LastEventID == "" {
		normalized.LastEventID = strings.TrimSpace(httpReq.GetHeader("Last-Event-ID"))
	}
	return normalized
}

func getWorkspaceEventCursor(ctx context.Context, projectID, eventID string) (workspaceEventCursor, bool, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return workspaceEventCursor{}, false, err
	}
	defer closeFn()

	row := db.QueryRowContext(
		ctx,
		`SELECT id, created_at
FROM `+dao.RunEventIndex.Table()+`
WHERE project_id = ? AND id = ?
LIMIT 1`,
		projectID,
		eventID,
	)

	var cursor workspaceEventCursor
	if err = row.Scan(&cursor.EventID, &cursor.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return workspaceEventCursor{}, false, nil
		}
		return workspaceEventCursor{}, false, gerror.Wrap(err, "query workspace event cursor failed")
	}
	return cursor, true, nil
}

func listWorkspaceProjectStreamEvents(ctx context.Context, projectID string, cursor workspaceEventCursor, limit int) ([]entity.RunEventIndex, error) {
	db, closeFn, err := openProjectDatabase(ctx)
	if err != nil {
		return nil, err
	}
	defer closeFn()

	rows, err := db.QueryContext(
		ctx,
		`SELECT
  id,
  project_id,
  run_binding_id,
  sequence_no,
  event_type,
  COALESCE(event_level, ''),
  summary,
  COALESCE(payload_json, ''),
  created_at
FROM `+dao.RunEventIndex.Table()+`
WHERE project_id = ?
  AND (
    ? = ''
    OR created_at > ?
    OR (created_at = ? AND id > ?)
  )
ORDER BY created_at ASC, id ASC
LIMIT ?`,
		projectID,
		cursor.EventID,
		cursor.CreatedAt,
		cursor.CreatedAt,
		cursor.EventID,
		limit,
	)
	if err != nil {
		return nil, gerror.Wrap(err, "query workspace project stream events failed")
	}
	defer rows.Close()

	items := make([]entity.RunEventIndex, 0, limit)
	for rows.Next() {
		var item entity.RunEventIndex
		if err = rows.Scan(
			&item.Id,
			&item.ProjectId,
			&item.RunBindingId,
			&item.SequenceNo,
			&item.EventType,
			&item.EventLevel,
			&item.Summary,
			&item.PayloadJson,
			&item.CreatedAt,
		); err != nil {
			return nil, gerror.Wrap(err, "scan workspace project stream event failed")
		}
		items = append(items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, gerror.Wrap(err, "iterate workspace project stream events failed")
	}
	return items, nil
}

func mapWorkspaceProjectStreamEvent(item entity.RunEventIndex) workspacev1.ProjectStreamEvent {
	return workspacev1.ProjectStreamEvent{
		EventID:      item.Id,
		ProjectID:    item.ProjectId,
		RunBindingID: item.RunBindingId,
		SequenceNo:   item.SequenceNo,
		EventType:    item.EventType,
		EventLevel:   item.EventLevel,
		Summary:      item.Summary,
		Payload:      decodeWorkspaceProjectEventPayload(item.PayloadJson),
		CreatedAt:    item.CreatedAt,
	}
}

func decodeWorkspaceProjectEventPayload(raw string) interface{} {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	var payload interface{}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return map[string]any{"raw": raw}
	}
	return payload
}

func writeWorkspaceSSEEvent(httpReq *ghttp.Request, eventID, eventName string, data interface{}) error {
	if eventName == "" {
		eventName = "workspace.event"
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return gerror.Wrap(err, "marshal workspace sse payload failed")
	}

	var frame strings.Builder
	if eventID != "" {
		frame.WriteString("id: ")
		frame.WriteString(cleanWorkspaceSSELine(eventID))
		frame.WriteString("\n")
	}
	frame.WriteString("event: ")
	frame.WriteString(cleanWorkspaceSSELine(eventName))
	frame.WriteString("\n")
	for _, line := range strings.Split(string(payload), "\n") {
		frame.WriteString("data: ")
		frame.WriteString(line)
		frame.WriteString("\n")
	}
	frame.WriteString("\n")
	_, err = httpReq.Response.RawWriter().Write([]byte(frame.String()))
	return err
}

func writeWorkspaceSSEControlFrame(httpReq *ghttp.Request, frame string) error {
	_, err := httpReq.Response.RawWriter().Write([]byte(frame))
	return err
}

func cleanWorkspaceSSELine(value string) string {
	return strings.NewReplacer("\r", " ", "\n", " ").Replace(strings.TrimSpace(value))
}
