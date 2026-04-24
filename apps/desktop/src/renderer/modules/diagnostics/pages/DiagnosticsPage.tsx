import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet } from "@/shared/lib/api";
import {
  exportDesktopDiagnostics,
  getDesktopRuntimeDiagnosis,
  getDesktopRuntimeInfo,
  getDesktopRuntimeRecoveryHint,
  relaunchDesktopNormalMode,
  relaunchDesktopSafeMode,
  restartManagedCore,
  showDesktopItemInFolder,
  startManagedCore,
} from "@/shared/lib/preferences";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import type {
  ProjectDiagnosticsView,
  RuntimeHealthView,
  SystemHealthView,
} from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function DiagnosticsPage() {
  const navigate = useNavigate();
  const { projectId, routes, buildRoute } = useProjectState();
  const [exportBusy, setExportBusy] = useState(false);
  const [exportError, setExportError] = useState("");
  const [exportMessage, setExportMessage] = useState("");
  const [actionBusy, setActionBusy] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const systemState = useQuery(
    () => apiGet<SystemHealthView>("/api/v3/system/healthz"),
    [],
  );
  const runtimeState = useQuery(
    () => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"),
    [],
  );
  const desktopRuntimeState = useQuery(() => getDesktopRuntimeInfo(), []);
  const desktopDiagnosis = desktopRuntimeState.data
    ? getDesktopRuntimeDiagnosis(desktopRuntimeState.data)
    : null;
  const diagnosticsState = useQuery(
    () =>
      apiGet<ProjectDiagnosticsView>(
        `/api/v3/projects/${encodeURIComponent(projectId)}/diagnostic-records?limit=12`,
      ),
    [projectId],
  );
  const primaryLinkedRun = diagnosticsState.data?.linked_runs?.[0];
  const primaryTaskID =
    primaryLinkedRun?.task_id ||
    diagnosticsState.data?.items.find((item) => item.task_id)?.task_id ||
    "";
  const primaryBindingID =
    primaryLinkedRun?.binding_id ||
    diagnosticsState.data?.items.find((item) => item.binding_id)?.binding_id ||
    "";
  const primaryRunID =
    primaryLinkedRun?.run_id ||
    diagnosticsState.data?.items.find((item) => item.run_id)?.run_id ||
    "";

  async function handleExport() {
    if (
      !systemState.data ||
      !runtimeState.data ||
      !diagnosticsState.data ||
      !desktopRuntimeState.data
    ) {
      setExportError("Diagnostics are not ready to export yet");
      setExportMessage("");
      return;
    }

    setExportBusy(true);
    setExportError("");
    setExportMessage("");
    try {
      const result = await exportDesktopDiagnostics(
        {
          exported_at: new Date().toISOString(),
          project_id: projectId,
          page: "diagnostics",
          system_health: systemState.data,
          runtime_health: runtimeState.data,
          desktop_runtime: desktopRuntimeState.data,
          desktop_diagnosis: desktopDiagnosis,
          project_diagnostics: diagnosticsState.data,
        },
        `easymvp-diagnostics-${projectId}-${Date.now()}.json`,
      );
      if (!result.ok) {
        if (result.canceled) {
          setExportMessage(result.error || "Diagnostic export canceled");
        } else {
          setExportError(result.error || "Diagnostic export failed");
        }
        return;
      }
      setExportMessage(
        result.path
          ? `Diagnostics exported to ${result.path}`
          : "Diagnostics export completed",
      );
    } catch (error) {
      setExportError(
        error instanceof Error ? error.message : "Diagnostic export failed",
      );
    } finally {
      setExportBusy(false);
    }
  }

  async function runAction(
    actionKey: string,
    action: () => Promise<{ ok: boolean; error?: string; path?: string; canceled?: boolean }>,
  ) {
    setActionBusy(actionKey);
    setActionError("");
    setActionMessage("");
    try {
      const result = await action();
      if (!result.ok) {
        if (result.canceled) {
          setActionMessage(result.error || "Desktop action canceled");
        } else {
          setActionError(result.error || "Desktop action failed");
        }
      } else {
        setActionMessage(result.path ? `Saved to ${result.path}` : "Action completed");
      }
    } catch (error) {
      setActionError(error instanceof Error ? error.message : "Desktop action failed");
    } finally {
      setActionBusy("");
    }
  }

  function runStructuredAction(
    actionID: string,
    options?: { dataDirectory?: string },
  ) {
    switch (actionID) {
      case "retry-health-probe":
        window.location.reload();
        return Promise.resolve();
      case "start-managed-core":
        return runAction(actionID, () => startManagedCore());
      case "restart-managed-core":
        return runAction(actionID, () => restartManagedCore());
      case "relaunch-safe-mode":
        return runAction(actionID, () => relaunchDesktopSafeMode());
      case "relaunch-normal-mode":
        return runAction(actionID, () => relaunchDesktopNormalMode());
      case "open-settings":
        navigate(routes.settings);
        return Promise.resolve();
      case "open-diagnostics":
        navigate(routes.diagnostics);
        return Promise.resolve();
      case "open-data-folder":
        return runAction(actionID, () =>
          showDesktopItemInFolder(options?.dataDirectory || ""),
        );
      default:
        setActionError(`Unsupported recovery action: ${actionID}`);
        return Promise.resolve();
    }
  }

  return (
    <QueryPanel
      loading={
        systemState.loading || runtimeState.loading || diagnosticsState.loading
      }
      error={systemState.error || runtimeState.error || diagnosticsState.error}
      title="Diagnostics"
      onRetry={() => window.location.reload()}
      secondaryActionLabel="Open Settings"
      onSecondaryAction={() => window.location.assign(routes.settings)}
      recoveryMessage="Diagnostics aggregates system health, runtime state, recovery signals, and exportable evidence from one entry point."
    >
      {systemState.data && runtimeState.data && diagnosticsState.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Diagnostics</p>
              <h3 className="placeholder-title">
                Recovery and health entry point
              </h3>
              <p className="placeholder-description">
                Use this page when a workbench page is stale, a runtime action
                fails, or local core connectivity needs verification.
              </p>
            </div>
            <div className="summary-stack">
              <button
                className="secondary-button"
                disabled={exportBusy}
                onClick={() => void handleExport()}
              >
                {exportBusy ? "Exporting..." : "Export Diagnostics"}
              </button>
              <span className="status-pill">
                system {systemState.data.status}
              </span>
              <span className="status-pill">
                runtime {runtimeState.data.status}
              </span>
              <span className="status-pill">
                {diagnosticsState.data.items.length} diagnostics
              </span>
              {Object.entries(diagnosticsState.data.category_counts || {}).slice(0, 3).map(([category, count]) => (
                <span key={category} className="status-pill">
                  {category} {count}
                </span>
              ))}
            </div>
          </div>
          {exportError ? <p className="error-copy">{exportError}</p> : null}
          {exportMessage ? <p className="muted-copy">{exportMessage}</p> : null}
          {actionError ? <p className="error-copy">{actionError}</p> : null}
          {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>System Health</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>{systemState.data.service}</strong>
                    <span className="status-pill">
                      {systemState.data.status}
                    </span>
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
                    <span className="status-pill">
                      {runtimeState.data.status}
                    </span>
                  </div>
                  <p>{runtimeState.data.base_url}</p>
                </article>
              </div>
            </section>

            {desktopRuntimeState.data ? (
              <section className="data-panel">
                <div className="panel-header">
                  <h3>Desktop Runtime</h3>
                  <span className="status-pill">
                    {desktopRuntimeState.data.source}
                  </span>
                </div>
                <div className="runtime-grid">
                  <div className="runtime-field">
                    <span>Launch mode</span>
                    <strong>{desktopRuntimeState.data.launchMode}</strong>
                  </div>
                  <div className="runtime-field">
                    <span>Packaged</span>
                    <strong>
                      {desktopRuntimeState.data.packaged ? "yes" : "no"}
                    </strong>
                  </div>
                  <div className="runtime-field">
                    <span>Platform</span>
                    <strong>{desktopRuntimeState.data.platform}</strong>
                  </div>
                  <div className="runtime-field">
                    <span>Version</span>
                    <strong>{desktopRuntimeState.data.version}</strong>
                  </div>
                  <div className="runtime-field runtime-field-wide">
                    <span>Core base URL</span>
                    <strong>{desktopRuntimeState.data.coreBaseUrl}</strong>
                  </div>
                  <div className="runtime-field">
                    <span>Core probe</span>
                    <strong>{desktopRuntimeState.data.coreStatus}</strong>
                  </div>
                  <div className="runtime-field">
                    <span>HTTP status</span>
                    <strong>
                      {desktopRuntimeState.data.coreHttpStatus || "n/a"}
                    </strong>
                  </div>
                  <div className="runtime-field">
                    <span>Core manager</span>
                    <strong>
                      {desktopRuntimeState.data.coreManagerStatus}
                    </strong>
                  </div>
                  <div className="runtime-field">
                    <span>Diagnosis</span>
                    <strong>{desktopDiagnosis?.status || "collecting"}</strong>
                  </div>
                </div>
                <div className="list-card runtime-recovery">
                  <div className="list-card-head">
                    <strong>Recovery hint</strong>
                    <span className="status-pill">
                      {desktopRuntimeState.data.issue
                        ? "check bridge"
                        : "runtime wiring"}
                    </span>
                  </div>
                  {desktopRuntimeState.data.coreError ? (
                    <p className="error-copy">
                      {desktopRuntimeState.data.coreError}
                    </p>
                  ) : null}
                  <p>
                    {desktopRuntimeState.data.issue ||
                      getDesktopRuntimeRecoveryHint(desktopRuntimeState.data)}
                  </p>
                  <p>
                    Health endpoint reports {runtimeState.data.base_url}.
                    Compare it with the renderer-resolved desktop runtime URL
                    before retrying a failing action.
                  </p>
                  {desktopRuntimeState.data.coreManagerCommand ? (
                    <p>
                      Managed command:{" "}
                      {desktopRuntimeState.data.coreManagerCommand}{" "}
                      {desktopRuntimeState.data.coreManagerArgs.join(" ")}
                    </p>
                  ) : null}
                </div>
                <div className="runtime-note">
                  <div className="runtime-note-header">
                    <strong>Structured startup diagnosis</strong>
                    <span className="status-pill">
                      {desktopDiagnosis?.issues.length || 0} issues
                    </span>
                  </div>
                  {desktopDiagnosis?.issues.length ? (
                    <div className="stack-list">
                      {desktopDiagnosis.issues.map((issue) => (
                        <article
                          key={`${issue.code}-${issue.summary}`}
                          className="list-card recovery-issue-card"
                        >
                          <div className="list-card-head">
                            <strong>{issue.summary}</strong>
                            <span
                              className={`severity-badge severity-${issue.severity}`}
                            >
                              {issue.severity}
                            </span>
                          </div>
                          <p>
                            {issue.code} · {issue.category} · {issue.source} · {issue.mode}
                          </p>
                          <p>{issue.detail}</p>
                          {issue.actions.length ? (
                            <div className="recovery-action-list">
                              {issue.actions.map((action) => (
                                <button
                                  key={`${issue.code}-${action.id}`}
                                  className="recovery-action-card"
                                  disabled={actionBusy !== ""}
                                  onClick={() =>
                                    void runStructuredAction(action.id, {
                                      dataDirectory: desktopRuntimeState.data?.dataDirectory || "",
                                    })
                                  }
                                >
                                  <strong>{action.label}</strong>
                                  <p>{action.description}</p>
                                </button>
                              ))}
                            </div>
                          ) : null}
                        </article>
                      ))}
                    </div>
                  ) : (
                    <article className="list-card recovery-issue-card">
                      <div className="list-card-head">
                        <strong>No structured startup issue available</strong>
                        <span className="severity-badge severity-info">
                          info
                        </span>
                      </div>
                      <p>
                        Renderer startup issue rendering is in place. The
                        current desktop snapshot did not report a structured
                        startup issue.
                      </p>
                    </article>
                  )}
                </div>
              </section>
            ) : null}

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
                  <p>
                    Return to the project cockpit after checking system and
                    runtime health.
                  </p>
                </a>
                <a
                  className="action-card"
                  href={buildRoute("/execution", {
                    binding: primaryBindingID || undefined,
                    run: primaryRunID || undefined,
                    task: primaryTaskID || undefined,
                  })}
                >
                  <div className="action-card-head">
                    <strong>Inspect Execution</strong>
                    <span className="status-pill">execution</span>
                  </div>
                  <p>
                    Open the runtime run board and verify the latest binding,
                    replay, and raw logs.
                  </p>
                </a>
                <a
                  className="action-card"
                  href={buildRoute("/replay", {
                    binding: primaryBindingID || undefined,
                    run: primaryRunID || undefined,
                    replay: primaryLinkedRun?.latest_replay_id,
                    task: primaryTaskID || undefined,
                  })}
                >
                  <div className="action-card-head">
                    <strong>Open Replay</strong>
                    <span className="status-pill">replay</span>
                  </div>
                  <p>
                    Jump directly into replay timeline inspection when the
                    runtime is healthy but artifact indexing needs review.
                  </p>
                </a>
                <a
                  className="action-card"
                  href={buildRoute("/audit", {
                    run: primaryRunID || undefined,
                    task: primaryTaskID || undefined,
                  })}
                >
                  <div className="action-card-head">
                    <strong>Open Audit</strong>
                    <span className="status-pill">audit</span>
                  </div>
                  <p>
                    Review persisted audit facts and release actions before
                    returning to execution or workspace triage.
                  </p>
                </a>
                <a className="action-card" href={routes.settings}>
                  <div className="action-card-head">
                    <strong>Edit Settings</strong>
                    <span className="status-pill">settings</span>
                  </div>
                  <p>
                    Adjust the current project id or local core base URL if the
                    shell points to the wrong instance.
                  </p>
                </a>
                <a className="action-card" href={routes.recovery}>
                  <div className="action-card-head">
                    <strong>Open Recovery</strong>
                    <span className="status-pill">recovery</span>
                  </div>
                  <p>
                    Switch the desktop shell into safe mode, retry the startup
                    probe, or inspect the local data directory.
                  </p>
                </a>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Category Counts</h3>
                <span className="status-pill">
                  {Object.keys(diagnosticsState.data.category_counts || {}).length} categories
                </span>
              </div>
              <div className="stack-list">
                {Object.entries(diagnosticsState.data.category_counts || {}).map(([category, count]) => (
                  <article key={category} className="list-card">
                    <div className="list-card-head">
                      <strong>{category}</strong>
                      <span className="status-pill">{count}</span>
                    </div>
                  </article>
                ))}
                {Object.keys(diagnosticsState.data.category_counts || {}).length === 0 ? (
                  <article className="list-card">
                    <p>No diagnostic categories have been counted yet.</p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Recent Diagnostics</h3>
                <span className="status-pill">
                  {diagnosticsState.data.refresh_hint}
                </span>
              </div>
              <div className="stack-list">
                {diagnosticsState.data.items.map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span
                        className={`severity-badge severity-${item.severity}`}
                      >
                        {item.severity}
                      </span>
                    </div>
                    <p>
                      {item.scope} · {item.error_code} · {item.created_at}
                    </p>
                    {item.category || item.component || item.field ? (
                      <p>
                        {item.category ? `category ${item.category}` : ""}
                        {item.component ? ` · component ${item.component}` : ""}
                        {item.field ? ` · field ${item.field}` : ""}
                      </p>
                    ) : null}
                    {item.project_id ||
                    item.task_id ||
                    item.run_id ||
                    item.binding_id ? (
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
                      <a
                        className="secondary-button"
                        href={buildRoute("/execution", {
                          task: item.task_id,
                          binding: item.binding_id,
                          run: item.run_id,
                        })}
                      >
                        Open Execution
                      </a>
                      <a
                        className="secondary-button"
                        href={buildRoute("/audit", {
                          task: item.task_id,
                          run: item.run_id,
                        })}
                      >
                        Open Audit
                      </a>
                      {item.related_page === "acceptance" ? (
                        <a
                          className="secondary-button"
                          href={buildRoute("/acceptance", { task: item.task_id })}
                        >
                          Open Acceptance
                        </a>
                      ) : null}
                      <a
                        className="secondary-button"
                        href={buildRoute("/replay", {
                          task: item.task_id,
                          binding: item.binding_id,
                          run: item.run_id,
                      })}
                    >
                        Open Replay
                      </a>
                    </div>
                    {item.recommended_action ? (
                      <p className="muted-copy">
                        Recommended: {item.recommended_action}
                      </p>
                    ) : null}
                    {item.detail_json ? (
                      <pre className="json-block">
                        {prettyJson(item.detail_json)}
                      </pre>
                    ) : null}
                  </article>
                ))}
                {diagnosticsState.data.items.length === 0 ? (
                  <article className="list-card">
                    <p>
                      No diagnostic records have been captured for this project
                      yet.
                    </p>
                  </article>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Verification Read Side</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>
                      {diagnosticsState.data.verification_read?.decision ||
                        diagnosticsState.data.verification_read?.completion_decision ||
                        "pending"}
                    </strong>
                    <span className="status-pill">
                      {diagnosticsState.data.verification_read?.fault_loop_detected
                        ? "fault loop"
                        : diagnosticsState.data.verification_read?.repair_draft_status || "steady"}
                    </span>
                  </div>
                  <p>
                    verification{" "}
                    {diagnosticsState.data.verification_read?.status || "n/a"} ·
                    completion{" "}
                    {diagnosticsState.data.verification_read?.completion_status || "n/a"}
                  </p>
                  {diagnosticsState.data.verification_read?.fault_summary ? (
                    <p>{diagnosticsState.data.verification_read.fault_summary}</p>
                  ) : null}
                  {diagnosticsState.data.verification_read?.missing_evidence?.length ? (
                    <p>
                      missing evidence:{" "}
                      {diagnosticsState.data.verification_read.missing_evidence.join(", ")}
                    </p>
                  ) : null}
                  {diagnosticsState.data.verification_read?.failed_checks?.length ? (
                    <p>
                      failed checks:{" "}
                      {diagnosticsState.data.verification_read.failed_checks.join(", ")}
                    </p>
                  ) : null}
                  <div className="action-row">
                    <a className="secondary-button" href={routes.acceptance}>
                      Open Acceptance
                    </a>
                    <a
                      className="secondary-button"
                      href={buildRoute("/repair-draft", { task: primaryTaskID || undefined })}
                    >
                      Open Repair
                    </a>
                  </div>
                  {diagnosticsState.data.verification_read?.verification_contract_json ? (
                    <pre className="json-block">
                      {prettyJson(
                        diagnosticsState.data.verification_read.verification_contract_json,
                      )}
                    </pre>
                  ) : null}
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Replay And Evidence</h3>
              </div>
              <div className="stack-list">
                {(diagnosticsState.data.linked_runs || []).map((item) => (
                  <article key={item.run_id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.run_id}</strong>
                      <span className="status-pill">{item.run_status || "unknown"}</span>
                    </div>
                    <p>
                      replay {item.replay_count} · logs {item.log_segment_count} · ready{" "}
                      {item.artifact_ready} · missing {item.artifact_missing}
                    </p>
                    {item.latest_replay_title || item.latest_replay_type ? (
                      <p>
                        {item.latest_replay_type || "replay"} ·{" "}
                        {item.latest_replay_title || item.latest_replay_id}
                      </p>
                    ) : null}
                    <div className="action-row">
                      <a
                        className="secondary-button"
                        href={buildRoute("/execution", {
                          binding: item.binding_id,
                          run: item.run_id,
                          task: item.task_id,
                        })}
                      >
                        Open Execution
                      </a>
                      <a
                        className="secondary-button"
                        href={buildRoute("/replay", {
                          binding: item.binding_id,
                          run: item.run_id,
                          replay: item.latest_replay_id,
                          task: item.task_id,
                        })}
                      >
                      Open Replay
                    </a>
                      <a
                        className="secondary-button"
                        href={buildRoute("/audit", {
                          run: item.run_id,
                          task: item.task_id,
                        })}
                      >
                        Open Audit
                      </a>
                    </div>
                  </article>
                ))}
                {(diagnosticsState.data.linked_runs || []).length === 0 ? (
                  <article className="list-card">
                    <p>No linked runtime runs are attached to the current diagnostics yet.</p>
                  </article>
                ) : null}
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>Evidence overview</strong>
                    <span className="status-pill">
                      {diagnosticsState.data.evidence_overview?.total_count || 0} evidence
                    </span>
                  </div>
                  {diagnosticsState.data.evidence_overview?.missing_required?.length ? (
                    <p>
                      missing required:{" "}
                      {diagnosticsState.data.evidence_overview.missing_required.join(", ")}
                    </p>
                  ) : null}
                  {diagnosticsState.data.evidence_overview?.failed_checks?.length ? (
                    <p>
                      failed checks:{" "}
                      {diagnosticsState.data.evidence_overview.failed_checks.join(", ")}
                    </p>
                  ) : null}
                  {(diagnosticsState.data.evidence_overview?.latest_evidence || []).map((item) => (
                    <p key={item.id}>
                      {item.surface} · {item.evidence_type} · {item.captured_at}
                    </p>
                  ))}
                  <div className="action-row">
                    <a className="secondary-button" href={routes.acceptance}>
                      Open Acceptance
                    </a>
                    <a
                      className="secondary-button"
                      href={buildRoute("/execution", {
                        binding: primaryBindingID || undefined,
                        run: primaryRunID || undefined,
                        task: primaryTaskID || undefined,
                      })}
                    >
                      Collect Evidence
                    </a>
                  </div>
                </article>
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Latest Audit Facts</h3>
              </div>
              <div className="stack-list">
                {(diagnosticsState.data.latest_audit_logs || []).map((item) => (
                  <article key={item.id} className="list-card">
                    <div className="list-card-head">
                      <strong>{item.summary}</strong>
                      <span className="status-pill">{item.event_type}</span>
                    </div>
                    <p>
                      {item.actor_kind} · {item.created_at}
                    </p>
                  </article>
                ))}
                {(diagnosticsState.data.latest_audit_logs || []).length === 0 ? (
                  <article className="list-card">
                    <p>No recent audit facts were returned with diagnostics.</p>
                  </article>
                ) : null}
                <div className="action-row">
                  <a
                    className="secondary-button"
                    href={buildRoute("/audit", {
                      run: primaryRunID || undefined,
                      task: primaryTaskID || undefined,
                    })}
                  >
                    Open Audit
                  </a>
                </div>
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
