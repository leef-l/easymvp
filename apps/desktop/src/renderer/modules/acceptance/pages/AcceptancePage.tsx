import { useEffect, useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type {
  AcceptanceView,
  CommandResponse,
  PlanView,
  ProjectDiagnosticsView,
} from "@/shared/lib/types";
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
  const verificationResult = state.data?.verification_result;
  const completionVerdict = state.data?.completion_verdict;
  const runtimeEscalation = state.data?.runtime_escalation;
  const faultSummary = state.data?.fault_summary;
  const repairPlanDraft = state.data?.repair_plan_draft;
  const acceptanceOverview = state.data?.overview;
  const acceptanceRun = state.data?.acceptance_run;
  const releaseGate = state.data?.release_gate;
  const acceptanceNextAction = firstText(completionVerdict?.next_action, "inspect_acceptance");

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
              <h3 className="placeholder-title">
                {firstText(completionVerdict?.decision, completionVerdict?.final_status, "pending")}
              </h3>
              <p className="placeholder-description">
                {firstText(completionVerdict?.summary, "Acceptance overview is waiting for the persisted completion verdict.")}
              </p>
            </div>
            <div className="summary-stack action-stack">
              {completionVerdict?.completed !== undefined ? (
                <span className="status-pill">completed {String(completionVerdict.completed)}</span>
              ) : null}
              {verificationResult?.preferred_verification_channel ? (
                <span className="status-pill">verify {verificationResult.preferred_verification_channel}</span>
              ) : null}
              <span className="status-pill">
                {firstText(acceptanceOverview?.functional_status, acceptanceRun?.functional_status, "functional pending")}
              </span>
              <span className="status-pill">
                {firstText(acceptanceOverview?.production_status, acceptanceRun?.production_status, "production pending")}
              </span>
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

          <div className="metrics-grid">
            <MetricCard
              label="Verification"
              value={firstText(verificationResult?.decision, verificationResult?.status, acceptanceOverview?.overall_status, "pending")}
              tone="calm"
            />
            <MetricCard
              label="Completion"
              value={firstText(completionVerdict?.decision, completionVerdict?.final_status, "pending")}
              tone="neutral"
            />
            <MetricCard
              label="Escalation"
              value={firstText(runtimeEscalation?.reason_class, runtimeEscalation?.status, runtimeEscalation?.policy_denied ? "policy_denied" : undefined, "none")}
              tone="warn"
            />
            <MetricCard
              label="Fault"
              value={firstText(faultSummary?.fault_kind, faultSummary?.status, faultSummary?.fault_loop_detected ? "fault_loop" : undefined, "none")}
              tone="warn"
            />
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Acceptance Commands</h3>
            </div>
            <div className="inline-metrics">
              <span className="status-pill">
                next {firstText(acceptanceNextAction, "inspect_acceptance")}
              </span>
              <span className="status-pill">run {acceptanceRun?.id || "none"}</span>
              <span className="status-pill">task {acceptanceRun?.task_id || "auto"}</span>
              <span className="status-pill">profile {acceptanceRun?.profile_version || "default"}</span>
              <span className="status-pill">
                blockers {firstNumber(acceptanceOverview?.blocking_issue_count, acceptanceRun?.blocking_issue_count, state.data.issues.length)}
              </span>
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
                {acceptanceNextAction === "start_acceptance_run" ? (
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
                {acceptanceNextAction === "apply_manual_release" ? (
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
                {acceptanceNextAction === "resolve_runtime_escalation" ? (
                  <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                    Open Runtime Diagnostics
                  </button>
                ) : null}
                {acceptanceNextAction === "resolve_verification_conflict" || acceptanceNextAction === "manual_checkpoint" ? (
                  <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                    Open Manual Checkpoint
                  </button>
                ) : null}
                {acceptanceNextAction === "collect_evidence" ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(
                        currentAcceptanceTaskId
                          ? buildRoute("/execution", { task: currentAcceptanceTaskId })
                          : routes.diagnostics,
                      )
                    }
                  >
                    Collect Evidence
                  </button>
                ) : null}
                {acceptanceNextAction === "prepare_rework" || acceptanceNextAction === "review_fault_loop" ? (
                  <button className="secondary-button" onClick={() => navigate(planState.data?.repair_draft.id ? routes.repairDraft : routes.diagnostics)}>
                    Open Rework Path
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
            {acceptanceRun?.latest_judgement_summary ? (
              <p className="muted-copy">
                Latest judgement · {acceptanceRun.latest_judgement_kind || "judgement"} · {acceptanceRun.latest_judgement_result || "n/a"} ·{" "}
                {acceptanceRun.latest_judgement_summary}
                {acceptanceRun.latest_judgement_at ? ` · ${acceptanceRun.latest_judgement_at}` : ""}
              </p>
            ) : null}
            {acceptanceRun?.finished_at ? (
              <p className="muted-copy">Acceptance finished at {acceptanceRun.finished_at}</p>
            ) : null}
          </section>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Total-Outline Summary</h3>
                <span className="status-pill">{countOutlineArtifacts(state.data)}</span>
              </div>
              <div className="stack-list">
                <SummaryCard
                  title="Acceptance Overview"
                  primary={firstText(
                    acceptanceOverview?.overall_status,
                    acceptanceOverview?.release_gate_status,
                    releaseGate?.status,
                    "not ready",
                  )}
                  secondary={firstText(
                    releaseGate?.summary,
                    "Structured acceptance overview is present when the backend emits overview fields.",
                  )}
                  pills={[
                    acceptanceOverview?.current_stage ? `stage ${acceptanceOverview.current_stage}` : undefined,
                    acceptanceOverview?.functional_status ? `functional ${acceptanceOverview.functional_status}` : undefined,
                    acceptanceOverview?.production_status ? `production ${acceptanceOverview.production_status}` : undefined,
                    acceptanceNextAction ? `next ${acceptanceNextAction}` : undefined,
                    acceptanceOverview ? `evidence ${acceptanceOverview.evidence_card_count}` : undefined,
                    acceptanceOverview?.manual_release_required !== undefined
                      ? `manual release ${String(acceptanceOverview.manual_release_required)}`
                      : undefined,
                  ]}
                  lines={[
                    acceptanceOverview
                      ? `Coverage items: ${acceptanceOverview.covered_item_count}/${acceptanceOverview.required_item_count}`
                      : undefined,
                    acceptanceOverview ? `Blocking issues: ${acceptanceOverview.blocking_issue_count}` : undefined,
                  ]}
                />
                <SummaryCard
                  title="Verification Result"
                  primary={firstText(verificationResult?.decision, verificationResult?.status, "not ready")}
                  secondary={firstText(verificationResult?.summary, "The current acceptance snapshot does not include a structured verification result.")}
                  pills={[
                    verificationResult?.completed !== undefined ? `completed ${String(verificationResult.completed)}` : undefined,
                    verificationResult?.preferred_verification_channel
                      ? `channel ${verificationResult.preferred_verification_channel}`
                      : undefined,
                    verificationResult?.source_run_id ? `run ${verificationResult.source_run_id}` : undefined,
                    verificationResult?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine("Required checks", verificationResult?.required_checks),
                    formatChecklistLine("Required evidence", verificationResult?.required_evidence),
                    formatChecklistLine("Missing evidence", verificationResult?.missing_evidence),
                    formatChecklistLine("Failed checks", verificationResult?.failed_checks),
                  ]}
                  rawJson={verificationResult?.verification_contract_json}
                  rawJsonLabel="Verification Contract JSON"
                />
                <SummaryCard
                  title="Completion Verdict"
                  primary={firstText(completionVerdict?.decision, completionVerdict?.final_status, "pending")}
                  secondary={firstText(
                    completionVerdict?.summary,
                    completionVerdict?.reason,
                    "The current acceptance snapshot does not include a structured completion verdict.",
                  )}
                  pills={[
                    completionVerdict?.completed !== undefined ? `completed ${String(completionVerdict.completed)}` : undefined,
                    completionVerdict?.release_ready !== undefined ? `release ${String(completionVerdict.release_ready)}` : undefined,
                    completionVerdict?.manual_review_required !== undefined
                      ? `manual review ${String(completionVerdict.manual_review_required)}`
                      : undefined,
                    completionVerdict?.manual_release_required !== undefined
                      ? `manual release ${String(completionVerdict.manual_release_required)}`
                      : undefined,
                    completionVerdict?.manual_release_completed !== undefined
                      ? `release done ${String(completionVerdict.manual_release_completed)}`
                      : undefined,
                    completionVerdict?.blocker_count !== undefined ? `blockers ${completionVerdict.blocker_count}` : undefined,
                    completionVerdict?.next_action ? `next ${completionVerdict.next_action}` : undefined,
                    completionVerdict?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title="Runtime Escalation"
                  primary={firstText(runtimeEscalation?.reason_class, runtimeEscalation?.status, "none")}
                  secondary={firstText(
                    runtimeEscalation?.summary,
                    "No structured runtime escalation is attached to the current acceptance snapshot.",
                  )}
                  pills={[
                    runtimeEscalation?.severity || undefined,
                    runtimeEscalation?.source_brain ? `brain ${runtimeEscalation.source_brain}` : undefined,
                    runtimeEscalation?.source_task_id ? `source task ${runtimeEscalation.source_task_id}` : undefined,
                    runtimeEscalation?.run_binding_id ? `binding ${runtimeEscalation.run_binding_id}` : undefined,
                    runtimeEscalation?.run_status ? `run status ${runtimeEscalation.run_status}` : undefined,
                    runtimeEscalation?.action ? `action ${runtimeEscalation.action}` : undefined,
                    runtimeEscalation?.task_id ? `task ${runtimeEscalation.task_id}` : undefined,
                    runtimeEscalation?.run_id ? `run ${runtimeEscalation.run_id}` : undefined,
                    runtimeEscalation?.policy_denied !== undefined ? `policy denied ${String(runtimeEscalation.policy_denied)}` : undefined,
                    runtimeEscalation?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title="Fault Summary"
                  primary={firstText(faultSummary?.fault_kind, faultSummary?.status, "none")}
                  secondary={firstText(faultSummary?.summary, faultSummary?.top_issue, "No aggregated fault summary is available yet.")}
                  pills={[
                    faultSummary?.severity || undefined,
                    faultSummary?.fault_loop_detected !== undefined ? `fault loop ${String(faultSummary.fault_loop_detected)}` : undefined,
                    faultSummary?.blocking_issue_count !== undefined ? `blocking ${faultSummary.blocking_issue_count}` : undefined,
                    faultSummary?.advisory_issue_count !== undefined ? `advisory ${faultSummary.advisory_issue_count}` : undefined,
                    faultSummary?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine("Failed checks", faultSummary?.failed_checks),
                    formatChecklistLine("Affected tasks", faultSummary?.affected_tasks),
                  ]}
                />
                <SummaryCard
                  title="Repair Plan Draft"
                  primary={firstText(
                    repairPlanDraft?.repair_strategy,
                    repairPlanDraft?.status,
                    planState.data?.repair_draft.status,
                    "idle",
                  )}
                  secondary={
                    firstText(
                      repairPlanDraft?.summary,
                      repairPlanDraft?.reasoning_summary,
                      planState.data?.repair_draft.reasoning_summary,
                      "The current acceptance issues have not produced a structured repair plan draft yet.",
                    )
                  }
                  pills={[
                    repairPlanDraft?.id || planState.data?.repair_draft.id || undefined,
                    repairPlanDraft?.reason_class || undefined,
                    repairPlanDraft?.manual_review_required !== undefined
                      ? `manual review ${String(repairPlanDraft.manual_review_required)}`
                      : undefined,
                    repairPlanDraft?.updated_at || planState.data?.repair_draft.updated_at || undefined,
                  ]}
                  lines={[formatChecklistLine("Updated tasks", repairPlanDraft?.updated_tasks)]}
                />
              </div>
            </section>

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
                    <p>
                      {item.blocking ? "blocking" : "non-blocking"}
                      {item.issue_kind ? ` · ${item.issue_kind}` : ""}
                      {item.created_at ? ` · ${item.created_at}` : ""}
                    </p>
                    {item.detail_json ? (
                      <details>
                        <summary>View issue detail</summary>
                        <pre className="json-block">{prettyJson(item.detail_json)}</pre>
                      </details>
                    ) : null}
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

function MetricCard(props: { label: string; value: string; tone: "calm" | "warn" | "neutral" }) {
  return (
    <article className={`metric-card metric-${props.tone}`}>
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </article>
  );
}

function SummaryCard(props: {
  title: string;
  primary: string;
  secondary: string;
  pills?: Array<string | undefined>;
  lines?: Array<string | undefined>;
  rawJson?: string;
  rawJsonLabel?: string;
}) {
  const pills = (props.pills ?? []).filter(Boolean) as string[];
  const lines = (props.lines ?? []).filter(Boolean) as string[];

  return (
    <article className="list-card">
      <div className="list-card-head">
        <strong>{props.title}</strong>
        <span className="status-pill">{props.primary}</span>
      </div>
      <p>{props.secondary}</p>
      {pills.length > 0 ? (
        <div className="inline-metrics">
          {pills.map((item) => (
            <span key={item} className="status-pill">
              {item}
            </span>
          ))}
        </div>
      ) : null}
      {lines.map((line) => (
        <p key={line} className="muted-copy">
          {line}
        </p>
      ))}
      {props.rawJson ? (
        <details>
          <summary>{props.rawJsonLabel || "View raw JSON"}</summary>
          <pre className="json-block">{prettyJson(props.rawJson)}</pre>
        </details>
      ) : null}
    </article>
  );
}

function countOutlineArtifacts(view: AcceptanceView) {
  return [
    view.verification_result,
    view.completion_verdict,
    view.runtime_escalation,
    view.fault_summary,
    view.repair_plan_draft,
  ].filter(Boolean).length;
}

function formatChecklistLine(label: string, items?: string[]) {
  if (!items || items.length === 0) {
    return undefined;
  }
  return `${label}: ${items.join(" / ")}`;
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

function firstText(...values: Array<string | undefined>) {
  for (const value of values) {
    if (value && value.trim() !== "") {
      return value;
    }
  }
  return "";
}

function firstNumber(...values: Array<number | undefined>) {
  for (const value of values) {
    if (typeof value === "number" && Number.isFinite(value)) {
      return String(value);
    }
  }
  return "0";
}
