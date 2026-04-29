import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useTranslation } from "react-i18next";
import { apiDelete, apiGet, getCoreBaseUrl } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type { WorkspaceHomeView, WorkspaceView } from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function WorkspacePage() {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { projectId, updateProjectId, routes } = useProjectState();

  if (!projectId) {
    return (
      <section className="placeholder-page">
        <div className="empty-state-panel">
          <h4>{t("workspace.noProjectTitle")}</h4>
          <p>{t("workspace.noProjectDescription")}</p>
          <div className="action-row">
            <button className="primary-button" onClick={() => navigate("/projects")}>
              {t("workspace.goToProjects")}
            </button>
            <button className="secondary-button" onClick={() => navigate("/settings")}>
              {t("settings.createProject")}
            </button>
          </div>
        </div>
      </section>
    );
  }
  const [refreshTick, setRefreshTick] = useState(0);
  const [streamState, setStreamState] = useState<"idle" | "connecting" | "live" | "disconnected">("idle");
  const [streamEvents, setStreamEvents] = useState<Array<{ id: string; type: string; summary: string }>>([]);
  const [expandedSections, setExpandedSections] = useState<Record<string, boolean>>({});
  const toggleSection = (key: string) => setExpandedSections((prev) => ({ ...prev, [key]: !prev[key] }));
  const [onboardingDismissed, setOnboardingDismissed] = useState(false);
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
  const isOnboarded = window.localStorage.getItem(`easymvp.project.onboarded.${projectId}`) === "true";
  const isEarlyStage =
    projectSnapshot?.current_stage === "created" ||
    projectSnapshot?.production_status === "pending" ||
    (projectSnapshot?.progress_percent ?? 0) === 0;
  const showOnboarding = Boolean(state.data) && !isOnboarded && !onboardingDismissed && isEarlyStage;
  const workspaceNextAction = firstText(completionVerdict?.next_action);
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
    const base = getCoreBaseUrl() || window.location.origin;
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

  useEffect(() => {
    setOnboardingDismissed(false);
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
      case "prepare_rework":
        navigate(buildProjectRoute(targetProjectId, "repair-draft"));
        return;
      case "open_project_plan":
        navigate(buildProjectRoute(targetProjectId, "plan"));
        return;
      case "open_task_review":
        navigate(buildProjectRoute(targetProjectId, "execution", { task: targetTaskId }));
        return;
      case "open_acceptance_center":
      case "open_acceptance_issue":
      case "collect_evidence":
      case "manual_checkpoint":
      case "resolve_verification_conflict":
      case "complete_project":
        navigate(buildProjectRoute(targetProjectId, "acceptance", { task: targetTaskId }));
        return;
      case "resolve_runtime_escalation":
      case "review_fault_loop":
        navigate(buildProjectRoute(targetProjectId, "diagnostics", { task: targetTaskId }));
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
      title={t("workspace.title")}
      onRetry={() => setRefreshTick((value) => value + 1)}
      secondaryActionLabel={t("workspace.openDiagnostics")}
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage={t("workspace.recovery")}
    >
      {state.data ? (
        showOnboarding ? (
          <OnboardingPanel
            projectName={firstText(state.data.workspace_explanation.headline, projectSnapshot?.name, projectId)}
            projectId={projectId}
            currentStage={projectSnapshot?.current_stage ?? ""}
            progressPercent={projectSnapshot?.progress_percent ?? 0}
            onDismiss={() => setOnboardingDismissed(true)}
            t={t}
            navigate={navigate}
          />
        ) : (
          <section className="dashboard-page">
          {/* Top Status Bar */}
          <div className="workspace-top-bar">
            <div style={{ minWidth: 0 }}>
              <p className="placeholder-section">{t("workspace.section")}</p>
              <h3 className="placeholder-title" style={{ fontSize: "20px", marginTop: "4px" }}>
                {firstText(state.data.workspace_explanation.headline, projectSnapshot?.name, projectId)}
              </h3>
            </div>
            <div className="action-row" style={{ flexWrap: "wrap", justifyContent: "flex-end" }}>
              <span className="status-pill">
                {firstText(workspaceOverview?.current_stage, projectSnapshot?.current_stage, t("unknown"))}
              </span>
              <span className="status-pill">
                {t("workspace.progress")} {projectSnapshot?.progress_percent ?? 0}%
              </span>
              <span className="status-pill">
                {t("workspace.risk")}{" "}
                {firstText(workspaceOverview?.risk_level, projectSnapshot?.risk_level, t("unknown"))}
              </span>
              {workspaceNextAction ? (
                <span className="status-pill">
                  {t("next")} {workspaceNextAction}
                </span>
              ) : null}
              <span className="status-pill">
                {t("stream")} {streamState}
              </span>
              {workspaceNextAction ? (
                <button
                  className="secondary-button"
                  onClick={() =>
                    handleAction(workspaceNextAction, {
                      targetId: runtimeEscalation?.task_id || runtimeEscalation?.source_task_id || projectId,
                    })
                  }
                >
                  {t("workspace.followNextAction")}
                </button>
              ) : null}
              <button className="secondary-button" onClick={() => setRefreshTick((value) => value + 1)}>
                {t("workspace.refreshView")}
              </button>
            </div>
          </div>

          {/* Three-column cockpit */}
          <div className="workspace-cockpit">
            {/* Left: Stage Rail */}
            <div className="workspace-column">
              <h4>{t("workspace.stageProgress")}</h4>
              <div className="stack-list">
                {state.data.stage_progress.map((item) => (
                  <article key={item.stage_key} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.stage_name}</strong>
                      <span className="status-pill">{item.status}</span>
                    </div>
                    <p>{item.active_item_title || t("noActiveItem")}</p>
                  </article>
                ))}
                {state.data.stage_progress.length === 0 ? (
                  <article className="list-card">
                    <p>{t("workspace.noStageProgress")}</p>
                  </article>
                ) : null}
              </div>
            </div>

            {/* Center: Live Activity */}
            <div className="workspace-column">
              <h4>{t("workspace.liveActivity")}</h4>
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
                {state.data.live_activity
                  .slice(0, expandedSections.activity ? undefined : 6)
                  .map((item) => (
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
                        <span
                          className={
                            item.requires_action ? "severity-badge severity-warning" : "status-pill"
                          }
                        >
                          {item.requires_action ? t("needsAction") : item.event_type}
                        </span>
                      </div>
                      <p>
                        {item.source_brain} · {item.occurred_at}
                        {item.source_task_id ? ` · ${t("task")} ${item.source_task_id}` : ""}
                      </p>
                    </button>
                  ))}
                {streamEvents.length === 0 && state.data.live_activity.length === 0 ? (
                  <article className="list-card">
                    <p>{t("workspace.noLiveActivity")}</p>
                  </article>
                ) : null}
                {state.data.live_activity.length > 6 && !expandedSections.activity ? (
                  <button className="show-more-btn" onClick={() => toggleSection("activity")}>
                    +{state.data.live_activity.length - 6} more
                  </button>
                ) : null}
              </div>
            </div>

            {/* Right: Action Inbox */}
            <div className="workspace-column">
              <h4>{t("workspace.actionInbox")}</h4>
              <div className="stack-list">
                {[...state.data.action_inbox]
                  .sort((a, b) => (a.is_blocking === b.is_blocking ? 0 : a.is_blocking ? -1 : 1))
                  .map((item) => (
                    <button
                      key={item.item_id}
                      className="action-card"
                      onClick={() => handleAction(item.recommended_action, { targetId: item.target_id })}
                    >
                      <div className="action-card-head">
                        <strong>{item.title}</strong>
                        <span className={`severity-badge severity-${item.severity}`}>
                          {formatInboxBadge(item)}
                        </span>
                      </div>
                      <p>
                        {describeInboxAction(item)} · {item.recommended_action}
                      </p>
                    </button>
                  ))}
                {state.data.action_inbox.length === 0 ? (
                  <article className="list-card">
                    <p>{t("workspace.noBlockingActions")}</p>
                  </article>
                ) : null}
              </div>
            </div>
          </div>

          {/* Bottom summary metrics */}
          <div className="metrics-grid" style={{ marginTop: "6px" }}>
            <MetricCard
              label={t("workspace.verification")}
              value={firstText(
                verificationResult?.decision,
                verificationResult?.status,
                workspaceOverview?.stage_status,
                projectSnapshot?.production_status,
                t("pending"),
              )}
              tone="neutral"
            />
            <MetricCard
              label={t("workspace.evidence")}
              value={`${state.data.acceptance_coverage.evidence_ready}/${state.data.acceptance_coverage.evidence_required}`}
              tone="calm"
            />
            <MetricCard
              label={t("workspace.completion")}
              value={firstText(
                completionVerdict?.decision,
                completionVerdict?.final_status,
                workspaceOverview?.production_status,
                projectSnapshot?.production_status,
                t("pending"),
              )}
              tone="neutral"
            />
            <MetricCard
              label={t("next")}
              value={workspaceNextAction || t("none")}
              tone={workspaceNextAction ? "warn" : "neutral"}
            />
          </div>

          {/* Workspace flags */}
          {workspaceFlags.length > 0 ? (
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("workspace.guards")}</h3>
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

          {/* Snapshot + Recommended Actions + Top Blockers + Explain Links */}
          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("workspace.snapshot")}</h3>
                <span className="status-pill">{countOutlineArtifacts(state.data)}</span>
              </div>
              <div className="stack-list">
                <SummaryCard
                  title={t("workspace.overview")}
                  primary={firstText(workspaceOverview?.stage_status, workspaceOverview?.production_status, projectSnapshot?.production_status, t("notReady"))}
                  secondary={firstText(
                    state.data.workspace_explanation.summary,
                    "Structured workspace overview becomes primary when the backend emits overview fields.",
                  )}
                  pills={[
                    workspaceOverview?.current_stage ? `${t("stage")} ${workspaceOverview.current_stage}` : undefined,
                    workspaceOverview?.risk_level ? `${t("risk")} ${workspaceOverview.risk_level}` : undefined,
                    workspaceOverview?.acceptance_run_status ? `${t("acceptance")} ${workspaceOverview.acceptance_run_status}` : undefined,
                    workspaceOverview?.next_action ? `${t("next")} ${workspaceOverview.next_action}` : undefined,
                    workspaceOverview ? `${t("actions")} ${workspaceOverview.action_required_count}` : undefined,
                    workspaceOverview?.manual_release_required !== undefined
                      ? `${t("manualRelease")} ${String(workspaceOverview.manual_release_required)}`
                      : undefined,
                  ]}
                  lines={[
                    workspaceOverview ? `${t("blockingIssues")}: ${workspaceOverview.blocking_issue_count}` : undefined,
                    formatBooleanLine(t("manualReviewRequired"), workspaceOverview?.manual_review_required),
                    formatBooleanLine(t("verificationConflict"), workspaceOverview?.verification_conflict),
                    formatBooleanLine(t("faultLoop"), workspaceOverview?.fault_loop_detected),
                    formatBooleanLine(t("policyDenied"), workspaceOverview?.policy_denied),
                  ]}
                />
                <SummaryCard
                  title={t("workspace.verificationResult")}
                  primary={firstText(verificationResult?.decision, verificationResult?.status, t("notReady"))}
                  secondary={
                    verificationResult?.summary ||
                    "Workspace is currently using production_status because no structured verification result is attached."
                  }
                  pills={[
                    verificationResult?.completed !== undefined
                      ? `${t("completed")} ${String(verificationResult.completed)}`
                      : undefined,
                    verificationResult?.preferred_verification_channel
                      ? `${t("channel")} ${verificationResult.preferred_verification_channel}`
                      : undefined,
                    verificationResult?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine(t("requiredChecks"), verificationResult?.required_checks),
                    formatChecklistLine(t("missingEvidence"), verificationResult?.missing_evidence),
                    formatChecklistLine(t("failedChecks"), verificationResult?.failed_checks),
                  ]}
                />
                <SummaryCard
                  title={t("workspace.completionVerdict")}
                  primary={firstText(completionVerdict?.decision, completionVerdict?.final_status, "not emitted")}
                  secondary={firstText(completionVerdict?.summary, completionVerdict?.reason, "Workspace explanation is currently using the latest available summary.")}
                  pills={[
                    completionVerdict?.completed !== undefined
                      ? `${t("completed")} ${String(completionVerdict.completed)}`
                      : undefined,
                    completionVerdict?.release_ready !== undefined
                      ? `${t("releaseReady")} ${String(completionVerdict.release_ready)}`
                      : undefined,
                    completionVerdict?.manual_review_required !== undefined
                      ? `${t("manualReviewRequired")} ${String(completionVerdict.manual_review_required)}`
                      : undefined,
                    completionVerdict?.manual_release_required !== undefined
                      ? `${t("manualRelease")} ${String(completionVerdict.manual_release_required)}`
                      : undefined,
                    completionVerdict?.manual_release_completed !== undefined
                      ? `${t("releaseReady")} ${String(completionVerdict.manual_release_completed)}`
                      : undefined,
                    completionVerdict?.blocker_count !== undefined
                      ? `${t("blockers")} ${completionVerdict.blocker_count}`
                      : undefined,
                    completionVerdict?.next_action ? `${t("next")} ${completionVerdict.next_action}` : undefined,
                    completionVerdict?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title={t("workspace.runtimeEscalation")}
                  primary={firstText(runtimeEscalation?.reason_class, runtimeEscalation?.status, t("none"))}
                  secondary={firstText(runtimeEscalation?.summary, "No structured runtime escalation is attached to this workspace snapshot.")}
                  pills={[
                    runtimeEscalation?.severity || undefined,
                    runtimeEscalation?.source_brain ? `${t("brain")} ${runtimeEscalation.source_brain}` : undefined,
                    runtimeEscalation?.source_task_id ? `${t("sourceTask")} ${runtimeEscalation.source_task_id}` : undefined,
                    runtimeEscalation?.run_binding_id ? `${t("binding")} ${runtimeEscalation.run_binding_id}` : undefined,
                    runtimeEscalation?.run_status ? `${t("runStatus")} ${runtimeEscalation.run_status}` : undefined,
                    runtimeEscalation?.action ? `${t("actions")} ${runtimeEscalation.action}` : undefined,
                    runtimeEscalation?.task_id ? `${t("task")} ${runtimeEscalation.task_id}` : undefined,
                    runtimeEscalation?.run_id ? `${t("run")} ${runtimeEscalation.run_id}` : undefined,
                    runtimeEscalation?.policy_denied !== undefined ? `${t("policyDenied")} ${String(runtimeEscalation.policy_denied)}` : undefined,
                    runtimeEscalation?.updated_at || undefined,
                  ]}
                />
                <SummaryCard
                  title={t("workspace.faultSummary")}
                  primary={firstText(faultSummary?.fault_kind, faultSummary?.status, t("none"))}
                  secondary={firstText(faultSummary?.summary, faultSummary?.top_issue, "No aggregated fault summary is attached to the current workspace view.")}
                  pills={[
                    faultSummary?.severity || undefined,
                    faultSummary?.fault_loop_detected !== undefined ? `${t("faultLoop")} ${String(faultSummary.fault_loop_detected)}` : undefined,
                    faultSummary?.blocking_issue_count !== undefined ? `${t("blockingIssues")} ${faultSummary.blocking_issue_count}` : undefined,
                    faultSummary?.advisory_issue_count !== undefined ? `${t("advisory")} ${faultSummary.advisory_issue_count}` : undefined,
                    faultSummary?.updated_at || undefined,
                  ]}
                  lines={[
                    formatChecklistLine(t("failedChecks"), faultSummary?.failed_checks),
                    formatChecklistLine(t("affectedTasks"), faultSummary?.affected_tasks),
                  ]}
                />
                <SummaryCard
                  title={t("workspace.repairPlan")}
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
                      ? `${t("manualReviewRequired")} ${String(repairPlanDraft.manual_review_required)}`
                      : undefined,
                    repairPlanDraft?.updated_at || undefined,
                  ]}
                  lines={[formatChecklistLine(t("updatedTasks"), repairPlanDraft?.updated_tasks)]}
                />
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("workspace.recommendedActions")}</h3>
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
                    <p>{t("workspace.noRecommendedActions")}</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("workspace.topBlockers")}</h3>
              </div>
              <div className="stack-list">
                {state.data.workspace_explanation.top_blockers.map((item) => (
                  <article key={item} className="list-card">
                    <p>{item}</p>
                  </article>
                ))}
                {state.data.workspace_explanation.top_blockers.length === 0 ? (
                  <article className="list-card">
                    <p>{t("workspace.noBlockers")}</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>{t("workspace.explainLinks")}</h3>
                <span className="status-pill">{state.data.workspace_explanation.explain_links?.length ?? 0}</span>
              </div>
              <div className="stack-list">
                {(state.data.workspace_explanation.explain_links ?? []).map((item) => (
                  <button key={item} className="action-card" onClick={() => handleExplainLink(item, projectId, navigate)}>
                    <div className="action-card-head">
                      <strong>{formatExplainLinkLabel(item, t)}</strong>
                      <span className="status-pill">{item}</span>
                    </div>
                    <p>{describeExplainLink(item, t)}</p>
                  </button>
                ))}
                {(state.data.workspace_explanation.explain_links?.length ?? 0) === 0 ? (
                  <article className="list-card">
                    <p>{t("workspace.noExplainLinks")}</p>
                  </article>
                ) : null}
              </div>
            </section>
          </div>

          {/* Multi-Project Overview (collapsible) */}
          {homeState.data ? (
            <section className="data-panel">
              <div
                className="panel-header"
                style={{ cursor: "pointer" }}
                onClick={() => toggleSection("homeOverview")}
              >
                <h3>{"Multi-Project Overview"}</h3>
                <span className="status-pill">{expandedSections.homeOverview ? "▾" : "▸"}</span>
              </div>
              {expandedSections.homeOverview ? (
                <>
                  <div className="metrics-grid" style={{ marginBottom: "14px" }}>
                    <MetricCard label={t("workspace.totalProjects")} value={String(homeState.data.summary.total_projects)} tone="calm" />
                    <MetricCard label={t("workspace.activeProjects")} value={String(homeState.data.summary.active_projects)} tone="calm" />
                    <MetricCard label={t("workspace.blockedProjects")} value={String(homeState.data.summary.blocked_projects)} tone={homeState.data.summary.blocked_projects > 0 ? "warn" : "neutral"} />
                    <MetricCard label={t("workspace.pendingActions")} value={String(homeState.data.summary.pending_actions)} tone={homeState.data.summary.pending_actions > 0 ? "warn" : "neutral"} />
                  </div>

                  <section className="data-panel" style={{ marginBottom: "14px" }}>
                    <div className="panel-header">
                      <h3>{t("workspace.activeProjects")}</h3>
                      <div className="action-row">
                        <span className="status-pill">{homeState.data.active_projects.length}</span>
                        <button className="secondary-button" onClick={() => navigate(routes.settings)}>+ {t("settings.createProject")}</button>
                      </div>
                    </div>
                    <div className="stack-list">
                      {homeState.data.active_projects.map((item) => (
                        <article key={item.project_id} className={`list-card${item.project_id === projectId ? " is-selected" : ""}`}>
                          <div className="list-card-head">
                            <strong style={{ cursor: "pointer" }} onClick={() => updateProjectId(item.project_id)}>{item.name}</strong>
                            <div className="action-row">
                              <span className="status-pill">{item.current_stage}</span>
                              <span className={`status-pill${item.production_status === "pending" ? " pill-advisory" : item.production_status === "ready" ? " pill-success" : ""}`}>{item.production_status}</span>
                            </div>
                          </div>
                          <p>{item.project_category} · {t("workspace.progress")} {item.progress_percent}%</p>
                          <div className="action-row">
                            <button className="secondary-button" onClick={() => { updateProjectId(item.project_id); navigate(routes.plan); }}>{t("nav.plan")}</button>
                            <button className="secondary-button" onClick={() => { updateProjectId(item.project_id); navigate(routes.execution); }}>{t("nav.execution")}</button>
                            <button className="secondary-button" onClick={() => { updateProjectId(item.project_id); navigate(routes.acceptance); }}>{t("nav.acceptance")}</button>
                          </div>
                        </article>
                      ))}
                      {homeState.data.active_projects.length === 0 ? (
                        <article className="list-card"><p>{t("workspace.noActiveProjects")}</p></article>
                      ) : null}
                    </div>
                  </section>

                  {homeState.data.need_attention.length > 0 ? (
                    <section className="data-panel" style={{ marginBottom: "14px" }}>
                      <div className="panel-header">
                        <h3>{t("workspace.needAttention")}</h3>
                        <span className="status-pill pill-advisory">{homeState.data.need_attention.length}</span>
                      </div>
                      <div className="stack-list">
                        {homeState.data.need_attention.map((item) => (
                          <button key={item.item_id} className="action-card" onClick={() => { updateProjectId(item.project_id); handleAction(item.recommended_action, {}); }}>
                            <div className="action-card-head">
                              <strong>{item.title}</strong>
                              <span className={`severity-badge severity-${item.severity}`}>{item.is_blocking ? t("blocking") : item.severity}</span>
                            </div>
                            <p>{item.project_id}</p>
                          </button>
                        ))}
                      </div>
                    </section>
                  ) : null}

                  <div className="content-grid">
                    <section className="data-panel">
                      <div className="panel-header">
                        <h3>{t("workspace.releaseReadiness")}</h3>
                      </div>
                      <div className="stack-list">
                        {homeState.data.release_readiness.map((item) => (
                          <article key={item.project_id} className="list-card">
                            <div className="list-card-head">
                              <strong>{item.name}</strong>
                              <span className={`status-pill${item.production_status === "ready" ? " pill-success" : ""}`}>{item.production_status}</span>
                            </div>
                            {item.missing_items > 0 ? <p>{item.missing_items} {t("workspace.missingItems")}</p> : null}
                          </article>
                        ))}
                        {homeState.data.release_readiness.length === 0 ? (
                          <article className="list-card"><p>{t("workspace.noReleaseItems")}</p></article>
                        ) : null}
                      </div>
                    </section>

                    <section className="data-panel">
                      <div className="panel-header">
                        <h3>{t("workspace.recentActivity")}</h3>
                      </div>
                      <div className="stack-list">
                        {homeState.data.recent_activity.slice(0, expandedSections.recentActivity ? undefined : 5).map((item) => (
                          <article key={item.event_id} className="list-card">
                            <div className="list-card-head">
                              <strong>{item.title}</strong>
                              <span className={item.needs_attention ? "severity-badge severity-warning" : "status-pill"}>{item.event_type}</span>
                            </div>
                            <p>{item.source_brain} · {item.occurred_at}</p>
                          </article>
                        ))}
                        {homeState.data.recent_activity.length === 0 ? (
                          <article className="list-card"><p>{t("workspace.noRecentActivity")}</p></article>
                        ) : null}
                        {homeState.data.recent_activity.length > 5 && !expandedSections.recentActivity ? (
                          <button className="show-more-btn" onClick={() => toggleSection("recentActivity")}>+{homeState.data.recent_activity.length - 5} more</button>
                        ) : null}
                      </div>
                    </section>
                  </div>
                </>
              ) : null}
            </section>
          ) : null}
        </section>
      )
      ) : null}
    </QueryPanel>
  );
}

