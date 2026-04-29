import { useMemo, useState } from "react";
import { useTranslation } from "react-i18next";
import { Link, useNavigate } from "react-router-dom";
import { apiGet } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { AuditLogsView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function AuditPage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, routes, searchParams, buildRoute } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("audit.noProjectTitle")}</h4>
          <p>{t("audit.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("audit.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }
  const runFromUrl = searchParams.get("run")?.trim() || "";
  const taskFromUrl = searchParams.get("task")?.trim() || "";
  const [eventTypeFilter, setEventTypeFilter] = useState("all");
  const [actorFilter, setActorFilter] = useState("all");
  const [searchTerm, setSearchTerm] = useState(runFromUrl || taskFromUrl);
  const [retryTick, setRetryTick] = useState(0);
  const state = useQuery(
    () => apiGet<AuditLogsView>(`/api/v3/projects/${encodeURIComponent(projectId)}/audit-logs?limit=50`),
    [projectId, retryTick],
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
  const activeContext = searchTerm.trim() || runFromUrl || taskFromUrl;

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title={t("audit.title")}
      onRetry={() => setRetryTick((v) => v + 1)}
      secondaryActionLabel={t("audit.openDiagnostics")}
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage={t("audit.recovery")}
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">{t("audit.title")}</p>
              <h3 className="placeholder-title">{t("audit.subtitle")}</h3>
              <p className="placeholder-description">
                {t("audit.description", { projectId })}
              </p>
            </div>
            <div className="summary-stack">
              <span className="status-pill">{state.data.items.length} items</span>
              <span className="status-pill">{state.data.refresh_hint}</span>
              {activeContext ? <span className="status-pill">context {activeContext}</span> : null}
            </div>
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Audit Records</h3>
              <span className="status-pill">{filteredItems.length}</span>
            </div>
            <div className="toolbar-filters">
              <label className="filter-field">
                <span>{t("audit.eventType")}</span>
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
                <span>{t("audit.actor")}</span>
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
                  placeholder={t("audit.filterPlaceholder")}
                />
              </label>
            </div>
            <div className="stack-list">
              <div className="action-row">
                <Link
                  className="secondary-button"
                  to={buildRoute("/replay", {
                    run: runFromUrl || undefined,
                    task: taskFromUrl || undefined,
                  })}
                >
                  Open Replay
                </Link>
                <Link
                  className="secondary-button"
                  to={buildRoute("/execution", {
                    run: runFromUrl || undefined,
                    task: taskFromUrl || undefined,
                  })}
                >
                  Open Execution
                </Link>
                <Link
                  className="secondary-button"
                  to={buildRoute("/diagnostics", {
                    run: runFromUrl || undefined,
                    task: taskFromUrl || undefined,
                  })}
                >
                  Open Diagnostics
                </Link>
              </div>
              {filteredItems.map((item) => {
                const auditContext = readAuditContext(item.payload_json);
                const runID = auditContext.run_id || runFromUrl;
                const taskID = auditContext.task_id || taskFromUrl;
                return (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className="status-pill">{item.event_type}</span>
                    </div>
                    <p>
                      {item.actor_kind} · {item.created_at}
                    </p>
                    {runID || taskID || auditContext.binding_id || auditContext.replay_id ? (
                      <p>
                        {taskID ? `task ${taskID}` : "task n/a"} · {runID ? `run ${runID}` : "run n/a"} ·{" "}
                        {auditContext.binding_id ? `binding ${auditContext.binding_id}` : "binding n/a"}
                      </p>
                    ) : null}
                    <div className="action-row">
                      <Link
                        className="secondary-button"
                        to={buildRoute("/execution", {
                          binding: auditContext.binding_id || undefined,
                          run: runID || undefined,
                          task: taskID || undefined,
                        })}
                      >
                        Open Execution
                      </Link>
                      <Link
                        className="secondary-button"
                        to={buildRoute("/replay", {
                          binding: auditContext.binding_id || undefined,
                          run: runID || undefined,
                          replay: auditContext.replay_id || undefined,
                          task: taskID || undefined,
                        })}
                      >
                        Open Replay
                      </Link>
                      <Link
                        className="secondary-button"
                        to={buildRoute("/diagnostics", {
                          binding: auditContext.binding_id || undefined,
                          run: runID || undefined,
                          task: taskID || undefined,
                        })}
                      >
                        Open Diagnostics
                      </Link>
                      {taskID ? (
                        <Link className="secondary-button" to={buildRoute("/acceptance", { task: taskID })}>
                          Open Acceptance
                        </Link>
                      ) : null}
                    </div>
                    {item.payload_json ? <pre className="json-block">{prettyJson(item.payload_json)}</pre> : null}
                  </article>
                );
              })}
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

function readAuditContext(raw?: string) {
  const parsed = safeParseJSONObject(raw);
  return {
    binding_id: readStringField(parsed, "binding_id", "run_binding_id"),
    replay_id: readStringField(parsed, "replay_id", "latest_replay_id"),
    run_id: readStringField(parsed, "run_id", "source_run_id"),
    task_id: readStringField(parsed, "task_id", "source_task_id", "domain_task_id"),
  };
}

function safeParseJSONObject(raw?: string): Record<string, unknown> | null {
  if (!raw || raw.trim() === "") {
    return null;
  }
  try {
    const parsed = JSON.parse(raw);
    if (parsed && typeof parsed === "object" && !Array.isArray(parsed)) {
      return parsed as Record<string, unknown>;
    }
  } catch {
    return null;
  }
  return null;
}

function readStringField(source: Record<string, unknown> | null, ...keys: string[]) {
  if (!source) {
    return "";
  }
  for (const key of keys) {
    const value = source[key];
    if (typeof value === "string" && value.trim() !== "") {
      return value.trim();
    }
  }
  return "";
}
