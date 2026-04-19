import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, PlanView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function PlanPage() {
  const navigate = useNavigate();
  const { projectId, routes, buildRoute } = useProjectState();
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
    <QueryPanel loading={state.loading} refreshing={state.refreshing} stale={state.stale} error={state.error} title="Plan view">
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Plan</p>
              <h3 className="placeholder-title">{state.data.draft.goal_summary || "Plan overview"}</h3>
              <p className="placeholder-description">{state.data.diff_summary.summary}</p>
            </div>
            <div className="action-row">
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={() =>
                  runAction("compile_plan", () =>
                    apiPost<CommandResponse>(`/api/v3/projects/${encodeURIComponent(projectId)}/plan/compile`, {
                      id: projectId,
                      force_recompile: true,
                    }),
                  )
                }
              >
                {busyAction === "compile_plan" ? "Compiling..." : "Compile Plan"}
              </button>
              {state.data.repair_draft.id ? (
                <button className="secondary-button" onClick={() => navigate(routes.repairDraft)}>
                  Open repair draft
                </button>
              ) : null}
            </div>
          </div>

          {actionError ? <section className="data-panel is-error"><p className="error-copy">{actionError}</p></section> : null}
          {actionMessage ? <section className="data-panel"><p className="muted-copy">{actionMessage}</p></section> : null}

          <div className="metrics-grid">
            <MetricCard label="Draft" value={state.data.draft.status} />
            <MetricCard label="Review" value={state.data.review.decision} />
            <MetricCard label="Compiled" value={`${state.data.compiled.status} · v${state.data.compiled.compiled_version || 0}`} />
            <MetricCard label="Repair" value={state.data.repair_draft.status || "idle"} />
          </div>

          {state.data.compiled.risk_summary ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>Compiled Risk Summary</h3>
              </div>
              <pre className="json-block">{state.data.compiled.risk_summary}</pre>
            </section>
          ) : null}

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Diff Summary</h3>
              </div>
              <div className="inline-metrics">
                <span className="status-pill">split {state.data.diff_summary.split_count}</span>
                <span className="status-pill">override {state.data.diff_summary.override_count}</span>
                <span className="status-pill">drop {state.data.diff_summary.drop_count}</span>
                <span className="status-pill">issues {state.data.diff_summary.review_issue_count}</span>
              </div>
              <div className="stack-list">
                {state.data.diff_summary.items.map((item, index) => (
                  <article key={`${item.diff_kind}-${index}`} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.after_label || item.before_label || item.diff_kind}</strong>
                      <span className="status-pill">{item.diff_kind}</span>
                    </div>
                    <p>{item.reason || "No explanation provided."}</p>
                  </article>
                ))}
                {state.data.diff_summary.items.length === 0 ? (
                  <article className="list-card">
                    <p>No diff summary items are available for the current plan.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Task Projection</h3>
                <span className="status-pill">{state.data.task_projection.length}</span>
              </div>
              <div className="stack-list">
                {state.data.task_projection.map((item) => (
                  <article key={item.task_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.task_name}</strong>
                      <span className="status-pill">{item.status}</span>
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
                    <p>No compiled task projection is available yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Create Repair Draft</h3>
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
          </div>
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
