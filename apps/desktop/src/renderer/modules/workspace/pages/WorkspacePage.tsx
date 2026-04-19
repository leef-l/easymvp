import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, getCoreBaseUrl } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { WorkspaceHomeView, WorkspaceView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function WorkspacePage() {
  const navigate = useNavigate();
  const { projectId, updateProjectId, routes } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [streamState, setStreamState] = useState<"idle" | "connecting" | "live" | "disconnected">("idle");
  const [streamEvents, setStreamEvents] = useState<Array<{ id: string; type: string; summary: string }>>([]);
  const state = useQuery(
    () => apiGet<WorkspaceView>(`/api/v3/projects/${encodeURIComponent(projectId)}/workspace-view`),
    [projectId, refreshTick],
  );
  const homeState = useQuery(() => apiGet<WorkspaceHomeView>("/api/v3/workspace/home-view"), [refreshTick]);

  useEffect(() => {
    setStreamState("connecting");
    const lastEventIdKey = `easymvp.workspace.events.${projectId}`;
    const lastEventID = window.sessionStorage.getItem(lastEventIdKey)?.trim();
    const base = getCoreBaseUrl();
    const url = new URL(`/api/v3/workspace/projects/${encodeURIComponent(projectId)}/events`, base);
    if (lastEventID) {
      url.searchParams.set("last_event_id", lastEventID);
    }

    const source = new EventSource(url.toString());
    source.onopen = () => {
      setStreamState("live");
    };
    source.onerror = () => {
      setStreamState("disconnected");
    };
    source.onmessage = (event) => {
      window.sessionStorage.setItem(lastEventIdKey, event.lastEventId || "");
      setStreamEvents((current) => [
        {
          id: event.lastEventId || `${Date.now()}`,
          type: "message",
          summary: event.data || "workspace event",
        },
        ...current,
      ].slice(0, 8));
      setRefreshTick((value) => value + 1);
    };
    source.addEventListener("workspace.snapshot_invalidated", (event) => {
      const typedEvent = event as MessageEvent<string>;
      window.sessionStorage.removeItem(lastEventIdKey);
      setStreamEvents((current) => [
        {
          id: typedEvent.lastEventId || `${Date.now()}`,
          type: "workspace.snapshot_invalidated",
          summary: typedEvent.data || "Workspace snapshot invalidated",
        },
        ...current,
      ].slice(0, 8));
      setRefreshTick((value) => value + 1);
    });

    return () => {
      source.close();
    };
  }, [projectId]);

  function handleAction(actionKey: string, options?: { deepLink?: string; targetId?: string }) {
    const deepLink = options?.deepLink?.trim();
    const explicitTargetId = options?.targetId?.trim();
    const targetProjectId = deepLink && !deepLink.startsWith("/") ? deepLink : explicitTargetId || projectId;

    if (deepLink) {
      if (deepLink.startsWith("/")) {
        navigate(deepLink);
        return;
      }
    }

    if (targetProjectId !== projectId) {
      updateProjectId(targetProjectId);
    }

    switch (actionKey) {
      case "open_repair_draft":
        navigate(buildProjectRoute(targetProjectId, "repair-draft"));
        return;
      case "open_project_plan":
        navigate(buildProjectRoute(targetProjectId, "plan"));
        return;
      case "open_task_review":
        navigate(buildProjectRoute(targetProjectId, "execution"));
        return;
      case "open_acceptance_center":
      case "open_acceptance_issue":
        navigate(buildProjectRoute(targetProjectId, "acceptance"));
        return;
      default:
        navigate(buildProjectRoute(targetProjectId, "workspace"));
    }
  }

  return (
    <QueryPanel
      loading={state.loading}
      refreshing={state.refreshing}
      stale={state.stale}
      error={state.error}
      title="Workspace overview"
      onRetry={() => setRefreshTick((value) => value + 1)}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage="Workspace keeps the last successful snapshot when realtime refresh fails."
    >
      {state.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Workspace</p>
              <h3 className="placeholder-title">{state.data.workspace_explanation.headline || state.data.project_snapshot.name || projectId}</h3>
              <p className="placeholder-description">{state.data.workspace_explanation.summary}</p>
            </div>
            <div className="summary-stack action-stack">
              <span className="status-pill">{state.data.project_snapshot.current_stage}</span>
              <span className="status-pill">core {getCoreBaseUrl()}</span>
              <span className="status-pill">stream {streamState}</span>
              <button className="secondary-button" onClick={() => setRefreshTick((value) => value + 1)}>
                Refresh View
              </button>
            </div>
          </div>

          <section className="data-panel">
            <div className="panel-header">
              <h3>Workspace Stream</h3>
              <span className="status-pill">{streamEvents.length}</span>
            </div>
            <div className="stack-list">
              {streamEvents.map((item) => (
                <article key={item.id} className="list-card">
                  <div className="list-card-head">
                    <strong>{item.type}</strong>
                    <span className="status-pill">{item.id}</span>
                  </div>
                  <p>{item.summary}</p>
                </article>
              ))}
              {streamEvents.length === 0 ? (
                <article className="list-card">
                  <p>No streamed workspace events have been observed in this session yet.</p>
                </article>
              ) : null}
            </div>
          </section>

          <div className="metrics-grid">
            <MetricCard label="Progress" value={`${state.data.project_snapshot.progress_percent}%`} tone="calm" />
            <MetricCard label="Risk" value={state.data.project_snapshot.risk_level} tone="warn" />
            <MetricCard label="Production" value={state.data.project_snapshot.production_status} tone="neutral" />
            <MetricCard
              label="Evidence"
              value={`${state.data.acceptance_coverage.evidence_ready}/${state.data.acceptance_coverage.evidence_required}`}
              tone="calm"
            />
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Action Inbox</h3>
                <span className="status-pill">{state.data.action_inbox.length}</span>
              </div>
              <div className="stack-list">
                {state.data.action_inbox.map((item) => (
                  <button
                    key={item.item_id}
                    className="action-card"
                    onClick={() => handleAction(item.recommended_action, { targetId: item.target_id })}
                  >
                    <div className="action-card-head">
                      <strong>{item.title}</strong>
                      <span className={`severity-badge severity-${item.severity}`}>{item.severity}</span>
                    </div>
                    <p>{item.recommended_action}</p>
                  </button>
                ))}
                {state.data.action_inbox.length === 0 ? (
                  <article className="list-card">
                    <p>No blocking actions in the inbox.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Recommended Actions</h3>
                <span className="status-pill">{state.data.workspace_explanation.recommended_actions.length}</span>
              </div>
              <div className="stack-list">
                {state.data.workspace_explanation.recommended_actions.map((item) => (
                  <button
                    key={`${item.action_key}-${item.label}`}
                    className="action-card"
                    onClick={() => handleAction(item.action_key, { deepLink: item.deep_link })}
                  >
                    <div className="action-card-head">
                      <strong>{item.label}</strong>
                      <span className="status-pill">{item.action_key}</span>
                    </div>
                    <p>{item.reason}</p>
                  </button>
                ))}
                {state.data.workspace_explanation.recommended_actions.length === 0 ? (
                  <article className="list-card">
                    <p>No recommended actions from the current workspace summary.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Active Projects</h3>
                <span className="status-pill">{homeState.data?.active_projects.length ?? 0}</span>
              </div>
              {homeState.error ? <p className="error-copy">{homeState.error}</p> : null}
              {homeState.stale ? <p className="muted-copy">Showing the last successful workspace home snapshot.</p> : null}
              <div className="stack-list">
                {(homeState.data?.active_projects ?? []).map((item) => (
                  <button key={item.project_id} className="action-card" onClick={() => updateProjectId(item.project_id)}>
                    <div className="action-card-head">
                      <strong>{item.name}</strong>
                      <span className="status-pill">{item.current_stage}</span>
                    </div>
                    <p>{item.project_category} · progress {item.progress_percent}% · {item.production_status}</p>
                  </button>
                ))}
                {(homeState.data?.active_projects.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>No active projects are available in the current workspace.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Stage Progress</h3>
              </div>
              <div className="stack-list">
                {state.data.stage_progress.map((item) => (
                  <article key={item.stage_key} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.stage_name}</strong>
                      <span className="status-pill">{item.status}</span>
                    </div>
                    <p>{item.active_item_title || "No active item"}</p>
                  </article>
                ))}
                {state.data.stage_progress.length === 0 ? (
                  <article className="list-card">
                    <p>No stage progress has been recorded yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Live Activity</h3>
              </div>
              <div className="stack-list">
                {state.data.live_activity.map((item) => (
                  <article key={item.event_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      <span className="status-pill">{item.event_type}</span>
                    </div>
                    <p>{item.source_brain} · {item.occurred_at}</p>
                  </article>
                ))}
                {state.data.live_activity.length === 0 ? (
                  <article className="list-card">
                    <p>No live activity is available yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Top Blockers</h3>
              </div>
              <div className="stack-list">
                {state.data.workspace_explanation.top_blockers.map((item) => (
                  <article key={item} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {state.data.workspace_explanation.top_blockers.length === 0 ? (
                  <article className="list-card">
                    <p>No blockers are reported in the current workspace explanation.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Explain Links</h3>
                <span className="status-pill">{state.data.workspace_explanation.explain_links?.length ?? 0}</span>
              </div>
              <div className="stack-list">
                {(state.data.workspace_explanation.explain_links ?? []).map((item) => (
                  <article key={item} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {(state.data.workspace_explanation.explain_links?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>No explanation links are attached to the current workspace summary.</p>
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

function MetricCard(props: { label: string; value: string; tone: "calm" | "warn" | "neutral" }) {
  return (
    <article className={`metric-card metric-${props.tone}`}>
      <span>{props.label}</span>
      <strong>{props.value}</strong>
    </article>
  );
}

function buildProjectRoute(projectId: string, section: "workspace" | "plan" | "execution" | "acceptance" | "repair-draft") {
  return `/${section}?project=${encodeURIComponent(projectId)}`;
}
