import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiDelete, apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, ExecutionView, LogSegmentRawView, ReplayDetailView, ReplayRawView, StartRunResponse, SyncRunResponse } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

const executionViewLimits = {
  eventLimit: 20,
  replayLimit: 10,
  logLimit: 10,
} as const;

export function ExecutionPage() {
  const navigate = useNavigate();
  const { projectId, searchParams, routes, buildRoute } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [expandedRawView, setExpandedRawView] = useState<ExpandedRawView | null>(null);
  const bindingFromUrl = searchParams.get("binding")?.trim() || "";
  const replayFromUrl = searchParams.get("replay")?.trim() || "";
  const segmentFromUrl = searchParams.get("segment")?.trim() || "";
  const taskFromUrl = searchParams.get("task")?.trim() || "";
  const [selectedBindingId, setSelectedBindingId] = useState(bindingFromUrl);
  const state = useQuery(
    () =>
      apiGet<ExecutionView>(
        buildExecutionViewPath(projectId, selectedBindingId, executionViewLimits.eventLimit, executionViewLimits.replayLimit, executionViewLimits.logLimit),
      ),
    [projectId, selectedBindingId, refreshTick],
  );
  const [selectedReplayId, setSelectedReplayId] = useState(replayFromUrl);
  const [selectedSegmentId, setSelectedSegmentId] = useState(segmentFromUrl);
  const [runStatusFilter, setRunStatusFilter] = useState("all");
  const [replayTypeFilter, setReplayTypeFilter] = useState("all");
  const [eventLevelFilter, setEventLevelFilter] = useState("all");
  const [startTaskId, setStartTaskId] = useState("");
  const [startPrompt, setStartPrompt] = useState("Review the assigned task and execute the required implementation steps.");
  const [startBrainKind, setStartBrainKind] = useState("coder");
  const [startProvider, setStartProvider] = useState("");
  const [mutationBusy, setMutationBusy] = useState("");
  const [mutationError, setMutationError] = useState("");
  const [mutationMessage, setMutationMessage] = useState("");

  useEffect(() => {
    if (!autoRefresh) {
      return;
    }
    const timer = window.setInterval(() => {
      setRefreshTick((value) => value + 1);
    }, 10000);
    return () => {
      window.clearInterval(timer);
    };
  }, [autoRefresh]);

  useEffect(() => {
    if (!state.data) {
      setSelectedBindingId("");
      setSelectedReplayId("");
      setSelectedSegmentId("");
      return;
    }
    if (!selectedBindingId && state.data.recent_bindings.length > 0) {
      const fromTask = taskFromUrl
        ? state.data.recent_bindings.find((item) => item.task_id === taskFromUrl)?.binding_id
        : "";
      setSelectedBindingId(fromTask || state.data.recent_bindings[0].binding_id);
    }
  }, [state.data, selectedBindingId, taskFromUrl]);

  const selectedBinding = state.data?.recent_bindings.find((item) => item.binding_id === selectedBindingId) ?? null;

  const bindingDetailState = useQuery(
    () =>
      selectedBindingId
        ? apiGet<ExecutionView["latest_binding"]>(
            `/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}/detail`,
          )
        : Promise.resolve(null),
    [selectedBindingId, refreshTick],
  );

  const activeRunID = bindingDetailState.data?.run_binding.run_id ?? selectedBinding?.run_id ?? "";

  const replaySummaryState = useQuery(
    () =>
      activeRunID
        ? apiGet<ExecutionView["replay_summary"]>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/replay-summary`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, refreshTick],
  );

  const replayTimelineState = useQuery(
    () =>
      activeRunID
        ? apiGet<{ items: ExecutionView["replay_timeline"] }>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/replay-timeline?limit=10`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, refreshTick],
  );

  const logSegmentsState = useQuery(
    () =>
      activeRunID
        ? apiGet<{ segments: ExecutionView["log_segments"] }>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/log-segments?limit=10`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, refreshTick],
  );

  useEffect(() => {
    setSelectedReplayId("");
    setSelectedSegmentId("");
  }, [selectedBindingId]);

  useEffect(() => {
    if (!selectedReplayId && (replayTimelineState.data?.items?.length ?? 0) > 0) {
      setSelectedReplayId(replayFromUrl || replayTimelineState.data!.items[0].replay_id);
    }
    if (!selectedSegmentId && (logSegmentsState.data?.segments?.length ?? 0) > 0) {
      setSelectedSegmentId(segmentFromUrl || logSegmentsState.data!.segments[0].segment_id);
    }
  }, [replayTimelineState.data, logSegmentsState.data, replayFromUrl, segmentFromUrl, selectedReplayId, selectedSegmentId]);

  useEffect(() => {
    const next = buildRoute("/execution", {
      binding: selectedBindingId || undefined,
      replay: selectedReplayId || undefined,
      segment: selectedSegmentId || undefined,
      task: taskFromUrl || undefined,
    });
    window.history.replaceState(null, "", next);
  }, [buildRoute, selectedBindingId, selectedReplayId, selectedSegmentId, taskFromUrl]);

  const replayDetailState = useQuery(
    () =>
      selectedReplayId && activeRunID
        ? apiGet<ReplayDetailView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/replay-items/${encodeURIComponent(selectedReplayId)}`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, selectedReplayId, refreshTick],
  );
  const replayRawState = useQuery(
    () =>
      selectedReplayId && activeRunID
        ? apiGet<ReplayRawView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/replay-items/${encodeURIComponent(selectedReplayId)}/raw?limit=12000`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, selectedReplayId, refreshTick],
  );
  const logRawState = useQuery(
    () =>
      selectedSegmentId && activeRunID
        ? apiGet<LogSegmentRawView>(
            `/api/v3/projects/${encodeURIComponent(projectId)}/runs/${encodeURIComponent(activeRunID)}/log-segments/${encodeURIComponent(selectedSegmentId)}/raw?limit=12000`,
          )
        : Promise.resolve(null),
    [projectId, activeRunID, selectedSegmentId, refreshTick],
  );

  const replayPreview = useMemo(() => {
    return formatRawPreview(replayRawState.data?.content, replayRawState.data?.truncated);
  }, [replayRawState.data]);
  const logPreview = useMemo(() => {
    return formatRawPreview(logRawState.data?.content, logRawState.data?.truncated);
  }, [logRawState.data]);
  const filteredBindings = useMemo(() => {
    return (state.data?.recent_bindings ?? []).filter((item) => runStatusFilter === "all" || item.run_status === runStatusFilter);
  }, [state.data?.recent_bindings, runStatusFilter]);
  const filteredEvents = useMemo(() => {
    return (bindingDetailState.data?.recent_events ?? []).filter(
      (item) => eventLevelFilter === "all" || (item.event_level || "info") === eventLevelFilter,
    );
  }, [bindingDetailState.data?.recent_events, eventLevelFilter]);
  const filteredReplayItems = useMemo(() => {
    return (replayTimelineState.data?.items ?? []).filter(
      (item) => replayTypeFilter === "all" || item.replay_type === replayTypeFilter,
    );
  }, [replayTimelineState.data?.items, replayTypeFilter]);
  const runStatusOptions = useMemo(() => buildOptions(state.data?.recent_bindings.map((item) => item.run_status) ?? []), [state.data?.recent_bindings]);
  const replayTypeOptions = useMemo(
    () => buildOptions(replayTimelineState.data?.items.map((item) => item.replay_type) ?? []),
    [replayTimelineState.data?.items],
  );
  const eventLevelOptions = useMemo(
    () => buildOptions((bindingDetailState.data?.recent_events ?? []).map((item) => item.event_level || "info")),
    [bindingDetailState.data?.recent_events],
  );

  async function runBindingMutation<T extends CommandResponse | StartRunResponse | SyncRunResponse>(
    actionKey: string,
    execute: () => Promise<T>,
  ) {
    setMutationBusy(actionKey);
    setMutationError("");
    try {
      const result = await execute();
      const nextAction = "next_action" in result ? result.next_action : "refresh_execution_view";
      setMutationMessage(`${actionKey} accepted · ${nextAction}`);
      if ("run_binding" in result && result.run_binding?.binding_id) {
        setSelectedBindingId(result.run_binding.binding_id);
      }
      setRefreshTick((value) => value + 1);
    } catch (error) {
      setMutationError(error instanceof Error ? error.message : "Runtime action failed");
    } finally {
      setMutationBusy("");
    }
  }

  return (
    <QueryPanel
      loading={state.loading}
      error={state.error}
      title="Execution"
      onRetry={() => setRefreshTick((value) => value + 1)}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage="Execution can stay partially usable even when runtime detail or replay fetches fail."
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Execution</p>
              <h3 className="placeholder-title">{state.data.runtime_health.status}</h3>
              <p className="placeholder-description">
                Runtime health, recent bindings, replay summary, and log segments for project {projectId}.
              </p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">{state.data.runtime_health.base_url}</span>
              <span className="status-pill">{state.data.recent_bindings.length} recent runs</span>
            </div>
          </div>

          <section className="data-panel execution-toolbar">
            <div className="toolbar-group">
              <button className="primary-button" onClick={() => setRefreshTick((value) => value + 1)}>
                Refresh Now
              </button>
              <label className="toggle-row">
                <input type="checkbox" checked={autoRefresh} onChange={(event) => setAutoRefresh(event.target.checked)} />
                <span>Auto refresh 10s</span>
              </label>
            </div>
            <div className="toolbar-filters">
              <label className="filter-field">
                <span>Run Status</span>
                <select className="project-input" value={runStatusFilter} onChange={(event) => setRunStatusFilter(event.target.value)}>
                  <option value="all">all</option>
                  {runStatusOptions.map((item) => (
                    <option key={item} value={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </label>
              <label className="filter-field">
                <span>Replay Type</span>
                <select className="project-input" value={replayTypeFilter} onChange={(event) => setReplayTypeFilter(event.target.value)}>
                  <option value="all">all</option>
                  {replayTypeOptions.map((item) => (
                    <option key={item} value={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </label>
              <label className="filter-field">
                <span>Event Level</span>
                <select className="project-input" value={eventLevelFilter} onChange={(event) => setEventLevelFilter(event.target.value)}>
                  <option value="all">all</option>
                  {eventLevelOptions.map((item) => (
                    <option key={item} value={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </label>
            </div>
          </section>

          <div className="metrics-grid">
            <MetricCard label="Runtime" value={state.data.runtime_health.status} />
            <MetricCard label="Selected Run" value={bindingDetailState.data?.run_binding.run_status ?? "idle"} />
            <MetricCard label="Replay Items" value={`${replaySummaryState.data?.replay_count ?? 0}`} />
            <MetricCard label="Log Segments" value={`${replaySummaryState.data?.log_segment_count ?? 0}`} />
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Runtime Commands</h3>
            </div>
            <div className="form-grid">
              <label className="form-field">
                <span>Task ID</span>
                <input className="project-input" value={startTaskId} onChange={(event) => setStartTaskId(event.target.value)} placeholder="task_xxx" />
              </label>
              <label className="form-field">
                <span>Brain Kind</span>
                <input className="project-input" value={startBrainKind} onChange={(event) => setStartBrainKind(event.target.value)} placeholder="coder" />
              </label>
              <label className="form-field">
                <span>Provider</span>
                <input className="project-input" value={startProvider} onChange={(event) => setStartProvider(event.target.value)} placeholder="optional provider" />
              </label>
              <label className="form-field">
                <span>Prompt</span>
                <textarea className="project-input project-textarea" value={startPrompt} onChange={(event) => setStartPrompt(event.target.value)} />
              </label>
              <div className="action-row">
                <button
                  className="primary-button"
                  disabled={mutationBusy !== "" || !startTaskId.trim() || !startPrompt.trim()}
                  onClick={() =>
                    runBindingMutation("start_runtime_run", () =>
                      apiPost<StartRunResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/runtime-runs`, {
                        id: projectId,
                        task_id: startTaskId.trim(),
                        brain_kind: startBrainKind.trim() || undefined,
                        prompt: startPrompt.trim(),
                        provider: startProvider.trim() || undefined,
                      }),
                    )
                  }
                >
                  {mutationBusy === "start_runtime_run" ? "Starting..." : "Start Run"}
                </button>
                <button
                  className="secondary-button"
                  disabled={mutationBusy !== "" || !selectedBindingId}
                  onClick={() =>
                    runBindingMutation("sync_runtime_run", () =>
                      apiPost<SyncRunResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}/sync`),
                    )
                  }
                >
                  {mutationBusy === "sync_runtime_run" ? "Syncing..." : "Sync"}
                </button>
                <button
                  className="secondary-button"
                  disabled={mutationBusy !== "" || !selectedBindingId}
                  onClick={() =>
                    runBindingMutation("resume_runtime_run", () =>
                      apiPost<CommandResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}/resume`),
                    )
                  }
                >
                  {mutationBusy === "resume_runtime_run" ? "Resuming..." : "Resume"}
                </button>
                <button
                  className="secondary-button danger-button"
                  disabled={mutationBusy !== "" || !selectedBindingId}
                  onClick={() =>
                    runBindingMutation("cancel_runtime_run", () =>
                      apiDelete<CommandResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}`),
                    )
                  }
                >
                  {mutationBusy === "cancel_runtime_run" ? "Cancelling..." : "Cancel"}
                </button>
              </div>
            </div>
            {mutationError ? <p className="error-copy">{mutationError}</p> : null}
            {mutationMessage ? <p className="muted-copy">{mutationMessage}</p> : null}
          </section>

          {bindingDetailState.data?.runtime_stale || bindingDetailState.data?.runtime_error || state.data.runtime_error ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>Runtime Status</h3>
              </div>
              <div className="stack-list">
                {bindingDetailState.data?.runtime_stale ? (
                  <article className="list-card">
                    <p>Runtime detail may be stale for the selected binding. Refresh or resync if the state looks outdated.</p>
                  </article>
                ) : null}
                {bindingDetailState.data?.runtime_error ? (
                  <article className="list-card">
                    <p>{bindingDetailState.data.runtime_error}</p>
                  </article>
                ) : null}
                {state.data.runtime_error ? (
                  <article className="list-card">
                    <p>{state.data.runtime_error}</p>
                  </article>
                ) : null}
              </div>
            </section>
          ) : null}

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Recent Bindings</h3>
              </div>
              <div className="stack-list">
                {filteredBindings.map((item) => (
                  <button
                    key={item.binding_id}
                    className={item.binding_id === selectedBindingId ? "action-card is-selected" : "action-card"}
                    onClick={() => setSelectedBindingId(item.binding_id)}
                  >
                    <div className="list-card-head">
                      <strong>{item.run_id || item.binding_id}</strong>
                      <span className="status-pill">{item.run_status}</span>
                    </div>
                    <p>
                      {item.brain_kind} · task {item.task_id || "n/a"} · {item.started_at}
                    </p>
                  </button>
                ))}
                {filteredBindings.length === 0 ? (
                  <article className="list-card">
                    <p>No recent run bindings match the current filter.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Recent Events</h3>
              </div>
              <div className="stack-list">
                {filteredEvents.map((item) => (
                  <article key={item.event_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className="status-pill">{item.event_type}</span>
                    </div>
                    <p>
                      {item.event_level || "info"} · {item.created_at}
                    </p>
                  </article>
                ))}
                {filteredEvents.length === 0 ? (
                  <article className="list-card">
                    <p>No recent runtime events match the current filter.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Replay Summary</h3>
              </div>
              {replaySummaryState.data ? (
                <div className="stack-list">
                  <article className="list-card">
                    <div className="list-card-head">
                      <strong>{replaySummaryState.data.run_id}</strong>
                      <span className="status-pill">{replaySummaryState.data.status}</span>
                    </div>
                    <p>
                      events {replaySummaryState.data.event_count} · replay {replaySummaryState.data.replay_count} · ready{" "}
                      {replaySummaryState.data.artifact_status_summary.available} · missing{" "}
                      {replaySummaryState.data.artifact_status_summary.missing} · pruned{" "}
                      {replaySummaryState.data.artifact_status_summary.pruned}
                    </p>
                    <p>
                      {replaySummaryState.data.brain_kind || "unknown brain"} · started{" "}
                      {replaySummaryState.data.started_at || "n/a"} · ended {replaySummaryState.data.ended_at || "n/a"}
                    </p>
                  </article>
                  {replaySummaryState.data.entry_points.map((item) => (
                    <button
                      key={item.replay_id}
                      className={item.replay_id === selectedReplayId ? "action-card is-selected" : "action-card"}
                      onClick={() => setSelectedReplayId(item.replay_id)}
                    >
                      <div className="list-card-head">
                        <strong>{item.summary || item.replay_id}</strong>
                        <span className="status-pill">{item.replay_type}</span>
                      </div>
                      <p>
                        {item.created_at} · {item.file_path || item.replay_id}
                      </p>
                    </button>
                  ))}
                  {replaySummaryState.data.entry_points.length === 0 ? (
                    <article className="list-card">
                      <p>No replay entry points were indexed for the selected run yet.</p>
                    </article>
                  ) : null}
                </div>
              ) : (
                <article className="list-card">
                  <p>{state.data.replay_error || replaySummaryState.error || "No replay artifacts available for the selected run."}</p>
                </article>
              )}
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Replay Timeline</h3>
              </div>
              <div className="stack-list">
                {filteredReplayItems.map((item) => (
                  <button
                    key={item.replay_id}
                    className={item.replay_id === selectedReplayId ? "action-card is-selected" : "action-card"}
                    onClick={() => setSelectedReplayId(item.replay_id)}
                  >
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      <span className="status-pill">{item.replay_type}</span>
                    </div>
                    <p>
                      seq {item.seq_no} · {item.summary || item.status} · {item.created_at}
                    </p>
                    <p>
                      preview {item.preview_available ? "ready" : "unavailable"} · raw {item.raw_target || "n/a"}
                    </p>
                  </button>
                ))}
                {filteredReplayItems.length === 0 ? (
                  <article className="list-card">
                    <p>No replay items are available for the selected run and filter.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Log Segments</h3>
              </div>
              <div className="stack-list">
                {(logSegmentsState.data?.segments ?? []).map((item) => (
                  <button
                    key={item.segment_id}
                    className={item.segment_id === selectedSegmentId ? "action-card is-selected" : "action-card"}
                    onClick={() => setSelectedSegmentId(item.segment_id)}
                  >
                    <div className="list-card-head">
                      <strong>{item.segment_id}</strong>
                      <span className="status-pill">{item.stream_kind}</span>
                    </div>
                    <p>
                      {item.status} · seq {item.seq_no} · {item.size} bytes
                    </p>
                    <p>
                      {item.started_at || "n/a"} → {item.ended_at || "n/a"} · raw {item.raw_target || "n/a"}
                    </p>
                  </button>
                ))}
                {(logSegmentsState.data?.segments?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>No log segments are available for the selected run.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Replay Detail</h3>
              </div>
              {replayDetailState.data ? (
                <div className="stack-list">
                  <article className="list-card">
                    <div className="list-card-head">
                      <strong>{replayDetailState.data.title}</strong>
                      <span className="status-pill">{replayDetailState.data.replay_kind}</span>
                    </div>
                    <p>{replayDetailState.data.summary || replayDetailState.data.status}</p>
                    <p>
                      source {replayDetailState.data.source_object_kind || "n/a"} · object {replayDetailState.data.source_object_id || "n/a"}
                    </p>
                    <p>
                      event {replayDetailState.data.event_id || "n/a"} · trace {replayDetailState.data.trace_id || "n/a"} · span{" "}
                      {replayDetailState.data.span_id || "n/a"}
                    </p>
                    <p>raw target {replayDetailState.data.raw_preview.raw_target || "n/a"}</p>
                    <div className="action-row">
                      <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                        Open Diagnostics
                      </button>
                    </div>
                  </article>
                  <RawPreviewCard
                    title="Replay Raw"
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
                        title: "Replay Raw",
                        subtitle: replayDetailState.data!.title,
                        preview: replayPreview,
                      })
                    }
                  />
                </div>
              ) : (
                <article className="list-card">
                  <p>{replayDetailState.error || "Select a replay item to inspect detail and raw preview."}</p>
                </article>
              )}
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Log Preview</h3>
              </div>
              <div className="stack-list">
                {(replayDetailState.data?.related_log_segments ?? []).map((item) => (
                  <article key={item.segment_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.segment_id}</strong>
                      <span className="status-pill">{item.stream_kind}</span>
                    </div>
                    <p>
                      seq {item.seq_no} · {item.status} · {item.started_at || "n/a"} → {item.ended_at || "n/a"}
                    </p>
                    <p>raw {item.raw_target || "n/a"}</p>
                  </article>
                ))}
                {(replayDetailState.data?.related_log_segments?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>No related log segments were linked to the selected replay item.</p>
                  </article>
                ) : null}
                <RawPreviewCard
                  title="Log Raw"
                  subtitle={selectedSegmentId || "Selected log segment"}
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
                      downloadRawContent(selectedSegmentId || "log-raw", logRawState.data.content);
                    }
                  }}
                  onExpand={() =>
                    setExpandedRawView({
                      title: "Log Raw",
                      subtitle: selectedSegmentId || "Selected log segment",
                      preview: logPreview,
                    })
                  }
                />
              </div>
            </section>
          </div>

          {expandedRawView ? (
            <div
              className="raw-overlay"
              role="dialog"
              aria-modal="true"
              aria-label={`${expandedRawView.title} preview`}
              onClick={() => setExpandedRawView(null)}
            >
              <div className="raw-overlay-panel" onClick={(event) => event.stopPropagation()}>
                <div className="raw-overlay-header">
                  <div>
                    <p className="placeholder-section">Expanded preview</p>
                    <h3 className="placeholder-title">{expandedRawView.title}</h3>
                    <p className="placeholder-description">{expandedRawView.subtitle}</p>
                  </div>
                  <button className="secondary-button" onClick={() => setExpandedRawView(null)}>
                    Close
                  </button>
                </div>
                <pre className="json-block raw-overlay-content">{expandedRawView.preview}</pre>
              </div>
            </div>
          ) : null}
        </section>
      ) : null}
    </QueryPanel>
  );
}

function MetricCard(props: { label: string; value: string }) {
  return (
    <article className="metric-card metric-neutral">
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </article>
  );
}

function buildExecutionViewPath(
  projectId: string,
  bindingId: string,
  eventLimit: number,
  replayLimit: number,
  logLimit: number,
) {
  const params = new URLSearchParams();
  if (bindingId) {
    params.set("binding_id", bindingId);
  }
  params.set("event_limit", String(eventLimit));
  params.set("replay_limit", String(replayLimit));
  params.set("log_limit", String(logLimit));

  const query = params.toString();
  const path = `/api/v3/projects/${encodeURIComponent(projectId)}/execution-view`;
  return query ? `${path}?${query}` : path;
}

function RawPreviewCard(props: {
  title: string;
  subtitle: string;
  preview: string;
  rawContent?: string;
  rawError?: string | null;
  onCopy: () => void;
  onDownload: () => void;
  onExpand: () => void;
}) {
  const hasContent = (props.rawContent ?? "").trim() !== "";
  return (
    <article className="raw-preview-card">
      <div className="panel-header raw-preview-header">
        <div>
          <h3>{props.title}</h3>
          <p className="raw-preview-note">{props.subtitle}</p>
        </div>
        <div className="raw-actions">
          <button className="secondary-button" onClick={props.onCopy} disabled={!hasContent}>
            Copy
          </button>
          <button className="secondary-button" onClick={props.onDownload} disabled={!hasContent}>
            Download
          </button>
          <button className="secondary-button" onClick={props.onExpand} disabled={!hasContent}>
            Expand
          </button>
        </div>
      </div>
      <pre className="json-block">{hasContent ? props.preview : props.rawError || "Select an item to inspect raw output."}</pre>
    </article>
  );
}

function formatRawPreview(content?: string, truncated?: boolean) {
  const text = content && content.trim() !== "" ? content : "<empty>";
  if (truncated) {
    return `${text}\n\n[truncated]`;
  }
  return text;
}

async function copyTextToClipboard(text: string) {
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(text);
    return;
  }

  const textArea = document.createElement("textarea");
  textArea.value = text;
  textArea.setAttribute("readonly", "true");
  textArea.style.position = "fixed";
  textArea.style.left = "-9999px";
  document.body.appendChild(textArea);
  textArea.select();
  document.execCommand("copy");
  document.body.removeChild(textArea);
}

function downloadRawContent(fileBaseName: string, content: string, mimeType?: string) {
  const safeName = sanitizeFileBaseName(fileBaseName);
  const extension = mimeType && mimeType.includes("json") ? "json" : "txt";
  const blob = new Blob([content], { type: mimeType || "text/plain;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = `${safeName}.${extension}`;
  anchor.click();
  window.setTimeout(() => URL.revokeObjectURL(url), 0);
}

function sanitizeFileBaseName(fileBaseName: string) {
  return fileBaseName
    .trim()
    .replace(/[^\w.-]+/g, "-")
    .replace(/^-+|-+$/g, "")
    .slice(0, 80) || "raw-content";
}

function buildOptions(values: string[]) {
  return [...new Set(values.filter((value) => value && value.trim() !== ""))].sort();
}

type ExpandedRawView = {
  title: string;
  subtitle: string;
  preview: string;
};
