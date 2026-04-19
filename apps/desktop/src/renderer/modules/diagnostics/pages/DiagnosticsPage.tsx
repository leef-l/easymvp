import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { ProjectDiagnosticsView, RuntimeHealthView, SystemHealthView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function DiagnosticsPage() {
  const { projectId, routes, buildRoute } = useProjectState();
  const systemState = useQuery(() => apiGet<SystemHealthView>("/api/v3/system/healthz"), []);
  const runtimeState = useQuery(() => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"), []);
  const diagnosticsState = useQuery(
    () => apiGet<ProjectDiagnosticsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/diagnostic-records?limit=12`),
    [projectId],
  );

  return (
    <QueryPanel
      loading={systemState.loading || runtimeState.loading || diagnosticsState.loading}
      error={systemState.error || runtimeState.error || diagnosticsState.error}
      title="Diagnostics"
      onRetry={() => window.location.reload()}
      secondaryActionLabel="Open Settings"
      onSecondaryAction={() => window.location.assign(routes.settings)}
      recoveryMessage="Diagnostics aggregates health endpoints first. Detailed audit and recovery history can be layered on later."
    >
      {systemState.data && runtimeState.data && diagnosticsState.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Diagnostics</p>
              <h3 className="placeholder-title">Recovery and health entry point</h3>
              <p className="placeholder-description">
                Use this page when a workbench page is stale, a runtime action fails, or local core connectivity needs verification.
              </p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">system {systemState.data.status}</span>
              <span className="status-pill">runtime {runtimeState.data.status}</span>
              <span className="status-pill">{diagnosticsState.data.items.length} diagnostics</span>
            </div>
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>System Health</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{systemState.data.service}</strong>
                    <span className="status-pill">{systemState.data.status}</span>
                  </div>
                  <p>{systemState.data.version}</p>
                  <p>{systemState.data.timestamp}</p>
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Runtime Health</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>brain-v3 runtime</strong>
                    <span className="status-pill">{runtimeState.data.status}</span>
                  </div>
                  <p>{runtimeState.data.base_url}</p>
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Recovery Actions</h3>
              </div>
              <div className="stack-list">
                <a className="action-card" href={routes.workspace}>
                  <div className="action-card-head">
                    <strong>Back to Workspace</strong>
                    <span className="status-pill">workspace</span>
                  </div>
                  <p>Return to the project cockpit after checking system and runtime health.</p>
                </a>
                <a className="action-card" href={routes.execution}>
                  <div className="action-card-head">
                    <strong>Inspect Execution</strong>
                    <span className="status-pill">execution</span>
                  </div>
                  <p>Open the runtime run board and verify the latest binding, replay, and raw logs.</p>
                </a>
                <a className="action-card" href={routes.replay}>
                  <div className="action-card-head">
                    <strong>Open Replay</strong>
                    <span className="status-pill">replay</span>
                  </div>
                  <p>Jump directly into replay timeline inspection when the runtime is healthy but artifact indexing needs review.</p>
                </a>
                <a className="action-card" href={routes.audit}>
                  <div className="action-card-head">
                    <strong>Open Audit</strong>
                    <span className="status-pill">audit</span>
                  </div>
                  <p>Review persisted audit facts and release actions before returning to execution or workspace triage.</p>
                </a>
                <a className="action-card" href={routes.settings}>
                  <div className="action-card-head">
                    <strong>Edit Settings</strong>
                    <span className="status-pill">settings</span>
                  </div>
                  <p>Adjust the current project id or local core base URL if the shell points to the wrong instance.</p>
                </a>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Recent Diagnostics</h3>
                <span className="status-pill">{diagnosticsState.data.refresh_hint}</span>
              </div>
              <div className="stack-list">
                {diagnosticsState.data.items.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className={`severity-badge severity-${item.severity}`}>{item.severity}</span>
                    </div>
                    <p>
                      {item.scope} · {item.error_code} · {item.created_at}
                    </p>
                    {(item.project_id || item.task_id || item.run_id || item.binding_id) ? (
                      <p>
                        {item.project_id ? `project ${item.project_id}` : ""}
                        {item.task_id ? ` · task ${item.task_id}` : ""}
                        {item.run_id ? ` · run ${item.run_id}` : ""}
                        {item.binding_id ? ` · binding ${item.binding_id}` : ""}
                      </p>
                    ) : null}
                    <div className="action-row">
                      <a className="secondary-button" href={routes.workspace}>
                        Open Workspace
                      </a>
                      {item.task_id ? (
                        <a className="secondary-button" href={buildRoute("/execution", { task: item.task_id })}>
                          Open Execution
                        </a>
                      ) : (
                        <a className="secondary-button" href={routes.execution}>
                          Open Execution
                        </a>
                      )}
                      <a className="secondary-button" href={routes.audit}>
                        Open Audit
                      </a>
                    </div>
                    {item.detail_json ? <pre className="json-block">{prettyJson(item.detail_json)}</pre> : null}
                  </article>
                ))}
                {diagnosticsState.data.items.length === 0 ? (
                  <article className="list-card">
                    <p>No diagnostic records have been captured for this project yet.</p>
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
