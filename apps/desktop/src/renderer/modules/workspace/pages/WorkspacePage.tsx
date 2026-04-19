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
  const workspaceOverview = state.data?.overview;
  const projectSnapshot = state.data?.project_snapshot;
  const verificationResult = state.data?.verification_result;
  const completionVerdict = state.data?.completion_verdict;
  const runtimeEscalation = state.data?.runtime_escalation;
  const faultSummary = state.data?.fault_summary;
  const repairPlanDraft = state.data?.repair_plan_draft;
  const workspaceFlags = [
    workspaceOverview?.manual_review_required || projectSnapshot?.manual_review_required ? "manual review required" : undefined,
    workspaceOverview?.verification_conflict || projectSnapshot?.verification_conflict ? "verification conflict" : undefined,
    workspaceOverview?.fault_loop_detected || projectSnapshot?.fault_loop_detected ? "fault loop detected" : undefined,
    workspaceOverview?.policy_denied || projectSnapshot?.policy_denied ? "policy denied" : undefined,
  ].filter(Boolean) as string[];

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
    const targetTaskId = explicitTargetId && explicitTargetId !== projectId ? explicitTargetId : undefined;

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
        navigate(buildProjectRoute(targetProjectId, "execution", { task: targetTaskId }));
        return;
      case "open_acceptance_center":
        navigate(buildProjectRoute(targetProjectId, "acceptance", { task: targetTaskId }));
        return;
      case "open_acceptance_issue":
        navigate(buildProjectRoute(targetProjectId, "acceptance", { task: targetTaskId }));
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
              <h3 className="placeholder-title">{firstText(state.data.workspace_explanation.headline, projectSnapshot?.name, projectId)}</h3>
              <p className="placeholder-description">
                {firstText(completionVerdict?.summary, verificationResult?.summary, state.data.workspace_explanation.summary)}
              </p>
            </div>
            <div className="summary-stack action-stack">
              <span className="status-pill">{firstText(workspaceOverview?.current_stage, projectSnapshot?.current_stage, "unknown")}</span>
              {firstText(completionVerdict?.decision, completionVerdict?.final_status) ? (
                <span className="status-pill">{firstText(completionVerdict?.decision, completionVerdict?.final_status)}</span>
              ) : null}
              {verificationResult?.preferred_verification_channel ? (
                <span className="status-pill">verify {verificationResult.preferred_verification_channel}</span>
              ) : null}
              {workspaceOverview?.next_action ? <span className="status-pill">next {workspaceOverview.next_action}</span> : null}
              <span className="status-pill">core {getCoreBaseUrl()}</span>
              <span className="status-pill">stream {streamState}</span>
              <button className="secondary-button" onClick={() => setRefreshTick((value) => value + 1)}>
                Refresh View
              </button>
            </div>
          </div>

          {workspaceFlags.length > 0 ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>Workspace Guards</h3>
                <span className="status-pill">{workspaceFlags.length}</span>
              </div>
              <div className="inline-metrics">
                {workspaceFlags.map((item) => (
                  <span key={item} className="severity-badge severity-warning">
                    {item}
                  </span>
                ))}
              </div>
            </section>
          ) : null}

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
            <MetricCard label="Progress" value={`${projectSnapshot?.progress_percent ?? 0}%`} tone="calm" />
            <MetricCard label="Risk" value={firstText(workspaceOverview?.risk_level, projectSnapshot?.risk_level, "unknown")} tone="warn" />
            <MetricCard
              label="Verification"
              value={firstText(verificationResult?.decision, verificationResult?.status, workspaceOverview?.stage_status, projectSnapshot?.production_status, "pending")}
              tone="neutral"
            />
            <MetricCard
              label="Completion"
              value={firstText(completionVerdict?.decision, completionVerdict?.final_status, workspaceOverview?.production_status, projectSnapshot?.production_status, "pending")}
              tone="neutral"
            />
            <MetricCard
              label="Evidence"
              value={`${state.data.acceptance_coverage.evidence_ready}/${state.data.acceptance_coverage.evidence_required}`}
              tone="calm"
            />
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Total-Outline Snapshot</h3>
                <span className="status-pill">{countOutlineArtifacts(state.data)}</span>
              </div>
              <div className="stack-list">
                <SummaryCard
                  title="Workspace Overview"
                  primary={firstText(workspaceOverview?.stage_status, workspaceOverview?.production_status, projectSnapshot?.production_status, "legacy only")}
                  secondary={firstText(
                    state.data.workspace_explanation.summary,
                    "Structured workspace overview becomes primary when the backend emits overview fields.",
                  )}
                  pills={[
                    workspaceOverview?.current_stage ? `stage ${workspaceOverview.current_stage}` : undefined,
                    workspaceOverview?.risk_level ? `risk ${workspaceOverview.risk_level}` : undefined,
                    workspaceOverview?.acceptance_run_status ? `acceptance ${workspaceOverview.acceptance_run_status}` : undefined,
                    workspaceOverview?.next_action ? `next ${workspaceOverview.next_action}` : undefined,
                    workspaceOverview ? `actions ${workspaceOverview.action_required_count}` : undefined,
                    workspaceOverview?.manual_release_required !== undefined
                      ? `manual release ${String(workspaceOverview.manual_release_required)}`
                      : undefined,
                  ]}
                  lines={[
                    workspaceOverview ? `Blocking issues: ${workspaceOverview.blocking_issue_count}` : undefined,
                    formatBooleanLine("Manual review required", workspaceOverview?.manual_review_required),
                    formatBooleanLine("Verification conflict", workspaceOverview?.verification_conflict),
                    formatBooleanLine("Fault loop detected", workspaceOverview?.fault_loop_detected),
                    formatBooleanLine("Policy denied", workspaceOverview?.policy_denied),
                  ]}
                />
                <SummaryCard
                  title="Verification Result"
                  primary={firstText(verificationResult?.decision, verificationResult?.status, "legacy only")}
                  secondary={
                    verificationResult?.summary ||
                    "Workspace is still falling back to production_status because no structured verification result is attached."
                  }
                  pills={[
                    verificationResult?.completed !== undefined
                      ? `completed ${String(verificationResult.completed)}`
                      : undefined,
                    verificationResult?.preferred_verification_channel
                      ? `channel ${verificationResult.preferred_verification_channel}`
                      : undefined,
                    verificationResult?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine("Required checks", verificationResult?.required_checks),
                    formatChecklistLine("Missing evidence", verificationResult?.missing_evidence),
                    formatChecklistLine("Failed checks", verificationResult?.failed_checks),
                  ]}
                />
                <SummaryCard
                  title="Completion Verdict"
                  primary={firstText(completionVerdict?.decision, completionVerdict?.final_status, "not emitted")}
                  secondary={firstText(completionVerdict?.summary, completionVerdict?.reason, "Workspace explanation is still the compatibility fallback.")}
                  pills={[
                    completionVerdict?.completed !== undefined
                      ? `completed ${String(completionVerdict.completed)}`
                      : undefined,
                    completionVerdict?.release_ready !== undefined
                      ? `release ${String(completionVerdict.release_ready)}`
                      : undefined,
                    completionVerdict?.manual_review_required !== undefined
                      ? `manual review ${String(completionVerdict.manual_review_required)}`
                      : undefined,
                    completionVerdict?.manual_release_required !== undefined
                      ? `manual release ${String(completionVerdict.manual_release_required)}`
                      : undefined,
                    completionVerdict?.manual_release_completed !== undefined
                      ? `release done ${String(completionVerdict.manual_release_completed)}`
                      : undefined,
                    completionVerdict?.blocker_count !== undefined
                      ? `blockers ${completionVerdict.blocker_count}`
                      : undefined,
                    completionVerdict?.next_action ? `next ${completionVerdict.next_action}` : undefined,
                    completionVerdict?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title="Runtime Escalation"
                  primary={firstText(runtimeEscalation?.reason_class, runtimeEscalation?.status, "none")}
                  secondary={firstText(runtimeEscalation?.summary, "No structured runtime escalation is attached to this workspace snapshot.")}
                  pills={[
                    runtimeEscalation?.severity || undefined,
                    runtimeEscalation?.source_brain ? `brain ${runtimeEscalation.source_brain}` : undefined,
                    runtimeEscalation?.source_task_id ? `source task ${runtimeEscalation.source_task_id}` : undefined,
                    runtimeEscalation?.run_binding_id ? `binding ${runtimeEscalation.run_binding_id}` : undefined,
                    runtimeEscalation?.run_status ? `run status ${runtimeEscalation.run_status}` : undefined,
                    runtimeEscalation?.action ? `action ${runtimeEscalation.action}` : undefined,
                    runtimeEscalation?.task_id ? `task ${runtimeEscalation.task_id}` : undefined,
                    runtimeEscalation?.run_id ? `run ${runtimeEscalation.run_id}` : undefined,
                    runtimeEscalation?.policy_denied !== undefined ? `policy denied ${String(runtimeEscalation.policy_denied)}` : undefined,
                    runtimeEscalation?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title="Fault Summary"
                  primary={firstText(faultSummary?.fault_kind, faultSummary?.status, "none")}
                  secondary={firstText(faultSummary?.summary, faultSummary?.top_issue, "No aggregated fault summary is attached to the current workspace view.")}
                  pills={[
                    faultSummary?.severity || undefined,
                    faultSummary?.fault_loop_detected !== undefined ? `fault loop ${String(faultSummary.fault_loop_detected)}` : undefined,
                    faultSummary?.blocking_issue_count !== undefined ? `blocking ${faultSummary.blocking_issue_count}` : undefined,
                    faultSummary?.advisory_issue_count !== undefined ? `advisory ${faultSummary.advisory_issue_count}` : undefined,
                    faultSummary?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine("Failed checks", faultSummary?.failed_checks),
                    formatChecklistLine("Affected tasks", faultSummary?.affected_tasks),
                  ]}
                />
                <SummaryCard
                  title="Repair Plan Draft"
                  primary={firstText(repairPlanDraft?.repair_strategy, repairPlanDraft?.status, "idle")}
                  secondary={firstText(
                    repairPlanDraft?.summary,
                    repairPlanDraft?.reasoning_summary,
                    "The workspace has not received a structured repair draft summary yet.",
                  )}
                  pills={[
                    repairPlanDraft?.id || undefined,
                    repairPlanDraft?.reason_class || undefined,
                    repairPlanDraft?.manual_review_required !== undefined
                      ? `manual review ${String(repairPlanDraft.manual_review_required)}`
                      : undefined,
                    repairPlanDraft?.updated_at || undefined,
                  ]}
                  lines={[formatChecklistLine("Updated tasks", repairPlanDraft?.updated_tasks)]}
                />
              </div>
            </section>

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
                      <span className={`severity-badge severity-${item.severity}`}>{formatInboxBadge(item)}</span>
                    </div>
                    <p>
                      {describeInboxAction(item)} · {item.recommended_action}
                    </p>
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
                    <p>{item.project_category} · progress {item.progress_percent}% · {firstText(item.stage_status, item.production_status, "pending")}</p>
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
                  <button
                    key={item.event_id}
                    className="action-card"
                    onClick={() =>
                      navigate(
                        item.source_task_id
                          ? buildProjectRoute(projectId, "execution", { task: item.source_task_id })
                          : buildProjectRoute(projectId, "execution"),
                      )
                    }
                  >
                    <div className="list-card-head">
                      <strong>{item.title}</strong>
                      <span className={item.requires_action ? "severity-badge severity-warning" : "status-pill"}>
                        {item.requires_action ? "needs action" : item.event_type}
                      </span>
                    </div>
                    <p>
                      {item.source_brain} · {item.occurred_at}
                      {item.source_task_id ? ` · task ${item.source_task_id}` : ""}
                    </p>
                  </button>
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
                  <button key={item} className="action-card" onClick={() => handleExplainLink(item, projectId, navigate)}>
                    <div className="action-card-head">
                      <strong>{formatExplainLinkLabel(item)}</strong>
                      <span className="status-pill">{item}</span>
                    </div>
                    <p>{describeExplainLink(item)}</p>
                  </button>
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

function SummaryCard(props: {
  title: string;
  primary: string;
  secondary: string;
  pills?: Array<string | undefined>;
  lines?: Array<string | undefined>;
}) {
  const pills = (props.pills ?? []).filter(Boolean) as string[];
  const lines = (props.lines ?? []).filter(Boolean) as string[];

  return (
    <article className="list-card">
      <div className="list-card-head">
        <strong>{props.title}</strong>
        <span className="status-pill">{props.primary}</span>
      </div>
      <p>{props.secondary}</p>
      {pills.length > 0 ? (
        <div className="inline-metrics">
          {pills.map((item) => (
            <span key={item} className="status-pill">
              {item}
            </span>
          ))}
        </div>
      ) : null}
      {lines.map((line) => (
        <p key={line} className="muted-copy">
          {line}
        </p>
      ))}
    </article>
  );
}

function countOutlineArtifacts(view: WorkspaceView) {
  return [
    view.verification_result,
    view.completion_verdict,
    view.runtime_escalation,
    view.fault_summary,
    view.repair_plan_draft,
  ].filter(Boolean).length;
}

function formatChecklistLine(label: string, items?: string[]) {
  if (!items || items.length === 0) {
    return undefined;
  }
  return `${label}: ${items.join(" / ")}`;
}

function buildProjectRoute(
  projectId: string,
  section: "workspace" | "plan" | "execution" | "acceptance" | "repair-draft",
  extraParams?: Record<string, string | undefined>,
) {
  const search = new URLSearchParams();
  search.set("project", projectId);
  for (const [key, value] of Object.entries(extraParams ?? {})) {
    if (!value) {
      continue;
    }
    search.set(key, value);
  }
  return `/${section}?${search.toString()}`;
}

function formatInboxBadge(item: WorkspaceView["action_inbox"][number]) {
  return item.is_blocking ? `${item.severity} · blocking` : item.severity;
}

function describeInboxAction(item: WorkspaceView["action_inbox"][number]) {
  if (item.target_id) {
    return `target ${item.target_id}`;
  }
  return item.is_blocking ? "blocking attention item" : "recommended follow-up";
}

function handleExplainLink(linkKey: string, projectId: string, navigate: ReturnType<typeof useNavigate>) {
  switch (linkKey) {
    case "runtime":
      navigate(`/execution?project=${encodeURIComponent(projectId)}`);
      return;
    case "task_review":
      navigate(`/execution?project=${encodeURIComponent(projectId)}`);
      return;
    case "diagnostics":
      navigate(`/diagnostics?project=${encodeURIComponent(projectId)}`);
      return;
    case "replay":
      navigate(`/replay?project=${encodeURIComponent(projectId)}`);
      return;
    case "audit":
      navigate(`/audit?project=${encodeURIComponent(projectId)}`);
      return;
    case "acceptance":
      navigate(`/acceptance?project=${encodeURIComponent(projectId)}`);
      return;
    case "plan":
      navigate(`/plan?project=${encodeURIComponent(projectId)}`);
      return;
    default:
      navigate(`/workspace?project=${encodeURIComponent(projectId)}`);
  }
}

function formatExplainLinkLabel(linkKey: string) {
  switch (linkKey) {
    case "runtime":
      return "Inspect runtime state";
    case "task_review":
      return "Review affected task";
    case "diagnostics":
      return "Open diagnostics";
    case "replay":
      return "Inspect replay timeline";
    case "audit":
      return "Review audit records";
    case "acceptance":
      return "Open acceptance center";
    case "plan":
      return "Open project plan";
    default:
      return "Open workspace detail";
  }
}

function describeExplainLink(linkKey: string) {
  switch (linkKey) {
    case "runtime":
      return "Jump to the execution board and verify the latest runtime binding, events, and replay artifacts.";
    case "task_review":
      return "Open task execution review when workspace explanation recommends manual follow-up.";
    case "diagnostics":
      return "Check runtime and system health before retrying the current project workflow.";
    case "replay":
      return "Inspect replay artifacts and timeline entries for the most recent run.";
    case "audit":
      return "Review persisted audit facts and release actions for this project.";
    case "acceptance":
      return "Open acceptance commands, gate status, and issue coverage.";
    case "plan":
      return "Return to the latest compiled plan and repair draft context.";
    default:
      return "Open the project workspace view.";
  }
}

function firstText(...values: Array<string | undefined>) {
  for (const value of values) {
    if (value && value.trim() !== "") {
      return value;
    }
  }
  return "";
}

function formatBooleanLine(label: string, value?: boolean) {
  if (value === undefined) {
    return undefined;
  }
  return `${label}: ${String(value)}`;
}
