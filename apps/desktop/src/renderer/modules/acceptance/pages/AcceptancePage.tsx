import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { AcceptanceView, CommandResponse, PlanView, ProjectDiagnosticsView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function AcceptancePage() {
  const navigate = useNavigate();
  const { projectId, searchParams, buildRoute, routes } = useProjectState();
  const selectedTaskFromUrl = searchParams.get("task")?.trim() || "";
  const [refreshTick, setRefreshTick] = useState(0);
  const [busyAction, setBusyAction] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [selectedTaskId, setSelectedTaskId] = useState(selectedTaskFromUrl);
  const [manualReleaseComment, setManualReleaseComment] = useState("");
  const state = useQuery(
    () => apiGet<AcceptanceView>(`/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-view`),
    [projectId, refreshTick],
  );
  const planState = useQuery(
    () => apiGet<PlanView>(`/api/v3/projects/${encodeURIComponent(projectId)}/plan-view`),
    [projectId, refreshTick],
  );
  const diagnosticsState = useQuery(
    () => apiGet<ProjectDiagnosticsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/diagnostic-records?limit=20`),
    [projectId, refreshTick],
  );

  const availableTaskOptions = useMemo(() => {
    return (planState.data?.task_projection ?? []).filter((item) => item.mapped_domain_task_id);
  }, [planState.data]);
  const currentAcceptanceTaskId = state.data?.acceptance_run.task_id || "";
  const relatedDiagnostics = useMemo(() => {
    const targetTaskID = currentAcceptanceTaskId || selectedTaskId || selectedTaskFromUrl;
    const targetRunID = state.data?.acceptance_run.id || "";
    return (diagnosticsState.data?.items ?? []).filter((item) => {
      if (targetTaskID && item.task_id === targetTaskID) {
        return true;
      }
      if (targetRunID && item.run_id === targetRunID) {
        return true;
      }
      return false;
    });
  }, [currentAcceptanceTaskId, diagnosticsState.data?.items, selectedTaskFromUrl, selectedTaskId, state.data?.acceptance_run.id]);

  useEffect(() => {
    if (selectedTaskFromUrl) {
      setSelectedTaskId(selectedTaskFromUrl);
      return;
    }
    if (state.data?.acceptance_run.task_id) {
      setSelectedTaskId(state.data.acceptance_run.task_id);
    }
  }, [selectedTaskFromUrl, state.data?.acceptance_run.task_id]);

  async function runAction(actionKey: string, execute: () => Promise<CommandResponse>) {
    setBusyAction(actionKey);
    setActionError("");
    try {
      const result = await execute();
      setActionMessage(`${actionKey} accepted · ${result.next_action}`);
      setRefreshTick((value) => value + 1);
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Action failed");
    } finally {
      setBusyAction("");
    }
  }

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title="Acceptance view"
      onRetry={() => setRefreshTick((value) => value + 1)}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage="Acceptance keeps prior data visible while refresh and adjudication commands settle."
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Acceptance</p>
              <h3 className="placeholder-title">{state.data.release_gate.status}</h3>
              <p className="placeholder-description">{state.data.release_gate.summary}</p>
            </div>
            <div className="summary-stack action-stack">
              <span className="status-pill">{state.data.acceptance_run.functional_status}</span>
              <span className="status-pill">{state.data.acceptance_run.production_status}</span>
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={() =>
                  runAction("refresh_acceptance_profiles", () =>
                    apiPost<CommandResponse>(
                      `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-profiles/refresh`,
                      { id: projectId },
                    ),
                  )
                }
              >
                {busyAction === "refresh_acceptance_profiles" ? "Refreshing..." : "Refresh Profiles"}
              </button>
              <button
                className="secondary-button"
                disabled={busyAction !== ""}
                onClick={() =>
                  runAction("adjudicate_acceptance", () =>
                    apiPost<CommandResponse>(
                      `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs/adjudicate`,
                      { id: projectId },
                    ),
                  )
                }
              >
                {busyAction === "adjudicate_acceptance" ? "Adjudicating..." : "Adjudicate Now"}
              </button>
            </div>
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Acceptance Commands</h3>
            </div>
            <div className="inline-metrics">
              <span className="status-pill">next {state.data.release_gate.next_action}</span>
              <span className="status-pill">run {state.data.acceptance_run.id || "none"}</span>
              <span className="status-pill">task {state.data.acceptance_run.task_id || "auto"}</span>
              <span className="status-pill">profile {state.data.acceptance_run.profile_version || "default"}</span>
            </div>
            <div className="form-grid">
              <label className="form-field">
                <span>Target Task</span>
                <select className="project-input" value={selectedTaskId} onChange={(event) => setSelectedTaskId(event.target.value)}>
                  <option value="">auto / latest</option>
                  {availableTaskOptions.map((item) => (
                    <option key={item.task_id} value={item.mapped_domain_task_id}>
                      {item.task_name} · {item.mapped_domain_task_id}
                    </option>
                  ))}
                </select>
              </label>
              <label className="form-field">
                <span>Manual Release Comment</span>
                <input
                  className="project-input"
                  value={manualReleaseComment}
                  onChange={(event) => setManualReleaseComment(event.target.value)}
                  placeholder="optional approval note"
                />
              </label>
              <div className="action-row">
                <button
                  className="primary-button"
                  disabled={busyAction !== ""}
                  onClick={() =>
                    runAction("start_acceptance", () =>
                      apiPost<CommandResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs`, {
                        id: projectId,
                        task_id: selectedTaskId || undefined,
                        mode: "running",
                      }),
                    )
                  }
                >
                  {busyAction === "start_acceptance" ? "Starting..." : "Start Acceptance"}
                </button>
                <button className="secondary-button" disabled={busyAction !== ""} onClick={() => setRefreshTick((value) => value + 1)}>
                  Refresh View
                </button>
                {state.data.release_gate.next_action === "start_acceptance_run" ? (
                  <button
                    className="secondary-button"
                    disabled={busyAction !== ""}
                    onClick={() =>
                      runAction("start_acceptance_from_gate", () =>
                        apiPost<CommandResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs`, {
                          id: projectId,
                          task_id: selectedTaskId || undefined,
                          mode: "running",
                        }),
                      )
                    }
                  >
                    Start from Release Gate
                  </button>
                ) : null}
                {state.data.release_gate.next_action === "apply_manual_release" ? (
                  <button
                    className="primary-button"
                    disabled={busyAction !== ""}
                    onClick={() =>
                      runAction("apply_manual_release", () =>
                        apiPost<CommandResponse>(
                          `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs/manual-release`,
                          { id: projectId, comment: manualReleaseComment.trim() || undefined },
                        ),
                      )
                    }
                  >
                    {busyAction === "apply_manual_release" ? "Applying..." : "Approve Manual Release"}
                  </button>
                ) : null}
                {currentAcceptanceTaskId ? (
                  <button
                    className="secondary-button"
                    onClick={() => navigate(buildRoute("/execution", { task: currentAcceptanceTaskId }))}
                  >
                    Open Task Execution
                  </button>
                ) : null}
                {planState.data?.repair_draft.id ? (
                  <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                    Open Repair Draft
                  </button>
                ) : null}
              </div>
            </div>
            {actionError ? <p className="error-copy">{actionError}</p> : null}
            {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}
            {state.data.acceptance_run.latest_judgement_summary ? (
              <p className="muted-copy">
                Latest judgement · {state.data.acceptance_run.latest_judgement_kind || "judgement"} ·{" "}
                {state.data.acceptance_run.latest_judgement_result || "n/a"} · {state.data.acceptance_run.latest_judgement_summary}
                {state.data.acceptance_run.latest_judgement_at ? ` · ${state.data.acceptance_run.latest_judgement_at}` : ""}
              </p>
            ) : null}
            {state.data.acceptance_run.finished_at ? (
              <p className="muted-copy">Acceptance finished at {state.data.acceptance_run.finished_at}</p>
            ) : null}
          </section>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Coverage Matrix</h3>
              </div>
              <div className="stack-list">
                {state.data.coverage_matrix.map((item) => (
                  <article key={`${item.kind}-${item.key}`} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.name}</strong>
                      <span className="status-pill">{item.coverage_status}</span>
                    </div>
                    <p>{item.kind} · evidence {item.evidence_count}</p>
                  </article>
                ))}
                {state.data.coverage_matrix.length === 0 ? (
                  <article className="list-card">
                    <p>No coverage matrix items have been recorded yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Issues</h3>
              </div>
              <div className="stack-list">
                {state.data.issues.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className={`severity-badge severity-${item.severity}`}>{item.severity}</span>
                    </div>
                    <p>{item.blocking ? "blocking" : "non-blocking"}</p>
                    <div className="action-row">
                      {currentAcceptanceTaskId ? (
                        <button
                          className="secondary-button"
                          onClick={() => navigate(buildRoute("/execution", { task: currentAcceptanceTaskId }))}
                        >
                          Inspect Execution
                        </button>
                      ) : null}
                      <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                        Open Diagnostics
                      </button>
                      {planState.data?.repair_draft.id ? (
                        <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                          Open Repair Draft
                        </button>
                      ) : null}
                    </div>
                  </article>
                ))}
                {state.data.issues.length === 0 ? (
                  <article className="list-card">
                    <p>No acceptance issues are currently open.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Evidence</h3>
              </div>
              <div className="stack-list">
                {state.data.evidence_cards.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.evidence_type}</strong>
                      <span className="status-pill">{item.surface || item.journey || "artifact"}</span>
                    </div>
                    <p>{item.file_path}</p>
                  </article>
                ))}
                {state.data.evidence_cards.length === 0 ? (
                  <article className="list-card">
                    <p>No evidence artifacts are available yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Repair Context</h3>
                <span className="status-pill">{planState.data?.repair_draft.id ? "ready" : "idle"}</span>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{planState.data?.repair_draft.id || "No repair draft yet"}</strong>
                    <span className="status-pill">{planState.data?.repair_draft.status || "idle"}</span>
                  </div>
                  <p>{planState.data?.repair_draft.reasoning_summary || "The latest failed acceptance run has not produced a repair draft yet."}</p>
                  <div className="action-row">
                    {planState.data?.repair_draft.id ? (
                      <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                        Open Repair Draft
                      </button>
                    ) : (
                      <button className="secondary-button" onClick={() => navigate(routes.plan)}>
                        Open Plan
                      </button>
                    )}
                    {currentAcceptanceTaskId ? (
                      <button
                        className="secondary-button"
                        onClick={() => navigate(buildRoute("/execution", { task: currentAcceptanceTaskId }))}
                      >
                        Open Execution
                      </button>
                    ) : null}
                  </div>
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Related Diagnostics</h3>
                <span className="status-pill">{relatedDiagnostics.length}</span>
              </div>
              <div className="stack-list">
                {relatedDiagnostics.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className={`severity-badge severity-${item.severity}`}>{item.severity}</span>
                    </div>
                    <p>
                      {item.scope} · {item.error_code} · {item.created_at}
                    </p>
                    <p>
                      {item.task_id ? `task ${item.task_id}` : "task n/a"} · {item.run_id ? `run ${item.run_id}` : "run n/a"} ·{" "}
                      {item.binding_id ? `binding ${item.binding_id}` : "binding n/a"}
                    </p>
                    <div className="action-row">
                      <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                        Open Diagnostics
                      </button>
                      {item.task_id ? (
                        <button
                          className="secondary-button"
                          onClick={() =>
                            navigate(
                              buildRoute("/execution", {
                                task: item.task_id,
                                binding: item.binding_id,
                              }),
                            )
                          }
                        >
                          Open Execution
                        </button>
                      ) : null}
                      {planState.data?.repair_draft.id ? (
                        <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                          Open Repair Draft
                        </button>
                      ) : null}
                    </div>
                    {item.detail_json ? <pre className="json-block">{prettyJson(item.detail_json)}</pre> : null}
                  </article>
                ))}
                {relatedDiagnostics.length === 0 ? (
                  <article className="list-card">
                    <p>No diagnostics are linked to the current acceptance task or run yet.</p>
                  </article>
                ) : null}
              </div>
            </section>
          </div>
        </section>
      ) : null}
    </QueryPanel>
  );
}

function prettyJson(raw?: string) {
  if (!raw || raw.trim() === "") {
    return "{}";
  }
  try {
    return JSON.stringify(JSON.parse(raw), null, 2);
  } catch {
    return raw;
  }
}
