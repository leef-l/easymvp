import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, DeliveryView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function DeliveryPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [busyAction, setBusyAction] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>No Project Selected</h4>
          <p>Select a project to view its delivery status.</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              Go to Projects
            </button>
          </div>
        </div>
      </section>
    );
  }

  const state = useQuery(
    () => apiGet<DeliveryView>(`/api/v3/projects/${encodeURIComponent(projectId)}/delivery-view`),
    [projectId, refreshTick],
  );

  async function runAction(actionKey: string, execute: () => Promise<CommandResponse>) {
    setBusyAction(actionKey);
    setActionError("");
    setActionMessage("");
    try {
      const result = await execute();
      setActionMessage(`${actionKey} accepted · ${result.next_action}`);
      setRefreshTick((v) => v + 1);
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
      title="Delivery"
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage="Check that the core service is running and the project exists."
    >
      {state.data ? (
        <section className="dashboard-page">
          {/* Header */}
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">MACCS Stage 6</p>
              <h3 className="placeholder-title">Delivery</h3>
              <p className="placeholder-description">
                Review deliverables, test reports, and project statistics before accepting or rejecting the delivery.
              </p>
            </div>
            <div className="action-row">
              <span className={`status-pill ${state.data.status === "accepted" ? "pill-success" : state.data.status === "rejected" ? "pill-danger" : "pill-advisory"}`}>
                {state.data.status}
              </span>
              <button className="secondary-button" onClick={() => setRefreshTick((v) => v + 1)}>
                Refresh
              </button>
            </div>
          </div>

          {/* Acceptance Summary */}
          {state.data.acceptance_summary ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>Acceptance Summary</h3>
              </div>
              <p>{state.data.acceptance_summary}</p>
            </section>
          ) : null}

          {/* Deliverable Artifacts */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Deliverable Artifacts</h3>
              <span className="status-pill">{state.data.artifacts.length}</span>
            </div>
            {state.data.artifacts.length > 0 ? (
              <div className="stack-list">
                {state.data.artifacts.map((artifact) => (
                  <article key={`${artifact.kind}-${artifact.path}`} className="list-card">
                    <div className="list-card-head">
                      <strong>{artifactKindLabel(artifact.kind)}</strong>
                      <span className={`status-pill ${artifact.status === "available" ? "pill-success" : artifact.status === "missing" ? "pill-danger" : "pill-advisory"}`}>
                        {artifact.status}
                      </span>
                    </div>
                    <p style={{ fontSize: 12, wordBreak: "break-all" }}>{artifact.path}</p>
                    {artifact.description ? <p className="muted-copy">{artifact.description}</p> : null}
                  </article>
                ))}
              </div>
            ) : (
              <p className="muted-copy">No deliverable artifacts registered.</p>
            )}
          </section>

          {/* Test Reports */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Test Reports</h3>
              <span className="status-pill">{state.data.test_reports.length} layers</span>
            </div>
            {state.data.test_reports.length > 0 ? (
              <div className="stack-list">
                {state.data.test_reports.map((report) => (
                  <article key={report.layer} className="list-card">
                    <div className="list-card-head">
                      <strong style={{ textTransform: "uppercase" }}>{report.layer}</strong>
                      <span className={`status-pill ${report.pass_rate >= 100 ? "pill-success" : report.pass_rate >= 80 ? "pill-advisory" : "pill-danger"}`}>
                        {report.pass_rate.toFixed(1)}%
                      </span>
                    </div>
                    <div className="inline-metrics">
                      <span className="status-pill">total {report.total}</span>
                      <span className="status-pill">passed {report.passed}</span>
                      <span className="status-pill">failed {report.failed}</span>
                      <span className="status-pill">skipped {report.skipped}</span>
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <p className="muted-copy">No test reports available.</p>
            )}
          </section>

          {/* Project Statistics */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Project Statistics</h3>
            </div>
            <div className="metrics-grid">
              <MetricCard
                label="Tasks"
                value={`${state.data.project_stats.completed_tasks}/${state.data.project_stats.total_tasks}`}
                tone={state.data.project_stats.completed_tasks === state.data.project_stats.total_tasks ? "calm" : "warn"}
              />
              <MetricCard
                label="Turns"
                value={String(state.data.project_stats.total_turns)}
                tone="neutral"
              />
              <MetricCard
                label="Tokens"
                value={formatNumber(state.data.project_stats.total_tokens)}
                tone="neutral"
              />
              <MetricCard
                label="Duration"
                value={formatDuration(state.data.project_stats.elapsed_seconds)}
                tone="neutral"
              />
              {state.data.project_stats.cost_usd !== undefined ? (
                <MetricCard
                  label="Cost"
                  value={`$${state.data.project_stats.cost_usd.toFixed(2)}`}
                  tone="neutral"
                />
              ) : null}
            </div>
          </section>

          {/* Actions */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Delivery Actions</h3>
            </div>
            <div className="action-row">
              <button
                className="primary-button"
                disabled={busyAction !== "" || state.data.status === "accepted"}
                onClick={() =>
                  runAction("accept_delivery", () =>
                    apiPost<CommandResponse>(
                      `/api/v3/deliveries/${encodeURIComponent(state.data!.id)}/accept`,
                      { project_id: projectId },
                    ),
                  )
                }
              >
                {busyAction === "accept_delivery" ? "Accepting..." : "Accept Delivery"}
              </button>
              <button
                className="secondary-button danger-button"
                disabled={busyAction !== "" || state.data.status === "rejected"}
                onClick={() =>
                  runAction("reject_delivery", () =>
                    apiPost<CommandResponse>(
                      `/api/v3/deliveries/${encodeURIComponent(state.data!.id)}/reject`,
                      { project_id: projectId },
                    ),
                  )
                }
              >
                {busyAction === "reject_delivery" ? "Rejecting..." : "Reject Delivery"}
              </button>
              <button
                className="secondary-button"
                onClick={() => navigate(routes.acceptance)}
              >
                Back to Acceptance
              </button>
            </div>
            {actionError ? <p className="error-copy">{actionError}</p> : null}
            {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}
          </section>

          {/* Timestamps */}
          {state.data.created_at || state.data.updated_at ? (
            <div className="inline-metrics" style={{ marginTop: 8 }}>
              {state.data.created_at ? <span className="status-pill">created {state.data.created_at}</span> : null}
              {state.data.updated_at ? <span className="status-pill">updated {state.data.updated_at}</span> : null}
            </div>
          ) : null}
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

function artifactKindLabel(kind: string): string {
  const labels: Record<string, string> = {
    code: "Source Code",
    readme: "README",
    architecture_doc: "Architecture Documentation",
    api_doc: "API Documentation",
    deploy_guide: "Deployment Guide",
  };
  return labels[kind] || kind;
}

function formatNumber(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(1)}K`;
  return String(n);
}

function formatDuration(seconds: number): string {
  if (seconds < 60) return `${seconds}s`;
  if (seconds < 3600) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  const h = Math.floor(seconds / 3600);
  const m = Math.floor((seconds % 3600) / 60);
  return `${h}h ${m}m`;
}
