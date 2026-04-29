import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiDelete, apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type {
  AcceptanceView,
  CommandResponse,
  ExecutionView,
  ProjectDiagnosticsView,
  StartRunResponse,
  SyncRunResponse,
} from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";
import { ReplayDrawer } from "@/modules/replay/components/ReplayDrawer";

export function ExecutionPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, searchParams, routes, buildRoute } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [autoRefresh, setAutoRefresh] = useState(false);
  const [replayOpen, setReplayOpen] = useState(false);
  const [replayRunId, setReplayRunId] = useState("");
  const bindingFromUrl = searchParams.get("binding")?.trim() || "";
  const runFromUrl = searchParams.get("run")?.trim() || "";
  const taskFromUrl = searchParams.get("task")?.trim() || "";
  const [selectedBindingId, setSelectedBindingId] = useState(bindingFromUrl);
  const [runStatusFilter, setRunStatusFilter] = useState("all");
  const [eventLevelFilter, setEventLevelFilter] = useState("all");
  const [startTaskId, setStartTaskId] = useState("");
  const [startPrompt, setStartPrompt] = useState(
    "Review the assigned task and execute the required implementation steps.",
  );
  const [startBrainKind, setStartBrainKind] = useState("coder");
  const [startProvider, setStartProvider] = useState("");
  const [mutationBusy, setMutationBusy] = useState("");
  const [mutationError, setMutationError] = useState("");
  const [mutationMessage, setMutationMessage] = useState("");

  useEffect(() => {
    if (!autoRefresh) return;
    const timer = window.setInterval(() => setRefreshTick((v) => v + 1), 10000);
    return () => window.clearInterval(timer);
  }, [autoRefresh]);

  const state = useQuery(
    () =>
      apiGet<ExecutionView>(
        buildExecutionViewPath(projectId, selectedBindingId, 20, 10, 10),
      ),
    [projectId, selectedBindingId, refreshTick],
  );

  useEffect(() => {
    if (!state.data) {
      setSelectedBindingId("");
      return;
    }
    if (!selectedBindingId && state.data.recent_bindings.length > 0) {
      const fromTask = taskFromUrl
        ? state.data.recent_bindings.find((i) => i.task_id === taskFromUrl)?.binding_id
        : "";
      const fromRun = runFromUrl
        ? state.data.recent_bindings.find((i) => i.run_id === runFromUrl)?.binding_id
        : "";
      setSelectedBindingId(fromTask || fromRun || state.data.recent_bindings[0].binding_id);
    }
  }, [runFromUrl, state.data, selectedBindingId, taskFromUrl]);

  const selectedBinding =
    state.data?.recent_bindings.find((i) => i.binding_id === selectedBindingId) ?? null;

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

  const diagnosticsState = useQuery(
    () => apiGet<ProjectDiagnosticsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/diagnostic-records?limit=20`),
    [projectId, refreshTick],
  );
  const acceptanceState = useQuery(
    () => apiGet<AcceptanceView>(`/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-view`),
    [projectId, refreshTick],
  );

  const filteredBindings = useMemo(
    () =>
      (state.data?.recent_bindings ?? []).filter(
        (i) => runStatusFilter === "all" || i.run_status === runStatusFilter,
      ),
    [state.data?.recent_bindings, runStatusFilter],
  );
  const filteredEvents = useMemo(
    () =>
      (bindingDetailState.data?.recent_events ?? []).filter(
        (i) => eventLevelFilter === "all" || (i.event_level || "info") === eventLevelFilter,
      ),
    [bindingDetailState.data?.recent_events, eventLevelFilter],
  );
  const eventLevelOptions = useMemo(
    () =>
      buildOptions(
        (bindingDetailState.data?.recent_events ?? []).map((i) => i.event_level || "info"),
      ),
    [bindingDetailState.data?.recent_events],
  );
  const runStatusOptions = useMemo(
    () => buildOptions(state.data?.recent_bindings.map((i) => i.run_status) ?? []),
    [state.data?.recent_bindings],
  );

  const relatedDiagnostics = useMemo(() => {
    return (diagnosticsState.data?.items ?? []).filter((item) => {
      if (selectedBindingId && item.binding_id === selectedBindingId) return true;
      if (activeRunID && item.run_id === activeRunID) return true;
      if (taskFromUrl && item.task_id === taskFromUrl) return true;
      if (selectedBinding?.task_id && item.task_id === selectedBinding.task_id) return true;
      return false;
    });
  }, [activeRunID, diagnosticsState.data?.items, selectedBinding?.task_id, selectedBindingId, taskFromUrl]);

  const verificationResult = acceptanceState.data?.verification_result;
  const completionVerdict = acceptanceState.data?.completion_verdict;
  const faultSummary = acceptanceState.data?.fault_summary;
  const repairPlanDraft = acceptanceState.data?.repair_plan_draft;
  const selectedTaskID = selectedBinding?.task_id || taskFromUrl || "";

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
      setRefreshTick((v) => v + 1);
    } catch (error) {
      setMutationError(error instanceof Error ? error.message : t("execution.runtimeActionFailed"));
    } finally {
      setMutationBusy("");
    }
  }

  function openReplay(runId: string) {
    setReplayRunId(runId);
    setReplayOpen(true);
  }

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("execution.noProjectTitle")}</h4>
          <p>{t("execution.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("execution.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }

  return (
    <>
      <QueryPanel
        loading={state.loading}
        error={state.error}
        title={t("execution.title")}
        onRetry={() => setRefreshTick((v) => v + 1)}
        secondaryActionLabel={t("execution.openDiagnostics")}
        onSecondaryAction={() => navigate(routes.diagnostics)}
        recoveryMessage={t("execution.executionRecovery")}
      >
        {state.data ? (
          <section className="dashboard-page">
            <div className="dashboard-intro">
              <div>
                <p className="placeholder-section">{t("execution.title")}</p>
                <h3 className="placeholder-title">{state.data.runtime_health.status}</h3>
                <p className="placeholder-description">
                  {t("execution.executionDescription", { projectId })}
                </p>
              </div>
              <div className="summary-stack">
                <span className="status-pill">{state.data.runtime_health.base_url}</span>
                <span className="status-pill">{state.data.recent_bindings.length} recent runs</span>
                {activeRunID ? <span className="status-pill">run {activeRunID}</span> : null}
                {selectedTaskID ? <span className="status-pill">task {selectedTaskID}</span> : null}
              </div>
            </div>

            <section className="data-panel execution-toolbar">
              <div className="toolbar-group">
                <button className="primary-button" onClick={() => setRefreshTick((v) => v + 1)}>
                  {t("execution.refreshNow")}
                </button>
                <label className="toggle-row">
                  <input type="checkbox" checked={autoRefresh} onChange={(e) => setAutoRefresh(e.target.checked)} />
                  <span>{t("execution.autoRefresh")}</span>
                </label>
              </div>
              <div className="toolbar-filters">
                <label className="filter-field">
                  <span>{t("execution.runStatus")}</span>
                  <select className="project-input" value={runStatusFilter} onChange={(e) => setRunStatusFilter(e.target.value)}>
                    <option value="all">all</option>
                    {runStatusOptions.map((item) => (
                      <option key={item} value={item}>{item}</option>
                    ))}
                  </select>
                </label>
                <label className="filter-field">
                  <span>{t("execution.eventLevel")}</span>
                  <select className="project-input" value={eventLevelFilter} onChange={(e) => setEventLevelFilter(e.target.value)}>
                    <option value="all">all</option>
                    {eventLevelOptions.map((item) => (
                      <option key={item} value={item}>{item}</option>
                    ))}
                  </select>
                </label>
              </div>
            </section>

            <div className="metrics-grid">
              <MetricCard label={t("execution.runtime")} value={state.data.runtime_health.status} />
              <MetricCard label={t("execution.selectedRun")} value={bindingDetailState.data?.run_binding.run_status ?? "idle"} />
              <MetricCard label={t("execution.replayItems")} value={`${state.data.replay_timeline?.length ?? 0}`} />
              <MetricCard label={t("execution.logSegments")} value={`${state.data.log_segments?.length ?? 0}`} />
            </div>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("execution.runtimeCommands")}</h3>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>{t("execution.taskId")}</span>
                  <input className="project-input" value={startTaskId} onChange={(e) => setStartTaskId(e.target.value)} placeholder="task_xxx" />
                </label>
                <label className="form-field">
                  <span>{t("execution.brainKind")}</span>
                  <input className="project-input" value={startBrainKind} onChange={(e) => setStartBrainKind(e.target.value)} placeholder="coder" />
                </label>
                <label className="form-field">
                  <span>{t("execution.provider")}</span>
                  <input className="project-input" value={startProvider} onChange={(e) => setStartProvider(e.target.value)} placeholder="optional provider" />
                </label>
                <label className="form-field">
                  <span>{t("execution.prompt")}</span>
                  <textarea className="project-input project-textarea" value={startPrompt} onChange={(e) => setStartPrompt(e.target.value)} />
                </label>
                <div className="action-row">
                  <button className="primary-button" disabled={mutationBusy !== "" || !startTaskId.trim() || !startPrompt.trim()} onClick={() => runBindingMutation("start_runtime_run", () => apiPost<StartRunResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/runtime-runs`, { id: projectId, task_id: startTaskId.trim(), brain_kind: startBrainKind.trim() || undefined, prompt: startPrompt.trim(), provider: startProvider.trim() || undefined }))}>
                    {mutationBusy === "start_runtime_run" ? t("execution.starting") : t("execution.startRun")}
                  </button>
                  <button className="secondary-button" disabled={mutationBusy !== "" || !selectedBindingId} onClick={() => runBindingMutation("sync_runtime_run", () => apiPost<SyncRunResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}/sync`))}>
                    {mutationBusy === "sync_runtime_run" ? t("execution.syncing") : t("execution.sync")}
                  </button>
                  <button className="secondary-button" disabled={mutationBusy !== "" || !selectedBindingId} onClick={() => runBindingMutation("resume_runtime_run", () => apiPost<CommandResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}/resume`))}>
                    {mutationBusy === "resume_runtime_run" ? t("execution.resuming") : t("execution.resume")}
                  </button>
                  <button className="secondary-button danger-button" disabled={mutationBusy !== "" || !selectedBindingId} onClick={() => runBindingMutation("cancel_runtime_run", () => apiDelete<CommandResponse>(`/api/v3/runtime-runs/${encodeURIComponent(selectedBindingId)}`))}>
                    {mutationBusy === "cancel_runtime_run" ? t("execution.cancelling") : t("execution.cancel")}
                  </button>
                </div>
              </div>
              {mutationError ? <p className="error-copy">{mutationError}</p> : null}
              {mutationMessage ? <p className="muted-copy">{mutationMessage}</p> : null}
            </section>

            {bindingDetailState.data?.runtime_stale || bindingDetailState.data?.runtime_error || state.data.runtime_error ? (
              <section className="data-panel">
                <div className="panel-header"><h3>{t("execution.runtimeStatus")}</h3></div>
                <div className="stack-list">
                  {bindingDetailState.data?.runtime_stale ? <article className="list-card"><p>{t("execution.runtimeStale")}</p></article> : null}
                  {bindingDetailState.data?.runtime_error ? <article className="list-card"><p>{bindingDetailState.data.runtime_error}</p></article> : null}
                  {state.data.runtime_error ? <article className="list-card"><p>{state.data.runtime_error}</p></article> : null}
                </div>
              </section>
            ) : null}

            <section className="data-panel">
              <div className="panel-header"><h3>{t("execution.verificationAndFaultClosure")}</h3></div>
              <div className="metrics-grid">
                <MetricCard label={t("execution.verification")} value={verificationResult?.decision || verificationResult?.status || "pending"} />
                <MetricCard label={t("execution.completion")} value={completionVerdict?.decision || completionVerdict?.final_status || "pending"} />
                <MetricCard label={t("execution.fault")} value={faultSummary?.fault_kind || (faultSummary?.fault_loop_detected ? "fault_loop" : faultSummary?.status) || "none"} />
                <MetricCard label={t("execution.repairLabel")} value={repairPlanDraft?.status || repairPlanDraft?.repair_strategy || "none"} />
              </div>
            </section>

            <div className="content-grid">
              <section className="data-panel">
                <div className="panel-header"><h3>{t("execution.recentBindings")}</h3></div>
                <div className="stack-list">
                  {filteredBindings.map((item) => (
                    <article key={item.binding_id} className={item.binding_id === selectedBindingId ? "action-card is-selected" : "action-card"} onClick={() => setSelectedBindingId(item.binding_id)}>
                      <div className="list-card-head">
                        <strong>{item.run_id || item.binding_id}</strong>
                        <span className="status-pill">{item.run_status}</span>
                      </div>
                      <p>{item.brain_kind} · task {item.task_id || "n/a"} · {item.started_at}</p>
                      <div className="action-row">
                        <button className="secondary-button" onClick={(e) => { e.stopPropagation(); setSelectedBindingId(item.binding_id); }}>{t("execution.selectRun")}</button>
                        <button className="secondary-button" onClick={(e) => { e.stopPropagation(); openReplay(item.run_id); }}>{t("execution.openReplay")}</button>
                        <button className="secondary-button" onClick={(e) => { e.stopPropagation(); navigate(buildRoute("/audit", { run: item.run_id, task: item.task_id })); }}>{t("execution.openAudit")}</button>
                      </div>
                    </article>
                  ))}
                  {filteredBindings.length === 0 ? (
                    <article className="list-card">
                      <p>{t("execution.noRecentBindings")}</p>
                      <div className="action-row" style={{ marginTop: 8 }}>
                        <button className="primary-button" disabled={mutationBusy !== "" || !startTaskId.trim()} onClick={() => runBindingMutation("start_runtime_run", () => apiPost<StartRunResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/runtime-runs`, { id: projectId, task_id: startTaskId.trim(), brain_kind: startBrainKind.trim() || undefined, prompt: startPrompt.trim(), provider: startProvider.trim() || undefined }))}>
                          {t("execution.startRun")}
                        </button>
                      </div>
                    </article>
                  ) : null}
                </div>
              </section>

              <section className="data-panel">
                <div className="panel-header"><h3>{t("execution.recentEvents")}</h3></div>
                <div className="stack-list">
                  {filteredEvents.map((item) => (
                    <article key={item.event_id} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.summary}</strong>
                        <span className="status-pill">{item.event_type}</span>
                      </div>
                      <p>{item.event_level || "info"} · {item.created_at}</p>
                      {item.payload ? <pre className="json-block">{prettyPayload(item.payload)}</pre> : null}
                    </article>
                  ))}
                  {filteredEvents.length === 0 ? <article className="list-card"><p>{t("execution.noRecentEvents")}</p></article> : null}
                </div>
              </section>

              <section className="data-panel">
                <div className="panel-header"><h3>{t("execution.relatedDiagnostics")}</h3><span className="status-pill">{relatedDiagnostics.length}</span></div>
                <div className="stack-list">
                  {relatedDiagnostics.map((item) => (
                    <article key={item.id} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.summary}</strong>
                        <span className={`severity-badge severity-${item.severity}`}>{item.severity}</span>
                      </div>
                      <p>{item.scope} · {item.error_code} · {item.created_at}</p>
                      <div className="action-row">
                        <button className="secondary-button" onClick={() => navigate(buildRoute("/diagnostics", { binding: item.binding_id, run: item.run_id, task: item.task_id }))}>{t("execution.openDiagnostics")}</button>
                        {item.task_id ? <button className="secondary-button" onClick={() => navigate(buildRoute("/acceptance", { task: item.task_id }))}>{t("execution.openAcceptance")}</button> : null}
                      </div>
                    </article>
                  ))}
                  {relatedDiagnostics.length === 0 ? <article className="list-card"><p>{t("execution.noDiagnostics")}</p></article> : null}
                </div>
              </section>
            </div>
          </section>
        ) : null}
      </QueryPanel>

      <ReplayDrawer
        projectId={projectId}
        runId={replayRunId}
        isOpen={replayOpen}
        onClose={() => setReplayOpen(false)}
      />
    </>
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
  if (bindingId) params.set("binding_id", bindingId);
  params.set("event_limit", String(eventLimit));
  params.set("replay_limit", String(replayLimit));
  params.set("log_limit", String(logLimit));
  const query = params.toString();
  const path = `/api/v3/projects/${encodeURIComponent(projectId)}/execution-view`;
  return query ? `${path}?${query}` : path;
}

function prettyPayload(payload?: string) {
  if (!payload) return "";
  try {
    return JSON.stringify(JSON.parse(payload), null, 2);
  } catch {
    return payload;
  }
}

function buildOptions(values: string[]) {
  return [...new Set(values.filter((v) => v && v.trim() !== ""))].sort();
}
