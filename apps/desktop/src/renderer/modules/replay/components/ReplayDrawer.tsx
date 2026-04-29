import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { apiGet } from "@/shared/lib/api";
import { useQuery } from "@/shared/lib/query";
import type {
  ExecutionView,
  LogSegmentRawView,
  ReplayDetailView,
  ReplayRawView,
} from "@/shared/lib/types";

type ReplayDrawerProps = {
  projectId: string;
  runId?: string;
  bindingId?: string;
  isOpen: boolean;
  onClose: () => void;
};

export function ReplayDrawer(props: ReplayDrawerProps) {
  const { t } = useTranslation();
  const { projectId, runId, isOpen, onClose } = props;
  const [refreshTick, setRefreshTick] = useState(0);
  const [selectedReplayId, setSelectedReplayId] = useState("");
  const [selectedSegmentId, setSelectedSegmentId] = useState("");
  const [expandedRawView, setExpandedRawView] = useState<{
    title: string;
    subtitle: string;
    preview: string;
  } | null>(null);

  const replaySummaryState = useQuery(
    () =>
      runId
        ? apiGet<ExecutionView["replay_summary"]>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/replay-summary`,
          )
        : Promise.resolve(null),
    [projectId, runId, refreshTick],
  );

  const replayTimelineState = useQuery(
    () =>
      runId
        ? apiGet<{ items: ExecutionView["replay_timeline"] }>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/replay-timeline?limit=10`,
          )
        : Promise.resolve(null),
    [projectId, runId, refreshTick],
  );

  const logSegmentsState = useQuery(
    () =>
      runId
        ? apiGet<{ segments: ExecutionView["log_segments"] }>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/log-segments?limit=10`,
          )
        : Promise.resolve(null),
    [projectId, runId, refreshTick],
  );

  useEffect(() => {
    if (!isOpen) {
      setSelectedReplayId("");
      setSelectedSegmentId("");
      return;
    }
    setRefreshTick((v) => v + 1);
  }, [isOpen, runId]);

  useEffect(() => {
    if (!selectedReplayId && (replayTimelineState.data?.items?.length ?? 0) > 0) {
      setSelectedReplayId(replayTimelineState.data!.items[0].replay_id);
    }
    if (!selectedSegmentId && (logSegmentsState.data?.segments?.length ?? 0) > 0) {
      setSelectedSegmentId(logSegmentsState.data!.segments[0].segment_id);
    }
  }, [replayTimelineState.data, logSegmentsState.data, selectedReplayId, selectedSegmentId]);

  const replayDetailState = useQuery(
    () =>
      selectedReplayId && runId
        ? apiGet<ReplayDetailView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/replay-items/${encodeURIComponent(selectedReplayId)}`,
          )
        : Promise.resolve(null),
    [projectId, runId, selectedReplayId, refreshTick],
  );

  const replayRawState = useQuery(
    () =>
      selectedReplayId && runId
        ? apiGet<ReplayRawView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/replay-items/${encodeURIComponent(selectedReplayId)}/raw?limit=12000`,
          )
        : Promise.resolve(null),
    [projectId, runId, selectedReplayId, refreshTick],
  );

  const logRawState = useQuery(
    () =>
      selectedSegmentId && runId
        ? apiGet<LogSegmentRawView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(runId)}/log-segments/${encodeURIComponent(selectedSegmentId)}/raw?limit=12000`,
          )
        : Promise.resolve(null),
    [projectId, runId, selectedSegmentId, refreshTick],
  );

  const replayPreview = useMemo(() => {
    return formatRawPreview(replayRawState.data?.content, replayRawState.data?.truncated);
  }, [replayRawState.data]);

  const logPreview = useMemo(() => {
    return formatRawPreview(logRawState.data?.content, logRawState.data?.truncated);
  }, [logRawState.data]);

  if (!isOpen) return null;

  return (
    <>
      <div className="replay-drawer-overlay" onClick={onClose} />
      <div className="replay-drawer">
        <div className="replay-drawer-header">
          <div>
            <p className="placeholder-section">{t("replay.title")}</p>
            <h3 className="placeholder-title">{runId || t("replay.noRun")}</h3>
          </div>
          <div className="action-row">
            <button className="secondary-button" onClick={() => setRefreshTick((v) => v + 1)}>
              {t("execution.refreshNow")}
            </button>
            <button className="secondary-button" onClick={onClose}>
              {t("execution.close")}
            </button>
          </div>
        </div>

        <div className="replay-drawer-body">
          {replaySummaryState.data ? (
            <div className="replay-drawer-summary">
              <span className="status-pill">{replaySummaryState.data.status}</span>
              <span className="status-pill">
                events {replaySummaryState.data.event_count}
              </span>
              <span className="status-pill">
                replay {replaySummaryState.data.replay_count}
              </span>
              <span className="status-pill">
                logs {replaySummaryState.data.log_segment_count}
              </span>
            </div>
          ) : null}

          <div className="replay-drawer-columns">
            <div className="replay-drawer-timeline">
              <h4>{t("execution.replayTimeline")}</h4>
              <div className="stack-list">
                {(replayTimelineState.data?.items ?? []).map((item) => (
                  <button
                    key={item.replay_id}
                    className={
                      item.replay_id === selectedReplayId
                        ? "action-card is-selected"
                        : "action-card"
                    }
                    onClick={() => {
                      setSelectedReplayId(item.replay_id);
                      setSelectedSegmentId("");
                    }}
                  >
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      <span className="status-pill">{item.replay_type}</span>
                    </div>
                    <p>
                      seq {item.seq_no} · {item.summary || item.status}
                    </p>
                  </button>
                ))}
                {(replayTimelineState.data?.items?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>{t("execution.noReplayItems")}</p>
                  </article>
                ) : null}
              </div>

              <h4 style={{ marginTop: 16 }}>{t("execution.logSegments")}</h4>
              <div className="stack-list">
                {(logSegmentsState.data?.segments ?? []).map((item) => (
                  <button
                    key={item.segment_id}
                    className={
                      item.segment_id === selectedSegmentId
                        ? "action-card is-selected"
                        : "action-card"
                    }
                    onClick={() => setSelectedSegmentId(item.segment_id)}
                  >
                    <div className="list-card-head">
                      <strong>{item.segment_id}</strong>
                      <span className="status-pill">{item.stream_kind}</span>
                    </div>
                    <p>
                      {item.status} · seq {item.seq_no} · {item.size} bytes
                    </p>
                  </button>
                ))}
                {(logSegmentsState.data?.segments?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>{t("execution.noLogSegments")}</p>
                  </article>
                ) : null}
              </div>
            </div>

            <div className="replay-drawer-detail">
              {replayDetailState.data ? (
                <div className="stack-list">
                  <article className="list-card">
                    <div className="list-card-head">
                      <strong>{replayDetailState.data.title}</strong>
                      <span className="status-pill">
                        {replayDetailState.data.replay_kind}
                      </span>
                    </div>
                    <p>{replayDetailState.data.summary || replayDetailState.data.status}</p>
                    <p>
                      domain {replayDetailState.data.domain_task_id || "n/a"} · compiled{" "}
                      {replayDetailState.data.compiled_task_id || "n/a"}
                    </p>
                  </article>

                  <RawPreviewCard
                    title={t("execution.replayRaw")}
                    subtitle={replayDetailState.data.title}
                    preview={replayPreview}
                    rawContent={replayRawState.data?.content}
                    rawError={replayRawState.error}
                    onCopy={() => {
                      if (replayRawState.data?.content) {
                        void copyTextToClipboard(replayRawState.data.content);
                      }
                    }}
                    onDownload={() => {
                      if (replayRawState.data?.content) {
                        downloadRawContent(
                          selectedReplayId || "replay-raw",
                          replayRawState.data.content,
                          replayRawState.data.mime_type,
                        );
                      }
                    }}
                    onExpand={() =>
                      setExpandedRawView({
                        title: t("execution.replayRaw"),
                        subtitle: replayDetailState.data!.title,
                        preview: replayPreview,
                      })
                    }
                  />

                  <RawPreviewCard
                    title={t("execution.logRaw")}
                    subtitle={selectedSegmentId || t("execution.selectedLogSegment")}
                    preview={logPreview}
                    rawContent={logRawState.data?.content}
                    rawError={logRawState.error}
                    onCopy={() => {
                      if (logRawState.data?.content) {
                        void copyTextToClipboard(logRawState.data.content);
                      }
                    }}
                    onDownload={() => {
                      if (logRawState.data?.content) {
                        downloadRawContent(
                          selectedSegmentId || "log-raw",
                          logRawState.data.content,
                        );
                      }
                    }}
                    onExpand={() =>
                      setExpandedRawView({
                        title: t("execution.logRaw"),
                        subtitle: selectedSegmentId || t("execution.selectedLogSegment"),
                        preview: logPreview,
                      })
                    }
                  />
                </div>
              ) : (
                <article className="list-card">
                  <p>{t("execution.selectReplayItem")}</p>
                </article>
              )}
            </div>
          </div>
        </div>
      </div>

      {expandedRawView ? (
        <div
          className="raw-overlay"
          role="dialog"
          aria-modal="true"
          onClick={() => setExpandedRawView(null)}
        >
          <div className="raw-overlay-panel" onClick={(e) => e.stopPropagation()}>
            <div className="raw-overlay-header">
              <div>
                <p className="placeholder-section">{t("execution.expandedPreview")}</p>
                <h3 className="placeholder-title">{expandedRawView.title}</h3>
                <p className="placeholder-description">{expandedRawView.subtitle}</p>
              </div>
              <button className="secondary-button" onClick={() => setExpandedRawView(null)}>
                {t("execution.close")}
              </button>
            </div>
            <pre className="json-block">{expandedRawView.preview}</pre>
          </div>
        </div>
      ) : null}
    </>
  );
}

