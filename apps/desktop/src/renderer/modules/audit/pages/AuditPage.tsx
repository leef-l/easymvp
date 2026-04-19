import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { AuditLogsView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function AuditPage() {
  const { projectId, routes } = useProjectState();
  const state = useQuery(
    () => apiGet<AuditLogsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/audit-logs?limit=50`),
    [projectId],
  );

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title="Audit"
      onRetry={() => window.location.reload()}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => window.location.assign(routes.diagnostics)}
      recoveryMessage="Audit currently exposes a thin project log stream backed by persisted audit facts."
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Audit</p>
              <h3 className="placeholder-title">Project audit log</h3>
              <p className="placeholder-description">
                Showing the latest persisted project audit records for {projectId}.
              </p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">{state.data.items.length} items</span>
              <span className="status-pill">{state.data.refresh_hint}</span>
            </div>
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Audit Records</h3>
            </div>
            <div className="stack-list">
              <div className="action-row">
                <a className="secondary-button" href={routes.replay}>
                  Open Replay
                </a>
                <a className="secondary-button" href={routes.execution}>
                  Open Execution
                </a>
              </div>
              {state.data.items.map((item) => (
                <article key={item.id} className="list-card">
                  <div className="list-card-head">
                    <strong>{item.summary}</strong>
                    <span className="status-pill">{item.event_type}</span>
                  </div>
                  <p>
                    {item.actor_kind} · {item.created_at}
                  </p>
                  {item.payload_json ? <pre className="json-block">{prettyJson(item.payload_json)}</pre> : null}
                </article>
              ))}
              {state.data.items.length === 0 ? (
                <article className="list-card">
                  <p>No audit records have been persisted for this project yet.</p>
                </article>
              ) : null}
            </div>
          </section>
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