function OnboardingPanel(props: {
  projectName: string;
  projectId: string;
  currentStage: string;
  progressPercent: number;
  onDismiss: () => void;
  t: (key: string, options?: Record<string, unknown>) => string;
  navigate: ReturnType<typeof useNavigate>;
}) {
  const handleDismiss = () => {
    window.localStorage.setItem(`easymvp.project.onboarded.${props.projectId}`, "true");
    props.onDismiss();
  };

  const handleSkip = () => {
    window.localStorage.setItem(`easymvp.project.onboarded.${props.projectId}`, "true");
    props.onDismiss();
  };

  const planCompiled = props.currentStage !== "created" && props.progressPercent > 0;

  return (
    <section className="onboarding-panel">
      <div className="onboarding-header">
        <h2 className="placeholder-title">{props.t("workspace.onboardingWelcome", { name: props.projectName })}</h2>
        <p className="muted-copy">{props.t("workspace.onboardingSubtitle")}</p>
      </div>

      <div className="onboarding-cards">
        <article className="data-panel onboarding-card">
          <h4>{props.t("workspace.onboardingStep1Title")}</h4>
          <p>{props.t("workspace.onboardingStep1Desc")}</p>
        </article>
        <article className="data-panel onboarding-card">
          <h4>{props.t("workspace.onboardingStep2Title")}</h4>
          <p>{props.t("workspace.onboardingStep2Desc")}</p>
        </article>
        <article className="data-panel onboarding-card">
          <h4>{props.t("workspace.onboardingStep3Title")}</h4>
          <p>{props.t("workspace.onboardingStep3Desc")}</p>
        </article>
      </div>

      <div className="action-row onboarding-actions">
        <button className="primary-button" onClick={handleDismiss}>
          {props.t("workspace.onboardingGetStarted")}
        </button>
        <button className="secondary-button" onClick={handleSkip}>
          {props.t("workspace.onboardingSkip")}
        </button>
        {!planCompiled ? (
          <button
            className="secondary-button"
            onClick={() => props.navigate(buildProjectRoute(props.projectId, "plan"))}
          >
            {props.t("nav.plan")}
          </button>
        ) : (
          <button
            className="secondary-button"
            onClick={() => props.navigate(buildProjectRoute(props.projectId, "execution"))}
          >
            {props.t("nav.execution")}
          </button>
        )}
      </div>
    </section>
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
          {pills.map((item, idx) => (
            <span key={idx} className="status-pill">
              {item}
            </span>
          ))}
        </div>
      ) : null}
      {lines.map((line, idx) => (
        <p key={idx} className="muted-copy">
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
  section:
    | "workspace"
    | "plan"
    | "execution"
    | "acceptance"
    | "diagnostics"
    | "repair-draft",
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

function formatExplainLinkLabel(linkKey: string, t: (key: string) => string) {
  switch (linkKey) {
    case "runtime":
      return t("explain.runtime");
    case "task_review":
      return t("explain.taskReview");
    case "diagnostics":
      return t("explain.diagnostics");
    case "replay":
      return t("explain.replay");
    case "audit":
      return t("explain.audit");
    case "acceptance":
      return t("explain.acceptance");
    case "plan":
      return t("explain.plan");
    default:
      return t("explain.default");
  }
}

function describeExplainLink(linkKey: string, t: (key: string) => string) {
  switch (linkKey) {
    case "runtime":
      return t("describe.runtime");
    case "task_review":
      return t("describe.taskReview");
    case "diagnostics":
      return t("describe.diagnostics");
    case "replay":
      return t("describe.replay");
    case "audit":
      return t("describe.audit");
    case "acceptance":
      return t("describe.acceptance");
    case "plan":
      return t("describe.plan");
    default:
      return t("describe.default");
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
