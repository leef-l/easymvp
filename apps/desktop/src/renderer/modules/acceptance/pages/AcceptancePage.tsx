import { useEffect, useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type {
  AcceptanceView,
  CommandResponse,
  MultiLayerAcceptanceView,
  PlanView,
  ProjectDiagnosticsView,
} from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function AcceptancePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, searchParams, buildRoute, routes } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("acceptance.noProjectTitle")}</h4>
          <p>{t("acceptance.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("acceptance.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }
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
  const multiLayerState = useQuery(
    () => apiGet<MultiLayerAcceptanceView>(`/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-layers`),
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
  const verificationRunID = firstText(verificationResult?.source_run_id, completionVerdict?.source_run_id, runtimeEscalation?.run_id);
  const runtimeBindingID = runtimeEscalation?.run_binding_id || "";
  const contextTaskID = firstText(currentAcceptanceTaskId, selectedTaskId, selectedTaskFromUrl, runtimeEscalation?.task_id);

  const evidenceGroups = useMemo(() => {
    const groups: Record<string, Array<{ id: string; surface: string; journey?: string; evidence_type: string; file_path: string; captured_at: string }>> = {};
    for (const card of state.data?.evidence_cards ?? []) {
      const group = card.evidence_type || "other";
      if (!groups[group]) groups[group] = [];
      groups[group].push(card);
    }
    return Object.entries(groups);
  }, [state.data?.evidence_cards]);

  const blockingIssues = useMemo(() => {
    return (state.data?.issues ?? []).filter((item) => item.blocking);
  }, [state.data?.issues]);

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
      setActionError(error instanceof Error ? error.message : t("acceptance.actionFailed"));
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
      title={t("acceptance.title")}
      onRetry={() => setRefreshTick((value) => value + 1)}
      secondaryActionLabel={t("acceptance.openDiagnostics")}
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage={t("acceptance.recovery")}
    >
      {state.data ? (
        <section className="acceptance-layout">
          {/* 1. Top Bar */}
          <div className="acceptance-top-bar">
            {/* Left: Verification Result */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("acceptance.verificationResult")}</h3>
                <span className="status-pill">
                  {firstText(verificationResult?.decision, verificationResult?.status, "not ready")}
                </span>
              </div>
              <p>{firstText(verificationResult?.summary, t("acceptance.verificationFallback"))}</p>
              <div className="inline-metrics">
                {verificationResult?.completed !== undefined ? (
                  <span className="status-pill">completed {String(verificationResult.completed)}</span>
                ) : null}
                {verificationResult?.preferred_verification_channel ? (
                  <span className="status-pill">channel {verificationResult.preferred_verification_channel}</span>
                ) : null}
                {verificationResult?.source_run_id ? <span className="status-pill">run {verificationResult.source_run_id}</span> : null}
                {verificationResult?.updated_at ? <span className="status-pill">{verificationResult.updated_at}</span> : null}
              </div>
              {formatChecklistLine(t("acceptance.requiredChecks"), verificationResult?.required_checks) ? (
                <p className="muted-copy">
                  {formatChecklistLine(t("acceptance.requiredChecks"), verificationResult?.required_checks)}
                </p>
              ) : null}
              {formatChecklistLine(t("acceptance.requiredEvidence"), verificationResult?.required_evidence) ? (
                <p className="muted-copy">
                  {formatChecklistLine(t("acceptance.requiredEvidence"), verificationResult?.required_evidence)}
                </p>
              ) : null}
              {formatChecklistLine(t("acceptance.missingEvidence"), verificationResult?.missing_evidence) ? (
                <p className="muted-copy">
                  {formatChecklistLine(t("acceptance.missingEvidence"), verificationResult?.missing_evidence)}
                </p>
              ) : null}
              {formatChecklistLine(t("acceptance.failedChecks"), verificationResult?.failed_checks) ? (
                <p className="muted-copy">
                  {formatChecklistLine(t("acceptance.failedChecks"), verificationResult?.failed_checks)}
                </p>
              ) : null}
              {verificationResult?.verification_contract_json ? (
                <details>
                  <summary>{t("acceptance.verificationContractJson")}</summary>
                  <pre className="json-block">{prettyJson(verificationResult.verification_contract_json)}</pre>
                </details>
              ) : null}
              <div className="action-row" style={{ marginTop: 12 }}>
                {contextTaskID ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(
                        buildRoute("/execution", {
                          task: contextTaskID,
                          run: verificationRunID || undefined,
                          binding: runtimeBindingID || undefined,
                        }),
                      )
                    }
                  >
                    {t("acceptance.openExecution")}
                  </button>
                ) : null}
                <button
                  className="secondary-button"
                  onClick={() =>
                    navigate(
                      buildRoute("/replay", {
                        binding: runtimeBindingID || undefined,
                        run: verificationRunID || undefined,
                        task: contextTaskID || undefined,
                      }),
                    )
                  }
                >
                  {t("acceptance.openReplay")}
                </button>
                <button
                  className="secondary-button"
                  onClick={() =>
                    navigate(
                      buildRoute("/audit", {
                        run: verificationRunID || undefined,
                        task: contextTaskID || undefined,
                      }),
                    )
                  }
                >
                  {t("acceptance.openAudit")}
                </button>
              </div>
            </section>

            {/* Right: Completion Verdict + Commands */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("acceptance.completionVerdict")}</h3>
                <span className="status-pill">
                  {firstText(completionVerdict?.decision, completionVerdict?.final_status, "pending")}
                </span>
              </div>
              <p>
                {firstText(
                  completionVerdict?.summary,
                  completionVerdict?.reason,
                  t("acceptance.completionFallback"),
                )}
              </p>
              <div className="inline-metrics">
                {completionVerdict?.completed !== undefined ? (
                  <span className="status-pill">completed {String(completionVerdict.completed)}</span>
                ) : null}
                {completionVerdict?.release_ready !== undefined ? (
                  <span className="status-pill">release {String(completionVerdict.release_ready)}</span>
                ) : null}
                {completionVerdict?.manual_review_required !== undefined ? (
                  <span className="status-pill">manual review {String(completionVerdict.manual_review_required)}</span>
                ) : null}
                {completionVerdict?.manual_release_required !== undefined ? (
                  <span className="status-pill">manual release {String(completionVerdict.manual_release_required)}</span>
                ) : null}
                {completionVerdict?.blocker_count !== undefined ? (
                  <span className="status-pill">blockers {completionVerdict.blocker_count}</span>
                ) : null}
                {completionVerdict?.next_action ? (
                  <span className="status-pill">next {completionVerdict.next_action}</span>
                ) : null}
              </div>

              <div className="form-grid" style={{ marginTop: 12 }}>
                <label className="form-field">
                  <span>{t("acceptance.targetTask")}</span>
                  <select
                    className="project-input"
                    value={selectedTaskId}
                    onChange={(event) => setSelectedTaskId(event.target.value)}
                  >
                    <option value="">auto / latest</option>
                    {availableTaskOptions.map((item) => (
                      <option key={item.task_id} value={item.mapped_domain_task_id}>
                        {item.task_name} · {item.mapped_domain_task_id}
                      </option>
                    ))}
                  </select>
                </label>
                <label className="form-field">
                  <span>{t("acceptance.manualReleaseComment")}</span>
                  <input
                    className="project-input"
                    value={manualReleaseComment}
                    onChange={(event) => setManualReleaseComment(event.target.value)}
                    placeholder="optional approval note"
                  />
                </label>
              </div>

              <div className="action-row" style={{ marginTop: 12 }}>
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
                  {busyAction === "refresh_acceptance_profiles"
                    ? t("acceptance.refreshing")
                    : t("acceptance.refreshProfiles")}
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
                  {busyAction === "adjudicate_acceptance"
                    ? t("acceptance.adjudicating")
                    : t("acceptance.adjudicateNow")}
                </button>
                <button
                  className="primary-button"
                  disabled={busyAction !== ""}
                  onClick={() =>
                    runAction("start_acceptance", () =>
                      apiPost<CommandResponse>(
                        `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs`,
                        {
                          id: projectId,
                          task_id: selectedTaskId || undefined,
                          mode: "running",
                        },
                      ),
                    )
                  }
                >
                  {busyAction === "start_acceptance"
                    ? t("acceptance.starting")
                    : t("acceptance.startAcceptance")}
                </button>
                <button
                  className="secondary-button"
                  disabled={busyAction !== ""}
                  onClick={() => setRefreshTick((value) => value + 1)}
                >
                  {t("acceptance.refreshView")}
                </button>
                <button
                  className="secondary-button"
                  disabled={busyAction !== ""}
                  onClick={() =>
                    runAction("collect_evidence", () =>
                      apiPost<CommandResponse>(
                        `/api/v3/projects/${encodeURIComponent(projectId)}/evidence/collect`,
                        { id: projectId },
                      ),
                    )
                  }
                >
                  {busyAction === "collect_evidence"
                    ? t("acceptance.refreshingEvidence")
                    : t("acceptance.collectEvidence")}
                </button>
                {acceptanceNextAction === "start_acceptance_run" ? (
                  <button
                    className="secondary-button"
                    disabled={busyAction !== ""}
                    onClick={() =>
                      runAction("start_acceptance_from_gate", () =>
                        apiPost<CommandResponse>(
                          `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs`,
                          {
                            id: projectId,
                            task_id: selectedTaskId || undefined,
                            mode: "running",
                          },
                        ),
                      )
                    }
                  >
                    {t("acceptance.startFromReleaseGate")}
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
                          {
                            id: projectId,
                            comment: manualReleaseComment.trim() || undefined,
                          },
                        ),
                      )
                    }
                  >
                    {busyAction === "apply_manual_release"
                      ? t("acceptance.applying")
                      : t("acceptance.approveManualRelease")}
                  </button>
                ) : null}
                {acceptanceNextAction === "resolve_runtime_escalation" ? (
                  <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                    {t("acceptance.openRuntimeDiagnostics")}
                  </button>
                ) : null}
                {acceptanceNextAction === "resolve_verification_conflict" ||
                acceptanceNextAction === "manual_checkpoint" ? (
                  <button className="secondary-button" onClick={() => navigate(routes.diagnostics)}>
                    {t("acceptance.openManualCheckpoint")}
                  </button>
                ) : null}
                {acceptanceNextAction === "collect_evidence" ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(
                        contextTaskID
                          ? buildRoute("/execution", {
                              task: contextTaskID,
                              run: verificationRunID || undefined,
                              binding: runtimeBindingID || undefined,
                            })
                          : routes.diagnostics,
                      )
                    }
                  >
                    {t("acceptance.collectEvidence")}
                  </button>
                ) : null}
                {acceptanceNextAction === "prepare_rework" ||
                acceptanceNextAction === "review_fault_loop" ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(planState.data?.repair_draft.id ? routes.repairDraft : routes.diagnostics)
                    }
                  >
                    {t("acceptance.openReworkPath")}
                  </button>
                ) : null}
                {contextTaskID ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(
                        buildRoute("/execution", {
                          task: contextTaskID,
                          run: verificationRunID || undefined,
                          binding: runtimeBindingID || undefined,
                        }),
                      )
                    }
                  >
                    {t("acceptance.openTaskExecution")}
                  </button>
                ) : null}
                {verificationRunID || runtimeBindingID || contextTaskID ? (
                  <button
                    className="secondary-button"
                    onClick={() =>
                      navigate(
                        buildRoute("/replay", {
                          binding: runtimeBindingID || undefined,
                          run: verificationRunID || undefined,
                          task: contextTaskID || undefined,
                        }),
                      )
                    }
                  >
                    {t("acceptance.openReplay")}
                  </button>
                ) : null}
                {planState.data?.repair_draft.id ? (
                  <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                    {t("acceptance.openRepairDraft")}
                  </button>
                ) : null}
              </div>
              {actionError ? <p className="error-copy">{actionError}</p> : null}
              {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}
              {acceptanceRun?.latest_judgement_summary ? (
                <p className="muted-copy">
                  {t("acceptance.latestJudgement")} · {acceptanceRun.latest_judgement_kind || "judgement"} ·{" "}
                  {acceptanceRun.latest_judgement_result || "n/a"} · {acceptanceRun.latest_judgement_summary}
                  {acceptanceRun.latest_judgement_at ? ` · ${acceptanceRun.latest_judgement_at}` : ""}
                </p>
              ) : null}
              {acceptanceRun?.finished_at ? (
                <p className="muted-copy">
                  {t("acceptance.finishedAt", { time: acceptanceRun.finished_at })}
                </p>
              ) : null}
            </section>
          </div>

          {/* 2. Middle */}
          <div className="acceptance-middle">
            {/* Left: Coverage */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("acceptance.coverageMatrix")}</h3>
                <span className="status-pill">{state.data.coverage_matrix.length}</span>
              </div>
              {state.data.coverage_matrix.length > 0 ? (
                <div className="coverage-grid">
                  {state.data.coverage_matrix.map((item) => (
                    <article key={`${item.kind}-${item.key}`} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.name}</strong>
                        <span className="status-pill">{item.coverage_status}</span>
                      </div>
                      <p>
                        {item.kind} · evidence {item.evidence_count}
                      </p>
                    </article>
                  ))}
                </div>
              ) : (
                <p className="muted-copy">{t("acceptance.noCoverageMatrix")}</p>
              )}

              <div className="panel-header" style={{ marginTop: 16 }}>
                <h3>{t("acceptance.issues")}</h3>
                <span className="status-pill">{blockingIssues.length}</span>
              </div>
              {blockingIssues.length > 0 ? (
                <div className="stack-list">
                  {blockingIssues.map((item) => (
                    <article key={item.id} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.summary}</strong>
                        <span className={`severity-badge severity-${item.severity}`}>
                          {item.severity}
                        </span>
                      </div>
                      <p>
                        blocking{item.issue_kind ? ` · ${item.issue_kind}` : ""}
                        {item.created_at ? ` · ${item.created_at}` : ""}
                      </p>
                      {item.detail_json ? (
                        <details>
                          <summary>{t("acceptance.viewIssueDetail")}</summary>
                          <pre className="json-block">{prettyJson(item.detail_json)}</pre>
                        </details>
                      ) : null}
                      <div className="action-row">
                        {contextTaskID ? (
                          <button
                            className="secondary-button"
                            onClick={() =>
                              navigate(
                                buildRoute("/execution", {
                                  task: contextTaskID,
                                  run: verificationRunID || undefined,
                                  binding: runtimeBindingID || undefined,
                                }),
                              )
                            }
                          >
                            {t("acceptance.inspectExecution")}
                          </button>
                        ) : null}
                        <button
                          className="secondary-button"
                          onClick={() => navigate(routes.diagnostics)}
                        >
                          {t("acceptance.openDiagnostics")}
                        </button>
                        {planState.data?.repair_draft.id ? (
                          <button
                            className="secondary-button"
                            onClick={() => navigate(routes.repairDraft)}
                          >
                            {t("acceptance.openRepairDraft")}
                          </button>
                        ) : null}
                      </div>
                    </article>
                  ))}
                </div>
              ) : (
                <p className="muted-copy">{t("acceptance.noIssues")}</p>
              )}
            </section>

            {/* Right: Evidence */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("acceptance.evidence")}</h3>
                <span className="status-pill">{state.data.evidence_cards.length}</span>
              </div>
              {state.data.evidence_cards.length > 0 ? (
                <div className="stack-list">
                  {evidenceGroups.map(([group, items]) => (
                    <div key={group}>
                      <h4
                        style={{
                          margin: "12px 0 8px",
                          fontSize: 13,
                          fontWeight: 700,
                          letterSpacing: "0.08em",
                          textTransform: "uppercase",
                          color: "#47607b",
                        }}
                      >
                        {group}
                      </h4>
                      <div className="evidence-grid">
                        {items.map((item) => (
                          <article key={item.id} className="list-card">
                            <div className="list-card-head">
                              <strong>{item.evidence_type}</strong>
                            </div>
                            <p style={{ fontSize: 12, wordBreak: "break-all" }}>
                              {item.file_path}
                            </p>
                            <span
                              className="status-pill"
                              style={{ marginTop: 8, display: "inline-flex" }}
                            >
                              {item.surface || item.journey || "artifact"}
                            </span>
                          </article>
                        ))}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="muted-copy">{t("acceptance.noEvidence")}</p>
              )}
            </section>
          </div>

          {/* 2b. Multi-Layer Acceptance Standards */}
          {multiLayerState.data ? (
            <div className="acceptance-middle">
              <section className="data-panel">
                <div className="panel-header">
                  <h3>Multi-Layer Acceptance</h3>
                  <span className="status-pill">{multiLayerState.data.layers.length} layers</span>
                </div>
                {multiLayerState.data.layers.length > 0 ? (
                  <div className="stack-list">
                    {multiLayerState.data.layers.map((layer) => (
                      <article key={layer.layer} className="list-card">
                        <div className="list-card-head">
                          <strong style={{ textTransform: "uppercase" }}>{layer.layer}</strong>
                          <span className={`status-pill ${layer.status === "passed" ? "pill-success" : layer.status === "failed" ? "pill-danger" : "pill-advisory"}`}>
                            {layer.status}
                          </span>
                        </div>
                        <div className="inline-metrics">
                          <span className="status-pill">passed {layer.passed_count}</span>
                          <span className="status-pill">failed {layer.failed_count}</span>
                          <span className="status-pill">missing {layer.missing_count}</span>
                          <span className="status-pill">total {layer.total_count}</span>
                        </div>
                        {layer.details ? <p className="muted-copy">{layer.details}</p> : null}
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="muted-copy">No acceptance layer results available.</p>
                )}
              </section>

              <section className="data-panel">
                <div className="panel-header">
                  <h3>Contract Gaps</h3>
                  <span className="status-pill">{multiLayerState.data.contract_gaps.length}</span>
                </div>
                {multiLayerState.data.contract_gaps.length > 0 ? (
                  <div className="stack-list">
                    {multiLayerState.data.contract_gaps.map((gap) => (
                      <article key={gap.check_name} className="list-card">
                        <div className="list-card-head">
                          <strong>{gap.check_name}</strong>
                          <span className={`status-pill ${gap.actual_status === "passed" ? "pill-success" : gap.actual_status === "failed" ? "pill-danger" : "pill-advisory"}`}>
                            {gap.actual_status}
                          </span>
                        </div>
                        <p>
                          {gap.required ? "required" : "optional"}
                          {gap.description ? ` · ${gap.description}` : ""}
                        </p>
                      </article>
                    ))}
                  </div>
                ) : (
                  <p className="muted-copy">No contract gaps detected.</p>
                )}

                {multiLayerState.data.repair_loop_progress ? (
                  <>
                    <div className="panel-header" style={{ marginTop: 16 }}>
                      <h3>Repair Loop Progress</h3>
                    </div>
                    <div className="inline-metrics">
                      <span className="status-pill">
                        completed {multiLayerState.data.repair_loop_progress.completed_repairs}/{multiLayerState.data.repair_loop_progress.total_repairs}
                      </span>
                      <span className="status-pill">
                        in progress {multiLayerState.data.repair_loop_progress.in_progress_repairs}
                      </span>
                      <span className="status-pill">
                        failed {multiLayerState.data.repair_loop_progress.failed_repairs}
                      </span>
                    </div>
                    {multiLayerState.data.repair_loop_progress.current_step ? (
                      <p className="muted-copy">
                        Current step: {multiLayerState.data.repair_loop_progress.current_step}
                      </p>
                    ) : null}
                    {multiLayerState.data.repair_loop_progress.updated_at ? (
                      <p className="muted-copy">
                        Updated: {multiLayerState.data.repair_loop_progress.updated_at}
                      </p>
                    ) : null}
                  </>
                ) : null}
              </section>
            </div>
          ) : null}

          {/* 3. Bottom Release Gate */}
          <section className="data-panel acceptance-gate">
            <div className="panel-header">
              <h3>Release Gate</h3>
              <span className="status-pill">{releaseGate?.status || "unknown"}</span>
            </div>
            <p>
              {firstText(
                releaseGate?.summary,
                acceptanceOverview?.release_gate_status,
                t("acceptance.overviewFallback"),
              )}
            </p>
            <div className="inline-metrics">
              <span className="status-pill">
                can release{" "}
                {String(
                  !!(releaseGate?.next_action && releaseGate.next_action !== "blocked"),
                )}
              </span>
              {acceptanceOverview?.manual_release_required !== undefined ? (
                <span className="status-pill">
                  manual release {String(acceptanceOverview.manual_release_required)}
                </span>
              ) : null}
              {completionVerdict?.manual_release_required !== undefined ? (
                <span className="status-pill">
                  manual release {String(completionVerdict.manual_release_required)}
                </span>
              ) : null}
              {completionVerdict?.blocker_count !== undefined ? (
                <span className="status-pill">
                  blockers {completionVerdict.blocker_count}
                </span>
              ) : null}
              {acceptanceOverview?.blocking_issue_count !== undefined ? (
                <span className="status-pill">
                  blocking issues {acceptanceOverview.blocking_issue_count}
                </span>
              ) : null}
            </div>
            <div className="action-row" style={{ marginTop: 12 }}>
              {acceptanceNextAction === "apply_manual_release" ? (
                <button
                  className="primary-button"
                  disabled={busyAction !== ""}
                  onClick={() =>
                    runAction("apply_manual_release", () =>
                      apiPost<CommandResponse>(
                        `/api/v3/projects/${encodeURIComponent(projectId)}/acceptance-runs/manual-release`,
                        {
                          id: projectId,
                          comment: manualReleaseComment.trim() || undefined,
                        },
                      ),
                    )
                  }
                >
                  {busyAction === "apply_manual_release"
                    ? t("acceptance.applying")
                    : t("acceptance.approveManualRelease")}
                </button>
              ) : null}
              {planState.data?.repair_draft.id ? (
                <button
                  className="secondary-button"
                  onClick={() => navigate(routes.repairDraft)}
                >
                  {t("acceptance.openRepairDraft")}
                </button>
              ) : (
                <button
                  className="secondary-button"
                  onClick={() => navigate(routes.plan)}
                >
                  {t("acceptance.openPlan")}
                </button>
              )}
              {contextTaskID ? (
                <button
                  className="secondary-button"
                  onClick={() =>
                    navigate(
                      buildRoute("/execution", {
                        task: contextTaskID,
                        run: verificationRunID || undefined,
                        binding: runtimeBindingID || undefined,
                      }),
                    )
                  }
                >
                  {t("acceptance.openExecution")}
                </button>
              ) : null}
              <button
                className="secondary-button"
                onClick={() =>
                  navigate(
                    buildRoute("/diagnostics", {
                      binding: runtimeBindingID || undefined,
                      run: verificationRunID || undefined,
                      task: contextTaskID || undefined,
                    }),
                  )
                }
              >
                {t("acceptance.openDiagnostics")}
              </button>
            </div>
          </section>
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
  const { t } = useTranslation();
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
          <summary>{props.rawJsonLabel || t("acceptance.viewRawJson")}</summary>
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
