import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { RuntimeHealthView, SystemHealthView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function DiagnosticsPage() {
  const { routes } = useProjectState();
  const systemState = useQuery(() => apiGet<SystemHealthView>("/api/v3/system/healthz"), []);
  const runtimeState = useQuery(() => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"), []);

  return (
    <QueryPanel
      loading={systemState.loading || runtimeState.loading}
      error={systemState.error || runtimeState.error}
      title="Diagnostics"
      onRetry={() => window.location.reload()}
      secondaryActionLabel="Open Settings"
      onSecondaryAction={() => window.location.assign(routes.settings)}
      recoveryMessage="Diagnostics aggregates health endpoints first. Detailed audit and recovery history can be layered on later."
    >
      {systemState.data && runtimeState.data ? (
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
          </div>
        </section>
      ) : null}
    </QueryPanel>
  );
}
