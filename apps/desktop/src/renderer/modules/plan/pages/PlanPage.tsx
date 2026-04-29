import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, PlanView, RedesignPlanResponse } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function PlanPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes, buildRoute } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("plan.noProjectTitle")}</h4>
          <p>{t("plan.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("plan.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }
  const [refreshTick, setRefreshTick] = useState(0);
  const [busyAction, setBusyAction] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const [repairPayload, setRepairPayload] = useState({
    failed_task_context_json: "{\"task_id\":\"\"}",
    failure_reason_json: "{\"summary\":\"\"}",
    original_contracts_json: "{\"contracts\":[]}",
    runtime_summary_json: "{\"status\":\"failed\"}",
    artifact_refs_json: "[]",
  });
  const [redesignFeedback, setRedesignFeedback] = useState("");
  const [showRedesignInput, setShowRedesignInput] = useState(false);
  const state = useQuery(
    () => apiGet<PlanView>(`/api/v3/projects/${encodeURIComponent(projectId)}/plan-view`),
    [projectId, refreshTick],
  );

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

  function compilePlan() {
    if (state.data && (state.data.review.decision === "rejected" || state.data.review.blocking_issue_count > 0)) {
      setActionError(t("plan.reviewBlocked"));
      return;
    }
    void runAction("compile_plan", () =>
      apiPost<CommandResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/plan/compile`, {
        id: projectId,
        force_recompile: true,
      }),
    );
  }

  async function redesignPlan() {
    await runAction("redesign_plan", () =>
      apiPost<RedesignPlanResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/plan/redesign`, {
        feedback: redesignFeedback,
      }),
    );
    setShowRedesignInput(false);
    setRedesignFeedback("");
  }

  async function createRepairDraft() {
    let artifactRefs: unknown = [];
    try {
      artifactRefs = JSON.parse(repairPayload.artifact_refs_json);
    } catch {
      setActionError("artifact_refs_json must be valid JSON");
      return;
    }
    await runAction("create_repair_draft", () =>
      apiPost<CommandResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/repair-draft`, {
        id: projectId,
        failed_task_context_json: repairPayload.failed_task_context_json,
        failure_reason_json: repairPayload.failure_reason_json,
        original_contracts_json: repairPayload.original_contracts_json,
        runtime_summary_json: repairPayload.runtime_summary_json,
        artifact_refs: artifactRefs,
      }),
    );
  }

  return (
    <QueryPanel loading={state.loading} refreshing={state.refreshing} stale={state.stale} error={state.error} title={t("plan.title")}>
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("plan.title")}</p>
              <h3 className="placeholder-title">{state.data.draft.goal_summary || t("plan.overview")}</h3>
              <p className="placeholder-description">{state.data.diff_summary.summary}</p>
            </div>
            <div className="action-row">
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={() => void compilePlan()}
              >
                {busyAction === "compile_plan" ? t("plan.compiling") : t("plan.compilePlan")}
              </button>
              {state.data.repair_draft.id ? (
                <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                  {t("plan.openRepairDraft")}
                </button>
              ) : null}
            </div>
          </div>

          {actionError ? <section className="data-panel is-error"><p className="error-copy">{actionError}</p></section> : null}
          {actionMessage ? <section className="data-panel"><p className="muted-copy">{actionMessage}</p></section> : null}

          <div className="plan-triple-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("plan.draft")}</h3>
                <span className="status-pill">{state.data.draft.status}</span>
              </div>
              <div style={{ padding: "12px 16px" }}>
                <p className="muted-copy">{state.data.draft.goal_summary || t("plan.overview")}</p>
                <p className="muted-copy">v{state.data.draft.version}</p>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("plan.review")}</h3>
                <span className="status-pill">{state.data.review.decision || "—"}</span>
              </div>
              <div style={{ padding: "12px 16px" }}>
                <div className="action-row">
                  <span className="status-pill pill-blocking">
                    {state.data.review.blocking_issue_count} {t("plan.blockingIssues")}
                  </span>
                  <span className="status-pill pill-advisory">
                    {state.data.review.advisory_issue_count} {t("plan.advisoryIssues")}
                  </span>
                </div>
                {state.data.review.split_suggestions_json ? (
                  <div style={{ marginTop: 12 }}>
                    <p className="muted-copy">{t("plan.rewriteHints")}</p>
                    <ul style={{ margin: "8px 0", paddingLeft: 16 }}>
                      {(() => {
                        try {
                          const parsed = JSON.parse(state.data.review.split_suggestions_json);
                          const items = Array.isArray(parsed) ? parsed : [parsed];
                          return items.map((item, idx) => (
                            <li key={idx} className="muted-copy">
                              {typeof item === "string" ? item : JSON.stringify(item)}
                            </li>
                          ));
                        } catch {
                          return <li className="muted-copy">{state.data.review.split_suggestions_json}</li>;
                        }
                      })()}
                    </ul>
                  </div>
                ) : null}
                {(state.data.review.decision === "rejected" || state.data.review.blocking_issue_count > 0) ? (
                  <div style={{ marginTop: 12 }}>
                    {!showRedesignInput ? (
                      <button
                        className="secondary-button"
                        disabled={busyAction !== ""}
                        onClick={() => setShowRedesignInput(true)}
                      >
                        {t("plan.redesignPlan")}
                      </button>
                    ) : (
                      <div style={{ display: "flex", flexDirection: "column", gap: 8 }}>
                        <p className="muted-copy">{t("plan.redesignHint")}</p>
                        <input
                          className="project-input"
                          type="text"
                          placeholder={t("plan.redesignPlaceholder")}
                          value={redesignFeedback}
                          onChange={(e) => setRedesignFeedback(e.target.value)}
                        />
                        <div className="action-row">
                          <button
                            className="secondary-button"
                            disabled={busyAction !== ""}
                            onClick={() => void redesignPlan()}
                          >
                            {busyAction === "redesign_plan" ? t("plan.redesigning") : t("plan.redesignPlan")}
                          </button>
                          <button
                            className="secondary-button"
                            disabled={busyAction !== ""}
                            onClick={() => {
                              setShowRedesignInput(false);
                              setRedesignFeedback("");
                            }}
                          >
                            {t("projects.cancel")}
                          </button>
                        </div>
                      </div>
                    )}
                  </div>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("plan.compiled")}</h3>
                <span className="status-pill">{state.data.compiled.status}</span>
              </div>
              <div style={{ padding: "12px 16px" }}>
                <p className="muted-copy">v{state.data.compiled.compiled_version || 0}</p>
                {state.data.compiled.risk_summary ? (
                  <pre className="json-block">{state.data.compiled.risk_summary}</pre>
                ) : null}
              </div>
            </section>
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("plan.createRepairDraft")}</h3>
            </div>
            <div className="form-grid">
              <label className="form-field">
                <span>Failed Task Context JSON</span>
                <textarea
                  className="project-input project-textarea"
                  value={repairPayload.failed_task_context_json}
                  onChange={(event) => setRepairPayload((current) => ({ ...current, failed_task_context_json: event.target.value }))}
                />
              </label>
              <label className="form-field">
                <span>Failure Reason JSON</span>
                <textarea
                  className="project-input project-textarea"
                  value={repairPayload.failure_reason_json}
                  onChange={(event) => setRepairPayload((current) => ({ ...current, failure_reason_json: event.target.value }))}
                />
              </label>
              <label className="form-field">
                <span>Original Contracts JSON</span>
                <textarea
                  className="project-input project-textarea"
                  value={repairPayload.original_contracts_json}
                  onChange={(event) => setRepairPayload((current) => ({ ...current, original_contracts_json: event.target.value }))}
                />
              </label>
              <label className="form-field">
                <span>Runtime Summary JSON</span>
                <textarea
                  className="project-input project-textarea"
                  value={repairPayload.runtime_summary_json}
                  onChange={(event) => setRepairPayload((current) => ({ ...current, runtime_summary_json: event.target.value }))}
                />
              </label>
              <label className="form-field">
                <span>Artifact Refs JSON</span>
                <textarea
                  className="project-input project-textarea"
                  value={repairPayload.artifact_refs_json}
                  onChange={(event) => setRepairPayload((current) => ({ ...current, artifact_refs_json: event.target.value }))}
                />
              </label>
              <div className="action-row">
                <button className="primary-button" disabled={busyAction !== ""} onClick={() => void createRepairDraft()}>
                  {busyAction === "create_repair_draft" ? "Creating..." : "Create Repair Draft"}
                </button>
              </div>
            </div>
          </section>

          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("plan.taskProjection")}</h3>
              <span className="status-pill">{state.data.task_projection.length}</span>
            </div>
            <div className="stack-list">
              {state.data.task_projection.map((item) => (
                <article key={item.task_id} className="list-card">
                  <div className="list-card-head">
                    <strong>{item.task_name}</strong>
                    <div className="action-row">
                      <span className="status-pill">{item.status}</span>
                      {item.manual_review_required ? (
                        <span className="status-pill pill-manual-review">{t("plan.manualReviewRequired")}</span>
                      ) : null}
                    </div>
                  </div>
                  <p>{item.role_type} · {item.brain_kind} · {item.risk_level}</p>
                  <p>{item.delivery_summary}</p>
                  <p>
                    domain {item.mapped_domain_task_id || "unmapped"} · state {item.mapped_domain_task_status || "n/a"}
                  </p>
                  <p>verify {item.verification_summary || "n/a"} · manual review {item.manual_review_required ? "yes" : "no"}</p>
                  <p>resources {(item.affected_resources ?? []).join(", ") || "none"}</p>
                  <div className="action-row">
                    {item.mapped_domain_task_id ? (
                      <button
                        className="secondary-button"
                        onClick={() => navigate(buildRoute("/acceptance", { task: item.mapped_domain_task_id || "" }))}
                      >
                        Start Acceptance
                      </button>
                    ) : null}
                    <button className="secondary-button" onClick={() => navigate(routes.execution)}>
                      Open Execution
                    </button>
                  </div>
                </article>
              ))}
              {state.data.task_projection.length === 0 ? (
                <article className="list-card">
                  <p>{t("plan.noTaskProjection")}</p>
                  <div className="action-row" style={{ marginTop: 8 }}>
                    <button
                      className="primary-button"
                      disabled={busyAction !== ""}
                      onClick={() => void compilePlan()}
                    >
                      {busyAction === "compile_plan" ? t("plan.compiling") : t("plan.compilePlan")}
                    </button>
                  </div>
                </article>
              ) : null}
            </div>
          </section>

          <section className="data-panel">
            <div className="panel-header">
              <h3>{t("plan.diffSummary")}</h3>
              {state.data.diff_summary.review_issue_count > 0 ? (
                <span className="status-pill pill-advisory">{state.data.diff_summary.review_issue_count} {t("plan.reviewIssues")}</span>
              ) : null}
            </div>
            <div className="inline-metrics">
              <span className="status-pill">split {state.data.diff_summary.split_count}</span>
              <span className="status-pill">override {state.data.diff_summary.override_count}</span>
              <span className="status-pill">drop {state.data.diff_summary.drop_count}</span>
              <span className="status-pill">issues {state.data.diff_summary.review_issue_count}</span>
            </div>
            <div className="stack-list">
              {(state.data.diff_summary.items || []).map((item, index) => (
                <article key={`${item.diff_kind}-${index}`} className="list-card">
                  <div className="list-card-head">
                    <strong>{item.after_label || item.before_label || item.diff_kind}</strong>
                    <div className="action-row">
                      <span className="status-pill">{item.diff_kind}</span>
                      {item.source_review_issue_id ? (
                        <span className="status-pill pill-advisory">{t("plan.linkedReviewIssue")}</span>
                      ) : null}
                    </div>
                  </div>
                  <p>{item.reason || "No explanation provided."}</p>
                </article>
              ))}
              {!(state.data.diff_summary.items || []).length ? (
                <article className="list-card">
                  <p>{t("plan.noDiffItems")}</p>
                  <div className="action-row" style={{ marginTop: 8 }}>
                    <button
                      className="primary-button"
                      disabled={busyAction !== ""}
                      onClick={() => void compilePlan()}
                    >
                      {busyAction === "compile_plan" ? t("plan.compiling") : t("plan.compilePlan")}
                    </button>
                  </div>
                </article>
              ) : null}
            </div>
          </section>
        </section>
      ) : null}
    </QueryPanel>
  );
}
