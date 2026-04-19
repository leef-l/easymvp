import { useMemo, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiGet, apiPost } from "@/shared/lib/api";
import { useProjectState } from "@/shared/lib/project";
import { useQuery } from "@/shared/lib/query";
import {
  getDesktopRuntimeInfo,
  getDesktopRuntimeRecoveryHint,
  getStoredCoreBaseUrl,
  openDesktopPath,
  selectDesktopDirectory,
  setStoredCoreBaseUrl,
  setStoredProjectId,
  showDesktopItemInFolder,
} from "@/shared/lib/preferences";
import type {
  CreateProjectPayload,
  CreateProjectResponse,
  RuntimeHealthView,
  SystemHealthView,
} from "@/shared/lib/types";
import { QueryPanel } from "@/shared/ui/QueryPanel";

export function SettingsPage() {
  const navigate = useNavigate();
  const { projectId, updateProjectId, buildRoute, routes } = useProjectState();
  const [draftProjectId, setDraftProjectId] = useState(projectId);
  const [draftBaseUrl, setDraftBaseUrl] = useState(getStoredCoreBaseUrl());
  const [saveTick, setSaveTick] = useState(0);
  const [createBusy, setCreateBusy] = useState(false);
  const [createError, setCreateError] = useState("");
  const [createMessage, setCreateMessage] = useState("");
  const [createdProjectId, setCreatedProjectId] = useState("");
  const [createPayload, setCreatePayload] = useState<CreateProjectPayload>({
    name: "",
    project_category: "web_app",
    goal_summary: "",
    workspace_root: "",
    repo_root: "",
  });

  const systemState = useQuery(
    () => apiGet<SystemHealthView>("/api/v3/system/healthz"),
    [saveTick],
  );
  const runtimeState = useQuery(
    () => apiGet<RuntimeHealthView>("/api/v3/runtime/healthz"),
    [saveTick],
  );
  const desktopRuntimeState = useQuery(() => getDesktopRuntimeInfo(), []);

  const saveSummary = useMemo(() => {
    return `Stored project ${projectId} and core ${getStoredCoreBaseUrl()}`;
  }, [projectId, saveTick]);

  function handleSave() {
    setStoredCoreBaseUrl(draftBaseUrl);
    setStoredProjectId(draftProjectId);
    updateProjectId(draftProjectId);
    setSaveTick((value) => value + 1);
  }

  async function handlePickDirectory(target: "workspace_root" | "repo_root") {
    const result = await selectDesktopDirectory();
    if (result.canceled || !result.path.trim()) {
      return;
    }
    setCreatePayload((current) => ({
      ...current,
      [target]: result.path,
    }));
  }

  async function handleOpenDataDirectory(mode: "open" | "show") {
    const targetPath = desktopRuntimeState.data?.dataDirectory?.trim() || "";
    if (!targetPath) {
      setCreateError("Desktop runtime data directory is unavailable");
      return;
    }
    const result =
      mode === "show"
        ? await showDesktopItemInFolder(targetPath)
        : await openDesktopPath(targetPath);
    if (!result.ok) {
      setCreateError(result.error || "Desktop shell path action failed");
      return;
    }
    setCreateError("");
  }

  async function handleCreateProject() {
    setCreateBusy(true);
    setCreateError("");
    setCreateMessage("");
    setCreatedProjectId("");
    try {
      const result = await apiPost<CreateProjectResponse>(
        "/api/v3/projects",
        createPayload,
      );
      const nextProjectId = result.resource_id.trim();
      if (nextProjectId) {
        setCreatedProjectId(nextProjectId);
        setDraftProjectId(nextProjectId);
        updateProjectId(nextProjectId);
        setStoredProjectId(nextProjectId);
        navigate(buildRoute("/workspace", { project: nextProjectId }));
      }
      setCreateMessage(
        `${result.next_action} · ${nextProjectId || "project created"}`,
      );
      setCreatePayload({
        name: "",
        project_category: createPayload.project_category,
        goal_summary: "",
        workspace_root: createPayload.workspace_root,
        repo_root: createPayload.repo_root,
      });
      setSaveTick((value) => value + 1);
    } catch (error) {
      setCreateError(
        error instanceof Error ? error.message : "Create project failed",
      );
    } finally {
      setCreateBusy(false);
    }
  }

  return (
    <QueryPanel
      loading={systemState.loading || runtimeState.loading}
      error={systemState.error || runtimeState.error}
      title="Settings"
      onRetry={() => setSaveTick((value) => value + 1)}
      secondaryActionLabel="Open Diagnostics"
      onSecondaryAction={() => navigate(routes.diagnostics)}
      recoveryMessage="If health checks fail, verify the core base URL and use diagnostics to inspect runtime connectivity."
    >
      {systemState.data && runtimeState.data ? (
        <section className="dashboard-page">
          <div className="dashboard-intro">
            <div>
              <p className="placeholder-section">Settings</p>
              <h3 className="placeholder-title">Local workbench preferences</h3>
              <p className="placeholder-description">{saveSummary}</p>
            </div>
            <button className="primary-button" onClick={handleSave}>
              Save and refresh
            </button>
          </div>

          <div className="content-grid">
            <section className="data-panel">
              <div className="panel-header">
                <h3>Workbench Defaults</h3>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>Project ID</span>
                  <input
                    className="project-input"
                    value={draftProjectId}
                    onChange={(event) => setDraftProjectId(event.target.value)}
                  />
                </label>
                <label className="form-field">
                  <span>Core Base URL</span>
                  <input
                    className="project-input"
                    value={draftBaseUrl}
                    onChange={(event) => setDraftBaseUrl(event.target.value)}
                  />
                </label>
              </div>
              {desktopRuntimeState.data ? (
                <div className="runtime-note">
                  <div className="runtime-note-header">
                    <strong>Desktop runtime</strong>
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
                    <div className="runtime-field runtime-field-wide">
                      <span>Runtime core base URL</span>
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
                      <span>Platform</span>
                      <strong>{desktopRuntimeState.data.platform}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>Version</span>
                      <strong>{desktopRuntimeState.data.version}</strong>
                    </div>
                    <div className="runtime-field">
                      <span>Core manager</span>
                      <strong>
                        {desktopRuntimeState.data.coreManagerStatus}
                      </strong>
                    </div>
                  </div>
                  <p className="muted-copy">
                    Renderer override currently resolves to{" "}
                    {getStoredCoreBaseUrl()}.
                  </p>
                  {desktopRuntimeState.data.coreError ? (
                    <p className="error-copy">
                      {desktopRuntimeState.data.coreError}
                    </p>
                  ) : null}
                  <p
                    className={
                      desktopRuntimeState.data.issue
                        ? "error-copy"
                        : "muted-copy"
                    }
                  >
                    {desktopRuntimeState.data.issue ||
                      getDesktopRuntimeRecoveryHint(desktopRuntimeState.data)}
                  </p>
                </div>
              ) : null}
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Create Project</h3>
                <span className="status-pill">
                  {createBusy
                    ? "creating"
                    : createdProjectId
                      ? "created"
                      : "ready"}
                </span>
              </div>
              <div className="form-grid">
                <label className="form-field">
                  <span>Name</span>
                  <input
                    className="project-input"
                    disabled={createBusy}
                    value={createPayload.name}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        name: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="form-field">
                  <span>Project Category</span>
                  <input
                    className="project-input"
                    disabled={createBusy}
                    value={createPayload.project_category}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        project_category: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="form-field">
                  <span>Goal Summary</span>
                  <textarea
                    className="project-input project-textarea"
                    disabled={createBusy}
                    value={createPayload.goal_summary}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        goal_summary: event.target.value,
                      }))
                    }
                  />
                </label>
                <label className="form-field">
                  <span>Workspace Root</span>
                  <input
                    className="project-input"
                    disabled={createBusy}
                    value={createPayload.workspace_root}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        workspace_root: event.target.value,
                      }))
                    }
                  />
                  <div className="action-row">
                    <button
                      className="secondary-button"
                      disabled={createBusy}
                      onClick={() => void handlePickDirectory("workspace_root")}
                    >
                      Choose Folder
                    </button>
                  </div>
                </label>
                <label className="form-field">
                  <span>Repo Root</span>
                  <input
                    className="project-input"
                    disabled={createBusy}
                    value={createPayload.repo_root}
                    onChange={(event) =>
                      setCreatePayload((current) => ({
                        ...current,
                        repo_root: event.target.value,
                      }))
                    }
                  />
                  <div className="action-row">
                    <button
                      className="secondary-button"
                      disabled={createBusy}
                      onClick={() => void handlePickDirectory("repo_root")}
                    >
                      Choose Folder
                    </button>
                  </div>
                </label>
                <div className="action-row">
                  <button
                    className="primary-button"
                    disabled={createBusy}
                    onClick={() => void handleCreateProject()}
                  >
                    {createBusy ? "Creating..." : "Create Project"}
                  </button>
                  {createdProjectId ? (
                    <>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/workspace", {
                              project: createdProjectId,
                            }),
                          )
                        }
                      >
                        Open Workspace
                      </button>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/plan", { project: createdProjectId }),
                          )
                        }
                      >
                        Open Plan
                      </button>
                      <button
                        className="secondary-button"
                        onClick={() =>
                          navigate(
                            buildRoute("/execution", {
                              project: createdProjectId,
                            }),
                          )
                        }
                      >
                        Open Execution
                      </button>
                    </>
                  ) : null}
                </div>
                {createError ? (
                  <p className="error-copy">{createError}</p>
                ) : null}
                {createMessage ? (
                  <p className="muted-copy">{createMessage}</p>
                ) : null}
                {!createError && !createMessage ? (
                  <p className="muted-copy">
                    Create a project, then jump directly into workspace, plan,
                    or execution triage from here.
                  </p>
                ) : null}
              </div>
            </section>

            <section className="data-panel">
              <div className="panel-header">
                <h3>Health Checks</h3>
              </div>
              <div className="stack-list">
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>System</strong>
                    <span className="status-pill">
                      {systemState.data.status}
                    </span>
                  </div>
                  <p>
                    {systemState.data.service} · {systemState.data.version} ·{" "}
                    {systemState.data.timestamp}
                  </p>
                </article>
                <article className="list-card">
                  <div className="list-card-head">
                    <strong>Runtime</strong>
                    <span className="status-pill">
                      {runtimeState.data.status}
                    </span>
                  </div>
                  <p>{runtimeState.data.base_url}</p>
                </article>
              </div>
              {desktopRuntimeState.data?.dataDirectory ? (
                <div className="action-row">
                  <button
                    className="secondary-button"
                    onClick={() => void handleOpenDataDirectory("show")}
                  >
                    Show Data Folder
                  </button>
                  <button
                    className="secondary-button"
                    onClick={() => void handleOpenDataDirectory("open")}
                  >
                    Open Data Path
                  </button>
                </div>
              ) : null}
            </section>
          </div>
        </section>
      ) : null}
    </QueryPanel>
  );
}
