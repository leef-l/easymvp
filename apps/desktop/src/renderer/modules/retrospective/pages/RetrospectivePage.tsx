import { useState } from "react";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { CommandResponse, RetrospectiveView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function RetrospectivePage() {
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
          <p>Select a project to view its retrospective.</p>
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
    () => apiGet<RetrospectiveView>(`/api/v3/projects/${encodeURIComponent(projectId)}/retrospective-view`),
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
      title="Retrospective"
      onRetry={() => setRefreshTick((v) => v + 1)}
      recoveryMessage="Check that the core service is running and the project exists."
    >
      {state.data ? (
        <section className="dashboard-page">
          {/* Header */}
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">MACCS Stage 7</p>
              <h3 className="placeholder-title">Retrospective</h3>
              <p className="placeholder-description">
                Review project performance, lessons learned, and extracted patterns for continuous improvement.
              </p>
            </div>
            <div className="action-row">
              <span className={`status-pill ${state.data.status === "finalized" ? "pill-success" : "pill-advisory"}`}>
                {state.data.status}
              </span>
              <button
                className="primary-button"
                disabled={busyAction !== ""}
                onClick={() =>
                  runAction("generate_retrospective", () =>
                    apiPost<CommandResponse>(
                      `/api/v3/retrospectives/generate`,
                      { project_id: projectId },
                    ),
                  )
                }
              >
                {busyAction === "generate_retrospective" ? "Generating..." : "Generate Retrospective"}
              </button>
              <button className="secondary-button" onClick={() => setRefreshTick((v) => v + 1)}>
                Refresh
              </button>
            </div>
          </div>

          {/* Summary */}
          {state.data.summary ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>Summary</h3>
              </div>
              <p>{state.data.summary}</p>
            </section>
          ) : null}

          {/* Plan vs Actual */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Plan vs Actual</h3>
              <span className="status-pill">{state.data.plan_vs_actual.length} metrics</span>
            </div>
            {state.data.plan_vs_actual.length > 0 ? (
              <div className="stack-list">
                {state.data.plan_vs_actual.map((item) => (
                  <article key={item.metric} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.metric}</strong>
                      {item.delta ? (
                        <span className={`status-pill ${item.delta.startsWith("-") ? "pill-danger" : item.delta.startsWith("+") ? "pill-success" : ""}`}>
                          {item.delta}
                        </span>
                      ) : null}
                    </div>
                    <div className="inline-metrics">
                      <span className="status-pill">planned: {item.planned}</span>
                      <span className="status-pill">actual: {item.actual}</span>
                    </div>
                  </article>
                ))}
              </div>
            ) : (
              <p className="muted-copy">No plan vs actual data available yet.</p>
            )}
          </section>

          <div className="content-grid">
            {/* Success Factors */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>Success Factors</h3>
                <span className="status-pill">{state.data.success_factors.length}</span>
              </div>
              {state.data.success_factors.length > 0 ? (
                <div className="stack-list">
                  {state.data.success_factors.map((item) => (
                    <article key={item.id} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.title}</strong>
                        {item.impact ? <span className="status-pill pill-success">{item.impact}</span> : null}
                      </div>
                      <p>{item.description}</p>
                      {item.tags && item.tags.length > 0 ? (
                        <div className="inline-metrics">
                          {item.tags.map((tag) => (
                            <span key={tag} className="status-pill">{tag}</span>
                          ))}
                        </div>
                      ) : null}
                    </article>
                  ))}
                </div>
              ) : (
                <p className="muted-copy">No success factors identified yet.</p>
              )}
            </section>

            {/* Failure Lessons */}
            <section className="data-panel">
              <div className="panel-header">
                <h3>Failure Lessons</h3>
                <span className="status-pill">{state.data.failure_lessons.length}</span>
              </div>
              {state.data.failure_lessons.length > 0 ? (
                <div className="stack-list">
                  {state.data.failure_lessons.map((item) => (
                    <article key={item.id} className="list-card">
                      <div className="list-card-head">
                        <strong>{item.title}</strong>
                        {item.impact ? <span className="status-pill pill-danger">{item.impact}</span> : null}
                      </div>
                      <p>{item.description}</p>
                      {item.tags && item.tags.length > 0 ? (
                        <div className="inline-metrics">
                          {item.tags.map((tag) => (
                            <span key={tag} className="status-pill">{tag}</span>
                          ))}
                        </div>
                      ) : null}
                    </article>
                  ))}
                </div>
              ) : (
                <p className="muted-copy">No failure lessons identified yet.</p>
              )}
            </section>
          </div>

          {/* Extracted Patterns */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Extracted Patterns</h3>
              <span className="status-pill">{state.data.extracted_patterns.length}</span>
            </div>
            {state.data.extracted_patterns.length > 0 ? (
              <div className="stack-list">
                {state.data.extracted_patterns.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      {item.impact ? <span className="status-pill">{item.impact}</span> : null}
                    </div>
                    <p>{item.description}</p>
                    {item.tags && item.tags.length > 0 ? (
                      <div className="inline-metrics">
                        {item.tags.map((tag) => (
                          <span key={tag} className="status-pill">{tag}</span>
                        ))}
                      </div>
                    ) : null}
                  </article>
                ))}
              </div>
            ) : (
              <p className="muted-copy">No patterns extracted yet. Generate a retrospective to extract patterns.</p>
            )}
          </section>

          {/* Statistics */}
          <section className="data-panel">
            <div className="panel-header">
              <h3>Statistics</h3>
            </div>
            <div className="metrics-grid">
              <MetricCard
                label="Task Completion"
                value={`${state.data.stats.task_completion_rate.toFixed(1)}%`}
                tone={state.data.stats.task_completion_rate >= 90 ? "calm" : state.data.stats.task_completion_rate >= 70 ? "neutral" : "warn"}
              />
              <MetricCard
                label="Total Cost"
                value={`$${state.data.stats.total_cost_usd.toFixed(2)}`}
                tone="neutral"
              />
              <MetricCard
                label="Duration"
                value={formatDuration(state.data.stats.total_duration_seconds)}
                tone="neutral"
              />
              <MetricCard
                label="Turns"
                value={String(state.data.stats.total_turns)}
                tone="neutral"
              />
              <MetricCard
                label="Tokens"
                value={formatNumber(state.data.stats.total_tokens)}
                tone="neutral"
              />
            </div>

            {/* Brain Usage Breakdown */}
            {Object.keys(state.data.stats.brain_usage).length > 0 ? (
              <>
                <div className="panel-header" style={{ marginTop: 16 }}>
                  <h3>Brain Usage</h3>
                </div>
                <div className="stack-list">
                  {Object.entries(state.data.stats.brain_usage).map(([brain, count]) => (
                    <article key={brain} className="list-card">
                      <div className="list-card-head">
                        <strong>{brain}</strong>
                        <span className="status-pill">{count} invocations</span>
                      </div>
                    </article>
                  ))}
                </div>
              </>
            ) : null}
          </section>

          {/* Action Errors/Messages */}
          {actionError ? <p className="error-copy">{actionError}</p> : null}
          {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}

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
