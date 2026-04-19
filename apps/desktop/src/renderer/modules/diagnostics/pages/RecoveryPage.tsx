import { useState } from "react";
import { useNavigate } from "react-router-dom";
import {
  exportDesktopDiagnostics,
  getDesktopRuntimeDiagnosis,
  getDesktopRuntimeInfo,
  getDesktopRuntimeRecoveryHint,
  openDesktopPath,
  relaunchDesktopNormalMode,
  relaunchDesktopSafeMode,
  restartManagedCore,
  startManagedCore,
  showDesktopItemInFolder,
} from "@/shared/lib/preferences";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";

export function RecoveryPage() {
  const navigate = useNavigate();
  const { routes } = useProjectState();
  const [refreshTick, setRefreshTick] = useState(0);
  const [actionBusy, setActionBusy] = useState("");
  const [actionError, setActionError] = useState("");
  const [actionMessage, setActionMessage] = useState("");
  const runtimeState = useQuery(() => getDesktopRuntimeInfo(), [refreshTick]);
  const diagnosis = runtimeState.data
    ? getDesktopRuntimeDiagnosis(runtimeState.data)
    : null;

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
          setActionError(result.error || "Desktop recovery action failed");
        }
      } else {
        setActionMessage(
          result.path ? `Saved to ${result.path}` : "Action completed",
        );
      }
    } catch (error) {
      setActionError(
        error instanceof Error
          ? error.message
          : "Desktop recovery action failed",
      );
    } finally {
      setActionBusy("");
    }
  }

  return (
    <section className="dashboard-page">
      <div className="dashboard-intro">
        <div>
          <p className="placeholder-section">Recovery</p>
          <h3 className="placeholder-title">Desktop runtime recovery</h3>
          <p className="placeholder-description">
            Use this page when the desktop shell starts before local core is
            ready, or when the shell needs to relaunch in safe mode for
            bootstrap triage.
          </p>
        </div>
        <div className="summary-stack">
          <span className="status-pill">
            {runtimeState.loading
              ? "probing"
              : runtimeState.data?.coreStatus || "unknown"}
          </span>
          <span className="status-pill">
            {runtimeState.data?.launchMode || "unknown"}
          </span>
          <span className="status-pill">
            {diagnosis?.status || "collecting"}
          </span>
        </div>
      </div>

      <div className="content-grid">
        <section className="data-panel">
          <div className="panel-header">
            <h3>Runtime Snapshot</h3>
            <span className="status-pill">
              {runtimeState.data?.source || "loading"}
            </span>
          </div>
          {runtimeState.loading ? (
            <p className="muted-copy">
              Probing desktop bridge and local core health.
            </p>
          ) : null}
          {runtimeState.error ? (
            <p className="error-copy">{runtimeState.error}</p>
          ) : null}
          {runtimeState.data ? (
            <>
              <div className="runtime-grid">
                <div className="runtime-field">
                  <span>Launch mode</span>
                  <strong>{runtimeState.data.launchMode}</strong>
                </div>
                <div className="runtime-field">
                  <span>Core probe</span>
                  <strong>{runtimeState.data.coreStatus}</strong>
                </div>
                <div className="runtime-field">
                  <span>HTTP status</span>
                  <strong>{runtimeState.data.coreHttpStatus || "n/a"}</strong>
                </div>
                <div className="runtime-field">
                  <span>Platform</span>
                  <strong>{runtimeState.data.platform}</strong>
                </div>
                <div className="runtime-field runtime-field-wide">
                  <span>Core base URL</span>
                  <strong>{runtimeState.data.coreBaseUrl}</strong>
                </div>
                <div className="runtime-field runtime-field-wide">
                  <span>Data directory</span>
                  <strong>{runtimeState.data.dataDirectory || "n/a"}</strong>
                </div>
              </div>
              {runtimeState.data.coreError ? (
                <p className="error-copy">{runtimeState.data.coreError}</p>
              ) : null}
              <p className="muted-copy">
                {diagnosis?.summary ||
                  runtimeState.data.issue ||
                  getDesktopRuntimeRecoveryHint(runtimeState.data)}
              </p>
            </>
          ) : null}
        </section>

        <section className="data-panel">
          <div className="panel-header">
            <h3>Structured Startup Diagnosis</h3>
            <span className="status-pill">
              {diagnosis?.issues.length || 0} issues
            </span>
          </div>
          {diagnosis?.issues.length ? (
            <div className="stack-list">
              {diagnosis.issues.map((issue) => (
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
                    {issue.code} · {issue.source} · {issue.mode}
                  </p>
                  <p>{issue.detail}</p>
                  {issue.evidence.length ? (
                    <div className="recovery-meta-list">
                      {issue.evidence.map((item) => (
                        <span key={item} className="recovery-meta-chip">
                          {item}
                        </span>
                      ))}
                    </div>
                  ) : null}
                  {issue.actions.length ? (
                    <div className="recovery-action-list">
                      {issue.actions.map((action) => (
                        <article
                          key={`${issue.code}-${action.id}`}
                          className="recovery-action-card"
                        >
                          <strong>{action.label}</strong>
                          <p>{action.description}</p>
                        </article>
                      ))}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          ) : (
            <article className="list-card recovery-issue-card">
              <div className="list-card-head">
                <strong>No active startup issue</strong>
                <span className="severity-badge severity-info">info</span>
              </div>
              <p>
                The current startup snapshot did not report a structured issue.
                Recovery stays available so you can still inspect health,
                managed core state, and exported diagnostics from one page.
              </p>
              <p>
                {runtimeState.data?.coreReachable
                  ? "Core probe is currently reachable."
                  : "Use the recovery actions below to keep triaging the current startup failure."}
              </p>
            </article>
          )}
        </section>

        <section className="data-panel">
          <div className="panel-header">
            <h3>Recovery Actions</h3>
          </div>
          {diagnosis?.mode === "managed" ? (
            <p className="muted-copy">
              Managed core mode is active. Focus on restart and log inspection
              before changing base URL settings.
            </p>
          ) : null}
          {diagnosis?.mode === "external" ? (
            <p className="muted-copy">
              External core mode is active. Focus on the configured core base
              URL and the target process outside Electron.
            </p>
          ) : null}
          <div className="stack-list">
            <button
              className="primary-button"
              disabled={actionBusy !== ""}
              onClick={() => setRefreshTick((value) => value + 1)}
            >
              Retry Health Probe
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() =>
                void runAction("start-core", () => startManagedCore())
              }
            >
              {actionBusy === "start-core"
                ? "Starting..."
                : "Start Managed Core"}
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() =>
                void runAction("restart-core", () => restartManagedCore())
              }
            >
              {actionBusy === "restart-core"
                ? "Restarting..."
                : "Restart Managed Core"}
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() =>
                void runAction("safe-mode", () => relaunchDesktopSafeMode())
              }
            >
              {actionBusy === "safe-mode"
                ? "Switching..."
                : "Relaunch in Safe Mode"}
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() =>
                void runAction("normal-mode", () => relaunchDesktopNormalMode())
              }
            >
              {actionBusy === "normal-mode"
                ? "Restarting..."
                : "Restart in Normal Mode"}
            </button>
            <button
              className="secondary-button"
              disabled={
                actionBusy !== "" || !runtimeState.data?.dataDirectory?.trim()
              }
              onClick={() =>
                void runAction("show-folder", () =>
                  showDesktopItemInFolder(
                    runtimeState.data?.dataDirectory || "",
                  ),
                )
              }
            >
              Open Data Folder
            </button>
            <button
              className="secondary-button"
              disabled={
                actionBusy !== "" || !runtimeState.data?.dataDirectory?.trim()
              }
              onClick={() =>
                void runAction("open-path", () =>
                  openDesktopPath(runtimeState.data?.dataDirectory || ""),
                )
              }
            >
              Open Data Path
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() =>
                void runAction("export-recovery", () =>
                  exportDesktopDiagnostics(
                    {
                      exported_at: new Date().toISOString(),
                      page: "recovery",
                      runtime: runtimeState.data,
                      diagnosis,
                    },
                    `easymvp-recovery-${Date.now()}.json`,
                  ),
                )
              }
            >
              {actionBusy === "export-recovery"
                ? "Exporting..."
                : "Export Recovery Snapshot"}
            </button>
            <button
              className="secondary-button"
              disabled={
                actionBusy !== "" ||
                !runtimeState.data?.coreManagerCwd?.trim()
              }
              onClick={() =>
                void runAction("show-core-cwd", () =>
                  showDesktopItemInFolder(
                    runtimeState.data?.coreManagerCwd || "",
                  ),
                )
              }
            >
              Open Core Working Dir
            </button>
            <button
              className="secondary-button"
              disabled={
                actionBusy !== "" ||
                !runtimeState.data?.coreManagerCwd?.trim()
              }
              onClick={() =>
                void runAction("open-core-cwd", () =>
                  openDesktopPath(runtimeState.data?.coreManagerCwd || ""),
                )
              }
            >
              Open Core CWD Path
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() => navigate(routes.settings)}
            >
              Open Settings
            </button>
            <button
              className="secondary-button"
              disabled={actionBusy !== ""}
              onClick={() => navigate(routes.diagnostics)}
            >
              Open Diagnostics
            </button>
          </div>
          {actionError ? <p className="error-copy">{actionError}</p> : null}
          {actionMessage ? <p className="muted-copy">{actionMessage}</p> : null}
        </section>

        {runtimeState.data?.coreManagerEnabled ? (
          <section className="data-panel">
            <div className="panel-header">
              <h3>Managed Core</h3>
              <span className="status-pill">
                {runtimeState.data.coreManagerStatus}
              </span>
            </div>
            <div className="runtime-grid">
              <div className="runtime-field">
                <span>Mode</span>
                <strong>{runtimeState.data.coreManagerMode}</strong>
              </div>
              <div className="runtime-field">
                <span>PID</span>
                <strong>{runtimeState.data.coreManagerPid || "n/a"}</strong>
              </div>
              <div className="runtime-field runtime-field-wide">
                <span>Command</span>
                <strong>
                  {runtimeState.data.coreManagerCommand || "not configured"}
                </strong>
              </div>
              <div className="runtime-field runtime-field-wide">
                <span>Args</span>
                <strong>
                  {runtimeState.data.coreManagerArgs.join(" ") || "n/a"}
                </strong>
              </div>
              <div className="runtime-field runtime-field-wide">
                <span>Working directory</span>
                <strong>{runtimeState.data.coreManagerCwd || "n/a"}</strong>
              </div>
              <div className="runtime-field">
                <span>Last exit code</span>
                <strong>
                  {runtimeState.data.coreManagerLastExitCode || "n/a"}
                </strong>
              </div>
            </div>
            <div className="recovery-callout">
              <strong>Managed core guidance</strong>
              <p>
                Managed mode means Electron owns the child process lifecycle.
                Prefer restart and log inspection here before assuming the base
                URL itself is wrong.
              </p>
            </div>
            {runtimeState.data.coreManagerLastError ? (
              <p className="error-copy">
                {runtimeState.data.coreManagerLastError}
              </p>
            ) : null}
            {runtimeState.data.coreManagerLogTail.length > 0 ? (
              <div className="list-card runtime-recovery">
                <div className="list-card-head">
                  <strong>Recent core logs</strong>
                  <span className="status-pill">
                    {runtimeState.data.coreManagerLogTail.length} lines
                  </span>
                </div>
                <pre className="runtime-log">
                  {runtimeState.data.coreManagerLogTail.join("\n")}
                </pre>
              </div>
            ) : null}
          </section>
        ) : null}
      </div>
    </section>
  );
}