function RawPreviewCard(props: {
  title: string;
  subtitle: string;
  preview: string;
  rawContent?: string;
  rawError?: string;
  onCopy: () => void;
  onDownload: () => void;
  onExpand: () => void;
}) {
  const { t } = useTranslation();
  return (
    <article className="list-card raw-preview-card">
      <div className="list-card-head raw-preview-header">
        <div>
          <h3>{props.title}</h3>
          <p className="raw-preview-note">{props.subtitle}</p>
        </div>
      </div>
      {props.rawError ? <p className="error-copy">{props.rawError}</p> : null}
      <pre className="json-block">{props.preview}</pre>
      <div className="raw-actions">
        <button className="secondary-button" onClick={props.onCopy} disabled={!props.rawContent}>
          {t("execution.copy")}
        </button>
        <button className="secondary-button" onClick={props.onDownload} disabled={!props.rawContent}>
          {t("execution.download")}
        </button>
        <button className="secondary-button" onClick={props.onExpand} disabled={!props.rawContent}>
          {t("execution.expand")}
        </button>
      </div>
    </article>
  );
}

function formatRawPreview(content?: string, truncated?: boolean) {
  if (!content) return "No content available.";
  const suffix = truncated ? "\n\n[truncated]" : "";
  return content + suffix;
}

async function copyTextToClipboard(text: string) {
  try {
    await navigator.clipboard.writeText(text);
  } catch {
    // ignore
  }
}

function downloadRawContent(id: string, content: string, mimeType?: string) {
  const blob = new Blob([content], { type: mimeType || "text/plain" });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = `${id}.txt`;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
