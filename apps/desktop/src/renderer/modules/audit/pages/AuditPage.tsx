import { useMemo, useState } from "react";
import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { AuditLogsView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function AuditPage() {
  const { projectId, routes } = useProjectState();
  const [eventTypeFilter, setEventTypeFilter] = useState("all");
  const [actorFilter, setActorFilter] = useState("all");
  const [searchTerm, setSearchTerm] = useState("");
  const state = useQuery(
    () => apiGet<AuditLogsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/audit-logs?limit=50`),
    [projectId],
  );
  const eventTypeOptions = useMemo(() => buildOptions(state.data?.items.map((item) => item.event_type) ?? []), [state.data?.items]);
  const actorOptions = useMemo(() => buildOptions(state.data?.items.map((item) => item.actor_kind) ?? []), [state.data?.items]);
  const filteredItems = useMemo(() => {
    const keyword = searchTerm.trim().toLowerCase();
    return (state.data?.items ?? []).filter((item) => {
      if (eventTypeFilter !== "all" && item.event_type !== eventTypeFilter) {
        return false;
      }
      if (actorFilter !== "all" && item.actor_kind !== actorFilter) {
        return false;
      }
      if (keyword === "") {
        return true;
      }
      return `${item.summary} ${item.payload_json ?? ""}`.toLowerCase().includes(keyword);
    });
  }, [actorFilter, eventTypeFilter, searchTerm, state.data?.items]);

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
      recoveryMessage="Audit exposes the persisted project log stream backed by current audit facts."
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
              <span className="status-pill">{filteredItems.length}</span>
            </div>
            <div className="toolbar-filters">
              <label className="filter-field">
                <span>Event Type</span>
                <select className="project-input" value={eventTypeFilter} onChange={(event) => setEventTypeFilter(event.target.value)}>
                  <option value="all">all</option>
                  {eventTypeOptions.map((item) => (
                    <option key={item} value={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </label>
              <label className="filter-field">
                <span>Actor</span>
                <select className="project-input" value={actorFilter} onChange={(event) => setActorFilter(event.target.value)}>
                  <option value="all">all</option>
                  {actorOptions.map((item) => (
                    <option key={item} value={item}>
                      {item}
                    </option>
                  ))}
                </select>
              </label>
              <label className="filter-field">
                <span>Search</span>
                <input
                  className="project-input"
                  value={searchTerm}
                  onChange={(event) => setSearchTerm(event.target.value)}
                  placeholder="summary or payload"
                />
              </label>
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
              {filteredItems.map((item) => (
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
              {filteredItems.length === 0 ? (
                <article className="list-card">
                  <p>No audit records match the current filters for this project.</p>
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

function buildOptions(items: string[]) {
  return Array.from(new Set(items.map((item) => item.trim()).filter(Boolean))).sort((left, right) => left.localeCompare(right));
}
